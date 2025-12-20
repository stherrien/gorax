# Sentry Error Tracking Integration

This document describes the Sentry error tracking integration in the Gorax project.

## Overview

Sentry error tracking has been integrated into both the backend (Go) and frontend (React) to provide comprehensive error monitoring, alerting, and debugging capabilities.

## Backend (Go)

### Components

#### 1. Error Tracking Package (`internal/errortracking/`)

The core error tracking functionality is implemented in `internal/errortracking/sentry.go`:

**Key Features:**
- Sentry SDK initialization with configuration
- Error capture with context enrichment (tenant ID, user ID, execution ID)
- Message capture with severity levels
- Breadcrumb tracking for audit trails
- User context tracking
- Panic recovery
- Scoped error tracking

**API:**
```go
// Initialize Sentry
tracker, err := errortracking.Initialize(config.ObservabilityConfig{
    SentryEnabled:     true,
    SentryDSN:         "https://...",
    SentryEnvironment: "production",
    SentrySampleRate:  1.0,
})

// Capture an error
eventID := tracker.CaptureError(ctx, err)

// Capture with custom tags
eventID := tracker.CaptureErrorWithTags(ctx, err, map[string]string{
    "tenant_id": "tenant-123",
    "workflow_id": "workflow-456",
})

// Add breadcrumb
tracker.AddBreadcrumb(ctx, errortracking.Breadcrumb{
    Type:     "http",
    Category: "request",
    Message:  "API request started",
    Level:    errortracking.LevelInfo,
    Data:     map[string]interface{}{"method": "POST"},
})

// Set user context
tracker.SetUser(ctx, errortracking.User{
    ID:    "user-123",
    Email: "user@example.com",
})
```

#### 2. Sentry Middleware (`internal/api/middleware/sentry.go`)

HTTP middleware that automatically:
- Captures and recovers from panics
- Enriches errors with request context
- Adds HTTP request breadcrumbs
- Sets user/tenant information from context

**Usage:**
```go
// Middleware is automatically added in app.go
r.Use(apiMiddleware.SentryMiddleware(errorTracker))
```

### Configuration

Backend configuration is set via environment variables in `.env`:

```bash
# Sentry Error Tracking
SENTRY_ENABLED=true
SENTRY_DSN=https://examplePublicKey@o0.ingest.sentry.io/0
SENTRY_ENVIRONMENT=production
SENTRY_SAMPLE_RATE=1.0
```

**Configuration Options:**

| Variable | Description | Default |
|----------|-------------|---------|
| `SENTRY_ENABLED` | Enable/disable Sentry tracking | `false` |
| `SENTRY_DSN` | Sentry Data Source Name (get from Sentry project) | `""` |
| `SENTRY_ENVIRONMENT` | Environment name (development, staging, production) | `"development"` |
| `SENTRY_SAMPLE_RATE` | Error sampling rate (0.0 to 1.0) | `1.0` |

### Error Context Enrichment

The middleware automatically enriches errors with:

**From Context:**
- `tenant_id` - Current tenant ID
- `user_id` - Authenticated user ID
- `execution_id` - Workflow execution ID
- `workflow_id` - Workflow ID
- `request_id` - HTTP request ID

**From Request:**
- HTTP method, path, URL
- User agent
- Remote address
- Selected headers (X-Request-ID, X-Tenant-ID, etc.)

### Integration Points

Sentry is integrated at several key points:

1. **HTTP Middleware** - Captures all panics and enriches request context
2. **Workflow Executor** - Can manually capture execution errors
3. **Queue Processing** - Can capture async processing errors
4. **External Integration Failures** - API calls, webhooks, etc.

Example of manual error capture in handlers:

```go
func (h *Handler) SomeMethod(w http.ResponseWriter, r *http.Request) {
    err := h.service.DoSomething(r.Context())
    if err != nil {
        // Log error
        h.logger.Error("operation failed", "error", err)

        // Capture in Sentry with additional context
        if h.errorTracker != nil {
            h.errorTracker.CaptureErrorWithTags(r.Context(), err, map[string]string{
                "operation": "some_operation",
                "component": "handler",
            })
        }

        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }
}
```

### Testing

Tests use a mock tracker to verify error tracking behavior without actually sending to Sentry:

