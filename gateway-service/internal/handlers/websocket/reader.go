package websocket

import (
	"context"

	"github.com/P3rCh1/chat-server/gateway-service/internal/models"
	msgpb "github.com/P3rCh1/chat-server/gateway-service/pkg/proto/gen/go/message"
	roomspb "github.com/P3rCh1/chat-server/gateway-service/pkg/proto/gen/go/rooms"
	"github.com/gorilla/websocket"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *connectionHandler) reader() {
	const op = "websocket.reader"
	defer h.close()
	defer h.cancel()
	for {
		select {
		case <-h.ctx.Done():
			return
		default:
		}
		r := new(models.WSRequest)
		if err := h.conn.ReadJSON(r); err != nil {
			if websocket.IsUnexpectedCloseError(err) {
				h.ws.services.Log.Error(
					op,
					"error", err,
					"uid", h.uid,
				)
			}
			return
		}
		h.route(r)
	}
}

func (h *connectionHandler) close() {
	defer close(h.closeDone)
	mu.Lock()
	h.delRoomMember()
	mu.Unlock()
	h.conn.Close()
}

func (h *connectionHandler) route(r *models.WSRequest) {
	switch r.Type {
	case "message":
		h.sendMessage(&msgpb.SendRequest{
			RoomID: h.roomID,
			UID:    h.uid,
			Type:   "message",
			Text:   r.Text,
		})
		return
	case "enter":
		h.enter(r.NewRoomID)
	default:
		h.SyncWriteJSON(models.NewWSError("invalid operation"))
	}
}

func (h *connectionHandler) internalErr(op string, err error) {
	h.SyncWriteJSON(models.NewWSError("internal error"))
	h.ws.services.Log.Error(
		op,
		"error", err,
	)
}

func (h *connectionHandler) sendMessage(msg *msgpb.SendRequest) {
	const op = "websocket.reader.sendMessage"
	if h.roomID == 0 {
		h.SyncWriteJSON(models.NewWSError("not in room"))
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), h.ws.services.Timeouts.Rooms)
	defer cancel()
	resp, err := h.ws.services.Message.Send(ctx, msg)
	if err != nil {
		if status, ok := status.FromError(err); ok && status.Code() != codes.Internal {
			h.SyncWriteJSON(models.NewWSError(status.Message()))
		}
		h.internalErr(op, err)
		return
	}
	h.SyncWriteJSON(models.NewSentResponse(resp.ID, resp.Timestamp.AsTime()))
}

func (h *connectionHandler) enter(newRoomID int64) {
	const op = "websocket.reader.enter"
	if err := h.validateEnter(newRoomID); err != nil {
		if err.Error == "internal error" {
			h.ws.services.Log.Error(
				op,
				"error", err,
			)
		}
		h.SyncWriteJSON(err)
		return
	}
	mu.Lock()
	h.delRoomMember()
	h.setRoomMember(newRoomID)
	mu.Unlock()
	h.SyncWriteJSON(models.NewEnterResponse())
}

func (h *connectionHandler) validateEnter(newRoomID int64) *models.WSError {
	if h.roomID == newRoomID {
		return models.NewWSError("already in room")
	}
	if newRoomID < 0 {
		return models.NewWSError("invalid room id")
	}
	ctx, cancel := context.WithTimeout(context.Background(), h.ws.services.Timeouts.Rooms)
	defer cancel()
	if newRoomID != 0 {
		isMember, err := h.ws.services.Rooms.IsMember(ctx, &roomspb.IsMemberRequest{
			UID:    h.uid,
			RoomID: newRoomID,
		})
		if err != nil {
			return models.NewWSError("internal error")
		}
		if !isMember.IsMember {
			return models.NewWSError("not room member")
		}
	}
	return nil
}

func (h *connectionHandler) setRoomMember(roomID int64) {
	handlers, ok := handlersInRoom[roomID]
	if !ok {
		m := make(map[*connectionHandler]struct{})
		m[h] = struct{}{}
		handlersInRoom[roomID] = m
	} else {
		handlers[h] = struct{}{}
	}
	h.roomID = roomID
}

func (h *connectionHandler) delRoomMember() {
	if h.roomID == 0 {
		return
	}
	handlers, ok := handlersInRoom[h.roomID]
	if ok {
		if len(handlers) == 1 {
			delete(handlersInRoom, h.roomID)
		} else {
			delete(handlers, h)
		}
	}
	h.roomID = 0
}
