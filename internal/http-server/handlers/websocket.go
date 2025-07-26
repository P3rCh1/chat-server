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

var (
	roomHandlers = make(map[int]*map[*handler]struct{})
	mu           sync.RWMutex
)

type handler struct {
	client    models.Client
	conn      *websocket.Conn
	tools     *models.Tools
	sendChan  chan *models.Message
	closeCtx  context.Context
	closeDone chan struct{}
	cancel    context.CancelFunc
}

func newHandler(userID int, tools *models.Tools) (*handler, error) {
	var h handler
	h.sendChan = make(chan *models.Message, tools.Cfg.WebSocket.MsgBufSize)
	h.closeCtx, h.cancel = context.WithCancel(context.Background())
	h.closeDone = make(chan struct{})
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

func ShutdownWS(ctx context.Context, tools *models.Tools) error {
	var grace = true
	var wg sync.WaitGroup
	mu.RLock()
	for _, handlersInRoom := range roomHandlers {
		wg.Add(len(*handlersInRoom))
		for h := range *handlersInRoom {
			go func(h *handler) {
				defer wg.Done()
				h.cancel()
				select {
				case <-ctx.Done():
					grace = false
					return
				case <-h.closeDone:
					return
				}
			}(h)

		}
	}
	mu.RUnlock()
	wg.Wait()
	if !grace {
		return errors.New("shutdown timer end before close websocket")
	}
	return nil
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
		h.toRoom(0)
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
		case msg := <-h.sendChan:
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
					"error", err.Error(),
				)
				h.notify(&models.Message{
					Username: h.tools.Cfg.PKG.ErrorUsername,
					Text:     err.Error(),
				}, h.tools.Pkg.ErrorUserID)
			}
		}
	}
}

func (h *handler) close() {
	defer close(h.closeDone)
	roomID := h.client.RoomID
	mu.Lock()
	handlers, ok := roomHandlers[roomID]
	if ok {
		delete(*handlers, h)
	}
	mu.Unlock()
	if roomID != 0 {
		msg := &models.Message{
			Username: h.tools.Cfg.PKG.SystemUsername,
			Text:     h.client.Username + " leaved the room",
		}
		h.broadcastToRoom(msg, roomID, h.tools.Pkg.SystemUserID)
	}
	close(h.sendChan)
	h.conn.Close()
	h.tools.Log.Info(
		"connection closed",
		"userID", h.client.UserID,
	)
}

func (h *handler) sendMessage(msg *models.Message, fromID int) error {
	select {
	case <-h.closeCtx.Done():
		return nil
	default:
	}
	select {
	case h.sendChan <- msg:
		return nil
	case <-time.After(h.tools.Cfg.WebSocket.WriteWait):
		h.tools.Log.Warn(
			"message send timeout",
			"userID", fromID,
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

func (h *handler) broadcastToRoom(msg *models.Message, roomID int, userID int) error {
	err := postgres.NewRepository(h.tools.DB).StoreMsg(msg, roomID, userID)
	if err != nil {
		return err
	}
	h.tools.Log.Debug(
		"broadcast",
		"roomID", roomID,
		"username", msg.Username,
		"text", msg.Text,
	)
	mu.RLock()
	defer mu.RUnlock()
	handlersInRoom := roomHandlers[roomID]
	for onlineHandler := range *handlersInRoom {
		onlineHandler.sendMessage(msg, userID)
	}
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
	h.broadcastToRoom(msg, h.client.RoomID, h.client.UserID)
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
	h.toRoom(newRoomID)
	msg := &models.Message{
		Username: h.tools.Cfg.PKG.SystemUsername,
		Text:     h.client.Username + " joined the room",
	}
	h.broadcastToRoom(msg, newRoomID, h.tools.Pkg.SystemUserID)
	return nil
}

func (h *handler) toRoom(roomID int) {
	h.client.RoomID = roomID
	mu.Lock()
	defer mu.Unlock()
	roomOnline, ok := roomHandlers[roomID]
	if !ok {
		m := make(map[*handler]struct{})
		m[h] = struct{}{}
		roomHandlers[roomID] = &m
	} else {
		(*roomOnline)[h] = struct{}{}
	}
}

func (h *handler) leave() error {
	roomID := h.client.RoomID
	if h.client.RoomID == 0 {
		return errors.New("not connected to any room")
	}
	mu.Lock()
	handlers, ok := roomHandlers[roomID]
	if ok {
		delete(*handlers, h)
	}
	mu.Unlock()
	h.toRoom(0)
	msg := &models.Message{
		Username: h.tools.Cfg.PKG.SystemUsername,
		Text:     "you leaved the room",
	}
	h.notify(msg, h.tools.Pkg.SystemUserID)
	msg = &models.Message{
		Username: h.tools.Cfg.PKG.SystemUsername,
		Text:     h.client.Username + " leaved the room",
	}
	msg.Text = h.client.Username + " leaved the room"
	h.broadcastToRoom(msg, roomID, h.tools.Pkg.SystemUserID)
	return nil
}

func (h *handler) notify(msg *models.Message, fromID int) {
	msg.Timestamp = time.Now()
	h.sendMessage(msg, fromID)
}
