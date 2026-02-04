# Salesforce Go SDK

A production-grade Go SDK for Salesforce with OAuth 2.0 authentication support.

## Installation

```bash
go get github.com/PramithaMJ/salesforce
```

## Features

- **Multiple Authentication Methods**: OAuth 2.0 refresh token, password flow, or direct token
- **Complete API Coverage**: SObject CRUD, SOQL queries, Bulk API 2.0, Tooling API, Apex REST

## Quick Start

```go
package main

import (
    "context"
    "log"
    
    sf "github.com/PramithaMJ/salesforce"
)

func main() {
    client, err := sf.NewClient(
        sf.WithOAuthRefresh(
            "your-client-id",
            "your-client-secret", 
            "your-refresh-token",
        ),
        sf.WithTokenURL("https://login.salesforce.com/services/oauth2/token"),
    )
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()
    if err := client.Connect(ctx); err != nil {
        log.Fatal(err)
    }

    // Create a record
    account, err := client.SObjects().Create(ctx, "Account", map[string]interface{}{
        "Name": "Test Account",
    })
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Created account: %s", account.ID())

    // Query records
    result, err := client.Query().Execute(ctx, "SELECT Id, Name FROM Account LIMIT 10")
    if err != nil {
        log.Fatal(err)
    }
    for _, record := range result.Records {
        log.Printf("Account: %s - %s", record.ID(), record.StringField("Name"))
    }
}
```

## Authentication Options

### OAuth 2.0 Refresh Token
```go
client, _ := sf.NewClient(
    sf.WithOAuthRefresh(clientID, clientSecret, refreshToken),
    sf.WithTokenURL("https://login.salesforce.com/services/oauth2/token"),
)
```

### Username/Password
```go
client, _ := sf.NewClient(
    sf.WithPasswordAuth(username, password, securityToken),
    sf.WithCredentials(clientID, clientSecret),
)
```

### Direct Token
```go
client, _ := sf.NewClient(
    sf.WithInstanceURL("https://yourinstance.salesforce.com"),
)
client.SetAccessToken(accessToken, instanceURL)
```

## Query Builder

```go
query := sf.NewQueryBuilder("Contact").
    Select("Id", "FirstName", "LastName", "Email").
    WhereEquals("AccountId", accountID).
    WhereLike("Email", "%@example.com").
    OrderByDesc("CreatedDate").
    Limit(100).
    Build()

result, _ := client.Query().Execute(ctx, query)
```

## Bulk API 2.0

```go
job, _ := client.Bulk().CreateJob(ctx, sf.CreateJobRequest{
    Object:    "Account",
    Operation: services.JobOperationInsert,
})
// Upload data, close job, wait for completion...
```

## License

MIT License
