package grant

import (
	"context"
	"fmt"

	"jijabot/internal/store"
)

type GrantHistoryRepository interface {
	WithExecutor(executor store.Executor) GrantHistoryRepository
	Record(ctx context.Context, granterID, granteeID, amount int64) error
}

type SQLiteGrantHistoryRepository struct {
	executor store.Executor
}

func NewSQLiteGrantHistoryRepository(executor store.Executor) *SQLiteGrantHistoryRepository {
	return &SQLiteGrantHistoryRepository{executor: executor}
}

func (r *SQLiteGrantHistoryRepository) WithExecutor(executor store.Executor) GrantHistoryRepository {
	return &SQLiteGrantHistoryRepository{executor: executor}
}

func (r *SQLiteGrantHistoryRepository) Record(ctx context.Context, granterID, granteeID, amount int64) error {
	const query = `
		INSERT INTO jija_coin_grant_history (granter_id, grantee_id, amount)
		VALUES (?, ?, ?)
	`

	_, err := r.executor.ExecContext(ctx, query, granterID, granteeID, amount)
	if err != nil {
		return fmt.Errorf("failed to record grant: %w", err)
	}

	return nil
}
