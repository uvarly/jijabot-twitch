package reward

import (
	"context"
	"errors"
	"fmt"
	"time"

	"jijabot/internal/dailywindow"
	"jijabot/internal/store"
	"jijabot/internal/users"
	"jijabot/internal/wallet"
)

type Clock interface {
	Now() time.Time
}

type RealClock struct{}

func (RealClock) Now() time.Time { return time.Now() }

type Option func(*DailyRedeemer)

func WithClock(clock Clock) Option {
	return func(dc *DailyRedeemer) { dc.clock = clock }
}

type DailyRedeemer struct {
	txBeginner store.TxBeginner
	users      users.Repository
	redeem     RedeemRepository
	wallet     wallet.Repository
	clock      Clock

	amount    int64
	resetHour int
}

func NewDailyRedeemer(
	txBeginner store.TxBeginner,
	userRepository users.Repository,
	redeemRepository RedeemRepository,
	walletRepository wallet.Repository,
	amount int64,
	resetHour int,
	options ...Option,
) *DailyRedeemer {
	dailyRedeemer := &DailyRedeemer{
		txBeginner: txBeginner,
		users:      userRepository,
		redeem:     redeemRepository,
		wallet:     walletRepository,
		clock:      RealClock{},
		amount:     amount,
		resetHour:  resetHour,
	}

	for _, o := range options {
		o(dailyRedeemer)
	}

	return dailyRedeemer
}

func (dc *DailyRedeemer) Amount() int64 { return dc.amount }

func (dc *DailyRedeemer) Redeem(ctx context.Context, twitchUserID, username string) (int64, error) {
	tx, err := dc.txBeginner.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	userTx := dc.users.WithExecutor(tx)
	redeemTx := dc.redeem.WithExecutor(tx)
	walletTx := dc.wallet.WithExecutor(tx)

	user, err := userTx.GetOrCreate(ctx, twitchUserID, username)
	if err != nil {
		return 0, fmt.Errorf("failed to get or create user: %w", err)
	}

	period := dailywindow.Key(dc.clock.Now(), dc.resetHour)

	if err := redeemTx.Redeem(ctx, user.ID, dc.amount, period); err != nil {
		if errors.Is(err, ErrAlreadyRedeemed) {
			return 0, ErrAlreadyRedeemed
		}

		return 0, fmt.Errorf("failed to record redeem: %w", err)
	}

	balance, err := walletTx.Credit(ctx, user.ID, dc.amount)
	if err != nil {
		return 0, fmt.Errorf("failed to credit wallet: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return balance, nil
}
