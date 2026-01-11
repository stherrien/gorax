# Gorax Configuration Files

This directory contains configuration files for monitoring and observability infrastructure.

## Contents

### Prometheus Configuration

- **`prometheus.yml`** - Main Prometheus configuration
  - Scrape configurations for Gorax API metrics
  - Global settings (scrape interval, timeouts)
  - Alert rule file references
  - Environment-specific examples

- **`prometheus-alerts.yml`** - Alert rules
  - Workflow performance alerts
  - Formula evaluation alerts
  - Database performance alerts
  - API performance alerts
  - Service health checks

### Grafana Configuration

- **`grafana/provisioning/datasources/prometheus.yml`** - Prometheus data source
  - Auto-provisioned Prometheus connection
  - Default data source configuration

- **`grafana/provisioning/dashboards/gorax.yml`** - Dashboard provisioning
  - Auto-loads dashboards from `dashboards/` directory
  - Organizes into "Gorax" folder

### Alertmanager Configuration (Optional)

- **`alertmanager.yml`** - Alert routing and notifications
  - Slack integration examples
  - PagerDuty integration examples
  - Email notification examples
  - Alert grouping and routing rules
  - Inhibition rules

## Quick Start

### Local Development

```bash
# Start monitoring stack with these configs
docker-compose -f docker-compose.monitoring.yml up -d

# Verify Prometheus configuration
docker exec gorax-prometheus promtool check config /etc/prometheus/prometheus.yml

# Verify alert rules
docker exec gorax-prometheus promtool check rules /etc/prometheus/alerts.yml
```

### Production Deployment

1. **Copy configuration files to server:**
   ```bash
   # Prometheus
   sudo cp prometheus.yml /etc/prometheus/
   sudo cp prometheus-alerts.yml /etc/prometheus/alerts.yml

   # Grafana
   sudo cp -r grafana/provisioning /etc/grafana/
   ```

2. **Modify for your environment:**
   - Update scrape targets in `prometheus.yml`
   - Set Slack webhook URLs in `alertmanager.yml`
   - Configure retention settings
   - Enable TLS if needed

3. **Restart services:**
   ```bash
   sudo systemctl restart prometheus
   sudo systemctl restart grafana-server
   ```

## Configuration Details

### Prometheus (`prometheus.yml`)

**Key Settings:**

- **Scrape Interval:** 15s (how often to collect metrics)
- **Evaluation Interval:** 15s (how often to evaluate alert rules)
- **Retention:** 30d (how long to keep metrics data)

**Target Configuration:**

For different environments, update the `static_configs` section:

```yaml
# Development (localhost)
- targets: ['localhost:9091']

# Docker Compose
- targets: ['gorax-api:9091']

# Docker Desktop (Mac/Windows)
- targets: ['host.docker.internal:9091']

# Production (multiple instances)
- targets:
    - 'gorax-api-1.example.com:9091'
    - 'gorax-api-2.example.com:9091'
    - 'gorax-api-3.example.com:9091'
```

### Alert Rules (`prometheus-alerts.yml`)

**Alert Groups:**

1. **gorax_workflow_performance** - Workflow execution metrics
2. **gorax_formula_performance** - Formula evaluation metrics
3. **gorax_database_performance** - Database query metrics
4. **gorax_api_performance** - HTTP API metrics
5. **gorax_service_health** - Service availability

**Severity Levels:**

- `critical` - Immediate action required (page on-call)
- `warning` - Investigate during business hours

**Customizing Thresholds:**

Edit the `expr` field to adjust thresholds:

```yaml
# Example: Change high error rate from 10% to 15%
- alert: HighWorkflowFailureRate
  expr: |
    (
      sum(rate(gorax_workflow_executions_total{status="failed"}[5m]))
      /
      sum(rate(gorax_workflow_executions_total[5m]))
    ) > 0.15  # Changed from 0.1 (10%) to 0.15 (15%)
```

### Alertmanager (`alertmanager.yml`)

**Before using:**

1. Replace placeholder webhook URLs:
   - Slack: `https://hooks.slack.com/services/YOUR/WEBHOOK/URL`
   - PagerDuty: `YOUR_PAGERDUTY_INTEGRATION_KEY`

2. Configure receivers for your notification channels

3. Adjust routing rules based on your team structure

**Testing Alerts:**

