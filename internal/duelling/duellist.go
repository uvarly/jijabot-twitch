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
	ErrInvalidStake             = errors.New("duelling: stake must be positive")
	ErrDailyLimitReached        = errors.New("duelling: daily duel limit reached") // TODO
	ErrCannotChallengeSelf      = errors.New("duelling: cannot challenge yourself")
	ErrChallengerNotFound       = errors.New("duelling: challenger not found")
	ErrOpponentNotFound         = errors.New("duelling: opponent not found")
	ErrChallengerHasPendingDuel = errors.New("duelling: challenger has a pending duel")
	ErrOpponentHasPendingDuel   = errors.New("duelling: opponent has a pending duel")
	ErrInsufficientFunds        = wallet.ErrInsufficientFunds
	ErrNoIncomingDuel           = errors.New("duelling: no incoming duel to respond to")
	ErrNoOutgoingDuel           = errors.New("duelling: no outgoing duel to cancel")
)

type ChallengeResult struct {
	ChallengerName string
	OpponentName   string
	Stake          int64
	// ExpiresAt      time.Time
	// NewBalance     int64
}

type AcceptResult struct {
	ChallengerName        string
	OpponentName          string
	Stake                 int64
	Pot                   int64
	ChallengerRoll        int
	OpponentRoll          int
	Result                DuelResult
	NewChallengerBalance  int64
	NewOpponentBalance    int64
	ChallengerRatingDelta int
	OpponentRatingDelta   int
}

type DeclineResult struct {
	ChallengerName string
	OpponentName   string
}

type CancelResult struct {
	ChallengerName string
	OpponentName   string
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
	txBeginner             store.TxBeginner
	userRepository         users.Repository
	walletRepository       wallet.Repository
	duelRepository         DuelRepository
	stakeHistoryRepository StakeHistoryRepository
	mmrRepository          MMRRepository
	mmrHistoryRepository   MMRHistoryRepository

	clock            Clock
	rng              RNG
	ratingCalculator RatingCalculator

	expiryDuration time.Duration
	defaultRating  int
}

