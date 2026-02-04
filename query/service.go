// Package query provides SOQL query execution and building.
package query

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// SObject represents a query result record.
type SObject struct {
	data map[string]interface{}
}

// FromMap creates an SObject from a map.
func FromMap(data map[string]interface{}) *SObject {
	return &SObject{data: data}
}

// Get returns a field value.
func (s *SObject) Get(key string) interface{} {
	if s.data == nil {
		return nil
	}
	return s.data[key]
}

// StringField returns a field as string.
func (s *SObject) StringField(key string) string {
	if v, ok := s.Get(key).(string); ok {
		return v
	}
	return ""
}

// ID returns the record ID.
func (s *SObject) ID() string { return s.StringField("Id") }

// ToMap returns the record as a map.
func (s *SObject) ToMap() map[string]interface{} { return s.data }

// Result contains SOQL query results.
type Result struct {
	TotalSize      int        `json:"totalSize"`
	Done           bool       `json:"done"`
	NextRecordsURL string     `json:"nextRecordsUrl,omitempty"`
	Records        []*SObject `json:"-"`
	RawRecords     []map[string]interface{} `json:"records"`
}

// HasMore returns true if more records are available.
func (r *Result) HasMore() bool {
	return !r.Done && r.NextRecordsURL != ""
}

// HTTPClient interface for dependency injection.
type HTTPClient interface {
	Get(ctx context.Context, path string) ([]byte, error)
	Post(ctx context.Context, path string, body interface{}) ([]byte, error)
}

// Service provides SOQL query operations.
type Service struct {
	client     HTTPClient
	apiVersion string
}

// NewService creates a new Query service.
func NewService(client HTTPClient, apiVersion string) *Service {
	return &Service{client: client, apiVersion: apiVersion}
}

// Execute runs a SOQL query.
func (s *Service) Execute(ctx context.Context, query string) (*Result, error) {
	path := fmt.Sprintf("/services/data/v%s/query?q=%s", s.apiVersion, url.QueryEscape(query))
	return s.executeQuery(ctx, path)
}

// ExecuteAll runs a SOQL query including deleted/archived records.
func (s *Service) ExecuteAll(ctx context.Context, query string) (*Result, error) {
	path := fmt.Sprintf("/services/data/v%s/queryAll?q=%s", s.apiVersion, url.QueryEscape(query))
	return s.executeQuery(ctx, path)
}

// QueryMore retrieves the next batch of query results.
func (s *Service) QueryMore(ctx context.Context, nextRecordsURL string) (*Result, error) {
	return s.executeQuery(ctx, nextRecordsURL)
}

// ExecuteWithCallback executes a query and calls fn for each record.
func (s *Service) ExecuteWithCallback(ctx context.Context, query string, fn func(*SObject) error) error {
	result, err := s.Execute(ctx, query)
	if err != nil {
		return err
	}
	for _, record := range result.Records {
		if err := fn(record); err != nil {
			return err
		}
	}
	for result.HasMore() {
		result, err = s.QueryMore(ctx, result.NextRecordsURL)
		if err != nil {
			return err
		}
		for _, record := range result.Records {
			if err := fn(record); err != nil {
				return err
			}
		}
	}
	return nil
}

// ExecuteAll fetches all records across pagination.
func (s *Service) ExecuteAllRecords(ctx context.Context, query string) ([]*SObject, error) {
	var allRecords []*SObject
	result, err := s.Execute(ctx, query)
	if err != nil {
		return nil, err
	}
	allRecords = append(allRecords, result.Records...)
	for result.HasMore() {
		result, err = s.QueryMore(ctx, result.NextRecordsURL)
		if err != nil {
			return nil, err
		}
		allRecords = append(allRecords, result.Records...)
	}
	return allRecords, nil
}

func (s *Service) executeQuery(ctx context.Context, path string) (*Result, error) {
	respBody, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var result Result
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	result.Records = make([]*SObject, len(result.RawRecords))
	for i, r := range result.RawRecords {
		result.Records[i] = FromMap(r)
	}
	return &result, nil
}

// NewBuilder creates a new SOQL query builder.
func (s *Service) NewBuilder(objectType string) *Builder {
	return NewBuilder(objectType)
}

// Builder provides fluent SOQL query building.
type Builder struct {
	objectType string
	fields     []string
	conditions []string
	orderBy    []string
	groupBy    []string
	having     []string
	limit      int
	offset     int
	forView    bool
	forRef     bool
	forUpdate  bool
}

// NewBuilder creates a new query builder.
func NewBuilder(objectType string) *Builder {
	return &Builder{objectType: objectType}
}

// Select adds fields to select.
func (b *Builder) Select(fields ...string) *Builder {
	b.fields = append(b.fields, fields...)
	return b
}

// Where adds a WHERE condition.
func (b *Builder) Where(condition string) *Builder {
	b.conditions = append(b.conditions, condition)
	return b
}

// WhereEquals adds an equality condition.
func (b *Builder) WhereEquals(field string, value interface{}) *Builder {
	b.conditions = append(b.conditions, fmt.Sprintf("%s = %s", field, formatValue(value)))
	return b
}

// WhereNotEquals adds a not-equal condition.
func (b *Builder) WhereNotEquals(field string, value interface{}) *Builder {
	b.conditions = append(b.conditions, fmt.Sprintf("%s != %s", field, formatValue(value)))
	return b
}

