# Gorax Observability Guide

This document describes the observability infrastructure for Gorax, including metrics, tracing, error tracking, and health checks.

## Overview

Gorax implements comprehensive observability using:

- **Prometheus** for metrics collection and monitoring
- **OpenTelemetry** for distributed tracing
- **Sentry** for error tracking and reporting
- **Structured logging** with slog for application logs

## Configuration

### Environment Variables

```bash
# Metrics Configuration
METRICS_ENABLED=true          # Enable Prometheus metrics (default: true)
METRICS_PORT=9090            # Metrics endpoint port (default: 9090)

# Tracing Configuration
TRACING_ENABLED=false        # Enable OpenTelemetry tracing (default: false)
TRACING_ENDPOINT=localhost:4317  # OTLP endpoint (default: localhost:4317)
TRACING_SAMPLE_RATE=1.0      # Sampling rate 0.0-1.0 (default: 1.0)
TRACING_SERVICE_NAME=gorax   # Service name in traces (default: gorax)

# Error Tracking Configuration
SENTRY_ENABLED=false         # Enable Sentry error tracking (default: false)
SENTRY_DSN=                  # Sentry DSN (required if enabled)
SENTRY_ENVIRONMENT=production # Environment tag (default: development)
SENTRY_SAMPLE_RATE=1.0       # Error sampling rate (default: 1.0)
```

## Metrics

### Available Metrics

#### Workflow Metrics

- `gorax_workflow_executions_total{tenant_id, workflow_id, status}` - Total workflow executions by status
- `gorax_workflow_execution_duration_seconds{tenant_id, workflow_id}` - Workflow execution duration histogram

#### Step Metrics

- `gorax_step_executions_total{tenant_id, workflow_id, step_type, status}` - Total step executions
- `gorax_step_execution_duration_seconds{tenant_id, workflow_id, step_type}` - Step execution duration histogram

#### Queue Metrics

- `gorax_queue_depth{queue}` - Current queue depth by queue name
- `gorax_active_workers` - Number of active workers processing jobs

#### HTTP Metrics

- `gorax_http_requests_total{method, path, status}` - Total HTTP requests by method, path, and status
- `gorax_http_request_duration_seconds{method, path}` - HTTP request latency histogram

### Metrics Endpoint

Metrics are exposed on a separate port for security:

```
http://localhost:9090/metrics
```

### Prometheus Configuration

Example `prometheus.yml`:

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'gorax-api'
    static_configs:
      - targets: ['localhost:9090']
        labels:
          service: 'gorax'
          component: 'api'

  - job_name: 'gorax-worker'
    static_configs:
      - targets: ['worker:9090']
        labels:
          service: 'gorax'
          component: 'worker'
```

## Distributed Tracing

### OpenTelemetry Setup

Gorax uses OpenTelemetry for distributed tracing, compatible with Jaeger, Zipkin, and other OTLP-compliant backends.

### Trace Spans

The following operations are automatically traced:

- **HTTP Requests**: Each incoming request creates a root span
- **Workflow Execution**: Complete workflow execution from trigger to completion
- **Step Execution**: Individual step execution within workflows
- **Database Queries**: SQL queries and transaction spans
- **Queue Operations**: SQS message send/receive operations
- **HTTP Client Calls**: Outgoing HTTP requests from actions

### Span Attributes

Standard attributes added to all spans:

- `tenant_id` - Tenant identifier
- `workflow_id` - Workflow identifier
- `execution_id` - Execution identifier
- `step_id` - Step identifier (for step spans)
- `action_type` - Action type (for action spans)
- `user_id` - User who triggered the execution

### Running with Jaeger

1. Start Jaeger:

```bash
docker run -d --name jaeger \
  -p 6831:6831/udp \
  -p 16686:16686 \
  -p 4317:4317 \
  jaegertracing/all-in-one:latest
