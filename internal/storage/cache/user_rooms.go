package cache

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type UserRoomsCacher struct {
	Client *redis.Client
	TTL    time.Duration
	CTX    context.Context
	Key    string
}

func NewUserRoomsCacher(cache *redis.Client, ttl time.Duration, tag string) *UserRoomsCacher {
	return &UserRoomsCacher{
		Client: cache,
		CTX:    context.Background(),
		TTL:    ttl,
		Key:    tag + `:%d`,
	}
}

func (c *UserRoomsCacher) Add(userID int, rooms []int) error {
	key := fmt.Sprintf(c.Key, userID)
	_, err := c.Client.TxPipelined(c.CTX, func(pipe redis.Pipeliner) error {
		for _, id := range rooms {
			pipe.SAdd(c.CTX, key, id)
		}
		pipe.Expire(c.CTX, key, c.TTL)
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to add user`s id:%d rooms in cache: %w", userID, err)
	}
	return nil
}

func (c *UserRoomsCacher) IsMember(userID int, roomID int) (bool, error) {
	key := fmt.Sprintf(c.Key, userID)
	pipe := c.Client.Pipeline()
	existsCmd := pipe.Exists(c.CTX, key)
	isMemberCmd := pipe.SIsMember(c.CTX, fmt.Sprintf(c.Key, userID), roomID)
	_, err := pipe.Exec(c.CTX)
	if err != nil {
		return false, fmt.Errorf("failed to search user`s id:%d room in cache: %w", userID, err)
	}
	if existsCmd.Val() == 0 {
		return false, fmt.Errorf("not found in cache")
	}
	return isMemberCmd.Val(), nil
}

func (c *UserRoomsCacher) Members(userID int) ([]int, error) {
	key := fmt.Sprintf(c.Key, userID)
	pipe := c.Client.Pipeline()
	existsCmd := pipe.Exists(c.CTX, key)
	roomsCmd := pipe.SMembers(c.CTX, key)
	_, err := pipe.Exec(c.CTX)
	if err != nil {
		return nil, fmt.Errorf("failed to search user`s id:%d rooms in cache: %w", userID, err)
	}
	if existsCmd.Val() == 0 {
		return nil, nil
	}
	val := roomsCmd.Val()
	rooms := make([]int, len(val))
	for i, v := range val {
		room, _ := strconv.Atoi(v)
		rooms[i] = room
	}
	return rooms, nil
}
