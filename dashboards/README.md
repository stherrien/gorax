# Gorax Grafana Dashboards

This directory contains Grafana dashboard configurations for monitoring Gorax performance metrics.

## Available Dashboards

### 1. Gorax Performance Overview (`gorax-performance-overview.json`)

Comprehensive performance monitoring dashboard with the following sections:

- **Workflow Execution Performance**: Track workflow execution rates and durations by trigger type
- **Formula Evaluation Performance**: Monitor formula cache hit rates and evaluation times
- **Database Performance**: Connection pool utilization and query performance
- **API Performance**: HTTP request rates and response times

## Installing Dashboards

### Option 1: Import via Grafana UI

1. Open Grafana web interface
2. Navigate to **Dashboards** → **Import**
3. Click **Upload JSON file**
4. Select the dashboard JSON file from this directory
5. Configure the data source (select your Prometheus instance)
6. Click **Import**

### Option 2: Provisioning (Recommended for Production)

1. Copy dashboard JSON files to Grafana's provisioning directory:
   ```bash
   cp dashboards/*.json /etc/grafana/provisioning/dashboards/
   ```

2. Create a dashboard provider configuration file:
   ```yaml
   # /etc/grafana/provisioning/dashboards/gorax.yml
   apiVersion: 1

   providers:
     - name: 'Gorax'
       orgId: 1
       folder: 'Gorax'
       type: file
       disableDeletion: false
       updateIntervalSeconds: 10
       allowUiUpdates: true
       options:
         path: /etc/grafana/provisioning/dashboards
         foldersFromFilesStructure: true
   ```

3. Restart Grafana:
   ```bash
   systemctl restart grafana-server
   ```

## Useful PromQL Queries

### Workflow Performance

#### Workflow execution success rate (last 5 minutes)
```promql
sum(rate(gorax_workflow_executions_total{status="completed"}[5m])) /
sum(rate(gorax_workflow_executions_total[5m]))
```

#### Average workflow execution duration by trigger type
```promql
rate(gorax_workflow_execution_duration_seconds_sum[5m]) /
rate(gorax_workflow_execution_duration_seconds_count[5m])
```

#### Slow workflows (P99 > 30 seconds)
```promql
histogram_quantile(0.99,
  rate(gorax_workflow_execution_duration_seconds_bucket[5m])
) > 30
```

#### Workflows by trigger type (last hour)
```promql
sum by (trigger_type) (
  increase(gorax_workflow_executions_total[1h])
)
```

### Formula Evaluation

#### Cache hit rate percentage
```promql
sum(rate(gorax_formula_cache_hits_total[5m])) /
(sum(rate(gorax_formula_cache_hits_total[5m])) +
 sum(rate(gorax_formula_cache_misses_total[5m]))) * 100
```

#### Formula evaluation error rate
```promql
sum(rate(gorax_formula_evaluations_total{status="error"}[5m])) /
sum(rate(gorax_formula_evaluations_total[5m]))
```

#### Average formula evaluation time
```promql
rate(gorax_formula_evaluation_duration_seconds_sum[5m]) /
rate(gorax_formula_evaluation_duration_seconds_count[5m])
```

### Database Performance

#### Connection pool utilization percentage
```promql
(gorax_db_connections_in_use / gorax_db_connections_open) * 100
```

#### High connection pool utilization (> 80%)
```promql
(gorax_db_connections_in_use / gorax_db_connections_open) > 0.8
```

#### Slow database queries (P99 > 1 second)
```promql
histogram_quantile(0.99,
  rate(gorax_db_query_duration_seconds_bucket[5m])
) > 1
```

#### Database query error rate by table
```promql
sum by (table) (
  rate(gorax_db_queries_total{status="error"}[5m])
) /
sum by (table) (
  rate(gorax_db_queries_total[5m])
)
```

#### Most frequently queried tables
```promql
topk(10, sum by (table, operation) (
  rate(gorax_db_queries_total[5m])
))
```

### API Performance

#### HTTP error rate (4xx and 5xx)
```promql
sum(rate(gorax_http_requests_total{status=~"4..|5.."}[5m])) /
sum(rate(gorax_http_requests_total[5m]))
```

#### Slowest API endpoints (P99)
```promql
topk(10, histogram_quantile(0.99,
  rate(gorax_http_request_duration_seconds_bucket[5m])
))
```

#### API requests per second by endpoint
```promql
sum by (method, path) (
  rate(gorax_http_requests_total[5m])
)
```

### Queue and Worker Metrics

#### Queue depth over time
```promql
gorax_queue_depth
```

#### Queue depth alert (> 1000 messages)
```promql
gorax_queue_depth > 1000
```

#### Active workers
```promql
gorax_active_workers
```

## Alerting Rules

Example Prometheus alerting rules for common performance issues:

