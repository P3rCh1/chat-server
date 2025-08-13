package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/P3rCh1/chat-server/message-service/internal/config"
	"github.com/P3rCh1/chat-server/message-service/internal/gRPC/server"
	"github.com/P3rCh1/chat-server/message-service/shared/logger"
)

func main() {
	cfg := config.MustLoad()
	log := logger.New(cfg.LogLevel)
	api, s := server.Run(cfg, log)
	defer api.Close()
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
