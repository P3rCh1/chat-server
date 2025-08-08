package app

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/P3rCh1/chat-server/gateway-service/internal/config"
	"github.com/P3rCh1/chat-server/gateway-service/internal/gateway"
	handlers "github.com/P3rCh1/chat-server/gateway-service/internal/handlers/user"
	mw "github.com/P3rCh1/chat-server/gateway-service/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func Run(cfg *config.Config) {
	services := gateway.MustNew(cfg)
	defer services.Close()
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(mw.CORS)
	r.Group(func(r chi.Router) {
		r.Use(mw.RequestID)
		r.Use(mw.Secure)
		r.Use(middleware.Throttle(cfg.HTTP.RateLimit))
		r.Use(mw.LogRequests(services.Log))
		r.Post("/register", handlers.Register(services))
		r.Get("/profile/{userID}", handlers.AnotherProfile(services))
		//r.Get("/room/{roomID}", handlers.GetRoom(services))
		r.With(middleware.Throttle(5)).Put("/login", handlers.Login(services))
		r.Group(func(r chi.Router) {
			r.Use(mw.Auth(services))
			r.Get("/profile", handlers.MyProfile(services))
			r.Put("/change-name", handlers.ChangeName(services))
			//r.Post("/create-room", handlers.Create(services))
			// r.Put("/invite", handlers.Invite(services))
			// r.Put("/join", handlers.Join(services))
			// r.Get("/rooms", handlers.GetUserRooms(services))
			// r.Get("/messages/{roomID}", handlers.Get(services))
		})
	})
	//r.HandleFunc("/ws", ws.HandlerFunc(services))
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
	}()
	<-done
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()
	services.Log.Info("shutting down server...")
	grace := true
	// if err := ws.Shutdown(shutdownCtx, services); err != nil {
	// 	services.Log.Error(
	// 		"websocket shutdown error",
	// 		"error", err,
	// 	)
	// 	grace = false
	// }
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
