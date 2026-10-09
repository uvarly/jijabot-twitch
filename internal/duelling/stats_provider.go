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
	WinCount       int64
	LossCount      int64
	DrawCount      int64
	WinRatePercent float64
}

type RankResult struct {
	MMR int64
}

type StatsProvider struct {
	txBeginner           store.TxBeginner
	userRepository       users.Repository
	duelRepository       DuelRepository
	mmrRepository        MMRRepository
	mmrHistoryRepository MMRHistoryRepository
	defaultRating        int
}

func NewStatsProvider(
	txBeginner store.TxBeginner,
	userRepository users.Repository,
	duelRepository DuelRepository,
	mmrRepository MMRRepository,
	mmrHistoryRepository MMRHistoryRepository,
	defaultRating int,
) *StatsProvider {
	return &StatsProvider{
		txBeginner:           txBeginner,
		userRepository:       userRepository,
		duelRepository:       duelRepository,
		mmrRepository:        mmrRepository,
		mmrHistoryRepository: mmrHistoryRepository,
		defaultRating:        defaultRating,
	}
}

func (sp *StatsProvider) Top(ctx context.Context) (TopResult, error) {
	tx, err := sp.txBeginner.BeginTx(ctx, nil)
	if err != nil {
		return TopResult{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	userTx := sp.userRepository.WithExecutor(tx)
	mmrTx := sp.mmrRepository.WithExecutor(tx)
	duelTx := sp.duelRepository.WithExecutor(tx)

	topMMREntry, err := mmrTx.GetTopMMR(ctx)
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

	record, err := duelTx.CountResultsBy(ctx, user.ID)
	if err != nil {
		return TopResult{}, fmt.Errorf("failed to count duel results: %w", err)
	}

	return TopResult{
		Username: user.Username,
		MMR:      topMMREntry.Rating,
		WinCount: record.Wins,
	}, nil
}

func (sp *StatsProvider) Stats(ctx context.Context, twitchUserID, username string) (StatsResult, error) {
	tx, err := sp.txBeginner.BeginTx(ctx, nil)
	if err != nil {
		return StatsResult{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	userTx := sp.userRepository.WithExecutor(tx)
	duelTx := sp.duelRepository.WithExecutor(tx)

	user, err := userTx.GetOrCreate(ctx, twitchUserID, username)
	if err != nil {
		return StatsResult{}, fmt.Errorf("failed to get or create user: %w", err)
	}

	record, err := duelTx.CountResultsBy(ctx, user.ID)
	if err != nil {
		return StatsResult{}, fmt.Errorf("failed to count duel results: %w", err)
	}

	var winRatePercent float64

	if total := record.Wins + record.Losses + record.Draws; total > 0 {
		winRatePercent = float64(record.Wins) / float64(total) * 100
	}

	return StatsResult{
		WinCount:       record.Wins,
		LossCount:      record.Losses,
		DrawCount:      record.Draws,
		WinRatePercent: winRatePercent,
	}, nil
}

func (sp *StatsProvider) Rating(ctx context.Context, twitchUserID, username string) (int, error) {
	tx, err := sp.txBeginner.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	userTx := sp.userRepository.WithExecutor(tx)
	mmrTx := sp.mmrRepository.WithExecutor(tx)

	user, err := userTx.GetOrCreate(ctx, twitchUserID, username)
	if err != nil {
		return 0, fmt.Errorf("failed to get or create user: %w", err)
	}

	rating, err := mmrTx.GetOrCreate(ctx, user.ID, sp.defaultRating)
	if err != nil {
		return 0, fmt.Errorf("failed to get MMR: %w", err)
	}

	return int(rating), nil
}
