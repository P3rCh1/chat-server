package config

import (
	"database/sql"
	"log/slog"
	"os"

	"github.com/P3rCh1/chat-server/internal/http-server/handlers"
	"github.com/P3rCh1/chat-server/internal/storage"
)

func initDB() (*sql.DB, error) {
	return storage.NewDB(storage.Config{
		Host:     "localhost",
		Port:     5432,
		User:     "chat",
		Password: "pass",
		DBName:   "chatdb",
	})
}

func initLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
}

func InitHandler() (*handlers.Handler, error) {
	db, err := initDB()
	return handlers.NewHandler(db, initLogger()), err
}
