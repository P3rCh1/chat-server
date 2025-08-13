package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/P3rCh1/chat-server/message-service/internal/config"
	"github.com/P3rCh1/chat-server/message-service/internal/models"
	msgpb "github.com/P3rCh1/chat-server/message-service/shared/proto/gen/go/message"
	_ "github.com/lib/pq"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const Limit = 100

type Postgres struct {
	db *sql.DB
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

		CREATE TABLE IF NOT EXISTS messages (
    		id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id),
    		room_id INTEGER REFERENCES rooms(id),
			type VARCHAR(15) NOT NULL,
    		text TEXT,
			timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`
	_, err := db.ExecContext(ctx, query)
	return err
}

func (p *Postgres) Close() error {
	return p.db.Close()
}

func (p *Postgres) StoreMsg(msg *models.Message) error {
	const query = `
        INSERT INTO messages (
            room_id,
            user_id,
			type,
            text
        ) VALUES ($1, $2, $3, $4)
		RETURNING id, timestamp
    `
	row := p.db.QueryRow(query, msg.RoomID, msg.UID, msg.Type, msg.Text)
	err := row.Scan(&msg.ID, &msg.Timestamp)
	if err != nil {
		return fmt.Errorf("store msg fail: %w", err)
	}
	return nil
}

func (p *Postgres) GetMsgs(roomID, lastID int64) ([]*msgpb.Message, error) {
	var rows *sql.Rows
	var err error
	if lastID != 0 {
		const query = `
        SELECT id, user_id, type, text, timestamp
		FROM messages
		WHERE room_id = $1 AND id <= $2
		ORDER BY timestamp DESC LIMIT $3
    `
		rows, err = p.db.Query(query, roomID, lastID, Limit)
	} else {
		const query = `
        SELECT id, user_id, type, text, timestamp
		FROM messages
		WHERE room_id = $1
		ORDER BY timestamp DESC LIMIT $2
    `
		rows, err = p.db.Query(query, roomID, Limit)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	defer rows.Close()
	msgs := make([]*msgpb.Message, 0, Limit)
	for rows.Next() {
		msg := msgpb.Message{}
		var timestamp time.Time
		if err := rows.Scan(&msg.ID, &msg.UID, &msg.Type, &msg.Text, timestamp); err != nil {
			return nil, fmt.Errorf("failed to scan msg: %w", err)
		}
		msg.Timestamp = timestamppb.New(timestamp)
		msgs = append(msgs, &msg)
	}
	return msgs, nil
}
