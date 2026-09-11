package commands

import (
	"context"
	"fmt"

	"jijabot/internal/logger"
)

const (
	streamerNickname = "unclekost"
	fallbackMessage  = "@%s, что-то пошло не так, попробуй позже."
)

type Payload struct {
	User   string
	UserID string
	Text   string
	Args   []string
}

type Command interface {
	Name() string
	Execute(ctx context.Context, p Payload, r Responder) error
}

type Responder interface {
	Say(message string) error
}

type PhrasePicker interface {
	Pick(command, scenario string, data any) (string, error)
}

func pickPhraseOrFallback(ctx context.Context, log logger.Logger, phrasePicker PhrasePicker, command, scenario, user string, data any) string {
	message, err := phrasePicker.Pick(command, scenario, data)
	if err != nil {
		log.ErrorContext(ctx, "failed to pick phrase", "command", command, "scenario", scenario, "error", err)
		return fmt.Sprintf(fallbackMessage, user)
	}

	return message
}
