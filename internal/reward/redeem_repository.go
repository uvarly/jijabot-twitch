package reward

import (
	"context"
	"errors"
	"fmt"

	"modernc.org/sqlite"
	sqlitelib "modernc.org/sqlite/lib"

	"jijabot/internal/store"
)

var ErrAlreadyRedeemed = errors.New("reward: already redeemed for this period")

type RedeemRepository interface {
	WithExecutor(executor store.Executor) RedeemRepository
	Record(ctx context.Context, userID, amount int64, period string) error
}

type SQLiteRedeemRepository struct {
	executor store.Executor
}

func NewSQLiteRedeemRepository(executor store.Executor) *SQLiteRedeemRepository {
	return &SQLiteRedeemRepository{executor: executor}
}

func (r *SQLiteRedeemRepository) WithExecutor(executor store.Executor) RedeemRepository {
	return &SQLiteRedeemRepository{executor: executor}
}

func (r *SQLiteRedeemRepository) Record(ctx context.Context, userID, amount int64, period string) error {
	const query = `
		INSERT INTO jija_coin_redeem_history (user_id, amount, redeem_period)
		VALUES (?, ?, ?)
	`

	if _, err := r.executor.ExecContext(ctx, query, userID, amount, period); err != nil {
		if isUniqueViolation(err) {
			return ErrAlreadyRedeemed
		}

		return fmt.Errorf("failed to record redeem: %w", err)
	}

	return nil
}

func isUniqueViolation(err error) bool {
	if sqliteErr, ok := errors.AsType[*sqlite.Error](err); ok {
		return sqliteErr.Code() == sqlitelib.SQLITE_CONSTRAINT_UNIQUE
	}

	return false
}
