package duelling

import "math"

const (
	WinScore  = 1.0
	DrawScore = 0.5
	LossScore = 0.0
)

type RatingCalculator interface {
	Delta(playerRating, opponentRating int, score float64) int
}

type EloCalculator struct {
	KFactor int
}

func NewEloCalculator(kFactor int) EloCalculator {
	return EloCalculator{KFactor: kFactor}
}

func (ec EloCalculator) Delta(playerRating, opponentRating int, score float64) int {
	expected := 1.0 / (1.0 + math.Pow(10, float64(opponentRating-playerRating)/400.0))

	return int(math.Round(float64(ec.KFactor) * (score - expected)))
}

func scoresFor(result DuelResult) (float64, float64) {
	switch result {
	case DuelResultChallengerWon:
		return WinScore, LossScore
	case DuelResultOpponentWon:
		return LossScore, WinScore
	default:
		return DrawScore, DrawScore
	}
}

func clampRating(rating int) int {
	if rating < 0 {
		return 0
	}

	return rating
}
