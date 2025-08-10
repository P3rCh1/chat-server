package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/P3rCh1/chat-server/user-service/internal/config"
	"github.com/P3rCh1/chat-server/user-service/internal/models"
	"github.com/redis/go-redis/v9"
)

type Cacher struct {
	client  *redis.Client
	timeout time.Duration
	ttl     time.Duration
	key     string
}

func New(cfg *config.Redis) (*Cacher, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		DB:       cfg.DB,
		Password: cfg.Password,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		client.Close()
		return nil, err
	}
	return &Cacher{
		client: client,
		ttl:    cfg.TTL,
		key:    "profile" + ":%d",
	}, nil
}

func (c *Cacher) Close() error {
	return c.client.Close()
}

func (c *Cacher) Set(ctx context.Context, p *models.Profile) error {
	key := fmt.Sprintf(c.key, p.ID)
	bytes, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}
	err = c.client.Set(ctx, key, bytes, c.ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set profile: %w", err)
	}
	return nil
}

func (c *Cacher) Get(ctx context.Context, id int) (*models.Profile, error) {
	key := fmt.Sprintf(c.key, id)
	str, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}
	p := &models.Profile{}
	if err := json.Unmarshal([]byte(str), p); err != nil {
		return nil, fmt.Errorf("failed to unmarshal profile: %w", err)
	}
	return p, nil
}
