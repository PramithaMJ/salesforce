// Package sobjects provides SObject CRUD operations for Salesforce.
package sobjects

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// SObject represents a Salesforce SObject record.
type SObject struct {
	data map[string]interface{}
}

// Attributes contains SObject metadata.
type Attributes struct {
	Type string `json:"type"`
	URL  string `json:"url,omitempty"`
}

// New creates a new SObject of the specified type.
func New(objectType string) *SObject {
	return &SObject{
		data: map[string]interface{}{
			"attributes": Attributes{Type: objectType},
		},
	}
}

// FromMap creates an SObject from a map.
func FromMap(data map[string]interface{}) *SObject {
	if data == nil {
		data = make(map[string]interface{})
	}
	return &SObject{data: data}
}

// Type returns the SObject type.
func (s *SObject) Type() string {
	if attrs := s.Attributes(); attrs != nil {
		return attrs.Type
	}
	return ""
}

// ID returns the record ID.
func (s *SObject) ID() string {
	return s.StringField("Id")
}

// Attributes returns the SObject attributes.
func (s *SObject) Attributes() *Attributes {
	if s.data == nil {
		return nil
	}
	switch v := s.data["attributes"].(type) {
	case Attributes:
		return &v
	case *Attributes:
		return v
	case map[string]interface{}:
		attrs := &Attributes{}
		if t, ok := v["type"].(string); ok {
			attrs.Type = t
		}
		if u, ok := v["url"].(string); ok {
			attrs.URL = u
		}
		return attrs
	}
	return nil
}

// Get returns a field value.
func (s *SObject) Get(key string) interface{} {
	if s.data == nil {
		return nil
	}
	return s.data[key]
}

// Set sets a field value.
func (s *SObject) Set(key string, value interface{}) *SObject {
	if s.data == nil {
		s.data = make(map[string]interface{})
	}
	s.data[key] = value
	return s
}

// StringField returns a field as string.
func (s *SObject) StringField(key string) string {
	if v, ok := s.Get(key).(string); ok {
		return v
	}
	return ""
}

// IntField returns a field as int.
func (s *SObject) IntField(key string) int {
	switch v := s.Get(key).(type) {
	case float64:
		return int(v)
	case int:
		return v
	case int64:
		return int(v)
	}
	return 0
}

// FloatField returns a field as float64.
func (s *SObject) FloatField(key string) float64 {
	switch v := s.Get(key).(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	}
	return 0
}

// BoolField returns a field as bool.
func (s *SObject) BoolField(key string) bool {
	if v, ok := s.Get(key).(bool); ok {
		return v
	}
	return false
}

// TimeField returns a field as time.Time.
func (s *SObject) TimeField(key string) time.Time {
	if v, ok := s.Get(key).(string); ok {
		t, _ := time.Parse(time.RFC3339, v)
		return t
	}
	return time.Time{}
}

// Related returns a related SObject.
func (s *SObject) Related(key string) *SObject {
	if v, ok := s.Get(key).(map[string]interface{}); ok {
		return FromMap(v)
	}
	return nil
}

// RelatedList returns a list of related SObjects.
func (s *SObject) RelatedList(key string) []*SObject {
	v := s.Get(key)
	if v == nil {
		return nil
	}
	if m, ok := v.(map[string]interface{}); ok {
		if records, ok := m["records"].([]interface{}); ok {
			result := make([]*SObject, len(records))
			for i, r := range records {
				if rm, ok := r.(map[string]interface{}); ok {
					result[i] = FromMap(rm)
				}
			}
			return result
		}
	}
	return nil
}

// ToMap returns the SObject as a map.
func (s *SObject) ToMap() map[string]interface{} {
	result := make(map[string]interface{}, len(s.data))
	for k, v := range s.data {
		result[k] = v
	}
	return result
}

// ToCreatePayload returns fields suitable for create/update.
func (s *SObject) ToCreatePayload() map[string]interface{} {
	systemFields := map[string]bool{
		"Id": true, "attributes": true, "IsDeleted": true,
		"CreatedDate": true, "CreatedById": true,
		"LastModifiedDate": true, "LastModifiedById": true,
		"SystemModstamp": true, "LastActivityDate": true,
		"LastViewedDate": true, "LastReferencedDate": true,
	}
	result := make(map[string]interface{})
	for k, v := range s.data {
		if !systemFields[k] {
			result[k] = v
		}
	}
	return result
}

// MarshalJSON implements json.Marshaler.
func (s *SObject) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.data)
}

// UnmarshalJSON implements json.Unmarshaler.
func (s *SObject) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &s.data); err != nil {
		return err
	}
	return nil
}

