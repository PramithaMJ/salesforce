// Package search provides SOSL search operations.
package search

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// Result contains SOSL search results.
type Result struct {
	SearchRecords []SearchRecord `json:"searchRecords"`
}

// SearchRecord represents a search result record.
type SearchRecord struct {
	Attributes map[string]interface{} `json:"attributes"`
	ID         string                 `json:"Id"`
	Name       string                 `json:"Name,omitempty"`
}

// ParameterizedSearchRequest contains parameterized search parameters.
type ParameterizedSearchRequest struct {
	Query           string   `json:"q"`
	Fields          []string `json:"fields,omitempty"`
	SObjects        []SObjSpec `json:"sobjects,omitempty"`
	In              string   `json:"in,omitempty"`
	Limit           int      `json:"overallLimit,omitempty"`
	DefaultLimit    int      `json:"defaultLimit,omitempty"`
	DataCategories  []DataCategory `json:"dataCategories,omitempty"`
}

// SObjSpec specifies search scope for an object.
type SObjSpec struct {
	Name   string   `json:"name"`
	Fields []string `json:"fields,omitempty"`
	Limit  int      `json:"limit,omitempty"`
}

// DataCategory specifies a data category filter.
type DataCategory struct {
	Group      string   `json:"groupName"`
	Categories []string `json:"categories"`
}

// HTTPClient interface for dependency injection.
type HTTPClient interface {
	Get(ctx context.Context, path string) ([]byte, error)
	Post(ctx context.Context, path string, body interface{}) ([]byte, error)
}

// Service provides SOSL search operations.
type Service struct {
	client     HTTPClient
	apiVersion string
}

// NewService creates a new Search service.
func NewService(client HTTPClient, apiVersion string) *Service {
	return &Service{client: client, apiVersion: apiVersion}
}

// Execute runs a SOSL search query.
func (s *Service) Execute(ctx context.Context, sosl string) (*Result, error) {
	path := fmt.Sprintf("/services/data/v%s/search?q=%s", s.apiVersion, url.QueryEscape(sosl))
	respBody, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var result Result
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse: %w", err)
	}
	return &result, nil
}

// Parameterized runs a parameterized search.
func (s *Service) Parameterized(ctx context.Context, req ParameterizedSearchRequest) (*Result, error) {
	path := fmt.Sprintf("/services/data/v%s/parameterizedSearch", s.apiVersion)
	respBody, err := s.client.Post(ctx, path, req)
	if err != nil {
		return nil, err
	}
	var result Result
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse: %w", err)
	}
	return &result, nil
}

// Builder provides fluent SOSL query building.
type Builder struct {
	searchTerm   string
	returning    []string
	inScope      string
	withDivision string
	limit        int
}

// NewBuilder creates a new SOSL query builder.
func NewBuilder(searchTerm string) *Builder {
	return &Builder{searchTerm: searchTerm}
}

// Returning adds objects to return.
func (b *Builder) Returning(objects ...string) *Builder {
	b.returning = append(b.returning, objects...)
	return b
}

// ReturningWithFields adds object with specific fields.
func (b *Builder) ReturningWithFields(object string, fields ...string) *Builder {
	if len(fields) > 0 {
		b.returning = append(b.returning, fmt.Sprintf("%s(%s)", object, strings.Join(fields, ", ")))
	} else {
		b.returning = append(b.returning, object)
	}
	return b
}

// In sets the search scope (ALL, NAME, EMAIL, PHONE, SIDEBAR).
func (b *Builder) In(scope string) *Builder {
	b.inScope = scope
	return b
}

// WithDivision filters by division.
func (b *Builder) WithDivision(division string) *Builder {
	b.withDivision = division
	return b
}

// Limit sets maximum results.
func (b *Builder) Limit(limit int) *Builder {
	b.limit = limit
	return b
}

// Build generates the SOSL query string.
func (b *Builder) Build() string {
	var sb strings.Builder
	sb.WriteString("FIND {")
	sb.WriteString(escapeSOSL(b.searchTerm))
	sb.WriteString("}")
	if b.inScope != "" {
		sb.WriteString(" IN ")
		sb.WriteString(b.inScope)
		sb.WriteString(" FIELDS")
	}
	if len(b.returning) > 0 {
		sb.WriteString(" RETURNING ")
		sb.WriteString(strings.Join(b.returning, ", "))
	}
	if b.withDivision != "" {
		sb.WriteString(" WITH DIVISION = '")
		sb.WriteString(b.withDivision)
		sb.WriteString("'")
	}
	if b.limit > 0 {
		sb.WriteString(fmt.Sprintf(" LIMIT %d", b.limit))
	}
	return sb.String()
}

func escapeSOSL(s string) string {
	replacer := strings.NewReplacer(
		"\\", "\\\\", "'", "\\'", "\"", "\\\"",
		"?", "\\?", "&", "\\&", "|", "\\|",
		"!", "\\!", "{", "\\{", "}", "\\}",
		"[", "\\[", "]", "\\]", "(", "\\(",
		")", "\\)", "^", "\\^", "~", "\\~",
		"*", "\\*", ":", "\\:", "-", "\\-",
	)
	return replacer.Replace(s)
}
