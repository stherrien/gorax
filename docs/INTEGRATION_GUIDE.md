# Gorax Integration Development Guide

This comprehensive guide teaches developers how to extend Gorax with new integrations, actions, and external service connectors.

## Table of Contents
1. [Overview](#overview)
2. [Action Development](#action-development)
3. [Integration Structure](#integration-structure)
4. [Common Integration Patterns](#common-integration-patterns)
5. [LLM Provider Integration](#llm-provider-integration)
6. [Credential Management](#credential-management)
7. [Expression Evaluation](#expression-evaluation)
8. [Real-World Examples](#real-world-examples)
9. [Testing Integrations](#testing-integrations)
10. [Publishing to Marketplace](#publishing-to-marketplace)

---

## Overview

### What are Integrations in Gorax?

Integrations in Gorax are connectors that allow workflows to interact with external services, APIs, and systems. They provide:

- **Actions**: Executable operations (e.g., "Send Slack message", "Upload to S3")
- **Triggers**: Event sources that start workflow execution (e.g., "Webhook received", "Schedule")
- **LLM Providers**: AI/LLM service integrations (e.g., OpenAI, Anthropic, AWS Bedrock)

### Types of Integrations

#### 1. **Built-in Actions** (`internal/executor/actions/`)
Core actions available in every workflow:
- `action:http` - HTTP requests
- `action:transform` - Data transformation
- `action:formula` - Expression evaluation
- `action:code` - Script execution

#### 2. **Service Integrations** (`internal/integrations/`)
Third-party service connectors:
- `slack:send_message` - Slack messaging
- `aws:s3:get_object` - AWS S3 operations
- `github:create_issue` - GitHub API
- `ai:chat_completion` - LLM chat completions

#### 3. **LLM Providers** (`internal/llm/`)
AI model providers:
- `openai` - OpenAI GPT models
- `anthropic` - Anthropic Claude models
- `bedrock` - AWS Bedrock models

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                      Workflow Executor                       │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────────┐      ┌──────────────┐      ┌───────────┐ │
│  │ Action       │      │ Credential   │      │ Expression│ │
│  │ Registry     │─────▶│ Injector     │─────▶│ Evaluator │ │
│  └──────────────┘      └──────────────┘      └───────────┘ │
│         │                      │                      │      │
└─────────┼──────────────────────┼──────────────────────┼─────┘
          │                      │                      │
          ▼                      ▼                      ▼
┌─────────────────┐    ┌──────────────────┐   ┌────────────┐
│  Built-in       │    │  Service         │   │  LLM       │
│  Actions        │    │  Integrations    │   │  Providers │
├─────────────────┤    ├──────────────────┤   ├────────────┤
│ • HTTP          │    │ • Slack          │   │ • OpenAI   │
│ • Transform     │    │ • AWS S3/SNS     │   │ • Anthropic│
│ • Formula       │    │ • GitHub         │   │ • Bedrock  │
│ • Code          │    │ • Google APIs    │   │ • Custom   │
└─────────────────┘    └──────────────────┘   └────────────┘
```

---

## Action Development

### Action Interface and Lifecycle

All actions must implement the `Action` interface:

```go
// internal/executor/actions/action.go
type Action interface {
    // Execute runs the action with the given context and input
    Execute(ctx context.Context, input *ActionInput) (*ActionOutput, error)
}

// ActionInput represents the input data for an action execution
type ActionInput struct {
    Config  interface{}            // Action-specific configuration
    Context map[string]interface{} // Data from previous steps and trigger
}

// ActionOutput represents the result of an action execution
type ActionOutput struct {
    Data     interface{}            // Output data from the action
    Metadata map[string]interface{} // Additional execution information
}
```

### Creating a New Action Type

Let's create a custom "SendEmail" action from scratch.

#### Step 1: Define Configuration Structures

```go
// internal/integrations/email/models.go
package email

// SendEmailConfig represents the configuration for sending an email
type SendEmailConfig struct {
    To      []string `json:"to" validate:"required"`
    Subject string   `json:"subject" validate:"required"`
    Body    string   `json:"body" validate:"required"`
    From    string   `json:"from,omitempty"`
    CC      []string `json:"cc,omitempty"`
    BCC     []string `json:"bcc,omitempty"`
}

// Validate validates the configuration
func (c *SendEmailConfig) Validate() error {
    if len(c.To) == 0 {
        return fmt.Errorf("at least one recipient is required")
    }
    if c.Subject == "" {
        return fmt.Errorf("subject is required")
    }
    if c.Body == "" {
        return fmt.Errorf("body is required")
    }
    return nil
}
```

#### Step 2: Implement the Action

```go
// internal/integrations/email/send_email.go
package email

import (
    "context"
    "fmt"
    "encoding/json"

    "github.com/gorax/gorax/internal/credential"
    "github.com/gorax/gorax/internal/executor/actions"
)

// SendEmailAction implements the email:send_email action
type SendEmailAction struct {
    credentialService credential.Service
    smtpHost          string
    smtpPort          int
}

// NewSendEmailAction creates a new SendEmail action
func NewSendEmailAction(credentialService credential.Service) *SendEmailAction {
    return &SendEmailAction{
        credentialService: credentialService,
        smtpHost:          "smtp.gmail.com",
        smtpPort:          587,
    }
}

// Execute implements the Action interface
func (a *SendEmailAction) Execute(ctx context.Context, input *actions.ActionInput) (*actions.ActionOutput, error) {
    // Parse config
    configBytes, err := json.Marshal(input.Config)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal config: %w", err)
    }

    var config SendEmailConfig
    if err := json.Unmarshal(configBytes, &config); err != nil {
        return nil, fmt.Errorf("invalid config type: expected SendEmailConfig: %w", err)
    }

    // Validate config
    if err := config.Validate(); err != nil {
        return nil, fmt.Errorf("invalid configuration: %w", err)
    }

    // Extract tenant_id and credential_id from context
    tenantID, err := extractString(input.Context, "env.tenant_id")
    if err != nil {
        return nil, fmt.Errorf("tenant_id is required in context: %w", err)
    }

    credentialID, err := extractString(input.Context, "credential_id")
    if err != nil {
        return nil, fmt.Errorf("credential_id is required in context: %w", err)
    }

    // Retrieve and decrypt credential
    decryptedCred, err := a.credentialService.GetValue(ctx, tenantID, credentialID, "system")
    if err != nil {
        return nil, fmt.Errorf("failed to retrieve credential: %w", err)
    }

    // Extract SMTP credentials
    username, _ := decryptedCred.Value["username"].(string)
    password, _ := decryptedCred.Value["password"].(string)

    // Send email using SMTP client
    if err := a.sendViaSMTP(ctx, config, username, password); err != nil {
        return nil, fmt.Errorf("failed to send email: %w", err)
    }

    // Build result
    result := map[string]interface{}{
        "success":    true,
        "recipients": config.To,
        "subject":    config.Subject,
    }

    // Create output with metadata
    output := actions.NewActionOutput(result)
    output.WithMetadata("recipient_count", len(config.To))
    output.WithMetadata("has_cc", len(config.CC) > 0)

    return output, nil
}

// sendViaSMTP sends the email using SMTP
func (a *SendEmailAction) sendViaSMTP(ctx context.Context, config SendEmailConfig, username, password string) error {
    // Implementation details...
    // Use net/smtp or a library like gomail
    return nil
}

// Helper function to extract string from nested context
func extractString(data map[string]interface{}, path string) (string, error) {
    keys := strings.Split(path, ".")
    current := data

    for i, key := range keys {
        if i == len(keys)-1 {
            if val, ok := current[key]; ok {
                if str, ok := val.(string); ok {
                    return str, nil
                }
                return "", fmt.Errorf("value at '%s' is not a string", path)
            }
            return "", fmt.Errorf("key '%s' not found in context", path)
        }

        if val, ok := current[key]; ok {
            if m, ok := val.(map[string]interface{}); ok {
                current = m
            } else {
                return "", fmt.Errorf("value at '%s' is not a map", key)
            }
        } else {
            return "", fmt.Errorf("key '%s' not found in context", key)
        }
    }

    return "", fmt.Errorf("failed to extract value from path '%s'", path)
}
```

#### Step 3: Register the Action

```go
// internal/integrations/email/registry.go
package email

import (
    "github.com/gorax/gorax/internal/credential"
    "github.com/gorax/gorax/internal/executor/actions"
)

// RegisterEmailActions registers email actions with the action registry
func RegisterEmailActions(registry *actions.Registry, credService credential.Service) error {
    registry.Register("email:send", func() actions.Action {
        return NewSendEmailAction(credService)
    })
    return nil
}
```

### Input/Output Handling

#### Accessing Execution Context

The `ActionInput.Context` map contains data from the workflow execution:

```go
// Available context data:
context := map[string]interface{}{
    "trigger": map[string]interface{}{
        "type": "webhook",
        "body": map[string]interface{}{
            "user_id": "123",
            "event": "order.created",
        },
    },
    "steps": map[string]interface{}{
        "fetch_user": map[string]interface{}{
            "output": map[string]interface{}{
                "name": "John Doe",
                "email": "john@example.com",
            },
        },
    },
    "env": map[string]interface{}{
        "tenant_id": "tenant-123",
        "execution_id": "exec-456",
    },
}

// Extract values
userID := input.Context["trigger"].(map[string]interface{})["body"].(map[string]interface{})["user_id"]
userName := input.Context["steps"].(map[string]interface{})["fetch_user"].(map[string]interface{})["output"].(map[string]interface{})["name"]
```

#### Returning Output Data

```go
// Simple output
return actions.NewActionOutput(map[string]interface{}{
    "status": "success",
    "message_id": "msg-123",
}), nil

// Output with metadata
output := actions.NewActionOutput(result)
output.WithMetadata("provider", "slack")
output.WithMetadata("channel_id", "C123456")
output.WithMetadata("timestamp", time.Now())
return output, nil
```

### Error Handling Patterns

#### Wrapped Errors

Always wrap errors with context:

```go
if err != nil {
    return nil, fmt.Errorf("failed to send email: %w", err)
}
```

#### Custom Error Types

```go
// internal/integrations/email/errors.go
package email

import "errors"

var (
    ErrInvalidRecipient = errors.New("invalid recipient email address")
    ErrSMTPAuth         = errors.New("SMTP authentication failed")
    ErrRateLimited      = errors.New("rate limit exceeded")
)

// IsRetryable determines if an error should trigger a retry
func IsRetryable(err error) bool {
    return errors.Is(err, ErrRateLimited) || errors.Is(err, ErrSMTPTimeout)
}
```

### Testing Actions

#### Unit Test with Mocks

```go
// internal/integrations/email/send_email_test.go
package email

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"

    "github.com/gorax/gorax/internal/credential"
    "github.com/gorax/gorax/internal/executor/actions"
)

// MockCredentialService is a mock implementation
type MockCredentialService struct {
    mock.Mock
}

func (m *MockCredentialService) GetValue(ctx context.Context, tenantID, credentialID, userID string) (*credential.DecryptedValue, error) {
    args := m.Called(ctx, tenantID, credentialID, userID)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*credential.DecryptedValue), args.Error(1)
}

func TestSendEmailAction_Execute(t *testing.T) {
    // Setup
    mockCredService := new(MockCredentialService)
    action := NewSendEmailAction(mockCredService)

    // Mock credential service response
    mockCredService.On("GetValue", mock.Anything, "tenant-123", "cred-456", "system").
        Return(&credential.DecryptedValue{
            Value: map[string]interface{}{
                "username": "test@example.com",
                "password": "secret123",
            },
        }, nil)

    // Prepare input
    config := SendEmailConfig{
        To:      []string{"recipient@example.com"},
        Subject: "Test Email",
        Body:    "This is a test",
    }

    input := actions.NewActionInput(config, map[string]interface{}{
        "env": map[string]interface{}{
            "tenant_id": "tenant-123",
        },
        "credential_id": "cred-456",
    })

    // Execute
    output, err := action.Execute(context.Background(), input)

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, output)
    assert.True(t, output.Data.(map[string]interface{})["success"].(bool))
    mockCredService.AssertExpectations(t)
}
```

---

## Integration Structure

### File Organization

Organize integrations by service:

```
internal/integrations/
├── integration.go           # Common interfaces
├── registry.go              # Global action registry
├── retry.go                 # Retry utilities
├── slack/
│   ├── client.go           # Slack API client
│   ├── client_test.go
│   ├── models.go           # Data structures
│   ├── errors.go           # Error types
│   ├── send_message.go     # SendMessage action
│   ├── send_message_test.go
│   ├── add_reaction.go     # AddReaction action
│   └── add_reaction_test.go
├── aws/
│   ├── s3.go               # S3 actions
│   ├── s3_test.go
│   ├── sns.go              # SNS actions
│   ├── lambda.go           # Lambda actions
│   └── registry.go         # AWS action registration
└── email/
    ├── client.go
    ├── send_email.go
    └── send_email_test.go
```

### Configuration and Credentials

#### Credential Structure

```go
// Stored in credential vault
{
    "type": "slack_oauth",
    "value": {
        "access_token": "xoxb-...",
        "team_id": "T123456",
        "bot_user_id": "U123456"
    }
}

// AWS credentials
{
    "type": "aws_access_key",
    "value": {
        "access_key_id": "AKIAIOSFODNN7EXAMPLE",
        "secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
        "region": "us-east-1"
    }
}
```

### Service Layer Patterns

#### Client Abstraction

```go
// internal/integrations/slack/client.go
package slack

import (
    "context"
    "net/http"
    "time"
)

type Client struct {
    accessToken string
    baseURL     string
    httpClient  *http.Client
    maxRetries  int
}

func NewClient(accessToken string) (*Client, error) {
    if accessToken == "" {
        return nil, ErrInvalidToken
    }

    return &Client{
        accessToken: accessToken,
        baseURL:     "https://slack.com/api",
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
        maxRetries: 3,
    }, nil
}

// SendMessage sends a message to a Slack channel
func (c *Client) SendMessage(ctx context.Context, req *SendMessageRequest) (*MessageResponse, error) {
    var resp MessageResponse
    if err := c.doRequest(ctx, "POST", "/chat.postMessage", req, &resp); err != nil {
        return nil, err
    }

    if !resp.OK {
        return nil, ParseSlackError(resp.Error)
    }

    return &resp, nil
}
```

---

## Common Integration Patterns

### HTTP/REST API Integrations

#### Basic HTTP Client Pattern

```go
// internal/integrations/myservice/client.go
package myservice

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
)

type Client struct {
    apiKey     string
    baseURL    string
    httpClient *http.Client
}

func NewClient(apiKey string) *Client {
    return &Client{
        apiKey:  apiKey,
        baseURL: "https://api.myservice.com/v1",
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}

func (c *Client) doRequest(ctx context.Context, method, endpoint string, body interface{}, result interface{}) error {
    // Prepare request body
    var bodyReader io.Reader
    if body != nil {
        bodyJSON, err := json.Marshal(body)
        if err != nil {
            return fmt.Errorf("failed to marshal request body: %w", err)
        }
        bodyReader = bytes.NewReader(bodyJSON)
    }

    // Create request
    req, err := http.NewRequestWithContext(ctx, method, c.baseURL+endpoint, bodyReader)
    if err != nil {
        return fmt.Errorf("failed to create request: %w", err)
    }

    // Set headers
    req.Header.Set("Authorization", "Bearer "+c.apiKey)
    req.Header.Set("Content-Type", "application/json")

    // Execute request
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()

    // Check status code
    if resp.StatusCode >= 400 {
        return c.parseError(resp)
    }

    // Parse response
    if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
        return fmt.Errorf("failed to decode response: %w", err)
    }

    return nil
}

func (c *Client) parseError(resp *http.Response) error {
    body, _ := io.ReadAll(resp.Body)
    return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
}
```

### OAuth 2.0 Authentication Flow

#### OAuth Client Implementation

```go
// internal/integrations/myservice/oauth.go
package myservice

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"
    "strings"
)

type OAuthClient struct {
    clientID     string
    clientSecret string
    redirectURI  string
    baseURL      string
}

// GetAuthorizationURL generates the OAuth authorization URL
func (o *OAuthClient) GetAuthorizationURL(state string, scopes []string) string {
    params := url.Values{
        "client_id":     {o.clientID},
        "redirect_uri":  {o.redirectURI},
        "response_type": {"code"},
        "scope":         {strings.Join(scopes, " ")},
        "state":         {state},
    }
    return fmt.Sprintf("%s/oauth/authorize?%s", o.baseURL, params.Encode())
}

// ExchangeCode exchanges an authorization code for an access token
func (o *OAuthClient) ExchangeCode(ctx context.Context, code string) (*TokenResponse, error) {
    data := url.Values{
        "grant_type":    {"authorization_code"},
        "code":          {code},
        "client_id":     {o.clientID},
        "client_secret": {o.clientSecret},
        "redirect_uri":  {o.redirectURI},
    }

    req, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/oauth/token",
        strings.NewReader(data.Encode()))
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }

    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

    client := &http.Client{Timeout: 30 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("token exchange failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("token exchange failed with status: %d", resp.StatusCode)
    }

    var tokenResp TokenResponse
    if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
        return nil, fmt.Errorf("failed to decode token response: %w", err)
    }

    return &tokenResp, nil
}

// RefreshToken refreshes an expired access token
func (o *OAuthClient) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
    data := url.Values{
        "grant_type":    {"refresh_token"},
        "refresh_token": {refreshToken},
        "client_id":     {o.clientID},
        "client_secret": {o.clientSecret},
    }

    req, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/oauth/token",
        strings.NewReader(data.Encode()))
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }

    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

    client := &http.Client{Timeout: 30 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("token refresh failed: %w", err)
    }
    defer resp.Body.Close()

    var tokenResp TokenResponse
    if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
        return nil, fmt.Errorf("failed to decode token response: %w", err)
    }

    return &tokenResp, nil
}

