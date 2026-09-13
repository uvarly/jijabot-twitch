package balance

import (
	"context"

	"jijabot/internal/store"
	"jijabot/internal/users"
	"jijabot/internal/wallet"
)

type BalanceGetter struct {
	txBeginner store.TxBeginner
	users      users.Repository
	wallet     wallet.Repository
}

func NewBalanceGetter(
	txBeginner store.TxBeginner,
	userRepository users.Repository,
	walletRepository wallet.Repository,
) *BalanceGetter {
	return &BalanceGetter{
		txBeginner: txBeginner,
		users:      userRepository,
		wallet:     walletRepository,
	}
}

func (bg *BalanceGetter) Get(ctx context.Context, twitchUserID string) (int64, error) {
	return 0, nil
}
