package grpcserver

import (
	"context"
	"log/slog"
	"net"
	"os"

	"github.com/P3rCh1/chat-server/session/internal/config"
	"github.com/P3rCh1/chat-server/session/internal/session"
	"google.golang.org/grpc"
)

func Run(cfg *config.Config, log *slog.Logger) *grpc.Server {
	s := grpc.NewServer()
	Register(s, session.New(cfg, log))
	lis, err := net.Listen("tcp", cfg.Port)
	if err != nil {
		log.Error(
			"failed to listen tcp",
			"port", cfg.Port,
			"error", err,
		)
		os.Exit(1)
	}
	go func() {
		log.Info("session-service started", "port", cfg.Port)
		err := s.Serve(lis)
		if err != grpc.ErrServerStopped {
			log.Error("server crashed", "error", err)
			os.Exit(1)
		}
	}()
	return s
}

func ShutdownWithContext(ctx context.Context, s *grpc.Server) bool {
	done := make(chan struct{})
	go func() {
		s.GracefulStop()
		close(done)
	}()
	select {
	case <-done:
		return true
	case <-ctx.Done():
		s.Stop()
		return false
	}
}
