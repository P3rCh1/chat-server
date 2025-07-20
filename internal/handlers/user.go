package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/P3rCh1/chat-server/internal/models"
	"github.com/P3rCh1/chat-server/internal/storage"
)

func ProfileHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("userID").(int)
		row := db.QueryRow("SELECT id, username, email, created_at FROM users WHERE id = $1", userID)
		var profile models.Profile
		err := row.Scan(&profile.ID, &profile.Username, &profile.Email, &profile.CreatedAt)
		if err != nil {
			if sql.ErrNoRows == err {
				http.Error(w, "Пользователь не найден", http.StatusBadRequest)
			} else {
				http.Error(w, "Неизвестная ошибка", http.StatusInternalServerError)
				log.Printf("Ошибка при доступе к профилю: %s", err.Error())
			}
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(profile)
	}
}

func ChangeNameHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.Context().Value("userID").(int)
		var request struct {
			NewName string `json:"username"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Ошибка чтения json", http.StatusBadRequest)
			return
		}
		if request.NewName == "" {
			http.Error(w, "Имя не может быть пустым", http.StatusBadRequest)
			return
		}
		err := storage.NewUserRepository(db).ChangeName(id, request.NewName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)

	}
}
