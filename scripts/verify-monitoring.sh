#!/bin/bash
# Gorax Monitoring Stack Verification Script
# This script verifies that the monitoring stack is properly configured and running

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print status
print_status() {
  local status=$1
  local message=$2

  if [ "$status" = "ok" ]; then
    echo -e "${GREEN}✓${NC} $message"
  elif [ "$status" = "fail" ]; then
    echo -e "${RED}✗${NC} $message"
  elif [ "$status" = "warn" ]; then
    echo -e "${YELLOW}!${NC} $message"
  else
    echo -e "  $message"
  fi
}

# Function to check if URL is reachable
check_url() {
  local url=$1
  local timeout=${2:-5}

  if curl -s --max-time "$timeout" "$url" > /dev/null 2>&1; then
    return 0
  else
    return 1
  fi
}

# Function to check if port is listening
check_port() {
  local host=$1
  local port=$2

  if command -v nc > /dev/null 2>&1; then
    if nc -z -w 2 "$host" "$port" 2>/dev/null; then
      return 0
    fi
  elif command -v telnet > /dev/null 2>&1; then
    if timeout 2 telnet "$host" "$port" 2>/dev/null | grep -q "Connected"; then
      return 0
    fi
  fi
  return 1
}

echo "=========================================="
echo "Gorax Monitoring Stack Verification"
echo "=========================================="
echo ""

# Check Docker
echo "Checking Docker..."
if command -v docker > /dev/null 2>&1; then
  print_status "ok" "Docker is installed ($(docker --version | cut -d' ' -f3 | tr -d ','))"
else
  print_status "fail" "Docker is not installed"
  exit 1
fi

if docker ps > /dev/null 2>&1; then
  print_status "ok" "Docker daemon is running"
else
  print_status "fail" "Docker daemon is not running"
  exit 1
fi

echo ""

# Check Docker Compose
echo "Checking Docker Compose..."
if command -v docker-compose > /dev/null 2>&1; then
  print_status "ok" "Docker Compose is installed ($(docker-compose --version | cut -d' ' -f4 | tr -d ','))"
elif docker compose version > /dev/null 2>&1; then
  print_status "ok" "Docker Compose (plugin) is installed ($(docker compose version | cut -d' ' -f4 | tr -d ','))"
else
  print_status "fail" "Docker Compose is not installed"
  exit 1
fi

echo ""

# Check if monitoring stack is running
echo "Checking monitoring services..."

# Prometheus
if docker ps | grep -q "gorax-prometheus"; then
  print_status "ok" "Prometheus container is running"

  # Check Prometheus health
  if check_url "http://localhost:9090/-/healthy"; then
    print_status "ok" "Prometheus is healthy"
  else
    print_status "fail" "Prometheus is not responding"
  fi

  # Check Prometheus targets
  if curl -s http://localhost:9090/api/v1/targets 2>/dev/null | grep -q "gorax-api"; then
    print_status "ok" "Prometheus is configured to scrape Gorax API"
  else
    print_status "warn" "Prometheus gorax-api target not found"
  fi
else
  print_status "fail" "Prometheus container is not running"
  echo "  Run: docker-compose -f docker-compose.monitoring.yml up -d"
fi

# Grafana
if docker ps | grep -q "gorax-grafana"; then
  print_status "ok" "Grafana container is running"

  # Check Grafana health
  if check_url "http://localhost:3000/api/health"; then
    print_status "ok" "Grafana is healthy"
  else
    print_status "fail" "Grafana is not responding"
  fi
else
  print_status "fail" "Grafana container is not running"
  echo "  Run: docker-compose -f docker-compose.monitoring.yml up -d"
fi

echo ""

# Check Gorax API
echo "Checking Gorax API..."

