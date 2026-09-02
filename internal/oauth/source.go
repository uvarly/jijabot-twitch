package oauth

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type refresher interface {
	Refresh(ctx context.Context, refreshToken string) (Token, error)
}

type SourceOption func(*Source)

func WithRefreshAhead(d time.Duration) SourceOption {
	return func(s *Source) { s.refreshAhead = d }
}

type Source struct {
	mu           sync.Mutex
	store        TokenStore
	refresher    refresher
	token        Token
	tokenLoaded  bool
	refreshAhead time.Duration
}

func NewSource(store TokenStore, refresher refresher, options ...SourceOption) *Source {
	source := &Source{
		store:        store,
		refresher:    refresher,
		refreshAhead: 5 * time.Minute,
	}

	for _, o := range options {
		o(source)
	}

	return source
}

func (s *Source) Token(ctx context.Context) (Token, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.tokenLoaded {
		t, err := s.store.Load(ctx)
		if err != nil {
			return Token{}, fmt.Errorf("failed to load token: %w", err)
		}

		s.token = t
		s.tokenLoaded = true
	}

	if s.needsRefresh() {
		refreshedToken, err := s.refresher.Refresh(ctx, s.token.RefreshToken)
		if err != nil {
			return Token{}, fmt.Errorf("failed to refresh token: %w", err)
		}

		if err := s.store.Save(ctx, refreshedToken); err != nil {
			return Token{}, fmt.Errorf("failed to persist refreshed token: %w", err)
		}

		s.token = refreshedToken
	}

	return s.token, nil
}

func (s *Source) needsRefresh() bool {
	if !s.token.IsValid() {
		return true
	}

	return time.Now().Add(s.refreshAhead).After(s.token.ExpiresIn)
}
