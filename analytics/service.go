// Package analytics provides Reports and Dashboards API operations.
package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// Report represents a report definition.
type Report struct {
	ID                string           `json:"id"`
	Name              string           `json:"name"`
	DescribeURL       string           `json:"describeUrl"`
	InstancesURL      string           `json:"instancesUrl"`
	ReportMetadata    ReportMetadata   `json:"reportMetadata,omitempty"`
	ReportTypeMetadata interface{}     `json:"reportTypeMetadata,omitempty"`
	ReportExtendedMetadata interface{} `json:"reportExtendedMetadata,omitempty"`
}

// ReportMetadata contains report configuration.
type ReportMetadata struct {
	ID                    string        `json:"id"`
	Name                  string        `json:"name"`
	ReportType            ReportType    `json:"reportType"`
	ReportFormat          string        `json:"reportFormat"`
	Description           string        `json:"description"`
	FolderID              string        `json:"folderId"`
	DeveloperName         string        `json:"developerName"`
	DetailColumns         []string      `json:"detailColumns"`
	SortBy                []SortColumn  `json:"sortBy,omitempty"`
	GroupingsDown         []Grouping    `json:"groupingsDown,omitempty"`
	GroupingsAcross       []Grouping    `json:"groupingsAcross,omitempty"`
	ReportFilters         []ReportFilter `json:"reportFilters,omitempty"`
	ReportBooleanFilter   string        `json:"reportBooleanFilter,omitempty"`
	Aggregates            []string      `json:"aggregates,omitempty"`
	StandardDateFilter    DateFilter    `json:"standardDateFilter,omitempty"`
}

// ReportType contains report type information.
type ReportType struct {
	Type  string `json:"type"`
	Label string `json:"label"`
}

// SortColumn represents a sort column.
type SortColumn struct {
	SortColumn string `json:"sortColumn"`
	SortOrder  string `json:"sortOrder"`
}

// Grouping represents a report grouping.
type Grouping struct {
	Name              string `json:"name"`
	SortOrder         string `json:"sortOrder"`
	DateGranularity   string `json:"dateGranularity,omitempty"`
}

// ReportFilter represents a report filter.
type ReportFilter struct {
	Column     string      `json:"column"`
	Operator   string      `json:"operator"`
	Value      interface{} `json:"value"`
	FilterType string      `json:"filterType,omitempty"`
}

// DateFilter represents a date filter.
type DateFilter struct {
	Column     string `json:"column"`
	DurationValue string `json:"durationValue"`
	StartDate  string `json:"startDate,omitempty"`
	EndDate    string `json:"endDate,omitempty"`
}

// ReportInstance represents a report run instance.
type ReportInstance struct {
	ID                string `json:"id"`
	Status            string `json:"status"`
	URL               string `json:"url"`
	OwnerId           string `json:"ownerId"`
	CompletionDate    string `json:"completionDate,omitempty"`
	RequestDate       string `json:"requestDate"`
	HasDetailRows     bool   `json:"hasDetailRows"`
}

// ReportResult contains report execution results.
type ReportResult struct {
	Attributes        map[string]interface{} `json:"attributes"`
	AllData           bool                   `json:"allData"`
	FactMap           map[string]FactEntry   `json:"factMap"`
	GroupingsDown     GroupingResults        `json:"groupingsDown"`
	GroupingsAcross   GroupingResults        `json:"groupingsAcross"`
	HasDetailRows     bool                   `json:"hasDetailRows"`
	ReportMetadata    ReportMetadata         `json:"reportMetadata"`
}

// FactEntry represents a fact map entry.
type FactEntry struct {
	Aggregates []AggregateResult `json:"aggregates"`
	Rows       []DataRow         `json:"rows,omitempty"`
}

// AggregateResult represents an aggregate value.
type AggregateResult struct {
	Label string      `json:"label"`
	Value interface{} `json:"value"`
}

// DataRow represents a data row.
type DataRow struct {
	DataCells []DataCell `json:"dataCells"`
}

// DataCell represents a data cell.
type DataCell struct {
	Label string      `json:"label"`
	Value interface{} `json:"value"`
}

// GroupingResults contains grouping results.
type GroupingResults struct {
	Groupings []GroupingValue `json:"groupings"`
}

// GroupingValue represents a grouping value.
type GroupingValue struct {
	Key     string          `json:"key"`
	Label   string          `json:"label"`
	Value   interface{}     `json:"value"`
	Groupings []GroupingValue `json:"groupings,omitempty"`
}

// Dashboard represents a dashboard.
type Dashboard struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	FolderID      string `json:"folderId"`
	FolderName    string `json:"folderName"`
	DeveloperName string `json:"developerName"`
	RunningUser   User   `json:"runningUser,omitempty"`
	StatusURL     string `json:"statusUrl"`
	ComponentsURL string `json:"componentsUrl"`
}

// User represents a user reference.
type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// DashboardResult contains dashboard execution results.
type DashboardResult struct {
	StatusURL         string              `json:"statusUrl"`
	ComponentData     []ComponentResult   `json:"componentData"`
	ComponentMetadata []ComponentMetadata `json:"componentMetadata"`
}

