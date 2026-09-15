package commands

import (
	"context"
	"errors"

	"jijabot/internal/logger"
	"jijabot/internal/reward"
)

const (
	phrasebookDaily                = "daily"
	phrasebookDailyInternalError   = "internal_error"
	phrasebookDailyAlreadyRedeemed = "already_redeemed"
	phrasebookDailySuccess         = "success"
)

type dailyErrorData struct {
	User string
}

type dailySuccessData struct {
	User    string
	Amount  int64
	Balance int64
}

type DailyRedeemer interface {
	Amount() int64
	Redeem(ctx context.Context, twitchUserID, username string) (int64, error)
}

type DailyCommand struct {
	redeemer     DailyRedeemer
	phrasePicker PhrasePicker
	log          logger.Logger
}

func NewDailyCommand(redeemer DailyRedeemer, phrasePicker PhrasePicker, log logger.Logger) *DailyCommand {
	return &DailyCommand{
		redeemer:     redeemer,
		phrasePicker: phrasePicker,
		log:          log.With("command", "!bet"),
	}
}

func (c *DailyCommand) Name() string {
	return "!daily"
}

func (c *DailyCommand) Execute(ctx context.Context, p Payload, r Responder) error {
	if p.UserID == "" {
		c.log.ErrorContext(ctx, "daily command attempted without a twitch user id", "user_id", p.UserID)
		return r.Say(c.pickError(ctx, phrasebookDailyInternalError, p))
	}

	balance, err := c.redeemer.Redeem(ctx, p.UserID, p.User)
	if errors.Is(err, reward.ErrAlreadyRedeemed) {
		return r.Say(c.pickError(ctx, phrasebookDailyAlreadyRedeemed, p))
	}

	if err != nil {
		c.log.ErrorContext(ctx, "failed to redeem daily reward", "user_id", p.UserID, "error", err)
		return r.Say(c.pickError(ctx, phrasebookDailyInternalError, p))
	}

	data := dailySuccessData{
		User:    p.User,
		Amount:  c.redeemer.Amount(),
		Balance: balance,
	}

	return r.Say(pickPhraseOrFallback(ctx, c.log, c.phrasePicker, phrasebookDaily, phrasebookDailySuccess, p.User, data))
}

func (c *DailyCommand) pickError(ctx context.Context, scenario string, p Payload) string {
	return pickPhraseOrFallback(ctx, c.log, c.phrasePicker, phrasebookDaily, scenario, p.User, dailyErrorData{User: p.User})
}
