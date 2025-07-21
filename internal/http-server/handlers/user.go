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
		userID := r.Context().Value("userID").(int)
		row := db.QueryRow("SELECT id, username, email, created_at FROM users WHERE id = $1", userID)
		var profile models.Profile
		err := row.Scan(&profile.ID, &profile.Username, &profile.Email, &profile.CreatedAt)
		if err != nil {
			if sql.ErrNoRows == err {
				msg.UserNotFound.Drop(w)
			} else {
				msg.ServerError.Drop(w)
			}
			return
		}
		msg.SendJSON(w, http.StatusOK, profile)
	}
}

func ChangeName(db *sql.DB, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.Context().Value("userID").(int)
		var request struct {
			NewName string `json:"username"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			msg.InvalidData.Drop(w)
			return
		}
		if request.NewName == "" {
			msg.EmptyFields.Drop(w)
			return
		}
		err := postgres.NewUserRepository(db).ChangeName(id, request.NewName)
		if err != nil {
			msg.UserAlreadyExist.Drop(w)
			return
		}
		response := struct {
			Message string `json:"message"`
		}{"name change succeed"}
		msg.SendJSON(w, http.StatusOK, response)
	}
}
