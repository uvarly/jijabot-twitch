package commands

import (
	"context"

	"jijabot/internal/logger"
)

const (
	phrasebookArenaRank              = "arena_rank"
	phrasebookArenaRankInternalError = "internal_error"
	phrasebookArenaRankSuccess       = "success"
)

type arenaRankErrorData struct {
	User string
}

type arenaRankData struct{}

type ArenaRankCommand struct {
	phrasePicker PhrasePicker
	log          logger.Logger
}

func NewArenaRankCommand(phrasePicker PhrasePicker, log logger.Logger) *ArenaRankCommand {
	return &ArenaRankCommand{
		phrasePicker: phrasePicker,
		log:          log.With("command", "!arena_rank"),
	}
}

func (c *ArenaRankCommand) Name() string {
	return "!arena_rank"
}

func (c *ArenaRankCommand) Execute(ctx context.Context, p Payload, r Responder) error {
	if p.UserID == "" {
		c.log.ErrorContext(ctx, "arena_rank command attempted without a twitch user id", "user_id", p.UserID)
		return r.Say(c.pickError(ctx, phrasebookArenaRankInternalError, p))
	}

	data := arenaRankData{}
	response := pickPhraseOrFallback(ctx, c.log, c.phrasePicker, phrasebookArenaRank, phrasebookArenaRankSuccess, p.User, data)

	return r.Say(response)
}

func (c *ArenaRankCommand) pickError(ctx context.Context, scenario string, p Payload) string {
	return pickPhraseOrFallback(ctx, c.log, c.phrasePicker, phrasebookArenaRankInternalError, scenario, p.User, arenaRankErrorData{User: p.User})
}
