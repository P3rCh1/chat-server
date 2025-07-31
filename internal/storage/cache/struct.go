package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type StructCacher[T any] struct {
	Client *redis.Client
	TTL    time.Duration
	CTX    context.Context
	Key    string
}

func NewStructCacher[T any](cache *redis.Client, ttl time.Duration, tag string) *StructCacher[T] {
	return &StructCacher[T]{
		Client: cache,
		CTX:    context.Background(),
		TTL:    ttl,
		Key:    tag + `:%d`,
	}
}

func (sc *StructCacher[T]) Get(id int) (*T, error) {
	data, err := sc.Client.Get(sc.CTX, fmt.Sprintf(sc.Key, id)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	s := new(T)
	json.Unmarshal(data, s)
	return s, nil
}

func (sc *StructCacher[T]) Set(id int, s *T) error {
	data, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("failed to marshal struct: %w", err)
	}
	if err := sc.Client.Set(sc.CTX, fmt.Sprintf(sc.Key, id), data, sc.TTL).Err(); err != nil {
		return fmt.Errorf("failed to set struct in cache: %w", err)
	}
	return nil
}

func (sc *StructCacher[T]) Delete(id int) error {
	return sc.Client.Del(sc.CTX, fmt.Sprintf(sc.Key, id)).Err()
}
