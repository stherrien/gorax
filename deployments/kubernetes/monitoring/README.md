# Gorax Kubernetes Monitoring Setup

This directory contains Kubernetes configurations for deploying the Gorax monitoring stack.

## Components

### 1. Prometheus
- **Purpose:** Metrics collection and storage
- **Port:** 9090
- **Configuration:** `prometheus-config.yaml`

### 2. Alertmanager
- **Purpose:** Alert routing and notification
- **Port:** 9093
- **Configuration:** Defined in Prometheus config

### 3. Grafana
- **Purpose:** Metrics visualization
- **Port:** 3000
- **Dashboards:** Pre-configured in `dashboards/` directory

### 4. Alert Rules
- **File:** `alerts.yaml`
- **Count:** 15 production-ready alerts
- **Categories:** Workflows, Queue, API, Infrastructure, Steps

## Quick Start

### Prerequisites

```bash
# Ensure you have kubectl configured
kubectl config current-context

# Create monitoring namespace
kubectl create namespace monitoring
```

### Deploy Monitoring Stack

```bash
# Deploy Prometheus configuration
kubectl apply -f prometheus-config.yaml -n monitoring

# Deploy alerting rules
kubectl apply -f alerts.yaml -n monitoring

# Deploy Prometheus (example using Helm)
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm install prometheus prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --set prometheus.prometheusSpec.additionalScrapeConfigs[0].job_name=gorax-api \
  --values prometheus-values.yaml

# Import Grafana dashboards
kubectl create configmap gorax-dashboards \
  --from-file=dashboards/ \
  --namespace monitoring
```

### Configure Application

Update your Gorax deployment to expose metrics:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: gorax-api-metrics
  namespace: gorax
  labels:
    app: gorax-api
    metrics: "true"
spec:
  selector:
    app: gorax-api
  ports:
    - name: metrics
      port: 9090
      targetPort: 9090
  type: ClusterIP
```

### Access Services

```bash
# Port-forward Prometheus
kubectl port-forward -n monitoring svc/prometheus-server 9090:9090

# Port-forward Grafana
kubectl port-forward -n monitoring svc/grafana 3000:3000

# Port-forward Alertmanager
kubectl port-forward -n monitoring svc/alertmanager 9093:9093
```

Then access:
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (default: admin/admin)
- Alertmanager: http://localhost:9093

## Configuration Files

### prometheus-config.yaml

Contains Prometheus server configuration including:
- Global settings (scrape interval, evaluation interval)
- Alertmanager integration
- Service discovery for Kubernetes pods
- Scrape configurations for Gorax API and Worker
- External exporters (PostgreSQL, Redis)

**Key Configuration:**
```yaml
scrape_interval: 15s        # How often to scrape metrics
evaluation_interval: 15s    # How often to evaluate rules
external_labels:
  cluster: 'gorax-production'
  environment: 'production'
```

### alerts.yaml

Contains 15 production-ready alerting rules:

**Severity Levels:**
- `critical` - Immediate action required (page on-call)
- `warning` - Investigation needed (notify team)
- `info` - Awareness only (log/dashboard)

**Alert Groups:**
1. `gorax_workflows` - Workflow execution health
2. `gorax_queue` - Queue and worker health
3. `gorax_api` - API performance and errors
4. `gorax_infrastructure` - Database, Redis, system health
5. `gorax_step_failures` - Step-level execution issues

## Dashboards

### gorax-overview.json

Comprehensive system overview dashboard with 7 panels:

1. **Workflow Execution Rate** - Real-time execution trends
2. **Workflow Success Rate** - System health indicator
3. **Active Workers** - Worker pool utilization
4. **HTTP Request Rate** - API traffic patterns
5. **Queue Depth** - Backlog monitoring
6. **Workflow Duration Percentiles** - Performance metrics
7. **Top Workflows** - Most active workflows

**Import Instructions:**
1. Open Grafana UI
2. Navigate to Dashboards → Import
3. Upload `dashboards/gorax-overview.json`
4. Select Prometheus datasource
5. Click Import

### Additional Dashboards (To Be Created)

- `gorax-workflows.json` - Detailed workflow metrics
- `gorax-api.json` - API performance deep-dive
- `gorax-workers.json` - Worker pool metrics

## Alerting Setup

### Configure Alertmanager

Create `alertmanager-config.yaml`:

```yaml
global:
  resolve_timeout: 5m

route:
  group_by: ['alertname', 'cluster', 'service']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 12h
  receiver: 'default'
  routes:
    - match:
        severity: critical
      receiver: 'pagerduty'
    - match:
        severity: warning
      receiver: 'slack'

receivers:
  - name: 'default'
    webhook_configs:
      - url: 'http://localhost:5001/'

  - name: 'pagerduty'
    pagerduty_configs:
      - service_key: '<your-pagerduty-key>'

  - name: 'slack'
    slack_configs:
      - api_url: '<your-slack-webhook>'
        channel: '#alerts'
        title: 'Gorax Alert'
        text: '{{ range .Alerts }}{{ .Annotations.description }}{{ end }}'
```

Apply configuration:

```bash
kubectl create secret generic alertmanager-config \
  --from-file=alertmanager.yml=alertmanager-config.yaml \
  --namespace monitoring
