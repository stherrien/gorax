# Distributed Tracing

This package provides distributed tracing support for the gorax workflow automation platform using OpenTelemetry.

## Features

- **OpenTelemetry Integration**: Full OTLP (OpenTelemetry Protocol) support for sending traces to compatible backends
- **HTTP Request Tracing**: Automatic instrumentation of incoming HTTP requests
- **Workflow Execution Tracing**: Detailed traces for workflow and step executions
- **Queue Message Tracing**: Tracing for queue message processing
- **Context Propagation**: W3C Trace Context propagation across service boundaries
- **Configurable Sampling**: Control sampling rate to manage trace volume

## Configuration

Tracing is configured via environment variables:

```bash
# Enable tracing
TRACING_ENABLED=true

# OTLP endpoint (supports both gRPC and HTTP)
TRACING_ENDPOINT=localhost:4317

# Service name (appears in traces)
TRACING_SERVICE_NAME=gorax

# Sample rate (0.0 to 1.0)
# 1.0 = trace all requests
# 0.5 = trace 50% of requests
# 0.0 = trace no requests
TRACING_SAMPLE_RATE=1.0
```

## Trace Backends

This implementation uses OTLP, which is compatible with many backends:

### Jaeger
```bash
# Run Jaeger all-in-one (includes OTLP receiver)
docker run -d --name jaeger \
  -e COLLECTOR_OTLP_ENABLED=true \
  -p 16686:16686 \
  -p 4317:4317 \
  -p 4318:4318 \
  jaegertracing/all-in-one:latest

# UI available at http://localhost:16686
```

### OpenTelemetry Collector
```yaml
# otel-collector-config.yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317

exporters:
  jaeger:
    endpoint: jaeger:14250
  logging:
    loglevel: debug

service:
  pipelines:
    traces:
      receivers: [otlp]
      exporters: [jaeger, logging]
```

### Grafana Tempo
```bash
# Use Grafana Cloud or self-hosted Tempo
TRACING_ENDPOINT=tempo.grafana.net:443
```

## Usage

### Initializing Tracing

Tracing is automatically initialized in `cmd/api/main.go` and `cmd/worker/main.go`:

```go
import "github.com/gorax/gorax/internal/tracing"

// Initialize global tracer
cleanup, err := tracing.InitGlobalTracer(ctx, &cfg.Observability)
if err != nil {
    log.Fatal(err)
}
defer cleanup()
```

### HTTP Middleware

HTTP tracing middleware is automatically added to the API server:

```go
import "github.com/gorax/gorax/internal/tracing"

r.Use(tracing.HTTPMiddleware())
```

This creates spans for all HTTP requests with:
- Request method and path
- Response status code
- Request duration
- User agent and other HTTP attributes

### Manual Span Creation

Create custom spans for specific operations:

```go
import "github.com/gorax/gorax/internal/tracing"

ctx, span := tracing.StartSpan(ctx, "custom.operation")
defer span.End()

// Add attributes
tracing.SetSpanAttributes(span, map[string]interface{}{
    "user_id": "123",
    "workflow_id": "wf-456",
})

// Record errors
if err != nil {
    tracing.RecordError(span, err)
    return err
}
```

### Workflow Execution Tracing

Workflows are automatically traced:

```go
import "github.com/gorax/gorax/internal/tracing"

err := tracing.TraceWorkflowExecution(ctx, tenantID, workflowID, executionID, func(ctx context.Context) error {
    // Execute workflow
    return executor.Execute(ctx, execution)
})
```

This creates a span with:
- Tenant ID
- Workflow ID
- Execution ID
- Duration and status

### Step Execution Tracing

Individual workflow steps are traced:

```go
output, err := tracing.TraceStepExecution(ctx, tenantID, workflowID, executionID, nodeID, nodeType, func(ctx context.Context) (interface{}, error) {
    // Execute step
    return executor.executeStep(ctx, step)
})
```

### Queue Message Tracing

Queue message processing is traced:

```go
err := tracing.TraceQueueMessage(ctx, queueName, messageID, func(ctx context.Context) error {
    // Process message
    return handler.HandleMessage(ctx, msg)
})
```

### HTTP Client Tracing

Trace outgoing HTTP requests:

