package message

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/P3rCh1/chat-server/gateway-service/internal/gateway"
	"github.com/P3rCh1/chat-server/gateway-service/internal/middleware"
	"github.com/P3rCh1/chat-server/gateway-service/internal/responses"
	msgpb "github.com/P3rCh1/chat-server/gateway-service/pkg/proto/gen/go/message"
	roomspb "github.com/P3rCh1/chat-server/gateway-service/pkg/proto/gen/go/rooms"
	"github.com/go-chi/chi/v5"
)

const URLParam = "roomID"

func Get(s *gateway.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		req := &msgpb.GetRequest{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			http.Error(w, "invalid data", http.StatusBadRequest)
			return
		}
		var err error
		req.RoomID, err = strconv.ParseInt(chi.URLParam(r, URLParam), 10, 64)
		if err != nil {
			http.Error(w, "invalid room id", http.StatusBadRequest)
			return
		}
		uid := r.Context().Value(middleware.UIDContextKey).(int64)
		ctx, cancel := context.WithTimeout(r.Context(), s.Timeouts.Message)
		defer cancel()
		isMember, err := s.Rooms.IsMember(ctx, &roomspb.IsMemberRequest{UID: uid, RoomID: req.RoomID})
		if err != nil {
			responses.GatewayGRPCErr(w, s.Log, "rooms", err)
			return
		}
		if !isMember.IsMember {
			http.Error(w, "not room member", http.StatusForbidden)
			return
		}
		msgs, err := s.Message.Get(ctx, req)
		if err != nil {
			responses.GatewayGRPCErr(w, s.Log, "messages", err)
			return
		}
		w.WriteHeader(http.StatusOK)
		writeMessages(w, msgs.Messages)
	}
}

func writeMessages(w io.Writer, messages []*msgpb.Message) error {
	_, err := w.Write([]byte{'['})
	if err != nil {
		return err
	}
	for i := 0; i < len(messages)-1; i++ {
		_, err := fmt.Fprintf(w, `{"ID":%d,"RoomID":%d,"UID":%d,"Type":%q,"Text":%q,"Timestamp":%q},`,
			messages[i].ID,
			messages[i].RoomID,
			messages[i].UID,
			messages[i].Type,
			messages[i].Text,
			messages[i].Timestamp.AsTime(),
		)
		if err != nil {
			return err
		}
	}
	lastIdx := len(messages) - 1
	if lastIdx >= 0 {
		_, err = fmt.Fprintf(w, `{"ID":%d,"RoomID":%d,"UID":%d,"Type":%q,"Text":%q,"Timestamp":%q}`,
			messages[len(messages)-1].ID,
			messages[len(messages)-1].RoomID,
			messages[len(messages)-1].UID,
			messages[len(messages)-1].Type,
			messages[len(messages)-1].Text,
			messages[len(messages)-1].Timestamp.AsTime(),
		)
		if err != nil {
			return err
		}
	}
	_, err = w.Write([]byte{']', '\n'})
	return err
}
