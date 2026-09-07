package redeem

import (
	"context"
	"errors"
	"fmt"

	"modernc.org/sqlite"
	sqlitelib "modernc.org/sqlite/lib"

	"jijabot/internal/store"
)

var ErrAlreadyClaimed = errors.New("redeem: already claimed for this period")

type Repository interface {
	WithExecutor(executor store.Executor) Repository
	Claim(ctx context.Context, userID, amount int64, period string) error
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

func (r *SQLiteRepository) Claim(ctx context.Context, userID, amount int64, period string) error {
	const query = `
		INSERT INTO jija_coin_redeem_history (user_id, amount, redeem_period)
		VALUES (?, ?, ?)
	`

	if _, err := r.executor.ExecContext(ctx, query, userID, amount, period); err != nil {
		if isUniqueViolation(err) {
			return ErrAlreadyClaimed
		}

		return fmt.Errorf("failed to record claim: %w", err)
	}

	return nil
}

func isUniqueViolation(err error) bool {
	if sqliteErr, ok := errors.AsType[*sqlite.Error](err); ok {
		return sqliteErr.Code() == sqlitelib.SQLITE_CONSTRAINT_UNIQUE
	}

	return false
}
