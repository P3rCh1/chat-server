package rooms

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/P3rCh1/chat-server/internal/models"
	"github.com/P3rCh1/chat-server/internal/pkg/logger"
	"github.com/P3rCh1/chat-server/internal/pkg/responses"
	"github.com/P3rCh1/chat-server/internal/pkg/tools"
	"github.com/P3rCh1/chat-server/internal/pkg/validate"
	"github.com/go-chi/chi/v5"
	"github.com/lib/pq"
)

func Create(tools *tools.Tools) http.HandlerFunc {
	const op = "internal.http-server.handlers.rooms.room.Create"
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		userID := r.Context().Value("userID").(int)
		var roomReq models.CreateRoomRequest
		err := json.NewDecoder(r.Body).Decode(&roomReq)
		if err != nil {
			responses.InvalidData.Drop(w)
			return
		}
		if !validate.Name(roomReq.Name) {
			responses.BadName.Drop(w)
			return
		}
		room := models.Room{
			Name:      roomReq.Name,
			IsPrivate: roomReq.IsPrivate,
			CreatorID: userID,
		}
		err = tools.Repository.CreateRoom(&room)
		if err != nil {
			var pqErr *pq.Error
			if errors.As(err, &pqErr) {
				if pqErr.Code == "23505" {
					responses.RoomAlreadyExist.Drop(w)
				} else {
					responses.ServerError.Drop(w)
					logger.LogError(tools.Log, op, err)
				}
				return
			}
		}
		responses.SendJSON(w, http.StatusCreated, room)
	}
}

func Join(tools *tools.Tools) http.HandlerFunc {
	const op = "internal.http-server.handlers.rooms.room.Create"
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		userID := r.Context().Value("userID").(int)
		var room struct {
			ID int `json:"room_id"`
		}
		err := json.NewDecoder(r.Body).Decode(&room)
		if err != nil {
			responses.InvalidData.Drop(w)
			return
		}
		isPrivate, err := tools.Repository.IsPrivate(room.ID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				responses.RoomNotFound.Drop(w)
			} else {
				responses.ServerError.Drop(w)
				logger.LogError(tools.Log, op, err)
			}
		}
		if isPrivate {
			responses.RoomIsPrivate.Drop(w)
			return
		}
		if !addToRoom(userID, room.ID, tools, w, op) {
			return
		}
		responses.SendOk(w, "join succeed")
	}
}

func Invite(tools *tools.Tools) http.HandlerFunc {
	const op = "internal.http-server.handlers.rooms.room.Invite"
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		userID := r.Context().Value("userID").(int)
		var invite models.InviteRequest
		err := json.NewDecoder(r.Body).Decode(&invite)
		if err != nil {
			responses.InvalidData.Drop(w)
			return
		}
		creatorID, err := tools.Repository.CreatorID(invite.RoomID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				responses.RoomNotFound.Drop(w)
			} else {
				responses.ServerError.Drop(w)
				logger.LogError(tools.Log, op, err)
			}
			return
		}
		if creatorID != userID {
			responses.NoAccessToRoom.Drop(w)
			return
		}
		if !addToRoom(invite.UserID, invite.RoomID, tools, w, op) {
			return
		}
		responses.SendOk(w, "invite succeed")
	}
}

func addToRoom(userID, roomID int, tools *tools.Tools, w http.ResponseWriter, op string) bool {
	if err := tools.Repository.AddToRoom(userID, roomID); err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case "23505":
				responses.AlreadyInRoom.Drop(w)
				return false
			case "23503":
				if _, err := tools.Repository.GetUsername(userID); err != nil {
					responses.UserNotFound.Drop(w)
				} else {
					responses.RoomNotFound.Drop(w)
				}
				return false
			}
		}
		responses.ServerError.Drop(w)
		logger.LogError(tools.Log, op, err)
		return false
	}
	return true
}

func GetUserRooms(tools *tools.Tools) http.HandlerFunc {
	const op = "internal.http-server.handlers.rooms.room.GetUsersRooms"
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		userID := r.Context().Value("userID").(int)
		rooms, err := tools.Repository.GetUserRooms(userID)
		if err != nil {
			responses.ServerError.Drop(w)
			logger.LogError(tools.Log, op, err)
			return
		}
		responses.SendJSON(w, http.StatusOK, rooms)
	}
}

func GetRoom(tools *tools.Tools) http.HandlerFunc {
	const op = "internal.http-server.handlers.rooms.room.GetRoom"
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		roomID, err := strconv.Atoi(chi.URLParam(r, "roomID"))
		if err != nil {
			responses.InvalidURL.Drop(w)
			return
		}
		room, err := tools.Repository.GetRoom(roomID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				responses.RoomNotFound.Drop(w)
			} else {
				responses.ServerError.Drop(w)
				logger.LogError(tools.Log, op, err)
			}
			return
		}
		responses.SendJSON(w, http.StatusOK, room)
	}
}
