# Distributed Tracing Guide

This guide explains how to use distributed tracing in the Gorax platform.

## Overview

Gorax uses OpenTelemetry for distributed tracing, allowing you to:
- Track requests across service boundaries
- Visualize workflow execution paths
- Identify performance bottlenecks
- Debug production issues
- Monitor SLAs and error rates

## Quick Start

### 1. Start Jaeger

```bash
# Start Jaeger using Docker Compose
docker-compose -f docker-compose.tracing.yml up -d

# Jaeger UI will be available at: http://localhost:16686
```

### 2. Enable Tracing

Update your `.env` file:

```bash
# Enable distributed tracing
TRACING_ENABLED=true
TRACING_ENDPOINT=localhost:4317
TRACING_SERVICE_NAME=gorax
TRACING_SAMPLE_RATE=1.0
```

### 3. Restart Services

```bash
# Restart API and worker
make restart-api
make restart-worker
```

### 4. Generate Traces

Execute a workflow and view traces in Jaeger UI.

## Understanding Traces

### Trace Structure

A trace represents a single request through the system:

```
Trace: Execute Workflow "notification-flow"
├─ Span: HTTP POST /api/v1/workflows/wf-123/execute
│  └─ Span: workflow.execute
│     ├─ Span: workflow.step.execute (http action)
│     │  └─ Span: http.request (outgoing)
│     ├─ Span: workflow.step.execute (transform)
│     └─ Span: workflow.step.execute (slack)
│        └─ Span: http.request (to Slack API)
```

### Span Attributes

Each span includes contextual information:

**HTTP Spans:**
- `http.method`: GET, POST, etc.
- `http.target`: /api/v1/workflows/...
- `http.status_code`: 200, 404, 500, etc.
- `http.user_agent`: Client identifier

**Workflow Spans:**
- `tenant_id`: Tenant identifier
- `workflow_id`: Workflow identifier
- `execution_id`: Execution identifier
- `node_id`: Step identifier
- `node_type`: Step type (http, slack, etc.)

**Queue Spans:**
- `queue.name`: Queue name
- `queue.message_id`: Message ID
- `retry_count`: Number of retries

## Configuration

### Sample Rate

Control what percentage of requests to trace:

```bash
# Trace all requests (100%)
TRACING_SAMPLE_RATE=1.0

# Trace 10% of requests
TRACING_SAMPLE_RATE=0.1

# Trace 1% of requests (production)
TRACING_SAMPLE_RATE=0.01

# Disable tracing
TRACING_ENABLED=false
```

### Backends

Gorax supports any OpenTelemetry-compatible backend:

#### Jaeger
```bash
TRACING_ENDPOINT=localhost:4317
```

#### Grafana Tempo
```bash
TRACING_ENDPOINT=tempo.example.com:443
```

#### AWS X-Ray
Use the OpenTelemetry Collector with X-Ray exporter.

#### Honeycomb
```bash
TRACING_ENDPOINT=api.honeycomb.io:443
```

## Advanced Usage

### Adding Custom Spans

```go
import "github.com/gorax/gorax/internal/tracing"

func processData(ctx context.Context, data []byte) error {
    // Create a custom span
    ctx, span := tracing.StartSpan(ctx, "data.process")
    defer span.End()

    // Add attributes
    tracing.SetSpanAttributes(span, map[string]interface{}{
        "data_size": len(data),
        "format": "json",
    })

    // Do work...
    result, err := parse(data)
    if err != nil {
        // Record error
        tracing.RecordError(span, err)
        return err
    }

    // Add more attributes
    tracing.SetSpanAttributes(span, map[string]interface{}{
        "records_processed": len(result),
    })

    return nil
}
```

### Recording Events

```go
// Record important events within a span
tracing.RecordWorkflowEvent(ctx, "validation.completed", map[string]interface{}{
    "errors_found": 3,
    "warnings_found": 7,
})
```

### Extracting Trace Context

```go
// Get trace ID for logging
traceID := tracing.GetTraceID(ctx)
logger.Info("processing request", "trace_id", traceID)

// Get span ID
spanID := tracing.GetSpanID(ctx)
```

### Propagating Context

```go
// When making HTTP requests, propagate trace context
req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
tracing.InjectHTTPTraceContext(req)

// The receiving service will see this as a child span
resp, err := client.Do(req)
```

## Production Best Practices

### 1. Use Appropriate Sampling

Don't trace every request in production:

```bash
# High traffic services (>1000 req/s)
TRACING_SAMPLE_RATE=0.01  # 1%

# Medium traffic services (100-1000 req/s)
TRACING_SAMPLE_RATE=0.1   # 10%

# Low traffic services (<100 req/s)
TRACING_SAMPLE_RATE=1.0   # 100%
```

### 2. Add Meaningful Attributes

```go
// ✅ Good: descriptive attributes
tracing.SetSpanAttributes(span, map[string]interface{}{
    "workflow_name": "customer-onboarding",
    "customer_tier": "enterprise",
    "data_source": "salesforce",
})

// ❌ Bad: vague attributes
tracing.SetSpanAttributes(span, map[string]interface{}{
    "id": "123",
    "type": "process",
})
```

