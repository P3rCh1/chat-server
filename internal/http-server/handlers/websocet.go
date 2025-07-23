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
	maxMessageSize  = 4096 // 4KB
	writeWait       = 10 * time.Second
	pongWait        = 60 * time.Second
	pingPeriod      = (pongWait * 9) / 10
	maxMessageLen   = 1000
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
	client models.Client
	conn *websocket.Conn
	tools models.Tools
	sendChan chan any
	closeChan chan struct{}
}

func (h *handler) writer(tools *models.Tools) {
	ticker := time.NewTicker(pingPeriod)
	for {
		select {
		case msg := <-h.sendChan:
			h.conn.SetWriteDeadline(time.Now().Add(writeWait))

			if err := h.conn.WriteJSON(msg); err != nil {
				return
			}
		}
	}
}

func Websocket(tools *models.Tools) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		userID, err := tools.TokenProvider.Verify(token)
		if err != nil {
			msg.UserNotFound.Drop(w)
			return
		}
		client := &models.Client{
			UserID: userID,
			RoomID: 0,
		}
		client.Username, err = postgres.NewRepository(tools.DB).GetUsername(userID)
		if err != nil {
			msg.UserNotFound.Drop(w)
			return
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			tools.Log.Info("upgrade fail", "error", err.Error())
			return
		}
		tools.Log.Info("new connection", "userID", userID)
		go handleConn(conn, client, tools)
	}
}

func handleConn(conn *websocket.Conn, client *models.Client, tools *models.Tools) {
	defer close(conn, client, tools)
	conn.SetReadLimit(4096) //TODO add to config file
	for {
		var r models.WSRequest
		if err := conn.ReadJSON(&r); err != nil {
			if websocket.IsUnexpectedCloseError(err) {
				tools.Log.Info("unexpected close",
					"userID", client.UserID,
					"error", err)
			}
			return
		}
		if err := route(&r, conn, client, tools); err != nil {
			tools.Log.Info("request processing failed",
				"userID", client.UserID,
				"error", err.Error())
			if err := conn.WriteJSON(map[string]string{"error": err.Error()}); err != nil {
				return
			}
		}
	}
}

func close(conn *websocket.Conn, client *models.Client, tools *models.Tools) {
	handleLeave(conn, client, tools)
	conn.Close()
	tools.Log.Info("connection closed", "userID", client.UserID)
}

func route(r *models.WSRequest, conn *websocket.Conn, client *models.Client, tools *models.Tools) error {
	switch r.Type {
	case "message":
		msg := models.Message{
			Username: client.Username,
			Text:     r.Content,
		}
		return handleMsg(client, &msg, tools)
	case "join":
		newRoomID, err := strconv.Atoi(r.Content)
		if err != nil {
			return errors.New("room id must be number")
		}
		return handleJoin(conn, client, newRoomID, tools)
	case "leave":
		return handleLeave(conn, client, tools)
	default:
		return errors.New("invalid operation")
	}
}

func broadcastToRoom(client *models.Client, msg *models.Message, tools *models.Tools) error {
	err := postgres.NewRepository(tools.DB).StoreMsg(client, msg)
	if err != nil {
		return err
	}
	tools.Log.Info("broadcast",
		"roomID", client.RoomID,
		"userID", client.UserID,
		"text", msg.Text,
	)
	val, ok := onlineInRooms.Load(client.RoomID)
	if !ok {
		return nil
	}
	online := val.(*sync.Map)
	online.Range(func(targetInterface, _ any) bool {
		conn := targetInterface.(*websocket.Conn)
		if err := conn.WriteJSON(msg); err != nil {
			close(conn, client, tools)
		}
		return true
	})
	return nil
}

func handleMsg(client *models.Client, msg *models.Message, tools *models.Tools) error {
	if client.RoomID == 0 {
		return errors.New("not connected to any room")
	}
	if msg.Text == "" {
		return errors.New("can not send empty message")
	}
	if len(msg.Text) > 1000 { //TODO move to config
		return fmt.Errorf("message should be less %d symbols long", 1000)
	}
	return broadcastToRoom(client, msg, tools)
}

func handleJoin(conn *websocket.Conn, client *models.Client, newRoomID int, tools *models.Tools) error {
	if client.RoomID == newRoomID {
		return errors.New("already in room")
	}
	handleLeave(conn, client, tools)
	if err := postgres.NewRepository(tools.DB).IsRoomMember(client.UserID, newRoomID); err != nil {
		return err
	}
	client.RoomID = newRoomID
	newOnline, ok := onlineInRooms.Load(newRoomID)
	if !ok {
		m := new(sync.Map)
		m.Store(conn, struct{}{})
		onlineInRooms.Store(newRoomID, m)
	} else {
		newOnline.(*sync.Map).Store(conn, struct{}{})
	}
	msg := &models.Message{
		Username: client.Username,
		Text:     "joined the room",
	}
	return broadcastToRoom(client, msg, tools)
}

func handleLeave(conn *websocket.Conn, client *models.Client, tools *models.Tools) error {
	online, ok := onlineInRooms.Load(client.RoomID)
	if ok {
		online.(*sync.Map).Delete(conn)
	}
	msg := &models.Message{
		Username: client.Username,
		Text:     "leaved the room",
	}
	err := broadcastToRoom(client, msg, tools)
	client.RoomID = 0
	return err
}
