package duelling

import (
	"context"
	"math/rand/v2"

	"jijabot/internal/store"
	"jijabot/internal/users"
)

type ChallengeResult struct{}
type CancelResult struct{}
type AcceptResult struct{}
type DeclineResult struct{}

type Duellist_ReferenceForClient_DeleteMe interface {
	Challenge(ctx context.Context, challengerName, opponentName string, stake int64) (ChallengeResult, error)
	Accept(ctx context.Context, opponentName string) (AcceptResult, error)
	Decline(ctx context.Context, opponentName string) (DeclineResult, error)
	Cancel(ctx context.Context, challengerName string) (CancelResult, error)
}

type RNG interface {
	Int64N(n int64) int64
}

type MathRandRNG struct{}

func (MathRandRNG) Int64N(n int64) int64 { return rand.Int64N(n) }

type Option func(*Duellist)

func WithRNG(rng RNG) Option {
	return func(d *Duellist) { d.rng = rng }
}

type Duellist struct {
	txBeginner store.TxBeginner
	users      users.Repository
	rng        RNG
}

func NewDuellist(txBeginner store.TxBeginner, userRepository users.Repository, options ...Option) *Duellist {
	duellist := &Duellist{
		txBeginner: txBeginner,
		users:      userRepository,
	}

	for _, o := range options {
		o(duellist)
	}

	return duellist
}

func (d *Duellist) Challenge(ctx context.Context, challengerName, opponentName string, stake int64) (ChallengeResult, error)
func (d *Duellist) Accept(ctx context.Context, opponentName string) (AcceptResult, error)
func (d *Duellist) Decline(ctx context.Context, opponentName string) (DeclineResult, error)
func (d *Duellist) Cancel(ctx context.Context, challengerName string) (CancelResult, error)
