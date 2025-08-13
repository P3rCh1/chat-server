package repository

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/P3rCh1/chat-server/rooms-service/internal/config"
	"github.com/P3rCh1/chat-server/rooms-service/internal/models"
	"github.com/P3rCh1/chat-server/rooms-service/internal/storage/cache"
	"github.com/P3rCh1/chat-server/rooms-service/internal/storage/database"
	"github.com/redis/go-redis/v9"
)

type Repository struct {
	log         *slog.Logger
	psql        *database.Postgres
	rooms       *cache.RedisRooms
	roomMembers *cache.RedisRoomMembers
}

func (r *Repository) Close() {
	if r.psql != nil {
		r.psql.Close()
	}
	if r.rooms != nil {
		r.rooms.Close()
	}
}

func New(log *slog.Logger, cfg *config.Config) (*Repository, error) {
	var errPQ, errRedis error
	repo := &Repository{log: log}
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		repo.psql, errPQ = database.New(cfg.Postgres)
	}()
	go func() {
		defer wg.Done()
		var redis *redis.Client
		redis, errRedis = cache.New(cfg.Redis)
		repo.roomMembers = cache.NewRoomMembersCacher(redis, cfg.Redis.TTL)
		repo.rooms = cache.NewRoomsCacher(redis, cfg.Redis.TTL)
	}()
	wg.Wait()
	if err := errors.Join(errPQ, errRedis); err != nil {
		repo.Close()
		return nil, err
	}
	return repo, nil
}

func (r *Repository) CreateRoom(ctx context.Context, room *models.Room) error {
	err := r.psql.CreateRoom(ctx, room)
	if err != nil {
		return err
	}
	go func() {
		r.rooms.Set(context.Background(), room)
	}()
	return nil
}

func (r *Repository) CreatorID(ctx context.Context, roomID int64) (int64, error) {
	room, err := r.GetRoom(ctx, roomID)
	if err != nil {
		return 0, err
	}
	return room.CreatorUID, nil
}

func (r *Repository) AddToRoom(ctx context.Context, uid, roomID int64) error {
	err := r.psql.AddToRoom(ctx, uid, roomID)
	if err != nil {
		return err
	}
	if err := r.roomMembers.AddSingle(ctx, uid, roomID); err != nil {
		if err == cache.NotFound {
			r.log.Debug("not found user`s rooms in cache", "UID", uid)
		}
		r.log.Error("add to room redis fail", "error", err)
	}
	return nil
}

func (r *Repository) IsPrivate(ctx context.Context, roomID int64) (bool, error) {
	room, err := r.GetRoom(ctx, roomID)
	if err != nil {
		return false, err
	}
	return room.IsPrivate, nil
}

func (r *Repository) GetUserRooms(ctx context.Context, uid int64) ([]int64, error) {
	rooms, err := r.roomMembers.Members(ctx, uid)
	if err == nil {
		return rooms, nil
	}
	if err == cache.NotFound {
		r.log.Debug("not found user`s rooms in cache", "UID", uid)
	} else {
		r.log.Error("get user`s rooms redis fail", "error", err)
	}
	rooms, err = r.psql.GetUserRooms(ctx, uid)
	if err != nil {
		return nil, err
	}
	go func() {
		err := r.roomMembers.Add(ctx, uid, rooms...)
		if err != nil {
			r.log.Error("set user`s rooms redis fail", "error", err)
		}
	}()
	return rooms, nil
}

func (r *Repository) GetRoom(ctx context.Context, roomID int64) (*models.Room, error) {
	room, err := r.rooms.Get(ctx, roomID)
	if err == nil {
		return room, nil
	}
	if err == cache.NotFound {
		r.log.Debug("not found room in cache", "roomID", roomID)
	} else {
		r.log.Error("get room redis fail", "error", err)
	}
	room, err = r.psql.GetRoom(ctx, roomID)
	if err != nil {
		return nil, err
	}
	go func() {
		err := r.rooms.Set(ctx, room)
		if err != nil {
			r.log.Error("set room redis fail", "error", err)
		}
	}()
	return room, nil
}

func (r *Repository) IsMember(ctx context.Context, uid, roomID int64) (bool, error) {
	isMember, err := r.roomMembers.IsMember(ctx, uid, roomID)
	if err == nil {
		return isMember, nil
	}
	if err == cache.NotFound {
		r.log.Debug("not found room in cache", "roomID", roomID)
	} else {
		r.log.Error("check room membership redis fail", "error", err)
	}
	return r.psql.IsMember(ctx, uid, roomID)
}
