package commands

import (
	"context"
	"errors"
	"fmt"

	"jijabot/internal/logger"
	"jijabot/internal/reward"
)

type DailyClaimer interface {
	Amount() int64
	Claim(ctx context.Context, twitchUserID, username string) (int64, error)
}

type DailyCommand struct {
	claimer DailyClaimer
	log     logger.Logger
}

func NewDailyCommand(claimer DailyClaimer, log logger.Logger) *DailyCommand {
	return &DailyCommand{
		claimer: claimer,
		log:     log,
	}
}

func (c *DailyCommand) Name() string {
	return "!daily"
}

func (c *DailyCommand) Execute(ctx context.Context, p Payload, r Responder) error {
	if p.UserID == "" {
		c.log.ErrorContext(ctx, "daily claim attempted without a twitch user id", "user_id", p.UserID)
		return r.Say(fmt.Sprintf("@%s, что-то пошло не так, попробуй позже.", p.User))
	}

	balance, err := c.claimer.Claim(ctx, p.UserID, p.User)
	if errors.Is(err, reward.ErrAlreadyClaimed) {
		return r.Say(fmt.Sprintf("@%s, на сегодня жижакоины уже получены! Заходи после 12:00 UTC+3.", p.User))
	}

	if err != nil {
		c.log.ErrorContext(ctx, "failed to claim daily reward", "user_id", p.UserID, "error", err)
		return r.Say(fmt.Sprintf("@%s, что-то пошло не так, попробуй позже.", p.User))
	}

	return r.Say(fmt.Sprintf("@%s, получено +%d жижакоинов! Твой баланс: %d.", p.User, c.claimer.Amount(), balance))
}
