package duelling

import (
	"context"
	"fmt"

	"jijabot/internal/store"
	"jijabot/internal/users"
)

type TopResult struct {
	Username string
	MMR      int64
	WinCount int64
}

type StatsResult struct {
	WinCount  int64
	LossCount int64
	DrawCount int64
	WinRate   float64
}

type RankResult struct {
	MMR int64
}

type StatsProvider struct {
	txBeginner           store.TxBeginner
	userRepository       users.Repository
	mmrRepository        MMRRepository
	mmrHistoryRepository MMRHistoryRepository
}

func NewStatsProvider(
	txBeginner store.TxBeginner,
	userRepository users.Repository,
	mmrRepository MMRRepository,
	mmrHistoryRepository MMRHistoryRepository,
) *StatsProvider {
	return &StatsProvider{
		txBeginner:           txBeginner,
		userRepository:       userRepository,
		mmrRepository:        mmrRepository,
		mmrHistoryRepository: mmrHistoryRepository,
	}
}

// userTx := sp.userRepository.WithExecutor(tx)
// mmrTx := sp.mmrRepository.WithExecutor(tx)
// mmrHistoryTx := sp.mmrHistoryRepository.WithExecutor(tx)

func (sp *StatsProvider) Top(ctx context.Context) (TopResult, error) {
	tx, err := sp.txBeginner.BeginTx(ctx, nil)
	if err != nil {
		return TopResult{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	userTx := sp.userRepository.WithExecutor(tx)
	mmrTx := sp.mmrRepository.WithExecutor(tx)
	mmrHistoryTx := sp.mmrHistoryRepository.WithExecutor(tx)

	topMMREntry, err := mmrTx.GetTopMMREntry(ctx)
	if err != nil {
		return TopResult{}, fmt.Errorf("failed to get top MMR entry: %w", err)
	}

	if topMMREntry.UserID == 0 {
		return TopResult{}, ErrNoTopUserYet
	}

	user, found, err := userTx.GetByID(ctx, topMMREntry.UserID)
	if err != nil {
		return TopResult{}, fmt.Errorf("failed to get user: %w", err)
	}

	if !found {
		return TopResult{}, ErrNoTopUserYet
	}

	winCount, err := mmrHistoryTx.CountWinsByUserID(ctx, user.ID)
	if err != nil {
		return TopResult{}, fmt.Errorf("failed to count wins: %w", err)
	}

	return TopResult{
		Username: user.Username,
		MMR:      topMMREntry.Rating,
		WinCount: winCount,
	}, nil
}

func (sp *StatsProvider) Stats(ctx context.Context, twitchUserID string) (StatsResult, error) {
	tx, err := sp.txBeginner.BeginTx(ctx, nil)
	if err != nil {
		return StatsResult{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// userTx := sp.userRepository.WithExecutor(tx)
	// duelTx := sp.duelRepository.WithExecutor(tx)
	// mmrTx := sp.mmrRepository.WithExecutor(tx)
	// mmrHistoryTx := sp.mmrHistoryRepository.WithExecutor(tx)

	return StatsResult{}, nil
}

func (sp *StatsProvider) Rank(ctx context.Context, twitchUserID string) (RankResult, error) {
	tx, err := sp.txBeginner.BeginTx(ctx, nil)
	if err != nil {
		return RankResult{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	return RankResult{}, nil
}
