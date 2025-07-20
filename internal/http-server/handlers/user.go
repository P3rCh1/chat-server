package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/P3rCh1/chat-server/internal/pkg/models"
	"github.com/P3rCh1/chat-server/internal/pkg/msg"
	"github.com/P3rCh1/chat-server/internal/storage"
)

func (h *Handler) Profile(w http.ResponseWriter, r *http.Request) {
	const op = "http-server.handlers.user.Profile"
	userID := r.Context().Value("userID").(int)
	row := h.DB.QueryRow("SELECT id, username, email, created_at FROM users WHERE id = $1", userID)
	var profile models.Profile
	err := row.Scan(&profile.ID, &profile.Username, &profile.Email, &profile.CreatedAt)
	if err != nil {
		if sql.ErrNoRows == err {
			msg.UserNotFound.DropWithLog(w, h.Log, op)
		} else {
			msg.ServerError.DropWithLog(w, h.Log, op)
		}
		return
	}
	msg.SendJSONWithLog(w, http.StatusOK, profile, h.Log, op)
}

func (h *Handler) ChangeName(w http.ResponseWriter, r *http.Request) {
	const op = "http-server.handlers.user.ChangeName"
	id := r.Context().Value("userID").(int)
	var request struct {
		NewName string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		msg.InvalidData.DropWithLog(w, h.Log, op)
		return
	}
	if request.NewName == "" {
		msg.EmptyFields.DropWithLog(w, h.Log, op)
		return
	}
	err := storage.NewUserRepository(h.DB).ChangeName(id, request.NewName)
	if err != nil {
		msg.UserAlreadyExist.DropWithLog(w, h.Log, op)
		return
	}
	response := struct {
		Message string `json:"message"`
	}{"name change succeed"}
	msg.SendJSONWithLog(w, http.StatusOK, response, h.Log, op)
}
