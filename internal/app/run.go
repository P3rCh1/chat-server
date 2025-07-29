package app

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/P3rCh1/chat-server/internal/config"
	"github.com/P3rCh1/chat-server/internal/http-server/handlers/rooms"
	"github.com/P3rCh1/chat-server/internal/http-server/handlers/users"
	ws "github.com/P3rCh1/chat-server/internal/http-server/handlers/websocket"
	mw "github.com/P3rCh1/chat-server/internal/http-server/middleware"
	"github.com/P3rCh1/chat-server/internal/pkg/tools"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func Run(cfg *config.Config) {
	tools := tools.MustPrepare(cfg)
	defer tools.Repository.DB.Close()
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Group(func(r chi.Router) {
		r.Use(mw.RequestID)
		r.Use(mw.Secure)
		r.Use(middleware.Throttle(cfg.HTTP.RateLimit))
		r.Use(mw.LogRequests(tools.Log))
		r.Use(mw.LogErrors(tools.Log))
		r.Post("/register", users.Register(tools))
		r.With(middleware.Throttle(5)).Put("/login", users.Login(tools))
		r.Group(func(authRouter chi.Router) {
			authRouter.Use(mw.Auth(tools.TokenProvider))
			authRouter.Get("/profile", users.Profile(tools))
			authRouter.Put("/change-name", users.ChangeName(tools))
			authRouter.Post("/create-room", rooms.Create(tools))
			authRouter.Put("/invite", rooms.Invite(tools))
			authRouter.Put("/join", rooms.Join(tools))
		})
	})
	r.HandleFunc("/ws", ws.HandlerFunc(tools))
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
			tools.Log.Error(
				"server error",
				"error", err,
			)
		}
	}()
	<-done
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()
	tools.Log.Info("shutting down server...")
	grace := true
	if err := ws.Shutdown(shutdownCtx, tools); err != nil {
		tools.Log.Error(
			"websocket shutdown error",
			"error", err,
		)
		grace = false
	}
	if err := server.Shutdown(shutdownCtx); err != nil {
		tools.Log.Error(
			"server shutdown error",
			"error", err,
		)
		grace = false
	}
	if grace {
		tools.Log.Info("server stopped gracefully")
	}
}
