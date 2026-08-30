package oauth

import (
	"context"
	"time"
)

type Token struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    time.Time
}

func (t Token) IsValid() bool {
	return t.AccessToken != "" && !t.IsExpired()
}

func (t Token) IsExpired() bool {
	return time.Now().After(t.ExpiresIn)
}

type TokenProvider interface {
	Token(ctx context.Context) (Token, error)
}

type TokenStore interface {
	Load(ctx context.Context) (Token, error)
	Save(ctx context.Context, t Token) error
}
