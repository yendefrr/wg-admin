package panel

import (
	"context"
	"database/sql"
	"go/wg-admin/internal/app/services/commands"
	"go/wg-admin/internal/app/store/sqlstore"
	"net/http"

	"github.com/segmentio/kafka-go"

	"github.com/gorilla/sessions"

	_ "github.com/go-sql-driver/mysql"
)

func Start(config *Config) error {
	db, err := newDB(config.DatabaseURL)
	if err != nil {
		return err
	}

	defer db.Close()

	store := sqlstore.New(db)
	sessionStore := sessions.NewCookieStore([]byte(config.SessionKey))
	command := commands.NewCommand(config.CommandsPath)
	events, err := kafka.DialLeader(context.Background(), "tcp", "localhost:9092", "requests", 0)
	if err != nil {
		return err
	}

	server := newServer(store, sessionStore, events, command)

	return http.ListenAndServe(config.BindAddr, server)
}

func newDB(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("mysql", databaseURL)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, err
}
