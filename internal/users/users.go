package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"jijabot/internal/store"
)

type User struct {
	ID           int64
	TwitchUserID string
	Username     string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Repository interface {
	WithExecutor(executor store.Executor) Repository
	GetOrCreate(ctx context.Context, twitchUserID, username string) (User, error)
	GetByUsername(ctx context.Context, username string) (User, bool, error)
}

type SQLiteRepository struct {
	executor store.Executor
}

func NewSQLiteRepository(executor store.Executor) *SQLiteRepository {
	return &SQLiteRepository{executor: executor}
}

func (r *SQLiteRepository) WithExecutor(executor store.Executor) Repository {
	return &SQLiteRepository{executor: executor}
}

func (r *SQLiteRepository) GetOrCreate(ctx context.Context, twitchUserID, username string) (User, error) {
	const query = `
		INSERT INTO users (twitch_user_id, username)
		VALUES (?, ?)
		ON CONFLICT (twitch_user_id) DO UPDATE SET
			username = EXCLUDED.username,
			updated_at = CURRENT_TIMESTAMP
		RETURNING id, twitch_user_id, username, created_at, updated_at
	`

	var user User

	err := r.executor.QueryRowContext(ctx, query, twitchUserID, username).
		Scan(&user.ID, &user.TwitchUserID, &user.Username, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return User{}, fmt.Errorf("failed to get or create user: %w", err)
	}

	return user, nil
}

func (r *SQLiteRepository) GetByUsername(ctx context.Context, username string) (User, bool, error) {
	const query = `
		SELECT id, twitch_user_id, username, created_at, updated_at
		FROM users
		WHERE username = ?
	`

	var user User

	err := r.executor.QueryRowContext(ctx, query, username).
		Scan(&user.ID, &user.TwitchUserID, &user.Username, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, false, nil
		}

		return User{}, false, fmt.Errorf("failed to get user: %w", err)
	}

	return user, true, nil
}
