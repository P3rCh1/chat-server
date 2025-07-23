package handlers

import (
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

const (
	maxMessageSize = 4096 // TODO move to config, change log msg
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageLen  = 1000
	maxFailedPings = 3
)

var upgrader = websocket.Upgrader{
	WriteBufferSize:   1024,
	ReadBufferSize:    1024,
	EnableCompression: true,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var onlineInRooms sync.Map

type handler struct {
	client    models.Client
	conn      *websocket.Conn
	tools     *models.Tools
	sendChan  chan *models.Message
	closeChan chan struct{}
}

func newHandler() *handler {
	var h handler
	h.sendChan = make(chan *models.Message, 100)
	h.closeChan = make(chan struct{})
	return &h
}

func Websocket(tools *models.Tools) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		userID, err := tools.TokenProvider.Verify(token)
		if err != nil {
			msg.UserNotFound.Drop(w)
			return
		}
		h := newHandler()
		h.tools = tools
		h.client = models.Client{
			UserID: userID,
			RoomID: 0,
		}
		h.client.Username, err = postgres.NewRepository(tools.DB).GetUsername(userID)
		if err != nil {
			msg.UserNotFound.Drop(w)
			return
		}
		h.conn, err = upgrader.Upgrade(w, r, nil)
		if err != nil {
			tools.Log.Info("upgrade fail", "error", err.Error())
			return
		}
		tools.Log.Info("new connection", "userID", userID)
		go h.writer()
		go h.reader()
	}
}

func (h *handler) writer() {
	defer h.conn.Close()
	ticker := time.NewTicker(pingPeriod)
	failedPings := 0
	for {
		select {
		case <-h.closeChan:
			return
		default:
		}
		select {
		case msg, ok := <-h.sendChan:
			if !ok {
				return
			}
			h.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := h.conn.WriteJSON(*msg); err != nil {
				return
			}
		case <-ticker.C:
			h.conn.SetWriteDeadline(time.Now().Add(writeWait))
			err := h.conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(writeWait))
			if err != nil {
				failedPings++
				if failedPings > maxFailedPings {
					return
				}
			} else {
				failedPings = 0
			}
		case <-h.closeChan:
			return
		}
	}
}

func (h *handler) reader() {
	defer func() {
		close(h.closeChan)
		close(h.sendChan)
		h.leave()
		h.conn.Close()
		h.tools.Log.Info("connection closed", "userID", h.client.UserID)
	}()
	h.conn.SetReadLimit(maxMessageSize)
	h.conn.SetReadDeadline(time.Now().Add(pongWait))
	h.conn.SetPongHandler(func(string) error {
		h.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		var r models.WSRequest
		if err := h.conn.ReadJSON(&r); err != nil {
			if websocket.IsUnexpectedCloseError(err) {
				h.tools.Log.Info("unexpected close",
					"userID", h.client.UserID,
					"error", err)
			}
			return
		}
		if err := h.route(&r); err != nil {
			h.tools.Log.Info("request processing failed",
				"userID", h.client.UserID,
				"error", err.Error())
			if err := h.conn.WriteJSON(map[string]string{"error": err.Error()}); err != nil {
				return
			}
		}
	}
}

func (h *handler) sendMessage(message *models.Message) error {
	select {
	case <-h.closeChan:
		return nil
	default:
	}
	select {
	case h.sendChan <- message:
		return nil
	case <-time.After(writeWait):
		h.tools.Log.Warn("message send timeout",
			"userID", h.client.UserID,
			"roomID", h.client.RoomID)
		return errors.New("the message delivery time has expired")
	case <-h.closeChan:
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

func (h *handler) broadcastToRoom(msg *models.Message) error {
	err := postgres.NewRepository(h.tools.DB).StoreMsg(&h.client, msg)
	if err != nil {
		return err
	}
	h.tools.Log.Debug("broadcast",
		"roomID", h.client.RoomID,
		"userID", h.client.UserID,
		"text", msg.Text,
	)
	val, ok := onlineInRooms.Load(h.client.RoomID)
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
	if len(msg.Text) > 1000 { //TODO move to config
		return fmt.Errorf("message should be less %d symbols long", 1000)
	}
	h.broadcastToRoom(msg)
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
	h.client.RoomID = newRoomID
	newOnline, ok := onlineInRooms.Load(newRoomID)
	if !ok {
		m := new(sync.Map)
		m.Store(h, struct{}{})
		onlineInRooms.Store(newRoomID, m)
	} else {
		newOnline.(*sync.Map).Store(h, struct{}{})
	}
	msg := &models.Message{
		Username: h.client.Username,
		Text:     "joined the room",
	}
	h.broadcastToRoom(msg)
	return nil
}

func (h *handler) leave() error {
	if h.client.RoomID == 0 {
		return errors.New("not connected to any room")
	}
	online, ok := onlineInRooms.Load(h.client.RoomID)
	if ok {
		online.(*sync.Map).Delete(h)
	}
	msg := &models.Message{
		Username: h.client.Username,
		Text:     "leaved the room",
	}
	h.broadcastToRoom(msg)
	h.client.RoomID = 0
	return nil
}