type TokenResponse struct {
    AccessToken  string `json:"access_token"`
    TokenType    string `json:"token_type"`
    ExpiresIn    int    `json:"expires_in"`
    RefreshToken string `json:"refresh_token,omitempty"`
    Scope        string `json:"scope,omitempty"`
}
```

### Webhook Receivers

Webhook receivers are handled at the API layer:

```go
// internal/api/handlers/webhook_handler.go
// See existing implementation for webhook handling patterns
```

### Polling vs Push Patterns

#### Polling Pattern (for APIs without webhooks)

```go
// internal/integrations/myservice/poller.go
package myservice

import (
    "context"
    "time"
)

type Poller struct {
    client   *Client
    interval time.Duration
    lastPoll time.Time
}

func NewPoller(client *Client, interval time.Duration) *Poller {
    return &Poller{
        client:   client,
        interval: interval,
    }
}

func (p *Poller) Poll(ctx context.Context) ([]Event, error) {
    // Fetch events since last poll
    events, err := p.client.GetEvents(ctx, p.lastPoll)
    if err != nil {
        return nil, err
    }

    p.lastPoll = time.Now()
    return events, nil
}
```

### Rate Limiting and Retries

#### Exponential Backoff Retry

```go
// internal/integrations/retry.go
package integrations

