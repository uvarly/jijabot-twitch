package duelling

import (
	"context"
	"fmt"

	"jijabot/internal/store"
)

type MMRHistoryRepository interface {
	WithExecutor(executor store.Executor) MMRHistoryRepository
	Record(ctx context.Context, duelID, userID int64, ratingDelta int) error
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