if check_port "localhost" "9091"; then
  print_status "ok" "Port 9091 is listening"

  # Check metrics endpoint
  if check_url "http://localhost:9091/metrics"; then
    print_status "ok" "Gorax metrics endpoint is accessible"

    # Check if metrics contain expected data
    metrics=$(curl -s http://localhost:9091/metrics 2>/dev/null)

    if echo "$metrics" | grep -q "gorax_workflow_executions_total"; then
      print_status "ok" "Workflow metrics are exposed"
    else
      print_status "warn" "Workflow metrics not found (API may be starting)"
    fi

    if echo "$metrics" | grep -q "gorax_http_requests_total"; then
      print_status "ok" "HTTP metrics are exposed"
    else
      print_status "warn" "HTTP metrics not found"
    fi

    if echo "$metrics" | grep -q "gorax_db_"; then
      print_status "ok" "Database metrics are exposed"
    else
      print_status "warn" "Database metrics not found"
    fi
  else
    print_status "fail" "Gorax metrics endpoint is not accessible"
    echo "  Make sure ENABLE_METRICS=true in .env"
    echo "  Make sure Gorax API is running: make run-api-dev"
  fi
else
  print_status "fail" "Port 9091 is not listening"
  echo "  Gorax API may not be running"
  echo "  Start with: make run-api-dev"
fi

echo ""

# Check Prometheus scraping
echo "Checking Prometheus data collection..."

if check_url "http://localhost:9090/api/v1/query?query=up{job=\"gorax-api\"}"; then
  result=$(curl -s "http://localhost:9090/api/v1/query?query=up{job=\"gorax-api\"}" 2>/dev/null | grep -o '"value":\[.*\]')

  if echo "$result" | grep -q '"1"'; then
    print_status "ok" "Prometheus is successfully scraping Gorax API"
  else
    print_status "fail" "Prometheus scrape target is down"
    echo "  Check: http://localhost:9090/targets"
  fi
else
  print_status "warn" "Cannot query Prometheus API"
fi

echo ""

# Check configuration files
echo "Checking configuration files..."

if [ -f "configs/prometheus.yml" ]; then
  print_status "ok" "configs/prometheus.yml exists"

  # Validate Prometheus config (if promtool is available)
  if command -v promtool > /dev/null 2>&1; then
    if promtool check config configs/prometheus.yml > /dev/null 2>&1; then
      print_status "ok" "Prometheus configuration is valid"
    else
      print_status "fail" "Prometheus configuration has errors"
    fi
  fi
else
  print_status "fail" "configs/prometheus.yml not found"
fi

if [ -f "configs/prometheus-alerts.yml" ]; then
  print_status "ok" "configs/prometheus-alerts.yml exists"

  # Validate alert rules (if promtool is available)
  if command -v promtool > /dev/null 2>&1; then
    if promtool check rules configs/prometheus-alerts.yml > /dev/null 2>&1; then
      print_status "ok" "Alert rules configuration is valid"
    else
      print_status "fail" "Alert rules configuration has errors"
    fi
  fi
else
  print_status "fail" "configs/prometheus-alerts.yml not found"
fi

if [ -f "dashboards/gorax-performance-overview.json" ]; then
  print_status "ok" "dashboards/gorax-performance-overview.json exists"
else
  print_status "fail" "dashboards/gorax-performance-overview.json not found"
fi

echo ""

# Check if Grafana dashboards are provisioned
echo "Checking Grafana dashboards..."

if check_url "http://localhost:3000/api/dashboards/uid/gorax-performance-overview"; then
  print_status "ok" "Gorax Performance Overview dashboard is provisioned"
else
  print_status "warn" "Dashboard may not be provisioned yet (or Grafana is not running)"
  echo "  Dashboard should auto-provision within 10 seconds of Grafana startup"
fi

echo ""

# Summary
echo "=========================================="
echo "Verification Complete"
echo "=========================================="
echo ""
echo "Access Points:"
echo "  • Prometheus:  http://localhost:9090"
echo "  • Grafana:     http://localhost:3000 (admin/admin)"
echo "  • Metrics:     http://localhost:9091/metrics"
echo ""
echo "Next Steps:"
echo "  1. Open Grafana: http://localhost:3000"
echo "  2. Navigate to Dashboards → Gorax → Performance Overview"
echo "  3. Generate some traffic to see metrics populate"
echo ""
echo "Documentation:"
echo "  • Monitoring Guide: docs/MONITORING.md"
echo "  • Deployment Guide: docs/GRAFANA_DEPLOYMENT_GUIDE.md"
echo "  • Dashboard Queries: dashboards/README.md"
echo ""
