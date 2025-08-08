package server

import (
	"context"

	"github.com/P3rCh1/chat-server/user-service/internal/gRPC/status_error"
	"github.com/P3rCh1/chat-server/user-service/internal/gRPC/validate"
	"github.com/P3rCh1/chat-server/user-service/internal/models"
	userpb "github.com/P3rCh1/chat-server/user-service/shared/proto/gen/go/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ServerAPI struct {
	userpb.UnimplementedUserServer
	user User
}

type User interface {
	Register(ctx context.Context, username, email, password string,
	) (int, error)
	Login(ctx context.Context, email, password string) (string, error)
	ChangeName(ctx context.Context, userID int, newName string) error
	Profile(ctx context.Context, userID int) (*models.Profile, error)
}

func Register(gRPCServer *grpc.Server, user User) {
	userpb.RegisterUserServer(gRPCServer, &ServerAPI{user: user})
}

func (s *ServerAPI) Register(ctx context.Context, r *userpb.RegisterRequest) (*userpb.RegisterResponse, error) {
	if err := validate.Register(r.Name, r.Email, r.Password); err != nil {
		return nil, err
	}
	if id, err := s.user.Register(ctx, r.Name, r.Email, r.Password); err != nil {
		if status_error.IsStatusError(err) {
			return nil, err
		}
		return nil, status.Errorf(codes.Internal, "unexpected error: %s", err)
	} else {
		return &userpb.RegisterResponse{Id: int32(id)}, nil
	}
}

func (s *ServerAPI) Login(ctx context.Context, r *userpb.LoginRequest) (*userpb.LoginResponse, error) {
	if !validate.Email(r.Email) {
		return nil, status.Error(codes.InvalidArgument, "invalid email")
	}
	if token, err := s.user.Login(ctx, r.Email, r.Password); err != nil {
		if status_error.IsStatusError(err) {
			return nil, err
		}
		return nil, status.Errorf(codes.Internal, "unexpected error: %s", err)
	} else {
		return &userpb.LoginResponse{Token: token}, nil
	}
}

func (s *ServerAPI) ChangeName(ctx context.Context, r *userpb.ChangeNameRequest) (*userpb.ChangeNameResponse, error) {
	if !validate.Name(r.NewName) {
		return nil, status.Error(codes.InvalidArgument, "invalid username")
	}
	if err := s.user.ChangeName(ctx, int(r.Id), r.NewName); err != nil {
		if status_error.IsStatusError(err) {
			return nil, err
		}
		return nil, status.Errorf(codes.Internal, "unexpected error: %s", err)
	}
	return &userpb.ChangeNameResponse{}, nil
}

func (s *ServerAPI) Profile(ctx context.Context, r *userpb.ProfileRequest) (*userpb.ProfileResponse, error) {
	if profile, err := s.user.Profile(ctx, int(r.Id)); err != nil {
		if status_error.IsStatusError(err) {
			return nil, err
		}
		return nil, status.Errorf(codes.Internal, "unexpected error: %s", err)
	} else {
		return &userpb.ProfileResponse{
			Id:        r.Id,
			Username:  profile.Username,
			Email:     profile.Email,
			CreatedAt: timestamppb.New(profile.CreatedAt),
		}, nil
	}
}
