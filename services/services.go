package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/PramithaMJ/salesforce/http"
	"github.com/PramithaMJ/salesforce/types"
)

type SObject struct {
	data map[string]interface{}
}

type SObjectAttributes struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

func NewSObject(objectType string) *SObject {
	return &SObject{data: map[string]interface{}{"attributes": SObjectAttributes{Type: objectType}}}
}

func NewSObjectFromMap(data map[string]interface{}) *SObject {
	return &SObject{data: data}
}

func (s *SObject) Type() string {
	if attrs := s.Attributes(); attrs != nil {
		return attrs.Type
	}
	return ""
}

func (s *SObject) ID() string { return s.StringField("Id") }

func (s *SObject) Attributes() *SObjectAttributes {
	if s.data == nil {
		return nil
	}
	if attrs, ok := s.data["attributes"]; ok {
		switch v := attrs.(type) {
		case SObjectAttributes:
			return &v
		case *SObjectAttributes:
			return v
		case map[string]interface{}:
			r := &SObjectAttributes{}
			if t, ok := v["type"].(string); ok {
				r.Type = t
			}
			if u, ok := v["url"].(string); ok {
				r.URL = u
			}
			return r
		}
	}
	return nil
}

func (s *SObject) Get(key string) interface{} {
	if s.data == nil {
		return nil
	}
	return s.data[key]
}

func (s *SObject) Set(key string, value interface{}) *SObject {
	if s.data == nil {
		s.data = make(map[string]interface{})
	}
	s.data[key] = value
	return s
}

func (s *SObject) StringField(key string) string {
	if v, ok := s.Get(key).(string); ok {
		return v
	}
	return ""
}

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

func (s *SObject) ToMap() map[string]interface{} {
	result := make(map[string]interface{}, len(s.data))
	for k, v := range s.data {
		result[k] = v
	}
	return result
}

func (s *SObject) ToCreatePayload() map[string]interface{} {
	sys := map[string]bool{"Id": true, "attributes": true, "IsDeleted": true, "CreatedDate": true, "CreatedById": true, "LastModifiedDate": true, "LastModifiedById": true, "SystemModstamp": true}
	result := make(map[string]interface{})
	for k, v := range s.data {
		if !sys[k] {
			result[k] = v
		}
	}
	return result
}

type SObjectMetadata struct {
	Name       string `json:"name"`
	Label      string `json:"label"`
	Createable bool   `json:"createable"`
	Updateable bool   `json:"updateable"`
	Deletable  bool   `json:"deletable"`
	Queryable  bool   `json:"queryable"`
}

type GlobalDescribeResult struct {
	SObjects []SObjectMetadata `json:"sobjects"`
}

type QueryResult struct {
	TotalSize      int        `json:"totalSize"`
	Done           bool       `json:"done"`
	NextRecordsURL string     `json:"nextRecordsUrl,omitempty"`
	Records        []*SObject `json:"records"`
}

func (r *QueryResult) HasMore() bool { return !r.Done && r.NextRecordsURL != "" }

type SObjectServiceImpl struct {
	client     *http.Client
	apiVersion string
}

func NewSObjectService(client *http.Client, apiVersion string) *SObjectServiceImpl {
	return &SObjectServiceImpl{client: client, apiVersion: apiVersion}
}

