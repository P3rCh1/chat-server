package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/P3rCh1/chat-server/internal/models"
	"github.com/P3rCh1/chat-server/internal/storage"
	"github.com/P3rCh1/chat-server/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

func RegisterHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user models.UserRequest
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			http.Error(w, "Ошибка распознавания json", http.StatusBadRequest)
			return
		}
		if user.Username == "" || user.Email == "" || user.Password == "" {
			http.Error(w, "Все поля обязательны", http.StatusBadRequest)
			return
		}
		hashedPass, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Ошибка при хэшировании пароля", http.StatusInternalServerError)
			return
		}
		user.Password = string(hashedPass)
		rep := storage.NewUserRepository(db)
		profile, err := rep.CreateUser(user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(profile)
	}
}

func LoginHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input models.LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "Ошибка распознавания json", http.StatusBadRequest)
			return
		}
		query := "SELECT id, password_hash FROM users WHERE email = $1"
		var id int
		var password string
		if err := db.QueryRow(query, input.Email).Scan(&id, &password); err != nil {
			http.Error(w, "Пользователь не найден", http.StatusUnauthorized)
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(password), []byte(input.Password)); err != nil {
			http.Error(w, "Неверный пароль", http.StatusUnauthorized)
			return
		}
		token, err := utils.GenJWT(id)
		if err != nil {
			http.Error(w, "Ошибка создания токена", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"token": token})
	}
}
