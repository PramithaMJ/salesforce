package salesforce

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/PramithaMJ/salesforce/auth"
	sfhttp "github.com/PramithaMJ/salesforce/http"
	"github.com/PramithaMJ/salesforce/services"
	"github.com/PramithaMJ/salesforce/types"
)

// SalesforceClient is the main client for interacting with Salesforce.
type SalesforceClient struct {
	config        *Config
	authenticator auth.Authenticator
	httpClient    *sfhttp.Client
	token         *types.Token

	sobjects SObjectService
	query    QueryService
	bulk     BulkService
	tooling  ToolingService
	apex     ApexService

	mu sync.RWMutex
}

// NewClient creates a new Salesforce client with the given options.
func NewClient(opts ...Option) (*SalesforceClient, error) {
	config := DefaultConfig()
	for _, opt := range opts {
		opt(config)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	client := &SalesforceClient{
		config: config,
	}

	// Setup HTTP client
	httpClientOpts := []sfhttp.ClientOption{
		sfhttp.WithAPIVersion(config.APIVersion),
	}

	if config.HTTPClient != nil {
		httpClientOpts = append(httpClientOpts, sfhttp.WithHTTPClient(config.HTTPClient))
	} else {
		httpClientOpts = append(httpClientOpts, sfhttp.WithHTTPClient(&http.Client{
			Timeout: config.Timeout,
		}))
	}

	if config.MaxRetries > 0 {
		httpClientOpts = append(httpClientOpts, sfhttp.WithRetry(
			config.MaxRetries,
			config.RetryWaitMin,
			config.RetryWaitMax,
		))
	}

	if config.Logger != nil {
		httpClientOpts = append(httpClientOpts, sfhttp.WithLogger(config.Logger))
		httpClientOpts = append(httpClientOpts, sfhttp.WithMiddleware(
			sfhttp.LoggingMiddleware(config.Logger),
		))
	}

	// Set token provider
	httpClientOpts = append(httpClientOpts, sfhttp.WithTokenProvider(client))

	client.httpClient = sfhttp.NewClient(httpClientOpts...)

	// Setup authenticator based on config
	if err := client.setupAuthenticator(); err != nil {
		return nil, err
	}

	// Setup services
	client.setupServices()

	return client, nil
}

// setupAuthenticator configures the appropriate authenticator.
func (c *SalesforceClient) setupAuthenticator() error {
	httpClient := c.config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: c.config.Timeout}
	}

	// Check for refresh token first
	if c.config.RefreshToken != "" {
		tokenURL := c.config.TokenURL
		if tokenURL == "" {
			tokenURL = c.config.BaseURL + "/services/oauth2/token"
		}
		c.authenticator = auth.NewOAuthRefreshAuthenticator(
			c.config.ClientID,
			c.config.ClientSecret,
			c.config.RefreshToken,
			tokenURL,
			httpClient,
		)
		return nil
	}

	// Check for password auth
	if c.config.Username != "" && c.config.Password != "" {
		c.authenticator = auth.NewPasswordAuthenticator(
			c.config.Username,
			c.config.Password,
			c.config.SecurityToken,
			c.config.ClientID,
			c.config.ClientSecret,
			c.config.BaseURL,
			httpClient,
		)
		return nil
	}

	// Check for pre-existing token (via instance URL)
	if c.config.InstanceURL != "" {
		c.authenticator = auth.NewTokenAuthenticator("", c.config.InstanceURL)
		return nil
	}

	return fmt.Errorf("no authentication method configured")
}

// setupServices initializes all service implementations.
func (c *SalesforceClient) setupServices() {
	c.sobjects = services.NewSObjectService(c.httpClient, c.config.APIVersion)
	c.query = services.NewQueryService(c.httpClient, c.config.APIVersion)
	c.bulk = services.NewBulkService(c.httpClient, c.config.APIVersion)
	c.tooling = services.NewToolingService(c.httpClient, c.config.APIVersion)
	c.apex = services.NewApexService(c.httpClient, c.config.APIVersion)
}

// Connect authenticates and establishes connection to Salesforce.
func (c *SalesforceClient) Connect(ctx context.Context) error {
	token, err := c.authenticator.Authenticate(ctx)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	c.mu.Lock()
	c.token = token
	c.httpClient.SetBaseURL(token.InstanceURL)
	c.httpClient.SetAccessToken(token.AccessToken)
	c.mu.Unlock()

	return nil
}

