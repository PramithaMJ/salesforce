package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/PramithaMJ/salesforce/types"
)

type Client struct {
	httpClient    *http.Client
	baseURL       string
	apiVersion    string
	accessToken   string
	tokenProvider TokenProvider
	logger        types.Logger
	middleware    []Middleware
	maxRetries    int
	retryWaitMin  time.Duration
	retryWaitMax  time.Duration
	mu            sync.RWMutex
}

type TokenProvider interface {
	GetToken() *types.Token
	RefreshToken(ctx context.Context) error
}

type Middleware func(next RoundTripperFunc) RoundTripperFunc
type RoundTripperFunc func(*http.Request) (*http.Response, error)

func (f RoundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type ClientOption func(*Client)

func NewClient(opts ...ClientOption) *Client {
	c := &Client{
		httpClient:   &http.Client{Timeout: 30 * time.Second},
		apiVersion:   types.DefaultAPIVersion,
		maxRetries:   3,
		retryWaitMin: 1 * time.Second,
		retryWaitMax: 30 * time.Second,
		middleware:   make([]Middleware, 0),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) { c.httpClient = client }
}

func WithBaseURL(url string) ClientOption {
	return func(c *Client) { c.baseURL = url }
}

func WithAPIVersion(version string) ClientOption {
	return func(c *Client) { c.apiVersion = version }
}

func WithAccessToken(token string) ClientOption {
	return func(c *Client) { c.accessToken = token }
}

func WithTokenProvider(provider TokenProvider) ClientOption {
	return func(c *Client) { c.tokenProvider = provider }
}

func WithLogger(logger types.Logger) ClientOption {
	return func(c *Client) { c.logger = logger }
}

func WithMiddleware(middleware ...Middleware) ClientOption {
	return func(c *Client) { c.middleware = append(c.middleware, middleware...) }
}

func WithRetry(maxRetries int, waitMin, waitMax time.Duration) ClientOption {
	return func(c *Client) {
		c.maxRetries = maxRetries
		c.retryWaitMin = waitMin
		c.retryWaitMax = waitMax
	}
}

func (c *Client) SetBaseURL(url string)      { c.mu.Lock(); c.baseURL = url; c.mu.Unlock() }
func (c *Client) SetAccessToken(token string) { c.mu.Lock(); c.accessToken = token; c.mu.Unlock() }

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	handler := RoundTripperFunc(c.httpClient.Do)
	for i := len(c.middleware) - 1; i >= 0; i-- {
		handler = c.middleware[i](handler)
	}
	return handler(req)
}

func (c *Client) Request(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		resp, err := c.doRequest(ctx, method, path, body)
		if err != nil {
			lastErr = err
			if !isRetryable(err) {
				return nil, err
			}
			if attempt < c.maxRetries {
				wait := c.calculateBackoff(attempt)
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(wait):
				}
			}
			continue
		}
		return resp, nil
	}
	return nil, fmt.Errorf("request failed after %d attempts: %w", c.maxRetries+1, lastErr)
}

func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	c.mu.RLock()
	baseURL, token := c.baseURL, c.accessToken
	c.mu.RUnlock()

	if c.tokenProvider != nil {
		if t := c.tokenProvider.GetToken(); t != nil {
			token = t.AccessToken
			if t.InstanceURL != "" {
				baseURL = t.InstanceURL
			}
		}
	}

	var bodyReader io.Reader
	if body != nil {
		switch v := body.(type) {
		case io.Reader:
			bodyReader = v
		case []byte:
			bodyReader = bytes.NewReader(v)
		case string:
			bodyReader = bytes.NewReader([]byte(v))
		default:
			jsonBody, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal body: %w", err)
			}
			bodyReader = bytes.NewReader(jsonBody)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if resp.StatusCode == http.StatusTooManyRequests {
			retryAfter := 60
			if ra := resp.Header.Get("Retry-After"); ra != "" {
				if s, e := strconv.Atoi(ra); e == nil {
					retryAfter = s
				}
			}
			return nil, &types.RateLimitError{RetryAfter: retryAfter, Message: string(respBody)}
		}
		if resp.StatusCode == http.StatusUnauthorized && c.tokenProvider != nil {
			if err := c.tokenProvider.RefreshToken(ctx); err == nil {
				return c.doRequest(ctx, method, path, body)
			}
		}
		return nil, types.ParseAPIError(resp.StatusCode, respBody)
	}
	return respBody, nil
}

func (c *Client) Get(ctx context.Context, path string) ([]byte, error) {
	return c.Request(ctx, http.MethodGet, path, nil)
}
func (c *Client) Post(ctx context.Context, path string, body interface{}) ([]byte, error) {
	return c.Request(ctx, http.MethodPost, path, body)
}
func (c *Client) Patch(ctx context.Context, path string, body interface{}) ([]byte, error) {
	return c.Request(ctx, http.MethodPatch, path, body)
}
func (c *Client) Put(ctx context.Context, path string, body interface{}) ([]byte, error) {
	return c.Request(ctx, http.MethodPut, path, body)
}
func (c *Client) Delete(ctx context.Context, path string) ([]byte, error) {
	return c.Request(ctx, http.MethodDelete, path, nil)
}

func (c *Client) calculateBackoff(attempt int) time.Duration {
	wait := float64(c.retryWaitMin) * math.Pow(2, float64(attempt))
	jitter := wait * 0.2 * (rand.Float64()*2 - 1)
	if time.Duration(wait+jitter) > c.retryWaitMax {
		return c.retryWaitMax
	}
	return time.Duration(wait + jitter)
}

func isRetryable(err error) bool {
	if _, ok := err.(*types.RateLimitError); ok {
		return true
	}
	if apiErr, ok := err.(*types.APIError); ok {
		return apiErr.IsRetryable()
	}
	return false
}

type DefaultLogger struct{ debug bool }

func NewDefaultLogger(debug bool) *DefaultLogger { return &DefaultLogger{debug: debug} }
func (l *DefaultLogger) Debug(msg string, args ...interface{}) {
	if l.debug {
		fmt.Printf("[DEBUG] "+msg+"\n", args...)
	}
}
func (l *DefaultLogger) Info(msg string, args ...interface{})  { fmt.Printf("[INFO] "+msg+"\n", args...) }
func (l *DefaultLogger) Warn(msg string, args ...interface{})  { fmt.Printf("[WARN] "+msg+"\n", args...) }
func (l *DefaultLogger) Error(msg string, args ...interface{}) { fmt.Printf("[ERROR] "+msg+"\n", args...) }

func LoggingMiddleware(logger types.Logger) Middleware {
	return func(next RoundTripperFunc) RoundTripperFunc {
		return func(req *http.Request) (*http.Response, error) {
			start := time.Now()
			logger.Debug("Request: %s %s", req.Method, req.URL.String())
			resp, err := next(req)
			if err != nil {
				logger.Error("Request failed: %s %s - %v (%s)", req.Method, req.URL.String(), err, time.Since(start))
			} else {
				logger.Debug("Response: %s %s - %d (%s)", req.Method, req.URL.String(), resp.StatusCode, time.Since(start))
			}
			return resp, err
		}
	}
}
