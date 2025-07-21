package app

import (
	"fmt"
	"log"
	"net/http"

	"github.com/P3rCh1/chat-server/internal/config"
	"github.com/P3rCh1/chat-server/internal/http-server/handlers"
	"github.com/P3rCh1/chat-server/internal/http-server/middleware"
	"github.com/P3rCh1/chat-server/internal/pkg/logger"
	"github.com/P3rCh1/chat-server/internal/pkg/tokens"
	"github.com/P3rCh1/chat-server/internal/storage/postgres"
)

func Run(cfg *config.Config) {
	logger := logger.New(&cfg.Logger)
	db, err := postgres.New(&cfg.DB)
	if err != nil {
		logger.Error("storage.postgres.New", "error", err.Error())
	}
	defer db.Close()
	err = postgres.ApplyMigrations(db)
	if err != nil {
		logger.Error("internal.storage.postgres.ApplyMigrations", "error", err.Error())
	}
	jwt := tokens.NewJWT(&cfg.JWT)
	http.HandleFunc("/register", handlers.Register(db, logger))
	http.HandleFunc("/login", handlers.Login(db, logger, jwt))
	http.HandleFunc("/profile", middleware.JWTAuth(logger, jwt, handlers.Profile(db, logger)))
	http.HandleFunc("/change-name", middleware.JWTAuth(logger, jwt, handlers.ChangeName(db, logger)))
	logger.Info(fmt.Sprintf("server is running on %s:%s", cfg.HTTP.Host, cfg.HTTP.Port))
	log.Fatal(http.ListenAndServe(":"+cfg.HTTP.Port, nil))
}
