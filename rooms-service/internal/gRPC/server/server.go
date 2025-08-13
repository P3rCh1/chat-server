package server

import (
	"context"

	"github.com/P3rCh1/chat-server/rooms-service/internal/gRPC/status_error"
	"github.com/P3rCh1/chat-server/rooms-service/internal/models"
	roomspb "github.com/P3rCh1/chat-server/rooms-service/shared/proto/gen/go/rooms"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ServerAPI struct {
	roomspb.UnimplementedRoomsServer
	rooms Rooms
}

type Rooms interface {
	Create(ctx context.Context, room *models.Room) (int64, error)
	Invite(ctx context.Context, requesterUID, invitedUID, roomID int64) error
	Join(ctx context.Context, UID, roomID int64) error
	Get(ctx context.Context, roomID int64) (*models.Room, error)
	UserIn(ctx context.Context, UID int64) ([]int64, error)
	IsMember(ctx context.Context, UID, roomID int64) (bool, error)
	Ping(ctx context.Context)
}

func Register(gRPCServer *grpc.Server, rooms Rooms) {
	roomspb.RegisterRoomsServer(gRPCServer, &ServerAPI{rooms: rooms})
}

func (s *ServerAPI) Create(ctx context.Context, r *roomspb.CreateRequest) (*roomspb.CreateResponse, error) {
	if r.Name == "" {
		return nil, status_error.EmptyName
	}
	if len(r.Name) < 3 || len(r.Name) > 20 {
		return nil, status_error.InvalidName
	}
	if roomID, err := s.rooms.Create(ctx, &models.Room{
		CreatorUID: r.UID,
		Name:       r.Name,
		IsPrivate:  r.IsPrivate},
	); err != nil {
		if status_error.IsStatusError(err) {
			return nil, err
		}
		return nil, status.Errorf(codes.Internal, "unexpected error: %s", err)
	} else {
		return &roomspb.CreateResponse{RoomID: roomID}, nil
	}
}

func (s *ServerAPI) Join(ctx context.Context, r *roomspb.JoinRequest) (*roomspb.JoinResponse, error) {
	if err := s.rooms.Join(ctx, r.UID, r.RoomID); err != nil {
		if status_error.IsStatusError(err) {
			return nil, err
		}
		return nil, status.Errorf(codes.Internal, "unexpected error: %s", err)
	} else {
		return &roomspb.JoinResponse{}, nil
	}
}

func (s *ServerAPI) Invite(ctx context.Context, r *roomspb.InviteRequest) (*roomspb.InviteResponse, error) {
	if err := s.rooms.Invite(ctx, r.CreatorUID, r.UID, r.RoomID); err != nil {
		if status_error.IsStatusError(err) {
			return nil, err
		}
		return nil, status.Errorf(codes.Internal, "unexpected error: %s", err)
	} else {
		return &roomspb.InviteResponse{}, nil
	}
}

func (s *ServerAPI) Get(ctx context.Context, r *roomspb.GetRequest) (*roomspb.GetResponse, error) {
	if room, err := s.rooms.Get(ctx, r.RoomID); err != nil {
		if status_error.IsStatusError(err) {
			return nil, err
		}
		return nil, status.Errorf(codes.Internal, "unexpected error: %s", err)
	} else {
		return &roomspb.GetResponse{
			RoomID:     room.RoomID,
			Name:       room.Name,
			CreatorUID: room.CreatorUID,
			IsPrivate:  room.IsPrivate,
			CreatedAt:  timestamppb.New(room.CreatedAt),
		}, nil
	}
}

func (s *ServerAPI) UserIn(ctx context.Context, r *roomspb.UserInRequest) (*roomspb.UserInResponse, error) {
	if IDs, err := s.rooms.UserIn(ctx, r.UID); err != nil {
		if status_error.IsStatusError(err) {
			return nil, err
		}
		return nil, status.Errorf(codes.Internal, "unexpected error: %s", err)
	} else {
		return &roomspb.UserInResponse{IDs: IDs}, nil
	}
}

func (s *ServerAPI) IsMember(ctx context.Context, r *roomspb.IsMemberRequest) (*roomspb.IsMemberResponse, error) {
	if isMember, err := s.rooms.IsMember(ctx, r.UID, r.RoomID); err != nil {
		if status_error.IsStatusError(err) {
			return nil, err
		}
		return nil, status.Errorf(codes.Internal, "unexpected error: %s", err)
	} else {
		return &roomspb.IsMemberResponse{IsMember: isMember}, nil
	}
}

func (s *ServerAPI) Ping(ctx context.Context, r *roomspb.Empty) (*roomspb.Empty, error) {
	s.rooms.Ping(ctx)
	return &roomspb.Empty{}, nil
}
