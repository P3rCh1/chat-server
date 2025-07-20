package handlers

import (
	"database/sql"
	"log/slog"
)

type Handler struct {
	DB     *sql.DB
	Log *slog.Logger
}

func NewHandler(db *sql.DB, logger *slog.Logger) *Handler {
	return &Handler{db, logger}
}
