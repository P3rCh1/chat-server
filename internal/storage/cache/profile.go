package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/P3rCh1/chat-server/internal/models"
	"github.com/redis/go-redis/v9"
)

type ProfileCacher struct {
	Client *redis.Client
	TTL    time.Duration
	CTX    context.Context
	Key    string
}

func NewProfileCacher(cache *redis.Client, ttl time.Duration, name string) *ProfileCacher {
	return &ProfileCacher{
		Client: cache,
		CTX:    context.Background(),
		TTL:    ttl,
		Key:    name + `:%d`,
	}
}

func (c *ProfileCacher) Get(userID int) (*models.Profile, error) {
	data, err := c.Client.Get(c.CTX, fmt.Sprintf(c.Key, userID)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var profile models.Profile
	json.Unmarshal(data, &profile)
	return &profile, nil
}

func (c *ProfileCacher) Set(profile *models.Profile) error {
	data, err := json.Marshal(profile)
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}
	if err := c.Client.Set(c.CTX, fmt.Sprintf(c.Key, profile.ID), data, c.TTL).Err(); err != nil {
		return fmt.Errorf("failed to set profile in cache: %w", err)
	}
	return nil
}

func (c *ProfileCacher) Delete(userID int) error {
	return c.Client.Del(c.CTX, fmt.Sprintf(c.Key, userID)).Err()
}
