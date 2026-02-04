// Package salesforce provides a comprehensive, production-grade Go SDK for Salesforce.
package salesforce

import (
	"context"
	"io"

	"github.com/PramithaMJ/salesforce/services"
	"github.com/PramithaMJ/salesforce/types"
)

// Re-export types for convenience
type (
	Token           = types.Token
	APIError        = types.APIError
	APIErrors       = types.APIErrors
	AuthError       = types.AuthError
	ValidationError = types.ValidationError
	RateLimitError  = types.RateLimitError
	Logger          = types.Logger
	ErrorCode       = types.ErrorCode
)

// Re-export constants
const (
	DefaultAPIVersion = types.DefaultAPIVersion
	DefaultTimeout    = types.DefaultTimeout
	DefaultMaxRetries = types.DefaultMaxRetries
)

// Re-export error codes
const (
	ErrorCodeInvalidSession       = types.ErrorCodeInvalidSession
	ErrorCodeSessionExpired       = types.ErrorCodeSessionExpired
	ErrorCodeInvalidField         = types.ErrorCodeInvalidField
	ErrorCodeMalformedQuery       = types.ErrorCodeMalformedQuery
	ErrorCodeInvalidType          = types.ErrorCodeInvalidType
	ErrorCodeEntityDeleted        = types.ErrorCodeEntityDeleted
	ErrorCodeDuplicateValue       = types.ErrorCodeDuplicateValue
	ErrorCodeRequiredFieldMissing = types.ErrorCodeRequiredFieldMissing
	ErrorCodeInvalidCrossRef      = types.ErrorCodeInvalidCrossRef
	ErrorCodeInsufficientAccess   = types.ErrorCodeInsufficientAccess
	ErrorCodeRequestLimit         = types.ErrorCodeRequestLimit
	ErrorCodeStorageLimit         = types.ErrorCodeStorageLimit
)

// Re-export helper functions
var (
	ParseAPIError    = types.ParseAPIError
	IsNotFoundError  = types.IsNotFoundError
	IsAuthError      = types.IsAuthError
	IsRateLimitError = types.IsRateLimitError
)

// Re-export service types
type (
	SObject               = services.SObject
	SObjectMetadata       = services.SObjectMetadata
	GlobalDescribeResult  = services.GlobalDescribeResult
	QueryResult           = services.QueryResult
	QueryBuilder          = services.QueryBuilder
	JobInfo               = services.JobInfo
	FailedRecord          = services.FailedRecord
	CreateJobRequest      = services.CreateJobRequest
	ExecuteAnonymousResult = services.ExecuteAnonymousResult
)

// NewSObject creates a new SObject.
var NewSObject = services.NewSObject

// NewQueryBuilder creates a new query builder.
var NewQueryBuilder = services.NewQueryBuilder

// Authenticator defines the interface for authentication strategies.
type Authenticator interface {
	Authenticate(ctx context.Context) (*Token, error)
	Refresh(ctx context.Context) (*Token, error)
	IsTokenValid() bool
}

// SObjectService defines operations on Salesforce SObjects.
type SObjectService interface {
	Create(ctx context.Context, objectType string, data map[string]interface{}) (*SObject, error)
	Get(ctx context.Context, objectType, id string, fields ...string) (*SObject, error)
	Update(ctx context.Context, objectType, id string, data map[string]interface{}) error
	Upsert(ctx context.Context, objectType, externalIDField, externalID string, data map[string]interface{}) (*SObject, error)
	Delete(ctx context.Context, objectType, id string) error
	Describe(ctx context.Context, objectType string) (*SObjectMetadata, error)
	DescribeGlobal(ctx context.Context) (*GlobalDescribeResult, error)
}

// QueryService defines operations for SOQL queries.
type QueryService interface {
	Execute(ctx context.Context, query string) (*QueryResult, error)
	ExecuteAll(ctx context.Context, query string) (*QueryResult, error)
	QueryMore(ctx context.Context, nextRecordsURL string) (*QueryResult, error)
	NewBuilder(objectType string) *QueryBuilder
}

// BulkService defines operations for Bulk API 2.0.
type BulkService interface {
	CreateJob(ctx context.Context, req CreateJobRequest) (*JobInfo, error)
	UploadData(ctx context.Context, jobID string, data io.Reader) error
	CloseJob(ctx context.Context, jobID string) (*JobInfo, error)
	GetJobStatus(ctx context.Context, jobID string) (*JobInfo, error)
	GetSuccessfulRecords(ctx context.Context, jobID string) ([]map[string]interface{}, error)
	GetFailedRecords(ctx context.Context, jobID string) ([]FailedRecord, error)
	AbortJob(ctx context.Context, jobID string) (*JobInfo, error)
	DeleteJob(ctx context.Context, jobID string) error
}

// ToolingService defines operations for the Tooling API.
type ToolingService interface {
	Query(ctx context.Context, query string) (*QueryResult, error)
	ExecuteAnonymous(ctx context.Context, apexCode string) (*ExecuteAnonymousResult, error)
	Describe(ctx context.Context, objectType string) (*SObjectMetadata, error)
}

// ApexService defines operations for executing Apex REST endpoints.
type ApexService interface {
	Execute(ctx context.Context, method, path string, body interface{}) ([]byte, error)
}

// Limits represents Salesforce API limits.
type Limits struct {
	DailyApiRequests LimitInfo `json:"DailyApiRequests"`
}

// LimitInfo contains limit details.
type LimitInfo struct {
	Max       int `json:"Max"`
	Remaining int `json:"Remaining"`
}

// Used returns the number of requests used.
func (l LimitInfo) Used() int { return l.Max - l.Remaining }

// PercentUsed returns the percentage of limit used.
func (l LimitInfo) PercentUsed() float64 {
	if l.Max == 0 {
		return 0
	}
	return float64(l.Used()) / float64(l.Max) * 100
}

// CompositeRequest represents a composite API request.
type CompositeRequest struct {
	AllOrNone        bool                  `json:"allOrNone"`
	CompositeRequest []CompositeSubrequest `json:"compositeRequest"`
}

// CompositeSubrequest represents a single subrequest.
type CompositeSubrequest struct {
	Method      string                 `json:"method"`
	URL         string                 `json:"url"`
	ReferenceId string                 `json:"referenceId"`
	Body        map[string]interface{} `json:"body,omitempty"`
}

// CompositeResponse represents the response from a composite API request.
type CompositeResponse struct {
	CompositeResponse []CompositeSubresponse `json:"compositeResponse"`
}

// CompositeSubresponse represents a single subresponse.
type CompositeSubresponse struct {
	Body           interface{} `json:"body"`
	HTTPStatusCode int         `json:"httpStatusCode"`
	ReferenceId    string      `json:"referenceId"`
}
