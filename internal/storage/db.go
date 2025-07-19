package storage

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/P3rCh1/chat-server/internal/models"
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
	fmt.Println("Successfully connected to PostgreSQL")
	return db, nil
}

func InitDB() (*sql.DB, error) {
	return NewDB(Config{
		Host:     "localhost",
		Port:     5432,
		User:     "chat",
		Password: "pass",
		DBName:   "chatdb",
	})
}

func ApplyMigrations(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS users (
	    	id SERIAL PRIMARY KEY,
    		username VARCHAR(50) UNIQUE NOT NULL,
    		email VARCHAR(100) UNIQUE NOT NULL,
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

func (r *UserRepository) CreateUser (user models.User) error {
	query := `
		INSERT INTO users (username, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING ID
	`
	err := r.db.QueryRow(query, user.Username, user.Email, user.Password).Scan(&user.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("Email или username уже заняты")
		} else {
			return fmt.Errorf("Ошибка при создании пользователя: %w", err)
		}
	}
	return nil
}