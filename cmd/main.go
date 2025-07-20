package main

import (
	"log"
	"net/http"

	"github.com/P3rCh1/chat-server/internal/handlers"
	"github.com/P3rCh1/chat-server/internal/middleware"
	"github.com/P3rCh1/chat-server/internal/storage"
)

func main() {
	db, err := storage.InitDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	err = storage.ApplyMigrations(db)
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/register", handlers.RegisterHandler(db))
	http.HandleFunc("/login", handlers.LoginHandler(db))
	http.HandleFunc("/profile", middleware.JWTAuth(handlers.ProfileHandler(db)))
	http.HandleFunc("/change-name", middleware.JWTAuth(handlers.ChangeNameHandler(db)))
	log.Println("Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
