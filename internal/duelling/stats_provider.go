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
}

func NewStatsProvider(
	txBeginner store.TxBeginner,
	userRepository users.Repository,
	duelRepository DuelRepository,
	mmrRepository MMRRepository,
	mmrHistoryRepository MMRHistoryRepository,
) *StatsProvider {
	return &StatsProvider{
		txBeginner:           txBeginner,
		userRepository:       userRepository,
		duelRepository:       duelRepository,
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

	resolvedDuels, err := duelTx.FindResolvedDuelsBy(ctx, user.ID)
	if err != nil {
		return StatsResult{}, fmt.Errorf("failed to find resolved duels: %w", err)
	}

	var (
		winCount, lossCount, drawCount int64
		winRatePercent                 float64
	)

	for _, duel := range resolvedDuels {
		if *duel.Result == DuelResultDraw {
			drawCount++
			continue
		}

		if duel.ChallengerID == user.ID {
			if *duel.Result == DuelResultChallengerWon {
				winCount++
			} else {
				lossCount++
			}

			continue
		}

		if duel.OpponentID == user.ID {
			if *duel.Result == DuelResultOpponentWon {
				winCount++
			} else {
				lossCount++
			}

			continue
		}
	}

	if winCount+lossCount+drawCount != 0 {
		winRatePercent = float64(winCount) / float64(winCount+lossCount+drawCount) * 100
	}

	return StatsResult{
		WinCount:       winCount,
		LossCount:      lossCount,
		DrawCount:      drawCount,
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

	if user.ID == 0 {
		return 0, nil
	}

	rating, err := mmrTx.GetMMR(ctx, user.ID)
	if err != nil {
		return 0, fmt.Errorf("failed to get MMR: %w", err)
	}

	return int(rating), nil
}
