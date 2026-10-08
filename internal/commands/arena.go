package commands

import (
	"context"

	"jijabot/internal/logger"
)

const (
	phrasebookArena              = "arena"
	phrasebookArenaInternalError = "internal_error"
	phrasebookArenaSuccess       = "success"
)

type arenaErrorData struct {
	User string
}

type arenaData struct{}

type ArenaCommand struct {
	phrasePicker PhrasePicker
	log          logger.Logger
}

func NewArenaCommand(phrasePicker PhrasePicker, log logger.Logger) *ArenaCommand {
	return &ArenaCommand{
		phrasePicker: phrasePicker,
		log:          log.With("command", "!arena"),
	}
}

func (c *ArenaCommand) Name() string {
	return "!arena"
}

func (c *ArenaCommand) Execute(ctx context.Context, p Payload, r Responder) error {
	if p.UserID == "" {
		c.log.ErrorContext(ctx, "arena command attempted without a twitch user id", "user_id", p.UserID)
		return r.Say(c.pickError(ctx, phrasebookArenaInternalError, p))
	}

	data := arenaData{}
	response := pickPhraseOrFallback(ctx, c.log, c.phrasePicker, phrasebookArena, phrasebookArenaSuccess, p.User, data)

	return r.Say(response)
}

func (c *ArenaCommand) pickError(ctx context.Context, scenario string, p Payload) string {
	return pickPhraseOrFallback(ctx, c.log, c.phrasePicker, phrasebookArenaInternalError, scenario, p.User, arenaErrorData{User: p.User})
}
