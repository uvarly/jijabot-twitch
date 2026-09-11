package commands

import (
	"context"
	"errors"

	"jijabot/internal/logger"
	"jijabot/internal/reward"
)

const (
	phrasebookDaily               = "daily"
	phrasebookDailyInternalError  = "internal_error"
	phrasebookDailyAlreadyClaimed = "already_claimed"
	phrasebookDailySuccess        = "success"
)

type dailyErrorData struct {
	User string
}

type dailySuccessData struct {
	User    string
	Amount  int64
	Balance int64
}

type DailyClaimer interface {
	Amount() int64
	Claim(ctx context.Context, twitchUserID, username string) (int64, error)
}

type DailyCommand struct {
	claimer      DailyClaimer
	phrasePicker PhrasePicker
	log          logger.Logger
}

func NewDailyCommand(claimer DailyClaimer, phrasePicker PhrasePicker, log logger.Logger) *DailyCommand {
	return &DailyCommand{
		claimer:      claimer,
		phrasePicker: phrasePicker,
		log:          log.With("command", "!bet"),
	}
}

func (c *DailyCommand) Name() string {
	return "!daily"
}

func (c *DailyCommand) Execute(ctx context.Context, p Payload, r Responder) error {
	if p.UserID == "" {
		c.log.ErrorContext(ctx, "daily claim attempted without a twitch user id", "user_id", p.UserID)
		return r.Say(c.pickError(ctx, phrasebookDailyInternalError, p))
	}

	balance, err := c.claimer.Claim(ctx, p.UserID, p.User)
	if errors.Is(err, reward.ErrAlreadyClaimed) {
		return r.Say(c.pickError(ctx, phrasebookDailyAlreadyClaimed, p))
	}

	if err != nil {
		c.log.ErrorContext(ctx, "failed to claim daily reward", "user_id", p.UserID, "error", err)
		return r.Say(c.pickError(ctx, phrasebookDailyInternalError, p))
	}

	data := dailySuccessData{
		User:    p.User,
		Amount:  c.claimer.Amount(),
		Balance: balance,
	}

	return r.Say(pickPhraseOrFallback(ctx, c.log, c.phrasePicker, phrasebookDaily, phrasebookDailySuccess, p.User, data))
}

func (c *DailyCommand) pickError(ctx context.Context, scenario string, p Payload) string {
	return pickPhraseOrFallback(ctx, c.log, c.phrasePicker, phrasebookDaily, scenario, p.User, dailyErrorData{User: p.User})
}
