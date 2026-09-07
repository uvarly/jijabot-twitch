package gambling

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"time"

	"jijabot/internal/bet"
	"jijabot/internal/dailywindow"
	"jijabot/internal/store"
	"jijabot/internal/users"
	"jijabot/internal/wallet"
)

var (
	ErrInvalidAmount     = errors.New("gambling: bet amount must be positive")
	ErrDailyLimitReached = errors.New("gambling: daily bet limit reached")
	ErrInsufficientFunds = wallet.ErrInsufficientFunds
)

type OddsCalculator interface {
	Odds(ctx context.Context, userID int64, base float64) (float64, error)
}

type StaticOddsCalculator struct {
}

func (StaticOddsCalculator) Odds(_ context.Context, _ int64, base float64) (float64, error) {
	return base, nil
}

type RNG interface {
	Float64() float64
}

type MathRandRNG struct{}

func (MathRandRNG) Float64() float64 { return rand.Float64() }

type Clock interface {
	Now() time.Time
}

type RealClock struct{}

func (RealClock) Now() time.Time { return time.Now() }

type Result struct {
	Won        bool
	Payout     int64
	NewBalance int64
}

type Option func(*BetPlacer)

func WithOddsCalculator(oddsCalculator OddsCalculator) Option {
	return func(bp *BetPlacer) { bp.oddsCalculator = oddsCalculator }
}

func WithRNG(rng RNG) Option {
	return func(bp *BetPlacer) { bp.rng = rng }
}

func WithClock(clock Clock) Option {
	return func(bp *BetPlacer) { bp.clock = clock }
}

type BetPlacer struct {
	txBeginner     store.TxBeginner
	users          users.Repository
	wallet         wallet.Repository
	bets           bet.Repository
	oddsCalculator OddsCalculator
	rng            RNG
	clock          Clock

	baseProbability float64
	payoutMultiple  float64
	dailyLimit      int
	resetHour       int
}

func NewBetPlacer(
	txBeginner store.TxBeginner,
	userRepository users.Repository,
	betRepository bet.Repository,
	walletRepository wallet.Repository,
	baseProbability float64,
	payoutMultiple float64,
	dailyLimit int,
	resetHour int,
	options ...Option,
) *BetPlacer {
	betPlacer := &BetPlacer{
		txBeginner:      txBeginner,
		users:           userRepository,
		bets:            betRepository,
		wallet:          walletRepository,
		oddsCalculator:  StaticOddsCalculator{},
		rng:             MathRandRNG{},
		clock:           RealClock{},
		baseProbability: baseProbability,
		payoutMultiple:  payoutMultiple,
		dailyLimit:      dailyLimit,
		resetHour:       resetHour,
	}

	for _, o := range options {
		o(betPlacer)
	}

	return betPlacer
}

func (bp *BetPlacer) Place(ctx context.Context, twitchUserID, username string, amount int64) (Result, error) {
	if amount <= 0 {
		return Result{}, ErrInvalidAmount
	}

	tx, err := bp.txBeginner.BeginTx(ctx, nil)
	if err != nil {
		return Result{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	userTx := bp.users.WithExecutor(tx)
	betTx := bp.bets.WithExecutor(tx)
	walletTx := bp.wallet.WithExecutor(tx)

	user, err := userTx.GetOrCreate(ctx, twitchUserID, username)
	if err != nil {
		return Result{}, fmt.Errorf("failed to get or create user: %w", err)
	}

	currentBalance, err := walletTx.Balance(ctx, user.ID)
	if err != nil {
		return Result{}, fmt.Errorf("failed to get balance: %w", err)
	}

	if currentBalance < amount {
		return Result{}, ErrInsufficientFunds
	}

	period := dailywindow.Key(bp.clock.Now(), bp.resetHour)

	count, err := betTx.CountInPeriod(ctx, user.ID, period)
	if err != nil {
		return Result{}, fmt.Errorf("failed to count bets in current period: %w", err)
	}

	if count >= bp.dailyLimit {
		return Result{}, ErrDailyLimitReached
	}

	probability, err := bp.oddsCalculator.Odds(ctx, user.ID, bp.baseProbability)
	if err != nil {
		return Result{}, fmt.Errorf("failed to calculate odds: %w", err)
	}

	won := bp.rng.Float64() < probability

	var payout int64

	if won {
		payout = int64(math.Round(float64(amount)*bp.payoutMultiple)) - amount
	} else {
		payout = -amount
	}

	newBalance, err := walletTx.Credit(ctx, user.ID, payout)
	if err != nil {
		if errors.Is(err, ErrInsufficientFunds) {
			return Result{}, ErrInsufficientFunds
		}

		return Result{}, fmt.Errorf("failed to settle bet: %w", err)
	}

	if err := betTx.Record(ctx, bet.Bet{
		UserID:      user.ID,
		Amount:      amount,
		Probability: probability,
		Won:         won,
		Payout:      payout,
		Period:      period,
	}); err != nil {
		return Result{}, fmt.Errorf("failed to record bet: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return Result{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return Result{
		Won:        won,
		Payout:     payout,
		NewBalance: newBalance,
	}, nil
}
