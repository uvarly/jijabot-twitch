package oauth

import (
	"context"
	"fmt"
	"io"
)

type deviceFlow interface {
	RequestCode(ctx context.Context) (DeviceCode, error)
	PollToken(ctx context.Context, deviceCode DeviceCode) (Token, error)
}

func Bootstrap(ctx context.Context, out io.Writer, flow deviceFlow, store TokenStore) (Token, error) {
	code, err := flow.RequestCode(ctx)
	if err != nil {
		return Token{}, fmt.Errorf("failed to request device code: %w", err)
	}

	_, _ = fmt.Fprintf(out, "To authorize Jija_Bot, open the following link:\n\n\t%s\n\n", code.VerificationURI)

	if code.UserCode != "" {
		_, _ = fmt.Fprintf(out, "If prompted, enter the code: %s\n\n", code.UserCode)
	}

	_, _ = fmt.Fprintf(out, "Waiting for you to authorize (code expires in %d minutes)...", int(code.ExpiresIn.Minutes()))

	token, err := flow.PollToken(ctx, code)
	if err != nil {
		return Token{}, fmt.Errorf("failed to obtain token: %w", err)
	}

	if err := store.Save(ctx, token); err != nil {
		return Token{}, fmt.Errorf("failed to persist token: %w", err)
	}

	_, _ = fmt.Fprintf(out, "Authorization complete, token saved.")

	return token, nil
}