// WhereIn adds an IN condition.
func (b *Builder) WhereIn(field string, values ...interface{}) *Builder {
	formatted := make([]string, len(values))
	for i, v := range values {
		formatted[i] = formatValue(v)
	}
	b.conditions = append(b.conditions, fmt.Sprintf("%s IN (%s)", field, strings.Join(formatted, ", ")))
	return b
}

// WhereNotIn adds a NOT IN condition.
func (b *Builder) WhereNotIn(field string, values ...interface{}) *Builder {
	formatted := make([]string, len(values))
	for i, v := range values {
		formatted[i] = formatValue(v)
	}
	b.conditions = append(b.conditions, fmt.Sprintf("%s NOT IN (%s)", field, strings.Join(formatted, ", ")))
	return b
}

// WhereLike adds a LIKE condition.
func (b *Builder) WhereLike(field, pattern string) *Builder {
	b.conditions = append(b.conditions, fmt.Sprintf("%s LIKE '%s'", field, escapeSoql(pattern)))
	return b
}

// WhereNull adds an IS NULL condition.
func (b *Builder) WhereNull(field string) *Builder {
	b.conditions = append(b.conditions, field+" = NULL")
	return b
}

// WhereNotNull adds an IS NOT NULL condition.
func (b *Builder) WhereNotNull(field string) *Builder {
	b.conditions = append(b.conditions, field+" != NULL")
	return b
}

// WhereGreaterThan adds a > condition.
func (b *Builder) WhereGreaterThan(field string, value interface{}) *Builder {
	b.conditions = append(b.conditions, fmt.Sprintf("%s > %s", field, formatValue(value)))
	return b
}

// WhereLessThan adds a < condition.
func (b *Builder) WhereLessThan(field string, value interface{}) *Builder {
	b.conditions = append(b.conditions, fmt.Sprintf("%s < %s", field, formatValue(value)))
	return b
}

// OrderByAsc adds ascending ORDER BY.
func (b *Builder) OrderByAsc(field string) *Builder {
	b.orderBy = append(b.orderBy, field+" ASC")
	return b
}

// OrderByDesc adds descending ORDER BY.
func (b *Builder) OrderByDesc(field string) *Builder {
	b.orderBy = append(b.orderBy, field+" DESC")
	return b
}

// OrderByNullsFirst adds NULLS FIRST ordering.
func (b *Builder) OrderByNullsFirst(field, direction string) *Builder {
	b.orderBy = append(b.orderBy, fmt.Sprintf("%s %s NULLS FIRST", field, direction))
	return b
}

// OrderByNullsLast adds NULLS LAST ordering.
func (b *Builder) OrderByNullsLast(field, direction string) *Builder {
	b.orderBy = append(b.orderBy, fmt.Sprintf("%s %s NULLS LAST", field, direction))
	return b
}

// GroupBy adds GROUP BY fields.
func (b *Builder) GroupBy(fields ...string) *Builder {
	b.groupBy = append(b.groupBy, fields...)
	return b
}

// Having adds HAVING conditions.
func (b *Builder) Having(condition string) *Builder {
	b.having = append(b.having, condition)
	return b
}

// Limit sets the result limit.
func (b *Builder) Limit(limit int) *Builder {
	b.limit = limit
	return b
}

// Offset sets the result offset.
func (b *Builder) Offset(offset int) *Builder {
	b.offset = offset
	return b
}

// ForView adds FOR VIEW.
func (b *Builder) ForView() *Builder {
	b.forView = true
	return b
}

// ForReference adds FOR REFERENCE.
func (b *Builder) ForReference() *Builder {
	b.forRef = true
	return b
}

// ForUpdate adds FOR UPDATE.
func (b *Builder) ForUpdate() *Builder {
	b.forUpdate = true
	return b
}

// Build generates the SOQL query string.
func (b *Builder) Build() string {
	var sb strings.Builder
	sb.WriteString("SELECT ")
	if len(b.fields) == 0 {
		sb.WriteString("Id")
	} else {
		sb.WriteString(strings.Join(b.fields, ", "))
	}
	sb.WriteString(" FROM ")
	sb.WriteString(b.objectType)
	if len(b.conditions) > 0 {
		sb.WriteString(" WHERE ")
		sb.WriteString(strings.Join(b.conditions, " AND "))
	}
	if len(b.groupBy) > 0 {
		sb.WriteString(" GROUP BY ")
		sb.WriteString(strings.Join(b.groupBy, ", "))
	}
	if len(b.having) > 0 {
		sb.WriteString(" HAVING ")
		sb.WriteString(strings.Join(b.having, " AND "))
	}
	if len(b.orderBy) > 0 {
		sb.WriteString(" ORDER BY ")
		sb.WriteString(strings.Join(b.orderBy, ", "))
	}
	if b.limit > 0 {
		sb.WriteString(fmt.Sprintf(" LIMIT %d", b.limit))
	}
	if b.offset > 0 {
		sb.WriteString(fmt.Sprintf(" OFFSET %d", b.offset))
	}
	if b.forView {
		sb.WriteString(" FOR VIEW")
	}
	if b.forRef {
		sb.WriteString(" FOR REFERENCE")
	}
	if b.forUpdate {
		sb.WriteString(" FOR UPDATE")
	}
	return sb.String()
}

func formatValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		return fmt.Sprintf("'%s'", escapeSoql(val))
	case bool:
		if val {
			return "TRUE"
		}
		return "FALSE"
	case nil:
		return "NULL"
	default:
		return fmt.Sprintf("%v", val)
	}
}

func escapeSoql(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "'", "\\'")
	return s
}
