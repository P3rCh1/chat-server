package cache

import (
	"context"
	"errors"

	"github.com/P3rCh1/chat-server/rooms-service/internal/config"
	"github.com/redis/go-redis/v9"
)

var NotFound = errors.New("not found in cache")

func New(cfg *config.Redis) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		DB:       cfg.DB,
		Password: cfg.Password,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		client.Close()
		return nil, err
	}
	return client, nil
}
