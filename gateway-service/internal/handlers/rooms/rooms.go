package rooms

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/P3rCh1/chat-server/gateway-service/internal/gateway"
	"github.com/P3rCh1/chat-server/gateway-service/internal/middleware"
	"github.com/P3rCh1/chat-server/gateway-service/internal/responses"
	roomspb "github.com/P3rCh1/chat-server/gateway-service/shared/proto/gen/go/rooms"
	"github.com/go-chi/chi/v5"
)

var URLParam = "roomID"

func Create(s *gateway.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		req := roomspb.CreateRequest{}
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "invalid data", http.StatusBadRequest)
			return
		}
		req.UID = r.Context().Value(middleware.UIDContextKey).(int64)
		ctx, cancel := context.WithTimeout(r.Context(), s.Timeouts.Rooms)
		defer cancel()
		resp, err := s.Rooms.Create(ctx, &req)
		if err != nil {
			responses.GatewayGRPCErr(w, s.Log, "rooms", err)
			return
		}
		responses.SendJSON(w, http.StatusCreated, resp)
	}
}

func Join(s *gateway.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		req := roomspb.JoinRequest{}
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "invalid data", http.StatusBadRequest)
			return
		}
		req.UID = r.Context().Value(middleware.UIDContextKey).(int64)
		ctx, cancel := context.WithTimeout(context.Background(), s.Timeouts.Rooms)
		defer cancel()
		_, err = s.Rooms.Join(ctx, &req)
		if err != nil {
			responses.GatewayGRPCErr(w, s.Log, "rooms", err)
			return
		}
		responses.SendJSON(w, http.StatusOK, map[string]string{"status": "joined"})
	}
}

func Invite(s *gateway.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		req := roomspb.InviteRequest{}
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "invalid data", http.StatusBadRequest)
			return
		}
		req.CreatorUID = r.Context().Value(middleware.UIDContextKey).(int64)
		ctx, cancel := context.WithTimeout(context.Background(), s.Timeouts.Rooms)
		defer cancel()
		_, err = s.Rooms.Invite(ctx, &req)
		if err != nil {
			responses.GatewayGRPCErr(w, s.Log, "rooms", err)
			return
		}
		responses.SendJSON(w, http.StatusOK, map[string]string{"status": "invited"})
	}
}

func Get(s *gateway.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		roomID, err := strconv.ParseInt(chi.URLParam(r, URLParam), 10, 64)
		if err != nil {
			http.Error(w, "invalid roomID", http.StatusBadRequest)
			return
		}
		req := roomspb.GetRequest{RoomID: int64(roomID)}
		ctx, cancel := context.WithTimeout(context.Background(), s.Timeouts.Rooms)
		defer cancel()
		respGRPC, err := s.Rooms.Get(ctx, &req)
		if err != nil {
			responses.GatewayGRPCErr(w, s.Log, "rooms", err)
			return
		}
		resp := struct {
			RoomID     int64     `json:"roomID"`
			Name       string    `json:"name"`
			CreatorUID int64     `json:"creatorUID"`
			IsPrivate  bool      `json:"isPrivate"`
			CreatedAt  time.Time `json:"createdAt"`
		}{
			RoomID:     respGRPC.RoomID,
			Name:       respGRPC.Name,
			CreatorUID: respGRPC.CreatorUID,
			IsPrivate:  respGRPC.IsPrivate,
			CreatedAt:  respGRPC.CreatedAt.AsTime(),
		}
		responses.SendJSON(w, http.StatusOK, resp)
	}
}

func UserIn(s *gateway.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		req := roomspb.UserInRequest{UID: r.Context().Value(middleware.UIDContextKey).(int64)}
		ctx, cancel := context.WithTimeout(context.Background(), s.Timeouts.Rooms)
		defer cancel()
		resp, err := s.Rooms.UserIn(ctx, &req)
		if err != nil {
			responses.GatewayGRPCErr(w, s.Log, "rooms", err)
			return
		}
		responses.SendJSON(w, http.StatusOK, resp)
	}
}
