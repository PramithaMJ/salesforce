// Package limits provides API Limits operations.
package limits

import (
	"context"
	"encoding/json"
	"fmt"
)

// Limit represents an API limit.
type Limit struct {
	Max       int `json:"Max"`
	Remaining int `json:"Remaining"`
}

// Used returns the number of uses consumed.
func (l Limit) Used() int {
	return l.Max - l.Remaining
}

// PercentUsed returns the percentage of limit used.
func (l Limit) PercentUsed() float64 {
	if l.Max == 0 {
		return 0
	}
	return float64(l.Used()) / float64(l.Max) * 100
}

// Limits contains all API limits.
type Limits struct {
	ActiveScratchOrgs               Limit `json:"ActiveScratchOrgs"`
	AnalyticsExternalDataSizeMB     Limit `json:"AnalyticsExternalDataSizeMB"`
	ConcurrentAsyncGetReportInstances Limit `json:"ConcurrentAsyncGetReportInstances"`
	ConcurrentEinsteinDataInsightsStoryCreation Limit `json:"ConcurrentEinsteinDataInsightsStoryCreation"`
	ConcurrentEinsteinDiscoveryStoryCreation Limit `json:"ConcurrentEinsteinDiscoveryStoryCreation"`
	ConcurrentSyncReportRuns        Limit `json:"ConcurrentSyncReportRuns"`
	DailyAnalyticsDataflowJobExecutions Limit `json:"DailyAnalyticsDataflowJobExecutions"`
	DailyAnalyticsUploadedFilesSizeMB Limit `json:"DailyAnalyticsUploadedFilesSizeMB"`
	DailyApiRequests                Limit `json:"DailyApiRequests"`
	DailyAsyncApexExecutions        Limit `json:"DailyAsyncApexExecutions"`
	DailyBulkApiRequests            Limit `json:"DailyBulkApiRequests"`
	DailyBulkV2QueryFileStorageMB   Limit `json:"DailyBulkV2QueryFileStorageMB"`
	DailyBulkV2QueryJobs            Limit `json:"DailyBulkV2QueryJobs"`
	DailyDeliveredPlatformEvents    Limit `json:"DailyDeliveredPlatformEvents"`
	DailyDurableGenericStreamingApiEvents Limit `json:"DailyDurableGenericStreamingApiEvents"`
	DailyDurableStreamingApiEvents  Limit `json:"DailyDurableStreamingApiEvents"`
	DailyEinsteinDataInsightsStoryCreation Limit `json:"DailyEinsteinDataInsightsStoryCreation"`
	DailyEinsteinDiscoveryPredictAPIAggregate Limit `json:"DailyEinsteinDiscoveryPredictAPICalls"`
	DailyEinsteinDiscoveryRecreatePredictions Limit `json:"DailyEinsteinDiscoveryRecreatePredictions"`
	DailyEinsteinDiscoveryStoryCreation Limit `json:"DailyEinsteinDiscoveryStoryCreation"`
	DailyFunctionsApiCallLimit      Limit `json:"DailyFunctionsApiCallLimit"`
	DailyGenericStreamingApiEvents  Limit `json:"DailyGenericStreamingApiEvents"`
	DailyScratchOrgs                Limit `json:"DailyScratchOrgs"`
	DailyStandardVolumePlatformEvents Limit `json:"DailyStandardVolumePlatformEvents"`
	DailyStreamingApiEvents         Limit `json:"DailyStreamingApiEvents"`
	DailyWorkflowEmails             Limit `json:"DailyWorkflowEmails"`
	DataStorageMB                   Limit `json:"DataStorageMB"`
	DurableStreamingApiConcurrentClients Limit `json:"DurableStreamingApiConcurrentClients"`
	FileStorageMB                   Limit `json:"FileStorageMB"`
	HourlyAsyncReportRuns           Limit `json:"HourlyAsyncReportRuns"`
	HourlyDashboardRefreshes        Limit `json:"HourlyDashboardRefreshes"`
	HourlyDashboardResults          Limit `json:"HourlyDashboardResults"`
	HourlyDashboardStatuses         Limit `json:"HourlyDashboardStatuses"`
	HourlyLongTermIdMapping         Limit `json:"HourlyLongTermIdMapping"`
	HourlyManagedContentPublicRequests Limit `json:"HourlyManagedContentPublicRequests"`
	HourlyODataCallout              Limit `json:"HourlyODataCallout"`
	HourlyPublishedPlatformEvents   Limit `json:"HourlyPublishedPlatformEvents"`
	HourlyPublishedStandardVolumePlatformEvents Limit `json:"HourlyPublishedStandardVolumePlatformEvents"`
	HourlyShortTermIdMapping        Limit `json:"HourlyShortTermIdMapping"`
	HourlySyncReportRuns            Limit `json:"HourlySyncReportRuns"`
	HourlyTimeBasedWorkflow         Limit `json:"HourlyTimeBasedWorkflow"`
	MassEmail                       Limit `json:"MassEmail"`
	MonthlyEinsteinDiscoveryStoryCreation Limit `json:"MonthlyEinsteinDiscoveryStoryCreation"`
	MonthlyPlatformEventsUsageEntitlement Limit `json:"MonthlyPlatformEventsUsageEntitlement"`
	Package2VersionCreates          Limit `json:"Package2VersionCreates"`
	Package2VersionCreatesWithoutValidation Limit `json:"Package2VersionCreatesWithoutValidation"`
	PermissionSets                  Limit `json:"PermissionSets"`
	PrivateConnectOutboundCalloutHourlyLimitMB Limit `json:"PrivateConnectOutboundCalloutHourlyLimitMB"`
	SingleEmail                     Limit `json:"SingleEmail"`
	StreamingApiConcurrentClients   Limit `json:"StreamingApiConcurrentClients"`
}

// HTTPClient interface for dependency injection.
type HTTPClient interface {
	Get(ctx context.Context, path string) ([]byte, error)
}

// Service provides Limits API operations.
type Service struct {
	client     HTTPClient
	apiVersion string
}

// NewService creates a new Limits service.
func NewService(client HTTPClient, apiVersion string) *Service {
	return &Service{client: client, apiVersion: apiVersion}
}

// GetLimits retrieves all API limits.
func (s *Service) GetLimits(ctx context.Context) (*Limits, error) {
	path := fmt.Sprintf("/services/data/v%s/limits", s.apiVersion)
	respBody, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var limits Limits
	if err := json.Unmarshal(respBody, &limits); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &limits, nil
}

// GetDailyApiRequests returns the daily API request limit.
func (s *Service) GetDailyApiRequests(ctx context.Context) (*Limit, error) {
	limits, err := s.GetLimits(ctx)
	if err != nil {
		return nil, err
	}
	return &limits.DailyApiRequests, nil
}

// GetDataStorage returns the data storage limit.
func (s *Service) GetDataStorage(ctx context.Context) (*Limit, error) {
	limits, err := s.GetLimits(ctx)
	if err != nil {
		return nil, err
	}
	return &limits.DataStorageMB, nil
}

// GetFileStorage returns the file storage limit.
func (s *Service) GetFileStorage(ctx context.Context) (*Limit, error) {
	limits, err := s.GetLimits(ctx)
	if err != nil {
		return nil, err
	}
	return &limits.FileStorageMB, nil
}
