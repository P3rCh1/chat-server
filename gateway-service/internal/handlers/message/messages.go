package message

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/P3rCh1/chat-server/gateway-service/internal/gateway"
	"github.com/P3rCh1/chat-server/gateway-service/internal/middleware"
	"github.com/P3rCh1/chat-server/gateway-service/internal/responses"
	msgpb "github.com/P3rCh1/chat-server/gateway-service/shared/proto/gen/go/message"
	roomspb "github.com/P3rCh1/chat-server/gateway-service/shared/proto/gen/go/rooms"
	"github.com/go-chi/chi/v5"
)

func Get(s *gateway.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		req := &msgpb.GetRequest{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			http.Error(w, "invalid data", http.StatusBadRequest)
			return
		}
		var err error
		req.RoomID, err = strconv.ParseInt(chi.URLParam(r, "roomID"), 10, 64)
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
		responses.SendJSON(w, http.StatusOK, msgs)
	}
}
