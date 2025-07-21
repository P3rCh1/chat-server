package app

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/P3rCh1/chat-server/internal/config"
	"github.com/P3rCh1/chat-server/internal/http-server/handlers"
	"github.com/P3rCh1/chat-server/internal/http-server/middleware"
	"github.com/P3rCh1/chat-server/internal/pkg/logger"
	"github.com/P3rCh1/chat-server/internal/pkg/tokens"
	"github.com/P3rCh1/chat-server/internal/storage/postgres"
	"github.com/go-chi/chi/v5"
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
	router := chi.NewRouter()
	router.Use(middleware.LogRequests(logger))
	router.Use(middleware.LogErrors(logger))
	router.Post("/register", handlers.Register(db, logger))
	router.Post("/login", handlers.Login(db, logger, jwt))
	router.Group(func(r chi.Router) {
		r.Use(middleware.Auth(logger, jwt))
		r.Get("/profile", handlers.Profile(db, logger))
		r.Put("/change-name", handlers.ChangeName(db, logger))
	})
	server := &http.Server{
		Addr:    cfg.HTTP.Host + ":" + cfg.HTTP.Port,
		Handler: router,
	}
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)
	go func() {
		logger.Info("starting server",
			"host", cfg.HTTP.Host,
			"port", cfg.HTTP.Port,
		)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
		}
	}()
	<-done
	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second) // * cfg...)
	defer cancel()
	logger.Info("shutting down server...")
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown error", "error", err)
	} else {
		logger.Info("server stopped gracefully")
	}
}
