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

// Authenticator defines the interface for authentication strategies.
type Authenticator interface {
	Authenticate(ctx context.Context) (*types.Token, error)
	Refresh(ctx context.Context) (*types.Token, error)
	IsTokenValid() bool
	GetToken() *types.Token
}

// OAuthRefreshAuthenticator implements OAuth 2.0 refresh token flow.
type OAuthRefreshAuthenticator struct {
	clientID     string
	clientSecret string
	refreshToken string
	tokenURL     string
	httpClient   *http.Client

	mu           sync.RWMutex
	currentToken *types.Token
}

// NewOAuthRefreshAuthenticator creates a new OAuth refresh token authenticator.
func NewOAuthRefreshAuthenticator(clientID, clientSecret, refreshToken, tokenURL string, httpClient *http.Client) *OAuthRefreshAuthenticator {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	if tokenURL == "" {
		tokenURL = "https://login.salesforce.com/services/oauth2/token"
	}
	return &OAuthRefreshAuthenticator{
		clientID:     clientID,
		clientSecret: clientSecret,
		refreshToken: refreshToken,
		tokenURL:     tokenURL,
		httpClient:   httpClient,
	}
}

// Authenticate performs authentication using the refresh token.
func (a *OAuthRefreshAuthenticator) Authenticate(ctx context.Context) (*types.Token, error) {
	return a.Refresh(ctx)
}

// Refresh refreshes the access token using the refresh token.
func (a *OAuthRefreshAuthenticator) Refresh(ctx context.Context) (*types.Token, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("client_id", a.clientID)
	data.Set("client_secret", a.clientSecret)
	data.Set("refresh_token", a.refreshToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var authErr types.AuthError
		if err := json.Unmarshal(body, &authErr); err == nil {
			authErr.StatusCode = resp.StatusCode
			return nil, &authErr
		}
		return nil, fmt.Errorf("authentication failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		InstanceURL  string `json:"instance_url"`
		IssuedAt     string `json:"issued_at"`
		TokenType    string `json:"token_type"`
	}

	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	token := &types.Token{
		AccessToken:  tokenResp.AccessToken,
		TokenType:    tokenResp.TokenType,
		RefreshToken: a.refreshToken,
		InstanceURL:  tokenResp.InstanceURL,
		IssuedAt:     time.Now(),
		ExpiresAt:    time.Now().Add(2 * time.Hour),
	}

	if tokenResp.RefreshToken != "" {
		token.RefreshToken = tokenResp.RefreshToken
		a.refreshToken = tokenResp.RefreshToken
	}

	a.mu.Lock()
	a.currentToken = token
	a.mu.Unlock()

	return token, nil
}

// IsTokenValid checks if the current token is valid.
func (a *OAuthRefreshAuthenticator) IsTokenValid() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.currentToken == nil {
		return false
	}
	return !a.currentToken.IsExpired()
}

// GetToken returns the current token.
func (a *OAuthRefreshAuthenticator) GetToken() *types.Token {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.currentToken
}

// PasswordAuthenticator implements username/password authentication.
type PasswordAuthenticator struct {
	username      string
	password      string
	securityToken string
	clientID      string
	clientSecret  string
	loginURL      string
	httpClient    *http.Client

	mu           sync.RWMutex
	currentToken *types.Token
}

// NewPasswordAuthenticator creates a new password authenticator.
func NewPasswordAuthenticator(username, password, securityToken, clientID, clientSecret, loginURL string, httpClient *http.Client) *PasswordAuthenticator {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	if loginURL == "" {
		loginURL = "https://login.salesforce.com"
	}
	return &PasswordAuthenticator{
		username:      username,
		password:      password,
		securityToken: securityToken,
		clientID:      clientID,
		clientSecret:  clientSecret,
		loginURL:      loginURL,
		httpClient:    httpClient,
	}
}

// Authenticate performs authentication using username and password.
func (a *PasswordAuthenticator) Authenticate(ctx context.Context) (*types.Token, error) {
	token, err := a.authenticateOAuth(ctx)
	if err == nil {
		return token, nil
	}
	return a.authenticateSOAP(ctx)
}

