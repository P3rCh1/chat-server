package gateway

import (
	"log/slog"
	"os"
	"sync"

	"github.com/P3rCh1/chat-server/gateway-service/internal/config"
	"github.com/P3rCh1/chat-server/gateway-service/shared/logger"
	roomspb "github.com/P3rCh1/chat-server/gateway-service/shared/proto/gen/go/rooms"
	sessionpb "github.com/P3rCh1/chat-server/gateway-service/shared/proto/gen/go/session"
	userpb "github.com/P3rCh1/chat-server/gateway-service/shared/proto/gen/go/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Services struct {
	Session  sessionpb.SessionClient
	User     userpb.UserClient
	Rooms    roomspb.RoomsClient
	Log      *slog.Logger
	Timeouts config.TimeoutsServices
	conns    []*grpc.ClientConn
}

func MustNew(cfg *config.Config) *Services {
	s := &Services{
		Log:      logger.New(cfg.LogLVL),
		Timeouts: cfg.Services.Timeouts,
	}
	wg := sync.WaitGroup{}
	ok := true
	wg.Add(3)
	go func() {
		defer wg.Done()
		conn := s.AddConn(s.Log, cfg.Services.SessionAddr)
		s.Session = sessionpb.NewSessionClient(conn)
		ok = ok && conn != nil
	}()
	go func() {
		defer wg.Done()
		conn := s.AddConn(s.Log, cfg.Services.RoomsAddr)
		s.Rooms = roomspb.NewRoomsClient(conn)
		ok = ok && conn != nil
	}()
	go func() {
		defer wg.Done()
		conn := s.AddConn(s.Log, cfg.Services.UserAddr)
		s.User = userpb.NewUserClient(conn)
		ok = ok && conn != nil
	}()
	wg.Wait()
	if !ok {
		s.Close()
		os.Exit(1)
	}
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
