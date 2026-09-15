package duelling

type RatingCalculator interface {
	CalculateRating(challengerRating, opponentRating int, score float64) int
}

type EloCalculator struct {
	KFactor int
}

func NewEloCalculator(kFactor int) EloCalculator {
	return EloCalculator{KFactor: kFactor}
}

func (ec EloCalculator) CalculateRating(challengerRating, opponentRating int, score float64) int {
	return 0
}