func NewDuellist(
	txBeginner store.TxBeginner,
	userRepository users.Repository,
	walletRepository wallet.Repository,
	duelRepository DuelRepository,
	stakeHistoryRepository StakeHistoryRepository,
	mmrRepository MMRRepository,
	mmrHistoryRepository MMRHistoryRepository,
	expiryDuration time.Duration,
	defaultRating int,
	kFactor int,
	options ...Option,
) *Duellist {
	duellist := &Duellist{
		txBeginner:             txBeginner,
		userRepository:         userRepository,
		walletRepository:       walletRepository,
		duelRepository:         duelRepository,
		stakeHistoryRepository: stakeHistoryRepository,
		mmrRepository:          mmrRepository,
		mmrHistoryRepository:   mmrHistoryRepository,
		clock:                  RealClock{},
		rng:                    MathRandRNG{},
		ratingCalculator:       NewEloCalculator(kFactor),
		expiryDuration:         expiryDuration,
		defaultRating:          defaultRating,
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

	userTx := d.userRepository.WithExecutor(tx)
	walletTx := d.walletRepository.WithExecutor(tx)
	duelTx := d.duelRepository.WithExecutor(tx)
	stakeHistoryTx := d.stakeHistoryRepository.WithExecutor(tx)

	if err := d.expirePending(ctx, duelTx, stakeHistoryTx, walletTx); err != nil {
		return ChallengeResult{}, fmt.Errorf("failed to expire pending duels: %w", err)
	}

	challenger, err := userTx.GetOrCreate(ctx, challengerTwitchID, challengerName)
	if err != nil {
		return ChallengeResult{}, fmt.Errorf("failed to get or create challenger: %w", err)
	}

	opponent, found, err := userTx.GetByUsername(ctx, opponentName)
	if err != nil {
		return ChallengeResult{}, fmt.Errorf("failed to get opponent: %w", err)
	}

	if !found {
		return ChallengeResult{}, ErrOpponentNotFound
	}

	if challenger.ID == opponent.ID {
		return ChallengeResult{}, ErrCannotChallengeSelf
	}

	if _, hasPendingDuels, err := duelTx.FindPendingDuelByChallenger(ctx, challenger.ID); err != nil {
		return ChallengeResult{}, fmt.Errorf("failed to check for a pending duel: %w", err)
	} else if hasPendingDuels {
		return ChallengeResult{}, ErrChallengerHasPendingDuel
	}

	if _, hasPendingDuel, err := duelTx.FindPendingDuelByOpponent(ctx, opponent.ID); err != nil {
		return ChallengeResult{}, fmt.Errorf("failed to check for a pending duel: %w", err)
	} else if hasPendingDuel {
		return ChallengeResult{}, ErrOpponentHasPendingDuel
	}

	challengerBalance, err := walletTx.Balance(ctx, challenger.ID)
	if err != nil {
		return ChallengeResult{}, fmt.Errorf("failed to get challenger balance: %w", err)
	}

	if challengerBalance < stake {
		return ChallengeResult{}, ErrInsufficientFunds
	}

	expiresAt := d.clock.Now().Add(d.expiryDuration)

	duel, err := duelTx.Create(ctx, challenger.ID, opponent.ID, stake, expiresAt)
	if err != nil {
		return ChallengeResult{}, fmt.Errorf("failed to create duel: %w", err)
	}

	_, err = walletTx.Credit(ctx, challenger.ID, -stake)
	if err != nil {
		if errors.Is(err, wallet.ErrInsufficientFunds) {
			return ChallengeResult{}, ErrInsufficientFunds
		}

		return ChallengeResult{}, fmt.Errorf("failed to withdraw stake: %w", err)
	}

	if err := stakeHistoryTx.Record(ctx, duel.ID, challenger.ID, stake, StakeReasonStake); err != nil {
		return ChallengeResult{}, fmt.Errorf("failed to record stake: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return ChallengeResult{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return ChallengeResult{
		ChallengerName: challenger.Username,
		OpponentName:   opponent.Username,
		Stake:          duel.Stake,
		// ExpiresAt:      duel.ExpiresAt,
		// NewBalance:     newBalance,
	}, nil
}

func (d *Duellist) Accept(ctx context.Context, opponentTwitchID, opponentName string) (AcceptResult, error) {
	tx, err := d.txBeginner.BeginTx(ctx, nil)
	if err != nil {
		return AcceptResult{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	userTx := d.userRepository.WithExecutor(tx)
	walletTx := d.walletRepository.WithExecutor(tx)
	duelTx := d.duelRepository.WithExecutor(tx)
	stakeHistoryTx := d.stakeHistoryRepository.WithExecutor(tx)
	mmrTx := d.mmrRepository.WithExecutor(tx)
	mmrHistoryTx := d.mmrHistoryRepository.WithExecutor(tx)

	if err := d.expirePending(ctx, duelTx, stakeHistoryTx, walletTx); err != nil {
		return AcceptResult{}, fmt.Errorf("failed to expire pending duels: %w", err)
	}

	opponent, err := userTx.GetOrCreate(ctx, opponentTwitchID, opponentName)
	if err != nil {
		return AcceptResult{}, fmt.Errorf("failed to get or create opponent: %w", err)
	}

	duel, found, err := duelTx.FindPendingDuelByOpponent(ctx, opponent.ID)
	if err != nil {
		return AcceptResult{}, fmt.Errorf("failed to find pending duel: %w", err)
	}

	if !found {
		return AcceptResult{}, ErrNoIncomingDuel
	}

	challenger, found, err := userTx.GetByID(ctx, duel.ChallengerID)
	if err != nil {
		return AcceptResult{}, fmt.Errorf("failed to get challenger: %w", err)
	}

	if !found {
		return AcceptResult{}, ErrChallengerNotFound
	}

	newOpponentBalance, err := walletTx.Credit(ctx, opponent.ID, -duel.Stake)
	if err != nil {
		if errors.Is(err, wallet.ErrInsufficientFunds) {
			return AcceptResult{}, ErrInsufficientFunds
		}

		return AcceptResult{}, fmt.Errorf("failed to withdraw stake: %w", err)
	}

	if err := stakeHistoryTx.Record(ctx, duel.ID, opponent.ID, duel.Stake, StakeReasonStake); err != nil {
		return AcceptResult{}, fmt.Errorf("failed to record stake: %w", err)
	}

	challengerRoll := d.roll()
	opponentRoll := d.roll()
	result := determineResult(challengerRoll, opponentRoll)

	newChallengerBalance := int64(0)
	pot := 2 * duel.Stake

	switch result {
	case DuelResultChallengerWon:
		newChallengerBalance, err = walletTx.Credit(ctx, duel.ChallengerID, pot)
		if err != nil {
			return AcceptResult{}, fmt.Errorf("failed to pay out challenger: %w", err)
		}

		if err := stakeHistoryTx.Record(ctx, duel.ID, duel.ChallengerID, pot, StakeReasonPayout); err != nil {
			return AcceptResult{}, fmt.Errorf("failed to record challenger payout: %w", err)
		}
	case DuelResultOpponentWon:
		newOpponentBalance, err = walletTx.Credit(ctx, duel.OpponentID, pot)
		if err != nil {
			return AcceptResult{}, fmt.Errorf("failed to pay out opponent: %w", err)
		}

		if err := stakeHistoryTx.Record(ctx, duel.ID, duel.OpponentID, pot, StakeReasonPayout); err != nil {
			return AcceptResult{}, fmt.Errorf("failed to record opponent payout: %w", err)
		}

		newChallengerBalance, err = walletTx.Balance(ctx, duel.ChallengerID)
		if err != nil {
			return AcceptResult{}, fmt.Errorf("failed to get challenger balance: %w", err)
		}
	case DuelResultDraw:
		newChallengerBalance, err = walletTx.Credit(ctx, duel.ChallengerID, duel.Stake)
		if err != nil {
			return AcceptResult{}, fmt.Errorf("failed to refund challenger: %w", err)
		}

		if err := stakeHistoryTx.Record(ctx, duel.ID, duel.ChallengerID, duel.Stake, StakeReasonRefund); err != nil {
			return AcceptResult{}, fmt.Errorf("failed to record challenger refund: %w", err)
		}

		newOpponentBalance, err = walletTx.Credit(ctx, duel.OpponentID, duel.Stake)
		if err != nil {
			return AcceptResult{}, fmt.Errorf("failed to refund opponent: %w", err)
		}

		if err := stakeHistoryTx.Record(ctx, duel.ID, duel.OpponentID, duel.Stake, StakeReasonRefund); err != nil {
			return AcceptResult{}, fmt.Errorf("failed to record opponent refund: %w", err)
		}
	}

	challengerDelta, opponentDelta, err := d.applyRatingChange(ctx, mmrTx, mmrHistoryTx, duel.ID, duel.ChallengerID, opponent.ID, result)
	if err != nil {
		return AcceptResult{}, fmt.Errorf("failed to apply rating change: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return AcceptResult{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return AcceptResult{
		ChallengerName:        challenger.Username,
		OpponentName:          opponent.Username,
		Stake:                 duel.Stake,
		Pot:                   pot,
		ChallengerRoll:        challengerRoll,
		OpponentRoll:          opponentRoll,
		Result:                result,
		NewChallengerBalance:  newChallengerBalance,
		NewOpponentBalance:    newOpponentBalance,
		ChallengerRatingDelta: challengerDelta,
		OpponentRatingDelta:   opponentDelta,
	}, nil
}

func (d *Duellist) Decline(ctx context.Context, opponentTwitchID, opponentName string) (DeclineResult, error) {
	tx, err := d.txBeginner.BeginTx(ctx, nil)
	if err != nil {
		return DeclineResult{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	userTx := d.userRepository.WithExecutor(tx)
	walletTx := d.walletRepository.WithExecutor(tx)
	duelTx := d.duelRepository.WithExecutor(tx)
	stakeHistoryTx := d.stakeHistoryRepository.WithExecutor(tx)

	if err := d.expirePending(ctx, duelTx, stakeHistoryTx, walletTx); err != nil {
		return DeclineResult{}, fmt.Errorf("failed to expire pending duels: %w", err)
	}

	opponent, err := userTx.GetOrCreate(ctx, opponentTwitchID, opponentName)
	if err != nil {
		return DeclineResult{}, fmt.Errorf("failed to get or create opponent: %w", err)
	}

	duel, found, err := duelTx.FindPendingDuelByOpponent(ctx, opponent.ID)
	if err != nil {
		return DeclineResult{}, fmt.Errorf("failed to find pending duel: %w", err)
	}

	if !found {
		return DeclineResult{}, ErrNoIncomingDuel
	}

	challenger, found, err := userTx.GetByID(ctx, duel.ChallengerID)
	if err != nil {
		return DeclineResult{}, fmt.Errorf("failed to get challenger: %w", err)
	}

	if !found {
		return DeclineResult{}, fmt.Errorf("failed to get challenger")
	}

	if err := duelTx.SetStatus(ctx, duel.ID, DuelStatusDeclined); err != nil {
		return DeclineResult{}, fmt.Errorf("failed to decline duel: %w", err)
	}

	if _, err := walletTx.Credit(ctx, duel.ChallengerID, duel.Stake); err != nil {
		return DeclineResult{}, fmt.Errorf("failed to refund challenger: %w", err)
	}

	if err := stakeHistoryTx.Record(ctx, duel.ID, duel.ChallengerID, duel.Stake, StakeReasonRefund); err != nil {
		return DeclineResult{}, fmt.Errorf("failed to record challenger refund: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return DeclineResult{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return DeclineResult{
		ChallengerName: challenger.Username,
		OpponentName:   opponent.Username,
	}, nil
}

func (d *Duellist) Cancel(ctx context.Context, challengerTwitchID, challengerName string) (CancelResult, error) {
	tx, err := d.txBeginner.BeginTx(ctx, nil)
	if err != nil {
		return CancelResult{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	userTx := d.userRepository.WithExecutor(tx)
	walletTx := d.walletRepository.WithExecutor(tx)
	duelTx := d.duelRepository.WithExecutor(tx)
	stakeHistoryTx := d.stakeHistoryRepository.WithExecutor(tx)

	if err := d.expirePending(ctx, duelTx, stakeHistoryTx, walletTx); err != nil {
		return CancelResult{}, fmt.Errorf("failed to expire pending duels: %w", err)
	}

	challenger, err := userTx.GetOrCreate(ctx, challengerTwitchID, challengerName)
	if err != nil {
		return CancelResult{}, fmt.Errorf("failed to get or create challenger: %w", err)
	}

	duel, found, err := duelTx.FindPendingDuelByChallenger(ctx, challenger.ID)
	if err != nil {
		return CancelResult{}, fmt.Errorf("failed to find pending duel: %w", err)
	}

	if !found {
		return CancelResult{}, ErrNoOutgoingDuel
	}

	opponent, found, err := userTx.GetByID(ctx, duel.OpponentID)

	if !found {
		return CancelResult{}, fmt.Errorf("failed to get opponent")
	}

	if err := duelTx.SetStatus(ctx, duel.ID, DuelStatusCancelled); err != nil {
		return CancelResult{}, fmt.Errorf("failed to cancel duel: %w", err)
	}

	if _, err := walletTx.Credit(ctx, challenger.ID, duel.Stake); err != nil {
		return CancelResult{}, fmt.Errorf("failed to refund challenger: %w", err)
	}

	if err := stakeHistoryTx.Record(ctx, duel.ID, challenger.ID, duel.Stake, StakeReasonRefund); err != nil {
		return CancelResult{}, fmt.Errorf("failed to record challenger refund: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return CancelResult{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return CancelResult{
		ChallengerName: challenger.Username,
		OpponentName:   opponent.Username,
	}, nil
}

func (d *Duellist) expirePending(ctx context.Context, duelTx DuelRepository, stakeHistoryTx StakeHistoryRepository, walletTx wallet.Repository) error {
	expired, err := duelTx.ExpirePending(ctx, d.clock.Now())
	if err != nil {
		return fmt.Errorf("failed to expire pending duels: %w", err)
	}

	for _, e := range expired {
		if _, err := walletTx.Credit(ctx, e.ChallengerID, e.Stake); err != nil {
			return fmt.Errorf("failed to refund expired duel stake: %w", err)
		}

		if err := stakeHistoryTx.Record(ctx, e.ID, e.ChallengerID, e.Stake, StakeReasonRefund); err != nil {
			return fmt.Errorf("failed to record expired duel refund: %w", err)
		}
	}

	return nil
}

func (d *Duellist) applyRatingChange(
	ctx context.Context,
	mmrTx MMRRepository,
	mmrHistoryTx MMRHistoryRepository,
	duelID, challengerID, opponentID int64,
	result DuelResult,
) (int, int, error) {
	challengerRating, err := mmrTx.GetOrCreate(ctx, challengerID, d.defaultRating)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get challenger MMR: %w", err)
	}

	opponentRating, err := mmrTx.GetOrCreate(ctx, opponentID, d.defaultRating)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get opponent MMR: %w", err)
	}

	challengerScore, opponentScore := scoresFor(result)

	newChallengerRating := clampRating(challengerRating + d.ratingCalculator.Delta(challengerRating, opponentRating, challengerScore))
	newOpponentRating := clampRating(opponentRating + d.ratingCalculator.Delta(opponentRating, challengerRating, opponentScore))

	if err := mmrTx.Set(ctx, challengerID, newChallengerRating); err != nil {
		return 0, 0, fmt.Errorf("failed to set challenger MMR: %w", err)
	}

	if err := mmrTx.Set(ctx, opponentID, newOpponentRating); err != nil {
		return 0, 0, fmt.Errorf("failed to set opponent MMR: %w", err)
	}

	challengerDelta := newChallengerRating - challengerRating
	opponentDelta := newOpponentRating - opponentRating

	if err := mmrHistoryTx.Record(ctx, duelID, challengerID, challengerDelta); err != nil {
		return 0, 0, fmt.Errorf("failed to record challenger MMR change: %w", err)
	}

	if err := mmrHistoryTx.Record(ctx, duelID, opponentID, opponentDelta); err != nil {
		return 0, 0, fmt.Errorf("failed to record opponent MMR change: %w", err)
	}

	return challengerDelta, opponentDelta, nil
}

func (d *Duellist) roll() int {
	return int(1 + d.rng.Int64N(20))
}

func determineResult(challengerRoll, opponentRoll int) DuelResult {
	switch {
	case challengerRoll > opponentRoll:
		return DuelResultChallengerWon
	case challengerRoll < opponentRoll:
		return DuelResultOpponentWon
	default:
		return DuelResultDraw
	}
}
