package commands

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"jijabot/internal/duelling"
	"jijabot/internal/logger"
)

const (
	phrasebookArenaTop              = "arena_top"
	phrasebookArenaTopInternalError = "internal_error"
	phrasebookArenaTopNoTopUserYet  = "no_top_user_yet"
	phrasebookArenaTopSuccess       = "success"

	arenaTopSeparator = " | "
)

type arenaTopErrorData struct {
	User string
}

type arenaTopData struct {
	Place       string
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

	lines := make([]string, 0, len(top.Entries))

	for i, entry := range top.Entries {
		data := arenaTopData{
			Place:       keycap(i + 1),
			User:        entry.Username,
			RatingValue: entry.MMR,
			WinCount:    entry.WinCount,
		}

		lines = append(lines, pickPhraseOrFallback(ctx, c.log, c.phrasePicker, phrasebookArenaTop, phrasebookArenaTopSuccess, p.User, data))
	}

	return r.Say(strings.Join(lines, arenaTopSeparator))
}

func (c *ArenaTopCommand) pickError(ctx context.Context, scenario string, p Payload) string {
	return pickPhraseOrFallback(ctx, c.log, c.phrasePicker, phrasebookArenaTop, scenario, p.User, arenaTopErrorData{User: p.User})
}

func keycap(n int) string {
	var b strings.Builder

	for _, digit := range strconv.Itoa(n) {
		b.WriteRune(digit)
		b.WriteString("\uFE0F\u20E3")
	}

	return b.String()
}
