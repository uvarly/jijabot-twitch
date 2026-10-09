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

type arenaRankData struct {
	User        string
	Rank        string
	RatingValue int
}

type ArenaRankCommand struct {
	rankProvider  RankProvider
	statsProvider StatsProvider
	phrasePicker  PhrasePicker
	log           logger.Logger
}

func NewArenaRankCommand(rankProvider RankProvider, statsProvider StatsProvider, phrasePicker PhrasePicker, log logger.Logger) *ArenaRankCommand {
	return &ArenaRankCommand{
		rankProvider:  rankProvider,
		statsProvider: statsProvider,
		phrasePicker:  phrasePicker,
		log:           log.With("command", "!arena_rank"),
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

	rating, err := c.statsProvider.Rating(ctx, p.UserID, p.User)
	if err != nil {
		c.log.ErrorContext(ctx, "failed to get arena rank", "user_id", p.UserID, "error", err)
		return r.Say(c.pickError(ctx, phrasebookArenaRankInternalError, p))
	}

	rank := c.rankProvider.Rank(rating)

	data := arenaRankData{
		User:        p.User,
		Rank:        rank,
		RatingValue: rating,
	}

	response := pickPhraseOrFallback(ctx, c.log, c.phrasePicker, phrasebookArenaRank, phrasebookArenaRankSuccess, p.User, data)

	return r.Say(response)
}

func (c *ArenaRankCommand) pickError(ctx context.Context, scenario string, p Payload) string {
	return pickPhraseOrFallback(ctx, c.log, c.phrasePicker, phrasebookArenaRank, scenario, p.User, arenaRankErrorData{User: p.User})
}
