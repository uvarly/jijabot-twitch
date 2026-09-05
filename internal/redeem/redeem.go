package redeem

import (
	"context"
	"errors"
	"fmt"
	"time"

	"modernc.org/sqlite"
	sqlitelib "modernc.org/sqlite/lib"

	"jijabot/internal/store"
)

var ErrAlreadyClaimed = errors.New("redeem: already claimed for this period")

func PeriodKey(t time.Time, resetHour int) string {
	shifted := t.UTC().Add(-time.Duration(resetHour) * time.Hour)
	return shifted.Format("2006-01-02")
}

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
		var sqliteErr *sqlite.Error

		if errors.As(err, &sqliteErr) && sqliteErr.Code() == sqlitelib.SQLITE_CONSTRAINT_UNIQUE {
			return ErrAlreadyClaimed
		}

		return fmt.Errorf("failed to record claim: %w", err)
	}

	return nil
}
