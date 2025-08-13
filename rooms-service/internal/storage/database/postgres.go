package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/P3rCh1/chat-server/rooms-service/internal/config"
	"github.com/P3rCh1/chat-server/rooms-service/internal/gRPC/status_error"
	"github.com/P3rCh1/chat-server/rooms-service/internal/models"
	_ "github.com/lib/pq"
)

type Postgres struct {
	db *sql.DB
}

func migrate(ctx context.Context, db *sql.DB) error {
	const query = `
		CREATE TABLE IF NOT EXISTS rooms (
    		id SERIAL PRIMARY KEY,
    		name VARCHAR(100) NOT NULL UNIQUE,
    		is_private BOOLEAN DEFAULT false,
    		creator_id INTEGER REFERENCES users(id) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS room_members (
    		user_id INTEGER REFERENCES users(id),
    		room_id INTEGER REFERENCES rooms(id),
    		PRIMARY KEY (user_id, room_id)
		);
	`
	_, err := db.ExecContext(ctx, query)
	return err
}

func New(cfg *config.Postgres) (*Postgres, error) {
	info := fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.DB, cfg.User, cfg.Password,
	)
	db, err := sql.Open("postgres", info)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgresql: %w", err)
	}
	if err := migrate(context.Background(), db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to migrate postgres: %w", err)
	}
	return &Postgres{db}, nil
}

func (p *Postgres) Close() error {
	return p.db.Close()
}

func (r *Postgres) CreateRoom(ctx context.Context, room *models.Room) error {
	const query = `
		INSERT INTO rooms (name, is_private, creator_id)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  false,
	})
	if err != nil {
		return fmt.Errorf("failed start transaction: %w", err)
	}
	defer tx.Rollback()
	err = tx.QueryRowContext(
		ctx,
		query,
		room.Name,
		room.IsPrivate,
		room.CreatorUID,
	).Scan(&room.RoomID, &room.CreatedAt)
	if err != nil {
		statErr := ExpectedPGErr(err, status_error.UserNotFound, status_error.NameExists)
		if statErr != nil {
			return statErr
		}
		return fmt.Errorf("failed to create room: %w", err)
	}
	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO room_members (user_id, room_id) VALUES ($1, $2)",
		room.CreatorUID, room.RoomID,
	)
	if err != nil {
		return fmt.Errorf("failed to add creator to room: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}
	return nil
}

func (r *Postgres) CreatorID(ctx context.Context, roomID int64) (int64, error) {
	const query = `
		SELECT creator_id FROM rooms WHERE id = $1
	`
	var creatorID int64
	err := r.db.QueryRowContext(ctx, query, roomID).Scan(&creatorID)
	if err != nil {
		statErr := ExpectedPGErr(err, status_error.UserNotFound, nil)
		if statErr != nil {
			return 0, statErr
		}
		return 0, fmt.Errorf("failed to find creator id: %w", err)
	}
	return creatorID, nil
}

func (r *Postgres) AddToRoom(ctx context.Context, uid, roomID int64) error {
	const query = `
		INSERT INTO room_members (user_id, room_id)
		VALUES ($1, $2)
	`
	_, err := r.db.ExecContext(ctx, query, uid, roomID)
	if err != nil {
		statErr := ExpectedPGErr(err, status_error.UserNotFound, status_error.AlreadyInRoom)
		if statErr != nil {
			return statErr
		}
		return fmt.Errorf("failed add to room: %w", err)
	}
	return nil
}

func (r *Postgres) IsPrivate(ctx context.Context, roomID int64) (bool, error) {
	const query = `
		SELECT is_private FROM rooms WHERE id = $1
	`
	var isPrivate bool
	err := r.db.QueryRowContext(ctx, query, roomID).Scan(&isPrivate)
	if err != nil {
		statErr := ExpectedPGErr(err, status_error.RoomNotFound, nil)
		if statErr != nil {
			return false, statErr
		}
		return false, fmt.Errorf("failed to check room's private status: %w", err)
	}
	return isPrivate, nil
}

func (r *Postgres) GetUserRooms(ctx context.Context, uid int64) ([]int64, error) {
	const query = `
        SELECT room_id FROM room_members WHERE user_id = $1
    `
	rows, err := r.db.Query(query, uid)
	if err != nil {
		statErr := ExpectedPGErr(err, status_error.UserNotFound, nil)
		if statErr != nil {
			return nil, statErr
		}
		return nil, fmt.Errorf("failed to get user`s rooms: %w", err)
	}
	defer rows.Close()
	var rooms []int64
	for rows.Next() {
		var roomID int64
		if err := rows.Scan(&roomID); err != nil {
			return nil, fmt.Errorf("failed to scan room: %w", err)
		}
		rooms = append(rooms, roomID)
	}
	return rooms, nil
}

func (r *Postgres) GetRoom(ctx context.Context, roomID int64) (*models.Room, error) {
	const query = `
        SELECT name, is_private, creator_id, created_at FROM rooms WHERE id = $1
    `
	room := &models.Room{
		RoomID: roomID,
	}
	err := r.db.QueryRowContext(ctx, query, roomID).Scan(
		&room.Name,
		&room.IsPrivate,
		&room.CreatorUID,
		&room.CreatedAt,
	)
	if err != nil {
		statErr := ExpectedPGErr(err, status_error.RoomNotFound, nil)
		if statErr != nil {
			return nil, statErr
		}
		return nil, fmt.Errorf("failed to get room: %w", err)
	}
	return room, nil
}

func (r *Postgres) IsMember(ctx context.Context, uid, roomID int64) (bool, error) {
	const query = `
        SELECT EXISTS(SELECT 1 FROM room_members WHERE user_id = $1 AND room_id = $2) 
    `
	var exists bool
	err := r.db.QueryRowContext(ctx, query, uid, roomID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check IsMember: %w", err)
	}
	return exists, nil
}
