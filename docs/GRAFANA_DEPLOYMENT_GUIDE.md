# Grafana Performance Monitoring Deployment Guide

This guide provides step-by-step instructions for deploying the Gorax Grafana performance monitoring dashboard with Prometheus.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Quick Start (Docker Compose)](#quick-start-docker-compose)
- [Manual Installation](#manual-installation)
- [Configuration](#configuration)
- [Dashboard Import](#dashboard-import)
- [Alerting Setup](#alerting-setup)
- [Verification](#verification)
- [Troubleshooting](#troubleshooting)

---

## Prerequisites

### Required Software

| Component | Minimum Version | Recommended Version |
|-----------|----------------|---------------------|
| Prometheus | 2.40+ | 2.48+ |
| Grafana | 8.0+ | 10.2+ |
| Docker | 20.10+ | 24.0+ (if using Docker) |
| Docker Compose | 2.0+ | 2.23+ (if using Docker) |

### Required Ports

| Service | Port | Purpose |
|---------|------|---------|
| Prometheus | 9090 | Metrics collection and querying |
| Grafana | 3000 | Dashboard visualization |
| Gorax API | 9091 | Metrics endpoint (`/metrics`) |

### Gorax Configuration

Ensure your Gorax API server exposes Prometheus metrics on `/metrics` endpoint:

```go
// Already configured in internal/api/app.go
router.GET("/metrics", gin.WrapH(promhttp.Handler()))
```

---

## Quick Start (Docker Compose)

The fastest way to get monitoring running locally.

### 1. Start the Monitoring Stack

```bash
# From the gorax project root
docker-compose -f docker-compose.monitoring.yml up -d
```

This starts:
- Prometheus (http://localhost:9090)
- Grafana (http://localhost:3000)
- Automatically configures data sources
- Provisions dashboards

### 2. Start Gorax with Metrics Enabled

```bash
# Ensure metrics are enabled in .env
export ENABLE_METRICS=true
export METRICS_PORT=9091

# Start the API server
make run-api-dev
```

### 3. Access Grafana

```
URL: http://localhost:3000
Default credentials:
  Username: admin
  Password: admin
```

On first login, you'll be prompted to change the password.

### 4. View the Dashboard

Navigate to: **Dashboards** → **Gorax** folder → **Gorax Performance Overview**

---

## Manual Installation

For production environments or custom setups.

### Step 1: Install Prometheus

#### Using Docker

```bash
docker run -d \
  --name prometheus \
  -p 9090:9090 \
  -v $(pwd)/configs/prometheus.yml:/etc/prometheus/prometheus.yml \
  -v $(pwd)/configs/prometheus-alerts.yml:/etc/prometheus/alerts.yml \
  -v prometheus-data:/prometheus \
  prom/prometheus:latest \
  --config.file=/etc/prometheus/prometheus.yml \
  --storage.tsdb.path=/prometheus \
  --storage.tsdb.retention.time=30d
```

#### Using Package Manager (Linux)

```bash
# Debian/Ubuntu
sudo apt-get update
sudo apt-get install prometheus

# RHEL/CentOS
sudo yum install prometheus

# Copy configuration
sudo cp configs/prometheus.yml /etc/prometheus/prometheus.yml
sudo cp configs/prometheus-alerts.yml /etc/prometheus/alerts.yml

# Restart service
sudo systemctl restart prometheus
sudo systemctl enable prometheus
```

#### Using Homebrew (macOS)

```bash
brew install prometheus

# Copy configuration
cp configs/prometheus.yml /usr/local/etc/prometheus.yml

# Start service
brew services start prometheus
```

### Step 2: Install Grafana

#### Using Docker

```bash
docker run -d \
  --name grafana \
  -p 3000:3000 \
  -v grafana-data:/var/lib/grafana \
  -v $(pwd)/configs/grafana/provisioning:/etc/grafana/provisioning \
  -e "GF_SECURITY_ADMIN_PASSWORD=admin" \
  grafana/grafana:latest
```

#### Using Package Manager (Linux)

```bash
# Debian/Ubuntu
sudo apt-get install -y software-properties-common
sudo add-apt-repository "deb https://packages.grafana.com/oss/deb stable main"
wget -q -O - https://packages.grafana.com/gpg.key | sudo apt-key add -
sudo apt-get update
sudo apt-get install grafana

# RHEL/CentOS
sudo yum install grafana

# Start service
sudo systemctl start grafana-server
sudo systemctl enable grafana-server
```

#### Using Homebrew (macOS)

```bash
brew install grafana

# Start service
brew services start grafana
```

### Step 3: Configure Prometheus Data Source in Grafana

#### Option A: Automatic Provisioning (Recommended)

1. Copy the provisioning configuration:
   ```bash
   sudo mkdir -p /etc/grafana/provisioning/datasources
   sudo cp configs/grafana/provisioning/datasources/prometheus.yml \
      /etc/grafana/provisioning/datasources/
   ```

2. Restart Grafana:
   ```bash
   sudo systemctl restart grafana-server
   ```

#### Option B: Manual Configuration via UI

1. Login to Grafana at http://localhost:3000
2. Navigate to **Configuration** → **Data Sources**
3. Click **Add data source**
4. Select **Prometheus**
5. Configure:
   - **Name**: `Prometheus`
   - **URL**: `http://localhost:9090` (or your Prometheus URL)
   - **Access**: `Server (default)`
6. Click **Save & Test**

---

## Configuration

### Prometheus Configuration

The `configs/prometheus.yml` file contains the scrape configuration for Gorax metrics.

**Key Configuration Points:**

```yaml
# Scrape interval (how often Prometheus collects metrics)
scrape_interval: 15s

# Evaluation interval (how often Prometheus evaluates rules)
evaluation_interval: 15s

# Scrape Gorax API metrics
scrape_configs:
  - job_name: 'gorax-api'
    static_configs:
      - targets: ['localhost:9091']
```

**For Production:**

Replace `localhost:9091` with your actual Gorax API server address:

```yaml
scrape_configs:
  - job_name: 'gorax-api'
    static_configs:
      - targets: ['gorax-api-1.example.com:9091', 'gorax-api-2.example.com:9091']
    relabel_configs:
      - source_labels: [__address__]
        target_label: instance
```

### Environment-Specific Configuration

**Development:**
```yaml
scrape_configs:
  - job_name: 'gorax-api'
    static_configs:
      - targets: ['localhost:9091']
        labels:
          environment: 'dev'
```

**Staging:**
```yaml
scrape_configs:
  - job_name: 'gorax-api'
    static_configs:
      - targets: ['staging-api.gorax.example.com:9091']
        labels:
          environment: 'staging'
```

**Production:**
```yaml
scrape_configs:
  - job_name: 'gorax-api'
    static_configs:
      - targets:
          - 'prod-api-1.gorax.example.com:9091'
          - 'prod-api-2.gorax.example.com:9091'
          - 'prod-api-3.gorax.example.com:9091'
        labels:
          environment: 'production'
    relabel_configs:
      - source_labels: [__address__]
        target_label: instance
```

---

## Dashboard Import

### Method 1: Automatic Provisioning (Recommended for Production)

1. Copy dashboard JSON to provisioning directory:
   ```bash
   sudo mkdir -p /etc/grafana/provisioning/dashboards
   sudo cp dashboards/gorax-performance-overview.json \
      /etc/grafana/provisioning/dashboards/
   sudo cp configs/grafana/provisioning/dashboards/gorax.yml \
      /etc/grafana/provisioning/dashboards/
   ```

2. Restart Grafana:
   ```bash
   sudo systemctl restart grafana-server
   ```

3. Dashboard will appear automatically in the **Gorax** folder

### Method 2: Manual Import via Grafana UI

1. Login to Grafana (http://localhost:3000)

2. Navigate to **Dashboards** → **Import**

3. Click **Upload JSON file**

4. Select `dashboards/gorax-performance-overview.json`

5. Configure import options:
   - **Name**: Gorax Performance Overview (default)
   - **Folder**: Create new folder "Gorax"
   - **UID**: gorax-performance-overview (default)
   - **Prometheus**: Select your Prometheus data source

6. Click **Import**

### Method 3: Import by JSON Content

1. Navigate to **Dashboards** → **Import**

2. Copy the entire contents of `dashboards/gorax-performance-overview.json`

3. Paste into the **Import via panel json** text area

4. Click **Load**

5. Configure data source and click **Import**

---

## Alerting Setup

### Step 1: Configure Alert Rules in Prometheus

1. Alert rules are defined in `configs/prometheus-alerts.yml`

2. Ensure the rules file is loaded in `prometheus.yml`:
   ```yaml
   rule_files:
     - "alerts.yml"
   ```

3. Restart Prometheus:
   ```bash
   sudo systemctl restart prometheus
   # OR for Docker:
   docker restart prometheus
   ```

4. Verify rules are loaded:
   ```bash
   # Check Prometheus UI
   open http://localhost:9090/alerts
   ```

### Step 2: Configure Alertmanager (Optional but Recommended)

Alertmanager handles alert routing and notifications.

**Install Alertmanager:**

```bash
# Docker
docker run -d \
  --name alertmanager \
  -p 9093:9093 \
  -v $(pwd)/configs/alertmanager.yml:/etc/alertmanager/alertmanager.yml \
  prom/alertmanager:latest
```

**Configure Prometheus to use Alertmanager:**

Add to `prometheus.yml`:

```yaml
alerting:
  alertmanagers:
    - static_configs:
        - targets: ['localhost:9093']
```

### Step 3: Configure Notification Channels

Example Alertmanager configuration for Slack notifications:

```yaml
# configs/alertmanager.yml
global:
  slack_api_url: 'https://hooks.slack.com/services/YOUR/WEBHOOK/URL'

route:
  group_by: ['alertname', 'severity']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 12h
  receiver: 'slack-notifications'

receivers:
  - name: 'slack-notifications'
    slack_configs:
      - channel: '#alerts'
        title: '{{ .GroupLabels.alertname }}'
        text: '{{ range .Alerts }}{{ .Annotations.description }}{{ end }}'
```

---

## Verification

### 1. Verify Metrics are Being Collected

**Check Gorax metrics endpoint:**
```bash
curl http://localhost:9091/metrics
```

Expected output should include:
```
# HELP gorax_workflow_executions_total Total number of workflow executions
# TYPE gorax_workflow_executions_total counter
gorax_workflow_executions_total{status="completed",trigger_type="manual"} 42
...
```

**Check Prometheus is scraping:**
```bash
# Open Prometheus UI
open http://localhost:9090/targets

# Or use curl
curl http://localhost:9090/api/v1/targets | jq '.data.activeTargets[] | select(.labels.job=="gorax-api")'
```

Expected status: **UP**

### 2. Verify Prometheus is Collecting Data

**Query a simple metric:**
```bash
curl -G http://localhost:9090/api/v1/query \
  --data-urlencode 'query=up{job="gorax-api"}' | jq
```

Expected: `"value": [<timestamp>, "1"]`

### 3. Verify Dashboard is Working

1. Open Grafana dashboard
2. Select time range: **Last 1 hour**
3. Verify panels are showing data:
   - Workflow Execution Rate should show data if workflows are running
   - Formula Cache Hit Rate should show percentage
   - Database Connection Pool should show connection counts
   - HTTP Request Rate should show API request metrics

**If panels show "No Data":**
- Check time range (try "Last 24 hours")
- Verify Gorax API is running and serving metrics
- Check Prometheus is successfully scraping
- Review browser console for errors

### 4. Verify Alerts are Active

**Check Prometheus alerts:**
```bash
curl http://localhost:9090/api/v1/rules | jq '.data.groups[].rules[] | select(.type=="alerting")'
```

**Trigger a test alert:**
```bash
# This query should return 1 if an alert is firing
curl -G http://localhost:9090/api/v1/query \
  --data-urlencode 'query=ALERTS{alertstate="firing"}' | jq
```

---

## Troubleshooting

### Dashboard Shows "No Data"

**Problem**: All panels are empty or show "No data"

**Diagnosis:**

1. Check Prometheus target status:
   ```bash
   curl http://localhost:9090/api/v1/targets | jq '.data.activeTargets[] | select(.labels.job=="gorax-api")'
   ```

2. Check if Gorax metrics endpoint is reachable:
   ```bash
   curl http://localhost:9091/metrics
   ```

3. Check Prometheus logs:
   ```bash
   # Docker
   docker logs prometheus

   # Systemd
   sudo journalctl -u prometheus -f
   ```

**Solutions:**

- Ensure Gorax API is running: `make run-api-dev`
- Verify metrics are enabled in `.env`: `ENABLE_METRICS=true`
- Check firewall rules: `sudo ufw status`
- Verify Prometheus scrape config: `configs/prometheus.yml`
- Check time range in Grafana (try "Last 24 hours")

### Connection Refused to Prometheus

**Problem**: Grafana cannot connect to Prometheus data source

**Diagnosis:**

1. Verify Prometheus is running:
   ```bash
   curl http://localhost:9090/-/healthy
   ```

2. Check if port is open:
   ```bash
   netstat -tuln | grep 9090
   ```

**Solutions:**

- Start Prometheus: `docker start prometheus` or `sudo systemctl start prometheus`
- Check Prometheus configuration: `promtool check config configs/prometheus.yml`
- Verify network connectivity: `telnet localhost 9090`
- For Docker: Ensure containers are on same network

### Metrics Not Updating

**Problem**: Dashboard shows old data or metrics are stale

**Diagnosis:**

1. Check Prometheus scrape interval:
   ```bash
   curl http://localhost:9090/api/v1/status/config | jq '.data.yaml' | grep scrape_interval
   ```

2. Check last scrape time:
   ```bash
   curl http://localhost:9090/api/v1/targets | jq '.data.activeTargets[] | select(.labels.job=="gorax-api") | .lastScrape'
   ```

**Solutions:**

- Restart Prometheus to reload configuration
- Verify Gorax API is receiving traffic (generate some test workflows)
- Check system time synchronization: `timedatectl status`
- Increase scrape interval if needed (reduce from 15s to 5s for development)

### Alerts Not Firing

**Problem**: Expected alerts are not triggering

**Diagnosis:**

1. Check alert rules syntax:
   ```bash
   promtool check rules configs/prometheus-alerts.yml
   ```

2. Verify rules are loaded:
   ```bash
   curl http://localhost:9090/api/v1/rules | jq '.data.groups[].rules[] | select(.type=="alerting") | .name'
   ```

3. Check alert evaluation:
   ```bash
   curl http://localhost:9090/api/v1/query \
     --data-urlencode 'query=gorax_workflow_executions_total' | jq
   ```

**Solutions:**

- Fix syntax errors in alert rules
- Restart Prometheus after rule changes
- Verify metrics exist: Query in Prometheus UI
- Check alert duration (`for: 5m` means alert must be true for 5 minutes)
- Lower thresholds temporarily for testing

### Dashboard Import Fails

**Problem**: Cannot import dashboard JSON

**Solutions:**

- Validate JSON syntax: `jq . dashboards/gorax-performance-overview.json`
- Try importing via file upload instead of paste
- Check Grafana logs: `docker logs grafana` or `sudo journalctl -u grafana-server`
- Update Grafana to latest version (compatibility issues)
- Remove `"id": null` field from JSON if present

### Permission Denied Errors

**Problem**: Cannot access metrics or configuration files

**Solutions:**

```bash
# Fix Prometheus config permissions
sudo chown prometheus:prometheus /etc/prometheus/prometheus.yml
sudo chmod 644 /etc/prometheus/prometheus.yml

# Fix Grafana provisioning permissions
sudo chown -R grafana:grafana /etc/grafana/provisioning
sudo chmod -R 755 /etc/grafana/provisioning

# Fix Docker volume permissions
docker exec -u root grafana chown -R grafana:grafana /var/lib/grafana
```

### High Memory Usage

**Problem**: Prometheus consuming too much memory

**Solutions:**

1. Reduce retention time:
   ```yaml
   # In prometheus.yml or command line
   --storage.tsdb.retention.time=15d
   ```

2. Reduce cardinality:
   ```yaml
   # Drop high-cardinality labels
   metric_relabel_configs:
     - source_labels: [__name__]
       regex: 'gorax_db_query_duration_seconds_bucket'
       target_label: query_id
       replacement: ''
   ```

3. Increase memory limits (Docker):
   ```yaml
   deploy:
     resources:
       limits:
         memory: 4G
   ```

---

## Security Best Practices

### 1. Secure Grafana

```ini
# /etc/grafana/grafana.ini

[security]
# Change default admin password immediately
admin_user = admin
admin_password = <strong-password>

# Disable user signup
allow_sign_up = false

# Enable HTTPS
[server]
protocol = https
cert_file = /etc/grafana/ssl/cert.pem
cert_key = /etc/grafana/ssl/key.pem
```

### 2. Secure Prometheus

```yaml
# Use authentication for Prometheus (requires reverse proxy)
# Example with Nginx:
# /etc/nginx/sites-available/prometheus

server {
    listen 9090 ssl;
    server_name prometheus.example.com;

    ssl_certificate /etc/ssl/certs/prometheus.crt;
    ssl_certificate_key /etc/ssl/private/prometheus.key;

    auth_basic "Prometheus";
    auth_basic_user_file /etc/nginx/.htpasswd;

    location / {
        proxy_pass http://localhost:9091;
    }
}
```

### 3. Network Isolation

```yaml
# docker-compose.monitoring.yml
networks:
  monitoring:
    driver: bridge
    internal: true  # Isolate from external access

  frontend:
    driver: bridge

services:
  prometheus:
    networks:
      - monitoring

  grafana:
    networks:
      - monitoring
      - frontend  # Only Grafana exposed
```

### 4. Restrict Metrics Endpoint

```go
// Protect /metrics endpoint with authentication
router.GET("/metrics", middleware.RequireAuth(), gin.WrapH(promhttp.Handler()))
```

---

## Production Considerations

### High Availability

**Prometheus HA:**
```yaml
# Run multiple Prometheus instances
# Use Thanos or Cortex for long-term storage and federation
```

**Grafana HA:**
```yaml
# Use external database (PostgreSQL)
[database]
type = postgres
host = postgres.example.com:5432
name = grafana
user = grafana
password = ${GRAFANA_DB_PASSWORD}

# Session storage in Redis
[session]
provider = redis
provider_config = addr=redis.example.com:6379,pool_size=100,db=grafana
```

### Backup and Restore

**Backup Prometheus data:**
```bash
# Create snapshot
curl -XPOST http://localhost:9090/api/v1/admin/tsdb/snapshot

# Backup TSDB directory
tar czf prometheus-backup-$(date +%Y%m%d).tar.gz /var/lib/prometheus/
```

**Backup Grafana dashboards:**
```bash
# Export all dashboards
for dash in $(curl -s http://admin:admin@localhost:3000/api/search?query=\& | jq -r '.[].uid'); do
  curl -s http://admin:admin@localhost:3000/api/dashboards/uid/$dash | jq . > dashboard-$dash.json
done
```

---

## Additional Resources

- [Prometheus Documentation](https://prometheus.io/docs/introduction/overview/)
- [Grafana Documentation](https://grafana.com/docs/grafana/latest/)
- [PromQL Query Examples](https://prometheus.io/docs/prometheus/latest/querying/examples/)
- [Gorax Performance Baseline](./PERFORMANCE_BASELINE.md)
- [Gorax Troubleshooting Guide](./TROUBLESHOOTING.md)

---

## Support

For issues related to monitoring setup:

1. Check [dashboards/README.md](../dashboards/README.md) for common queries
2. Review this troubleshooting section
3. Check Prometheus and Grafana logs
4. Open an issue on GitHub with:
   - Prometheus version
   - Grafana version
   - Error messages from logs
   - Screenshots of issues

---

**Next Steps:**
- Configure alerting channels (Slack, PagerDuty, email)
- Set up long-term metrics storage (Thanos, Cortex)
- Create custom dashboards for specific workflows
- Set up automated dashboard backups
- Implement metrics-based autoscaling
