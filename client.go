// Package salesforce provides a comprehensive Go SDK for Salesforce.
package salesforce

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/PramithaMJ/salesforce/analytics"
	"github.com/PramithaMJ/salesforce/apex"
	"github.com/PramithaMJ/salesforce/auth"
	"github.com/PramithaMJ/salesforce/bulk"
	"github.com/PramithaMJ/salesforce/composite"
	"github.com/PramithaMJ/salesforce/connect"
	sfhttp "github.com/PramithaMJ/salesforce/http"
	"github.com/PramithaMJ/salesforce/limits"
	"github.com/PramithaMJ/salesforce/query"
	"github.com/PramithaMJ/salesforce/search"
	"github.com/PramithaMJ/salesforce/sobjects"
	"github.com/PramithaMJ/salesforce/tooling"
	"github.com/PramithaMJ/salesforce/types"
	"github.com/PramithaMJ/salesforce/uiapi"
)

// Client is the main Salesforce API client.
type Client struct {
	config     *Config
	httpClient *sfhttp.Client
	auth       auth.Authenticator

	// Services
	sobjects   *sobjects.Service
	query      *query.Service
	bulk       *bulk.Service
	composite  *composite.Service
	analytics  *analytics.Service
	tooling    *tooling.Service
	connect    *connect.Service
	limits     *limits.Service
	uiapi      *uiapi.Service
	search     *search.Service
	apex       *apex.Service
}

// NewClient creates a new Salesforce client with the given options.
func NewClient(opts ...Option) (*Client, error) {
	cfg := &Config{
		APIVersion: types.DefaultAPIVersion,
		Timeout:    types.DefaultTimeout,
		MaxRetries: types.DefaultMaxRetries,
		TokenURL:   "https://login.salesforce.com/services/oauth2/token",
	}
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: cfg.Timeout}
	}
	client := &Client{config: cfg}

	// Create authenticator
	switch {
	case cfg.RefreshToken != "":
		client.auth = auth.NewRefreshTokenAuthenticator(
			cfg.ClientID, cfg.ClientSecret, cfg.RefreshToken, cfg.TokenURL)
	case cfg.Username != "" && cfg.Password != "":
		client.auth = auth.NewPasswordAuthenticator(
			cfg.ClientID, cfg.ClientSecret, cfg.Username,
			cfg.Password, cfg.SecurityToken, cfg.TokenURL)
	case cfg.AccessToken != "":
		client.auth = auth.NewTokenAuthenticator(cfg.AccessToken, cfg.InstanceURL)
	default:
		return nil, errors.New("no authentication method configured")
	}

	// Create HTTP client
	client.httpClient = sfhttp.NewClient(sfhttp.Config{
		HTTPClient: httpClient,
		APIVersion: cfg.APIVersion,
		Logger:     cfg.Logger,
		MaxRetries: cfg.MaxRetries,
	})
	return client, nil
}

// Connect authenticates and establishes connection to Salesforce.
func (c *Client) Connect(ctx context.Context) error {
	token, err := c.auth.Authenticate(ctx)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}
	c.httpClient.SetBaseURL(token.InstanceURL)
	c.httpClient.SetAccessToken(token.AccessToken)
	c.initServices()
	return nil
}

// SetAccessToken sets the access token directly.
func (c *Client) SetAccessToken(token, instanceURL string) {
	c.httpClient.SetBaseURL(instanceURL)
	c.httpClient.SetAccessToken(token)
	c.initServices()
}

func (c *Client) initServices() {
	apiVersion := c.config.APIVersion
	c.sobjects = sobjects.NewService(c.httpClient, apiVersion)
	c.query = query.NewService(c.httpClient, apiVersion)
	c.bulk = bulk.NewService(c.httpClient, apiVersion)
	c.composite = composite.NewService(c.httpClient, apiVersion)
	c.analytics = analytics.NewService(c.httpClient, apiVersion)
	c.tooling = tooling.NewService(c.httpClient, apiVersion)
	c.connect = connect.NewService(c.httpClient, apiVersion)
	c.limits = limits.NewService(c.httpClient, apiVersion)
	c.uiapi = uiapi.NewService(c.httpClient, apiVersion)
	c.search = search.NewService(c.httpClient, apiVersion)
	c.apex = apex.NewService(c.httpClient)
}

// Services access methods

// SObjects returns the SObject service.
func (c *Client) SObjects() *sobjects.Service { return c.sobjects }

// Query returns the Query service.
func (c *Client) Query() *query.Service { return c.query }

// Bulk returns the Bulk API 2.0 service.
func (c *Client) Bulk() *bulk.Service { return c.bulk }

// Composite returns the Composite API service.
func (c *Client) Composite() *composite.Service { return c.composite }

// Analytics returns the Analytics/Reports service.
func (c *Client) Analytics() *analytics.Service { return c.analytics }

// Tooling returns the Tooling API service.
func (c *Client) Tooling() *tooling.Service { return c.tooling }

// Connect returns the Connect/Chatter service.
func (c *Client) Connect() *connect.Service { return c.connect }

// Limits returns the Limits service.
func (c *Client) Limits() *limits.Service { return c.limits }

// UIAPI returns the User Interface API service.
func (c *Client) UIAPI() *uiapi.Service { return c.uiapi }

// Search returns the SOSL Search service.
func (c *Client) Search() *search.Service { return c.search }

// Apex returns the Apex REST service.
func (c *Client) Apex() *apex.Service { return c.apex }

// GetToken returns the current access token.
func (c *Client) GetToken() *types.Token { return c.auth.GetToken() }

// APIVersion returns the configured API version.
func (c *Client) APIVersion() string { return c.config.APIVersion }

// InstanceURL returns the Salesforce instance URL.
func (c *Client) InstanceURL() string { return c.httpClient.BaseURL() }

// RefreshToken refreshes the access token.
func (c *Client) RefreshToken(ctx context.Context) error {
	token, err := c.auth.Refresh(ctx)
	if err != nil {
		return err
	}
	c.httpClient.SetAccessToken(token.AccessToken)
	return nil
}
