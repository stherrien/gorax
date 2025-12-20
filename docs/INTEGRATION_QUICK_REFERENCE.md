# Integration Quick Reference Guide

## Available Integrations

### Slack Integration
**Actions:** 4
- `slack:send_message` - Send message to channel
- `slack:send_dm` - Send direct message to user
- `slack:add_reaction` - Add emoji reaction to message
- `slack:update_message` - Update existing message

**Authentication:** OAuth Bearer token

**Example Config (send_message):**
```json
{
  "channel": "#general",
  "text": "Hello from Gorax!",
  "username": "Gorax Bot",
  "icon_emoji": ":robot_face:"
}
```

### Jira Integration
**Actions:** 5
- `jira:create_issue` - Create new issue
- `jira:update_issue` - Update existing issue
- `jira:add_comment` - Add comment to issue
- `jira:transition_issue` - Change issue status
- `jira:search_issues` - Search using JQL

**Authentication:** Basic Auth (email + API token)

**Example Config (create_issue):**
```json
{
  "project": "PROJ",
  "issue_type": "Bug",
  "summary": "Issue from workflow",
  "description": "Automated issue creation",
  "priority": "High",
  "labels": ["automated", "gorax"]
}
```

### GitHub Integration
**Actions:** 3
- `github:create_issue` - Create new issue
- `github:create_pr_comment` - Comment on pull request
- `github:add_label` - Add labels to issue/PR

**Authentication:** Personal Access Token (PAT)

**Example Config (create_issue):**
```json
{
  "owner": "myorg",
  "repo": "myrepo",
  "title": "Automated issue",
  "body": "Created by Gorax workflow",
  "labels": ["bug", "automated"]
}
```

### PagerDuty Integration
**Actions:** 4
- `pagerduty:create_incident` - Create new incident
- `pagerduty:acknowledge_incident` - Acknowledge incident
- `pagerduty:resolve_incident` - Resolve incident
- `pagerduty:add_note` - Add note to incident

**Authentication:** API key + From email

**Example Config (create_incident):**
```json
{
  "title": "Critical alert from Gorax",
  "service": "PXXXXXX",
  "urgency": "high",
  "body": "Automated incident from workflow"
}
```

## Integration Action Interface

All integration actions implement the following interface:

```go
type Action interface {
    Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error)
    Validate(config map[string]interface{}) error
    Name() string
    Description() string
}
```

## Using Integrations in Workflows

### 1. Register Actions (Startup)
```go
import (
    "github.com/gorax/gorax/internal/integrations"
    "github.com/gorax/gorax/internal/integrations/slack"
    "github.com/gorax/gorax/internal/integrations/jira"
)

// During application startup
func registerIntegrationActions(registry *integrations.Registry) {
    // Slack actions
    registry.Register(slack.NewSendMessageAction(credService))
    registry.Register(slack.NewSendDMAction(credService))

    // Jira actions
    registry.Register(jira.NewCreateIssueAction(baseURL, email, token))
    registry.Register(jira.NewUpdateIssueAction(baseURL, email, token))

    // ... register other actions
}
```

### 2. Execute Action from Workflow
```go
// Get action from registry
action, err := integrations.GlobalRegistry.Get("slack:send_message")
if err != nil {
    return err
}

// Prepare configuration
config := map[string]interface{}{
    "channel": "#alerts",
    "text": "Workflow completed successfully",
}

// Prepare input from previous step
input := map[string]interface{}{
    "execution_id": executionID,
    "status": "completed",
}

// Execute action
result, err := action.Execute(ctx, config, input)
if err != nil {
    return err
}

// Use result in next step
nextStepInput := result
```

### 3. Validate Configuration
```go
config := map[string]interface{}{
    "channel": "#general",
    "text": "Hello!",
}

if err := action.Validate(config); err != nil {
    // Handle validation error
    return fmt.Errorf("invalid config: %w", err)
}
```

## Error Handling

### Common Error Types
```go
integrations.ErrInvalidConfig     // Configuration is invalid
integrations.ErrAuthFailed        // Authentication failed
integrations.ErrRateLimitExceeded // Rate limit exceeded (retryable)
integrations.ErrNotFound          // Resource not found
integrations.ErrPermissionDenied  // Permission denied
```

