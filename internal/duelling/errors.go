package duelling

import (
	"errors"

	"jijabot/internal/wallet"
)

var (
	ErrInvalidStake        = errors.New("duelling: stake must be positive")
	ErrDailyLimitReached   = errors.New("duelling: daily duel limit reached")
	ErrCannotChallengeSelf = errors.New("duelling: cannot challenge yourself")
	// ErrUserNotFound             = errors.New("duelling: user not found")
	ErrChallengerNotFound       = errors.New("duelling: challenger not found")
	ErrOpponentNotFound         = errors.New("duelling: opponent not found")
	ErrChallengerHasPendingDuel = errors.New("duelling: challenger has a pending duel")
	ErrOpponentHasPendingDuel   = errors.New("duelling: opponent has a pending duel")
	ErrInsufficientFunds        = wallet.ErrInsufficientFunds
	ErrNoIncomingDuel           = errors.New("duelling: no incoming duel to respond to")
	ErrNoOutgoingDuel           = errors.New("duelling: no outgoing duel to cancel")
	ErrNoTopUserYet             = errors.New("duelling: no top user yet")
)
