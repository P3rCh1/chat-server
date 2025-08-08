package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/P3rCh1/chat-server/shared/logger"
	"github.com/P3rCh1/chat-server/user-service/internal/config"
	"github.com/P3rCh1/chat-server/user-service/internal/gRPC/server"
)

func main() {
	cfg := config.MustLoad()
	log := logger.New(cfg.LogLevel)
	s := server.Run(cfg, log)
	close := make(chan os.Signal, 1)
	signal.Notify(close, syscall.SIGINT, syscall.SIGTERM)
	<-close
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if server.ShutdownWithContext(ctx, s) {
		log.Info("server stopped gracefully")
	} else {
		log.Warn("forced shutdown due to timeout")
	}
}
