package users

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/P3rCh1/chat-server/internal/models"
	"github.com/P3rCh1/chat-server/internal/pkg/logger"
	"github.com/P3rCh1/chat-server/internal/pkg/responses"
	"github.com/P3rCh1/chat-server/internal/pkg/tools"
	"github.com/P3rCh1/chat-server/internal/pkg/validate"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func ValidateRegister(r *models.RegisterRequest) responses.ErrorHTTP {
	if r.Username == "" || r.Password == "" || r.Email == "" {
		return responses.EmptyFields
	}
	if !validate.Email(r.Email) {
		return responses.BadEmail
	}
	if !validate.Name(r.Username) {
		return responses.BadName
	}
	if !validate.Password(r.Password) {
		return responses.BadPassword
	}
	return nil
}

func Register(tools *tools.Tools) http.HandlerFunc {
	const op = "internal.http-server.handlers.users.auth.Register"
	return func(w http.ResponseWriter, r *http.Request) {
		var user models.RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			responses.InvalidData.Drop(w)
			return
		}
		if err := ValidateRegister(&user); err != nil {
			err.Drop(w)
			return
		}
		hashedPass, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			responses.ServerError.Drop(w)
			logger.LogError(tools.Log, op, err)
			return
		}
		user.Password = string(hashedPass)
		profile, err := tools.Repository.CreateUser(user)
		if err != nil {
			var pqErr *pq.Error
			if errors.As(err, &pqErr) && pqErr.Code == "23505" {
				responses.UserOrEmailAlreadyExist.Drop(w)
			} else {
				responses.ServerError.Drop(w)
				logger.LogError(tools.Log, op, err)
			}
			return
		}
		responses.SendJSON(w, http.StatusCreated, profile)
	}
}

func Login(tools *tools.Tools) http.HandlerFunc {
	const op = "internal.http-server.handlers.users.auth.Login"
	return func(w http.ResponseWriter, r *http.Request) {
		var input models.LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			responses.InvalidData.Drop(w)
			return
		}
		if input.Email == "" || input.Password == "" {
			responses.EmptyFields.Drop(w)
			return
		}
		query := "SELECT id, password_hash FROM users WHERE email = $1"
		var id int
		var password string
		if err := tools.Repository.DB.QueryRow(query, input.Email).Scan(&id, &password); err != nil {
			if err == sql.ErrNoRows {
				responses.UserNotFound.Drop(w)
			} else {
				responses.ServerError.Drop(w)
				logger.LogError(tools.Log, op, err)
			}
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(password), []byte(input.Password)); err != nil {
			responses.InvalidPassword.Drop(w)
			return
		}
		jwt, err := tools.TokenProvider.Gen(id)
		if err != nil {
			responses.ServerError.Drop(w)
			logger.LogError(tools.Log, op, err)
			return
		}
		responses.SendJSON(w, http.StatusOK, map[string]string{"token": jwt})
	}
}
