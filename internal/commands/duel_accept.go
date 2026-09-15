package commands

import (
	"context"
	"errors"

	"jijabot/internal/duelling"
	"jijabot/internal/logger"
)

const (
	phrasebookDuelAccept                  = "duel_accept"
	phrasebookDuelAcceptInternalError     = "internal_error"
	phrasebookDuelAcceptNoIncomingDuel    = "no_incoming_duel"
	phrasebookDuelAcceptInsufficientFunds = "insufficient_funds"
	phrasebookDuelAcceptSuccess           = "success"
)

type duelAcceptErrorData struct {
	User string
}

type duelAcceptSuccessData struct {
	ChallengerUser string
	OpponentUser   string
}

type DuelAcceptCommand struct {
	duellist     Duellist
	phrasePicker PhrasePicker
	log          logger.Logger
}

func NewDuelAcceptCommand(duellist Duellist, phrasePicker PhrasePicker, log logger.Logger) *DuelAcceptCommand {
	return &DuelAcceptCommand{
		duellist:     duellist,
		phrasePicker: phrasePicker,
		log:          log.With("command", "!duel_accept"),
	}
}

func (c *DuelAcceptCommand) Name() string {
	return "!duel_accept"
}

func (c *DuelAcceptCommand) Execute(ctx context.Context, p Payload, r Responder) error {
	if p.UserID == "" {
		c.log.ErrorContext(ctx, "duel command attempted without a twitch user id", "user_id", p.UserID)
		return r.Say(c.pickError(ctx, phrasebookDuelAcceptInternalError, p))
	}

	result, err := c.duellist.Accept(ctx, p.UserID, p.User)

	switch {
	case errors.Is(err, duelling.ErrNoIncomingDuel):
		return r.Say(c.pickError(ctx, phrasebookDuelAcceptNoIncomingDuel, p))
	case errors.Is(err, duelling.ErrInsufficientFunds):
		return r.Say(c.pickError(ctx, phrasebookDuelAcceptInsufficientFunds, p))
	case err != nil:
		c.log.ErrorContext(ctx, "failed to accept duel", "user", p.User, "target_user", "error", err)
		return r.Say(c.pickError(ctx, phrasebookDuelAcceptInternalError, p))
	}

	data := duelAcceptSuccessData{
		ChallengerUser: result.ChallengerName,
		OpponentUser:   result.OpponentName,
	}

	return r.Say(pickPhraseOrFallback(ctx, c.log, c.phrasePicker, phrasebookDuelAccept, phrasebookDuelAcceptSuccess, p.User, data))
}

func (c *DuelAcceptCommand) pickError(ctx context.Context, scenario string, p Payload) string {
	return pickPhraseOrFallback(ctx, c.log, c.phrasePicker, phrasebookDuelAccept, scenario, p.User, duelErrorData{User: p.User})
}