// Metadata contains SObject describe information.
type Metadata struct {
	Name               string           `json:"name"`
	Label              string           `json:"label"`
	LabelPlural        string           `json:"labelPlural"`
	KeyPrefix          string           `json:"keyPrefix"`
	Createable         bool             `json:"createable"`
	Updateable         bool             `json:"updateable"`
	Deletable          bool             `json:"deletable"`
	Queryable          bool             `json:"queryable"`
	Searchable         bool             `json:"searchable"`
	Retrieveable       bool             `json:"retrieveable"`
	Undeletable        bool             `json:"undeletable"`
	Mergeable          bool             `json:"mergeable"`
	Replicateable      bool             `json:"replicateable"`
	Triggerable        bool             `json:"triggerable"`
	FeedEnabled        bool             `json:"feedEnabled"`
	HasSubtypes        bool             `json:"hasSubtypes"`
	IsSubtype          bool             `json:"isSubtype"`
	Custom             bool             `json:"custom"`
	CustomSetting      bool             `json:"customSetting"`
	Fields             []FieldMetadata  `json:"fields,omitempty"`
	ChildRelationships []ChildRelation  `json:"childRelationships,omitempty"`
	RecordTypeInfos    []RecordTypeInfo `json:"recordTypeInfos,omitempty"`
}

// FieldMetadata describes a field.
type FieldMetadata struct {
	Name             string          `json:"name"`
	Label            string          `json:"label"`
	Type             string          `json:"type"`
	Length           int             `json:"length"`
	Precision        int             `json:"precision"`
	Scale            int             `json:"scale"`
	Createable       bool            `json:"createable"`
	Updateable       bool            `json:"updateable"`
	Nillable         bool            `json:"nillable"`
	Unique           bool            `json:"unique"`
	Custom           bool            `json:"custom"`
	ExternalId       bool            `json:"externalId"`
	AutoNumber       bool            `json:"autoNumber"`
	Calculated       bool            `json:"calculated"`
	NameField        bool            `json:"nameField"`
	IdLookup         bool            `json:"idLookup"`
	DefaultValue     interface{}     `json:"defaultValue"`
	ReferenceTo      []string        `json:"referenceTo,omitempty"`
	RelationshipName string          `json:"relationshipName,omitempty"`
	PicklistValues   []PicklistValue `json:"picklistValues,omitempty"`
}

// PicklistValue represents a picklist option.
type PicklistValue struct {
	Active       bool   `json:"active"`
	DefaultValue bool   `json:"defaultValue"`
	Label        string `json:"label"`
	Value        string `json:"value"`
}

// ChildRelation describes a child relationship.
type ChildRelation struct {
	ChildSObject        string `json:"childSObject"`
	Field               string `json:"field"`
	RelationshipName    string `json:"relationshipName"`
	CascadeDelete       bool   `json:"cascadeDelete"`
	RestrictedDelete    bool   `json:"restrictedDelete"`
	DeprecatedAndHidden bool   `json:"deprecatedAndHidden"`
}

// RecordTypeInfo describes a record type.
type RecordTypeInfo struct {
	Name                     string `json:"name"`
	RecordTypeId             string `json:"recordTypeId"`
	Available                bool   `json:"available"`
	DefaultRecordTypeMapping bool   `json:"defaultRecordTypeMapping"`
	Master                   bool   `json:"master"`
}

// GlobalDescribe contains all accessible SObjects.
type GlobalDescribe struct {
	Encoding     string     `json:"encoding"`
	MaxBatchSize int        `json:"maxBatchSize"`
	SObjects     []Metadata `json:"sobjects"`
}

// DeletedRecords contains deleted record information.
type DeletedRecords struct {
	DeletedRecords        []DeletedRecord `json:"deletedRecords"`
	EarliestDateAvailable string          `json:"earliestDateAvailable"`
	LatestDateCovered     string          `json:"latestDateCovered"`
}

// DeletedRecord represents a deleted record.
type DeletedRecord struct {
	ID          string `json:"id"`
	DeletedDate string `json:"deletedDate"`
}

// UpdatedRecords contains updated record IDs.
type UpdatedRecords struct {
	IDs               []string `json:"ids"`
	LatestDateCovered string   `json:"latestDateCovered"`
}

// CreateResult contains the result of a create operation.
type CreateResult struct {
	ID      string  `json:"id"`
	Success bool    `json:"success"`
	Errors  []Error `json:"errors,omitempty"`
}

// Error represents an operation error.
type Error struct {
	StatusCode string   `json:"statusCode"`
	Message    string   `json:"message"`
	Fields     []string `json:"fields,omitempty"`
}

// HTTPClient interface for dependency injection.
type HTTPClient interface {
	Get(ctx context.Context, path string) ([]byte, error)
	Post(ctx context.Context, path string, body interface{}) ([]byte, error)
	Patch(ctx context.Context, path string, body interface{}) ([]byte, error)
	Put(ctx context.Context, path string, body interface{}) ([]byte, error)
	Delete(ctx context.Context, path string) ([]byte, error)
}

