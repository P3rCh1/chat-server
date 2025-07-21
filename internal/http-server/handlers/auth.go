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
	"github.com/P3rCh1/chat-server/internal/storage/postgres"
	"github.com/P3rCh1/chat-server/pkg/email"
	"golang.org/x/crypto/bcrypt"
)

func Register(db *sql.DB, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.auth.Register"

		var user models.UserRequest
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			msg.InvalidData.DropWithLog(w, log, op)
			return
		}
		if user.Username == "" || user.Password == "" || user.Email == "" {
			msg.EmptyFields.DropWithLog(w, log, op)
			return
		}
		if !email.Check(user.Email) {
			msg.BadEmail.DropWithLog(w, log, op)
			return
		}
		hashedPass, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			msg.ServerError.DropWithLog(w, log, op+": hash password")
			return
		}
		user.Password = string(hashedPass)
		rep := postgres.NewUserRepository(db)
		profile, err := rep.CreateUser(user)
		var errHTTP msg.ErrorHTTP
		if err != nil && errors.As(err, &errHTTP) {
			errHTTP.DropWithLog(w, log, op)
			return
		}
		msg.SendJSONWithLog(w, http.StatusCreated, profile, log, op)
	}
}

func Login(db *sql.DB, log *slog.Logger, jwt tokens.TokenProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.auth.Login"
		var input models.LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			msg.InvalidData.DropWithLog(w, log, op)
			return
		}
		if input.Email == "" || input.Password == "" {
			msg.EmptyFields.DropWithLog(w, log, op)
			return
		}
		query := "SELECT id, password_hash FROM users WHERE email = $1"
		var id int
		var password string
		if err := db.QueryRow(query, input.Email).Scan(&id, &password); err != nil {
			msg.UserNotFound.DropWithLog(w, log, op)
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(password), []byte(input.Password)); err != nil {
			msg.InvalidPassword.DropWithLog(w, log, op)
			return
		}
		jwt, err := jwt.Gen(id)
		if err != nil {
			msg.ServerError.DropWithLog(w, log, op+": jwt generation")
			return
		}
		msg.SendJSON(w, http.StatusOK, map[string]string{"token": jwt})
		log.Info(op, "status", http.StatusOK)
	}
}
