package grant

import (
	"context"
	"errors"
	"fmt"
	"jijabot/internal/store"
	"jijabot/internal/users"
	"jijabot/internal/wallet"
)

var (
	ErrInvalidAmount = errors.New("grant: amount must be positive")
	ErrUserNotFound  = errors.New("grant: user not found")
)

type Result struct {
	TargetUser string
	NewBalance int64
}

type Granter struct {
	txBeginner store.TxBeginner
	users      users.Repository
	wallet     wallet.Repository
}

func NewGranter(
	txBeginner store.TxBeginner,
	userRepository users.Repository,
	walletRepository wallet.Repository,
) *Granter {
	return &Granter{
		txBeginner: txBeginner,
		users:      userRepository,
		wallet:     walletRepository,
	}
}

func (g *Granter) Grant(ctx context.Context, username string, amount int64) (Result, error) {
	if amount <= 0 {
		return Result{}, ErrInvalidAmount
	}

	tx, err := g.txBeginner.BeginTx(ctx, nil)
	if err != nil {
		return Result{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	userTx := g.users.WithExecutor(tx)
	walletTx := g.wallet.WithExecutor(tx)

	user, found, err := userTx.GetByUsername(ctx, username)
	if err != nil {
		return Result{}, fmt.Errorf("failed to get user: %w", err)
	}

	if !found {
		return Result{}, ErrUserNotFound
	}

	newBalance, err := walletTx.Credit(ctx, user.ID, amount)
	if err != nil {
		return Result{}, fmt.Errorf("failed to credit wallet: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return Result{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return Result{
		TargetUser: username,
		NewBalance: newBalance,
	}, nil
}
