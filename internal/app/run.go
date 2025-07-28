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
	"github.com/P3rCh1/chat-server/internal/models"
	"github.com/P3rCh1/chat-server/internal/pkg/logger"
	"github.com/P3rCh1/chat-server/internal/pkg/tokens"
	"github.com/P3rCh1/chat-server/internal/storage/postgres"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/websocket"
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

		r.Post("/register", users.Register(tools))
		r.With(middleware.Throttle(5)).Post("/login", users.Login(tools))

		r.Group(func(authRouter chi.Router) {
			authRouter.Use(mw.Auth(tools.TokenProvider))
			authRouter.Get("/profile", users.Profile(tools))
			authRouter.Put("/change-name", users.ChangeName(tools))
			authRouter.Put("/create-room", rooms.Create(tools))
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

func mustPrepareTools(cfg *config.Config) *models.Tools {
	log := logger.New(&cfg.Logger)
	db := postgres.MustOpen(log, &cfg.DB)
	postgres.MustApplyMigrations(log, db)
	jwt := tokens.NewJWT(&cfg.JWT)
	ws := &websocket.Upgrader{
		WriteBufferSize:   cfg.WebSocket.WriteBufSize,
		ReadBufferSize:    cfg.WebSocket.ReadBufSize,
		EnableCompression: cfg.WebSocket.EnableCompression,
	}
	if cfg.WebSocket.CheckOrigin {
		ws.CheckOrigin = func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			for _, allowed := range cfg.WebSocket.AllowedOrigins {
				if origin == allowed {
					return true
				}
			}
			return false
		}
	} else {
		ws.CheckOrigin = func(r *http.Request) bool {
			return true
		}
	}
	repo := postgres.NewRepository(db)
	pkg := &models.Package{
		SystemUserID: repo.MustCreateInternalUser(log, cfg.PKG.SystemUsername),
		ErrorUserID:  repo.MustCreateInternalUser(log, cfg.PKG.ErrorUsername),
	}
	return &models.Tools{
		Log:           log,
		TokenProvider: jwt,
		DB:            db,
		WSUpgrader:    ws,
		Cfg:           cfg,
		Pkg:           pkg,
	}
}
