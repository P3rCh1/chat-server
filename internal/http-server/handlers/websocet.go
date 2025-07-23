package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/P3rCh1/chat-server/internal/models"
	"github.com/P3rCh1/chat-server/internal/pkg/msg"
	"github.com/P3rCh1/chat-server/internal/storage/postgres"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	WriteBufferSize:   1024,
	ReadBufferSize:    1024,
	EnableCompression: true,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var conns, onlineInRooms sync.Map

func Websocket(tools *models.Tools) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		userID, err := tools.TokenProvider.Verify(token)
		if err != nil {
			msg.UserNotFound.Drop(w)
			return
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			tools.Log.Info("upgrade fail", "error", err.Error())
			return
		}
		conns.Store(conn, 0)
		tools.Log.Info("new connection", "userID", userID)
		go handleConn(conn, userID, tools)
	}
}

func handleConn(conn *websocket.Conn, userID int, tools *models.Tools) {
	defer close(userID, conn, tools)
	conn.SetReadLimit(512) //TODO add to config file
	for {
		var r models.WSRequest
		if err := conn.ReadJSON(&r); err != nil {
			if websocket.IsUnexpectedCloseError(err) {
				tools.Log.Info("unexpected close", "userID", userID, "error", err)
			}
			return
		}
		if err := route(userID, conn, &r, tools); err != nil {
			tools.Log.Info("request processing failed", "userID", userID, "error", err.Error())
			if err := conn.WriteJSON(map[string]string{"error": err.Error()}); err != nil {
				return
			}
		}
	}
}

func close(userID int, conn *websocket.Conn, tools *models.Tools) {
	handleLeave(conn)
	conn.Close()
	tools.Log.Info("connection closed", "userID", userID)
}

func route(userID int, conn *websocket.Conn, r *models.WSRequest, tools *models.Tools) error {
	val, ok := conns.Load(conn)
	if !ok {
		return errors.New("connection interrupted")
	}
	roomID := val.(int)
	switch r.Type {
	case "message":
		return handleMsg(&models.Message{
			RoomID: roomID,
			UserID: userID,
			Text:   &r.Content,
		}, tools)
	case "join":
		newRoom, err := strconv.Atoi(r.Content)
		if err != nil {
			return errors.New("room id must be number")
		}
		return handleJoin(conn, userID, newRoom, tools)
	case "leave":
		return handleLeave(conn)
	default:
		return errors.New("invalid operation")
	}
}

func broadcastToRoom(msg *models.Message, tools *models.Tools) {
	val, ok := onlineInRooms.Load(msg.RoomID)
	if !ok {
		return
	}
	online := val.(*sync.Map)
	online.Range(func(targetInterface, _ any) bool {
		conn := targetInterface.(*websocket.Conn)
		if err := conn.WriteJSON(msg); err != nil {
			close(msg.UserID, conn, tools)
		}
		return true
	})
}

func handleMsg(msg *models.Message, tools *models.Tools) error {
	if msg.RoomID == 0 {
		return errors.New("user not connected to any room")
	}
	if *msg.Text == "" {
		return errors.New("can not send empty message")
	}
	if len(*msg.Text) > 1000 { //TODO move to config
		return fmt.Errorf("message should be less %d symbols long", 1000)
	}
	err := postgres.NewRepository(tools.DB).StoreMsg(msg)
	if err != nil {
		return err
	}
	broadcastToRoom(msg, tools)
	return nil
}

func handleJoin(conn *websocket.Conn, userID, newRoomID int, tools *models.Tools) error {
	handleLeave(conn)
	if err := postgres.NewRepository(tools.DB).IsRoomMember(userID, newRoomID); err != nil {
		return err
	}
	conns.Store(conn, newRoomID)
	newOnline, ok := onlineInRooms.Load(newRoomID)
	if !ok {
		m := new(sync.Map)
		m.Store(conn, struct{}{})
		onlineInRooms.Store(newRoomID, m)
	} else {
		newOnline.(*sync.Map).Store(conn, struct{}{})
	}
	return nil
}

func handleLeave(conn *websocket.Conn) error {
	roomID, ok := conns.Load(conn)
	if !ok {
		return errors.New("connection interrupted")
	}
	if roomID.(int) == 0 {
		return nil
	}
	online, ok := onlineInRooms.Load(roomID.(int))
	if ok {
		online.(*sync.Map).Delete(conn)
	}
	conns.Store(conn, 0)
	return nil
}