```

## Monitoring Best Practices

### 1. Alert Tuning

**Start Conservative:**
- Enable only critical alerts initially
- Monitor false positive rate
- Adjust thresholds based on actual behavior

**Avoid Alert Fatigue:**
- Group related alerts
- Set appropriate repeat intervals
- Use severity levels correctly

### 2. Dashboard Organization

**Create Role-Based Dashboards:**
- Ops: System health and infrastructure
- Dev: Application performance and errors
- Business: Workflow execution and throughput

**Use Variables:**
- Add tenant_id, workflow_id filters
- Enable time range selection
- Support environment switching

### 3. Retention and Storage

**Prometheus Retention:**
```yaml
prometheus:
  retention: 15d  # Keep 15 days of detailed metrics
  retentionSize: 50GB
```

**Long-term Storage:**
- Consider Thanos or Cortex for long-term metrics
- Use downsampling for historical data
- Archive to S3 for compliance

### 4. High Availability

**Prometheus HA:**
```bash
# Deploy multiple Prometheus replicas
helm install prometheus prometheus-community/kube-prometheus-stack \
  --set prometheus.prometheusSpec.replicas=2
```

**Alertmanager HA:**
```bash
# Deploy Alertmanager cluster
helm install alertmanager prometheus-community/alertmanager \
  --set replicaCount=3
```

## Troubleshooting

### Metrics Not Appearing

1. **Check pod metrics endpoint:**
   ```bash
   kubectl port-forward -n gorax pod/<pod-name> 9090:9090
   curl http://localhost:9090/metrics
   ```

2. **Verify service discovery:**
   ```bash
   kubectl get servicemonitors -n monitoring
   kubectl describe servicemonitor gorax-api -n monitoring
   ```

3. **Check Prometheus targets:**
   - Open Prometheus UI
   - Navigate to Status → Targets
   - Verify gorax-api and gorax-worker are UP

### Alerts Not Firing

1. **Check rule evaluation:**
   - Open Prometheus UI
   - Navigate to Status → Rules
   - Verify rule state (Inactive/Pending/Firing)

2. **Test alert query:**
   ```promql
   # Copy alert expression from alerts.yaml
   rate(gorax_workflow_executions_total{status="failed"}[5m]) > 0.1
   ```

3. **Verify Alertmanager connection:**
   ```bash
   kubectl logs -n monitoring deployment/prometheus-server | grep alertmanager
   ```

### Dashboard Shows No Data

1. **Verify datasource:**
   - Grafana → Configuration → Data Sources
   - Test connection to Prometheus

2. **Check query syntax:**
   - Open panel edit mode
   - Run query in Prometheus UI first

3. **Verify time range:**
   - Ensure time range includes data
   - Check Prometheus retention period

## Security Considerations

### Network Policies

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: prometheus-ingress
  namespace: monitoring
spec:
  podSelector:
    matchLabels:
      app: prometheus
  policyTypes:
  - Ingress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: gorax
    ports:
    - protocol: TCP
      port: 9090
```

### RBAC

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: prometheus
rules:
- apiGroups: [""]
  resources:
  - nodes
  - nodes/proxy
  - services
  - endpoints
  - pods
  verbs: ["get", "list", "watch"]
```

### TLS/mTLS

Enable TLS for Prometheus scraping:

```yaml
tls_config:
  ca_file: /etc/prometheus/ca.crt
  cert_file: /etc/prometheus/client.crt
  key_file: /etc/prometheus/client.key
```

## Maintenance

### Backup Prometheus Data

```bash
# Create snapshot
kubectl exec -n monitoring prometheus-server-0 -- \
  promtool tsdb create-blocks-from data/ /snapshots/backup-$(date +%Y%m%d)

# Copy snapshot
kubectl cp monitoring/prometheus-server-0:/snapshots/backup-20250120 ./backup-20250120
```

### Update Alert Rules

```bash
# Edit alerts.yaml
vim alerts.yaml

# Apply changes
kubectl apply -f alerts.yaml -n monitoring

# Reload Prometheus config
kubectl exec -n monitoring prometheus-server-0 -- \
  curl -X POST http://localhost:9090/-/reload
```

### Update Dashboards

```bash
# Update dashboard JSON
vim dashboards/gorax-overview.json

# Update ConfigMap
kubectl create configmap gorax-dashboards \
  --from-file=dashboards/ \
  --namespace monitoring \
  --dry-run=client -o yaml | kubectl apply -f -
```

## Monitoring Costs

### Resource Requirements

**Prometheus:**
- CPU: 2 cores (4 cores for HA)
- Memory: 4GB (8GB for HA)
- Storage: 50GB (scales with retention)

**Grafana:**
- CPU: 1 core
- Memory: 2GB
- Storage: 10GB

**Alertmanager:**
- CPU: 500m
- Memory: 1GB
- Storage: 5GB

### Optimization Tips

1. **Reduce cardinality:**
   - Limit label values
   - Use path normalization
   - Avoid IDs in labels

2. **Adjust retention:**
   - Shorter retention for high-resolution data
   - Use recording rules for long-term queries

3. **Use recording rules:**
   ```yaml
   groups:
   - name: gorax_recordings
     interval: 30s
     rules:
     - record: gorax:workflow_success_rate:5m
       expr: |
         sum(rate(gorax_workflow_executions_total{status="completed"}[5m]))
         /
         sum(rate(gorax_workflow_executions_total[5m]))
   ```

## Support

- **Documentation:** `/docs/observability.md`
- **Implementation Guide:** `/docs/OBSERVABILITY_IMPLEMENTATION.md`
- **Issues:** GitHub Issues
- **Slack:** #gorax-monitoring

## References

- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [Alertmanager Documentation](https://prometheus.io/docs/alerting/latest/alertmanager/)
- [Kubernetes Service Discovery](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#kubernetes_sd_config)
