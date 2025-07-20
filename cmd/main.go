package main

import (
	"log"
	"net/http"

	"github.com/P3rCh1/chat-server/internal/config"
	_ "github.com/P3rCh1/chat-server/internal/http-server/handlers"
	"github.com/P3rCh1/chat-server/internal/http-server/middleware"
	"github.com/P3rCh1/chat-server/internal/storage"
)

func main() {
	h, err := config.InitHandler()
	if err != nil {
		log.Fatal(err)
	}
	defer h.DB.Close()
	err = storage.ApplyMigrations(h.DB)
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/register", h.Register)
	http.HandleFunc("/login", h.Login)
	http.HandleFunc("/profile", middleware.JWTAuth(h.Log, h.Profile))
	http.HandleFunc("/change-name", middleware.JWTAuth(h.Log, h.ChangeName))
	h.Log.Info("Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
