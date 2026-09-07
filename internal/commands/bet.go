package commands

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"jijabot/internal/gambling"
	"jijabot/internal/logger"
)

type BetPlacer interface {
	Place(ctx context.Context, twitchUserID, username string, amount int64) (gambling.Result, error)
}

type BetCommand struct {
	betPlacer BetPlacer
	log       logger.Logger
}

func (c *BetCommand) Name() string {
	return "!bet"
}

func NewBetCommand(betPlacer BetPlacer, log logger.Logger) *BetCommand {
	return &BetCommand{
		betPlacer: betPlacer,
		log:       log.With("command", "!bet"),
	}
}

func (c *BetCommand) Execute(ctx context.Context, p Payload, r Responder) error {
	if p.UserID == "" {
		c.log.ErrorContext(ctx, "bet attempted without a twitch user id", "user_id", p.UserID)
		return r.Say(fmt.Sprintf("@%s, что-то пошло не так, попробуй позже.", p.User))
	}

	if len(p.Args) != 1 {
		return r.Say(fmt.Sprintf("@%s, укажи ставку: !bet <кол-во>.", p.User))
	}

	amount, err := strconv.ParseInt(p.Args[0], 10, 64)
	if err != nil {
		return r.Say(fmt.Sprintf("@%s, ставка должна быть числом. Пример: !bet 20.", p.User))
	}

	result, err := c.betPlacer.Place(ctx, p.UserID, p.User, amount)
	switch {
	case errors.Is(err, gambling.ErrDailyLimitReached):
		return r.Say(fmt.Sprintf("@%s, на сегодня лимит ставок исчерпан.", p.User))
	case errors.Is(err, gambling.ErrInsufficientFunds):
		return r.Say(fmt.Sprintf("@%s, у тебя недостаточно жижа-коинов для такой ставки.", p.User))
	case errors.Is(err, gambling.ErrInvalidAmount):
		return r.Say(fmt.Sprintf("@%s, ставка должна быть больше нуля. Пример: !bet 20.", p.User))
	case err != nil:
		c.log.ErrorContext(ctx, "failed to place bet", "user", p.User, "amount", amount, "error", err)
		return r.Say(fmt.Sprintf("@%s, что-то пошло не так, попробуй позже.", p.User))
	}

	if result.Won {
		return r.Say(fmt.Sprintf("@%s, поздравляю! Твой выигрыш: %d жижа-коинов! Баланс: %d.", p.User, result.Payout, result.NewBalance))
	}

	return r.Say(fmt.Sprintf("@%s, %d жижа-коинов улетели в трубу. Баланс: %d.", p.User, -result.Payout, result.NewBalance))
}