func (a *PasswordAuthenticator) authenticateOAuth(ctx context.Context) (*types.Token, error) {
	tokenURL := a.loginURL + "/services/oauth2/token"

	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("client_id", a.clientID)
	data.Set("client_secret", a.clientSecret)
	data.Set("username", a.username)
	data.Set("password", a.password+a.securityToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var authErr types.AuthError
		if err := json.Unmarshal(body, &authErr); err == nil {
			authErr.StatusCode = resp.StatusCode
			return nil, &authErr
		}
		return nil, fmt.Errorf("authentication failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		InstanceURL string `json:"instance_url"`
		TokenType   string `json:"token_type"`
	}

	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	token := &types.Token{
		AccessToken: tokenResp.AccessToken,
		TokenType:   tokenResp.TokenType,
		InstanceURL: tokenResp.InstanceURL,
		IssuedAt:    time.Now(),
		ExpiresAt:   time.Now().Add(2 * time.Hour),
	}

	a.mu.Lock()
	a.currentToken = token
	a.mu.Unlock()

	return token, nil
}

func (a *PasswordAuthenticator) authenticateSOAP(ctx context.Context) (*types.Token, error) {
	soapBody := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8" ?>
<env:Envelope
    xmlns:xsd="http://www.w3.org/2001/XMLSchema"
    xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
    xmlns:env="http://schemas.xmlsoap.org/soap/envelope/"
    xmlns:urn="urn:partner.soap.sforce.com">
    <env:Header>
        <urn:CallOptions>
            <urn:client>%s</urn:client>
            <urn:defaultNamespace>sf</urn:defaultNamespace>
        </urn:CallOptions>
    </env:Header>
    <env:Body>
        <n1:login xmlns:n1="urn:partner.soap.sforce.com">
            <n1:username>%s</n1:username>
            <n1:password>%s%s</n1:password>
        </n1:login>
    </env:Body>
</env:Envelope>`, a.clientID, a.username, escapeXML(a.password), a.securityToken)

	soapURL := a.loginURL + "/services/Soap/u/59.0"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, soapURL, strings.NewReader(soapBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create SOAP request: %w", err)
	}
	req.Header.Set("Content-Type", "text/xml")
	req.Header.Set("SOAPAction", "login")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute SOAP request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read SOAP response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("SOAP login failed with status %d: %s", resp.StatusCode, string(body))
	}

	sessionID := extractXMLValue(string(body), "sessionId")
	serverURL := extractXMLValue(string(body), "serverUrl")

	if sessionID == "" {
		return nil, fmt.Errorf("failed to extract session ID from SOAP response")
	}

	instanceURL := extractInstanceURL(serverURL)

	token := &types.Token{
		AccessToken: sessionID,
		TokenType:   "Bearer",
		InstanceURL: instanceURL,
		IssuedAt:    time.Now(),
		ExpiresAt:   time.Now().Add(2 * time.Hour),
	}

	a.mu.Lock()
	a.currentToken = token
	a.mu.Unlock()

	return token, nil
}

// Refresh re-authenticates using credentials.
func (a *PasswordAuthenticator) Refresh(ctx context.Context) (*types.Token, error) {
	return a.Authenticate(ctx)
}

// IsTokenValid checks if the current token is valid.
func (a *PasswordAuthenticator) IsTokenValid() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.currentToken == nil {
		return false
	}
	return !a.currentToken.IsExpired()
}

// GetToken returns the current token.
func (a *PasswordAuthenticator) GetToken() *types.Token {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.currentToken
}

// TokenAuthenticator uses a pre-existing access token.
type TokenAuthenticator struct {
	token *types.Token
	mu    sync.RWMutex
}

// NewTokenAuthenticator creates an authenticator with a pre-existing token.
func NewTokenAuthenticator(accessToken, instanceURL string) *TokenAuthenticator {
	return &TokenAuthenticator{
		token: &types.Token{
			AccessToken: accessToken,
			TokenType:   "Bearer",
			InstanceURL: instanceURL,
			IssuedAt:    time.Now(),
		},
	}
}

// Authenticate returns the existing token.
func (a *TokenAuthenticator) Authenticate(ctx context.Context) (*types.Token, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.token == nil {
		return nil, &types.AuthError{
			ErrorType:   "no_token",
			Description: "no token available",
		}
	}
	return a.token, nil
}

// Refresh cannot refresh a static token.
func (a *TokenAuthenticator) Refresh(ctx context.Context) (*types.Token, error) {
	return nil, &types.AuthError{
		ErrorType:   "refresh_not_supported",
		Description: "static token authenticator does not support refresh",
	}
}

// IsTokenValid checks if the token is valid.
func (a *TokenAuthenticator) IsTokenValid() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.token != nil && !a.token.IsExpired()
}

// GetToken returns the current token.
func (a *TokenAuthenticator) GetToken() *types.Token {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.token
}

// SetToken updates the token.
func (a *TokenAuthenticator) SetToken(token *types.Token) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.token = token
}

// Helper functions

func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}

func extractXMLValue(xml, tag string) string {
	startTag := "<" + tag + ">"
	endTag := "</" + tag + ">"

	startIdx := strings.Index(xml, startTag)
	if startIdx == -1 {
		return ""
	}
	startIdx += len(startTag)

	endIdx := strings.Index(xml[startIdx:], endTag)
	if endIdx == -1 {
		return ""
	}

	return xml[startIdx : startIdx+endIdx]
}

func extractInstanceURL(serverURL string) string {
	if serverURL == "" {
		return ""
	}
	parsed, err := url.Parse(serverURL)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%s://%s", parsed.Scheme, parsed.Host)
}