```bash
# Send test alert to Alertmanager
curl -XPOST http://localhost:9093/api/v1/alerts \
  -H 'Content-Type: application/json' \
  -d '[{
    "labels": {
      "alertname": "TestAlert",
      "severity": "warning"
    },
    "annotations": {
      "summary": "This is a test alert"
    }
  }]'
```

## Environment Variables

### Prometheus

No environment variables required for basic setup.

### Grafana

Set in `docker-compose.monitoring.yml` or `/etc/grafana/grafana.ini`:

```yaml
GF_SECURITY_ADMIN_USER=admin
GF_SECURITY_ADMIN_PASSWORD=<strong-password>
GF_SERVER_ROOT_URL=http://localhost:3000
```

### Alertmanager

For sensitive values (Slack tokens, API keys):

```yaml
# Use environment variable substitution
slack_configs:
  - api_url: ${SLACK_WEBHOOK_URL}
```

Then pass environment variables:

```bash
export SLACK_WEBHOOK_URL='https://hooks.slack.com/services/YOUR/WEBHOOK/URL'
```

## Validation

### Validate Prometheus Config

```bash
# Using promtool directly
promtool check config configs/prometheus.yml

# Using Docker
docker run --rm -v $(pwd)/configs:/configs prom/prometheus:latest \
  promtool check config /configs/prometheus.yml
```

### Validate Alert Rules

```bash
# Using promtool
promtool check rules configs/prometheus-alerts.yml

# Using Docker
docker run --rm -v $(pwd)/configs:/configs prom/prometheus:latest \
  promtool check rules /configs/prometheus-alerts.yml
```

### Validate Alertmanager Config

```bash
# Using amtool
amtool check-config configs/alertmanager.yml

# Using Docker
docker run --rm -v $(pwd)/configs:/configs prom/alertmanager:latest \
  amtool check-config /configs/alertmanager.yml
```

## Security Considerations

### 1. Protect Configuration Files

```bash
# Restrict file permissions
chmod 600 configs/alertmanager.yml  # Contains sensitive webhook URLs
chmod 644 configs/prometheus.yml
chmod 644 configs/prometheus-alerts.yml
```

### 2. Use Secrets Management

For production, use secrets management:

```yaml
# Kubernetes
apiVersion: v1
kind: Secret
metadata:
  name: alertmanager-config
type: Opaque
stringData:
  alertmanager.yml: |
    global:
      slack_api_url: ${SLACK_WEBHOOK_URL}
```

### 3. Enable Authentication

Add authentication to Prometheus and Grafana in production. See [GRAFANA_DEPLOYMENT_GUIDE.md](../docs/GRAFANA_DEPLOYMENT_GUIDE.md#security-best-practices).

## Troubleshooting

### Prometheus Won't Start

**Check configuration syntax:**
```bash
promtool check config configs/prometheus.yml
```

**Check logs:**
```bash
docker logs gorax-prometheus
```

**Common issues:**
- Invalid YAML syntax
- Missing alert rules file
- Invalid target addresses

### Alerts Not Firing

**Verify alert rules are loaded:**
```bash
curl http://localhost:9090/api/v1/rules | jq
```

**Check alert evaluation:**
```bash
# See pending/firing alerts
curl http://localhost:9090/api/v1/alerts | jq
```

**Test alert expression:**
```bash
# Query Prometheus directly
curl -G http://localhost:9090/api/v1/query \
  --data-urlencode 'query=gorax_workflow_executions_total'
```

### Grafana Can't Connect to Prometheus

**Check data source configuration:**
- URL should be `http://prometheus:9090` (Docker) or `http://localhost:9090` (local)
- Access mode should be "Server" (default)

**Test connection from Grafana container:**
```bash
docker exec gorax-grafana wget -O- http://prometheus:9090/-/healthy
```

## Documentation

- [Monitoring Guide](../docs/MONITORING.md) - Complete monitoring setup guide
- [Grafana Deployment Guide](../docs/GRAFANA_DEPLOYMENT_GUIDE.md) - Detailed Grafana setup
- [Dashboard README](../dashboards/README.md) - Dashboard usage and queries
- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)

## Contributing

When modifying configurations:

1. Validate syntax before committing
2. Test changes in development environment
3. Update this README if adding new files
4. Document any new alert rules or thresholds
5. Add comments explaining non-obvious configuration choices
