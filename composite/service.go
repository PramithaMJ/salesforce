// Package composite provides Composite API operations.
package composite

import (
	"context"
	"encoding/json"
	"fmt"
)

// Request represents a composite API request.
type Request struct {
	AllOrNone          bool               `json:"allOrNone"`
	CollateSubrequests bool               `json:"collateSubrequests,omitempty"`
	CompositeRequest   []Subrequest       `json:"compositeRequest"`
}

// Subrequest represents a single subrequest in a composite request.
type Subrequest struct {
	Method      string                 `json:"method"`
	URL         string                 `json:"url"`
	ReferenceId string                 `json:"referenceId"`
	Body        interface{}            `json:"body,omitempty"`
	HTTPHeaders map[string]string      `json:"httpHeaders,omitempty"`
}

// Response represents a composite API response.
type Response struct {
	CompositeResponse []Subresponse `json:"compositeResponse"`
}

// Subresponse represents a single subresponse.
type Subresponse struct {
	Body           interface{}       `json:"body"`
	HTTPHeaders    map[string]string `json:"httpHeaders"`
	HTTPStatusCode int               `json:"httpStatusCode"`
	ReferenceId    string            `json:"referenceId"`
}

// IsSuccess returns true if the subresponse was successful.
func (s *Subresponse) IsSuccess() bool {
	return s.HTTPStatusCode >= 200 && s.HTTPStatusCode < 300
}

// BatchRequest represents a batch request.
type BatchRequest struct {
	BatchRequests []BatchSubrequest `json:"batchRequests"`
	HaltOnError   bool              `json:"haltOnError,omitempty"`
}

// BatchSubrequest represents a single batch subrequest.
type BatchSubrequest struct {
	Method      string      `json:"method"`
	URL         string      `json:"url"`
	RichInput   interface{} `json:"richInput,omitempty"`
}

// BatchResponse represents a batch response.
type BatchResponse struct {
	HasErrors bool               `json:"hasErrors"`
	Results   []BatchSubresponse `json:"results"`
}

// BatchSubresponse represents a single batch subresponse.
type BatchSubresponse struct {
	StatusCode int         `json:"statusCode"`
	Result     interface{} `json:"result"`
}

// TreeRequest represents an SObject Tree request.
type TreeRequest struct {
	Records []TreeRecord `json:"records"`
}

// TreeRecord represents a record in a tree request.
type TreeRecord struct {
	Attributes  TreeAttributes `json:"attributes"`
	ReferenceId string         `json:"referenceId"`
	Fields      map[string]interface{}
}

// TreeAttributes contains record attributes for tree requests.
type TreeAttributes struct {
	Type        string `json:"type"`
	ReferenceId string `json:"referenceId,omitempty"`
}

// MarshalJSON implements custom JSON marshaling for TreeRecord.
func (t TreeRecord) MarshalJSON() ([]byte, error) {
	m := make(map[string]interface{})
	m["attributes"] = t.Attributes
	for k, v := range t.Fields {
		m[k] = v
	}
	return json.Marshal(m)
}

// TreeResponse represents an SObject Tree response.
type TreeResponse struct {
	HasErrors bool         `json:"hasErrors"`
	Results   []TreeResult `json:"results"`
}

// TreeResult represents a single result in a tree response.
type TreeResult struct {
	ID          string  `json:"id"`
	ReferenceId string  `json:"referenceId"`
	Errors      []Error `json:"errors,omitempty"`
}

// Error represents an API error.
type Error struct {
	StatusCode string   `json:"statusCode"`
	Message    string   `json:"message"`
	Fields     []string `json:"fields,omitempty"`
}

// GraphRequest represents a Composite Graph request.
type GraphRequest struct {
	Graphs []Graph `json:"graphs"`
}

// Graph represents a single graph in a graph request.
type Graph struct {
	GraphId     string           `json:"graphId"`
	CompositeRequest []Subrequest `json:"compositeRequest"`
}

// GraphResponse represents a Composite Graph response.
type GraphResponse struct {
	Graphs []GraphResult `json:"graphs"`
}

// GraphResult represents a single graph result.
type GraphResult struct {
	GraphId           string        `json:"graphId"`
	IsSuccessful      bool          `json:"isSuccessful"`
	GraphResponse     Response      `json:"graphResponse"`
}

// CollectionRequest represents an SObject Collections request.
type CollectionRequest struct {
	AllOrNone bool          `json:"allOrNone"`
	Records   []interface{} `json:"records"`
}

// CollectionResponse represents an SObject Collections response.
type CollectionResponse []CollectionResult

