package commands

import (
	"context"
	"errors"
	"strconv"

	"jijabot/internal/gambling"
	"jijabot/internal/logger"
)

const (
	phrasebookBet                          = "bet"
	phrasebookBetInternalError             = "internal_error"
	phrasebookBetInvalidAmountMissingArgs  = "invalid_amount_missing_args"
	phrasebookBetInvalidAmountNotInteger   = "invalid_amount_not_integer"
	phrasebookBetInvalidAmountLessThanZero = "invalid_amount_less_than_zero"
	phrasebookBetInsufficientFunds         = "insufficient_funds"
	phrasebookBetDailyLimitReached         = "daily_limit_reached"
	phrasebookBetWin                       = "win"
	phrasebookBetLose                      = "lose"
)

type betErrorData struct {
	User string
}

type betResultData struct {
	User       string
	Payout     int64
	NewBalance int64
}

type BetPlacer interface {
	Place(ctx context.Context, twitchUserID, username string, amount int64) (gambling.Result, error)
}

type BetCommand struct {
	betPlacer    BetPlacer
	phrasePicker PhrasePicker
	log          logger.Logger
}

func (c *BetCommand) Name() string {
	return "!bet"
}

func NewBetCommand(betPlacer BetPlacer, phrasePicker PhrasePicker, log logger.Logger) *BetCommand {
	return &BetCommand{
		betPlacer:    betPlacer,
		phrasePicker: phrasePicker,
		log:          log.With("command", "!bet"),
	}
}

func (c *BetCommand) Execute(ctx context.Context, p Payload, r Responder) error {
	if p.UserID == "" {
		c.log.ErrorContext(ctx, "bet attempted without a twitch user id", "user_id", p.UserID)
		return r.Say(c.pickError(ctx, phrasebookBetInternalError, p))
	}

	if len(p.Args) != 1 {
		return r.Say(c.pickError(ctx, phrasebookBetInvalidAmountMissingArgs, p))
	}

	amount, err := strconv.ParseInt(p.Args[0], 10, 64)
	if err != nil {
		return r.Say(c.pickError(ctx, phrasebookBetInvalidAmountNotInteger, p))
	}

	result, err := c.betPlacer.Place(ctx, p.UserID, p.User, amount)
	switch {
	case errors.Is(err, gambling.ErrDailyLimitReached):
		return r.Say(c.pickError(ctx, phrasebookBetDailyLimitReached, p))
	case errors.Is(err, gambling.ErrInsufficientFunds):
		return r.Say(c.pickError(ctx, phrasebookBetInsufficientFunds, p))
	case errors.Is(err, gambling.ErrInvalidAmount):
		return r.Say(c.pickError(ctx, phrasebookBetInvalidAmountLessThanZero, p))
	case err != nil:
		c.log.ErrorContext(ctx, "failed to place bet", "user", p.User, "amount", amount, "error", err)
		return r.Say(c.pickError(ctx, phrasebookBetInternalError, p))
	}

	if result.Won {
		data := betResultData{
			User:       p.User,
			Payout:     result.Payout,
			NewBalance: result.NewBalance,
		}

		return r.Say(pickPhraseOrFallback(ctx, c.log, c.phrasePicker, phrasebookBet, phrasebookBetWin, p.User, data))
	}

	data := betResultData{
		User:       p.User,
		Payout:     -result.Payout,
		NewBalance: result.NewBalance,
	}

	return r.Say(pickPhraseOrFallback(ctx, c.log, c.phrasePicker, phrasebookBet, phrasebookBetLose, p.User, data))
}

func (c *BetCommand) pickError(ctx context.Context, scenario string, p Payload) string {
	return pickPhraseOrFallback(ctx, c.log, c.phrasePicker, phrasebookBet, scenario, p.User, betErrorData{User: p.User})
}
