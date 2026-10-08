package commands

import (
	"context"
	"fmt"

	"jijabot/internal/duelling"
	"jijabot/internal/gambling"
	"jijabot/internal/logger"
)

const fallbackMessage = "@%s, что-то пошло не так, попробуй позже."

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

type StreamerChecker interface {
	IsStreamer(twitchUserID string) bool
}

type PhrasePicker interface {
	Pick(command, scenario string, data any) (string, error)
}

type BalanceGetter interface {
	Get(ctx context.Context, twitchUserID, username string) (int64, error)
}

type DailyRedeemer interface {
	Amount() int64
	Redeem(ctx context.Context, twitchUserID, username string) (int64, error)
}

type BetPlacer interface {
	Place(ctx context.Context, twitchUserID, username string, amount int64) (gambling.Result, error)
}

type Duellist interface {
	Challenge(ctx context.Context, challengerTwitchID, challengerName, opponentName string, stake int64) (duelling.ChallengeResult, error)
	Accept(ctx context.Context, opponentTwitchID, opponentName string) (duelling.AcceptResult, error)
	Decline(ctx context.Context, opponentTwitchID, opponentName string) (duelling.DeclineResult, error)
	Cancel(ctx context.Context, challengerTwitchID, challengerName string) (duelling.CancelResult, error)
}

type StatsProvider interface {
	Top(ctx context.Context) (duelling.TopResult, error)
	Stats(ctx context.Context, twitchUserID string) (duelling.StatsResult, error)
	Rank(ctx context.Context, twitchUserID string) (duelling.RankResult, error)
}

func pickPhraseOrFallback(ctx context.Context, log logger.Logger, phrasePicker PhrasePicker, command, scenario, user string, data any) string {
	message, err := phrasePicker.Pick(command, scenario, data)
	if err != nil {
		log.ErrorContext(ctx, "failed to pick phrase", "command", command, "scenario", scenario, "error", err)
		return fmt.Sprintf(fallbackMessage, user)
	}

	return message
}
