# gorax Action System

The action system provides a flexible, extensible framework for executing workflow actions with support for HTTP requests, data transformations, and JSONPath-based data extraction.

## Table of Contents

- [Overview](#overview)
- [Action Interface](#action-interface)
- [Built-in Actions](#built-in-actions)
  - [HTTP Action](#http-action)
  - [Transform Action](#transform-action)
- [JSONPath Support](#jsonpath-support)
- [Action Registry](#action-registry)
- [Usage Examples](#usage-examples)

## Overview

The action system is built around three core concepts:

1. **Action Interface**: Defines how actions are executed
2. **Action Registry**: Maps action types to implementations
3. **Interpolation Engine**: Supports dynamic data extraction using JSONPath

## Action Interface

All actions implement the `Action` interface:

```go
type Action interface {
    Execute(ctx context.Context, input *ActionInput) (*ActionOutput, error)
}
```

### ActionInput

Contains the configuration and context for action execution:

```go
type ActionInput struct {
    Config  interface{}            // Action-specific configuration
    Context map[string]interface{} // Data from trigger and previous steps
}
```

### ActionOutput

Contains the result of action execution:

```go
type ActionOutput struct {
    Data     interface{}            // Output data
    Metadata map[string]interface{} // Additional execution metadata
}
```

## Built-in Actions

### HTTP Action

Executes HTTP requests with full support for all HTTP methods, authentication, headers, and body.

#### Configuration

```go
type HTTPActionConfig struct {
    Method   string            // GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS
    URL      string            // Request URL (supports interpolation)
    Headers  map[string]string // Custom headers (supports interpolation)
    Body     json.RawMessage   // Request body (supports interpolation)
    Timeout  int               // Timeout in seconds (default: 30)
    Auth     *HTTPAuth         // Authentication configuration
    FollowRedirects bool       // Whether to follow redirects (default: true)
}
```

#### Authentication

Supports three authentication types:

1. **Basic Auth**:
```go
Auth: &HTTPAuth{
    Type:     "basic",
    Username: "admin",
    Password: "secret",
}
```

2. **Bearer Token**:
```go
Auth: &HTTPAuth{
    Type:  "bearer",
    Token: "{{env.API_TOKEN}}",
}
```

3. **API Key**:
```go
Auth: &HTTPAuth{
    Type:   "api_key",
    APIKey: "{{env.API_KEY}}",
    Header: "X-API-Key", // Optional, defaults to "X-API-Key"
}
```

#### Example: GET Request

```go
config := HTTPActionConfig{
    Method: "GET",
    URL:    "https://api.example.com/users/{{trigger.user_id}}",
    Headers: map[string]string{
        "Accept": "application/json",
    },
}

action := &HTTPAction{}
input := NewActionInput(config, execContext)
output, err := action.Execute(ctx, input)
```

#### Example: POST Request with Body

```go
config := HTTPActionConfig{
    Method: "POST",
    URL:    "https://api.example.com/users",
    Headers: map[string]string{
        "Content-Type": "application/json",
    },
    Body: json.RawMessage(`{
        "name": "{{trigger.name}}",
        "email": "{{trigger.email}}",
        "role": "user"
    }`),
    Auth: &HTTPAuth{
        Type:  "bearer",
        Token: "{{env.API_TOKEN}}",
    },
}
```

#### Result Structure

```go
type HTTPActionResult struct {
    StatusCode int               // HTTP status code
    Headers    map[string]string // Response headers
    Body       interface{}       // Response body (parsed as JSON if possible)
}
```

### Transform Action

Performs data transformations using JSONPath expressions or field mappings.

#### Configuration

```go
type TransformActionConfig struct {
    Expression string            // JSONPath expression to extract a value
    Mapping    map[string]string // Field mappings from source to target
    Default    interface{}       // Default value if extraction fails
}
```

#### Example: Extract Value with Expression

```go
config := TransformActionConfig{
    Expression: "steps.http-1.body.user",
}

action := &TransformAction{}
input := NewActionInput(config, execContext)
output, err := action.Execute(ctx, input)
// output.Data contains the extracted user object
```

#### Example: Map Fields

```go
config := TransformActionConfig{
    Mapping: map[string]string{
        "user_id":    "steps.http-1.body.id",
        "first_name": "trigger.user.first_name",
        "last_name":  "trigger.user.last_name",
        "email":      "trigger.user.email",
    },
}

// Result will be:
// {
//   "user_id": 123,
//   "first_name": "John",
//   "last_name": "Doe",
//   "email": "john@example.com"
// }
```

#### Example: With Default Value

```go
config := TransformActionConfig{
    Expression: "steps.http-1.body.optional_field",
    Default:    "default_value",
}
// If the field doesn't exist, "default_value" is returned
```

## JSONPath Support

The action system includes a powerful JSONPath-like interpolation engine.

### Syntax

- **Dot notation**: `trigger.user.name`
- **Array access**: `steps.http-1.body.users[0].name`
- **Nested paths**: `steps.http-1.body.data.items[2].value`
- **Escaped dots**: `user.file\.name` (for keys containing dots)

### Interpolation in Strings

Use `{{expression}}` syntax to interpolate values:

```go
url := "https://api.example.com/users/{{trigger.user_id}}/posts"
// If trigger.user_id = 123, becomes:
// "https://api.example.com/users/123/posts"
```

### Multiple Interpolations

```go
message := "User {{trigger.name}} from {{trigger.country}} logged in"
// Becomes: "User Alice from USA logged in"
```

### Available Context

The execution context contains:

- `trigger.*`: Data from the workflow trigger (webhook, schedule, etc.)
- `steps.<node-id>.*`: Output from previous workflow steps
- `env.*`: Environment variables (tenant_id, execution_id, workflow_id)

Example context structure:

```go
{
    "trigger": {
        "event": "user.created",
        "user": {
            "id": 123,
            "name": "Alice"
        }
    },
    "steps": {
        "http-1": {
            "status_code": 200,
            "body": {
                "data": [...]
            }
        }
    },
    "env": {
        "tenant_id": "tenant-123",
        "execution_id": "exec-456",
        "workflow_id": "workflow-789"
    }
}
```

## Action Registry

The registry manages action type registration and instantiation.

### Creating Actions

```go
registry := NewRegistry()

// Built-in actions are automatically registered:
// - action:http
// - action:transform

action, err := registry.Create("action:http")
if err != nil {
    // Handle unknown action type
}
```

### Registering Custom Actions

```go
type CustomAction struct{}

func (a *CustomAction) Execute(ctx context.Context, input *ActionInput) (*ActionOutput, error) {
    // Custom implementation
    return NewActionOutput(result), nil
}

// Register the custom action
registry.Register("action:custom", func() Action {
    return &CustomAction{}
})
```

### Default Registry

A global registry is available for convenience:

```go
action, err := DefaultRegistry.Create("action:http")
```

## Usage Examples

### Example 1: Webhook to HTTP Request

```go
// Workflow receives webhook with user data
// Step 1: Transform data
transformConfig := TransformActionConfig{
    Mapping: map[string]string{
        "userId":    "trigger.user.id",
        "userName":  "trigger.user.name",
        "userEmail": "trigger.user.email",
    },
}

// Step 2: Send HTTP request
httpConfig := HTTPActionConfig{
    Method: "POST",
    URL:    "https://api.example.com/users",
    Body:   json.RawMessage(`{
        "id": "{{steps.transform-1.userId}}",
        "name": "{{steps.transform-1.userName}}",
        "email": "{{steps.transform-1.userEmail}}"
    }`),
    Auth: &HTTPAuth{
        Type:  "bearer",
        Token: "{{env.API_TOKEN}}",
    },
}
```

### Example 2: Chained HTTP Requests

```go
// Step 1: Get user details
step1Config := HTTPActionConfig{
    Method: "GET",
    URL:    "https://api.example.com/users/{{trigger.user_id}}",
}

// Step 2: Use user data in another request
step2Config := HTTPActionConfig{
    Method: "POST",
    URL:    "https://api.example.com/notifications",
    Body:   json.RawMessage(`{
        "user_email": "{{steps.get-user.body.email}}",
        "message": "Hello {{steps.get-user.body.name}}!"
    }`),
}
```

### Example 3: Array Processing

```go
// Extract specific items from array
config := TransformActionConfig{
    Mapping: map[string]string{
        "first_user":  "steps.list-users.body.users[0].name",
        "second_user": "steps.list-users.body.users[1].name",
        "user_count":  "steps.list-users.body.total",
    },
}
```

### Example 4: Error Handling with Default

```go
// Safely extract optional fields
config := TransformActionConfig{
    Expression: "steps.http-1.body.optional_field",
    Default:    "N/A",
}
// If the field doesn't exist, returns "N/A" instead of error
```

## Testing

The action system includes comprehensive unit tests:

```bash
# Run all action tests
go test ./internal/executor/actions/...

# Run specific test
go test ./internal/executor/actions/... -run TestHTTPAction_Execute_GET

# Run with verbose output
go test -v ./internal/executor/actions/...
```

## Error Handling

All actions return descriptive errors:

```go
output, err := action.Execute(ctx, input)
if err != nil {
    // Errors are wrapped with context
    // Example: "failed to create request: invalid URL"
    log.Error("Action execution failed", "error", err)
}
```

### Timeout Support

Actions respect context cancellation and timeouts:

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

output, err := action.Execute(ctx, input)
// Will timeout after 10 seconds
```

## Performance Considerations

- **HTTP Client Reuse**: Each HTTP action creates a client with configured timeout
- **Context Cancellation**: Always pass context for proper cancellation handling
- **Memory**: Large response bodies are loaded into memory; consider streaming for large files
- **Concurrency**: Actions are safe for concurrent use

## Security Considerations

- **Secrets**: Use environment variables for API keys and tokens
- **URL Validation**: HTTP action validates URLs before making requests
- **Timeout Protection**: All actions support timeouts to prevent hanging
- **Error Messages**: Errors don't leak sensitive configuration data

## Future Enhancements

Planned features:

- [ ] Code action (sandboxed JavaScript/Lua execution)
- [ ] Email action (send emails via SMTP or API)
- [ ] Database action (query SQL databases)
- [ ] File action (read/write files, S3 operations)
- [ ] Advanced JSONPath (filters, functions, wildcards)
- [ ] Action retries with exponential backoff
- [ ] Action caching (memoization)
- [ ] Parallel action execution
