package commands

import (
	"context"
	"jijabot/internal/logger"
)

const (
	phrasebookBalance              = "balance"
	phrasebookBalanceInternalError = "internal_error"
)

type balanceData struct {
	User    string
	Balance int64
}

type BalanceGetter interface {
	Get(ctx context.Context, twitchUserID string) (int64, error)
}

type BalanceCommand struct {
	balanceGetter BalanceGetter
	phrasePicker  PhrasePicker
	log           logger.Logger
}

func NewBalanceCommand(balanceGetter BalanceGetter, phrasePicker PhrasePicker, log logger.Logger) *BalanceCommand {
	return &BalanceCommand{
		balanceGetter: balanceGetter,
		phrasePicker:  phrasePicker,
		log:           log.With("command", "!balance"),
	}
}

func (c *BalanceCommand) Name() string {
	return "!balance"
}

func (c *BalanceCommand) Execute(ctx context.Context, p Payload, r Responder) error {
	if p.UserID == "" {
		c.log.ErrorContext(ctx, "balance attempted without a twitch user id", "user_id", p.UserID)
		return r.Say(c.pickError(ctx, phrasebookBalanceInternalError, p))
	}

	// balance, err := c.balanceGetter.Get(ctx, p.UserID)
	// if err != nil {
	// 	c.log.ErrorContext(ctx, "failed to get balance", "user_id", p.UserID, "error", err)
	// 	return r.Say(c.pickError(ctx, phrasebookBalanceInternalError, p))
	// }

	// // data := balanceData{
	// // 	User:    p.User,
	// // 	Balance: balance,
	// // }

	return nil
}

func (c *BalanceCommand) pickError(ctx context.Context, scenario string, p Payload) string {
	return pickPhraseOrFallback(ctx, c.log, c.phrasePicker, phrasebookBalance, scenario, p.User, balanceData{User: p.User})
}
