package commands

import (
	"context"

	"jijabot/internal/logger"
)

const (
	phrasebookArenaStats              = "arena_stats"
	phrasebookArenaStatsInternalError = "internal_error"
	phrasebookArenaStatsSuccess       = "success"
)

type arenaStatsErrorData struct {
	User string
}

type arenaStatsData struct{}

type ArenaStatsCommand struct {
	phrasePicker PhrasePicker
	log          logger.Logger
}

func NewArenaStatsCommand(phrasePicker PhrasePicker, log logger.Logger) *ArenaStatsCommand {
	return &ArenaStatsCommand{
		phrasePicker: phrasePicker,
		log:          log.With("command", "!arena_stats"),
	}
}

func (c *ArenaStatsCommand) Name() string {
	return "!arena_stats"
}

func (c *ArenaStatsCommand) Execute(ctx context.Context, p Payload, r Responder) error {
	if p.UserID == "" {
		c.log.ErrorContext(ctx, "arena_stats command attempted without a twitch user id", "user_id", p.UserID)
		return r.Say(c.pickError(ctx, phrasebookArenaStatsInternalError, p))
	}

	data := arenaStatsData{}
	response := pickPhraseOrFallback(ctx, c.log, c.phrasePicker, phrasebookArenaStats, phrasebookArenaStatsSuccess, p.User, data)

	return r.Say(response)
}

func (c *ArenaStatsCommand) pickError(ctx context.Context, scenario string, p Payload) string {
	return pickPhraseOrFallback(ctx, c.log, c.phrasePicker, phrasebookArenaStatsInternalError, scenario, p.User, arenaStatsErrorData{User: p.User})
}