```yaml
# /etc/prometheus/rules/gorax-alerts.yml
groups:
  - name: gorax_performance
    interval: 30s
    rules:
      # High workflow error rate
      - alert: HighWorkflowErrorRate
        expr: |
          sum(rate(gorax_workflow_executions_total{status="failed"}[5m])) /
          sum(rate(gorax_workflow_executions_total[5m])) > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High workflow error rate detected"
          description: "More than 10% of workflows are failing (current: {{ $value | humanizePercentage }})"

      # Low formula cache hit rate
      - alert: LowFormulaCacheHitRate
        expr: |
          sum(rate(gorax_formula_cache_hits_total[5m])) /
          (sum(rate(gorax_formula_cache_hits_total[5m])) +
           sum(rate(gorax_formula_cache_misses_total[5m]))) < 0.7
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Formula cache hit rate is low"
          description: "Cache hit rate is {{ $value | humanizePercentage }}, consider increasing cache size"

      # High database connection pool utilization
      - alert: HighDBConnectionPoolUtilization
        expr: |
          (gorax_db_connections_in_use / gorax_db_connections_open) > 0.9
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Database connection pool nearly exhausted"
          description: "Connection pool {{ $labels.pool }} is {{ $value | humanizePercentage }} utilized"

      # Slow database queries
      - alert: SlowDatabaseQueries
        expr: |
          histogram_quantile(0.99,
            rate(gorax_db_query_duration_seconds_bucket[5m])
          ) > 2
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Slow database queries detected"
          description: "P99 query duration is {{ $value }}s for {{ $labels.operation }} on {{ $labels.table }}"

      # High API error rate
      - alert: HighAPIErrorRate
        expr: |
          sum(rate(gorax_http_requests_total{status=~"5.."}[5m])) /
          sum(rate(gorax_http_requests_total[5m])) > 0.05
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High API error rate detected"
          description: "More than 5% of API requests are returning 5xx errors"

      # High queue depth
      - alert: HighQueueDepth
        expr: gorax_queue_depth > 1000
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Queue depth is high"
          description: "Queue {{ $labels.queue }} has {{ $value }} messages pending"
```

## Performance Troubleshooting

### Slow Workflow Executions

1. **Check workflow execution duration**:
   ```promql
   histogram_quantile(0.99, rate(gorax_workflow_execution_duration_seconds_bucket{workflow_id="YOUR_ID"}[5m]))
   ```

2. **Identify slow steps**:
   ```promql
   topk(10, histogram_quantile(0.99,
     rate(gorax_step_execution_duration_seconds_bucket[5m])
   ))
   ```

3. **Check by trigger type**:
   ```promql
   avg by (trigger_type) (
     rate(gorax_workflow_execution_duration_seconds_sum[5m]) /
     rate(gorax_workflow_execution_duration_seconds_count[5m])
   )
   ```

### Low Formula Cache Hit Rate

1. **Check current hit rate**:
   ```promql
   sum(rate(gorax_formula_cache_hits_total[5m])) /
   (sum(rate(gorax_formula_cache_hits_total[5m])) +
    sum(rate(gorax_formula_cache_misses_total[5m])))
   ```

2. **Actions**:
   - Increase cache size in configuration
   - Check for unique formula variations
   - Review formula patterns for optimization

### Database Connection Pool Issues

1. **Check pool utilization**:
   ```promql
   gorax_db_connections_in_use / gorax_db_connections_open
   ```

2. **Check wait times** (if instrumented):
   ```promql
   rate(gorax_db_connection_wait_duration_seconds_sum[5m]) /
   rate(gorax_db_connection_wait_duration_seconds_count[5m])
   ```

3. **Actions**:
   - Increase `MaxOpenConns` if consistently high
   - Review slow queries
   - Check for connection leaks

### Slow Database Queries

1. **Find slowest queries**:
   ```promql
   topk(10, histogram_quantile(0.99,
     rate(gorax_db_query_duration_seconds_bucket[5m])
   ))
   ```

2. **Check query rate by table**:
   ```promql
   sum by (table, operation) (
     rate(gorax_db_queries_total[5m])
   )
   ```

3. **Actions**:
   - Add database indexes
   - Optimize query patterns
   - Consider query caching
   - Review N+1 query issues

## Best Practices

1. **Set Up Alerts**: Configure Prometheus alerts for critical thresholds
2. **Monitor Trends**: Watch for gradual performance degradation over time
3. **Baseline Performance**: Establish performance baselines after optimization
4. **Regular Reviews**: Review dashboards during incident retrospectives
5. **Correlate Metrics**: Look for correlations between different metrics (e.g., high queue depth + slow workflows)
6. **Tag by Environment**: Use labels to separate dev/staging/production metrics

## Support

For issues or questions about metrics and dashboards:
- Check [TROUBLESHOOTING.md](../docs/TROUBLESHOOTING.md) for performance issues
- Review [PERFORMANCE_BASELINE.md](../docs/PERFORMANCE_BASELINE.md) for expected metrics
- Open an issue in the GitHub repository
