package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/P3rCh1/chat-server/gateway-service/internal/config"
	"github.com/P3rCh1/chat-server/gateway-service/internal/gateway"
	"github.com/P3rCh1/chat-server/gateway-service/internal/handlers/message"
	"github.com/P3rCh1/chat-server/gateway-service/internal/handlers/rooms"
	"github.com/P3rCh1/chat-server/gateway-service/internal/handlers/user"
	"github.com/P3rCh1/chat-server/gateway-service/internal/handlers/websocket"
	mw "github.com/P3rCh1/chat-server/gateway-service/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func Run(cfg *config.Config) {
	services := gateway.MustNew(cfg)
	defer services.Close()
	r := AddHandlers(cfg, services)
	server := &http.Server{
		Addr:         cfg.HTTP.Host + ":" + cfg.HTTP.Port,
		Handler:      r,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)
	go serve(cfg, services, server)
	<-done
	Shutdown(cfg, services, server)
}

func AddHandlers(cfg *config.Config, services *gateway.Services) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(mw.CORS)
	r.Group(func(r chi.Router) {
		r.Use(mw.RequestID)
		r.Use(mw.Secure)
		r.Use(middleware.Throttle(cfg.HTTP.RateLimit))
		r.Use(mw.LogRequests(services.Log))
		r.Use(mw.LogInternalErrors(services.Log))
		r.Post("/register", user.Register(services))
		r.Get(fmt.Sprintf("/profile/{%s}", user.URLParam), user.AnotherProfile(services))
		r.Get(fmt.Sprintf("/room/{%s}", rooms.URLParam), rooms.Get(services))
		r.With(middleware.Throttle(5)).Put("/login", user.Login(services))
		r.Group(func(r chi.Router) {
			r.Use(mw.Auth(services))
			r.Get("/profile", user.MyProfile(services))
			r.Put("/change-name", user.ChangeName(services))
			r.Post("/create-room", rooms.Create(services))
			r.Put("/invite", rooms.Invite(services))
			r.Put("/join", rooms.Join(services))
			r.Get("/rooms", rooms.UserIn(services))
			r.Get(fmt.Sprintf("/messages/{%s}", message.URLParam), message.Get(services))
		})
	})
	r.HandleFunc("/ws", websocket.Connector(cfg, services))
	return r
}

func serve(cfg *config.Config, services *gateway.Services, server *http.Server) {
	services.Log.Info("starting server",
		"host", cfg.HTTP.Host,
		"port", cfg.HTTP.Port,
	)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		services.Log.Error(
			"server error",
			"error", err,
		)
	}
}

func Shutdown(cfg *config.Config, services *gateway.Services, server *http.Server) {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()
	services.Log.Info("shutting down server...")
	grace := true
	if err := websocket.Shutdown(shutdownCtx); err != nil {
		services.Log.Error(
			"websocket shutdown error",
			"error", err,
		)
		grace = false
	}
	if err := server.Shutdown(shutdownCtx); err != nil {
		services.Log.Error(
			"server shutdown error",
			"error", err,
		)
		grace = false
	}
	if grace {
		services.Log.Info("server stopped gracefully")
	}
}
