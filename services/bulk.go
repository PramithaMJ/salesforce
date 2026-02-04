package services

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/PramithaMJ/salesforce/http"
)

type JobOperation string
type JobState string
type ContentType string
type LineEnding string

const (
	JobOperationInsert JobOperation = "insert"
	JobOperationUpdate JobOperation = "update"
	JobOperationUpsert JobOperation = "upsert"
	JobOperationDelete JobOperation = "delete"
	JobOperationQuery  JobOperation = "query"
)

const (
	JobStateOpen           JobState = "Open"
	JobStateUploadComplete JobState = "UploadComplete"
	JobStateInProgress     JobState = "InProgress"
	JobStateJobComplete    JobState = "JobComplete"
	JobStateFailed         JobState = "Failed"
	JobStateAborted        JobState = "Aborted"
)

const (
	ContentTypeCSV ContentType = "CSV"
	LineEndingLF   LineEnding  = "LF"
)

type CreateJobRequest struct {
	Object              string       `json:"object"`
	Operation           JobOperation `json:"operation"`
	ExternalIdFieldName string       `json:"externalIdFieldName,omitempty"`
	ContentType         ContentType  `json:"contentType,omitempty"`
	LineEnding          LineEnding   `json:"lineEnding,omitempty"`
}

type JobInfo struct {
	ID                     string       `json:"id"`
	Object                 string       `json:"object"`
	Operation              JobOperation `json:"operation"`
	State                  JobState     `json:"state"`
	ContentType            ContentType  `json:"contentType"`
	NumberRecordsProcessed int          `json:"numberRecordsProcessed"`
	NumberRecordsFailed    int          `json:"numberRecordsFailed"`
}

func (j *JobInfo) IsComplete() bool {
	return j.State == JobStateJobComplete || j.State == JobStateFailed || j.State == JobStateAborted
}

type FailedRecord struct {
	ID    string                 `json:"sf__Id"`
	Error string                 `json:"sf__Error"`
	Data  map[string]interface{} `json:"-"`
}

type BulkServiceImpl struct {
	client     *http.Client
	apiVersion string
}

func NewBulkService(client *http.Client, apiVersion string) *BulkServiceImpl {
	return &BulkServiceImpl{client: client, apiVersion: apiVersion}
}

func (b *BulkServiceImpl) CreateJob(ctx context.Context, req CreateJobRequest) (*JobInfo, error) {
	if req.ContentType == "" {
		req.ContentType = ContentTypeCSV
	}
	if req.LineEnding == "" {
		req.LineEnding = LineEndingLF
	}
	path := fmt.Sprintf("/services/data/v%s/jobs/ingest", b.apiVersion)
	respBody, err := b.client.Post(ctx, path, req)
	if err != nil {
		return nil, err
	}
	var job JobInfo
	return &job, json.Unmarshal(respBody, &job)
}

func (b *BulkServiceImpl) UploadData(ctx context.Context, jobID string, data io.Reader) error {
	path := fmt.Sprintf("/services/data/v%s/jobs/ingest/%s/batches", b.apiVersion, jobID)
	_, err := b.client.Put(ctx, path, data)
	return err
}

func (b *BulkServiceImpl) CloseJob(ctx context.Context, jobID string) (*JobInfo, error) {
	path := fmt.Sprintf("/services/data/v%s/jobs/ingest/%s", b.apiVersion, jobID)
	respBody, err := b.client.Patch(ctx, path, map[string]string{"state": string(JobStateUploadComplete)})
	if err != nil {
		return nil, err
	}
	var job JobInfo
	return &job, json.Unmarshal(respBody, &job)
}

func (b *BulkServiceImpl) GetJobStatus(ctx context.Context, jobID string) (*JobInfo, error) {
	path := fmt.Sprintf("/services/data/v%s/jobs/ingest/%s", b.apiVersion, jobID)
	respBody, err := b.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var job JobInfo
	return &job, json.Unmarshal(respBody, &job)
}

func (b *BulkServiceImpl) WaitForCompletion(ctx context.Context, jobID string, poll time.Duration) (*JobInfo, error) {
	if poll == 0 {
		poll = 5 * time.Second
	}
	ticker := time.NewTicker(poll)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			job, err := b.GetJobStatus(ctx, jobID)
			if err != nil {
				return nil, err
			}
			if job.IsComplete() {
				return job, nil
			}
		}
	}
}

func (b *BulkServiceImpl) GetSuccessfulRecords(ctx context.Context, jobID string) ([]map[string]interface{}, error) {
	path := fmt.Sprintf("/services/data/v%s/jobs/ingest/%s/successfulResults", b.apiVersion, jobID)
	respBody, err := b.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	return parseCSV(respBody)
}

func (b *BulkServiceImpl) GetFailedRecords(ctx context.Context, jobID string) ([]FailedRecord, error) {
	path := fmt.Sprintf("/services/data/v%s/jobs/ingest/%s/failedResults", b.apiVersion, jobID)
	respBody, err := b.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	records, err := parseCSV(respBody)
	if err != nil {
		return nil, err
	}
	result := make([]FailedRecord, len(records))
	for i, r := range records {
		result[i] = FailedRecord{ID: getString(r, "sf__Id"), Error: getString(r, "sf__Error"), Data: r}
	}
	return result, nil
}

func (b *BulkServiceImpl) AbortJob(ctx context.Context, jobID string) (*JobInfo, error) {
	path := fmt.Sprintf("/services/data/v%s/jobs/ingest/%s", b.apiVersion, jobID)
	respBody, err := b.client.Patch(ctx, path, map[string]string{"state": string(JobStateAborted)})
	if err != nil {
		return nil, err
	}
	var job JobInfo
	return &job, json.Unmarshal(respBody, &job)
}

func (b *BulkServiceImpl) DeleteJob(ctx context.Context, jobID string) error {
	path := fmt.Sprintf("/services/data/v%s/jobs/ingest/%s", b.apiVersion, jobID)
	_, err := b.client.Delete(ctx, path)
	return err
}

func parseCSV(data []byte) ([]map[string]interface{}, error) {
	reader := csv.NewReader(bytes.NewReader(data))
	headers, err := reader.Read()
	if err == io.EOF {
		return []map[string]interface{}{}, nil
	}
	if err != nil {
		return nil, err
	}
	var records []map[string]interface{}
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		rec := make(map[string]interface{})
		for i, h := range headers {
			if i < len(row) {
				rec[h] = row[i]
			}
		}
		records = append(records, rec)
	}
	return records, nil
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}
