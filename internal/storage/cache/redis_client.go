package cache

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/P3rCh1/chat-server/internal/config"
	"github.com/redis/go-redis/v9"
)

func MustCreate(log *slog.Logger, cfg *config.Redis) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		log.Error(
			"failed to ping redis",
			"error",
			err,
		)
		os.Exit(1)
	}
	return client
}