// Service provides SObject CRUD operations.
type Service struct {
	client     HTTPClient
	apiVersion string
}

// NewService creates a new SObject service.
func NewService(client HTTPClient, apiVersion string) *Service {
	return &Service{client: client, apiVersion: apiVersion}
}

// Create creates a new SObject record.
func (s *Service) Create(ctx context.Context, objectType string, data map[string]interface{}) (*CreateResult, error) {
	path := fmt.Sprintf("/services/data/v%s/sobjects/%s", s.apiVersion, objectType)
	respBody, err := s.client.Post(ctx, path, data)
	if err != nil {
		return nil, err
	}
	var result CreateResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &result, nil
}

// Get retrieves an SObject by ID.
func (s *Service) Get(ctx context.Context, objectType, id string, fields ...string) (*SObject, error) {
	path := fmt.Sprintf("/services/data/v%s/sobjects/%s/%s", s.apiVersion, objectType, id)
	if len(fields) > 0 {
		path += "?fields=" + url.QueryEscape(strings.Join(fields, ","))
	}
	respBody, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var data map[string]interface{}
	if err := json.Unmarshal(respBody, &data); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return FromMap(data), nil
}

// Update updates an existing SObject.
func (s *Service) Update(ctx context.Context, objectType, id string, data map[string]interface{}) error {
	path := fmt.Sprintf("/services/data/v%s/sobjects/%s/%s", s.apiVersion, objectType, id)
	_, err := s.client.Patch(ctx, path, data)
	return err
}

// Upsert upserts an SObject by external ID.
func (s *Service) Upsert(ctx context.Context, objectType, extIDField, extID string, data map[string]interface{}) (*CreateResult, error) {
	path := fmt.Sprintf("/services/data/v%s/sobjects/%s/%s/%s", s.apiVersion, objectType, extIDField, url.PathEscape(extID))
	respBody, err := s.client.Patch(ctx, path, data)
	if err != nil {
		return nil, err
	}
	if len(respBody) == 0 {
		return &CreateResult{Success: true}, nil
	}
	var result CreateResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &result, nil
}

// Delete deletes an SObject by ID.
func (s *Service) Delete(ctx context.Context, objectType, id string) error {
	path := fmt.Sprintf("/services/data/v%s/sobjects/%s/%s", s.apiVersion, objectType, id)
	_, err := s.client.Delete(ctx, path)
	return err
}

// Describe returns metadata for an SObject type.
func (s *Service) Describe(ctx context.Context, objectType string) (*Metadata, error) {
	path := fmt.Sprintf("/services/data/v%s/sobjects/%s/describe", s.apiVersion, objectType)
	respBody, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var meta Metadata
	if err := json.Unmarshal(respBody, &meta); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}
	return &meta, nil
}

// DescribeGlobal returns all accessible SObject types.
func (s *Service) DescribeGlobal(ctx context.Context) (*GlobalDescribe, error) {
	path := fmt.Sprintf("/services/data/v%s/sobjects", s.apiVersion)
	respBody, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var result GlobalDescribe
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &result, nil
}

// GetDeleted retrieves deleted records for an SObject type.
func (s *Service) GetDeleted(ctx context.Context, objectType string, start, end time.Time) (*DeletedRecords, error) {
	path := fmt.Sprintf("/services/data/v%s/sobjects/%s/deleted/?start=%s&end=%s",
		s.apiVersion, objectType,
		url.QueryEscape(start.Format(time.RFC3339)),
		url.QueryEscape(end.Format(time.RFC3339)))
	respBody, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var result DeletedRecords
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &result, nil
}

// GetUpdated retrieves updated record IDs for an SObject type.
func (s *Service) GetUpdated(ctx context.Context, objectType string, start, end time.Time) (*UpdatedRecords, error) {
	path := fmt.Sprintf("/services/data/v%s/sobjects/%s/updated/?start=%s&end=%s",
		s.apiVersion, objectType,
		url.QueryEscape(start.Format(time.RFC3339)),
		url.QueryEscape(end.Format(time.RFC3339)))
	respBody, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var result UpdatedRecords
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &result, nil
}

// GetByExternalID retrieves an SObject by external ID.
func (s *Service) GetByExternalID(ctx context.Context, objectType, extIDField, extID string) (*SObject, error) {
	path := fmt.Sprintf("/services/data/v%s/sobjects/%s/%s/%s", s.apiVersion, objectType, extIDField, url.PathEscape(extID))
	respBody, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var data map[string]interface{}
	if err := json.Unmarshal(respBody, &data); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return FromMap(data), nil
}
