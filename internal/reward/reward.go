package reward

import (
	"context"
	"errors"
	"fmt"
	"jijabot/internal/redeem"
	"jijabot/internal/store"
	"jijabot/internal/users"
	"jijabot/internal/wallet"
	"time"
)

var ErrAlreadyClaimed = redeem.ErrAlreadyClaimed

type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (rc realClock) Now() time.Time { return time.Now() }

type Option func(*DailyClaimer)

func WithClock(clock Clock) Option {
	return func(dc *DailyClaimer) { dc.clock = clock }
}

type DailyClaimer struct {
	txBeginner store.TxBeginner
	user       users.Repository
	redeem     redeem.Repository
	wallet     wallet.Repository
	clock      Clock
	amount     int64
	resetHour  int
}

func NewDailyClaimer(
	txBeginner store.TxBeginner,
	userRepository users.Repository,
	redeemRepository redeem.Repository,
	walletRepository wallet.Repository,
	amount int64,
	resetHour int,
	options ...Option,
) *DailyClaimer {
	dailyClaimer := &DailyClaimer{
		txBeginner: txBeginner,
		user:       userRepository,
		redeem:     redeemRepository,
		wallet:     walletRepository,
		clock:      realClock{},
		amount:     amount,
		resetHour:  resetHour,
	}

	for _, o := range options {
		o(dailyClaimer)
	}

	return dailyClaimer
}

func (dc *DailyClaimer) Amount() int64 { return dc.amount }

func (dc *DailyClaimer) Claim(ctx context.Context, twitchUserID, username string) (int64, error) {
	tx, err := dc.txBeginner.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	txUser := dc.user.WithExecutor(tx)
	txRedeem := dc.redeem.WithExecutor(tx)
	txWallet := dc.wallet.WithExecutor(tx)

	user, err := txUser.GetOrCreate(ctx, twitchUserID, username)
	if err != nil {
		return 0, fmt.Errorf("failed to get or create user: %w", err)
	}

	period := redeem.PeriodKey(dc.clock.Now(), dc.resetHour)

	if err := txRedeem.Claim(ctx, user.ID, dc.amount, period); err != nil {
		if errors.Is(err, redeem.ErrAlreadyClaimed) {
			return 0, ErrAlreadyClaimed
		}

		return 0, fmt.Errorf("failed to record claim: %w", err)
	}

	balance, err := txWallet.Credit(ctx, user.ID, dc.amount)
	if err != nil {
		return 0, fmt.Errorf("failed to credit wallet: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return balance, nil
}
