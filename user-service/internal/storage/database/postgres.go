package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/P3rCh1/chat-server/user-service/internal/config"
	"github.com/P3rCh1/chat-server/user-service/internal/gRPC/status_error"
	"github.com/P3rCh1/chat-server/user-service/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type Postgres struct {
	db *sql.DB
}

func migrate(ctx context.Context, db *sql.DB) error {
	const query = `
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(100) NOT NULL UNIQUE,
			email VARCHAR(100) NOT NULL UNIQUE,
			password VARCHAR(255) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
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
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()
	if err := migrate(ctx, db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to migrate postgres: %w", err)
	}
	return &Postgres{db}, nil
}

func (p *Postgres) Ping() {
	fmt.Println(p.db.Ping())
}

func (p *Postgres) Close() error {
	return p.db.Close()
}

func (p *Postgres) CreateUser(ctx context.Context, profile *models.Profile, password string) error {
	const query = `
		INSERT INTO users (username, email, password)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	row := p.db.QueryRowContext(ctx, query, profile.Username, profile.Email, hash)
	err = row.Scan(&profile.ID, &profile.CreatedAt)
	if err != nil {
		if errExists := AsUsernameOrEmailExistsErr(err); errExists != nil {
			return errExists
		}
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (p *Postgres) Login(ctx context.Context, email, password string) (*models.Profile, error) {
	query := `
		SELECT id, username, created_at, password
	 	FROM users
	  	WHERE email = $1
	`
	profile := &models.Profile{
		Email: email,
	}
	var hash string
	row := p.db.QueryRowContext(ctx, query, email)
	if err := row.Scan(&profile.ID, &profile.Username, &profile.CreatedAt, &hash); err != nil {
		if err == sql.ErrNoRows {
			return nil, status_error.NotFound
		} else {
			return nil, fmt.Errorf("unexpected database error: %w", err)
		}
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return nil, status_error.InvalidPassword
	}
	return profile, nil
}

func (p *Postgres) Profile(ctx context.Context, id int) (*models.Profile, error) {
	const query = `
		SELECT username, email, created_at
		FROM users
		WHERE id = $1
	`
	row := p.db.QueryRowContext(ctx, query, id)
	profile := &models.Profile{
		ID: id,
	}
	if err := row.Scan(&profile.Username, &profile.Email, &profile.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status_error.NotFound
		}
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}
	return profile, nil
}

func (p *Postgres) ChangeName(ctx context.Context, id int, newName string) error {
	const query = `
		UPDATE users 
        SET username = $1 
        WHERE id = $2 
	`
	res, err := p.db.ExecContext(ctx, query, newName, id)
	if err != nil {
		if isAlreadyExists(err) {
			return status_error.NameExists
		}
		return fmt.Errorf("update failed: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return status_error.NotFound
	}
	return nil
}
