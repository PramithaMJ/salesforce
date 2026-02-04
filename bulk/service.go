// Package bulk provides Bulk API 2.0 operations.
package bulk

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// Operation represents bulk job operation types.
type Operation string

const (
	OperationInsert   Operation = "insert"
	OperationUpdate   Operation = "update"
	OperationUpsert   Operation = "upsert"
	OperationDelete   Operation = "delete"
	OperationHardDelete Operation = "hardDelete"
)

// State represents bulk job states.
type State string

const (
	StateOpen           State = "Open"
	StateUploadComplete State = "UploadComplete"
	StateInProgress     State = "InProgress"
	StateJobComplete    State = "JobComplete"
	StateFailed         State = "Failed"
	StateAborted        State = "Aborted"
)

// ContentType represents data content types.
type ContentType string

const (
	ContentTypeCSV  ContentType = "CSV"
	ContentTypeJSON ContentType = "JSON"
)

// LineEnding represents line ending types.
type LineEnding string

const (
	LineEndingLF   LineEnding = "LF"
	LineEndingCRLF LineEnding = "CRLF"
)

// ColumnDelimiter represents CSV column delimiters.
type ColumnDelimiter string

const (
	DelimiterComma     ColumnDelimiter = "COMMA"
	DelimiterTab       ColumnDelimiter = "TAB"
	DelimiterSemicolon ColumnDelimiter = "SEMICOLON"
	DelimiterPipe      ColumnDelimiter = "PIPE"
	DelimiterBackquote ColumnDelimiter = "BACKQUOTE"
	DelimiterCaret     ColumnDelimiter = "CARET"
)

// CreateJobRequest contains job creation parameters.
type CreateJobRequest struct {
	Object              string          `json:"object"`
	Operation           Operation       `json:"operation"`
	ExternalIdFieldName string          `json:"externalIdFieldName,omitempty"`
	ContentType         ContentType     `json:"contentType,omitempty"`
	LineEnding          LineEnding      `json:"lineEnding,omitempty"`
	ColumnDelimiter     ColumnDelimiter `json:"columnDelimiter,omitempty"`
}

// JobInfo contains bulk job information.
type JobInfo struct {
	ID                      string      `json:"id"`
	Object                  string      `json:"object"`
	Operation               Operation   `json:"operation"`
	State                   State       `json:"state"`
	ContentType             ContentType `json:"contentType"`
	ColumnDelimiter         string      `json:"columnDelimiter"`
	LineEnding              LineEnding  `json:"lineEnding"`
	ExternalIdFieldName     string      `json:"externalIdFieldName,omitempty"`
	CreatedById             string      `json:"createdById"`
	CreatedDate             string      `json:"createdDate"`
	SystemModstamp          string      `json:"systemModstamp"`
	ConcurrencyMode         string      `json:"concurrencyMode"`
	ContentURL              string      `json:"contentUrl,omitempty"`
	NumberRecordsProcessed  int         `json:"numberRecordsProcessed"`
	NumberRecordsFailed     int         `json:"numberRecordsFailed"`
	Retries                 int         `json:"retries"`
	TotalProcessingTime     int         `json:"totalProcessingTime"`
	ApiActiveProcessingTime int         `json:"apiActiveProcessingTime"`
	ApexProcessingTime      int         `json:"apexProcessingTime"`
	ErrorMessage            string      `json:"errorMessage,omitempty"`
}

// IsComplete returns true if the job has finished.
func (j *JobInfo) IsComplete() bool {
	return j.State == StateJobComplete || j.State == StateFailed || j.State == StateAborted
}

// IsSuccess returns true if the job completed successfully.
func (j *JobInfo) IsSuccess() bool {
	return j.State == StateJobComplete && j.NumberRecordsFailed == 0
}

// FailedRecord represents a failed record.
type FailedRecord struct {
	ID    string                 `json:"sf__Id"`
	Error string                 `json:"sf__Error"`
	Data  map[string]interface{} `json:"-"`
}

// SuccessRecord represents a successful record.
type SuccessRecord struct {
	ID      string                 `json:"sf__Id"`
	Created bool                   `json:"sf__Created"`
	Data    map[string]interface{} `json:"-"`
}

// QueryJobRequest contains query job creation parameters.
type QueryJobRequest struct {
	Query       string      `json:"query"`
	Operation   Operation   `json:"operation,omitempty"`
	ContentType ContentType `json:"contentType,omitempty"`
}

// QueryJobInfo contains query job information.
type QueryJobInfo struct {
	ID                     string      `json:"id"`
	Operation              Operation   `json:"operation"`
	Object                 string      `json:"object"`
	State                  State       `json:"state"`
	ContentType            ContentType `json:"contentType"`
	CreatedById            string      `json:"createdById"`
	CreatedDate            string      `json:"createdDate"`
	SystemModstamp         string      `json:"systemModstamp"`
	NumberRecordsProcessed int         `json:"numberRecordsProcessed"`
}

