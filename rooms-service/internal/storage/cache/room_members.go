package cache

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisRoomMembers struct {
	client *redis.Client
	ttl    time.Duration
	key    string
}

func NewRoomMembersCacher(cache *redis.Client, ttl time.Duration) *RedisRoomMembers {
	return &RedisRoomMembers{
		client: cache,
		ttl:    ttl,
		key:    "room_members" + `:%d`,
	}
}

func (r *RedisRoomMembers) Add(ctx context.Context, uid int64, rooms ...int64) error {
	key := fmt.Sprintf(r.key, uid)
	_, err := r.client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		for _, id := range rooms {
			pipe.SAdd(ctx, key, id)
		}
		pipe.Expire(ctx, key, r.ttl)
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to add user`s id:%d rooms in cache: %w", uid, err)
	}
	return nil
}

func (r *RedisRoomMembers) IsMember(ctx context.Context, uid int64, roomID int64) (bool, error) {
	key := fmt.Sprintf(r.key, uid)
	pipe := r.client.Pipeline()
	existsCmd := pipe.Exists(ctx, key)
	isMemberCmd := pipe.SIsMember(ctx, fmt.Sprintf(r.key, uid), roomID)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to search user`s id:%d room in cache: %w", uid, err)
	}
	if existsCmd.Val() == 0 {
		return false, NotFound
	}
	return isMemberCmd.Val(), nil
}

func (r *RedisRoomMembers) Members(ctx context.Context, uid int64) ([]int64, error) {
	key := fmt.Sprintf(r.key, uid)
	pipe := r.client.Pipeline()
	existsCmd := pipe.Exists(ctx, key)
	roomsCmd := pipe.SMembers(ctx, key)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to search user`s id:%d rooms in cache: %w", uid, err)
	}
	if existsCmd.Val() == 0 {
		return nil, NotFound
	}
	val := roomsCmd.Val()
	rooms := make([]int64, len(val))
	for i, v := range val {
		room, _ := strconv.ParseInt(v, 10, 64)
		rooms[i] = room
	}
	return rooms, nil
}

func (r *RedisRoomMembers) AddSingle(ctx context.Context, uid, roomID int64) error {
	key := fmt.Sprintf(r.key, uid)
	err := r.client.Watch(ctx, func(tx *redis.Tx) error {
		exists, err := tx.Exists(ctx, key).Result()
		if err != nil {
			return err
		}
		if exists == 0 {
			return NotFound
		}
		_, err = tx.SAdd(ctx, key, roomID).Result()
		return err
	}, key)
	if err != nil {
		r.client.Del(ctx, key)
	}
	return err
}