func (s *SObjectServiceImpl) Create(ctx context.Context, objectType string, data map[string]interface{}) (*SObject, error) {
	path := fmt.Sprintf("/services/data/v%s/sobjects/%s", s.apiVersion, objectType)
	respBody, err := s.client.Post(ctx, path, data)
	if err != nil {
		return nil, err
	}
	var result struct {
		ID      string `json:"id"`
		Success bool   `json:"success"`
		Errors  []struct {
			StatusCode string   `json:"statusCode"`
			Message    string   `json:"message"`
			Fields     []string `json:"fields"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}
	if !result.Success && len(result.Errors) > 0 {
		return nil, &types.APIError{Message: result.Errors[0].Message, ErrorCode: types.ErrorCode(result.Errors[0].StatusCode)}
	}
	obj := NewSObject(objectType)
	obj.Set("Id", result.ID)
	for k, v := range data {
		obj.Set(k, v)
	}
	return obj, nil
}

func (s *SObjectServiceImpl) Get(ctx context.Context, objectType, id string, fields ...string) (*SObject, error) {
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
		return nil, err
	}
	return NewSObjectFromMap(data), nil
}

func (s *SObjectServiceImpl) Update(ctx context.Context, objectType, id string, data map[string]interface{}) error {
	path := fmt.Sprintf("/services/data/v%s/sobjects/%s/%s", s.apiVersion, objectType, id)
	_, err := s.client.Patch(ctx, path, data)
	return err
}

func (s *SObjectServiceImpl) Upsert(ctx context.Context, objectType, extField, extID string, data map[string]interface{}) (*SObject, error) {
	path := fmt.Sprintf("/services/data/v%s/sobjects/%s/%s/%s", s.apiVersion, objectType, extField, extID)
	respBody, err := s.client.Patch(ctx, path, data)
	if err != nil {
		return nil, err
	}
	obj := NewSObject(objectType)
	obj.Set(extField, extID)
	for k, v := range data {
		obj.Set(k, v)
	}
	if len(respBody) > 0 {
		var result struct{ ID string `json:"id"` }
		json.Unmarshal(respBody, &result)
		obj.Set("Id", result.ID)
	}
	return obj, nil
}

func (s *SObjectServiceImpl) Delete(ctx context.Context, objectType, id string) error {
	path := fmt.Sprintf("/services/data/v%s/sobjects/%s/%s", s.apiVersion, objectType, id)
	_, err := s.client.Delete(ctx, path)
	return err
}

func (s *SObjectServiceImpl) Describe(ctx context.Context, objectType string) (*SObjectMetadata, error) {
	path := fmt.Sprintf("/services/data/v%s/sobjects/%s/describe", s.apiVersion, objectType)
	respBody, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var meta SObjectMetadata
	return &meta, json.Unmarshal(respBody, &meta)
}

func (s *SObjectServiceImpl) DescribeGlobal(ctx context.Context) (*GlobalDescribeResult, error) {
	path := fmt.Sprintf("/services/data/v%s/sobjects", s.apiVersion)
	respBody, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var result GlobalDescribeResult
	return &result, json.Unmarshal(respBody, &result)
}

type QueryServiceImpl struct {
	client     *http.Client
	apiVersion string
}

func NewQueryService(client *http.Client, apiVersion string) *QueryServiceImpl {
	return &QueryServiceImpl{client: client, apiVersion: apiVersion}
}

func (q *QueryServiceImpl) Execute(ctx context.Context, query string) (*QueryResult, error) {
	return q.executeQuery(ctx, fmt.Sprintf("/services/data/v%s/query?q=%s", q.apiVersion, url.QueryEscape(query)))
}

func (q *QueryServiceImpl) ExecuteAll(ctx context.Context, query string) (*QueryResult, error) {
	return q.executeQuery(ctx, fmt.Sprintf("/services/data/v%s/queryAll?q=%s", q.apiVersion, url.QueryEscape(query)))
}

func (q *QueryServiceImpl) QueryMore(ctx context.Context, nextURL string) (*QueryResult, error) {
	return q.executeQuery(ctx, nextURL)
}

func (q *QueryServiceImpl) NewBuilder(objectType string) *QueryBuilder {
	return NewQueryBuilder(objectType)
}

func (q *QueryServiceImpl) executeQuery(ctx context.Context, path string) (*QueryResult, error) {
	respBody, err := q.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var raw struct {
		TotalSize      int                      `json:"totalSize"`
		Done           bool                     `json:"done"`
		NextRecordsURL string                   `json:"nextRecordsUrl"`
		Records        []map[string]interface{} `json:"records"`
	}
	if err := json.Unmarshal(respBody, &raw); err != nil {
		return nil, err
	}
	records := make([]*SObject, len(raw.Records))
	for i, r := range raw.Records {
		records[i] = NewSObjectFromMap(r)
	}
	return &QueryResult{TotalSize: raw.TotalSize, Done: raw.Done, NextRecordsURL: raw.NextRecordsURL, Records: records}, nil
}

type ToolingServiceImpl struct {
	client     *http.Client
	apiVersion string
}

func NewToolingService(client *http.Client, apiVersion string) *ToolingServiceImpl {
	return &ToolingServiceImpl{client: client, apiVersion: apiVersion}
}

func (t *ToolingServiceImpl) Query(ctx context.Context, query string) (*QueryResult, error) {
	path := fmt.Sprintf("/services/data/v%s/tooling/query?q=%s", t.apiVersion, url.QueryEscape(query))
	respBody, err := t.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var raw struct {
		TotalSize int                      `json:"totalSize"`
		Done      bool                     `json:"done"`
		Records   []map[string]interface{} `json:"records"`
	}
	if err := json.Unmarshal(respBody, &raw); err != nil {
		return nil, err
	}
	records := make([]*SObject, len(raw.Records))
	for i, r := range raw.Records {
		records[i] = NewSObjectFromMap(r)
	}
	return &QueryResult{TotalSize: raw.TotalSize, Done: raw.Done, Records: records}, nil
}

type ExecuteAnonymousResult struct {
	Line           int         `json:"line"`
	Column         int         `json:"column"`
	Compiled       bool        `json:"compiled"`
	Success        bool        `json:"success"`
	CompileProblem interface{} `json:"compileProblem"`
}

func (t *ToolingServiceImpl) ExecuteAnonymous(ctx context.Context, apex string) (*ExecuteAnonymousResult, error) {
	path := fmt.Sprintf("/services/data/v%s/tooling/executeAnonymous/?anonymousBody=%s", t.apiVersion, url.QueryEscape(apex))
	respBody, err := t.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var result ExecuteAnonymousResult
	return &result, json.Unmarshal(respBody, &result)
}

func (t *ToolingServiceImpl) Describe(ctx context.Context, objectType string) (*SObjectMetadata, error) {
	path := fmt.Sprintf("/services/data/v%s/tooling/sobjects/%s/describe", t.apiVersion, objectType)
	respBody, err := t.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var meta SObjectMetadata
	return &meta, json.Unmarshal(respBody, &meta)
}

type ApexServiceImpl struct {
	client     *http.Client
	apiVersion string
}

func NewApexService(client *http.Client, apiVersion string) *ApexServiceImpl {
	return &ApexServiceImpl{client: client, apiVersion: apiVersion}
}

func (a *ApexServiceImpl) Execute(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if !strings.HasPrefix(path, "/services/apexrest") {
		path = "/services/apexrest" + path
	}
	var reader io.Reader
	if r, ok := body.(io.Reader); ok {
		reader = r
	}
	return a.client.Request(ctx, method, path, reader)
}

type QueryBuilder struct {
	objectType string
	fields     []string
	conditions []string
	orderBy    []string
	limit      int
	offset     int
}

func NewQueryBuilder(objectType string) *QueryBuilder {
	return &QueryBuilder{objectType: objectType}
}

func (q *QueryBuilder) Select(fields ...string) *QueryBuilder { q.fields = append(q.fields, fields...); return q }
func (q *QueryBuilder) Where(c string) *QueryBuilder          { q.conditions = append(q.conditions, c); return q }
func (q *QueryBuilder) WhereEquals(f string, v interface{}) *QueryBuilder {
	q.conditions = append(q.conditions, fmt.Sprintf("%s = %s", f, formatValue(v)))
	return q
}
func (q *QueryBuilder) WhereNotNull(f string) *QueryBuilder {
	q.conditions = append(q.conditions, f+" != NULL")
	return q
}
func (q *QueryBuilder) WhereLike(f, p string) *QueryBuilder {
	q.conditions = append(q.conditions, fmt.Sprintf("%s LIKE '%s'", f, escapeSoql(p)))
	return q
}
func (q *QueryBuilder) OrderByAsc(f string) *QueryBuilder  { q.orderBy = append(q.orderBy, f+" ASC"); return q }
func (q *QueryBuilder) OrderByDesc(f string) *QueryBuilder { q.orderBy = append(q.orderBy, f+" DESC"); return q }
func (q *QueryBuilder) Limit(l int) *QueryBuilder          { q.limit = l; return q }
func (q *QueryBuilder) Offset(o int) *QueryBuilder         { q.offset = o; return q }

func (q *QueryBuilder) Build() string {
	var sb strings.Builder
	sb.WriteString("SELECT ")
	if len(q.fields) == 0 {
		sb.WriteString("Id")
	} else {
		sb.WriteString(strings.Join(q.fields, ", "))
	}
	sb.WriteString(" FROM " + q.objectType)
	if len(q.conditions) > 0 {
		sb.WriteString(" WHERE " + strings.Join(q.conditions, " AND "))
	}
	if len(q.orderBy) > 0 {
		sb.WriteString(" ORDER BY " + strings.Join(q.orderBy, ", "))
	}
	if q.limit > 0 {
		sb.WriteString(fmt.Sprintf(" LIMIT %d", q.limit))
	}
	if q.offset > 0 {
		sb.WriteString(fmt.Sprintf(" OFFSET %d", q.offset))
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
	return strings.ReplaceAll(strings.ReplaceAll(s, "\\", "\\\\"), "'", "\\'")
}
