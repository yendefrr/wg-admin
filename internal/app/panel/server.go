package panel

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"
	"go/wg-admin/internal/app/model"
	"go/wg-admin/internal/app/services"
	"go/wg-admin/internal/app/store"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	sessionName     string = "session"
	ctxKeyUser      ctxKey = iota
	ctcKeyRequestID ctxKey = iota
)

var (
	errIncorrectEmailOrPassword = errors.New("incorrect email or password")
	errNotAuthenticated         = errors.New("not authenticated")
)

type ctxKey int8

type server struct {
	router        *mux.Router
	logger        *logrus.Logger
	store         store.Store
	sessionsStore sessions.Store
	command       services.Command
}

func newServer(store store.Store, sessionsStore sessions.Store, command services.Command) *server {
	s := &server{
		router:        mux.NewRouter(),
		logger:        logrus.New(),
		store:         store,
		sessionsStore: sessionsStore,
		command:       command,
	}

	s.configureRouter()

	return s
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *server) configureRouter() {
	s.router.Use(s.setRequestID)
	s.router.Use(s.logRequest)
	s.router.Use(handlers.CORS(handlers.AllowedOrigins([]string{"*"})))

	s.router.HandleFunc("/", s.handleIndexPage()).Methods("GET")

	s.router.HandleFunc("/get-file", s.handleStreamConfig()).Methods("GET")

	s.router.HandleFunc("/register", s.handleUserCreate()).Methods("POST")
	s.router.HandleFunc("/login", s.handleAuthorize()).Methods("POST")

	s.router.HandleFunc("/create-config", s.handleCreateConfigPage()).Methods("GET")
	s.router.HandleFunc("/create-config", s.handleCreateConfig()).Methods("POST")

	private := s.router.PathPrefix("/admin").Subrouter()
	private.Use(s.authenticateUser)
	//private.HandleFunc("/create-config", s.handleCreateConfigPage()).Methods("GET")
	//private.HandleFunc("/create-config", s.handleCreateConfig()).Methods("POST")
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

		profiles, err := s.store.Profile().GetAll()
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		if err := t.ExecuteTemplate(w, "index", profiles); err != nil {
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
			return
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

func (s *server) authenticateUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := s.sessionsStore.Get(r, sessionName)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		id, ok := session.Values["user_id"]
		if !ok {
			s.error(w, r, http.StatusUnauthorized, errNotAuthenticated)
			return
		}

		u, err := s.store.User().Find(id.(int))
		if err != nil {
			s.error(w, r, http.StatusUnauthorized, errNotAuthenticated)
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKeyUser, u)))
	})
}

func (s *server) handleCreateConfig() http.HandlerFunc {
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
		}

		if err := p.Validate(); err != nil {
			s.error(w, r, http.StatusUnprocessableEntity, err)
			return
		}

		if err := s.command.Keygen(req.Name, req.ConfigType); err != nil {
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

		//TODO: Think about transactions
		if err := s.store.Profile().Create(p); err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		if err := p.AppendPear(); err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		if err := s.command.RestartWG(); err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		if err := p.GenProfile(); err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func (s *server) handleUserCreate() http.HandlerFunc {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		u := &model.User{
			Email:    req.Email,
			Password: req.Password,
		}
		if err := s.store.User().Create(u); err != nil {
			s.error(w, r, http.StatusUnprocessableEntity, err)
			return
		}

		u.Sanitize()
		s.respond(w, r, http.StatusCreated, u)
	}
}

func (s *server) handleAuthorize() http.HandlerFunc {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		u, err := s.store.User().FindByEmail(req.Email)
		if err != nil || !u.ComparePassword(req.Password) {
			s.error(w, r, http.StatusUnauthorized, errIncorrectEmailOrPassword)
			return
		}

		session, err := s.sessionsStore.Get(r, sessionName)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		session.Values["user_id"] = u.ID
		if err := s.sessionsStore.Save(r, w, session); err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		s.respond(w, r, http.StatusOK, nil)
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
