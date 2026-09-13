package commands

import (
	"context"

	"jijabot/internal/logger"
)

const (
	phrasebookHi                 = "hi"
	phrasebookHiInternalError    = "internal_error"
	phrasebookHiGreeting         = "greeting"
	phrasebookHiGreetingStreamer = "greeting_streamer"
)

type hiGreetingData struct {
	User string
}

type HiCommand struct {
	phrasePicker PhrasePicker
	log          logger.Logger
}

func NewHiCommand(phrasePicker PhrasePicker, log logger.Logger) *HiCommand {
	return &HiCommand{
		phrasePicker: phrasePicker,
		log:          log.With("command", "!hi"),
	}
}

func (c *HiCommand) Name() string {
	return "!hi"
}

func (c *HiCommand) Execute(ctx context.Context, p Payload, r Responder) error {
	if p.UserID == "" {
		c.log.ErrorContext(ctx, "hi attempted without a twitch user id", "user_id", p.UserID)
		return r.Say(c.pickError(ctx, phrasebookHiInternalError, p))
	}

	scenario := phrasebookHiGreeting

	if isStreamer(p.User) {
		scenario = phrasebookHiGreetingStreamer
	}

	data := hiGreetingData{User: p.User}
	response := pickPhraseOrFallback(ctx, c.log, c.phrasePicker, phrasebookHi, scenario, p.User, data)

	return r.Say(response)
}

func (c *HiCommand) pickError(ctx context.Context, scenario string, p Payload) string {
	return pickPhraseOrFallback(ctx, c.log, c.phrasePicker, phrasebookHiInternalError, scenario, p.User, dailyErrorData{User: p.User})
}