```

2. Configure Gorax:

```bash
TRACING_ENABLED=true
TRACING_ENDPOINT=localhost:4317
TRACING_SERVICE_NAME=gorax
```

3. Access Jaeger UI: http://localhost:16686

## Error Tracking

### Sentry Integration

Sentry automatically captures:

- **Panics**: Application panics with full stack traces
- **Errors**: Workflow execution errors
- **Step Failures**: Failed action executions
- **HTTP Errors**: 5xx responses

### Error Context

Errors include contextual information:

- Request ID
- Trace ID
- Tenant ID
- Workflow ID
- Execution ID
- User ID
- Environment tags

### Setup

1. Create Sentry project at https://sentry.io

2. Configure Gorax:

```bash
SENTRY_ENABLED=true
SENTRY_DSN=https://[key]@sentry.io/[project]
SENTRY_ENVIRONMENT=production
SENTRY_SAMPLE_RATE=1.0
```

## Health Checks

### Endpoints

- `GET /health` - Liveness probe (is process running)
- `GET /ready` - Readiness probe (can accept traffic)
- `GET /health/live` - Kubernetes liveness probe
- `GET /health/ready` - Kubernetes readiness probe
- `GET /health/startup` - Kubernetes startup probe

### Health Check Details

#### Liveness (`/health/live`)
Returns 200 if the process is running. Used by Kubernetes to restart unhealthy pods.

```json
{
  "status": "ok",
  "timestamp": "2025-01-15T10:30:00Z"
}
```

#### Readiness (`/health/ready`)
Returns 200 if the service can accept traffic. Checks:
- Database connectivity
- Redis connectivity
- Queue connectivity (if enabled)

```json
{
  "status": "ok",
  "checks": {
    "database": "ok",
    "redis": "ok",
    "queue": "ok"
  },
  "timestamp": "2025-01-15T10:30:00Z"
}
```

## Alerting

### Prometheus Alerts

Example alerting rules in `/deployments/kubernetes/monitoring/alerts.yaml`:

```yaml
groups:
- name: gorax
  rules:
  - alert: HighWorkflowFailureRate
    expr: rate(gorax_workflow_executions_total{status="failed"}[5m]) > 0.1
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: High workflow failure rate detected
      description: Workflow failure rate is {{ $value }} per second

  - alert: QueueBacklog
    expr: gorax_queue_depth > 1000
    for: 10m
    labels:
      severity: warning
    annotations:
      summary: Queue backlog growing
      description: Queue depth is {{ $value }} messages

  - alert: SlowWorkflowExecutions
    expr: histogram_quantile(0.95, gorax_workflow_execution_duration_seconds) > 60
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: Workflow executions are slow
      description: 95th percentile execution time is {{ $value }} seconds

  - alert: HighHTTPErrorRate
    expr: rate(gorax_http_requests_total{status=~"5.."}[5m]) > 0.05
    for: 5m
    labels:
      severity: critical
    annotations:
      summary: High HTTP error rate
      description: HTTP 5xx error rate is {{ $value }} per second
