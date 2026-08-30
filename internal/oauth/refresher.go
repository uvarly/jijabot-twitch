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

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return Token{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// curl -X POST -H "Content-Type: application/x-www-form-urlencoded" -d "client_id=CLIENT_ID&client_secret=CLIENT_SECRET&refresh_token=REFRESH_TOKEN&grant_type=refresh_token" https://api.twitch.tv/kraken/oauth2/token
	resp, err := r.httpDoer.Do(req)
	if err != nil {
		return Token{}, fmt.Errorf("failed to do a request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Token{}, fmt.Errorf("failed to refresh token, received status: %s", resp.Status)
	}

	var token struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return Token{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return Token{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresIn:    time.Now().Add(time.Duration(token.ExpiresIn) * time.Second),
	}, nil
}
