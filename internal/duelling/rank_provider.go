package duelling

import "sort"

type Rank struct {
	Name      string
	MinRating int
}

type RankProvider struct {
	ranks []Rank
}

func NewRankProvider(ranks []Rank) *RankProvider {
	sort.Slice(ranks, func(i, j int) bool {
		return ranks[i].MinRating > ranks[j].MinRating
	})

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