```

## Grafana Dashboards

Pre-built dashboards are available in `/deployments/kubernetes/monitoring/dashboards/`:

### 1. Gorax Overview (`gorax-overview.json`)
- System health overview
- Request rate and error rate
- Queue depth and worker utilization
- Top workflows by execution count

### 2. Workflow Metrics (`gorax-workflows.json`)
- Workflow execution rate by status
- Execution duration percentiles (p50, p95, p99)
- Success rate over time
- Top failing workflows
- Execution breakdown by trigger type

### 3. API Performance (`gorax-api.json`)
- Request rate by endpoint
- Response time percentiles
- Error rate by endpoint
- Request volume heatmap
- Slowest endpoints

### 4. Worker Metrics (`gorax-workers.json`)
- Active workers gauge
- Queue depth over time
- Message processing rate
- Worker errors
- Processing duration

## Structured Logging

### Log Format

Gorax uses structured JSON logging in production:

```json
{
  "time": "2025-01-15T10:30:00Z",
  "level": "INFO",
  "msg": "workflow execution completed",
  "trace_id": "abc123",
  "tenant_id": "tenant-123",
  "workflow_id": "wf-456",
  "execution_id": "exec-789",
  "duration_ms": 1250,
  "status": "completed"
}
```

### Log Correlation

All logs include `trace_id` for correlation with distributed traces.

### Log Levels

- **DEBUG**: Detailed diagnostic information
- **INFO**: General informational messages
- **WARN**: Warning messages for potentially harmful situations
- **ERROR**: Error messages for error events

## Best Practices

### Metrics

1. **Use labels wisely**: Avoid high-cardinality labels (like execution_id)
2. **Set appropriate histogram buckets**: Based on expected latency distribution
3. **Monitor scrape health**: Ensure Prometheus can scrape all targets

### Tracing

1. **Use sampling in production**: Set sample rate < 1.0 for high-traffic services
2. **Add custom attributes**: Enrich spans with business context
3. **Limit span count**: Avoid creating too many spans per trace

### Error Tracking

1. **Filter noisy errors**: Use Sentry's filtering to reduce noise
2. **Set proper sampling**: Avoid overwhelming Sentry with errors
3. **Add custom context**: Include relevant business data

### Alerting

1. **Start with critical alerts**: Add more alerts as you learn patterns
2. **Set appropriate thresholds**: Balance sensitivity vs. false positives
3. **Include runbooks**: Add links to troubleshooting guides in alert annotations

## Troubleshooting

### Metrics not appearing

1. Check metrics are enabled: `METRICS_ENABLED=true`
2. Verify metrics endpoint is accessible: `curl http://localhost:9090/metrics`
3. Check Prometheus can scrape the target
4. Verify registry is properly initialized

### Traces not appearing

1. Check tracing is enabled: `TRACING_ENABLED=true`
2. Verify OTLP endpoint is reachable: `TRACING_ENDPOINT=localhost:4317`
3. Check sampling rate: `TRACING_SAMPLE_RATE > 0`
4. Verify trace backend (Jaeger/Zipkin) is running

### High cardinality warnings

If Prometheus warns about high cardinality:

1. Review label values - avoid IDs in labels
2. Use URL path normalization for HTTP metrics
3. Limit number of unique label combinations
4. Consider using exemplars instead of labels

## Integration Examples

### Custom Metrics in Code

```go
// Record custom workflow metric
metrics.RecordWorkflowExecution(
    tenantID,
    workflowID,
    "completed",
    durationSeconds,
)

// Update queue depth
metrics.SetQueueDepth("default", 42)

// Record HTTP request
metrics.RecordHTTPRequest("GET", "/api/v1/workflows", "200", 0.15)
```

### Custom Trace Spans

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
)

func ProcessWorkflow(ctx context.Context, workflow *Workflow) error {
    tracer := otel.Tracer("gorax")
    ctx, span := tracer.Start(ctx, "process_workflow")
    defer span.End()

    span.SetAttributes(
        attribute.String("workflow_id", workflow.ID),
        attribute.String("tenant_id", workflow.TenantID),
    )

    // Your processing logic
    return nil
}
```

### Error Reporting

```go
import "github.com/getsentry/sentry-go"

func handleError(err error) {
    sentry.CaptureException(err)
}

func handleErrorWithContext(ctx context.Context, err error) {
    hub := sentry.GetHubFromContext(ctx)
    if hub != nil {
        hub.CaptureException(err)
    }
}
```

## Production Checklist

- [ ] Metrics endpoint is secured (not publicly accessible)
- [ ] Tracing sampling rate is appropriate for traffic volume
- [ ] Sentry DSN is configured and tested
- [ ] Health check endpoints are configured in load balancer
- [ ] Prometheus alerting rules are deployed
- [ ] Grafana dashboards are imported
- [ ] Log aggregation is configured (e.g., ELK, Loki)
- [ ] On-call rotation is set up for critical alerts
- [ ] Runbooks are created for common issues
- [ ] Observability costs are monitored

## Additional Resources

- [Prometheus Documentation](https://prometheus.io/docs/)
- [OpenTelemetry Documentation](https://opentelemetry.io/docs/)
- [Sentry Documentation](https://docs.sentry.io/)
- [Grafana Documentation](https://grafana.com/docs/)