import (
    "context"
    "time"
)

// WithRetry executes a function with exponential backoff retry logic
func WithRetry(ctx context.Context, config RetryConfig, fn RetryableFunc) error {
    var lastErr error

    for attempt := 0; attempt < config.MaxAttempts; attempt++ {
        // Check context cancellation
        if ctx.Err() != nil {
            return ctx.Err()
        }

        // Execute the function
        err := fn()
        if err == nil {
            return nil
        }

        lastErr = err

        // Don't retry if error is not retryable
        if !IsRetryableError(err) {
            return err
        }

        // Don't sleep after the last attempt
        if attempt < config.MaxAttempts-1 {
            delay := calculateDelay(attempt, config)
            select {
            case <-ctx.Done():
                return ctx.Err()
            case <-time.After(delay):
                // Continue to next attempt
            }
        }
    }

    return fmt.Errorf("max retry attempts exceeded: %w", lastErr)
}

// calculateDelay calculates the delay for exponential backoff
func calculateDelay(attempt int, config RetryConfig) time.Duration {
    if attempt < 0 {
        attempt = 0
    }
    if attempt > 30 {
        attempt = 30
    }
    // Exponential backoff: baseDelay * 2^attempt
    delay := config.BaseDelay * time.Duration(1<<attempt)

    // Cap at max delay
    if delay > config.MaxDelay {
        return config.MaxDelay
    }

    return delay
}
```

---

## LLM Provider Integration

### Provider Interface

All LLM providers implement the `Provider` interface:

```go
// internal/llm/provider.go
type Provider interface {
    // ChatCompletion performs a chat-style completion
    ChatCompletion(ctx context.Context, req *ChatRequest) (*ChatResponse, error)

    // GenerateEmbeddings generates embeddings for input texts
    GenerateEmbeddings(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error)

    // CountTokens estimates the token count for the given text and model
    CountTokens(text string, model string) (int, error)

    // ListModels returns available models for this provider
    ListModels(ctx context.Context) ([]Model, error)

    // Name returns the provider name (e.g., "openai", "anthropic", "bedrock")
    Name() string

    // HealthCheck verifies the provider connection is valid
    HealthCheck(ctx context.Context) error
}
```

### Adding a New LLM Provider

Let's add a Cohere provider as an example.

#### Step 1: Create Provider Implementation

```go
// internal/llm/providers/cohere/provider.go
package cohere

