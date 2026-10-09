package duelling

type Rank struct {
	Name      string
	MinRating int
}

type RankProvider struct {
	ranks []Rank
}

func NewRankProvider(ranks []Rank) *RankProvider {
	return &RankProvider{ranks: ranks}
}

func (r *RankProvider) Rank(rating int) string {
	for _, rank := range r.ranks {
		if rating >= rank.MinRating {
			return rank.Name
		}
	}

	return ""
}
