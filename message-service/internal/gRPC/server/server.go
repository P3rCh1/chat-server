package server //TODO think about join msg loss

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/P3rCh1/chat-server/message-service/internal/config"
	"github.com/P3rCh1/chat-server/message-service/internal/models"
	"github.com/P3rCh1/chat-server/message-service/internal/storage/database"
	"github.com/P3rCh1/chat-server/message-service/internal/storage/kafka"
	"github.com/P3rCh1/chat-server/message-service/shared/logger"
	msgpb "github.com/P3rCh1/chat-server/message-service/shared/proto/gen/go/message"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var ErrInternal = status.Error(codes.Internal, "internal error")

type ServerAPI struct {
	msgpb.UnimplementedMessageServiceServer
	log      *slog.Logger
	psql     *database.Postgres
	producer *kafka.Producer
}

func New(gRPCServer *grpc.Server, cfg *config.Config) (*ServerAPI, error) {
	s := &ServerAPI{
		log: logger.New(cfg.LogLevel),
	}
	var err error
	if s.psql, err = database.New(cfg.Postgres); err != nil {
		return nil, fmt.Errorf("postgres open fail %w", err)
	}
	s.producer = kafka.NewProducer(cfg.Kafka)
	msgpb.RegisterMessageServiceServer(gRPCServer, s)
	return s, err
}

func (s *ServerAPI) Close() {
	s.psql.Close()
}

func (s *ServerAPI) Send(ctx context.Context, r *msgpb.SendRequest) (*msgpb.SendResponse, error) {
	msg := &models.Message{
		UID:    r.UID,
		RoomID: r.RoomID,
		Text:   r.Text,
		Type:   r.Type,
	}
	if err := s.psql.StoreMsg(msg); err != nil {
		s.log.Error("send msg db error", "error", err)
		return nil, ErrInternal
	}
	if err := s.producer.Send(ctx, msg); err != nil {
		s.log.Error("send msg kafka error", "error", err)
		return nil, ErrInternal
	}
	return &msgpb.SendResponse{
		ID:        msg.ID,
		Timestamp: timestamppb.New(msg.Timestamp),
	}, nil
}

func (s *ServerAPI) Get(ctx context.Context, r *msgpb.GetRequest) (*msgpb.GetResponse, error) {
	msgs, err := s.psql.GetMsgs(r.RoomID, r.LastID)
	if err != nil {
		s.log.Error("get msgs db error", "error", err)
		return nil, ErrInternal
	}
	return &msgpb.GetResponse{Messages: msgs}, nil
}

func (s *ServerAPI) Ping(ctx context.Context, r *msgpb.Empty) (*msgpb.Empty, error) {
	return &msgpb.Empty{}, nil
}
