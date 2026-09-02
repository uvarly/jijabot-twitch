package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type HTTPDoer interface {
	Do(r *http.Request) (*http.Response, error)
}

type Refresher struct {
	clientID     string
	clientSecret string
	endpoint     string
	httpDoer     HTTPDoer
}

func NewRefresher(clientID, clientSecret, endpoint string, httpDoer HTTPDoer) *Refresher {
	return &Refresher{
		clientID:     clientID,
		clientSecret: clientSecret,
		endpoint:     endpoint,
		httpDoer:     httpDoer,
	}
}

func (r *Refresher) Refresh(ctx context.Context, refreshToken string) (Token, error) {
	form := url.Values{
		"client_id":     {r.clientID},
		"client_secret": {r.clientSecret},
		"refresh_token": {refreshToken},
		"grant_type":    {"refresh_token"},
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, r.endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return Token{}, fmt.Errorf("failed to create request: %w", err)
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := r.httpDoer.Do(request)
	if err != nil {
		return Token{}, fmt.Errorf("failed to do a request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return Token{}, fmt.Errorf("failed to refresh token, received status: %s", response.Status)
	}

	var token struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}

	if err := json.NewDecoder(response.Body).Decode(&token); err != nil {
		return Token{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return Token{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresIn:    time.Now().Add(time.Duration(token.ExpiresIn) * time.Second),
	}, nil
}
