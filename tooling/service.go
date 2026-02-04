// Package tooling provides Tooling API operations.
package tooling

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// ExecuteAnonymousResult contains execute anonymous results.
type ExecuteAnonymousResult struct {
	Line                int    `json:"line"`
	Column              int    `json:"column"`
	Compiled            bool   `json:"compiled"`
	Success             bool   `json:"success"`
	CompiledClass       string `json:"compiledClass,omitempty"`
	CompileProblem      string `json:"compileProblem,omitempty"`
	ExceptionMessage    string `json:"exceptionMessage,omitempty"`
	ExceptionStackTrace string `json:"exceptionStackTrace,omitempty"`
}

// TestResult contains unit test results.
type TestResult struct {
	ApexTestResults []ApexTestResult `json:"apexTestResults,omitempty"`
	ApexTestClassId string           `json:"apexTestClassId"`
	AsyncApexJobId  string           `json:"asyncApexJobId"`
	Status          string           `json:"status"`
	NumberRun       int              `json:"numberRun"`
	NumberFailed    int              `json:"numberFailed"`
	TotalTime       float64          `json:"totalTime"`
}

// ApexTestResult contains individual test results.
type ApexTestResult struct {
	ID            string  `json:"id"`
	ApexClassId   string  `json:"apexClassId"`
	ApexClassName string  `json:"apexClassName"`
	MethodName    string  `json:"methodName"`
	Outcome       string  `json:"outcome"`
	Message       string  `json:"message,omitempty"`
	StackTrace    string  `json:"stackTrace,omitempty"`
	RunTime       float64 `json:"runTime"`
}

// TestQueueItem represents an item in the test queue.
type TestQueueItem struct {
	Id              string `json:"Id"`
	ApexClassId     string `json:"ApexClassId"`
	Status          string `json:"Status"`
	ExtendedStatus  string `json:"ExtendedStatus,omitempty"`
	ParentJobId     string `json:"ParentJobId,omitempty"`
	TestRunResultId string `json:"TestRunResultId,omitempty"`
}

// ApexLog represents an Apex debug log.
type ApexLog struct {
	Id             string `json:"Id"`
	Application    string `json:"Application"`
	DurationMillis int    `json:"DurationMilliseconds"`
	Location       string `json:"Location"`
	LogLength      int    `json:"LogLength"`
	LogUserId      string `json:"LogUserId"`
	Operation      string `json:"Operation"`
	Request        string `json:"Request"`
	StartTime      string `json:"StartTime"`
	Status         string `json:"Status"`
}

// Completions contains code completion results.
type Completions struct {
	Completions []Completion `json:"completions"`
}

// Completion represents a code completion suggestion.
type Completion struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Signature string `json:"signature,omitempty"`
}

// SObjectMetadata contains describe metadata for tooling objects.
type SObjectMetadata struct {
	Name       string          `json:"name"`
	Label      string          `json:"label"`
	Createable bool            `json:"createable"`
	Updateable bool            `json:"updateable"`
	Queryable  bool            `json:"queryable"`
	Fields     []FieldMetadata `json:"fields,omitempty"`
}

// FieldMetadata describes a tooling field.
type FieldMetadata struct {
	Name       string `json:"name"`
	Label      string `json:"label"`
	Type       string `json:"type"`
	Createable bool   `json:"createable"`
	Updateable bool   `json:"updateable"`
}

// QueryResult contains tooling query results.
type QueryResult struct {
	Size           int                      `json:"size"`
	TotalSize      int                      `json:"totalSize"`
	Done           bool                     `json:"done"`
	NextRecordsUrl string                   `json:"nextRecordsUrl,omitempty"`
	Records        []map[string]interface{} `json:"records"`
}

// ApexClass represents an Apex class.
type ApexClass struct {
	Id                    string `json:"Id"`
	Name                  string `json:"Name"`
	Body                  string `json:"Body"`
	ApiVersion            string `json:"ApiVersion"`
	Status                string `json:"Status"`
	IsValid               bool   `json:"IsValid"`
	LengthWithoutComments int    `json:"LengthWithoutComments"`
	NamespacePrefix       string `json:"NamespacePrefix,omitempty"`
}

// ApexTrigger represents an Apex trigger.
type ApexTrigger struct {
	Id            string `json:"Id"`
	Name          string `json:"Name"`
	Body          string `json:"Body"`
	ApiVersion    string `json:"ApiVersion"`
	Status        string `json:"Status"`
	IsValid       bool   `json:"IsValid"`
	TableEnumOrId string `json:"TableEnumOrId"`
}

// HTTPClient interface for dependency injection.
type HTTPClient interface {
	Get(ctx context.Context, path string) ([]byte, error)
	Post(ctx context.Context, path string, body interface{}) ([]byte, error)
	Patch(ctx context.Context, path string, body interface{}) ([]byte, error)
	Delete(ctx context.Context, path string) ([]byte, error)
}

// Service provides Tooling API operations.
type Service struct {
	client     HTTPClient
	apiVersion string
}

// NewService creates a new Tooling service.
func NewService(client HTTPClient, apiVersion string) *Service {
	return &Service{client: client, apiVersion: apiVersion}
}

// Query executes a Tooling API query.
func (s *Service) Query(ctx context.Context, query string) (*QueryResult, error) {
	path := fmt.Sprintf("/services/data/v%s/tooling/query?q=%s", s.apiVersion, url.QueryEscape(query))
	respBody, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var result QueryResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &result, nil
}

