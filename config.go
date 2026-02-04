package salesforce

import (
	"errors"
	"net/http"
	"time"

	"github.com/PramithaMJ/salesforce/types"
)

// Config holds the configuration for the Salesforce client.
type Config struct {
	// Authentication
	ClientID      string
	ClientSecret  string
	Username      string
	Password      string
	SecurityToken string
	RefreshToken  string
	AccessToken   string
	TokenURL      string
	InstanceURL   string

	// Connection
	APIVersion string
	Timeout    time.Duration
	MaxRetries int
	HTTPClient *http.Client
	Logger     types.Logger
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	hasRefreshToken := c.RefreshToken != ""
	hasPasswordAuth := c.Username != "" && c.Password != ""
	hasDirectToken := c.AccessToken != ""

	if !hasRefreshToken && !hasPasswordAuth && !hasDirectToken {
		return errors.New("authentication required: provide refresh token, username/password, or access token")
	}
	if (hasRefreshToken || hasPasswordAuth) && c.ClientID == "" {
		return errors.New("client_id required for OAuth flows")
	}
	if hasDirectToken && c.InstanceURL == "" {
		return errors.New("instance_url required when using direct access token")
	}
	return nil
}

// Option configures the Salesforce client.
type Option func(*Config) error

// WithOAuthRefresh configures OAuth 2.0 refresh token authentication.
func WithOAuthRefresh(clientID, clientSecret, refreshToken string) Option {
	return func(c *Config) error {
		c.ClientID = clientID
		c.ClientSecret = clientSecret
		c.RefreshToken = refreshToken
		return nil
	}
}

// WithPasswordAuth configures username-password authentication.
func WithPasswordAuth(username, password, securityToken string) Option {
	return func(c *Config) error {
		c.Username = username
		c.Password = password
		c.SecurityToken = securityToken
		return nil
	}
}

// WithCredentials sets OAuth client credentials.
func WithCredentials(clientID, clientSecret string) Option {
	return func(c *Config) error {
		c.ClientID = clientID
		c.ClientSecret = clientSecret
		return nil
	}
}

// WithAccessToken sets a direct access token.
func WithAccessToken(accessToken, instanceURL string) Option {
	return func(c *Config) error {
		c.AccessToken = accessToken
		c.InstanceURL = instanceURL
		return nil
	}
}

// WithTokenURL sets the OAuth token endpoint URL.
func WithTokenURL(url string) Option {
	return func(c *Config) error {
		c.TokenURL = url
		return nil
	}
}

// WithInstanceURL sets the Salesforce instance URL.
func WithInstanceURL(url string) Option {
	return func(c *Config) error {
		c.InstanceURL = url
		return nil
	}
}

// WithAPIVersion sets the API version.
func WithAPIVersion(version string) Option {
	return func(c *Config) error {
		c.APIVersion = version
		return nil
	}
}

// WithTimeout sets the HTTP timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Config) error {
		c.Timeout = timeout
		return nil
	}
}

// WithMaxRetries sets maximum retry attempts.
func WithMaxRetries(retries int) Option {
	return func(c *Config) error {
		c.MaxRetries = retries
		return nil
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Config) error {
		c.HTTPClient = client
		return nil
	}
}

// WithLogger sets the logger.
func WithLogger(logger types.Logger) Option {
	return func(c *Config) error {
		c.Logger = logger
		return nil
	}
}

// WithSandbox configures for sandbox environment.
func WithSandbox() Option {
	return func(c *Config) error {
		c.TokenURL = "https://test.salesforce.com/services/oauth2/token"
		return nil
	}
}

// WithCustomDomain configures for a custom My Domain.
func WithCustomDomain(domain string) Option {
	return func(c *Config) error {
		c.TokenURL = "https://" + domain + ".my.salesforce.com/services/oauth2/token"
		c.InstanceURL = "https://" + domain + ".my.salesforce.com"
		return nil
	}
}
