// Package apex provides Apex REST endpoint operations.
package apex

import (
	"context"
	"encoding/json"
	"fmt"
)

// HTTPClient interface for dependency injection.
type HTTPClient interface {
	Get(ctx context.Context, path string) ([]byte, error)
	Post(ctx context.Context, path string, body interface{}) ([]byte, error)
	Patch(ctx context.Context, path string, body interface{}) ([]byte, error)
	Put(ctx context.Context, path string, body interface{}) ([]byte, error)
	Delete(ctx context.Context, path string) ([]byte, error)
}

// Service provides Apex REST endpoint operations.
type Service struct {
	client HTTPClient
}

// NewService creates a new Apex REST service.
func NewService(client HTTPClient) *Service {
	return &Service{client: client}
}

// Get calls GET on an Apex REST endpoint.
func (s *Service) Get(ctx context.Context, path string) ([]byte, error) {
	fullPath := fmt.Sprintf("/services/apexrest%s", ensureLeadingSlash(path))
	return s.client.Get(ctx, fullPath)
}

// GetJSON calls GET and unmarshals JSON response.
func (s *Service) GetJSON(ctx context.Context, path string, result interface{}) error {
	body, err := s.Get(ctx, path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, result); err != nil {
		return err
	}
	return nil
}

// Post calls POST on an Apex REST endpoint.
func (s *Service) Post(ctx context.Context, path string, body interface{}) ([]byte, error) {
	fullPath := fmt.Sprintf("/services/apexrest%s", ensureLeadingSlash(path))
	return s.client.Post(ctx, fullPath, body)
}

// PostJSON calls POST and unmarshals JSON response.
func (s *Service) PostJSON(ctx context.Context, path string, body, result interface{}) error {
	respBody, err := s.Post(ctx, path, body)
	if err != nil {
		return err
	}
	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return err
		}
	}
	return nil
}

// Patch calls PATCH on an Apex REST endpoint.
func (s *Service) Patch(ctx context.Context, path string, body interface{}) ([]byte, error) {
	fullPath := fmt.Sprintf("/services/apexrest%s", ensureLeadingSlash(path))
	return s.client.Patch(ctx, fullPath, body)
}

// Put calls PUT on an Apex REST endpoint.
func (s *Service) Put(ctx context.Context, path string, body interface{}) ([]byte, error) {
	fullPath := fmt.Sprintf("/services/apexrest%s", ensureLeadingSlash(path))
	return s.client.Put(ctx, fullPath, body)
}

// Delete calls DELETE on an Apex REST endpoint.
func (s *Service) Delete(ctx context.Context, path string) ([]byte, error) {
	fullPath := fmt.Sprintf("/services/apexrest%s", ensureLeadingSlash(path))
	return s.client.Delete(ctx, fullPath)
}

func ensureLeadingSlash(path string) string {
	if len(path) > 0 && path[0] != '/' {
		return "/" + path
	}
	return path
}