```go
import "github.com/gorax/gorax/internal/tracing"

// Wrap HTTP transport
client := &http.Client{
    Transport: tracing.HTTPClientMiddleware(http.DefaultTransport),
}
```

## Trace Context Propagation

Trace context is automatically propagated:

### Incoming Requests
The HTTP middleware extracts trace context from incoming request headers and creates child spans.

### Outgoing Requests
Use the HTTP client middleware to inject trace context into outgoing requests:

```go
req, _ := http.NewRequest("GET", "https://api.example.com", nil)
tracing.InjectHTTPTraceContext(req)
```

### Manual Propagation
```go
// Extract from headers
headers := map[string]string{
    "traceparent": req.Header.Get("traceparent"),
}
ctx = tracing.ExtractTraceContext(ctx, headers)

// Inject into headers
headers := make(map[string]string)
tracing.InjectTraceContext(ctx, headers)
```

## Span Hierarchy

Traces are organized hierarchically:

```
HTTP Request (api.request)
└─ Workflow Execution (workflow.execute)
   ├─ Step Execution (workflow.step.execute)
   │  └─ HTTP Action (http.action)
   ├─ Step Execution (workflow.step.execute)
   │  └─ Transform Action (transform.action)
   └─ Step Execution (workflow.step.execute)
      └─ Sub-workflow (workflow.sub_workflow)
```

## Span Attributes

Standard attributes added to spans:

### HTTP Spans
- `http.method`: Request method (GET, POST, etc.)
- `http.target`: Request path
- `http.status_code`: Response status code
- `http.user_agent`: User agent string

### Workflow Spans
- `tenant_id`: Tenant identifier
- `workflow_id`: Workflow identifier
- `execution_id`: Execution identifier
- `node_id`: Node/step identifier
- `node_type`: Node type (http, transform, etc.)
- `component`: Component name (executor, queue_consumer)

### Queue Spans
- `queue.name`: Queue name
- `queue.message_id`: Message identifier

## Events

Record events within spans:

```go
tracing.RecordWorkflowEvent(ctx, "step.started", map[string]interface{}{
    "node_id": "node-123",
    "node_type": "http",
})
```

## Performance Considerations

### Sampling
Use sampling to reduce overhead in high-traffic environments:

```bash
# Trace 10% of requests
TRACING_SAMPLE_RATE=0.1
```

### Overhead
- Minimal overhead when disabled (no-op tracer)
- ~1-2ms overhead per span when enabled
- Batched export to minimize network calls

### Resource Usage
- Spans are buffered in memory before export
- Automatic backpressure handling
- Configurable batch size and timeout

## Testing

Tests are included for all tracing functionality:

```bash
# Run tracing tests
go test ./internal/tracing/... -v

# Run with short mode (faster, may skip some tests)
go test ./internal/tracing/... -short
```

## Troubleshooting

### No traces appearing

1. **Check configuration**:
   ```bash
   echo $TRACING_ENABLED
   echo $TRACING_ENDPOINT
   ```

2. **Verify backend is running**:
   ```bash
   # For Jaeger
   curl http://localhost:16686

   # For OTLP endpoint
   telnet localhost 4317
   ```

3. **Check logs**:
   Look for tracing initialization messages in application logs.

### Sampling issues

If too few/many traces:
```bash
# Adjust sample rate
TRACING_SAMPLE_RATE=1.0  # Trace everything
TRACING_SAMPLE_RATE=0.1  # Trace 10%
```

### Context not propagating

Ensure context is passed through all function calls:
```go
// ❌ Bad - loses context
doWork()

// ✅ Good - passes context
doWork(ctx)
```

## Architecture

```
┌─────────────────┐
│   Application   │
│   (API/Worker)  │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Tracing Package │
│   (internal/    │
│    tracing)     │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  OpenTelemetry  │
│      SDK        │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  OTLP Exporter  │
│   (gRPC/HTTP)   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Trace Backend  │
│ (Jaeger/Tempo/  │
│   Collector)    │
└─────────────────┘
```

## Future Enhancements

- [ ] Add Jaeger native exporter (currently OTLP only)
- [ ] Implement trace baggage for cross-service metadata
- [ ] Add trace sampling based on attributes (e.g., only trace errors)
- [ ] Implement trace exemplars for linking to logs
- [ ] Add support for distributed context propagation in async jobs
- [ ] Implement custom span processors for sensitive data masking
