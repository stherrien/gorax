# Monitoring Quick Start

Get Gorax monitoring running in 5 minutes.

## Prerequisites

- Docker and Docker Compose installed
- Gorax development environment set up

## Steps

### 1. Start Monitoring Stack

```bash
# Start Prometheus and Grafana
docker-compose -f docker-compose.monitoring.yml up -d

# Verify services started
docker-compose -f docker-compose.monitoring.yml ps
```

Expected output:
```
NAME                STATUS    PORTS
gorax-prometheus    healthy   0.0.0.0:9090->9090/tcp
gorax-grafana       healthy   0.0.0.0:3000->3000/tcp
```

### 2. Start Gorax with Metrics

```bash
# Start database dependencies
docker-compose -f docker-compose.dev.yml up -d

# Enable metrics in environment
echo "ENABLE_METRICS=true" >> .env
echo "METRICS_PORT=9091" >> .env

# Start API server
make run-api-dev
```

### 3. Verify Setup

```bash
# Run verification script
./scripts/verify-monitoring.sh

# Or manually check:
# - Metrics endpoint: curl http://localhost:9091/metrics
# - Prometheus targets: http://localhost:9090/targets
# - Grafana: http://localhost:3000 (admin/admin)
```

### 4. View Dashboard

1. Open Grafana: http://localhost:3000
2. Login with `admin` / `admin` (change password on first login)
3. Navigate to **Dashboards** → **Gorax** folder → **Gorax Performance Overview**

## Troubleshooting

### No Data in Dashboard?

```bash
# Check if Gorax metrics are exposed
curl http://localhost:9091/metrics | grep gorax

# Check if Prometheus is scraping
curl http://localhost:9090/api/v1/targets | jq '.data.activeTargets[] | select(.labels.job=="gorax-api")'

# Check Prometheus logs
docker logs gorax-prometheus
```

### Grafana Can't Connect?

```bash
# Verify containers are on same network
docker network inspect gorax-monitoring

# Test connectivity from Grafana
docker exec gorax-grafana wget -O- http://prometheus:9090/-/healthy
```

## Full Documentation

- [Complete Monitoring Guide](../docs/MONITORING.md)
- [Grafana Deployment Guide](../docs/GRAFANA_DEPLOYMENT_GUIDE.md)
- [Dashboard README](../dashboards/README.md)

## Quick Commands

```bash
# View all running monitoring services
docker-compose -f docker-compose.monitoring.yml ps

# View logs
docker-compose -f docker-compose.monitoring.yml logs -f

# Restart Prometheus (after config change)
docker-compose -f docker-compose.monitoring.yml restart prometheus

# Stop monitoring stack
docker-compose -f docker-compose.monitoring.yml down

# Remove all data (WARNING: destructive)
docker-compose -f docker-compose.monitoring.yml down -v
```
