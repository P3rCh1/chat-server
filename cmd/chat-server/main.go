package main

import (
	"github.com/P3rCh1/chat-server/internal/app"
	"github.com/P3rCh1/chat-server/internal/config"
)

func main() {
	cfg := config.MustLoad()
	app.Run(cfg)
}

//TODO check RWMutex and Redis availability in ws
//TODO check hack safety in ws
//TODO add room creator and add handler
//TODO create default config options
