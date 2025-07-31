package tools

import (
	"log/slog"
	"net/http"

	"github.com/P3rCh1/chat-server/internal/config"
	"github.com/P3rCh1/chat-server/internal/models"
	"github.com/P3rCh1/chat-server/internal/pkg/logger"
	"github.com/P3rCh1/chat-server/internal/pkg/tokens"
	"github.com/P3rCh1/chat-server/internal/storage/cache"
	"github.com/P3rCh1/chat-server/internal/storage/postgres"
	"github.com/P3rCh1/chat-server/internal/storage/repository"
	"github.com/gorilla/websocket"
)

type Tools struct {
	Repository    *repository.Repository
	TokenProvider tokens.TokenProvider
	Log           *slog.Logger
	WSUpgrader    *websocket.Upgrader
	Cfg           *config.Config
	PKG           *Package
}

type Package struct {
	SystemUserID int
	ErrorUserID  int
}

func MustPrepare(cfg *config.Config) *Tools {
	log := logger.New(&cfg.Logger)
	db := postgres.MustOpen(log, &cfg.DB)
	postgres.MustApplyMigrations(log, db)
	redis := cache.MustCreate(log, &cfg.Redis)
	userCacher := cache.NewStructCacher[models.Profile](redis, cfg.Redis.TTL, "user")
	roomCacher := cache.NewStructCacher[models.Room](redis, cfg.Redis.TTL, "room")
	userRoomsCacher := cache.NewUserRoomsCacher(redis, cfg.Redis.TTL, "user`s-rooms")
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
	repo := repository.NewRepository(db, userCacher, roomCacher, userRoomsCacher)
	pkg := &Package{
		SystemUserID: repo.MustCreateInternalUser(log, cfg.PKG.SystemUsername),
		ErrorUserID:  repo.MustCreateInternalUser(log, cfg.PKG.ErrorUsername),
	}
	return &Tools{
		Log:           log,
		TokenProvider: jwt,
		Repository:    repo,
		WSUpgrader:    ws,
		Cfg:           cfg,
		PKG:           pkg,
	}
}
