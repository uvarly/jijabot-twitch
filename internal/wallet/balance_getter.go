package wallet

import (
	"context"
	"fmt"

	"jijabot/internal/store"
	"jijabot/internal/users"
)

type BalanceGetter struct {
	txBeginner store.TxBeginner
	users      users.Repository
	wallet     Repository
}

func NewBalanceGetter(
	txBeginner store.TxBeginner,
	userRepository users.Repository,
	walletRepository Repository,
) *BalanceGetter {
	return &BalanceGetter{
		txBeginner: txBeginner,
		users:      userRepository,
		wallet:     walletRepository,
	}
}

func (bg *BalanceGetter) Get(ctx context.Context, twitchUserID, username string) (int64, error) {
	tx, err := bg.txBeginner.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	userTx := bg.users.WithExecutor(tx)
	walletTx := bg.wallet.WithExecutor(tx)

	user, err := userTx.GetOrCreate(ctx, twitchUserID, username)
	if err != nil {
		return 0, fmt.Errorf("failed to get or create user: %w", err)
	}

	balance, err := walletTx.Balance(ctx, user.ID)
	if err != nil {
		return 0, fmt.Errorf("failed to get balance: %w", err)
	}

	return balance, nil
}