// JobListResult contains a list of jobs.
type JobListResult struct {
	Done           bool      `json:"done"`
	Records        []JobInfo `json:"records"`
	NextRecordsURL string    `json:"nextRecordsUrl,omitempty"`
}

// HTTPClient interface for dependency injection.
type HTTPClient interface {
	Get(ctx context.Context, path string) ([]byte, error)
	Post(ctx context.Context, path string, body interface{}) ([]byte, error)
	Patch(ctx context.Context, path string, body interface{}) ([]byte, error)
	Put(ctx context.Context, path string, body interface{}) ([]byte, error)
	Delete(ctx context.Context, path string) ([]byte, error)
}

// Service provides Bulk API 2.0 operations.
type Service struct {
	client     HTTPClient
	apiVersion string
}

// NewService creates a new Bulk service.
func NewService(client HTTPClient, apiVersion string) *Service {
	return &Service{client: client, apiVersion: apiVersion}
}

// CreateJob creates a new ingest job.
func (s *Service) CreateJob(ctx context.Context, req CreateJobRequest) (*JobInfo, error) {
	if req.ContentType == "" {
		req.ContentType = ContentTypeCSV
	}
	if req.LineEnding == "" {
		req.LineEnding = LineEndingLF
	}
	path := fmt.Sprintf("/services/data/v%s/jobs/ingest", s.apiVersion)
	respBody, err := s.client.Post(ctx, path, req)
	if err != nil {
		return nil, err
	}
	var job JobInfo
	if err := json.Unmarshal(respBody, &job); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &job, nil
}

// UploadData uploads data to an ingest job.
func (s *Service) UploadData(ctx context.Context, jobID string, data io.Reader) error {
	path := fmt.Sprintf("/services/data/v%s/jobs/ingest/%s/batches", s.apiVersion, jobID)
	_, err := s.client.Put(ctx, path, data)
	return err
}

// UploadCSV uploads CSV data to an ingest job.
func (s *Service) UploadCSV(ctx context.Context, jobID string, records []map[string]interface{}, columns []string) error {
	if len(records) == 0 {
		return nil
	}
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	if len(columns) == 0 {
		for key := range records[0] {
			columns = append(columns, key)
		}
	}
	if err := writer.Write(columns); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}
	for _, record := range records {
		row := make([]string, len(columns))
		for i, col := range columns {
			if val, ok := record[col]; ok {
				row[i] = fmt.Sprintf("%v", val)
			}
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}
	writer.Flush()
	return s.UploadData(ctx, jobID, &buf)
}

// CloseJob closes an ingest job to begin processing.
func (s *Service) CloseJob(ctx context.Context, jobID string) (*JobInfo, error) {
	path := fmt.Sprintf("/services/data/v%s/jobs/ingest/%s", s.apiVersion, jobID)
	respBody, err := s.client.Patch(ctx, path, map[string]string{"state": string(StateUploadComplete)})
	if err != nil {
		return nil, err
	}
	var job JobInfo
	if err := json.Unmarshal(respBody, &job); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &job, nil
}

// GetJob retrieves job information.
func (s *Service) GetJob(ctx context.Context, jobID string) (*JobInfo, error) {
	path := fmt.Sprintf("/services/data/v%s/jobs/ingest/%s", s.apiVersion, jobID)
	respBody, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var job JobInfo
	if err := json.Unmarshal(respBody, &job); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &job, nil
}

// ListJobs lists all ingest jobs.
func (s *Service) ListJobs(ctx context.Context, concurrencyMode string, isPkChunkingEnabled bool) (*JobListResult, error) {
	path := fmt.Sprintf("/services/data/v%s/jobs/ingest", s.apiVersion)
	respBody, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var result JobListResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &result, nil
}

// AbortJob aborts an ingest job.
func (s *Service) AbortJob(ctx context.Context, jobID string) (*JobInfo, error) {
	path := fmt.Sprintf("/services/data/v%s/jobs/ingest/%s", s.apiVersion, jobID)
	respBody, err := s.client.Patch(ctx, path, map[string]string{"state": string(StateAborted)})
	if err != nil {
		return nil, err
	}
	var job JobInfo
	if err := json.Unmarshal(respBody, &job); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &job, nil
}

// DeleteJob deletes an ingest job.
func (s *Service) DeleteJob(ctx context.Context, jobID string) error {
	path := fmt.Sprintf("/services/data/v%s/jobs/ingest/%s", s.apiVersion, jobID)
	_, err := s.client.Delete(ctx, path)
	return err
}

// WaitForCompletion waits for a job to complete.
func (s *Service) WaitForCompletion(ctx context.Context, jobID string, pollInterval time.Duration) (*JobInfo, error) {
	if pollInterval == 0 {
		pollInterval = 5 * time.Second
	}
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			job, err := s.GetJob(ctx, jobID)
			if err != nil {
				return nil, err
			}
			if job.IsComplete() {
				return job, nil
			}
		}
	}
}

