package gateway

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"sync/atomic"

	"github.com/P3rCh1/chat-server/gateway-service/internal/config"
	"github.com/P3rCh1/chat-server/gateway-service/internal/kafka"
	"github.com/P3rCh1/chat-server/gateway-service/pkg/logger"
	msgpb "github.com/P3rCh1/chat-server/gateway-service/pkg/proto/gen/go/message"
	roomspb "github.com/P3rCh1/chat-server/gateway-service/pkg/proto/gen/go/rooms"
	sessionpb "github.com/P3rCh1/chat-server/gateway-service/pkg/proto/gen/go/session"
	userpb "github.com/P3rCh1/chat-server/gateway-service/pkg/proto/gen/go/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const ServicesCount = 4

type Services struct {
	Session  sessionpb.SessionClient
	User     userpb.UserClient
	Rooms    roomspb.RoomsClient
	Message  msgpb.MessageServiceClient
	Kafka    *kafka.Consumer
	Log      *slog.Logger
	Timeouts *config.TimeoutsServices
	conns    []*grpc.ClientConn
}

func MustNew(cfg *config.Config) *Services {
	s := &Services{
		Log:      logger.New(cfg.LogLVL),
		Timeouts: &cfg.Services.Timeouts,
	}
	wg := sync.WaitGroup{}
	var ok atomic.Bool
	ok.Store(true)
	s.conns = make([]*grpc.ClientConn, 0, ServicesCount)
	wg.Add(ServicesCount)
	go func() {
		defer wg.Done()
		conn := s.AddConn(s.Log, cfg.Services.SessionAddr)
		s.Session = sessionpb.NewSessionClient(conn)
		ok.CompareAndSwap(conn == nil, false)
	}()
	go func() {
		defer wg.Done()
		conn := s.AddConn(s.Log, cfg.Services.RoomsAddr)
		s.Rooms = roomspb.NewRoomsClient(conn)
		if conn == nil {
			ok.Store(false)
		}
	}()
	go func() {
		defer wg.Done()
		conn := s.AddConn(s.Log, cfg.Services.UserAddr)
		s.User = userpb.NewUserClient(conn)
		ok.CompareAndSwap(conn == nil, false)
		if conn == nil {
			ok.Store(false)
		}
	}()
	go func() {
		defer wg.Done()
		conn := s.AddConn(s.Log, cfg.Services.MessageAddr)
		s.Message = msgpb.NewMessageServiceClient(conn)
		ok.CompareAndSwap(conn == nil, false)
		if conn == nil {
			ok.Store(false)
		}
	}()
	s.Kafka = kafka.NewConsumer(cfg.Kafka)
	wg.Wait()
	if !ok.Load() {
		s.Close()
		os.Exit(1)
	}
	s.WarmUpAsync()
	return s
}

func (s *Services) Close() {
	for _, v := range s.conns {
		v.Close()
	}
}

func (s *Services) AddConn(log *slog.Logger, addr string) *grpc.ClientConn {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error(
			"failed to open",
			"error", err,
			"address", addr,
		)
		return nil
	}
	s.conns = append(s.conns, conn)
	return conn
}

func (s *Services) WarmUpAsync() {
	ctx := context.Background()
	go s.Message.Ping(ctx, &msgpb.Empty{})
	go s.Rooms.Ping(ctx, &roomspb.Empty{})
	go s.User.Ping(ctx, &userpb.Empty{})
	go s.Session.Ping(ctx, &sessionpb.Empty{})
}
