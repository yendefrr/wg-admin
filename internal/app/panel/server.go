package panel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go/wg-admin/internal/app/model"
	"go/wg-admin/internal/app/services"
	"go/wg-admin/internal/app/store"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

const (
	ctcKeyRequestID ctxKey = iota
)

type ctxKey int8

type server struct {
	router        *mux.Router
	logger        *logrus.Logger
	store         store.Store
	command       services.Command
	events        *kafka.Conn
	storage       *redis.Client
	sessionsStore sessions.Store
}

func newServer(store store.Store, sessionsStore sessions.Store, command services.Command, events *kafka.Conn, storage *redis.Client) *server {
	s := &server{
		router:        mux.NewRouter(),
		logger:        logrus.New(),
		store:         store,
		sessionsStore: sessionsStore,
		command:       command,
		events:        events,
		storage:       storage,
	}

	s.configureRouter()

	return s
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *server) configureRouter() {
	path, _ := os.Getwd()

	s.router.Use(s.setRequestID)
	s.router.Use(s.logRequest)
	s.router.Use(handlers.CORS(handlers.AllowedOrigins([]string{"*"})))

	s.router.HandleFunc("/", s.handleIndexPage()).Methods("GET")

	s.router.Handle("/img/{file}", http.StripPrefix("/img/", http.FileServer(http.Dir(path+"/web/img")))).Methods("GET")

	s.router.HandleFunc("/get-file", s.handleStreamConfig()).Methods("GET")
	s.router.HandleFunc("/remove-config", s.handleConfigDelete()).Methods("GET")
	s.router.HandleFunc("/remove-config-request", s.handleConfigRequestDelete()).Methods("GET")

	s.router.HandleFunc("/create-user", s.handleUserCreate()).Methods("POST")

	s.router.HandleFunc("/make-config", s.handleCreateConfigPage()).Methods("GET")
	s.router.HandleFunc("/create-config-request", s.handleConfigCreateRequest()).Methods("POST")
	s.router.HandleFunc("/create-config", s.handleConfigCreate()).Methods("GET")
}

func (s *server) setRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New().String()
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctcKeyRequestID, id)))
	})
}

func (s *server) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := s.logger.WithFields(logrus.Fields{
			"remote_addr": r.RemoteAddr,
			"request_id":  r.Context().Value(ctcKeyRequestID),
		})

		logger.Infof("started %s %s", r.Method, r.RequestURI)

		start := time.Now()
		rw := &responseWriter{w, http.StatusOK}
		next.ServeHTTP(rw, r)

		logger.Infof(
			"completed with %d %s in %v",
			rw.code,
			http.StatusText(rw.code),
			time.Now().Sub(start),
		)
	})
}

func (s *server) handleIndexPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path, err := os.Getwd()
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		t, err := template.ParseFiles(path+"/web/templates/index.html", path+"/web/templates/header.html", path+"/web/templates/footer.html")
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		users, err := s.store.User().GetAll()
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}
		profilesActive, err := s.store.Profile().GetAll(true)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}
		profilesInActive, err := s.store.Profile().GetAll(false)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		params := map[string]interface{}{}
		params["users"] = users
		params["profilesActive"] = profilesActive
		params["profilesInActive"] = profilesInActive

		fmt.Println(profilesInActive)
		fmt.Println(profilesActive)

		if err := t.ExecuteTemplate(w, "index", params); err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}
	}
}

func (s *server) handleStreamConfig() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.FormValue("id"))
		if err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		profile, err := s.store.Profile().Find(id)
		if err != nil {
			logrus.Debugln(err)
		}

		file, err := ioutil.ReadFile(profile.Path + "wg.conf")
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		b := bytes.NewBuffer(file)

		w.Header().Set("Content-Disposition", "attachment; filename=\"wg.conf\"")
		w.Header().Set("Content-Type", "application/octet-stream")

		if _, err := b.WriteTo(w); err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
		}
	}
}

func (s *server) handleCreateConfigPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path, _ := os.Getwd()
		t, _ := template.ParseFiles(path+"/web/templates/create-config.html", path+"/web/templates/header.html", path+"/web/templates/footer.html")

		err := t.ExecuteTemplate(w, "create-config", nil)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
		}
	}
}

func (s *server) handleConfigCreateRequest() http.HandlerFunc {
	type request struct {
		Name       string
		ConfigType string
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{
			Name:       r.FormValue("name"),
			ConfigType: r.FormValue("config_type"),
		}

		p := &model.Profile{
			Username: req.Name,
			Type:     req.ConfigType,
			Path:     fmt.Sprintf("/etc/wireguard/%s/%s/", req.Name, req.ConfigType),
			IsActive: false,
		}

		if err := p.Validate(); err != nil {
			s.error(w, r, http.StatusUnprocessableEntity, err)
			return
		}

		s.events.SetWriteDeadline(time.Now().Add(time.Second * 10))
		s.events.WriteMessages(
			kafka.Message{
				Key:   []byte(p.Username),
				Value: []byte(p.Type),
			},
		)

		//TODO: Think about transactions
		if err := s.store.Profile().Create(p); err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func (s *server) handleConfigCreate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.Atoi(r.FormValue("id"))

		p, err := s.store.Profile().Find(id)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		if err := s.command.Keygen(p.Username, p.Type); err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		publickey, privatekey, err := p.ReadKeys()
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}
		p.Publickey = publickey
		p.Privatekey = privatekey
		p.IsActive = true

		if err := p.AppendPear(); err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		if err := s.command.RestartWG(); err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		if err := s.store.Profile().Update(p); err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func (s *server) handleConfigRequestDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.Atoi(r.FormValue("id"))

		err := s.store.Profile().Delete(id)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func (s *server) handleConfigDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.Atoi(r.FormValue("id"))

		p, err := s.store.Profile().Find(id)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		if err := p.DelProfileFiles(); err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		err = s.store.Profile().Delete(id)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func (s *server) handleUserCreate() http.HandlerFunc {
	type request struct {
		Username string `json:"username"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		u := &model.User{
			Username: req.Username,
		}
		if err := s.store.User().Create(u); err != nil {
			s.error(w, r, http.StatusUnprocessableEntity, err)
			return
		}

		s.respond(w, r, http.StatusCreated, u)
	}
}

func (s *server) error(w http.ResponseWriter, r *http.Request, code int, err error) {
	s.respond(w, r, code, map[string]string{"error": err.Error()})
}

func (s *server) respond(w http.ResponseWriter, r *http.Request, code int, data interface{}) {
	w.WriteHeader(code)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}