import (
    "context"
    "fmt"

    "github.com/gorax/gorax/internal/llm"
)

type CohereProvider struct {
    apiKey     string
    baseURL    string
    httpClient *http.Client
}

func NewCohereProvider(config *llm.ProviderConfig) (llm.Provider, error) {
    if config.APIKey == "" {
        return nil, fmt.Errorf("API key is required")
    }

    baseURL := "https://api.cohere.ai/v1"
    if config.BaseURL != "" {
        baseURL = config.BaseURL
    }

    return &CohereProvider{
        apiKey:  config.APIKey,
        baseURL: baseURL,
        httpClient: &http.Client{
            Timeout: config.Timeout,
        },
    }, nil
}

func (p *CohereProvider) Name() string {
    return "cohere"
}

func (p *CohereProvider) ChatCompletion(ctx context.Context, req *llm.ChatRequest) (*llm.ChatResponse, error) {
    // Convert request to Cohere format
    cohereReq := p.buildCohereRequest(req)

    // Make API call
    cohereResp, err := p.doRequest(ctx, "/chat", cohereReq)
    if err != nil {
        return nil, err
    }

    // Convert response to standard format
    return p.parseCohereResponse(cohereResp)
}

func (p *CohereProvider) GenerateEmbeddings(ctx context.Context, req *llm.EmbeddingRequest) (*llm.EmbeddingResponse, error) {
    // Implementation...
    return nil, fmt.Errorf("not implemented")
}

func (p *CohereProvider) CountTokens(text string, model string) (int, error) {
    // Rough estimation (provider-specific)
    return len(text) / 4, nil
}