```go
func TestHandler(t *testing.T) {
    mockTracker := &mockTracker{
        errors: []error{},
    }

    // ... test code ...

    assert.Len(t, mockTracker.errors, 1)
}
```

## Frontend (React/TypeScript)

### Components

#### 1. Sentry Library (`web/src/lib/sentry.ts`)

Core Sentry functionality for the frontend:

**Key Features:**
- Sentry SDK initialization
- Error and message capture
- User context management
- Breadcrumb tracking
- Session replay
- Performance monitoring
- Browser extension filtering

**API:**
```typescript
import {
  initializeSentry,
  setUser,
  captureError,
  addBreadcrumb,
  withExecutionScope
} from './lib/sentry';

// Initialize (done automatically in main.tsx)
initializeSentry({
  dsn: 'https://...',
  environment: 'production',
  enabled: true,
  sampleRate: 1.0,
  tracesSampleRate: 1.0,
});

// Set user on login
setUser({
  id: 'user-123',
  email: 'user@example.com',
  username: 'username',
});

// Manually capture error
try {
  // ... code ...
} catch (error) {
  captureError(error, { component: 'WorkflowEditor' });
}

// Add breadcrumb
addBreadcrumb({
  message: 'User clicked save button',
  category: 'ui',
  level: 'info',
  data: { workflowId: 'workflow-123' },
});

// Scoped tracking
withExecutionScope(executionId, workflowId, () => {
  // Errors in this scope will be tagged with execution/workflow IDs
});
```

#### 2. ErrorBoundary Component (`web/src/components/ErrorBoundary.tsx`)

React Error Boundary that catches component errors and reports them to Sentry:

**Features:**
- Catches React component errors
- Reports to Sentry with component stack
- Displays user-friendly error UI
- Provides retry and navigation options
- Shows error details in development mode

**Usage:**
```tsx
// Wrap your app (done automatically in main.tsx)
<ErrorBoundary>
  <App />
</ErrorBoundary>

// Or wrap specific components
<ErrorBoundary fallback={<CustomErrorUI />}>
  <SomeComponent />
</ErrorBoundary>
```

### Configuration

Frontend configuration uses Vite environment variables in `.env`:

```bash
# Environment
VITE_APP_ENV=production

# Sentry Error Tracking
VITE_SENTRY_ENABLED=true
VITE_SENTRY_DSN=https://examplePublicKey@o0.ingest.sentry.io/0
VITE_SENTRY_SAMPLE_RATE=1.0
VITE_SENTRY_TRACES_SAMPLE_RATE=1.0
```

**Configuration Options:**

| Variable | Description | Default |
|----------|-------------|---------|
| `VITE_SENTRY_ENABLED` | Enable/disable Sentry tracking | `false` |
| `VITE_SENTRY_DSN` | Sentry Data Source Name | `""` |
| `VITE_APP_ENV` | Environment name | `"development"` |
| `VITE_SENTRY_SAMPLE_RATE` | Error sampling rate | `1.0` |
| `VITE_SENTRY_TRACES_SAMPLE_RATE` | Performance trace sampling | `1.0` |

### Source Maps

Source maps are enabled in production builds (`vite.config.ts`) to provide readable stack traces:

```typescript
export default defineConfig({
  build: {
    sourcemap: true,
  },
});
```

### Features

1. **Session Replay** - Records user sessions when errors occur (10% sampling, 100% on errors)
2. **Performance Monitoring** - Tracks page load times and API calls
3. **Browser Extension Filtering** - Automatically filters out errors from browser extensions
4. **User Context** - Automatically captures user information on login
5. **Breadcrumbs** - Tracks user actions before errors occur

## Setup Guide

### 1. Create Sentry Project

