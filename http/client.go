// Package http provides HTTP client functionality.
package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/PramithaMJ/salesforce/types"
)

// Client provides HTTP operations for Salesforce API.
type Client struct {
	httpClient  *http.Client
	baseURL     string
	accessToken string
	apiVersion  string
	logger      types.Logger
	maxRetries  int
	retryDelay  time.Duration
}

// Config holds HTTP client configuration.
type Config struct {
	HTTPClient *http.Client
	BaseURL    string
	APIVersion string
	Logger     types.Logger
	MaxRetries int
	RetryDelay time.Duration
}

// NewClient creates a new HTTP client.
func NewClient(cfg Config) *Client {
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: 30 * time.Second}
	}
	if cfg.APIVersion == "" {
		cfg.APIVersion = types.DefaultAPIVersion
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = types.DefaultMaxRetries
	}
	if cfg.RetryDelay == 0 {
		cfg.RetryDelay = 1 * time.Second
	}
	return &Client{
		httpClient: cfg.HTTPClient,
		baseURL:    strings.TrimSuffix(cfg.BaseURL, "/"),
		apiVersion: cfg.APIVersion,
		logger:     cfg.Logger,
		maxRetries: cfg.MaxRetries,
		retryDelay: cfg.RetryDelay,
	}
}

// SetBaseURL sets the base URL.
func (c *Client) SetBaseURL(url string) {
	c.baseURL = strings.TrimSuffix(url, "/")
}

// SetAccessToken sets the access token.
func (c *Client) SetAccessToken(token string) {
	c.accessToken = token
}

// Get performs a GET request.
func (c *Client) Get(ctx context.Context, path string) ([]byte, error) {
	return c.doRequest(ctx, http.MethodGet, path, nil, "")
}

// Post performs a POST request.
func (c *Client) Post(ctx context.Context, path string, body interface{}) ([]byte, error) {
	return c.doRequest(ctx, http.MethodPost, path, body, "application/json")
}

// Patch performs a PATCH request.
func (c *Client) Patch(ctx context.Context, path string, body interface{}) ([]byte, error) {
	return c.doRequest(ctx, http.MethodPatch, path, body, "application/json")
}

// Put performs a PUT request.
func (c *Client) Put(ctx context.Context, path string, body interface{}) ([]byte, error) {
	contentType := "application/json"
	if _, ok := body.(io.Reader); ok {
		contentType = "text/csv"
	}
	return c.doRequest(ctx, http.MethodPut, path, body, contentType)
}

// Delete performs a DELETE request.
func (c *Client) Delete(ctx context.Context, path string) ([]byte, error) {
	return c.doRequest(ctx, http.MethodDelete, path, nil, "")
}

func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}, contentType string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			delay := c.retryDelay * time.Duration(math.Pow(2, float64(attempt-1)))
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}
		respBody, err := c.executeRequest(ctx, method, path, body, contentType)
		if err == nil {
			return respBody, nil
		}
		if !types.IsRetryableError(err) {
			return nil, err
		}
		lastErr = err
		if c.logger != nil {
			c.logger.Warn("Request failed, retrying", "attempt", attempt+1, "error", err)
		}
	}
	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

func (c *Client) executeRequest(ctx context.Context, method, path string, body interface{}, contentType string) ([]byte, error) {
	url := c.baseURL + path
	if !strings.HasPrefix(path, "http") && strings.HasPrefix(path, "/services/") {
		url = c.baseURL + path
	} else if strings.HasPrefix(path, "http") {
		url = path
	}
	var reqBody io.Reader
	if body != nil {
		switch v := body.(type) {
		case io.Reader:
			reqBody = v
		case []byte:
			reqBody = bytes.NewReader(v)
		case string:
			reqBody = strings.NewReader(v)
		default:
			data, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal body: %w", err)
			}
			reqBody = bytes.NewReader(data)
		}
	}
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	req.Header.Set("Accept", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, types.ParseAPIError(resp.StatusCode, respBody)
	}
	return respBody, nil
}

// APIVersion returns the API version.
func (c *Client) APIVersion() string { return c.apiVersion }

// BaseURL returns the base URL.
func (c *Client) BaseURL() string { return c.baseURL }
