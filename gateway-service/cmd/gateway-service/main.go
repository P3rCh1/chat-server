package main

import (
	"github.com/P3rCh1/chat-server/gateway-service/internal/app"
	"github.com/P3rCh1/chat-server/gateway-service/internal/config"
)

func main() {
	cfg := config.MustLoad()
	app.Run(cfg)
}