1. Go to [sentry.io](https://sentry.io) and create an account
2. Create a new project:
   - For backend: Choose "Go" platform
   - For frontend: Choose "React" platform
3. Copy the DSN from project settings

### 2. Configure Backend

Update `.env` file:

```bash
SENTRY_ENABLED=true
SENTRY_DSN=<your-go-project-dsn>
SENTRY_ENVIRONMENT=production
SENTRY_SAMPLE_RATE=1.0
```

### 3. Configure Frontend

Update `web/.env` file:

```bash
VITE_SENTRY_ENABLED=true
VITE_SENTRY_DSN=<your-react-project-dsn>
VITE_APP_ENV=production
VITE_SENTRY_SAMPLE_RATE=1.0
VITE_SENTRY_TRACES_SAMPLE_RATE=1.0
```

### 4. Upload Source Maps (Optional)

For better stack traces in production, upload source maps to Sentry:

```bash
# Install Sentry CLI
npm install -g @sentry/cli

# Configure Sentry CLI
export SENTRY_AUTH_TOKEN=<your-auth-token>
export SENTRY_ORG=<your-org>
export SENTRY_PROJECT=<your-project>

# Upload source maps after build
npm run build
sentry-cli sourcemaps upload --validate dist/
```

## Best Practices

### Backend

1. **Context Enrichment** - Always pass context to error tracking functions
2. **Meaningful Tags** - Add custom tags for filtering (component, operation, etc.)
3. **Breadcrumbs** - Add breadcrumbs before critical operations
4. **Error Wrapping** - Wrap errors with context: `fmt.Errorf("operation failed: %w", err)`

### Frontend

1. **User Context** - Set user context immediately after login
2. **Clear on Logout** - Clear user context on logout
3. **Scoped Tracking** - Use scoped functions for workflow/execution tracking
4. **Breadcrumbs** - Add breadcrumbs for user actions
5. **Error Boundaries** - Wrap major components in error boundaries

## Monitoring and Alerts

### Key Metrics to Monitor

1. **Error Rate** - Track overall error rate and set alerts
2. **Failed Workflows** - Monitor workflow execution failures
3. **API Errors** - Track HTTP 5xx errors
4. **Integration Failures** - Monitor external API failures
5. **Panics** - Alert on any panic recovery

### Recommended Alerts

- Error rate exceeds threshold (e.g., >1% of requests)
- Critical workflow failures
- Multiple panics in short time
- External integration failures

## Troubleshooting

### Backend Issues

**Sentry not capturing errors:**
- Verify `SENTRY_ENABLED=true`
- Check DSN is correct
- Ensure error tracking is initialized before errors occur
- Check network connectivity to Sentry

**Missing context:**
- Ensure middleware is loaded before handlers
- Verify context values are set correctly
- Check tag names match expected keys

### Frontend Issues

**Errors not appearing:**
- Check `VITE_SENTRY_ENABLED=true` in `.env`
- Verify DSN is correct
- Check browser console for Sentry initialization errors
- Ensure ErrorBoundary wraps components

**Source maps not working:**
- Verify `sourcemap: true` in vite.config.ts
- Upload source maps to Sentry
- Check source map URL in production bundles

## Testing

### Backend Tests

```bash
# Run error tracking tests
go test ./internal/errortracking/...

# Run middleware tests
go test ./internal/api/middleware -run TestSentry
```

### Frontend Tests

```bash
# Run all tests (Sentry is mocked in tests)
npm run test
```

## Architecture Decisions

### Why Sentry?

1. **Comprehensive** - Covers both backend and frontend
2. **Context-Rich** - Excellent context enrichment capabilities
3. **Performance** - Includes performance monitoring
4. **Integrations** - Integrates with issue trackers, Slack, etc.
5. **Session Replay** - Helps reproduce frontend issues

### Design Choices

1. **Non-Blocking** - Sentry initialization errors don't crash the app
2. **Testable** - Interface-based design allows mocking
3. **Contextual** - Heavy use of context for enrichment
4. **Filtered** - Browser extensions and known issues filtered out
5. **Sampled** - Configurable sampling to control costs

## Cost Optimization

1. **Sampling** - Reduce sample rates in high-traffic environments
2. **Filtering** - Filter out non-actionable errors
3. **Quotas** - Set up Sentry quotas to control costs
4. **Environments** - Use different projects for dev/staging/prod
5. **Retention** - Adjust data retention policies

## Future Enhancements

- [ ] Add custom Sentry tags for workflow types
- [ ] Implement release tracking
- [ ] Add commit SHA to error reports
- [ ] Set up Sentry alerts integration with Slack
- [ ] Implement custom error grouping rules
- [ ] Add performance monitoring for critical paths
- [ ] Integrate with JIRA for automatic ticket creation
