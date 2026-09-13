package wallet

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"modernc.org/sqlite"
	sqlitelib "modernc.org/sqlite/lib"

	"jijabot/internal/store"
)

var ErrInsufficientFunds = errors.New("wallet: insufficient funds")

type Repository interface {
	WithExecutor(executor store.Executor) Repository
	Balance(ctx context.Context, userID int64) (int64, error)
	Credit(ctx context.Context, userID int64, amount int64) (int64, error)
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

func (r *SQLiteRepository) Balance(ctx context.Context, userID int64) (int64, error) {
	const query = `SELECT balance FROM jija_coin_wallets WHERE user_id = ?`

	var balance int64

	err := r.executor.QueryRowContext(ctx, query, userID).Scan(&balance)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}

	if err != nil {
		return 0, fmt.Errorf("failed to get balance: %w", err)
	}

	return balance, nil
}

func (r *SQLiteRepository) Credit(ctx context.Context, userID int64, amount int64) (int64, error) {
	var query = `
		INSERT OR IGNORE INTO jija_coin_wallets (user_id, balance)
		VALUES (?, 0)
	`

	if _, err := r.executor.ExecContext(ctx, query, userID); err != nil {
		return 0, fmt.Errorf("failed to ensure wallet exists: %w", err)
	}

	query = `
		UPDATE jija_coin_wallets
		SET balance = balance + ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE user_id = ?
		RETURNING balance
	`

	var balance int64

	if err := r.executor.QueryRowContext(ctx, query, amount, userID).Scan(&balance); err != nil {
		if isCheckViolation(err) {
			return 0, ErrInsufficientFunds
		}

		return 0, fmt.Errorf("failed to credit balance: %w", err)
	}

	return balance, nil
}

func isCheckViolation(err error) bool {
	if sqliteErr, ok := errors.AsType[*sqlite.Error](err); ok {
		return sqliteErr.Code() == sqlitelib.SQLITE_CONSTRAINT_CHECK
	}

	return false
}
