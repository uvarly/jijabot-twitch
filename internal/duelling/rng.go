package duelling

import "math/rand/v2"

type RNG interface {
	Int64N(n int64) int64
}

type MathRandRNG struct{}

func (MathRandRNG) Int64N(n int64) int64 { return rand.Int64N(n) }
