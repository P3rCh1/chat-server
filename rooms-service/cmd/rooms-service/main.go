package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/P3rCh1/chat-server/rooms-service/internal/config"
	"github.com/P3rCh1/chat-server/rooms-service/internal/gRPC/server"
	"github.com/P3rCh1/chat-server/rooms-service/internal/rooms"
	"github.com/P3rCh1/chat-server/rooms-service/pkg/logger"
)

func main() {
	cfg := config.MustLoad()
	log := logger.New(cfg.LogLevel)
	rooms := rooms.MustPrepare(log, cfg)
	defer rooms.Close()
	s := server.Run(cfg, log, rooms)
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
