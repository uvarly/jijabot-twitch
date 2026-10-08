package commands

import (
	"context"
	"errors"

	"jijabot/internal/duelling"
	"jijabot/internal/logger"
)

const (
	phrasebookArenaTop              = "arena_top"
	phrasebookArenaTopInternalError = "internal_error"
	phrasebookArenaTopNoTopUserYet  = "no_top_user_yet"
	phrasebookArenaTopSuccess       = "success"
)

type arenaTopErrorData struct {
	User string
}

type arenaTopData struct {
	User        string
	RatingValue int64
	WinCount    int64
}

type ArenaTopCommand struct {
	statsProvider StatsProvider
	phrasePicker  PhrasePicker
	log           logger.Logger
}

func NewArenaTopCommand(statsProvider StatsProvider, phrasePicker PhrasePicker, log logger.Logger) *ArenaTopCommand {
	return &ArenaTopCommand{
		statsProvider: statsProvider,
		phrasePicker:  phrasePicker,
		log:           log.With("command", "!arena_top"),
	}
}

func (c *ArenaTopCommand) Name() string {
	return "!arena_top"
}

func (c *ArenaTopCommand) Execute(ctx context.Context, p Payload, r Responder) error {
	if p.UserID == "" {
		c.log.ErrorContext(ctx, "arena_top command attempted without a twitch user id", "user_id", p.UserID)
		return r.Say(c.pickError(ctx, phrasebookArenaTopInternalError, p))
	}

	top, err := c.statsProvider.Top(ctx)

	switch {
	case errors.Is(err, duelling.ErrNoTopUserYet):
		return r.Say(c.pickError(ctx, phrasebookArenaTopNoTopUserYet, p))
	case err != nil:
		c.log.ErrorContext(ctx, "failed to get arena top", "error", err)
		return r.Say(c.pickError(ctx, phrasebookArenaTopInternalError, p))
	}

	data := arenaTopData{
		User:        top.Username,
		RatingValue: top.MMR,
		WinCount:    top.WinCount,
	}

	response := pickPhraseOrFallback(ctx, c.log, c.phrasePicker, phrasebookArenaTop, phrasebookArenaTopSuccess, p.User, data)

	return r.Say(response)
}

func (c *ArenaTopCommand) pickError(ctx context.Context, scenario string, p Payload) string {
	return pickPhraseOrFallback(ctx, c.log, c.phrasePicker, phrasebookArenaTop, scenario, p.User, arenaTopErrorData{User: p.User})
}
