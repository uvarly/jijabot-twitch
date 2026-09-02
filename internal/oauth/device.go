package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	TwitchDeviceAuthEndpoint = "https://id.twitch.tv/oauth2/device"
	TwitchTokenEndpoint      = "https://id.twitch.tv/oauth2/token"
	deviceCodeGrantType      = "urn:ietf:params:oauth:grant-type:device_code"
)

var ErrDeviceCodeExpired = errors.New("oauth: device code expired before authorization")

type DeviceCode struct {
	DeviceCode      string
	UserCode        string
	VerificationURI string
	ExpiresIn       time.Duration
	Interval        time.Duration
}

type DeviceAuthorizerOption func(*DeviceAuthorizer)

func WithDeviceEndpoint(endpoint string) DeviceAuthorizerOption {
	return func(a *DeviceAuthorizer) { a.deviceEndpoint = endpoint }
}

func WithTokenEndpoint(endpoint string) DeviceAuthorizerOption {
	return func(a *DeviceAuthorizer) { a.tokenEndpoint = endpoint }
}

type DeviceAuthorizer struct {
	clientID string
	scopes   []string
	httpDoer HTTPDoer

	deviceEndpoint string
	tokenEndpoint  string
}

func NewDeviceAuthorizer(clientID string, scopes []string, httpDoer HTTPDoer, options ...DeviceAuthorizerOption) *DeviceAuthorizer {
	deviceAuthorizer := &DeviceAuthorizer{
		clientID:       clientID,
		scopes:         scopes,
		httpDoer:       httpDoer,
		deviceEndpoint: TwitchDeviceAuthEndpoint,
		tokenEndpoint:  TwitchTokenEndpoint,
	}

	for _, o := range options {
		o(deviceAuthorizer)
	}

	return deviceAuthorizer
}

func (da *DeviceAuthorizer) RequestCode(ctx context.Context) (DeviceCode, error) {
	form := url.Values{
		"client_id": {da.clientID},
		"scopes":    {strings.Join(da.scopes, " ")},
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, da.deviceEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return DeviceCode{}, fmt.Errorf("failed to create request: %w", err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := da.httpDoer.Do(request)
	if err != nil {
		return DeviceCode{}, fmt.Errorf("failed request device code: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return DeviceCode{}, fmt.Errorf("failed to request device code, received status: %s", response.Status)
	}

	var body struct {
		DeviceCode      string `json:"device_code"`
		UserCode        string `json:"user_code"`
		VerificationURI string `json:"verification_uri"`
		ExpiresIn       int    `json:"expires_in"`
		Interval        int    `json:"interval"`
	}

	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		return DeviceCode{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return DeviceCode{
		DeviceCode:      body.DeviceCode,
		UserCode:        body.UserCode,
		VerificationURI: body.VerificationURI,
		ExpiresIn:       time.Duration(body.ExpiresIn) * time.Second,
		Interval:        time.Duration(body.Interval) * time.Second,
	}, nil
}

func (da *DeviceAuthorizer) PollToken(ctx context.Context, deviceCode DeviceCode) (Token, error) {
	deadline := time.Now().Add(deviceCode.ExpiresIn)

	for {
		if time.Now().After(deadline) {
			return Token{}, ErrDeviceCodeExpired
		}

		token, pending, err := da.tryToken(ctx, deviceCode.DeviceCode)
		if err != nil {
			return Token{}, fmt.Errorf("failed to poll token: %w", err)
		}

		if !pending {
			return token, nil
		}

		select {
		case <-ctx.Done():
			return Token{}, fmt.Errorf("failed to poll token: %w", ctx.Err())
		case <-time.After(deviceCode.Interval):
		}
	}
}

func (da *DeviceAuthorizer) tryToken(ctx context.Context, deviceCode string) (Token, bool, error) {
	form := url.Values{
		"client_id":   {da.clientID},
		"scopes":      {strings.Join(da.scopes, " ")},
		"device_code": {deviceCode},
		"grant_type":  {deviceCodeGrantType},
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, da.tokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return Token{}, false, fmt.Errorf("failed to create request: %w", err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := da.httpDoer.Do(request)
	if err != nil {
		return Token{}, false, fmt.Errorf("failed to poll for token: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusBadRequest {
		var body struct {
			Message string `json:"message"`
		}

		if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
			return Token{}, false, fmt.Errorf("failed to unmarshal pending response: %w", err)
		}

		if body.Message == "authorization_pending" {
			return Token{}, true, nil
		}

		return Token{}, false, fmt.Errorf("failed to poll for token: %s", body.Message)
	}

	if response.StatusCode != http.StatusOK {
		return Token{}, false, fmt.Errorf("failed to poll for token, received status: %s", response.Status)
	}

	var body struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}

	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		return Token{}, false, fmt.Errorf("failed to unmarshal token response: %w", err)
	}

	return Token{
		AccessToken:  body.AccessToken,
		RefreshToken: body.RefreshToken,
		ExpiresIn:    time.Now().Add(time.Duration(body.ExpiresIn) * time.Second),
	}, false, nil
}
