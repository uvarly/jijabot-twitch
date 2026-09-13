package grant

import (
	"context"
	"fmt"

	"jijabot/internal/store"
)

type GrantRepository interface {
	WithExecutor(executor store.Executor) GrantRepository
	Record(ctx context.Context, granterID, granteeID, amount int64) error
}

type SQLiteGrantRepository struct {
	executor store.Executor
}

func NewSQLiteGrantRepository(executor store.Executor) *SQLiteGrantRepository {
	return &SQLiteGrantRepository{executor: executor}
}

func (r *SQLiteGrantRepository) WithExecutor(executor store.Executor) GrantRepository {
	return &SQLiteGrantRepository{executor: executor}
}

func (r *SQLiteGrantRepository) Record(ctx context.Context, granterID, granteeID, amount int64) error {
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
