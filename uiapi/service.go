// Package uiapi provides User Interface API operations.
package uiapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// RecordUI contains full record UI information.
type RecordUI struct {
	Layouts     map[string]LayoutRepresentation `json:"layouts"`
	ObjectInfos map[string]ObjectInfo           `json:"objectInfos"`
	Records     map[string]RecordRepresentation `json:"records"`
}

// RecordRepresentation represents a record.
type RecordRepresentation struct {
	ID             string                `json:"id"`
	APIName        string                `json:"apiName"`
	Fields         map[string]FieldValue `json:"fields"`
	RecordTypeId   string                `json:"recordTypeId,omitempty"`
	SystemModstamp string                `json:"systemModstamp"`
}

// FieldValue represents a field value.
type FieldValue struct {
	DisplayValue string      `json:"displayValue"`
	Value        interface{} `json:"value"`
}

// ObjectInfo contains object metadata.
type ObjectInfo struct {
	APIName     string               `json:"apiName"`
	Label       string               `json:"label"`
	LabelPlural string               `json:"labelPlural"`
	KeyPrefix   string               `json:"keyPrefix"`
	Fields      map[string]FieldInfo `json:"fields"`
	Createable  bool                 `json:"createable"`
	Updateable  bool                 `json:"updateable"`
	Deletable   bool                 `json:"deletable"`
}

// FieldInfo contains field metadata.
type FieldInfo struct {
	APIName    string `json:"apiName"`
	Label      string `json:"label"`
	DataType   string `json:"dataType"`
	Createable bool   `json:"createable"`
	Updateable bool   `json:"updateable"`
	Required   bool   `json:"required"`
}

// LayoutRepresentation contains layout information.
type LayoutRepresentation struct {
	ID         string          `json:"id"`
	Sections   []LayoutSection `json:"sections"`
	LayoutType string          `json:"layoutType"`
	Mode       string          `json:"mode"`
}

// LayoutSection represents a layout section.
type LayoutSection struct {
	Heading    string      `json:"heading"`
	Columns    int         `json:"columns"`
	UseHeading bool        `json:"useHeading"`
	LayoutRows []LayoutRow `json:"layoutRows"`
}

// LayoutRow represents a layout row.
type LayoutRow struct {
	LayoutItems []LayoutItem `json:"layoutItems"`
}

// LayoutItem represents a layout item.
type LayoutItem struct {
	Field       string `json:"field,omitempty"`
	Label       string `json:"label"`
	Editability string `json:"editability"`
}

// PicklistValues contains picklist values.
type PicklistValues struct {
	ETag                string                          `json:"eTag"`
	PicklistFieldValues map[string]PicklistFieldValue `json:"picklistFieldValues"`
}

// PicklistFieldValue contains values for a picklist field.
type PicklistFieldValue struct {
	DefaultValue *PicklistValue  `json:"defaultValue"`
	Values       []PicklistValue `json:"values"`
}

// PicklistValue represents a picklist option.
type PicklistValue struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// HTTPClient interface for dependency injection.
type HTTPClient interface {
	Get(ctx context.Context, path string) ([]byte, error)
	Post(ctx context.Context, path string, body interface{}) ([]byte, error)
	Patch(ctx context.Context, path string, body interface{}) ([]byte, error)
	Delete(ctx context.Context, path string) ([]byte, error)
}

// Service provides User Interface API operations.
type Service struct {
	client     HTTPClient
	apiVersion string
}

// NewService creates a new UI API service.
func NewService(client HTTPClient, apiVersion string) *Service {
	return &Service{client: client, apiVersion: apiVersion}
}

// GetRecordUI retrieves record UI data.
func (s *Service) GetRecordUI(ctx context.Context, recordIds []string) (*RecordUI, error) {
	path := fmt.Sprintf("/services/data/v%s/ui-api/record-ui/%s", s.apiVersion, strings.Join(recordIds, ","))
	respBody, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var ui RecordUI
	if err := json.Unmarshal(respBody, &ui); err != nil {
		return nil, fmt.Errorf("failed to parse: %w", err)
	}
	return &ui, nil
}

// GetRecord retrieves a single record.
func (s *Service) GetRecord(ctx context.Context, recordId string, fields []string) (*RecordRepresentation, error) {
	path := fmt.Sprintf("/services/data/v%s/ui-api/records/%s", s.apiVersion, recordId)
	if len(fields) > 0 {
		path += "?fields=" + strings.Join(fields, ",")
	}
	respBody, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var record RecordRepresentation
	if err := json.Unmarshal(respBody, &record); err != nil {
		return nil, fmt.Errorf("failed to parse: %w", err)
	}
	return &record, nil
}

// CreateRecord creates a new record.
func (s *Service) CreateRecord(ctx context.Context, objectAPIName string, fields map[string]interface{}) (*RecordRepresentation, error) {
	path := fmt.Sprintf("/services/data/v%s/ui-api/records", s.apiVersion)
	body := map[string]interface{}{"apiName": objectAPIName, "fields": fields}
	respBody, err := s.client.Post(ctx, path, body)
	if err != nil {
		return nil, err
	}
	var record RecordRepresentation
	if err := json.Unmarshal(respBody, &record); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &record, nil
}

// UpdateRecord updates a record.
func (s *Service) UpdateRecord(ctx context.Context, recordId string, fields map[string]interface{}) (*RecordRepresentation, error) {
	path := fmt.Sprintf("/services/data/v%s/ui-api/records/%s", s.apiVersion, recordId)
	respBody, err := s.client.Patch(ctx, path, map[string]interface{}{"fields": fields})
	if err != nil {
		return nil, err
	}
	var record RecordRepresentation
	if err := json.Unmarshal(respBody, &record); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &record, nil
}

// DeleteRecord deletes a record.
func (s *Service) DeleteRecord(ctx context.Context, recordId string) error {
	path := fmt.Sprintf("/services/data/v%s/ui-api/records/%s", s.apiVersion, recordId)
	_, err := s.client.Delete(ctx, path)
	return err
}

// GetObjectInfo retrieves object metadata.
func (s *Service) GetObjectInfo(ctx context.Context, objectAPIName string) (*ObjectInfo, error) {
	path := fmt.Sprintf("/services/data/v%s/ui-api/object-info/%s", s.apiVersion, objectAPIName)
	respBody, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var info ObjectInfo
	if err := json.Unmarshal(respBody, &info); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &info, nil
}

// GetPicklistValues retrieves picklist values.
func (s *Service) GetPicklistValues(ctx context.Context, objectAPIName, recordTypeId string) (*PicklistValues, error) {
	path := fmt.Sprintf("/services/data/v%s/ui-api/object-info/%s/picklist-values/%s", s.apiVersion, objectAPIName, recordTypeId)
	respBody, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var values PicklistValues
	if err := json.Unmarshal(respBody, &values); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &values, nil
}

// GetLayout retrieves layout information.
func (s *Service) GetLayout(ctx context.Context, objectAPIName, layoutType, mode string) (*LayoutRepresentation, error) {
	path := fmt.Sprintf("/services/data/v%s/ui-api/layout/%s", s.apiVersion, objectAPIName)
	params := url.Values{}
	if layoutType != "" {
		params.Set("layoutType", layoutType)
	}
	if mode != "" {
		params.Set("mode", mode)
	}
	if len(params) > 0 {
		path += "?" + params.Encode()
	}
	respBody, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var layout LayoutRepresentation
	if err := json.Unmarshal(respBody, &layout); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &layout, nil
}
