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
	grant      GrantRepository
	wallet     wallet.Repository
}

func NewGranter(
	txBeginner store.TxBeginner,
	userRepository users.Repository,
	grantRepository GrantRepository,
	walletRepository wallet.Repository,
) *Granter {
	return &Granter{
		txBeginner: txBeginner,
		users:      userRepository,
		grant:      grantRepository,
		wallet:     walletRepository,
	}
}

func (g *Granter) Grant(ctx context.Context, granterName, granteeName string, amount int64) (Result, error) {
	if amount <= 0 {
		return Result{}, ErrInvalidAmount
	}

	tx, err := g.txBeginner.BeginTx(ctx, nil)
	if err != nil {
		return Result{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	userTx := g.users.WithExecutor(tx)
	grantTx := g.grant.WithExecutor(tx)
	walletTx := g.wallet.WithExecutor(tx)

	granter, found, err := userTx.GetByUsername(ctx, granterName)
	if err != nil {
		return Result{}, fmt.Errorf("failed to get granter: %w", err)
	}

	if !found {
		return Result{}, ErrUserNotFound
	}

	grantee, found, err := userTx.GetByUsername(ctx, granteeName)
	if err != nil {
		return Result{}, fmt.Errorf("failed to get grantee: %w", err)
	}

	if !found {
		return Result{}, ErrUserNotFound
	}

	if err := grantTx.Record(ctx, granter.ID, grantee.ID, amount); err != nil {
		return Result{}, fmt.Errorf("failed to record grant: %w", err)
	}

	newBalance, err := walletTx.Credit(ctx, grantee.ID, amount)
	if err != nil {
		return Result{}, fmt.Errorf("failed to credit wallet: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return Result{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return Result{
		TargetUser: grantee.Username,
		NewBalance: newBalance,
	}, nil
}