func (p *CohereProvider) ListModels(ctx context.Context) ([]llm.Model, error) {
    return []llm.Model{
        {
            ID:            "command-r-plus",
            Name:          "Command R+",
            Provider:      "cohere",
            MaxTokens:     4096,
            ContextWindow: 128000,
        },
    }, nil
}

func (p *CohereProvider) HealthCheck(ctx context.Context) error {
    _, err := p.ListModels(ctx)
    return err
}

// Helper methods
func (p *CohereProvider) buildCohereRequest(req *llm.ChatRequest) map[string]interface{} {
    // Convert standard ChatRequest to Cohere-specific format
    return map[string]interface{}{
        "model":   req.Model,
        "message": req.Messages[len(req.Messages)-1].Content,
        // ... other fields
    }
}

func (p *CohereProvider) parseCohereResponse(resp map[string]interface{}) (*llm.ChatResponse, error) {
    // Convert Cohere response to standard format
    return &llm.ChatResponse{
        ID:    resp["id"].(string),
        Model: resp["model"].(string),
        Message: llm.ChatMessage{
            Role:    "assistant",
            Content: resp["text"].(string),
        },
        FinishReason: "stop",
        Usage: llm.TokenUsage{
            PromptTokens:     int(resp["meta"].(map[string]interface{})["tokens"].(map[string]interface{})["input_tokens"].(float64)),
            CompletionTokens: int(resp["meta"].(map[string]interface{})["tokens"].(map[string]interface{})["output_tokens"].(float64)),
        },
    }, nil
}
```

#### Step 2: Register the Provider

```go
// internal/llm/providers/cohere/init.go
package cohere

import (
    "github.com/gorax/gorax/internal/llm"
)

func init() {
    // Register Cohere provider with the global registry
    llm.RegisterProvider("cohere", func(config *llm.ProviderConfig) (llm.Provider, error) {
        return NewCohereProvider(config)
    })
}
```

#### Step 3: Use the Provider

```go
// In your application startup
import _ "github.com/gorax/gorax/internal/llm/providers/cohere"

// In action execution
provider, err := llm.GetGlobalProvider("cohere", &llm.ProviderConfig{
    APIKey: "your-api-key",
})

resp, err := provider.ChatCompletion(ctx, &llm.ChatRequest{
    Model: "command-r-plus",
    Messages: []llm.ChatMessage{
        {Role: "user", Content: "Hello!"},
    },
})
```

### Request/Response Mapping

Each provider has unique request/response formats. Example mappings:

```go
// OpenAI format
{
    "model": "gpt-4o",
    "messages": [{"role": "user", "content": "Hello"}],
    "temperature": 0.7
}

// Anthropic format
{
    "model": "claude-3-sonnet-20240229",
    "max_tokens": 1024,
    "messages": [{"role": "user", "content": "Hello"}]
}

// Cohere format
{
    "model": "command-r-plus",
    "message": "Hello",
    "temperature": 0.7
}
```

### Token Tracking

```go
// internal/llm/types.go
type TokenUsage struct {
    PromptTokens     int `json:"prompt_tokens"`
    CompletionTokens int `json:"completion_tokens"`
    TotalTokens      int `json:"total_tokens"`
}

// Track usage in response
return &ChatResponse{
    // ... other fields
    Usage: TokenUsage{
        PromptTokens:     100,
        CompletionTokens: 50,
        TotalTokens:      150,
    },
}
```

---

## Credential Management

### Using the Credential Service

#### Injecting Credentials into Actions

The credential injector automatically replaces `{{credentials.name}}` references:

```go
// internal/credential/injector.go

// Action configuration with credential reference
config := map[string]interface{}{
    "api_key": "{{credentials.slack_token}}",
    "channel": "#general",
}

// Inject credentials
injector := credential.NewInjector(repo, encryptionService)
result, err := injector.InjectCredentials(ctx, configJSON, &credential.InjectionContext{
    TenantID:    "tenant-123",
    WorkflowID:  "wf-456",
    ExecutionID: "exec-789",
    AccessedBy:  "user-001",
})

// Result contains injected config
// config["api_key"] = "xoxb-actual-token-value"
```

### Template Injection

#### Supported Template Syntax

```json
{
  "action": "slack:send_message",
  "config": {
    "token": "{{credentials.slack_token}}",
    "channel": "{{steps.get_channel.output.channel_id}}",
    "text": "Hello {{trigger.body.user_name}}!"
  }
}
```

Credentials use special `credentials.` prefix:
- `{{credentials.name}}` - Injects credential value
- `{{steps.step_id.output.field}}` - Access step output
- `{{trigger.body.field}}` - Access trigger data

### Encryption Best Practices

#### Storing Credentials

```go
// Create credential with encryption
credData := &credential.CredentialData{
    Type: "api_key",
    Value: map[string]interface{}{
        "api_key": "secret-key-value",
    },
}

// Encrypt using envelope encryption
encrypted, err := encryptionService.Encrypt(ctx, tenantID, credData)

// Store in database
cred := &credential.Credential{
    ID:           uuid.New().String(),
    TenantID:     tenantID,
    Name:         "my_api_key",
    Type:         "api_key",
    Ciphertext:   encrypted.Ciphertext,
    EncryptedDEK: encrypted.EncryptedDEK,
}
```

#### Masking Sensitive Output

```go
// Mask credential values in output
credentialValues := []string{"xoxb-secret-token", "password123"}
maskedOutput := injector.MaskOutput(actionOutput, credentialValues)

// All occurrences of credential values are replaced with "[REDACTED]"
```

### Testing with Mock Credentials

```go
func TestActionWithMockCredentials(t *testing.T) {
    mockCredService := &MockCredentialService{}
    mockCredService.On("GetValue", mock.Anything, "tenant-123", "cred-456", "system").
        Return(&credential.DecryptedValue{
            Value: map[string]interface{}{
                "api_key": "test-key-123",
            },
        }, nil)

    action := NewMyAction(mockCredService)
    // ... test action
}
```

---

## Expression Evaluation

### Using CEL Expressions in Actions

Gorax uses Google's Common Expression Language (CEL) for dynamic expressions.

#### Evaluating Conditions

```go
// internal/executor/expression/evaluator.go
evaluator := expression.NewEvaluator()

