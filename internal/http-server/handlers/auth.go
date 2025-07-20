package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/P3rCh1/chat-server/internal/pkg/models"
	"github.com/P3rCh1/chat-server/internal/pkg/msg"
	"github.com/P3rCh1/chat-server/internal/storage"
	"github.com/P3rCh1/chat-server/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.auth.Register"

	var user models.UserRequest
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		msg.InvalidData.DropWithLog(w, h.Log, op)
		return
	}
	if user.Username == "" || user.Password == "" || user.Email == "" {
		msg.EmptyFields.DropWithLog(w, h.Log, op)
		return
	}
	if !utils.CheckEmail(user.Email) {
		msg.BadEmail.DropWithLog(w, h.Log, op)
		return
	}
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		msg.ServerError.DropWithLog(w, h.Log, op+": hash password")
		return
	}
	user.Password = string(hashedPass)
	rep := storage.NewUserRepository(h.DB)
	profile, err := rep.CreateUser(user)
	var errHTTP msg.ErrorHTTP
	if err != nil && errors.As(err, &errHTTP) {
		errHTTP.DropWithLog(w, h.Log, op)
		return
	}
	msg.SendJSONWithLog(w, http.StatusCreated, profile, h.Log, op)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.auth.Login"
	var input models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		msg.InvalidData.DropWithLog(w, h.Log, op)
		return
	}
	if input.Email == "" || input.Password == "" {
		msg.EmptyFields.DropWithLog(w, h.Log, op)
		return
	}
	query := "SELECT id, password_hash FROM users WHERE email = $1"
	var id int
	var password string
	if err := h.DB.QueryRow(query, input.Email).Scan(&id, &password); err != nil {
		msg.UserNotFound.DropWithLog(w, h.Log, op)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(password), []byte(input.Password)); err != nil {
		msg.InvalidPassword.DropWithLog(w, h.Log, op)
		return
	}
	token, err := utils.GenJWT(id)
	if err != nil {
		msg.ServerError.DropWithLog(w, h.Log, op+": jwt generation")
		return
	}
	msg.SendJSON(w, http.StatusOK, map[string]string{"token": token})
	h.Log.Info(op,
		"status", http.StatusOK,
	)
}
