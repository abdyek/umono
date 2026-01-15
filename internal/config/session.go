package config

import (
	"os"
	"time"

	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/sqlite3"
)

func NewSessionStore() *session.Store {
	driver := os.Getenv("SESSION_DRIVER")

	cfg := session.Config{
		Expiration: 72 * time.Hour,
	}

	if driver == "db" {
		cfg.Storage = sqlite3.New(sqlite3.Config{
			Database: "./session_storage.db",
			Table:    "sessions",
		})
	}

	return session.New(cfg)
}
