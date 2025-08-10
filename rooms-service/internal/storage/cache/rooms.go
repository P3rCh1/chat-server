package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/P3rCh1/chat-server/rooms-service/internal/models"
	"github.com/redis/go-redis/v9"
)

type RedisRooms struct {
	client *redis.Client
	ttl    time.Duration
	key    string
}

func NewRoomsCacher(cache *redis.Client, ttl time.Duration) *RedisRooms {
	return &RedisRooms{
		client: cache,
		ttl:    ttl,
		key:    "rooms" + `:%d`,
	}
}

func (c *RedisRooms) Close() error {
	return c.client.Close()
}

func (c *RedisRooms) Set(ctx context.Context, r *models.Room) error {
	key := fmt.Sprintf(c.key, r.RoomID)
	bytes, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}
	err = c.client.Set(ctx, key, bytes, c.ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set profile: %w", err)
	}
	return nil
}

func (c *RedisRooms) Get(ctx context.Context, id int) (*models.Room, error) {
	key := fmt.Sprintf(c.key, id)
	str, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, NotFound
		}
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}
	r := &models.Room{}
	if err := json.Unmarshal([]byte(str), r); err != nil {
		return nil, fmt.Errorf("failed to unmarshal profile: %w", err)
	}
	return r, nil
}
