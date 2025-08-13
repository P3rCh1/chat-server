package rooms

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/P3rCh1/chat-server/rooms-service/internal/config"
	"github.com/P3rCh1/chat-server/rooms-service/internal/gRPC/status_error"
	"github.com/P3rCh1/chat-server/rooms-service/internal/models"
	"github.com/P3rCh1/chat-server/rooms-service/internal/storage/repository"
)

type RoomsService struct {
	log  *slog.Logger
	repo *repository.Repository
}

func MustPrepare(log *slog.Logger, cfg *config.Config) *RoomsService {
	const op = "user.MustPrepare"
	room := &RoomsService{log: log}
	var err error
	room.repo, err = repository.New(log, cfg)
	if err != nil {
		log.Error(op, "error", err)
		room.Close()
		os.Exit(1)
	}
	return room
}

func (s *RoomsService) Close() {
	s.repo.Close()
}

func (s *RoomsService) Create(
	ctx context.Context,
	room *models.Room,
) (int64, error) {
	const op = "user.Create"
	err := s.repo.CreateRoom(ctx, room)
	if err != nil {
		if status_error.IsStatusError(err) {
			return 0, err
		}
		s.log.Error(op, "error", err)
		return 0, fmt.Errorf("create room error: %w", err)
	}
	return room.RoomID, nil
}

func (s *RoomsService) Invite(
	ctx context.Context,
	requesterUID, invitedUID, roomID int64,
) error {
	const op = "user.Invite"
	creatorID, err := s.repo.CreatorID(ctx, roomID)
	if err != nil {
		if status_error.IsStatusError(err) {
			return err
		}
		s.log.Error(op, "error", err)
		return fmt.Errorf("get creator error: %w", err)
	}
	if creatorID != requesterUID {
		return status_error.NoAccess
	}
	err = s.repo.AddToRoom(ctx, invitedUID, roomID)
	if err != nil {
		if status_error.IsStatusError(err) {
			return err
		}
		return fmt.Errorf("add to room error: %w", err)
	}
	return nil
}

func (s *RoomsService) Join(
	ctx context.Context,
	UID, roomID int64,
) error {
	const op = "user.Join"
	isPrivate, err := s.repo.IsPrivate(ctx, roomID)
	if err != nil {
		if status_error.IsStatusError(err) {
			return err
		}
		s.log.Error(op, "error", err)
		return fmt.Errorf("get creator error: %w", err)
	}
	if isPrivate {
		return status_error.Private
	}
	err = s.repo.AddToRoom(ctx, UID, roomID)
	if err != nil {
		if status_error.IsStatusError(err) {
			return err
		}
		return fmt.Errorf("add to room error: %w", err)
	}
	return nil
}

func (s *RoomsService) Get(
	ctx context.Context,
	roomID int64,
) (*models.Room, error) {
	const op = "user.Get"
	room, err := s.repo.GetRoom(ctx, roomID)
	if err != nil {
		if status_error.IsStatusError(err) {
			return nil, err
		}
		s.log.Error(op, "error", err)
		return nil, fmt.Errorf("get creator error: %w", err)
	}
	return room, nil
}

func (s *RoomsService) UserIn(
	ctx context.Context,
	uid int64,
) ([]int64, error) {
	const op = "user.UserIn"
	rooms, err := s.repo.GetUserRooms(ctx, uid)
	if err != nil {
		if status_error.IsStatusError(err) {
			return nil, err
		}
		s.log.Error(op, "error", err)
		return nil, fmt.Errorf("get creator error: %w", err)
	}
	resp := make([]int64, len(rooms))
	for i, v := range rooms {
		resp[i] = v
	}
	return resp, nil
}

func (s *RoomsService) IsMember(
	ctx context.Context,
	uid, roomID int64,
) (bool, error) {
	const op = "user.UserIn"
	isMember, err := s.repo.IsMember(ctx, uid, roomID)
	if err != nil {
		if status_error.IsStatusError(err) {
			return false, err
		}
		s.log.Error(op, "error", err)
		return false, fmt.Errorf("check IsMember error: %w", err)
	}
	return isMember, nil
}

func (s *RoomsService) Ping(ctx context.Context) {
	return
}
