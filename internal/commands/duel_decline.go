package commands

import (
	"context"
	"errors"

	"jijabot/internal/duelling"
	"jijabot/internal/logger"
)

const (
	phrasebookDuelDecline               = "duel_decline"
	phrasebookDuelDeclineInternalError  = "internal_error"
	phrasebookDuelDeclineNoIncomingDuel = "no_incoming_duel"
	phrasebookDuelDeclineSuccess        = "success"
)

type duelDeclineErrorData struct {
	User string
}

type duelDeclineSuccessData struct {
	ChallengerUser string
	OpponentUser   string
}

type DuelDeclineCommand struct {
	duellist     Duellist
	phrasePicker PhrasePicker
	log          logger.Logger
}

func NewDuelDeclineCommand(duellist Duellist, phrasePicker PhrasePicker, log logger.Logger) *DuelDeclineCommand {
	return &DuelDeclineCommand{
		duellist:     duellist,
		phrasePicker: phrasePicker,
		log:          log.With("command", "!duel_decline"),
	}
}

func (c *DuelDeclineCommand) Name() string {
	return "!duel_decline"
}

func (c *DuelDeclineCommand) Execute(ctx context.Context, p Payload, r Responder) error {
	if p.UserID == "" {
		c.log.ErrorContext(ctx, "duel command attempted without a twitch user id", "user_id", p.UserID)
		return r.Say(c.pickError(ctx, phrasebookDuelDeclineInternalError, p))
	}

	result, err := c.duellist.Decline(ctx, p.UserID, p.User)

	switch {
	case errors.Is(err, duelling.ErrNoIncomingDuel):
		return r.Say(c.pickError(ctx, phrasebookDuelDeclineNoIncomingDuel, p))
	case err != nil:
		c.log.ErrorContext(ctx, "failed to decline duel", "user", p.User, "target_user", "error", err)
		return r.Say(c.pickError(ctx, phrasebookDuelDeclineInternalError, p))
	}

	data := duelDeclineSuccessData{
		ChallengerUser: result.ChallengerName,
		OpponentUser:   result.OpponentName,
	}

	return r.Say(pickPhraseOrFallback(ctx, c.log, c.phrasePicker, phrasebookDuelDecline, phrasebookDuelDeclineSuccess, p.User, data))
}

func (c *DuelDeclineCommand) pickError(ctx context.Context, scenario string, p Payload) string {
	return pickPhraseOrFallback(ctx, c.log, c.phrasePicker, phrasebookDuelDecline, scenario, p.User, duelErrorData{User: p.User})
}
