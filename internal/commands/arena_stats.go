package commands

import (
	"context"
	"fmt"

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

type arenaStatsData struct {
	User           string
	WinCount       int64
	LossCount      int64
	DrawCount      int64
	WinRatePercent string
}

type ArenaStatsCommand struct {
	statsProvider StatsProvider
	phrasePicker  PhrasePicker
	log           logger.Logger
}

func NewArenaStatsCommand(statsProvider StatsProvider, phrasePicker PhrasePicker, log logger.Logger) *ArenaStatsCommand {
	return &ArenaStatsCommand{
		statsProvider: statsProvider,
		phrasePicker:  phrasePicker,
		log:           log.With("command", "!arena_stats"),
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

	stats, err := c.statsProvider.Stats(ctx, p.UserID, p.User)
	if err != nil {
		c.log.ErrorContext(ctx, "failed to get arena stats", "user_id", p.UserID, "error", err)
		return r.Say(c.pickError(ctx, phrasebookArenaStatsInternalError, p))
	}

	data := arenaStatsData{
		User:           p.User,
		WinCount:       stats.WinCount,
		LossCount:      stats.LossCount,
		DrawCount:      stats.DrawCount,
		WinRatePercent: fmt.Sprintf("%.2f", stats.WinRatePercent),
	}

	response := pickPhraseOrFallback(ctx, c.log, c.phrasePicker, phrasebookArenaStats, phrasebookArenaStatsSuccess, p.User, data)

	return r.Say(response)
}

func (c *ArenaStatsCommand) pickError(ctx context.Context, scenario string, p Payload) string {
	return pickPhraseOrFallback(ctx, c.log, c.phrasePicker, phrasebookArenaStatsInternalError, scenario, p.User, arenaStatsErrorData{User: p.User})
}
