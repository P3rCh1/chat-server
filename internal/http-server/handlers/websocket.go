package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/P3rCh1/chat-server/internal/models"
	"github.com/P3rCh1/chat-server/internal/pkg/msg"
	"github.com/P3rCh1/chat-server/internal/storage/postgres"
	"github.com/gorilla/websocket"
)

var onlineInRooms sync.Map

type handler struct {
	client   models.Client
	conn     *websocket.Conn
	tools    *models.Tools
	sendChan chan *models.Message
	closeCtx context.Context
	cancel   context.CancelFunc
}

func newHandler(userID int, tools *models.Tools) (*handler, error) {
	var h handler
	h.sendChan = make(chan *models.Message, tools.Cfg.WebSocket.MsgBufSize)
	h.closeCtx, h.cancel = context.WithCancel(context.Background())
	h.tools = tools
	h.client = models.Client{
		UserID: userID,
		RoomID: 0,
	}
	var err error
	h.client.Username, err = postgres.NewRepository(tools.DB).GetUsername(userID)
	if err != nil {
		return nil, err
	}
	return &h, nil
}

func ShutdownWS(ctx context.Context, tools *models.Tools) bool {
	success := true
	onlineInRooms.Range(func(_, room interface{}) bool {
		room.(*sync.Map).Range(func(h, _ interface{}) bool {
			select {
			case <-ctx.Done():
				success = false
				return false
			default:
				h.(*handler).cancel()
				return true
			}
		})
		return true
	})
	if !success {
		tools.Log.Warn("shutdown timer end before close websocket")
	}
	return success
}

func Websocket(tools *models.Tools) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		userID, err := tools.TokenProvider.Verify(token)
		if err != nil {
			msg.UserNotFound.Drop(w)
			tools.Log.Info(
				"user not found",
				"token", token,
			)
			return
		}
		h, err := newHandler(userID, tools)
		if err != nil {
			msg.UserNotFound.Drop(w)
			tools.Log.Error(
				"user not found",
				"userID", userID,
			)
			return
		}
		h.conn, err = tools.WSUpgrader.Upgrade(w, r, nil)
		if err != nil {
			tools.Log.Info(
				"upgrade fail",
				"error", err.Error(),
			)
			return
		}
		tools.Log.Info(
			"new connection",
			"userID", userID,
		)
		h.storeToRoom(0)
		go h.writer()
		go h.reader()
	}
}

func (h *handler) writer() {
	defer h.close()
	ticker := time.NewTicker(h.tools.Cfg.WebSocket.PingPeriod)
	defer ticker.Stop()
	failedPings := 0
	for {
		var err error
		select {
		case <-h.closeCtx.Done():
			return
		default:
		}
		select {
		case msg, ok := <-h.sendChan:
			if !ok {
				return
			}
			h.conn.SetWriteDeadline(time.Now().Add(h.tools.Cfg.WebSocket.WriteWait))
			err = h.conn.WriteJSON(*msg)
		case <-ticker.C:
			h.conn.SetWriteDeadline(time.Now().Add(h.tools.Cfg.WebSocket.WriteWait))
			err = h.conn.WriteMessage(websocket.PingMessage, nil)
		case <-h.closeCtx.Done():
			return
		}
		if err != nil {
			failedPings++
			if failedPings > h.tools.Cfg.WebSocket.MaxFailedPings {
				h.cancel()
				return
			}
		} else {
			failedPings = 0
			continue
		}
	}
}

func (h *handler) setConnOptions() {
	h.conn.SetReadLimit(int64(h.tools.Cfg.WebSocket.MsgMaxSize))
	h.conn.SetReadDeadline(time.Now().Add(h.tools.Cfg.WebSocket.PongWait))
	h.conn.SetPongHandler(func(string) error {
		h.conn.SetReadDeadline(time.Now().Add(h.tools.Cfg.WebSocket.PongWait))
		return nil
	})
}

