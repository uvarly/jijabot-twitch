package commands

import (
	"context"
	"errors"

	"jijabot/internal/duelling"
	"jijabot/internal/logger"
)

const (
	phrasebookDuelCancel               = "duel_cancel"
	phrasebookDuelCancelInternalError  = "internal_error"
	phrasebookDuelCancelNoOutgoingDuel = "no_outgoing_duel"
	phrasebookDuelCancelSuccess        = "success"
)

type duelCancelErrorData struct {
	User string
}

type duelCancelSuccessData struct {
	ChallengerUser string
	OpponentUser   string
}

type DuelCancelCommand struct {
	duellist     Duellist
	phrasePicker PhrasePicker
	log          logger.Logger
}

func NewDuelCancelCommand(duellist Duellist, phrasePicker PhrasePicker, log logger.Logger) *DuelCancelCommand {
	return &DuelCancelCommand{
		duellist:     duellist,
		phrasePicker: phrasePicker,
		log:          log.With("command", "!duel_cancel"),
	}
}

func (c *DuelCancelCommand) Name() string {
	return "!duel_cancel"
}

func (c *DuelCancelCommand) Execute(ctx context.Context, p Payload, r Responder) error {
	if p.UserID == "" {
		c.log.ErrorContext(ctx, "duel command attempted without a twitch user id", "user_id", p.UserID)
		return r.Say(c.pickError(ctx, phrasebookDuelCancelInternalError, p))
	}

	result, err := c.duellist.Cancel(ctx, p.UserID, p.User)

	switch {
	case errors.Is(err, duelling.ErrNoOutgoingDuel):
		return r.Say(c.pickError(ctx, phrasebookDuelCancelNoOutgoingDuel, p))
	case err != nil:
		c.log.ErrorContext(ctx, "failed to cancel duel", "user", p.User, "target_user", "error", err)
		return r.Say(c.pickError(ctx, phrasebookDuelCancelInternalError, p))
	}

	data := duelCancelSuccessData{
		ChallengerUser: result.ChallengerName,
		OpponentUser:   result.OpponentName,
	}

	return r.Say(pickPhraseOrFallback(ctx, c.log, c.phrasePicker, phrasebookDuelCancel, phrasebookDuelCancelSuccess, p.User, data))
}

func (c *DuelCancelCommand) pickError(ctx context.Context, scenario string, p Payload) string {
	return pickPhraseOrFallback(ctx, c.log, c.phrasePicker, phrasebookDuelCancel, scenario, p.User, duelErrorData{User: p.User})
}