// QueryMore retrieves additional query results.
func (s *Service) QueryMore(ctx context.Context, nextRecordsURL string) (*QueryResult, error) {
	respBody, err := s.client.Get(ctx, nextRecordsURL)
	if err != nil {
		return nil, err
	}
	var result QueryResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &result, nil
}

// ExecuteAnonymous executes anonymous Apex code.
func (s *Service) ExecuteAnonymous(ctx context.Context, apexCode string) (*ExecuteAnonymousResult, error) {
	path := fmt.Sprintf("/services/data/v%s/tooling/executeAnonymous?anonymousBody=%s",
		s.apiVersion, url.QueryEscape(apexCode))
	respBody, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var result ExecuteAnonymousResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &result, nil
}

// RunTestsAsynchronous runs Apex tests asynchronously.
func (s *Service) RunTestsAsynchronous(ctx context.Context, classIds []string) (string, error) {
	path := fmt.Sprintf("/services/data/v%s/tooling/runTestsAsynchronous", s.apiVersion)
	body := map[string]interface{}{"classids": classIds}
	respBody, err := s.client.Post(ctx, path, body)
	if err != nil {
		return "", err
	}
	return string(respBody), nil
}

// RunTestsSynchronous runs Apex tests synchronously.
func (s *Service) RunTestsSynchronous(ctx context.Context, classNames []string) (*TestResult, error) {
	path := fmt.Sprintf("/services/data/v%s/tooling/runTestsSynchronous", s.apiVersion)
	body := map[string]interface{}{"tests": classNames}
	respBody, err := s.client.Post(ctx, path, body)
	if err != nil {
		return nil, err
	}
	var result TestResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &result, nil
}

// GetCompletions retrieves code completions.
func (s *Service) GetCompletions(ctx context.Context, completionType string) (*Completions, error) {
	path := fmt.Sprintf("/services/data/v%s/tooling/completions?type=%s", s.apiVersion, completionType)
	respBody, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var result Completions
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &result, nil
}

// Describe retrieves metadata for a Tooling API object.
func (s *Service) Describe(ctx context.Context, objectType string) (*SObjectMetadata, error) {
	path := fmt.Sprintf("/services/data/v%s/tooling/sobjects/%s/describe", s.apiVersion, objectType)
	respBody, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var meta SObjectMetadata
	if err := json.Unmarshal(respBody, &meta); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &meta, nil
}

// DescribeGlobal lists all Tooling API objects.
func (s *Service) DescribeGlobal(ctx context.Context) ([]SObjectMetadata, error) {
	path := fmt.Sprintf("/services/data/v%s/tooling/sobjects", s.apiVersion)
	respBody, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var result struct {
		SObjects []SObjectMetadata `json:"sobjects"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return result.SObjects, nil
}

// CreateApexClass creates a new Apex class.
func (s *Service) CreateApexClass(ctx context.Context, name, body string, apiVersion float64) (*ApexClass, error) {
	path := fmt.Sprintf("/services/data/v%s/tooling/sobjects/ApexClass", s.apiVersion)
	data := map[string]interface{}{
		"Name":       name,
		"Body":       body,
		"ApiVersion": apiVersion,
	}
	respBody, err := s.client.Post(ctx, path, data)
	if err != nil {
		return nil, err
	}
	var result struct {
		Id      string `json:"id"`
		Success bool   `json:"success"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return s.GetApexClass(ctx, result.Id)
}

// GetApexClass retrieves an Apex class by ID.
func (s *Service) GetApexClass(ctx context.Context, id string) (*ApexClass, error) {
	path := fmt.Sprintf("/services/data/v%s/tooling/sobjects/ApexClass/%s", s.apiVersion, id)
	respBody, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var cls ApexClass
	if err := json.Unmarshal(respBody, &cls); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &cls, nil
}

// UpdateApexClass updates an Apex class body.
func (s *Service) UpdateApexClass(ctx context.Context, id, body string) error {
	path := fmt.Sprintf("/services/data/v%s/tooling/sobjects/ApexClass/%s", s.apiVersion, id)
	data := map[string]interface{}{"Body": body}
	_, err := s.client.Patch(ctx, path, data)
	return err
}

// DeleteApexClass deletes an Apex class.
func (s *Service) DeleteApexClass(ctx context.Context, id string) error {
	path := fmt.Sprintf("/services/data/v%s/tooling/sobjects/ApexClass/%s", s.apiVersion, id)
	_, err := s.client.Delete(ctx, path)
	return err
}

// GetApexLogs retrieves debug logs.
func (s *Service) GetApexLogs(ctx context.Context, limit int) ([]ApexLog, error) {
	query := fmt.Sprintf("SELECT Id,Application,DurationMilliseconds,Location,LogLength,LogUserId,Operation,Request,StartTime,Status FROM ApexLog ORDER BY StartTime DESC LIMIT %d", limit)
	result, err := s.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	logs := make([]ApexLog, len(result.Records))
	for i, r := range result.Records {
		data, err := json.Marshal(r)
		if err != nil {
			continue
		}
		if err := json.Unmarshal(data, &logs[i]); err != nil {
			continue
		}
	}
	return logs, nil
}

// GetApexLogBody retrieves the body of a debug log.
func (s *Service) GetApexLogBody(ctx context.Context, logId string) (string, error) {
	path := fmt.Sprintf("/services/data/v%s/tooling/sobjects/ApexLog/%s/Body", s.apiVersion, logId)
	respBody, err := s.client.Get(ctx, path)
	if err != nil {
		return "", err
	}
	return string(respBody), nil
}
