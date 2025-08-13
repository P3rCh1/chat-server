package websocket

import (
	"net/http"

	"github.com/P3rCh1/chat-server/gateway-service/internal/config"
	"github.com/P3rCh1/chat-server/gateway-service/internal/gateway"
	"github.com/gorilla/websocket"
)

type WS struct {
	services *gateway.Services
	cfg      *config.Websocket
}

func newWS(cfg *config.Websocket, s *gateway.Services) *WS {
	return &WS{
		cfg:      cfg,
		services: s,
	}
}

func newUpgrader(cfg config.Websocket) *websocket.Upgrader {
	ws := &websocket.Upgrader{
		WriteBufferSize:   cfg.WriteBufSize,
		ReadBufferSize:    cfg.ReadBufSize,
		EnableCompression: cfg.EnableCompression,
	}
	if cfg.CheckOrigin {
		ws.CheckOrigin = func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			for _, allowed := range cfg.AllowedOrigins {
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
	return ws
}
