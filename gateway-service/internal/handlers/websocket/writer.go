package websocket

import (
	"time"

	"github.com/P3rCh1/chat-server/gateway-service/internal/models"
	"github.com/gorilla/websocket"
)

func (h *connectionHandler) pinger() {
	defer h.cancel()
	ticker := time.NewTicker(h.ws.cfg.PingPeriod)
	defer ticker.Stop()
	failedPings := 0
	for {
		var err error
		select {
		case <-h.ctx.Done():
			return
		default:
		}
		select {
		case <-ticker.C:
			err = h.SyncWriteMessage(websocket.PingMessage, nil)
		case <-h.ctx.Done():
			return
		}
		if err != nil {
			failedPings++
			if failedPings > h.ws.cfg.MaxFailedPings {
				return
			}
		} else {
			failedPings = 0
		}
	}
}

func broadcast(ws *WS, msg *models.Message) {
	const op = "websocket.writer.broadcast"
	mu.RLock()
	defer mu.RUnlock()
	handlers, ok := handlersInRoom[msg.RoomID]
	if !ok {
		return
	}
	for h := range handlers {
		if err := h.SyncWriteJSON(msg); err != nil {
			ws.services.Log.Warn(
				op,
				"error", err,
				"messageID", msg.ID,
			)
		}
	}
}

func (h *connectionHandler) SyncWriteJSON(v any) error {
	h.writeMutex.Lock()
	defer h.writeMutex.Unlock()
	h.conn.SetWriteDeadline(time.Now().Add(h.ws.cfg.WriteWait))
	return h.conn.WriteJSON(v)
}

func (h *connectionHandler) SyncWriteMessage(messageType int, data []byte) error {
	h.writeMutex.Lock()
	defer h.writeMutex.Unlock()
	h.conn.SetWriteDeadline(time.Now().Add(h.ws.cfg.WriteWait))
	return h.conn.WriteMessage(messageType, data)
}
