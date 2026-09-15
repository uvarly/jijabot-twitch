package commands

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"jijabot/internal/duelling"
	"jijabot/internal/logger"
)

const (
	phrasebookDuel                          = "duel"
	phrasebookDuelInternalError             = "internal_error"
	phrasebookDuelMissingArgs               = "missing_args"
	phrasebookDuelInvalidAmountNotInteger   = "invalid_amount_not_integer"
	phrasebookDuelInvalidAmountLessThanZero = "invalid_amount_less_than_zero"
	phrasebookDuelTargetUserNotFound        = "target_user_not_found"
	phrasebookDuelCannotTargetSelf          = "cannot_target_self"
	phrasebookDuelUserHasPendingDuel        = "user_has_pending_duel"
	phrasebookDuelTargetHasPendingDuel      = "target_has_pending_duel"
	phrasebookDuelInsufficientFunds         = "insufficient_funds"
	phrasebookDuelDailyLimitReached         = "daily_limit_reached"
	phrasebookDuelSuccess                   = "success"
)

type duelErrorData struct {
	User string
}

type duelSuccessData struct {
	ChallengerUser string
	OpponentUser   string
	Amount         int64
}

type Duellist interface {
	Challenge(ctx context.Context, challengerTwitchID, challengerName, opponentName string, stake int64) (duelling.ChallengeResult, error)
	Accept(ctx context.Context, opponentTwitchID, opponentName string) (duelling.AcceptResult, error)
	Decline(ctx context.Context, opponentTwitchID, opponentName string) (duelling.DeclineResult, error)
	Cancel(ctx context.Context, challengerTwitchID, challengerName string) (duelling.CancelResult, error)
}

type DuelCommand struct {
	duellist     Duellist
	phrasePicker PhrasePicker
	log          logger.Logger
}

func NewDuelCommand(duellist Duellist, phrasePicker PhrasePicker, log logger.Logger) *DuelCommand {
	return &DuelCommand{
		duellist:     duellist,
		phrasePicker: phrasePicker,
		log:          log.With("command", "!duel"),
	}
}

func (c *DuelCommand) Name() string {
	return "!duel"
}

func (c *DuelCommand) Execute(ctx context.Context, p Payload, r Responder) error {
	if p.UserID == "" {
		c.log.ErrorContext(ctx, "duel command attempted without a twitch user id", "user_id", p.UserID)
		return r.Say(c.pickError(ctx, phrasebookDuelInternalError, p))
	}

	if len(p.Args) != 2 {
		return r.Say(c.pickError(ctx, phrasebookDuelMissingArgs, p))
	}

	targetUser := strings.TrimPrefix(p.Args[0], "@")
	if targetUser == "" {
		return r.Say(c.pickError(ctx, phrasebookDuelMissingArgs, p))
	}

	amount, err := strconv.ParseInt(p.Args[0], 10, 64)
	if err != nil {
		return r.Say(c.pickError(ctx, phrasebookDuelInvalidAmountNotInteger, p))
	}

	result, err := c.duellist.Challenge(ctx, p.UserID, p.User, targetUser, amount)

	switch {
	case errors.Is(err, duelling.ErrCannotChallengeSelf):
		return r.Say(c.pickError(ctx, phrasebookDuelCannotTargetSelf, p))
	case errors.Is(err, duelling.ErrChallengerHasPendingDuel):
		return r.Say(c.pickError(ctx, phrasebookDuelUserHasPendingDuel, p))
	case errors.Is(err, duelling.ErrDailyLimitReached):
		return r.Say(c.pickError(ctx, phrasebookDuelDailyLimitReached, p))
	case errors.Is(err, duelling.ErrInsufficientFunds):
		return r.Say(c.pickError(ctx, phrasebookDuelInsufficientFunds, p))
	case errors.Is(err, duelling.ErrInvalidStake):
		return r.Say(c.pickError(ctx, phrasebookDuelInvalidAmountLessThanZero, p))
	case errors.Is(err, duelling.ErrOpponentNotFound):
		return r.Say(c.pickError(ctx, phrasebookDuelTargetUserNotFound, p))
	case errors.Is(err, duelling.ErrOpponentHasPendingDuel):
		return r.Say(c.pickError(ctx, phrasebookDuelTargetHasPendingDuel, p))
	case err != nil:
		c.log.ErrorContext(ctx, "failed to challenge user", "user", p.User, "target_user", targetUser, "amount", amount, "error", err)
		return r.Say(c.pickError(ctx, phrasebookDuelInternalError, p))
	}

	data := duelSuccessData{
		ChallengerUser: result.ChallengerName,
		OpponentUser:   result.OpponentName,
		Amount:         amount,
	}

	return r.Say(pickPhraseOrFallback(ctx, c.log, c.phrasePicker, phrasebookDuel, phrasebookDuelSuccess, p.User, data))
}

func (c *DuelCommand) pickError(ctx context.Context, scenario string, p Payload) string {
	return pickPhraseOrFallback(ctx, c.log, c.phrasePicker, phrasebookDuel, scenario, p.User, duelErrorData{User: p.User})
}
