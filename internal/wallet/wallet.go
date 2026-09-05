package wallet

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"jijabot/internal/store"
)

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
	const query = `
		INSERT INTO jija_coin_wallets (user_id, balance)
		VALUES (?, ?)
		ON CONFLICT (user_id) DO UPDATE SET
			balance = balance + EXCLUDED.balance,
			updated_at = CURRENT_TIMESTAMP
		RETURNING balance
	`

	var balance int64

	if err := r.executor.QueryRowContext(ctx, query, userID, amount).Scan(&balance); err != nil {
		return 0, fmt.Errorf("failed to credit balance: %w", err)
	}

	return balance, nil
}
