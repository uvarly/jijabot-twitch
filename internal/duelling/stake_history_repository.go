package duelling

import (
	"context"
	"fmt"

	"jijabot/internal/store"
)

type StakeReason string

const (
	StakeReasonStake  = "stake"
	StakeReasonRefund = "refund"
	StakeReasonPayout = "payout"
)

type StakeHistoryRepository interface {
	WithExecutor(executor store.Executor) StakeHistoryRepository
	Record(ctx context.Context, duelID int64, userID int64, amount int64, reason StakeReason) error
}

type SQLiteStakeHistoryRepository struct {
	executor store.Executor
}

func NewSQLiteStakeHistoryRepository(executor store.Executor) *SQLiteStakeHistoryRepository {
	return &SQLiteStakeHistoryRepository{executor: executor}
}

func (r *SQLiteStakeHistoryRepository) WithExecutor(executor store.Executor) StakeHistoryRepository {
	return &SQLiteStakeHistoryRepository{executor: executor}
}

func (r *SQLiteStakeHistoryRepository) Record(ctx context.Context, duelID int64, userID int64, amount int64, stakeReason StakeReason) error {
	const query = `
		INSERT INTO arena_duel_stake_history (duel_id, user_id, amount, reason)
		VALUES (?, ?, ?, ?)
	`

	if _, err := r.executor.ExecContext(ctx, query, duelID, userID, amount, string(stakeReason)); err != nil {
		return fmt.Errorf("failed to record stake history: %w", err)
	}

	return nil
}
