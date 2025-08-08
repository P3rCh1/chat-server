package grpcserver

import (
	"context"

	sessionpb "github.com/P3rCh1/chat-server/session/shared/proto/gen/go/session"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type serverAPI struct {
	sessionpb.UnimplementedSessionServer
	session Session
}

type Session interface {
	Verify(ctx context.Context, token string) (int, error)
	Generate(ctx context.Context, userID int) (string, error)
}

func Register(gRPCServer *grpc.Server, session Session) {
	sessionpb.RegisterSessionServer(gRPCServer, &serverAPI{session: session})
}

func (s *serverAPI) Verify(ctx context.Context, r *sessionpb.VerifyRequest) (
	*sessionpb.VerifyResponse,
	error,
) {
	id, err := s.session.Verify(ctx, r.GetToken())
	if err != nil {
		return nil, status.Error(codes.Unknown, "invalid token")
	}
	return &sessionpb.VerifyResponse{Id: int32(id)}, nil
}

func (s *serverAPI) Generate(ctx context.Context, r *sessionpb.GenerateRequest) (
	*sessionpb.GenerateResponse,
	error,
) {
	token, err := s.session.Generate(ctx, int(r.Id))
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate token")
	}
	return &sessionpb.GenerateResponse{Token: token}, nil
}
