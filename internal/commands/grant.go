package commands

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"jijabot/internal/grant"
	"jijabot/internal/logger"
)

const (
	phrasebookGrant                          = "grant"
	phrasebookGrantInternalError             = "internal_error"
	phrasebookGrantPermissionDenied          = "permission_denied"
	phrasebookGrantMissingArgs               = "missing_args"
	phrasebookGrantInvalidAmountNotInteger   = "invalid_amount_not_integer"
	phrasebookGrantInvalidAmountLessThanZero = "invalid_amount_less_than_zero"
	phrasebookGrantTargetUserNotFound        = "target_user_not_found"
	phrasebookGrantSuccess                   = "success"
)

type grantErrorData struct {
	User string
}

type grantSuccessData struct {
	User       string
	TargetUser string
	Amount     int64
	NewBalance int64
}

type Granter interface {
	Grant(ctx context.Context, username string, amount int64) (grant.Result, error)
}

type GrantCommand struct {
	granter      Granter
	phrasePicker PhrasePicker
	log          logger.Logger
}

func NewGrantCommand(granter Granter, phrasePicker PhrasePicker, log logger.Logger) *GrantCommand {
	return &GrantCommand{
		granter:      granter,
		phrasePicker: phrasePicker,
		log:          log.With("command", "!grant"),
	}
}

func (c *GrantCommand) Name() string {
	return "!grant"
}

func (c *GrantCommand) Execute(ctx context.Context, p Payload, r Responder) error {
	if !isStreamer(p.User) {
		c.log.InfoContext(ctx, "grant command attempted by a non-streamer user", "user", p.User)
		return r.Say(c.pickError(ctx, phrasebookGrantPermissionDenied, p))
	}

	if len(p.Args) != 2 {
		return r.Say(c.pickError(ctx, phrasebookGrantMissingArgs, p))
	}

	targetUser := strings.TrimPrefix(p.Args[0], "@")
	if targetUser == "" {
		return r.Say(c.pickError(ctx, phrasebookGrantMissingArgs, p))
	}

	amount, err := strconv.ParseInt(p.Args[1], 10, 64)
	if err != nil {
		return r.Say(c.pickError(ctx, phrasebookGrantInvalidAmountNotInteger, p))
	}

	result, err := c.granter.Grant(ctx, targetUser, amount)

	switch {
	case errors.Is(err, grant.ErrInvalidAmount):
		return r.Say(c.pickError(ctx, phrasebookGrantInvalidAmountLessThanZero, p))
	case errors.Is(err, grant.ErrUserNotFound):
		return r.Say(c.pickError(ctx, phrasebookGrantTargetUserNotFound, p))
	case err != nil:
		c.log.ErrorContext(ctx, "failed to grant jija-coins", "target_user", targetUser, "amount", amount, "error", err)
		return r.Say(c.pickError(ctx, phrasebookGrantInternalError, p))
	}

	data := grantSuccessData{
		User:       p.User,
		TargetUser: result.TargetUser,
		Amount:     amount,
		NewBalance: result.NewBalance,
	}

	return r.Say(pickPhraseOrFallback(ctx, c.log, c.phrasePicker, phrasebookGrant, phrasebookGrantSuccess, p.User, data))
}

func (c *GrantCommand) pickError(ctx context.Context, scenario string, p Payload) string {
	return pickPhraseOrFallback(ctx, c.log, c.phrasePicker, phrasebookGrant, scenario, p.User, grantErrorData{User: p.User})
}