// CollectionResult represents a single collection result.
type CollectionResult struct {
	ID      string  `json:"id"`
	Success bool    `json:"success"`
	Errors  []Error `json:"errors,omitempty"`
}

// HTTPClient interface for dependency injection.
type HTTPClient interface {
	Get(ctx context.Context, path string) ([]byte, error)
	Post(ctx context.Context, path string, body interface{}) ([]byte, error)
	Patch(ctx context.Context, path string, body interface{}) ([]byte, error)
	Delete(ctx context.Context, path string) ([]byte, error)
}

// Service provides Composite API operations.
type Service struct {
	client     HTTPClient
	apiVersion string
}

// NewService creates a new Composite service.
func NewService(client HTTPClient, apiVersion string) *Service {
	return &Service{client: client, apiVersion: apiVersion}
}

// Execute executes a composite request.
func (s *Service) Execute(ctx context.Context, req Request) (*Response, error) {
	path := fmt.Sprintf("/services/data/v%s/composite", s.apiVersion)
	respBody, err := s.client.Post(ctx, path, req)
	if err != nil {
		return nil, err
	}
	var resp Response
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &resp, nil
}

// ExecuteBatch executes a batch request.
func (s *Service) ExecuteBatch(ctx context.Context, req BatchRequest) (*BatchResponse, error) {
	path := fmt.Sprintf("/services/data/v%s/composite/batch", s.apiVersion)
	respBody, err := s.client.Post(ctx, path, req)
	if err != nil {
		return nil, err
	}
	var resp BatchResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &resp, nil
}

// CreateTree creates records using SObject Tree.
func (s *Service) CreateTree(ctx context.Context, objectType string, records []TreeRecord) (*TreeResponse, error) {
	path := fmt.Sprintf("/services/data/v%s/composite/tree/%s", s.apiVersion, objectType)
	req := TreeRequest{Records: records}
	respBody, err := s.client.Post(ctx, path, req)
	if err != nil {
		return nil, err
	}
	var resp TreeResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &resp, nil
}

// ExecuteGraph executes a composite graph request.
func (s *Service) ExecuteGraph(ctx context.Context, req GraphRequest) (*GraphResponse, error) {
	path := fmt.Sprintf("/services/data/v%s/composite/graph", s.apiVersion)
	respBody, err := s.client.Post(ctx, path, req)
	if err != nil {
		return nil, err
	}
	var resp GraphResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &resp, nil
}

// CreateCollection creates multiple records using SObject Collections.
func (s *Service) CreateCollection(ctx context.Context, records []interface{}, allOrNone bool) (CollectionResponse, error) {
	path := fmt.Sprintf("/services/data/v%s/composite/sobjects", s.apiVersion)
	req := CollectionRequest{AllOrNone: allOrNone, Records: records}
	respBody, err := s.client.Post(ctx, path, req)
	if err != nil {
		return nil, err
	}
	var resp CollectionResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return resp, nil
}

// UpdateCollection updates multiple records using SObject Collections.
func (s *Service) UpdateCollection(ctx context.Context, records []interface{}, allOrNone bool) (CollectionResponse, error) {
	path := fmt.Sprintf("/services/data/v%s/composite/sobjects", s.apiVersion)
	req := CollectionRequest{AllOrNone: allOrNone, Records: records}
	respBody, err := s.client.Patch(ctx, path, req)
	if err != nil {
		return nil, err
	}
	var resp CollectionResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return resp, nil
}

// DeleteCollection deletes multiple records using SObject Collections.
func (s *Service) DeleteCollection(ctx context.Context, ids []string, allOrNone bool) (CollectionResponse, error) {
	path := fmt.Sprintf("/services/data/v%s/composite/sobjects?ids=%s&allOrNone=%t",
		s.apiVersion, joinIDs(ids), allOrNone)
	respBody, err := s.client.Delete(ctx, path)
	if err != nil {
		return nil, err
	}
	var resp CollectionResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return resp, nil
}

// GetCollection retrieves multiple records using SObject Collections.
func (s *Service) GetCollection(ctx context.Context, objectType string, ids []string, fields []string) ([]map[string]interface{}, error) {
	path := fmt.Sprintf("/services/data/v%s/composite/sobjects/%s?ids=%s",
		s.apiVersion, objectType, joinIDs(ids))
	if len(fields) > 0 {
		path += "&fields=" + joinIDs(fields)
	}
	respBody, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var resp []map[string]interface{}
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return resp, nil
}

func joinIDs(ids []string) string {
	result := ""
	for i, id := range ids {
		if i > 0 {
			result += ","
		}
		result += id
	}
	return result
}
