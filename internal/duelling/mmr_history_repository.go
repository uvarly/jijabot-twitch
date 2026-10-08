package duelling

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"jijabot/internal/store"
)

type MMRHistoryRepository interface {
	WithExecutor(executor store.Executor) MMRHistoryRepository
	Record(ctx context.Context, duelID, userID int64, ratingDelta int) error
	CountWinsByUserID(ctx context.Context, userID int64) (int64, error)
}

type SQLiteMMRHistoryRepository struct {
	executor store.Executor
}

func NewSQLiteMMRHistoryRepository(executor store.Executor) *SQLiteMMRHistoryRepository {
	return &SQLiteMMRHistoryRepository{executor: executor}
}

func (r *SQLiteMMRHistoryRepository) WithExecutor(executor store.Executor) MMRHistoryRepository {
	return &SQLiteMMRHistoryRepository{executor: executor}
}

func (r *SQLiteMMRHistoryRepository) Record(ctx context.Context, duelID, userID int64, ratingDelta int) error {
	const query = `
		INSERT INTO arena_duel_mmr_history (duel_id, user_id, rating_delta)
		VALUES (?, ?, ?)
	`

	if _, err := r.executor.ExecContext(ctx, query, duelID, userID, ratingDelta); err != nil {
		return fmt.Errorf("failed to record mmr history: %w", err)
	}

	return nil
}

func (r *SQLiteMMRHistoryRepository) CountWinsByUserID(ctx context.Context, userID int64) (int64, error) {
	const query = `
		SELECT COUNT(*)
		FROM arena_duel_mmr_history
		WHERE user_id = ? AND rating_delta > 0
	`

	var count int64

	if err := r.executor.QueryRowContext(ctx, query, userID).Scan(&count); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}

		return 0, fmt.Errorf("failed to count wins: %w", err)
	}

	// if err != nil {
	// 	if errors.Is(err, sql.ErrNoRows) {
	// 		return Duel{}, false, nil
	// 	}

	// 	return Duel{}, false, fmt.Errorf("failed to scan pending duel: %w", err)
	// }

	return count, nil
}