// Context with workflow data
context := map[string]interface{}{
    "steps": map[string]interface{}{
        "fetch_user": map[string]interface{}{
            "output": map[string]interface{}{
                "age": 25,
                "status": "active",
            },
        },
    },
}

// Evaluate boolean condition
result, err := evaluator.EvaluateCondition("steps.fetch_user.output.age >= 18", context)
// result = true

// Evaluate any expression
value, err := evaluator.Evaluate("steps.fetch_user.output.status", context)
// value = "active"
```

### Accessing Execution Context

Available context variables:

```go
{
    "trigger": {
        "type": "webhook",
        "body": {...},
        "headers": {...}
    },
    "steps": {
        "step_id": {
            "output": {...},
            "status": "success",
            "error": null
        }
    },
    "env": {
        "tenant_id": "...",
        "execution_id": "...",
        "workflow_id": "..."
    }
}
```

### Template Variable Syntax

```javascript
// Dot notation
{{ steps.fetch_user.output.name }}

// Array access
{{ steps.get_items.output.items[0].id }}

// Conditional
{{ steps.check_status.output.success ? "OK" : "FAILED" }}
```

### Custom Functions

Add custom CEL functions:

```go
// Register custom function
import "github.com/expr-lang/expr"

env := map[string]interface{}{
    "uppercase": strings.ToUpper,
    "contains": func(haystack, needle string) bool {
        return strings.Contains(haystack, needle)
    },
}

program, err := expr.Compile(expression, expr.Env(env))
```

---

## Real-World Examples

### Example 1: Building a Slack Integration (Step-by-Step)

This example shows the complete implementation of a Slack "SendMessage" action.

#### Step 1: Define Models

```go
// internal/integrations/slack/models.go
package slack

type SendMessageConfig struct {
    Channel        string                   `json:"channel"`
    Text           string                   `json:"text,omitempty"`
    Blocks         []map[string]interface{} `json:"blocks,omitempty"`
    ThreadTS       string                   `json:"thread_ts,omitempty"`
    IconEmoji      string                   `json:"icon_emoji,omitempty"`
}

func (c *SendMessageConfig) Validate() error {
    if c.Channel == "" {
        return ErrChannelRequired
    }
    if c.Text == "" && len(c.Blocks) == 0 {
        return ErrTextOrBlocksRequired
    }
    return nil
}
```

#### Step 2: Implement HTTP Client

```go
// internal/integrations/slack/client.go
package slack

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

const DefaultBaseURL = "https://slack.com/api"

type Client struct {
    accessToken string
    baseURL     string
    httpClient  *http.Client
}

func NewClient(accessToken string) (*Client, error) {
    if accessToken == "" {
        return nil, ErrInvalidToken
    }

    return &Client{
        accessToken: accessToken,
        baseURL:     DefaultBaseURL,
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
    }, nil
}

func (c *Client) SendMessage(ctx context.Context, req *SendMessageRequest) (*MessageResponse, error) {
    var resp MessageResponse
    if err := c.doRequest(ctx, "POST", "/chat.postMessage", req, &resp); err != nil {
        return nil, err
    }

    if !resp.OK {
        return nil, ParseSlackError(resp.Error)
    }

    return &resp, nil
}

func (c *Client) doRequest(ctx context.Context, method, endpoint string, body interface{}, result interface{}) error {
    // Marshal body
    bodyJSON, err := json.Marshal(body)
    if err != nil {
        return fmt.Errorf("failed to marshal body: %w", err)
    }

    // Create request
    req, err := http.NewRequestWithContext(ctx, method, c.baseURL+endpoint, bytes.NewReader(bodyJSON))
    if err != nil {
        return fmt.Errorf("failed to create request: %w", err)
    }

    // Set headers
    req.Header.Set("Authorization", "Bearer "+c.accessToken)
    req.Header.Set("Content-Type", "application/json")

    // Execute
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()

    // Decode response
    if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
        return fmt.Errorf("failed to decode response: %w", err)
    }

    return nil
}
```

#### Step 3: Implement Action

```go
// internal/integrations/slack/send_message.go
package slack

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/gorax/gorax/internal/credential"
    "github.com/gorax/gorax/internal/executor/actions"
)

type SendMessageAction struct {
    credentialService credential.Service
}

func NewSendMessageAction(credentialService credential.Service) *SendMessageAction {
    return &SendMessageAction{
        credentialService: credentialService,
    }
}

func (a *SendMessageAction) Execute(ctx context.Context, input *actions.ActionInput) (*actions.ActionOutput, error) {
    // Parse config
    configBytes, err := json.Marshal(input.Config)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal config: %w", err)
    }

    var config SendMessageConfig
    if err := json.Unmarshal(configBytes, &config); err != nil {
        return nil, fmt.Errorf("invalid config: %w", err)
    }

    // Validate
    if err := config.Validate(); err != nil {
        return nil, err
    }

    // Get credentials
    tenantID := input.Context["env"].(map[string]interface{})["tenant_id"].(string)
    credentialID := input.Context["credential_id"].(string)

    decryptedCred, err := a.credentialService.GetValue(ctx, tenantID, credentialID, "system")
    if err != nil {
        return nil, fmt.Errorf("failed to retrieve credential: %w", err)
    }

    accessToken := decryptedCred.Value["access_token"].(string)

    // Create client and send message
    client, err := NewClient(accessToken)
    if err != nil {
        return nil, err
    }

    resp, err := client.SendMessage(ctx, &SendMessageRequest{
        Channel:   config.Channel,
        Text:      config.Text,
        Blocks:    config.Blocks,
        ThreadTS:  config.ThreadTS,
        IconEmoji: config.IconEmoji,
    })
    if err != nil {
        return nil, err
    }

    // Return output
    output := actions.NewActionOutput(map[string]interface{}{
        "ok":        resp.OK,
        "channel":   resp.Channel,
        "timestamp": resp.TS,
    })
    output.WithMetadata("channel", resp.Channel)

    return output, nil
}
```

### Example 2: Building a Custom HTTP Action

```go
// internal/integrations/custom/http_get.go
package custom

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"

    "github.com/gorax/gorax/internal/executor/actions"
)

