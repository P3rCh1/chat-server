package handlers

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/P3rCh1/chat-server/internal/models"
	"github.com/P3rCh1/chat-server/internal/pkg/msg"
	"github.com/P3rCh1/chat-server/internal/storage/postgres"
)

func Profile(db *sql.DB, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handlers.user.Profile"
		userID := r.Context().Value("userID").(int)
		row := db.QueryRow("SELECT id, username, email, created_at FROM users WHERE id = $1", userID)
		var profile models.Profile
		err := row.Scan(&profile.ID, &profile.Username, &profile.Email, &profile.CreatedAt)
		if err != nil {
			if sql.ErrNoRows == err {
				msg.UserNotFound.DropWithLog(w, log, op)
			} else {
				msg.ServerError.DropWithLog(w, log, op)
			}
			return
		}
		msg.SendJSONWithLog(w, http.StatusOK, profile, log, op)
	}
}

func ChangeName(db *sql.DB, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handlers.user.ChangeName"
		id := r.Context().Value("userID").(int)
		var request struct {
			NewName string `json:"username"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			msg.InvalidData.DropWithLog(w, log, op)
			return
		}
		if request.NewName == "" {
			msg.EmptyFields.DropWithLog(w, log, op)
			return
		}
		err := postgres.NewUserRepository(db).ChangeName(id, request.NewName)
		if err != nil {
			msg.UserAlreadyExist.DropWithLog(w, log, op)
			return
		}
		response := struct {
			Message string `json:"message"`
		}{"name change succeed"}
		msg.SendJSONWithLog(w, http.StatusOK, response, log, op)
	}
}
