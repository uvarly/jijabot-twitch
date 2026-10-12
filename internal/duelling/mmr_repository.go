package duelling

import (
	"context"
	"fmt"

	"jijabot/internal/store"
)

type MMREntry struct {
	UserID int64
	Rating int64
}

type MMRRepository interface {
	WithExecutor(executor store.Executor) MMRRepository
	GetOrCreate(ctx context.Context, userID int64, defaultRating int) (int, error)
	Set(ctx context.Context, userID int64, rating int) error
	GetTopMMR(ctx context.Context, limit int) ([]MMREntry, error)
}

type SQLiteMMRRepository struct {
	executor store.Executor
}

func NewSQLiteMMRRepository(executor store.Executor) *SQLiteMMRRepository {
	return &SQLiteMMRRepository{executor: executor}
}

func (r *SQLiteMMRRepository) WithExecutor(executor store.Executor) MMRRepository {
	return &SQLiteMMRRepository{executor: executor}
}

func (r *SQLiteMMRRepository) GetOrCreate(ctx context.Context, userID int64, defaultRating int) (int, error) {
	var query = `
		INSERT OR IGNORE INTO arena_duel_mmr (user_id, rating)
		VALUES (?, ?)
	`

	if _, err := r.executor.ExecContext(ctx, query, userID, defaultRating); err != nil {
		return 0, fmt.Errorf("failed to ensure mmr row exists: %w", err)
	}

	query = `
		SELECT rating
		FROM arena_duel_mmr
		WHERE user_id = ?
	`

	var rating int

	if err := r.executor.QueryRowContext(ctx, query, userID).Scan(&rating); err != nil {
		return 0, fmt.Errorf("failed to get mmr: %w", err)
	}

	return rating, nil
}

func (r *SQLiteMMRRepository) Set(ctx context.Context, userID int64, rating int) error {
	const query = `
		UPDATE arena_duel_mmr
		SET rating = ?, updated_at = CURRENT_TIMESTAMP
		WHERE user_id = ?
	`

	if _, err := r.executor.ExecContext(ctx, query, rating, userID); err != nil {
		return fmt.Errorf("failed to set mmr: %w", err)
	}

	return nil
}

func (r *SQLiteMMRRepository) GetTopMMR(ctx context.Context, limit int) ([]MMREntry, error) {
	const query = `
		SELECT user_id, rating
		FROM arena_duel_mmr
		ORDER BY rating DESC, updated_at ASC, user_id ASC
		LIMIT ?
	`

	rows, err := r.executor.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get top mmr entries: %w", err)
	}
	defer rows.Close()

	var entries []MMREntry

	for rows.Next() {
		var entry MMREntry

		if err := rows.Scan(&entry.UserID, &entry.Rating); err != nil {
			return nil, fmt.Errorf("failed to scan top mmr entry: %w", err)
		}

		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate top mmr entries: %w", err)
	}

	return entries, nil
}