// GetToken returns the current token (implements TokenProvider).
func (c *SalesforceClient) GetToken() *types.Token {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.token
}

// RefreshToken refreshes the access token (implements TokenProvider).
func (c *SalesforceClient) RefreshToken(ctx context.Context) error {
	token, err := c.authenticator.Refresh(ctx)
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.token = token
	c.httpClient.SetAccessToken(token.AccessToken)
	if token.InstanceURL != "" {
		c.httpClient.SetBaseURL(token.InstanceURL)
	}
	c.mu.Unlock()

	return nil
}

// SetAccessToken sets the access token directly.
func (c *SalesforceClient) SetAccessToken(accessToken, instanceURL string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.token = &types.Token{
		AccessToken: accessToken,
		InstanceURL: instanceURL,
		TokenType:   "Bearer",
		IssuedAt:    time.Now(),
	}

	c.httpClient.SetAccessToken(accessToken)
	c.httpClient.SetBaseURL(instanceURL)

	// Update authenticator if using token auth
	if tokenAuth, ok := c.authenticator.(*auth.TokenAuthenticator); ok {
		tokenAuth.SetToken(c.token)
	}
}

// SObjects returns the SObject service.
func (c *SalesforceClient) SObjects() SObjectService {
	return c.sobjects
}

// Query returns the Query service.
func (c *SalesforceClient) Query() QueryService {
	return c.query
}

// Bulk returns the Bulk API service.
func (c *SalesforceClient) Bulk() BulkService {
	return c.bulk
}

// Tooling returns the Tooling API service.
func (c *SalesforceClient) Tooling() ToolingService {
	return c.tooling
}

// Apex returns the Apex REST service.
func (c *SalesforceClient) Apex() ApexService {
	return c.apex
}

// InstanceURL returns the Salesforce instance URL.
func (c *SalesforceClient) InstanceURL() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.token != nil {
		return c.token.InstanceURL
	}
	return c.config.InstanceURL
}

// APIVersion returns the API version being used.
func (c *SalesforceClient) APIVersion() string {
	return c.config.APIVersion
}

// Close cleans up any resources held by the client.
func (c *SalesforceClient) Close() error {
	return nil
}

// GetLimits retrieves the current API limits for the org.
func (c *SalesforceClient) GetLimits(ctx context.Context) (*Limits, error) {
	path := fmt.Sprintf("/services/data/v%s/limits", c.config.APIVersion)

	respBody, err := c.httpClient.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get limits: %w", err)
	}

	var limits Limits
	if err := json.Unmarshal(respBody, &limits); err != nil {
		return nil, fmt.Errorf("failed to parse limits: %w", err)
	}

	return &limits, nil
}

// GetVersions retrieves available API versions.
func (c *SalesforceClient) GetVersions(ctx context.Context) ([]APIVersionInfo, error) {
	respBody, err := c.httpClient.Get(ctx, "/services/data/")
	if err != nil {
		return nil, fmt.Errorf("failed to get versions: %w", err)
	}

	var versions []APIVersionInfo
	if err := json.Unmarshal(respBody, &versions); err != nil {
		return nil, fmt.Errorf("failed to parse versions: %w", err)
	}

	return versions, nil
}

// APIVersionInfo represents a Salesforce API version.
type APIVersionInfo struct {
	Label   string `json:"label"`
	URL     string `json:"url"`
	Version string `json:"version"`
}

// ExecuteComposite executes a composite request.
func (c *SalesforceClient) ExecuteComposite(ctx context.Context, req CompositeRequest) (*CompositeResponse, error) {
	path := fmt.Sprintf("/services/data/v%s/composite", c.config.APIVersion)

	respBody, err := c.httpClient.Post(ctx, path, req)
	if err != nil {
		return nil, fmt.Errorf("composite request failed: %w", err)
	}

	var resp CompositeResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse composite response: %w", err)
	}

	return &resp, nil
}

// ApexREST makes a request to a custom Apex REST endpoint.
func (c *SalesforceClient) ApexREST(ctx context.Context, method, path string, body io.Reader) ([]byte, error) {
	return c.apex.Execute(ctx, method, path, body)
}