type HTTPGetAction struct {
    httpClient *http.Client
}

func NewHTTPGetAction() *HTTPGetAction {
    return &HTTPGetAction{
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}

type HTTPGetConfig struct {
    URL     string            `json:"url"`
    Headers map[string]string `json:"headers,omitempty"`
}

func (a *HTTPGetAction) Execute(ctx context.Context, input *actions.ActionInput) (*actions.ActionOutput, error) {
    // Parse config
    configBytes, _ := json.Marshal(input.Config)
    var config HTTPGetConfig
    json.Unmarshal(configBytes, &config)

    // Create request
    req, err := http.NewRequestWithContext(ctx, "GET", config.URL, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }

    // Set headers
    for key, value := range config.Headers {
        req.Header.Set(key, value)
    }

    // Execute
    resp, err := a.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()

    // Read body
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read body: %w", err)
    }

    // Parse JSON if possible
    var jsonBody interface{}
    if err := json.Unmarshal(body, &jsonBody); err != nil {
        jsonBody = string(body)
    }

    return actions.NewActionOutput(map[string]interface{}{
        "status_code": resp.StatusCode,
        "body":        jsonBody,
        "headers":     resp.Header,
    }), nil
}
```

### Example 3: Adding a New LLM Provider (Gemini)

```go
// internal/llm/providers/gemini/provider.go
package gemini

import (
    "context"
    "fmt"

    "github.com/gorax/gorax/internal/llm"
)

type GeminiProvider struct {
    apiKey  string
    baseURL string
}

func NewGeminiProvider(config *llm.ProviderConfig) (llm.Provider, error) {
    if config.APIKey == "" {
        return nil, fmt.Errorf("API key is required")
    }

    return &GeminiProvider{
        apiKey:  config.APIKey,
        baseURL: "https://generativelanguage.googleapis.com/v1",
    }, nil
}

func (p *GeminiProvider) Name() string {
    return "gemini"
}

func (p *GeminiProvider) ChatCompletion(ctx context.Context, req *llm.ChatRequest) (*llm.ChatResponse, error) {
    // Build Gemini request
    geminiReq := map[string]interface{}{
        "contents": p.convertMessages(req.Messages),
        "generationConfig": map[string]interface{}{
            "temperature":    req.Temperature,
            "maxOutputTokens": req.MaxTokens,
        },
    }

    // Make API call
    // ... implementation details

    return &llm.ChatResponse{
        ID:    "gemini-response-id",
        Model: req.Model,
        Message: llm.ChatMessage{
            Role:    "assistant",
            Content: "response content",
        },
    }, nil
}

func (p *GeminiProvider) convertMessages(messages []llm.ChatMessage) []map[string]interface{} {
    result := make([]map[string]interface{}, len(messages))
    for i, msg := range messages {
        result[i] = map[string]interface{}{
            "role": map[string]string{"user": "user", "assistant": "model"}[msg.Role],
            "parts": []map[string]string{
                {"text": msg.Content},
            },
        }
    }
    return result
}

// Implement other Provider methods...
```

---

## Testing Integrations

### Unit Testing with Mocks

#### Mock Credential Service

```go
// internal/integrations/slack/send_message_test.go
package slack

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"

    "github.com/gorax/gorax/internal/credential"
    "github.com/gorax/gorax/internal/executor/actions"
)

type MockCredentialService struct {
    mock.Mock
}

func (m *MockCredentialService) GetValue(ctx context.Context, tenantID, credentialID, userID string) (*credential.DecryptedValue, error) {
    args := m.Called(ctx, tenantID, credentialID, userID)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*credential.DecryptedValue), args.Error(1)
}

func TestSendMessageAction_Success(t *testing.T) {
    // Setup
    mockCredService := new(MockCredentialService)
    action := NewSendMessageAction(mockCredService)

    mockCredService.On("GetValue", mock.Anything, "tenant-123", "cred-456", "system").
        Return(&credential.DecryptedValue{
            Value: map[string]interface{}{
                "access_token": "xoxb-test-token",
            },
        }, nil)

    // Prepare input
    config := SendMessageConfig{
        Channel: "#general",
        Text:    "Hello, world!",
    }

    input := actions.NewActionInput(config, map[string]interface{}{
        "env": map[string]interface{}{
            "tenant_id": "tenant-123",
        },
        "credential_id": "cred-456",
    })

    // Execute
    output, err := action.Execute(context.Background(), input)

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, output)
    assert.True(t, output.Data.(map[string]interface{})["ok"].(bool))
    mockCredService.AssertExpectations(t)
}
```

### Integration Testing with Real Services

#### Using Test Containers

```go
// internal/integrations/database/postgres_integration_test.go
//go:build integration

package database

import (
    "context"
    "testing"

    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/wait"
)

func TestPostgresAction_Integration(t *testing.T) {
    ctx := context.Background()

    // Start PostgreSQL container
    req := testcontainers.ContainerRequest{
        Image:        "postgres:15",
        ExposedPorts: []string{"5432/tcp"},
        Env: map[string]string{
            "POSTGRES_PASSWORD": "test",
            "POSTGRES_DB":       "testdb",
        },
        WaitingFor: wait.ForLog("database system is ready to accept connections"),
    }

    container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: req,
        Started:          true,
    })
    if err != nil {
        t.Fatal(err)
    }
    defer container.Terminate(ctx)

    // Get connection details
    host, _ := container.Host(ctx)
    port, _ := container.MappedPort(ctx, "5432")

    // Test action
    action := NewPostgresQueryAction(host, port.Port(), "testdb", "postgres", "test")
    // ... execute test queries
}
```

### Test Fixtures and Helpers

```go
// internal/integrations/slack/testdata/fixtures.go
package testdata

