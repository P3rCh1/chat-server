package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/P3rCh1/chat-server/internal/models"
	"github.com/P3rCh1/chat-server/internal/pkg/msg"
	"github.com/P3rCh1/chat-server/internal/pkg/tokens"
	"github.com/P3rCh1/chat-server/internal/pkg/validate"
	"github.com/P3rCh1/chat-server/internal/storage/postgres"
	"golang.org/x/crypto/bcrypt"
)

func Register(db *sql.DB, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user models.UserRequest
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			msg.InvalidData.Drop(w)
			return
		}
		if user.Username == "" || user.Password == "" || user.Email == "" {
			msg.EmptyFields.Drop(w)
			return
		}
		if !validate.Email(user.Email) {
			msg.BadEmail.Drop(w)
			return
		}
		if !validate.Username(user.Username) {
			msg.BadUsername.Drop(w)
			return
		}
		if !validate.Password(user.Password) {
			msg.BadPassword.Drop(w)
			return
		}
		hashedPass, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			msg.ServerError.Drop(w)
			return
		}
		user.Password = string(hashedPass)
		rep := postgres.NewUserRepository(db)
		profile, err := rep.CreateUser(user)
		var errHTTP msg.ErrorHTTP
		if err != nil && errors.As(err, &errHTTP) {
			errHTTP.Drop(w)
			return
		}
		msg.SendJSON(w, http.StatusCreated, profile)
	}
}

func Login(db *sql.DB, log *slog.Logger, jwt tokens.TokenProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input models.LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			msg.InvalidData.Drop(w)
			return
		}
		if input.Email == "" || input.Password == "" {
			msg.EmptyFields.Drop(w)
			return
		}
		query := "SELECT id, password_hash FROM users WHERE email = $1"
		var id int
		var password string
		if err := db.QueryRow(query, input.Email).Scan(&id, &password); err != nil {
			msg.UserNotFound.Drop(w)
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(password), []byte(input.Password)); err != nil {
			msg.InvalidPassword.Drop(w)
			return
		}
		jwt, err := jwt.Gen(id)
		if err != nil {
			msg.ServerError.Drop(w)
			return
		}
		msg.SendJSON(w, http.StatusOK, map[string]string{"token": jwt})
	}
}
