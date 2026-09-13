package bet

import (
	"context"
	"fmt"
	"jijabot/internal/store"
	"time"
)

type Bet struct {
	ID          int64
	UserID      int64
	Amount      int64
	Probability float64
	Won         bool
	Payout      int64
	Period      string
	PlacedAt    time.Time
}

type Repository interface {
	CountInPeriod(ctx context.Context, userID int64, period string) (int, error)
	Record(ctx context.Context, b Bet) error
	WithExecutor(executor store.Executor) Repository
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

func (r *SQLiteRepository) CountInPeriod(ctx context.Context, userID int64, period string) (int, error) {
	const query = `
		SELECT COUNT(*) FROM jija_coin_bet_history
		WHERE user_id = ? AND bet_period = ?
	`

	var count int

	if err := r.executor.QueryRowContext(ctx, query, userID, period).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count bets: %w", err)
	}

	return count, nil
}

func (r *SQLiteRepository) Record(ctx context.Context, b Bet) error {
	const query = `
		INSERT INTO jija_coin_bet_history (user_id, amount, probability, won, payout, bet_period)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	won := 0
	if b.Won {
		won = 1
	}

	if _, err := r.executor.ExecContext(ctx, query, b.UserID, b.Amount, b.Probability, won, b.Payout, b.Period); err != nil {
		return fmt.Errorf("failed to record bet: %w", err)
	}

	return nil
}
