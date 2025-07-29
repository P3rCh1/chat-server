package repository

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	"github.com/P3rCh1/chat-server/internal/models"
	"github.com/P3rCh1/chat-server/internal/pkg/responses"
	"github.com/P3rCh1/chat-server/internal/storage/cache"
)

type Repository struct {
	DB           *sql.DB
	ProfileCache *cache.ProfileCacher
}

func NewRepository(db *sql.DB, profileCache *cache.ProfileCacher) *Repository {
	return &Repository{db, profileCache}
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
	err := r.DB.QueryRow(query, username).Scan(&id)
	if err != nil {
		log.Error(
			"internal.storage.postgres.MustCreateInternalUser",
			"error", err.Error(),
		)
		os.Exit(1)
	}
	return id
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
	err := r.DB.QueryRow(query, user.Username, user.Email, user.Password).Scan(&profile.ID, &profile.CreatedAt)
	if err != nil {
		err = fmt.Errorf("fail to create user: %w", err)
	}
	go func() {
		r.ProfileCache.Set(&profile)
	}()
	return profile, err
}

func (r *Repository) ChangeName(userID int, newName string) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction to change name: %w", err)
	}
	defer tx.Rollback()
	row := tx.QueryRow("SELECT count(*) FROM users WHERE username = $1", newName)
	var count int
	if err := row.Scan(&count); err != nil {
		return fmt.Errorf("failed to check name existence: %w", err)
	}
	if count != 0 {
		return responses.UserAlreadyExist
	}
	_, err = tx.Exec("UPDATE users SET username = $1 WHERE id = $2", newName, userID)
	if err != nil {
		return fmt.Errorf("fail to start update name: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	go func() {
		profile, err := r.Profile(userID)
		if err == nil {
			profile.Username = newName
			r.ProfileCache.Set(profile)
		} else {
			r.ProfileCache.Delete(profile.ID)
		}
	}()
	return nil
}

func (r *Repository) IsRoomMember(userID int, roomID int) (bool, error) {
	var isRoomMember bool
	err := r.DB.QueryRow(
		`SELECT EXISTS(
            SELECT 1 FROM room_members 
            WHERE user_id = $1 AND room_id = $2
        )`,
		userID,
		roomID,
	).Scan(&isRoomMember)
	if err != nil {
		err = fmt.Errorf("fail to check room membership: %w", err)
	}
	return isRoomMember, err
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
	row := r.DB.QueryRow(query, roomID, userID, msg.Text)
	err := row.Scan(&msg.Timestamp)
	if err != nil {
		return fmt.Errorf("store msg fail: %w", err)
	}
	return nil
}

func (r *Repository) Profile(userID int) (*models.Profile, error) {
	if profile, _ := r.ProfileCache.Get(userID); profile != nil {
		return profile, nil
	}
	const query = `
		SELECT username, email, created_at
		FROM users
		WHERE id = $1
	`
	row := r.DB.QueryRow(query, userID)
	profile := &models.Profile{
		ID: userID,
	}
	err := row.Scan(&profile.Username, &profile.Email, &profile.CreatedAt)
	if err != nil {
		err = fmt.Errorf("fail to get profile: %w", err)
	}
	go func() {
		r.ProfileCache.Set(profile)
	}()
	return profile, err
}

func (r *Repository) GetUsername(userID int) (string, error) {
	profile, err := r.Profile(userID)
	return profile.Username, err
}

func (r *Repository) CreateRoom(room *models.Room) error {
	const query = `
		INSERT INTO rooms (name, is_private, creator_id)
		VALUES ($1, $2, $3)
		RETURNING id
	`
	tx, err := r.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed start transaction: %w", err)
	}
	defer tx.Rollback()
	err = tx.QueryRow(query, room.Name, room.IsPrivate, room.CreatorID).Scan(&room.ID)
	if err != nil {
		return fmt.Errorf("failed to create room: %w", err)
	}
	_, err = tx.Exec("INSERT INTO room_members (user_id, room_id) VALUES ($1, $2)", room.CreatorID, room.ID)
	if err != nil {
		return fmt.Errorf("failed to add creator to room: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}
	return nil
}

func (r *Repository) CreatorID(roomID int) (int, error) {
	const query = `
		SELECT creator_id FROM rooms WHERE id = $1
	`
	var creatorID int
	err := r.DB.QueryRow(query, roomID).Scan(&creatorID)
	if err != nil {
		return 0, fmt.Errorf("failed to find creator id: %w", err)
	}
	return creatorID, nil
}

func (r *Repository) AddToRoom(userID, roomID int) error {
	const query = `
		INSERT INTO room_members (user_id, room_id)
		VALUES ($1, $2)
		RETURNING user_id
	`
	_, err := r.DB.Exec(query, userID, roomID)
	if err != nil {
		return fmt.Errorf("failed add to room: %w", err)
	}
	return nil
}

func (r *Repository) IsPrivate(roomID int) (bool, error) {
	const query = `
		SELECT is_private FROM rooms WHERE id = $1
	`
	var isPrivate bool
	err := r.DB.QueryRow(query, roomID).Scan(&isPrivate)
	if err != nil {
		return false, fmt.Errorf("failed to check room's private status: %w", err)
	}
	return isPrivate, nil
}