// ComponentResult contains component data.
type ComponentResult struct {
	ComponentId string      `json:"componentId"`
	Status      string      `json:"status"`
	ReportResult *ReportResult `json:"reportResult,omitempty"`
}

// ComponentMetadata contains component metadata.
type ComponentMetadata struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	ReportID string `json:"reportId"`
}

// HTTPClient interface for dependency injection.
type HTTPClient interface {
	Get(ctx context.Context, path string) ([]byte, error)
	Post(ctx context.Context, path string, body interface{}) ([]byte, error)
	Patch(ctx context.Context, path string, body interface{}) ([]byte, error)
	Delete(ctx context.Context, path string) ([]byte, error)
}

// Service provides Analytics API operations.
type Service struct {
	client     HTTPClient
	apiVersion string
}

// NewService creates a new Analytics service.
func NewService(client HTTPClient, apiVersion string) *Service {
	return &Service{client: client, apiVersion: apiVersion}
}

// ListReports lists all reports.
func (s *Service) ListReports(ctx context.Context) ([]Report, error) {
	path := fmt.Sprintf("/services/data/v%s/analytics/reports", s.apiVersion)
	respBody, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var reports []Report
	if err := json.Unmarshal(respBody, &reports); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return reports, nil
}

// GetReport retrieves a report definition.
func (s *Service) GetReport(ctx context.Context, reportID string) (*Report, error) {
	path := fmt.Sprintf("/services/data/v%s/analytics/reports/%s/describe", s.apiVersion, reportID)
	respBody, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var report Report
	if err := json.Unmarshal(respBody, &report); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &report, nil
}

// RunReport runs a report synchronously.
func (s *Service) RunReport(ctx context.Context, reportID string, includeDetails bool) (*ReportResult, error) {
	path := fmt.Sprintf("/services/data/v%s/analytics/reports/%s?includeDetails=%t",
		s.apiVersion, reportID, includeDetails)
	respBody, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var result ReportResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &result, nil
}

// RunReportWithFilters runs a report with filters.
func (s *Service) RunReportWithFilters(ctx context.Context, reportID string, metadata ReportMetadata, includeDetails bool) (*ReportResult, error) {
	path := fmt.Sprintf("/services/data/v%s/analytics/reports/%s?includeDetails=%t",
		s.apiVersion, reportID, includeDetails)
	body := map[string]interface{}{"reportMetadata": metadata}
	respBody, err := s.client.Post(ctx, path, body)
	if err != nil {
		return nil, err
	}
	var result ReportResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &result, nil
}

// RunReportAsync runs a report asynchronously.
func (s *Service) RunReportAsync(ctx context.Context, reportID string) (*ReportInstance, error) {
	path := fmt.Sprintf("/services/data/v%s/analytics/reports/%s/instances", s.apiVersion, reportID)
	respBody, err := s.client.Post(ctx, path, nil)
	if err != nil {
		return nil, err
	}
	var instance ReportInstance
	if err := json.Unmarshal(respBody, &instance); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &instance, nil
}

// GetReportInstance retrieves an async report instance.
func (s *Service) GetReportInstance(ctx context.Context, reportID, instanceID string) (*ReportResult, error) {
	path := fmt.Sprintf("/services/data/v%s/analytics/reports/%s/instances/%s",
		s.apiVersion, reportID, instanceID)
	respBody, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var result ReportResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &result, nil
}

// ListReportInstances lists async report instances.
func (s *Service) ListReportInstances(ctx context.Context, reportID string) ([]ReportInstance, error) {
	path := fmt.Sprintf("/services/data/v%s/analytics/reports/%s/instances", s.apiVersion, reportID)
	respBody, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var instances []ReportInstance
	if err := json.Unmarshal(respBody, &instances); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return instances, nil
}

// ListDashboards lists all dashboards.
func (s *Service) ListDashboards(ctx context.Context) ([]Dashboard, error) {
	path := fmt.Sprintf("/services/data/v%s/analytics/dashboards", s.apiVersion)
	respBody, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Dashboards []Dashboard `json:"dashboards"`
	}
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return resp.Dashboards, nil
}

// GetDashboard retrieves a dashboard.
func (s *Service) GetDashboard(ctx context.Context, dashboardID string) (*DashboardResult, error) {
	path := fmt.Sprintf("/services/data/v%s/analytics/dashboards/%s", s.apiVersion, dashboardID)
	respBody, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var result DashboardResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &result, nil
}

// RefreshDashboard refreshes a dashboard.
func (s *Service) RefreshDashboard(ctx context.Context, dashboardID string) (*DashboardResult, error) {
	path := fmt.Sprintf("/services/data/v%s/analytics/dashboards/%s", s.apiVersion, dashboardID)
	respBody, err := s.client.Put(ctx, path, nil)
	if err != nil {
		return nil, err
	}
	var result DashboardResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &result, nil
}

// SearchReports searches reports by name.
func (s *Service) SearchReports(ctx context.Context, searchText string) ([]Report, error) {
	path := fmt.Sprintf("/services/data/v%s/analytics/reports?q=%s",
		s.apiVersion, url.QueryEscape(searchText))
	respBody, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var reports []Report
	if err := json.Unmarshal(respBody, &reports); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return reports, nil
}
