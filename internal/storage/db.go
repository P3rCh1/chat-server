package storage

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/P3rCh1/chat-server/internal/pkg/models"
	"github.com/P3rCh1/chat-server/internal/pkg/msg"
	_ "github.com/lib/pq"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

func NewDB(cfg Config) (*sql.DB, error) {
	psqlInfo := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName,
	)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}

func ApplyMigrations(db *sql.DB) error {
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
    		name VARCHAR(100) NOT NULL,
    		is_private BOOLEAN DEFAULT false,
    		creator_id INTEGER REFERENCES users(id)
		);

		CREATE TABLE IF NOT EXISTS messages (
    		id SERIAL PRIMARY KEY,
	    	room_id INTEGER REFERENCES rooms(id),
    		user_id INTEGER REFERENCES users(id),
    		text TEXT NOT NULL,
    		timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`
	_, err := db.Exec(query)
	return err
}

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db}
}

func (r *UserRepository) CreateUser(user models.UserRequest) (models.Profile, error) {
	query := `
		INSERT INTO users (username, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	profile := models.Profile{
		Username: user.Username,
		Email:    user.Email,
	}
	err := r.db.QueryRow(query, user.Username, user.Email, user.Password).Scan(&profile.ID, &profile.CreatedAt)
	if err != nil {
		return models.Profile{}, msg.UserOrEmailAlreadyExist
	}
	return profile, nil
}

func (r *UserRepository) ChangeName(id int, newName string) error {
	row := r.db.QueryRow("SELECT username FROM users WHERE id = $1", id)
	var username string
	if err := row.Scan(&username); err != nil {
		return msg.UserNotFound
	}
	if username == newName {
		return msg.New(http.StatusBadRequest, "new name matches the current")
	}
	row = r.db.QueryRow("SELECT count(*) FROM users WHERE username = $1", newName)
	var count int
	if err := row.Scan(&count); err != nil {
		return msg.ServerError
	}
	if count != 0 {
		return msg.UserAlreadyExist
	}
	res, err := r.db.Exec("UPDATE users SET username = $1 WHERE id = $2", newName, id)
	if err != nil {
		return msg.ServerError
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return msg.UserNotFound
	}
	return nil
}
