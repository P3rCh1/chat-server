package app

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/P3rCh1/chat-server/internal/config"
	"github.com/P3rCh1/chat-server/internal/http-server/handlers"
	mw "github.com/P3rCh1/chat-server/internal/http-server/middleware"
	"github.com/P3rCh1/chat-server/internal/models"
	"github.com/P3rCh1/chat-server/internal/pkg/logger"
	"github.com/P3rCh1/chat-server/internal/pkg/tokens"
	"github.com/P3rCh1/chat-server/internal/storage/postgres"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func Run(cfg *config.Config) {
	tools := mustPrepareTools(cfg)
	defer tools.DB.Close()
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Group(func(r chi.Router) {
		r.Use(mw.RequestID)
		r.Use(mw.Secure)
		r.Use(middleware.Throttle(cfg.HTTP.RateLimit))
		r.Use(mw.LogRequests(tools.Log))
		r.Use(mw.LogErrors(tools.Log))

		r.Post("/register", handlers.Register(tools))
		r.With(middleware.Throttle(5)).Post("/login", handlers.Login(tools))

		r.Group(func(authRouter chi.Router) {
			authRouter.Use(mw.Auth(tools.TokenProvider))
			authRouter.Get("/profile", handlers.Profile(tools))
			authRouter.Put("/change-name", handlers.ChangeName(tools))
		})
	})
	r.HandleFunc("/ws", handlers.Websocket(tools))
	server := &http.Server{
		Addr:         cfg.HTTP.Host + ":" + cfg.HTTP.Port,
		Handler:      r,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)
	go func() {
		tools.Log.Info("starting server",
			"host", cfg.HTTP.Host,
			"port", cfg.HTTP.Port,
		)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			tools.Log.Error("server error", "error", err)
		}
	}()
	<-done
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()
	tools.Log.Info("shutting down server...")
	if err := server.Shutdown(shutdownCtx); err != nil {
		tools.Log.Error("server shutdown error", "error", err)
	} else {
		tools.Log.Info("server stopped gracefully")
	}
}

func mustPrepareTools(cfg *config.Config) *models.Tools {
	log := logger.New(&cfg.Logger)
	db, err := postgres.New(&cfg.DB)
	if err != nil {
		log.Error("storage.postgres.New", "error", err.Error())
		os.Exit(1)
	}
	err = postgres.ApplyMigrations(db)
	if err != nil {
		log.Error("internal.storage.postgres.ApplyMigrations", "error", err.Error())
		os.Exit(1)
	}
	jwt := tokens.NewJWT(&cfg.JWT)
	return &models.Tools{
		Log:           log,
		TokenProvider: jwt,
		DB:            db,
	}
}
