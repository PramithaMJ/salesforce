package salesforce

import (
	"errors"
	"net/http"
	"time"

	sfhttp "github.com/PramithaMJ/salesforce/http"
	"github.com/PramithaMJ/salesforce/types"
)

// Config holds the configuration for the Salesforce client.
type Config struct {
	// Authentication
	Username      string
	Password      string
	SecurityToken string
	ClientID      string
	ClientSecret  string
	RefreshToken  string
	TokenURL      string

	// Connection
	BaseURL     string
	InstanceURL string
	APIVersion  string
	Timeout     time.Duration

	// Retry settings
	MaxRetries   int
	RetryWaitMin time.Duration
	RetryWaitMax time.Duration

	// Advanced
	Logger     types.Logger
	HTTPClient *http.Client
	Debug      bool
}

// DefaultConfig returns a Config with default values.
func DefaultConfig() *Config {
	return &Config{
		BaseURL:      "https://login.salesforce.com",
		APIVersion:   types.DefaultAPIVersion,
		Timeout:      types.DefaultTimeout,
		MaxRetries:   types.DefaultMaxRetries,
		RetryWaitMin: 1 * time.Second,
		RetryWaitMax: 30 * time.Second,
	}
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	hasRefresh := c.RefreshToken != "" && c.ClientID != ""
	hasPassword := c.Username != "" && c.Password != ""
	hasToken := c.InstanceURL != ""

	if !hasRefresh && !hasPassword && !hasToken {
		return errors.New("at least one authentication method must be configured")
	}

	if c.APIVersion == "" {
		return errors.New("API version is required")
	}

	return nil
}

// Option is a functional option for configuring the client.
type Option func(*Config)

// WithOAuthRefresh configures OAuth 2.0 refresh token authentication.
func WithOAuthRefresh(clientID, clientSecret, refreshToken string) Option {
	return func(c *Config) {
		c.ClientID = clientID
		c.ClientSecret = clientSecret
		c.RefreshToken = refreshToken
	}
}

// WithPasswordAuth configures username/password authentication.
func WithPasswordAuth(username, password, securityToken string) Option {
	return func(c *Config) {
		c.Username = username
		c.Password = password
		c.SecurityToken = securityToken
	}
}

// WithCredentials sets the OAuth client credentials.
func WithCredentials(clientID, clientSecret string) Option {
	return func(c *Config) {
		c.ClientID = clientID
		c.ClientSecret = clientSecret
	}
}

// WithTokenURL sets the OAuth token endpoint URL.
func WithTokenURL(tokenURL string) Option {
	return func(c *Config) {
		c.TokenURL = tokenURL
	}
}

// WithInstanceURL sets the Salesforce instance URL.
func WithInstanceURL(instanceURL string) Option {
	return func(c *Config) {
		c.InstanceURL = instanceURL
	}
}

// WithBaseURL sets the base URL for authentication.
func WithBaseURL(baseURL string) Option {
	return func(c *Config) {
		c.BaseURL = baseURL
	}
}

// WithAPIVersion sets the Salesforce API version.
func WithAPIVersion(version string) Option {
	return func(c *Config) {
		c.APIVersion = version
	}
}

// WithTimeout sets the HTTP timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

// WithRetry configures retry behavior.
func WithRetry(maxRetries int, waitMin, waitMax time.Duration) Option {
	return func(c *Config) {
		c.MaxRetries = maxRetries
		c.RetryWaitMin = waitMin
		c.RetryWaitMax = waitMax
	}
}

// WithLogger sets the logger.
func WithLogger(logger types.Logger) Option {
	return func(c *Config) {
		c.Logger = logger
	}
}

// WithHTTPClient sets the HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Config) {
		c.HTTPClient = client
	}
}

// WithDebug enables debug logging.
func WithDebug(debug bool) Option {
	return func(c *Config) {
		c.Debug = debug
		if debug && c.Logger == nil {
			c.Logger = sfhttp.NewDefaultLogger(true)
		}
	}
}
