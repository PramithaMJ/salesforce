# Salesforce Go SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/PramithaMJ/salesforce.svg)](https://pkg.go.dev/github.com/PramithaMJ/salesforce)

A comprehensive, production-grade Go SDK for Salesforce covering all major APIs with SOLID principles.

## Installation

```bash
go get github.com/PramithaMJ/salesforce
```

## Features

- **Multiple Authentication Methods**: OAuth 2.0 refresh token, username-password, and direct access token
- **Complete API Coverage**:
  - SObject CRUD operations
  - SOQL queries with builder pattern
  - SOSL search with builder pattern  
  - Bulk API 2.0 (ingest and query)
  - Composite API (batch, tree, graph, collections)
  - Analytics (Reports & Dashboards)
  - Tooling API (Apex, tests, metadata)
  - Connect/Chatter API
  - User Interface API
  - Apex REST endpoints
  - API Limits monitoring
- **Production Ready**: Retry with exponential backoff, comprehensive error handling, context support

## Quick Start

### OAuth 2.0 Refresh Token

```go
client, _ := salesforce.NewClient(
    salesforce.WithOAuthRefresh(
        os.Getenv("SF_CLIENT_ID"),
        os.Getenv("SF_CLIENT_SECRET"),
        os.Getenv("SF_REFRESH_TOKEN"),
    ),
)
client.Connect(context.Background())
```

### Username-Password

```go
client, _ := salesforce.NewClient(
    salesforce.WithCredentials(clientID, clientSecret),
    salesforce.WithPasswordAuth(username, password, securityToken),
)
client.Connect(context.Background())
```

### Direct Access Token

```go
client, _ := salesforce.NewClient(
    salesforce.WithAccessToken(accessToken, instanceURL),
)
client.SetAccessToken(accessToken, instanceURL)
```

## Usage Examples

### Query Records

```go
// Simple query
result, _ := client.Query().Execute(ctx, "SELECT Id, Name FROM Account LIMIT 10")
for _, record := range result.Records {
    fmt.Println(record.StringField("Name"))
}

// Query builder
query := client.Query().NewBuilder("Contact").
    Select("Id", "FirstName", "LastName", "Email").
    WhereEquals("AccountId", accountID).
    OrderByAsc("LastName").
    Limit(50).
    Build()
result, _ := client.Query().Execute(ctx, query)
```

### Create/Update Records

```go
// Create
result, _ := client.SObjects().Create(ctx, "Account", map[string]interface{}{
    "Name": "Acme Corp",
    "Industry": "Technology",
})
fmt.Println("Created:", result.ID)

// Update
client.SObjects().Update(ctx, "Account", accountID, map[string]interface{}{
    "Description": "Updated description",
})

// Upsert by external ID
client.SObjects().Upsert(ctx, "Account", "External_ID__c", "EXT-001", data)
```

### Bulk Operations

```go
// Create bulk job
job, _ := client.Bulk().CreateJob(ctx, bulk.CreateJobRequest{
    Object:    "Account",
    Operation: bulk.OperationInsert,
})

// Upload data
client.Bulk().UploadCSV(ctx, job.ID, records, []string{"Name", "Industry"})

// Close and wait
client.Bulk().CloseJob(ctx, job.ID)
job, _ = client.Bulk().WaitForCompletion(ctx, job.ID, 5*time.Second)
```

### Composite API

```go
resp, _ := client.Composite().Execute(ctx, composite.Request{
    AllOrNone: true,
    CompositeRequest: []composite.Subrequest{
        {Method: "POST", URL: "/services/data/v59.0/sobjects/Account",
         ReferenceId: "newAccount", Body: map[string]interface{}{"Name": "New Account"}},
        {Method: "POST", URL: "/services/data/v59.0/sobjects/Contact",
         ReferenceId: "newContact", Body: map[string]interface{}{
             "LastName": "Smith", "AccountId": "@{newAccount.id}"}},
    },
})
```

### Search (SOSL)

```go
// Direct SOSL
result, _ := client.Search().Execute(ctx, "FIND {Acme} IN NAME FIELDS RETURNING Account(Id, Name)")

// Search builder
sosl := search.NewBuilder("Acme").
    In("NAME").
    ReturningWithFields("Account", "Id", "Name").
    ReturningWithFields("Contact", "Id", "FirstName", "LastName").
    Limit(20).
    Build()
result, _ := client.Search().Execute(ctx, sosl)
```

### Tooling API

```go
// Execute anonymous Apex
result, _ := client.Tooling().ExecuteAnonymous(ctx, `System.debug('Hello');`)

// Query Apex classes
classes, _ := client.Tooling().Query(ctx, "SELECT Id, Name FROM ApexClass LIMIT 10")
```

### Analytics

```go
// Run report
result, _ := client.Analytics().RunReport(ctx, reportID, true)

// List dashboards
dashboards, _ := client.Analytics().ListDashboards(ctx)
```

### Chatter

```go
// Get news feed
feed, _ := client.Chatter().GetNewsFeed(ctx)

// Post to feed
client.Chatter().PostFeedElement(ctx, connect.FeedInput{
    SubjectId: userID,
    Body: connect.MessageBodyInput{
        MessageSegments: []connect.MessageSegmentInput{
            {Type: "Text", Text: "Hello from Go SDK!"},
        },
    },
})
```

### Limits

```go
limits, _ := client.Limits().GetLimits(ctx)
fmt.Printf("API Requests: %d/%d (%.1f%% used)\n",
    limits.DailyApiRequests.Used(),
    limits.DailyApiRequests.Max,
    limits.DailyApiRequests.PercentUsed())
```

### Apex REST

```go
// Call custom Apex REST endpoint
var result MyResponse
client.Apex().GetJSON(ctx, "/MyEndpoint/v1/data", &result)

// POST to endpoint
client.Apex().PostJSON(ctx, "/MyEndpoint/v1/process", requestData, &result)
```

## Configuration Options

| Option | Description |
|--------|-------------|
| `WithOAuthRefresh` | OAuth 2.0 refresh token flow |
| `WithPasswordAuth` | Username-password flow |
| `WithCredentials` | OAuth client credentials |
| `WithAccessToken` | Direct access token |
| `WithTokenURL` | Custom OAuth token endpoint |
| `WithAPIVersion` | API version (default: 59.0) |
| `WithTimeout` | HTTP timeout |
| `WithMaxRetries` | Retry attempts (default: 3) |
| `WithHTTPClient` | Custom HTTP client |
| `WithLogger` | Custom logger |
| `WithSandbox` | Use sandbox environment |
| `WithCustomDomain` | My Domain configuration |

## Error Handling

```go
result, err := client.SObjects().Get(ctx, "Account", "invalid-id")
if err != nil {
    if types.IsNotFoundError(err) {
        // Handle not found
    } else if types.IsAuthError(err) {
        // Handle auth error - maybe refresh token
    } else if types.IsRateLimitError(err) {
        // Handle rate limit - back off
    }
}
```

## License

MIT License - see [LICENSE](LICENSE)
