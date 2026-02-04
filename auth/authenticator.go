// Package auth provides authentication strategies for Salesforce.
package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/PramithaMJ/salesforce/types"
)

// Authenticator defines the authentication interface.
type Authenticator interface {
	Authenticate(ctx context.Context) (*types.Token, error)
	Refresh(ctx context.Context) (*types.Token, error)
	IsTokenValid() bool
	GetToken() *types.Token
}

// BaseAuthenticator provides common authentication functionality.
type BaseAuthenticator struct {
	mu         sync.RWMutex
	token      *types.Token
	httpClient *http.Client
	tokenURL   string
}

// GetToken returns the current token.
func (a *BaseAuthenticator) GetToken() *types.Token {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.token
}

// IsTokenValid checks if the current token is valid.
func (a *BaseAuthenticator) IsTokenValid() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.token != nil && !a.token.IsExpired()
}

// SetToken sets the current token.
func (a *BaseAuthenticator) SetToken(token *types.Token) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.token = token
}

// RefreshTokenAuthenticator uses OAuth 2.0 refresh token flow.
type RefreshTokenAuthenticator struct {
	BaseAuthenticator
	clientID     string
	clientSecret string
	refreshToken string
}

// NewRefreshTokenAuthenticator creates a refresh token authenticator.
func NewRefreshTokenAuthenticator(clientID, clientSecret, refreshToken, tokenURL string) *RefreshTokenAuthenticator {
	return &RefreshTokenAuthenticator{
		BaseAuthenticator: BaseAuthenticator{
			httpClient: &http.Client{Timeout: 30 * time.Second},
			tokenURL:   tokenURL,
		},
		clientID:     clientID,
		clientSecret: clientSecret,
		refreshToken: refreshToken,
	}
}

// Authenticate performs initial authentication.
func (a *RefreshTokenAuthenticator) Authenticate(ctx context.Context) (*types.Token, error) {
	return a.Refresh(ctx)
}

// Refresh refreshes the access token.
func (a *RefreshTokenAuthenticator) Refresh(ctx context.Context) (*types.Token, error) {
	data := url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {a.clientID},
		"client_secret": {a.clientSecret},
		"refresh_token": {a.refreshToken},
	}
	token, err := a.doTokenRequest(ctx, data)
	if err != nil {
		return nil, err
	}
	a.SetToken(token)
	return token, nil
}

func (a *RefreshTokenAuthenticator) doTokenRequest(ctx context.Context, data url.Values) (*types.Token, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		var authErr types.AuthError
		json.Unmarshal(body, &authErr)
		authErr.StatusCode = resp.StatusCode
		return nil, &authErr
	}
	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		InstanceURL string `json:"instance_url"`
		ID          string `json:"id"`
		IssuedAt    string `json:"issued_at"`
		Scope       string `json:"scope"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}
	issuedAt := time.Now()
	if tokenResp.IssuedAt != "" {
		if ts, err := parseTimestamp(tokenResp.IssuedAt); err == nil {
			issuedAt = ts
		}
	}
	return &types.Token{
		AccessToken:  tokenResp.AccessToken,
		TokenType:    tokenResp.TokenType,
		InstanceURL:  tokenResp.InstanceURL,
		ID:           tokenResp.ID,
		IssuedAt:     issuedAt,
		Scope:        tokenResp.Scope,
		RefreshToken: a.refreshToken,
	}, nil
}

// PasswordAuthenticator uses username-password flow.
type PasswordAuthenticator struct {
	BaseAuthenticator
	clientID      string
	clientSecret  string
	username      string
	password      string
	securityToken string
}

// NewPasswordAuthenticator creates a password authenticator.
func NewPasswordAuthenticator(clientID, clientSecret, username, password, securityToken, tokenURL string) *PasswordAuthenticator {
	return &PasswordAuthenticator{
		BaseAuthenticator: BaseAuthenticator{
			httpClient: &http.Client{Timeout: 30 * time.Second},
			tokenURL:   tokenURL,
		},
		clientID:      clientID,
		clientSecret:  clientSecret,
		username:      username,
		password:      password,
		securityToken: securityToken,
	}
}

// Authenticate performs username-password authentication.
func (a *PasswordAuthenticator) Authenticate(ctx context.Context) (*types.Token, error) {
	data := url.Values{
		"grant_type":    {"password"},
		"client_id":     {a.clientID},
		"client_secret": {a.clientSecret},
		"username":      {a.username},
		"password":      {a.password + a.securityToken},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("auth request failed: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		var authErr types.AuthError
		json.Unmarshal(body, &authErr)
		authErr.StatusCode = resp.StatusCode
		return nil, &authErr
	}
	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		InstanceURL string `json:"instance_url"`
		ID          string `json:"id"`
		IssuedAt    string `json:"issued_at"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}
	token := &types.Token{
		AccessToken: tokenResp.AccessToken,
		TokenType:   tokenResp.TokenType,
		InstanceURL: tokenResp.InstanceURL,
		ID:          tokenResp.ID,
		IssuedAt:    time.Now(),
	}
	a.SetToken(token)
	return token, nil
}

// Refresh re-authenticates using credentials.
func (a *PasswordAuthenticator) Refresh(ctx context.Context) (*types.Token, error) {
	return a.Authenticate(ctx)
}

// TokenAuthenticator uses a pre-existing token.
type TokenAuthenticator struct {
	BaseAuthenticator
}

// NewTokenAuthenticator creates a token authenticator.
func NewTokenAuthenticator(accessToken, instanceURL string) *TokenAuthenticator {
	return &TokenAuthenticator{
		BaseAuthenticator: BaseAuthenticator{
			token: &types.Token{
				AccessToken: accessToken,
				InstanceURL: instanceURL,
				IssuedAt:    time.Now(),
			},
		},
	}
}

// Authenticate returns the existing token.
func (a *TokenAuthenticator) Authenticate(ctx context.Context) (*types.Token, error) {
	return a.GetToken(), nil
}

// Refresh is not supported for token authenticator.
func (a *TokenAuthenticator) Refresh(ctx context.Context) (*types.Token, error) {
	return nil, fmt.Errorf("token refresh not supported")
}

func parseTimestamp(ts string) (time.Time, error) {
	if len(ts) > 10 {
		ms, err := time.Parse("2006-01-02T15:04:05.000Z", ts)
		if err == nil {
			return ms, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid timestamp")
}