// GetSuccessfulRecords retrieves successfully processed records.
func (s *Service) GetSuccessfulRecords(ctx context.Context, jobID string) ([]SuccessRecord, error) {
	path := fmt.Sprintf("/services/data/v%s/jobs/ingest/%s/successfulResults", s.apiVersion, jobID)
	respBody, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	records, err := parseCSV(respBody)
	if err != nil {
		return nil, err
	}
	result := make([]SuccessRecord, len(records))
	for i, r := range records {
		result[i] = SuccessRecord{
			ID:      getString(r, "sf__Id"),
			Created: getString(r, "sf__Created") == "true",
			Data:    r,
		}
	}
	return result, nil
}

// GetFailedRecords retrieves failed records.
func (s *Service) GetFailedRecords(ctx context.Context, jobID string) ([]FailedRecord, error) {
	path := fmt.Sprintf("/services/data/v%s/jobs/ingest/%s/failedResults", s.apiVersion, jobID)
	respBody, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	records, err := parseCSV(respBody)
	if err != nil {
		return nil, err
	}
	result := make([]FailedRecord, len(records))
	for i, r := range records {
		result[i] = FailedRecord{
			ID:    getString(r, "sf__Id"),
			Error: getString(r, "sf__Error"),
			Data:  r,
		}
	}
	return result, nil
}

// GetUnprocessedRecords retrieves unprocessed records.
func (s *Service) GetUnprocessedRecords(ctx context.Context, jobID string) ([]map[string]interface{}, error) {
	path := fmt.Sprintf("/services/data/v%s/jobs/ingest/%s/unprocessedrecords", s.apiVersion, jobID)
	respBody, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	return parseCSV(respBody)
}

// CreateQueryJob creates a bulk query job.
func (s *Service) CreateQueryJob(ctx context.Context, req QueryJobRequest) (*QueryJobInfo, error) {
	if req.Operation == "" {
		req.Operation = "query"
	}
	if req.ContentType == "" {
		req.ContentType = ContentTypeCSV
	}
	path := fmt.Sprintf("/services/data/v%s/jobs/query", s.apiVersion)
	respBody, err := s.client.Post(ctx, path, req)
	if err != nil {
		return nil, err
	}
	var job QueryJobInfo
	if err := json.Unmarshal(respBody, &job); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &job, nil
}

// GetQueryJob retrieves query job information.
func (s *Service) GetQueryJob(ctx context.Context, jobID string) (*QueryJobInfo, error) {
	path := fmt.Sprintf("/services/data/v%s/jobs/query/%s", s.apiVersion, jobID)
	respBody, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var job QueryJobInfo
	if err := json.Unmarshal(respBody, &job); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &job, nil
}

// GetQueryResults retrieves query job results.
func (s *Service) GetQueryResults(ctx context.Context, jobID string, maxRecords int, locator string) ([]map[string]interface{}, string, error) {
	path := fmt.Sprintf("/services/data/v%s/jobs/query/%s/results", s.apiVersion, jobID)
	if maxRecords > 0 || locator != "" {
		path += "?"
		if maxRecords > 0 {
			path += fmt.Sprintf("maxRecords=%d", maxRecords)
		}
		if locator != "" {
			if maxRecords > 0 {
				path += "&"
			}
			path += fmt.Sprintf("locator=%s", locator)
		}
	}
	respBody, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, "", err
	}
	records, err := parseCSV(respBody)
	return records, "", err
}

// AbortQueryJob aborts a query job.
func (s *Service) AbortQueryJob(ctx context.Context, jobID string) (*QueryJobInfo, error) {
	path := fmt.Sprintf("/services/data/v%s/jobs/query/%s", s.apiVersion, jobID)
	respBody, err := s.client.Patch(ctx, path, map[string]string{"state": string(StateAborted)})
	if err != nil {
		return nil, err
	}
	var job QueryJobInfo
	if err := json.Unmarshal(respBody, &job); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &job, nil
}

// DeleteQueryJob deletes a query job.
func (s *Service) DeleteQueryJob(ctx context.Context, jobID string) error {
	path := fmt.Sprintf("/services/data/v%s/jobs/query/%s", s.apiVersion, jobID)
	_, err := s.client.Delete(ctx, path)
	return err
}

func parseCSV(data []byte) ([]map[string]interface{}, error) {
	reader := csv.NewReader(bytes.NewReader(data))
	headers, err := reader.Read()
	if err == io.EOF {
		return []map[string]interface{}{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}
	var records []map[string]interface{}
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read row: %w", err)
		}
		record := make(map[string]interface{})
		for i, h := range headers {
			if i < len(row) {
				record[h] = row[i]
			}
		}
		records = append(records, record)
	}
	return records, nil
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}
