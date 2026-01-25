# Gorax Performance Monitoring Guide

Complete guide for monitoring Gorax performance using Prometheus and Grafana.

## Table of Contents

- [Quick Start](#quick-start)
- [Architecture Overview](#architecture-overview)
- [Metrics Reference](#metrics-reference)
- [Dashboards](#dashboards)
- [Alerting](#alerting)
- [Production Setup](#production-setup)
- [Troubleshooting](#troubleshooting)

---

## Quick Start

Get monitoring running in 5 minutes for local development.

### Prerequisites

- Docker and Docker Compose installed
- Gorax project cloned locally
- Port 3000 (Grafana), 9090 (Prometheus), and 9091 (Gorax metrics) available

### Start Monitoring Stack

```bash
# 1. Start the monitoring stack (Prometheus + Grafana)
docker-compose -f docker-compose.monitoring.yml up -d

# 2. Verify services are running
docker-compose -f docker-compose.monitoring.yml ps

# Expected output:
# NAME                STATUS    PORTS
# gorax-prometheus    healthy   0.0.0.0:9090->9090/tcp
# gorax-grafana       healthy   0.0.0.0:3000->3000/tcp
```

### Start Gorax with Metrics Enabled

```bash
# 3. Start database dependencies
docker-compose -f docker-compose.dev.yml up -d

# 4. Configure metrics in .env
cat >> .env << EOF
ENABLE_METRICS=true
METRICS_PORT=9091
EOF

# 5. Start the Gorax API
make run-api-dev

# 6. Verify metrics endpoint is working
curl http://localhost:9091/metrics
```

### Access Dashboards

```bash
# Open Grafana in your browser
open http://localhost:3000

# Default credentials:
# Username: admin
# Password: admin
```

**Navigate to:** Dashboards → Gorax folder → **Gorax Performance Overview**

---

## Architecture Overview

### Components

```
┌─────────────┐      ┌─────────────┐      ┌─────────────┐
│   Gorax     │      │ Prometheus  │      │  Grafana    │
│   API       │─────▶│   (TSDB)    │─────▶│ (Dashboards)│
│             │      │             │      │             │
│ :9091       │      │ :9090       │      │ :3000       │
│ /metrics    │      │             │      │             │
└─────────────┘      └─────────────┘      └─────────────┘
      │                     │
      │                     │
      ▼                     ▼
┌─────────────┐      ┌─────────────┐
│ PostgreSQL  │      │Alertmanager │
│             │      │  (Optional) │
│ :5432       │      │ :9093       │
└─────────────┘      └─────────────┘
```

### Data Flow

1. **Gorax API** exposes metrics at `/metrics` endpoint (Prometheus format)
2. **Prometheus** scrapes metrics every 15 seconds and stores in time-series database
3. **Grafana** queries Prometheus and visualizes metrics in dashboards
4. **Alertmanager** (optional) receives alerts from Prometheus and sends notifications

### Ports

| Service | Port | Purpose |
|---------|------|---------|
| Gorax API | 9091 | Metrics endpoint (`/metrics`) |
| Prometheus | 9090 | Metrics collection, querying, alerting |
| Grafana | 3000 | Dashboard visualization |
| Alertmanager | 9093 | Alert routing and notifications |

---

## Metrics Reference

### Workflow Metrics

#### `gorax_workflow_executions_total`
**Type:** Counter
**Description:** Total number of workflow executions
**Labels:**
- `status`: `completed`, `failed`, `cancelled`
- `trigger_type`: `manual`, `webhook`, `schedule`, `api`

**Example Query:**
```promql
# Workflow execution rate by status
rate(gorax_workflow_executions_total[5m])

# Success rate
sum(rate(gorax_workflow_executions_total{status="completed"}[5m]))
/
sum(rate(gorax_workflow_executions_total[5m]))
```

#### `gorax_workflow_execution_duration_seconds`
**Type:** Histogram
**Description:** Workflow execution duration in seconds
**Labels:**
- `trigger_type`: Type of trigger that started the workflow

**Example Query:**
```promql
# P99 execution duration
histogram_quantile(0.99,
  rate(gorax_workflow_execution_duration_seconds_bucket[5m])
)

# Average execution time
rate(gorax_workflow_execution_duration_seconds_sum[5m])
/
rate(gorax_workflow_execution_duration_seconds_count[5m])
```

### Formula Evaluation Metrics

#### `gorax_formula_evaluations_total`
**Type:** Counter
**Description:** Total number of formula evaluations
**Labels:**
- `status`: `success`, `error`

**Example Query:**
```promql
# Formula evaluation rate
rate(gorax_formula_evaluations_total[5m])

# Error rate
sum(rate(gorax_formula_evaluations_total{status="error"}[5m]))
/
sum(rate(gorax_formula_evaluations_total[5m]))
```

#### `gorax_formula_evaluation_duration_seconds`
**Type:** Histogram
**Description:** Formula evaluation duration in seconds

**Example Query:**
```promql
# P99 evaluation latency
histogram_quantile(0.99,
  rate(gorax_formula_evaluation_duration_seconds_bucket[5m])
)
```

#### `gorax_formula_cache_hits_total`
**Type:** Counter
**Description:** Total number of formula cache hits

#### `gorax_formula_cache_misses_total`
**Type:** Counter
**Description:** Total number of formula cache misses

**Example Query:**
```promql
# Cache hit rate
sum(rate(gorax_formula_cache_hits_total[5m]))
/
(
  sum(rate(gorax_formula_cache_hits_total[5m])) +
  sum(rate(gorax_formula_cache_misses_total[5m]))
)
```

### Database Metrics

#### `gorax_db_queries_total`
**Type:** Counter
**Description:** Total number of database queries
**Labels:**
- `operation`: `select`, `insert`, `update`, `delete`
- `table`: Table name
- `status`: `success`, `error`

**Example Query:**
```promql
# Query rate by table
sum by (table) (rate(gorax_db_queries_total[5m]))

# Error rate by table
sum by (table) (
  rate(gorax_db_queries_total{status="error"}[5m])
)
```

#### `gorax_db_query_duration_seconds`
**Type:** Histogram
**Description:** Database query duration in seconds
**Labels:**
- `operation`: Query operation type
- `table`: Table name

**Example Query:**
```promql
# Slowest tables (P95)
topk(10,
  histogram_quantile(0.95,
    rate(gorax_db_query_duration_seconds_bucket[5m])
  )
)
```

#### `gorax_db_connections_open`
**Type:** Gauge
**Description:** Current number of open database connections
**Labels:**
- `pool`: Connection pool name

#### `gorax_db_connections_in_use`
**Type:** Gauge
**Description:** Current number of database connections in use
**Labels:**
- `pool`: Connection pool name

#### `gorax_db_connections_idle`
**Type:** Gauge
**Description:** Current number of idle database connections
**Labels:**
- `pool`: Connection pool name

**Example Query:**
```promql
# Connection pool utilization
(gorax_db_connections_in_use / gorax_db_connections_open) * 100
```

### HTTP API Metrics

#### `gorax_http_requests_total`
**Type:** Counter
**Description:** Total number of HTTP requests
**Labels:**
- `method`: HTTP method (`GET`, `POST`, `PUT`, `DELETE`)
- `path`: Request path
- `status`: HTTP status code

**Example Query:**
```promql
# Request rate by endpoint
sum by (method, path) (
  rate(gorax_http_requests_total[5m])
)

# Error rate (5xx errors)
sum(rate(gorax_http_requests_total{status=~"5.."}[5m]))
/
sum(rate(gorax_http_requests_total[5m]))
```

#### `gorax_http_request_duration_seconds`
**Type:** Histogram
**Description:** HTTP request duration in seconds
**Labels:**
- `method`: HTTP method
- `path`: Request path

**Example Query:**
```promql
# P99 latency by endpoint
histogram_quantile(0.99,
  sum by (method, path, le) (
    rate(gorax_http_request_duration_seconds_bucket[5m])
  )
)

# Slowest endpoints
topk(10,
  histogram_quantile(0.99,
    rate(gorax_http_request_duration_seconds_bucket[5m])
  )
)
```

---

## Dashboards

### Gorax Performance Overview

**Location:** `dashboards/gorax-performance-overview.json`

**Sections:**

1. **Workflow Execution Performance**
   - Workflow execution rate by trigger type
   - Workflow execution duration (P50, P90, P99)

2. **Formula Evaluation Performance**
   - Formula cache hit rate (gauge)
   - Formula evaluation duration (P50, P90, P99)
   - Formula evaluation rate

3. **Database Performance**
   - Connection pool utilization
   - Database query duration (P90, P99)
   - Database query rate

4. **API Performance**
   - HTTP request rate by endpoint
   - HTTP request duration (P50, P90, P99)

**Time Range:** Last 1 hour (default), refresh every 10 seconds

**Import Instructions:**

See [GRAFANA_DEPLOYMENT_GUIDE.md](./GRAFANA_DEPLOYMENT_GUIDE.md#dashboard-import) for detailed instructions.

---

## Alerting

### Configured Alerts

All alerts are defined in `configs/prometheus-alerts.yml`.

#### Workflow Alerts

| Alert | Threshold | Duration | Severity |
|-------|-----------|----------|----------|
| HighWorkflowFailureRate | > 10% | 5m | warning |
| CriticalWorkflowFailureRate | > 25% | 3m | critical |
| SlowWorkflowExecutions | P99 > 60s | 10m | warning |
| NoWorkflowExecutions | 0 executions | 15m | warning |

#### Formula Alerts

| Alert | Threshold | Duration | Severity |
|-------|-----------|----------|----------|
| LowFormulaCacheHitRate | < 80% | 10m | warning |
| VeryLowFormulaCacheHitRate | < 50% | 5m | critical |
| HighFormulaEvaluationLatency | P99 > 500ms | 5m | warning |
| HighFormulaEvaluationErrorRate | > 5% | 5m | warning |

#### Database Alerts

| Alert | Threshold | Duration | Severity |
|-------|-----------|----------|----------|
| HighDBConnectionPoolUtilization | > 90% | 5m | warning |
| DBConnectionPoolExhausted | 100% | 2m | critical |
| SlowDatabaseQueries | P95 > 1s | 5m | warning |
| VerySlowDatabaseQueries | P99 > 5s | 3m | critical |
| HighDatabaseErrorRate | > 1% | 5m | warning |

#### API Alerts

| Alert | Threshold | Duration | Severity |
|-------|-----------|----------|----------|
| HighAPIErrorRate | 5xx > 5% | 5m | critical |
| HighAPIClientErrorRate | 4xx > 20% | 10m | warning |
| HighAPILatency | P99 > 2s | 5m | warning |
| VeryHighAPILatency | P99 > 5s | 3m | critical |
| NoAPITraffic | 0 requests | 10m | warning |

#### Service Health Alerts

| Alert | Threshold | Duration | Severity |
|-------|-----------|----------|----------|
| GoraxAPIDown | Target down | 2m | critical |

### Viewing Active Alerts

**Prometheus UI:**
```bash
open http://localhost:9090/alerts
```

**Grafana:**
Navigate to **Alerting** → **Alert rules**

### Setting Up Notifications

See [GRAFANA_DEPLOYMENT_GUIDE.md](./GRAFANA_DEPLOYMENT_GUIDE.md#alerting-setup) for:
- Alertmanager configuration
- Slack integration
- PagerDuty integration
- Email notifications

---

## Production Setup

### High Availability

**Prometheus:**
- Run multiple Prometheus instances for redundancy
- Use Thanos or Cortex for long-term storage and federation
- Configure remote write to external storage

**Grafana:**
- Use external PostgreSQL database instead of SQLite
- Configure session storage in Redis
- Run multiple Grafana instances behind a load balancer

### Security

**Authentication:**
```yaml
# Grafana: Enable OAuth
[auth.generic_oauth]
enabled = true
name = OAuth
client_id = YOUR_CLIENT_ID
client_secret = YOUR_CLIENT_SECRET
auth_url = https://auth.example.com/authorize
token_url = https://auth.example.com/token
```

**Network Security:**
```yaml
# Restrict metrics endpoint
# Add authentication middleware to /metrics
router.GET("/metrics", middleware.RequireAuth(), gin.WrapH(promhttp.Handler()))
```

**TLS Encryption:**
```yaml
# Prometheus
tls_config:
  cert_file: /etc/prometheus/certs/cert.pem
  key_file: /etc/prometheus/certs/key.pem

# Grafana
[server]
protocol = https
cert_file = /etc/grafana/ssl/cert.pem
cert_key = /etc/grafana/ssl/key.pem
```

### Retention and Storage

**Prometheus:**
```yaml
# Keep 90 days of data
--storage.tsdb.retention.time=90d

# Or limit by size
--storage.tsdb.retention.size=50GB
```

**Long-Term Storage:**
```yaml
# Thanos Sidecar
thanos:
  image: thanosio/thanos:v0.32.5
  command:
    - 'sidecar'
    - '--tsdb.path=/prometheus'
    - '--objstore.config-file=/etc/thanos/bucket.yml'
    - '--prometheus.url=http://prometheus:9090'
```

### Backup Strategy

**Prometheus Snapshots:**
```bash
# Create snapshot
curl -XPOST http://localhost:9090/api/v1/admin/tsdb/snapshot

# Backup snapshot
tar czf prometheus-backup-$(date +%Y%m%d).tar.gz \
  /var/lib/prometheus/snapshots/
```

**Grafana Dashboards:**
```bash
# Automated backup script
./scripts/backup-grafana-dashboards.sh
```

### Scaling Considerations

| Component | Scaling Strategy |
|-----------|------------------|
| Prometheus | Horizontal sharding by environment/service |
| Grafana | Horizontal with shared database |
| Alertmanager | Cluster mode for HA |
| Gorax API | Ensure all instances expose metrics on same path |

---

## Troubleshooting

### No Data in Dashboards

**Check Prometheus targets:**
```bash
curl http://localhost:9090/api/v1/targets | jq
```

**Verify Gorax metrics endpoint:**
```bash
curl http://localhost:9091/metrics
```

**Check Prometheus logs:**
```bash
docker logs gorax-prometheus
```

### High Cardinality Issues

**Symptom:** Prometheus using excessive memory

**Solution:** Reduce label cardinality
```yaml
# prometheus.yml
metric_relabel_configs:
  - source_labels: [user_id]
    action: drop
```

### Alert Fatigue

**Symptom:** Too many alerts firing

**Solutions:**
- Increase alert thresholds
- Extend alert duration (`for: 10m` → `for: 15m`)
- Use inhibition rules to suppress lower-severity alerts
- Implement time-based muting during maintenance windows

### Slow Dashboard Loading

**Solutions:**
- Reduce time range (e.g., 1 hour instead of 24 hours)
- Use recording rules for complex queries
- Increase Grafana memory limit
- Add dashboard-level caching

**Recording Rules Example:**
```yaml
# prometheus.yml
groups:
  - name: gorax_recordings
    interval: 30s
    rules:
      - record: job:gorax_workflow_success_rate:5m
        expr: |
          sum(rate(gorax_workflow_executions_total{status="completed"}[5m]))
          /
          sum(rate(gorax_workflow_executions_total[5m]))
```

---

## Best Practices

### 1. Start with Standard Metrics

Don't over-instrument initially. Focus on:
- Request rate, errors, duration (RED method)
- Workflow execution metrics
- Database query performance

### 2. Use Consistent Labels

```go
// Good: Consistent label names
prometheus.NewCounterVec(prometheus.CounterOpts{
  Name: "gorax_http_requests_total",
}, []string{"method", "path", "status"})

// Bad: Inconsistent naming
prometheus.NewCounterVec(prometheus.CounterOpts{
  Name: "gorax_http_requests_total",
}, []string{"httpMethod", "endpoint", "statusCode"})
```

### 3. Keep Cardinality Low

```go
// Good: Low cardinality
labels := prometheus.Labels{
  "status": status,  // ~3 values: success, error, timeout
}

// Bad: High cardinality
labels := prometheus.Labels{
  "user_id": userID,  // Potentially millions of values
  "request_id": reqID,  // Unique per request
}
```

### 4. Set Appropriate Alert Thresholds

- Start conservative (high thresholds, long durations)
- Iterate based on actual incidents
- Document why each threshold was chosen

### 5. Use Runbooks

Every alert should have a runbook URL:
```yaml
annotations:
  runbook_url: "https://github.com/stherrien/gorax/wiki/Runbook-High-API-Latency"
```

---

## Additional Resources

### Documentation

- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [PromQL Tutorial](https://prometheus.io/docs/prometheus/latest/querying/basics/)
- [Dashboards README](../dashboards/README.md)
- [Grafana Deployment Guide](./GRAFANA_DEPLOYMENT_GUIDE.md)

### Example Queries

See [dashboards/README.md](../dashboards/README.md) for:
- Common PromQL queries
- Troubleshooting queries
- Performance investigation examples

### Community

- [Prometheus GitHub](https://github.com/prometheus/prometheus)
- [Grafana GitHub](https://github.com/grafana/grafana)
- [CNCF Slack #prometheus](https://slack.cncf.io/)
- [Grafana Community Forums](https://community.grafana.com/)

---

## Quick Reference Commands

```bash
# Start monitoring stack
docker-compose -f docker-compose.monitoring.yml up -d

# View logs
docker-compose -f docker-compose.monitoring.yml logs -f

# Stop monitoring stack
docker-compose -f docker-compose.monitoring.yml down

# Remove all data (WARNING: destructive)
docker-compose -f docker-compose.monitoring.yml down -v

# Restart Prometheus (reload config)
docker-compose -f docker-compose.monitoring.yml restart prometheus

# Access Prometheus
open http://localhost:9090

# Access Grafana
open http://localhost:3000

# Query Gorax metrics
curl http://localhost:9091/metrics

# Check Prometheus targets
curl http://localhost:9090/api/v1/targets | jq

# Check active alerts
curl http://localhost:9090/api/v1/alerts | jq

# Validate Prometheus config
docker exec gorax-prometheus promtool check config /etc/prometheus/prometheus.yml

# Validate alert rules
docker exec gorax-prometheus promtool check rules /etc/prometheus/alerts.yml
```

---

**Need Help?**

- Check [GRAFANA_DEPLOYMENT_GUIDE.md](./GRAFANA_DEPLOYMENT_GUIDE.md) for detailed setup instructions
- Review [dashboards/README.md](../dashboards/README.md) for query examples
- Open an issue on GitHub for monitoring-related questions