### Retry Behavior
- Automatic retry for rate limit errors
- Exponential backoff: 1s, 2s, 4s, 8s...
- Max delay: 30s
- Max attempts: 3 (configurable)

## Frontend Usage

### Adding Integration Node to Canvas
```typescript
import { SlackNode, JiraNode, GitHubNode, PagerDutyNode } from '@/components/nodes';

// Register node types with React Flow
const nodeTypes = {
  slack: SlackNode,
  jira: JiraNode,
  github: GitHubNode,
  pagerduty: PagerDutyNode,
};

// Create node
const newNode = {
  id: generateId(),
  type: 'slack',
  position: { x: 100, y: 100 },
  data: {
    action: 'send_message',
    config: {
      channel: '#general',
      text: 'Hello!',
    },
  },
};
```

### Action Selection UI
```typescript
import { integrationActions } from '@/types/integrations';

// Get all Slack actions
const slackActions = integrationActions.filter(
  action => action.integration === 'slack'
);

// Render action selector
<select>
  {slackActions.map(action => (
    <option key={action.id} value={action.id}>
      {action.name}
    </option>
  ))}
</select>
```

## Best Practices

### Configuration
1. **Store sensitive data in credentials vault**, not in workflow config
2. **Validate configuration before saving** workflow
3. **Use credential references** instead of hardcoded tokens
4. **Test with real credentials** before deploying

### Error Handling
1. **Always wrap errors** with context
2. **Log integration failures** with execution ID
3. **Implement circuit breakers** for repeated failures
4. **Provide clear error messages** to users

### Performance
1. **Use connection pooling** for HTTP clients
2. **Cache credential lookups** when possible
3. **Implement request batching** where supported
4. **Monitor rate limits** proactively

### Security
1. **Never log credentials** or tokens
2. **Rotate credentials regularly**
3. **Use principle of least privilege** for API tokens
4. **Validate all input** before sending to external APIs

## Testing

### Unit Tests
```go
func TestSlackSendMessage(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Mock Slack API response
        json.NewEncoder(w).Encode(map[string]interface{}{
            "ok": true,
            "ts": "1234567890.123456",
        })
    }))
    defer server.Close()

    action := slack.NewSendMessageAction(credService)
    action.client.baseURL = server.URL

    result, err := action.Execute(ctx, config, input)

    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

### Integration Tests
```go
// Test with real API (optional, requires credentials)
func TestSlackIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    token := os.Getenv("SLACK_TOKEN")
    if token == "" {
        t.Skip("SLACK_TOKEN not set")
    }

    client, err := slack.NewClient(token)
    require.NoError(t, err)

    err = client.Authenticate(context.Background())
    assert.NoError(t, err)
}
```

## Troubleshooting

### Slack Issues
- **Error: "channel_not_found"** - Check channel name includes # or verify bot is invited
- **Error: "invalid_auth"** - Verify OAuth token is valid and has required scopes
- **Rate limiting** - Implement exponential backoff, current limit ~1 msg/sec

### Jira Issues
- **Error: "Unauthorized"** - Check email/API token combination
- **Error: "Field 'x' cannot be set"** - Verify field is available for issue type
- **Transition not found** - Check transition name exactly matches Jira workflow

### GitHub Issues
- **Error: "Bad credentials"** - Verify PAT is valid and not expired
- **Error: "Resource not found"** - Check owner/repo names are correct
- **Rate limiting** - GitHub allows 5000 requests/hour for authenticated users

### PagerDuty Issues
- **Error: "Unauthorized"** - Verify API key and From email
- **Error: "Service not found"** - Check service ID format (PXXXXXX)
- **Incident not created** - Check service has integration key configured

## Additional Resources

- [Slack API Documentation](https://api.slack.com/)
- [Jira REST API Documentation](https://developer.atlassian.com/cloud/jira/platform/rest/v3/)
- [GitHub REST API Documentation](https://docs.github.com/en/rest)
- [PagerDuty API Documentation](https://developer.pagerduty.com/docs/rest-api-v2/)

## Support

For issues or questions about integrations:
1. Check the integration test files for examples
2. Review the implementation summary document
3. Check API documentation for the specific service
4. File an issue in the repository with logs and configuration (redact sensitive data)