### 3. Use Span Events for Key Milestones

```go
tracing.RecordWorkflowEvent(ctx, "validation.started", nil)
// ... validation logic
tracing.RecordWorkflowEvent(ctx, "validation.completed", map[string]interface{}{
    "duration_ms": elapsed.Milliseconds(),
})
```

### 4. Avoid Sensitive Data

```go
// ❌ Bad: includes PII
tracing.SetSpanAttributes(span, map[string]interface{}{
    "email": "user@example.com",
    "ssn": "123-45-6789",
})

// ✅ Good: anonymized identifiers
tracing.SetSpanAttributes(span, map[string]interface{}{
    "user_id": "usr_abc123",
    "customer_hash": "sha256:...",
})
```

### 5. Set Span Status Appropriately

```go
if err != nil {
    span.RecordError(err)
    span.SetStatus(codes.Error, err.Error())
    return err
}
span.SetStatus(codes.Ok, "operation completed")
```

## Querying Traces

### Jaeger UI

Access the Jaeger UI at `http://localhost:16686`.

**Find traces by:**
- Service name: `gorax`
- Operation: `workflow.execute`, `http.request`, etc.
- Tags: `tenant_id=tenant-123`
- Duration: `> 1s`
- Status: `error=true`

**Common queries:**
```
# All failed workflow executions
service=gorax operation=workflow.execute error=true

# Slow HTTP requests
service=gorax operation=http.request AND min_duration=1s

# Specific tenant
service=gorax tag=tenant_id:tenant-123

# Recent executions
service=gorax operation=workflow.execute lookback=1h
```

## Troubleshooting

### No traces appearing

**Check configuration:**
```bash
echo $TRACING_ENABLED
echo $TRACING_ENDPOINT
```

**Verify Jaeger is running:**
```bash
curl http://localhost:16686/api/traces?service=gorax
```

**Check application logs:**
```bash
# Should see: "distributed tracing enabled"
docker logs gorax-api | grep tracing
```

### Traces are incomplete

**Context not propagated:**
Ensure context is passed to all functions:

```go
// ❌ Bad
func doWork() error {
    // No context!
}

// ✅ Good
func doWork(ctx context.Context) error {
    // Context available for tracing
}
```

### High overhead

**Reduce sampling:**
```bash
TRACING_SAMPLE_RATE=0.1
```

**Batch more aggressively:**
Update `internal/tracing/tracer.go` to adjust batch settings.

## Integration Examples

### With Structured Logging

```go
ctx, span := tracing.StartSpan(ctx, "user.authenticate")
defer span.End()

traceID := tracing.GetTraceID(ctx)
logger := logger.With("trace_id", traceID)

logger.Info("authentication started")
// Logs will include trace_id for correlation
```

### With Metrics

```go
ctx, span := tracing.StartSpan(ctx, "order.process")
defer span.End()

start := time.Now()
err := processOrder(ctx, order)
duration := time.Since(start)

// Record both trace and metric
if err != nil {
    metrics.RecordError("order_processing")
    tracing.RecordError(span, err)
} else {
    metrics.RecordDuration("order_processing", duration)
    tracing.SetSpanAttributes(span, map[string]interface{}{
        "duration_ms": duration.Milliseconds(),
    })
}
```

### With Error Tracking

```go
ctx, span := tracing.StartSpan(ctx, "payment.charge")
defer span.End()

if err != nil {
    // Record in both systems
    tracing.RecordError(span, err)
    sentry.CaptureException(err)

    // Include trace context in Sentry
    sentry.ConfigureScope(func(scope *sentry.Scope) {
        scope.SetTag("trace_id", tracing.GetTraceID(ctx))
    })
}
```

## Performance Impact

### Overhead by Sampling Rate

| Sample Rate | Overhead | Use Case |
|-------------|----------|----------|
| 0% (disabled) | ~0ms | - |
| 1% | <0.1ms | High-volume production |
| 10% | <0.5ms | Medium-volume production |
| 100% | 1-2ms | Development, debugging |

### Optimization Tips

1. **Batch exports** (already configured)
2. **Use head-based sampling** (current approach)
3. **Consider tail-based sampling** for advanced use cases
4. **Export traces asynchronously** (already implemented)
5. **Limit span attributes** to essential data

## Monitoring Tracing Health

### Check Exporter Status

```bash
# View tracer provider metrics (if enabled)
curl http://localhost:9090/metrics | grep otel
```

### Common Issues

**Memory growth:**
- Check batch size and export interval
- Ensure cleanup functions are called

**Missing spans:**
- Verify context propagation
- Check for panics that skip span.End()

**Slow exports:**
- Check network connectivity to backend
- Increase batch timeout
- Use async export (already enabled)

## Further Reading

- [OpenTelemetry Documentation](https://opentelemetry.io/docs/)
- [Jaeger Documentation](https://www.jaegertracing.io/docs/)
- [W3C Trace Context](https://www.w3.org/TR/trace-context/)
- [Distributed Tracing Best Practices](https://opentelemetry.io/docs/concepts/signals/traces/)

## Support

For issues or questions:
- Create an issue on GitHub
- Join our Discord community
- Check existing documentation in `/docs`
