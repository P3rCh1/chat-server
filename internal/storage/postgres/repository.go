package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/P3rCh1/chat-server/internal/models"
	"github.com/P3rCh1/chat-server/internal/pkg/msg"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db}
}

func (r *Repository) CreateUser(user models.RegisterRequest) (models.Profile, error) {
	const query = `
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

func (r *Repository) MustCreateInternalUser(log *slog.Logger, username string) int {
	const query = `
        WITH insert_attempt AS (
            INSERT INTO users (username, email, password_hash)
            VALUES ($1, $1, '')
            ON CONFLICT (username) DO NOTHING
            RETURNING id
        )
        SELECT id FROM insert_attempt
        UNION ALL
        SELECT id FROM users WHERE username = $1
        LIMIT 1
    `
	var id int
	err := r.db.QueryRow(query, username).Scan(&id)
	if err != nil {
		log.Error(
			"internal.storage.postgres.MustCreateInternalUser",
			"error", err.Error(),
		)
		os.Exit(1)
	}
	return id
}

func (r *Repository) ChangeName(id int, newName string) error {
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

func (r *Repository) IsRoomMember(userID int, roomID int) error {
	var exists bool
	err := r.db.QueryRow(
		`SELECT EXISTS(
            SELECT 1 FROM room_members 
            WHERE user_id = $1 AND room_id = $2
        )`,
		userID,
		roomID,
	).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check room access: %w", err)
	}
	if !exists {
		return errors.New("user is not a member of the room")
	}
	return nil
}

func (r *Repository) StoreMsg(msg *models.Message, roomID, userID int) error {
	const query = `
        INSERT INTO messages (
            room_id, 
            user_id,
            text
        ) VALUES ($1, $2, $3)
		RETURNING timestamp
    `
	row := r.db.QueryRow(query, roomID, userID, msg.Text)
	if err := row.Scan(&msg.Timestamp); err != nil {
		return errors.New("failed to store message")
	}
	return nil
}

func (r *Repository) GetUsername(userID int) (string, error) {
	var username string
	err := r.db.QueryRow("SELECT username FROM users WHERE id = $1", userID).Scan(&username)
	if err != nil {
		return "", errors.New("can not find user")
	}
	return username, nil
}