func MockSlackMessageResponse() *slack.MessageResponse {
    return &slack.MessageResponse{
        OK:      true,
        Channel: "C123456",
        TS:      "1234567890.123456",
        Message: slack.Message{
            Type: "message",
            Text: "Hello, world!",
        },
    }
}

func MockSlackCredential() map[string]interface{} {
    return map[string]interface{}{
        "access_token": "xoxb-test-token",
        "team_id":      "T123456",
        "bot_user_id":  "U123456",
    }
}
```

### CI/CD Considerations

```yaml
# .github/workflows/integration-tests.yml
name: Integration Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run integration tests
        run: go test -tags=integration ./...
        env:
          DATABASE_URL: postgres://postgres:test@localhost:5432/testdb
```

---

## Publishing to Marketplace

### Creating Workflow Templates

#### Template Structure

```json
{
  "name": "Slack Notification on GitHub Issue",
  "description": "Send a Slack notification when a GitHub issue is created",
  "category": "notification",
  "tags": ["slack", "github", "notifications"],
  "version": "1.0.0",
  "author": "Your Name",
  "definition": {
    "trigger": {
      "type": "webhook",
      "config": {
        "path": "/github/issues",
        "method": "POST"
      }
    },
    "steps": [
      {
        "id": "parse_issue",
        "type": "action:transform",
        "config": {
          "operations": [
            {
              "type": "extract",
              "path": "$.issue",
              "output": "issue_data"
            }
          ]
        }
      },
      {
        "id": "send_slack",
        "type": "slack:send_message",
        "config": {
          "channel": "#notifications",
          "text": "New issue created: {{steps.parse_issue.output.issue_data.title}}"
        },
        "credentials": {
          "slack": "{{credentials.slack_token}}"
        }
      }
    ]
  }
}
```

### Template Metadata

```go
// internal/marketplace/model.go
type MarketplaceTemplate struct {
    ID            string          `json:"id"`
    Name          string          `json:"name"`
    Description   string          `json:"description"`
    Category      string          `json:"category"`
    Tags          []string        `json:"tags"`
    Version       string          `json:"version"`
    AuthorID      string          `json:"author_id"`
    AuthorName    string          `json:"author_name"`
    Definition    json.RawMessage `json:"definition"`
    DownloadCount int             `json:"download_count"`
    AverageRating float64         `json:"average_rating"`
    IsVerified    bool            `json:"is_verified"`
}
```

### Documentation Requirements

Every marketplace template should include:

1. **README.md**: Overview and usage instructions
2. **SETUP.md**: Required credentials and configuration
3. **EXAMPLE.json**: Sample input/output
4. **CHANGELOG.md**: Version history

### Submission Process

1. **Create Template**: Define workflow JSON
2. **Test Thoroughly**: Ensure all actions work
3. **Document**: Add comprehensive documentation
4. **Submit**: POST to `/api/v1/marketplace/templates`
5. **Review**: Wait for admin verification
6. **Publish**: Template goes live

```bash
# Submit template via API
curl -X POST https://gorax.example.com/api/v1/marketplace/templates \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d @template.json
```

---

## Best Practices Summary

### Action Development
- ✅ Implement proper error handling with wrapped errors
- ✅ Validate all configuration inputs
- ✅ Use context for cancellation and timeouts
- ✅ Return structured output with metadata
- ✅ Write comprehensive unit tests

### Security
- ✅ Never log sensitive data (credentials, tokens)
- ✅ Use credential injection for secrets
- ✅ Validate and sanitize all inputs
- ✅ Implement rate limiting for external APIs
- ✅ Mask credentials in output

### Performance
- ✅ Use connection pooling for HTTP clients
- ✅ Implement exponential backoff for retries
- ✅ Set appropriate timeouts
- ✅ Cache expensive operations
- ✅ Limit payload sizes

### Testing
- ✅ Write unit tests with mocks
- ✅ Add integration tests for critical paths
- ✅ Test error scenarios
- ✅ Verify credential masking
- ✅ Check for goroutine leaks

### Documentation
- ✅ Document all configuration options
- ✅ Provide usage examples
- ✅ Explain credential requirements
- ✅ Include troubleshooting guide
- ✅ Maintain changelog

---

## Additional Resources

- **Gorax Architecture Guide**: `/docs/DEVELOPER_GUIDE.md`
- **Security Guidelines**: `/docs/WEBSOCKET_SECURITY.md`
- **Collaboration Features**: `/docs/COLLABORATION.md`
- **API Reference**: `/docs/API.md`
- **Credential Management**: `internal/credential/`
- **Action Examples**: `internal/integrations/`

## Getting Help

- **GitHub Issues**: Report bugs or request features
- **Discussions**: Ask questions and share ideas
- **Documentation**: Browse the docs folder
- **Source Code**: Review existing integrations

---

## Appendix: Quick Reference

### Action Interface
```go
type Action interface {
    Execute(ctx context.Context, input *ActionInput) (*ActionOutput, error)
}
```

### LLM Provider Interface
```go
type Provider interface {
    ChatCompletion(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
    GenerateEmbeddings(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error)
    Name() string
    HealthCheck(ctx context.Context) error
}
```

### Credential Injection
```go
injector.InjectCredentials(ctx, config, &InjectionContext{...})
```

### Expression Evaluation
```go
evaluator.EvaluateCondition("steps.check.output.status == 'success'", context)
```

### Retry Pattern
```go
WithRetry(ctx, RetryConfig{MaxAttempts: 3}, func() error {
    return doRequest()
})
```

---

**Version**: 1.0.0
**Last Updated**: 2026-01-01
**Maintainer**: Gorax Development Team
