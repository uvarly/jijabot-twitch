package duelling

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

	"jijabot/internal/store"
	"jijabot/internal/users"
	"jijabot/internal/wallet"
)

var (
	ErrInvalidStake        = errors.New("duelling: stake must be positive")
	ErrCannotChallengeSelf = errors.New("duelling: cannot challenge yourself")
)

type ChallengeResult struct{}
type CancelResult struct{}
type AcceptResult struct{}
type DeclineResult struct{}

type Duellist_ReferenceForClient_DeleteMe interface {
	Challenge(ctx context.Context, challengerTwitchID, challengerName, opponentName string, stake int64) (ChallengeResult, error)
	Accept(ctx context.Context, opponentTwitchID, opponentName string) (AcceptResult, error)
	Decline(ctx context.Context, opponentTwitchID, opponentName string) (DeclineResult, error)
	Cancel(ctx context.Context, challengerTwitchID, challengerName string) (CancelResult, error)
}

type Clock interface {
	Now() time.Time
}

type RealClock struct{}

func (RealClock) Now() time.Time { return time.Now() }

type RNG interface {
	Int64N(n int64) int64
}

type MathRandRNG struct{}

func (MathRandRNG) Int64N(n int64) int64 { return rand.Int64N(n) }

type Option func(*Duellist)

func WithClock(clock Clock) Option {
	return func(d *Duellist) { d.clock = clock }
}

func WithRNG(rng RNG) Option {
	return func(d *Duellist) { d.rng = rng }
}

func WithRatingCalculator(ratingCalculator RatingCalculator) Option {
	return func(d *Duellist) { d.ratingCalculator = ratingCalculator }
}

type Duellist struct {
	txBeginner       store.TxBeginner
	userRepository   users.Repository
	walletRepository wallet.Repository
	// duelRepository         DuelRepository
	// stakeHistoryRepository StakeHistoryRepository
	// mmrRepository          MMRRepository
	// mmrHistoryRepository   MMRHistoryRepository

	clock            Clock
	rng              RNG
	ratingCalculator RatingCalculator

	// expiresIn     time.Duration
	// defaultRating int
}

func NewDuellist(
	txBeginner store.TxBeginner,
	userRepository users.Repository,
	walletRepository wallet.Repository,
	kFactor int,
	options ...Option,
) *Duellist {
	duellist := &Duellist{
		txBeginner:       txBeginner,
		userRepository:   userRepository,
		walletRepository: walletRepository,
		clock:            RealClock{},
		rng:              MathRandRNG{},
		ratingCalculator: NewEloCalculator(kFactor),
	}

	for _, o := range options {
		o(duellist)
	}

	return duellist
}

func (d *Duellist) Challenge(ctx context.Context, challengerTwitchID, challengerName, opponentName string, stake int64) (ChallengeResult, error) {
	if stake <= 0 {
		return ChallengeResult{}, ErrInvalidStake
	}

	if strings.EqualFold(challengerName, opponentName) {
		return ChallengeResult{}, ErrCannotChallengeSelf
	}

	tx, err := d.txBeginner.BeginTx(ctx, nil)
	if err != nil {
		return ChallengeResult{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	_ = d.userRepository.WithExecutor(tx)
	_ = d.walletRepository.WithExecutor(tx)
	// duelTx := d.duelRepository.WithExecutor(tx)
	// stakeTx := d.stakeRepository.WithExecutor(tx)

	return ChallengeResult{}, nil
}

func (d *Duellist) Accept(ctx context.Context, opponentTwitchID, opponentName string) (AcceptResult, error)
func (d *Duellist) Decline(ctx context.Context, opponentTwitchID, opponentName string) (DeclineResult, error)
func (d *Duellist) Cancel(ctx context.Context, challengerTwitchID, challengerName string) (CancelResult, error)

// func (d *Duellist) expirePending(ctx context.Context, duelTx DuelRepository, stakeTx StakeRepository, walletTx WalletRepository) error {
// 	expired, err := duelTx.ExpirePending(ctx, d.clock.Now())
// 	if err != nil {
// 		return fmt.Errorf("failed to expire pending duels: %w", err)
// 	}

// 	for _, e := range expired {
// 		if _, err := walletTx.Credit(ctx, e.ChallengerID, e.Stake); err != nil {
// 			return fmt.Errorf("failed to refund expired duel stake: %w", err)
// 		}

// 		if err := stakeTx.Record(); err != nil {
// 		}
// 	}

// 	return nil
// }