func (h *handler) reader() {
	defer h.cancel()
	h.setConnOptions()
	for {
		select {
		case <-h.closeCtx.Done():
			return
		default:
			var r models.WSRequest
			if err := h.conn.ReadJSON(&r); err != nil {
				if websocket.IsUnexpectedCloseError(err) {
					h.tools.Log.Info(
						"unexpected close",
						"userID", h.client.UserID,
						"error", err)
				}
				return
			}
			if err := h.route(&r); err != nil {
				h.tools.Log.Info(
					"request processing failed",
					"userID", h.client.UserID,
					"error", err.Error())
				if err := h.conn.WriteJSON(map[string]string{"error": err.Error()}); err != nil {
					return
				}
			}
		}

	}
}

func (h *handler) close() {
	h.leave()
	close(h.sendChan)
	h.conn.Close()
	h.tools.Log.Info(
		"connection closed",
		"userID", h.client.UserID,
	)
}

func (h *handler) sendMessage(message *models.Message) error {
	select {
	case <-h.closeCtx.Done():
		return nil
	default:
	}
	select {
	case h.sendChan <- message:
		return nil
	case <-time.After(h.tools.Cfg.WebSocket.WriteWait):
		h.tools.Log.Warn(
			"message send timeout",
			"userID", h.client.UserID,
			"roomID", h.client.RoomID)
		return errors.New("the message delivery time has expired")
	case <-h.closeCtx.Done():
		return nil
	}
}

func (h *handler) route(r *models.WSRequest) error {
	switch r.Type {
	case "message":
		msg := models.Message{
			Username: h.client.Username,
			Text:     r.Content,
		}
		return h.message(&msg)
	case "join":
		newRoomID, err := strconv.Atoi(r.Content)
		if err != nil {
			return errors.New("room id must be number")
		}
		return h.join(newRoomID)
	case "leave":
		return h.leave()
	default:
		return errors.New("invalid operation")
	}
}

func (h *handler) broadcastToRoom(msg *models.Message, roomID int) error {
	err := postgres.NewRepository(h.tools.DB).StoreMsg(&h.client, msg)
	if err != nil {
		return err
	}
	h.tools.Log.Debug(
		"broadcast",
		"roomID", roomID,
		"username", msg.Username,
		"text", msg.Text,
	)
	val, ok := onlineInRooms.Load(roomID)
	if !ok {
		return nil
	}
	inRoom := val.(*sync.Map)
	inRoom.Range(func(targetInterface, _ any) bool {
		h := targetInterface.(*handler)
		h.sendMessage(msg)
		return true
	})
	return nil
}

func (h *handler) message(msg *models.Message) error {
	if h.client.RoomID == 0 {
		return errors.New("not connected to any room")
	}
	if msg.Text == "" {
		return errors.New("can not send empty message")
	}
	if len(msg.Text) > h.tools.Cfg.WebSocket.MsgMaxLength {
		return fmt.Errorf("message should be less %d symbols long", h.tools.Cfg.WebSocket.MsgMaxLength)
	}
	h.broadcastToRoom(msg, h.client.RoomID)
	return nil
}

func (h *handler) join(newRoomID int) error {
	if h.client.RoomID == newRoomID {
		return errors.New("already in room")
	}
	h.leave()
	if err := postgres.NewRepository(h.tools.DB).IsRoomMember(h.client.UserID, newRoomID); err != nil {
		return err
	}
	h.storeToRoom(newRoomID)
	msg := &models.Message{
		Username: h.tools.Cfg.PKG.SystemUsername,
		Text:     h.client.Username + " joined the room",
	}
	h.broadcastToRoom(msg, newRoomID)
	return nil
}

func (h *handler) storeToRoom(roomID int) {
	h.client.RoomID = roomID
	newOnline, ok := onlineInRooms.Load(roomID)
	if !ok {
		m := new(sync.Map)
		m.Store(h, struct{}{})
		onlineInRooms.Store(roomID, m)
	} else {
		newOnline.(*sync.Map).Store(h, struct{}{})
	}
}

func (h *handler) leave() error {
	roomID := h.client.RoomID
	if h.client.RoomID == 0 {
		return errors.New("not connected to any room")
	}
	handlers, ok := onlineInRooms.Load(roomID)
	if ok {
		handlers.(*sync.Map).Delete(h)
	}
	h.storeToRoom(0)
	msg := &models.Message{
		Username: h.tools.Cfg.PKG.SystemUsername,
		Text:     h.client.Username + " leaved the room",
	}
	h.broadcastToRoom(msg, roomID)
	return nil
}
