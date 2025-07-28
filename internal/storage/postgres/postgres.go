package postgres

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	"github.com/P3rCh1/chat-server/internal/config"
	_ "github.com/lib/pq"
)

func MustOpen(log *slog.Logger, cfg *config.DB) *sql.DB {
	var err error
	defer func() {
		if err != nil {
			log.Error(
				"storage.postgres.New",
				"error", err.Error(),
			)
			os.Exit(1)
		}
	}()
	var db *sql.DB
	psqlInfo := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name,
	)
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil
	}
	err = db.Ping()
	if err != nil {
		return nil
	}
	return db
}

func MustApplyMigrations(log *slog.Logger, db *sql.DB) {
	query := `
		CREATE TABLE IF NOT EXISTS users (
	    	id SERIAL PRIMARY KEY,
    		username VARCHAR(50) UNIQUE NOT NULL,
    		email VARCHAR(100) UNIQUE NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
	    	password_hash VARCHAR(255) NOT NULL
		);

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
	    	room_id INTEGER REFERENCES rooms(id),
    		user_id INTEGER REFERENCES users(id),
    		text TEXT NOT NULL,
    		timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`
	_, err := db.Exec(query)
	if err != nil {
		log.Error(
			"internal.storage.postgres.ApplyMigrations",
			"error", err.Error(),
		)
		os.Exit(1)
	}
}
