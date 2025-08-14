package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/P3rCh1/chat-server/session/internal/config"
	grpcserver "github.com/P3rCh1/chat-server/session/internal/gRPCServer"
	"github.com/P3rCh1/chat-server/session/pkg/logger"
)

func main() {
	cfg := config.MustLoad()
	log := logger.New(cfg.LogLevel)
	s := grpcserver.Run(cfg, log)
	close := make(chan os.Signal, 1)
	signal.Notify(close, syscall.SIGINT, syscall.SIGTERM)
	<-close
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if grpcserver.ShutdownWithContext(ctx, s) {
		log.Info("server stopped gracefully")
	} else {
		log.Warn("forced shutdown due to timeout")
	}
}
