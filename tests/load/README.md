# Gorax Load Testing Suite

Comprehensive performance and load testing suite for the Gorax workflow automation platform using k6.

## Overview

This test suite provides systematic load and performance testing across all critical platform components:

- **Workflow API**: CRUD operations on workflows
- **Execution API**: Workflow execution throughput and performance
- **Webhook Triggers**: Webhook ingestion rate and reliability
- **WebSocket Connections**: Real-time connection scaling
- **Authentication**: Auth endpoint performance under load

## Prerequisites

### Install k6

**macOS:**
```bash
brew install k6
```

**Linux:**
```bash
sudo gpg -k
sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6
```

**Windows:**
```powershell
choco install k6
```

**Docker:**
```bash
docker pull grafana/k6:latest
```

### Setup Test Environment

1. **Start the Gorax platform:**
```bash
cd /Users/shawntherrien/Projects/gorax
make run
```

2. **Configure environment variables** (optional):
```bash
# Updated defaults for Gorax (port 8181)
export BASE_URL="http://localhost:8181"
export WS_URL="ws://localhost:8181"

# Authentication mode ('dev' for DevAuth, 'kratos' for production)
export AUTH_MODE="dev"  # Default: 'dev'

# DevAuth credentials (for development)
export TEST_TENANT_ID="default-tenant"  # Default tenant ID
export TEST_USER_ID="default-test-user"  # Default user ID

# Kratos credentials (for production)
export TEST_USER_EMAIL="loadtest@example.com"
export TEST_USER_PASSWORD="loadtest123"
export KRATOS_PUBLIC_URL="http://localhost:4433"
```

## Authentication Modes

Gorax supports two authentication modes depending on the environment:

### Development Mode (DevAuth)

**Default mode** for local development and testing.

Uses header-based authentication:
- `X-Tenant-ID`: Identifies the tenant (default: `default-tenant`)
- `X-User-ID`: Identifies the user (default: `default-test-user`)

**No login endpoints required** - authentication is handled by middleware that reads headers.

**Configuration:**
```bash
export AUTH_MODE="dev"
export TEST_TENANT_ID="your-tenant-id"
export TEST_USER_ID="your-user-id"
```

**When to use:**
- Local development
- Testing in development environments
- CI/CD tests against development servers
- Quick smoke tests

### Production Mode (Kratos)

Uses Ory Kratos for production authentication.

Requires login via Kratos endpoints:
- Login flow initialization
- Credential submission
- Session token management

**Configuration:**
```bash
export AUTH_MODE="kratos"
export KRATOS_PUBLIC_URL="https://auth.gorax.io"
export TEST_USER_EMAIL="loadtest@example.com"
export TEST_USER_PASSWORD="loadtest123"
```

**When to use:**
- Staging environment testing
- Production performance testing
- Pre-release validation
- Load testing with production-like auth overhead

### Authentication in Tests

The test suite automatically detects and uses the appropriate authentication mode:

**DevAuth Example:**
```javascript
// No setup required - auth is handled by middleware
const headers = {
  'X-Tenant-ID': 'default-tenant',
  'X-User-ID': 'default-test-user',
};
```

**Kratos Example** (future):
```javascript
// Login and get session token
const token = await authenticate();
const headers = {
  'Authorization': `Bearer ${token}`,
};
```

### Test Compatibility

| Test Suite | DevAuth | Kratos |
|------------|---------|--------|
| workflow_api.js | ✅ | ✅ |
| execution_api.js | ✅ | ✅ |
| webhook_trigger.js | ✅ | ✅ |
| websocket_connections.js | ✅ | ✅ |
| auth_endpoints.js | ⚠️ Skipped | ✅ |

**Note:** The `auth_endpoints.js` test is skipped in DevAuth mode since DevAuth doesn't use login endpoints.

## Test Scenarios

### Smoke Test
**Purpose:** Verify system functionality with minimal load
**Duration:** 1 minute
**VUs:** 1
**Use Case:** Quick health check before major tests

```bash
./run_tests.sh --scenario smoke
```

### Load Test (Default)
**Purpose:** Test under normal expected load
**Duration:** 9 minutes
**VUs:** 10 (ramped)
**Use Case:** Standard performance validation

```bash
./run_tests.sh --scenario load
```

### Stress Test
**Purpose:** Push system beyond normal capacity
**Duration:** 26 minutes
**VUs:** 10 → 20 → 50 → 100 (staged)
**Use Case:** Find breaking points and bottlenecks

```bash
./run_tests.sh --scenario stress
```

### Spike Test
**Purpose:** Test sudden traffic spikes
**Duration:** 8 minutes
**VUs:** 10 → 100 → 10 (rapid changes)
**Use Case:** Validate auto-scaling and recovery

```bash
./run_tests.sh --scenario spike
```

### Soak Test
**Purpose:** Extended duration testing for memory leaks
**Duration:** 70 minutes
**VUs:** 20 (sustained)
**Use Case:** Long-term stability validation

```bash
./run_tests.sh --scenario soak
```

## Running Tests

### Run All Tests
```bash
./run_tests.sh
```

### Run Specific Test
```bash
# Run only workflow API tests
./run_tests.sh workflow

# Run only execution tests
./run_tests.sh execution

# Run only webhook tests
./run_tests.sh webhook

# Run only WebSocket tests
./run_tests.sh websocket

# Run only auth tests
./run_tests.sh auth
```

### Custom Configuration
```bash
# Run against staging environment
./run_tests.sh --url https://staging.gorax.io --scenario stress

# Custom output directory
./run_tests.sh --output /tmp/load-results

# Run smoke test with custom user
./run_tests.sh --scenario smoke --email test@example.com --password testpass123

# Combine options
./run_tests.sh workflow --scenario load --url https://api.gorax.io
```

### Using Docker
```bash
# Run with Docker
docker run --rm -v $(pwd):/tests grafana/k6:latest run /tests/workflow_api.js \
  -e BASE_URL=http://host.docker.internal:8080 \
  -e SCENARIO=load
```

## Test Details

### Workflow API Test (`workflow_api.js`)

Tests workflow CRUD operations:
- ✓ Create workflow
- ✓ Get workflow by ID
- ✓ List workflows
- ✓ Update workflow
- ✓ Delete workflow
- ✓ Verify deletion

**Key Metrics:**
- `workflow_create_duration`: Time to create workflow (p95 < 1s)
- `workflow_read_duration`: Time to retrieve workflow (p95 < 500ms)
- `workflow_update_duration`: Time to update workflow (p95 < 1s)
- `workflow_delete_duration`: Time to delete workflow (p95 < 500ms)
- `workflow_success_rate`: Overall success rate (> 95%)

### Execution API Test (`execution_api.js`)

Tests workflow execution performance:
- ✓ Start execution
- ✓ Poll execution status
- ✓ Get execution history
- ✓ Concurrent executions

**Key Metrics:**
- `execution_start_duration`: Time to start execution (p95 < 2s)
- `execution_status_check_duration`: Time to check status (p95 < 500ms)
- `execution_throughput`: Executions per second
- `execution_success_rate`: Successful completions (> 95%)

### Webhook Trigger Test (`webhook_trigger.js`)

Tests webhook ingestion performance:
- ✓ Single webhook trigger
- ✓ Batch webhook triggers
- ✓ Different content types
- ✓ Large payload handling
- ✓ Validation errors
- ✓ High-frequency bursts

**Key Metrics:**
- `webhook_ingestion_duration`: Time to accept webhook (p95 < 200ms)
- `webhook_throughput`: Webhooks per second
- `webhook_success_rate`: Acceptance rate (> 99%)

### WebSocket Test (`websocket_connections.js`)

Tests WebSocket scaling:
- ✓ Connection establishment
- ✓ Message sending/receiving
- ✓ Message latency
- ✓ Connection reliability
- ✓ Message bursts

**Key Metrics:**
- `ws_connection_duration`: Time to connect (p95 < 1s)
- `ws_message_latency`: Message round-trip time (p95 < 100ms)
- `ws_connection_success_rate`: Connection success (> 95%)

### Authentication Test (`auth_endpoints.js`)

Tests authentication performance:
- ✓ User login
- ✓ Token validation
- ✓ Protected endpoint access
- ✓ Token refresh
- ✓ Invalid credentials handling
- ✓ Logout
- ✓ Concurrent logins

**Key Metrics:**
- `login_duration`: Time to login (p95 < 300ms)
- `token_refresh_duration`: Time to refresh token (p95 < 200ms)
- `auth_success_rate`: Auth success rate (> 95%)

## Performance Thresholds

### Global Thresholds
```javascript
http_req_duration: ['p(95)<500', 'p(99)<1000']  // 95% < 500ms, 99% < 1s
http_req_failed: ['rate<0.01']                   // Error rate < 1%
iteration_duration: ['p(95)<5000']               // Iteration < 5s
```

### Component-Specific Thresholds
- **Workflow Create:** p95 < 1s
- **Workflow Execute:** p95 < 2s
- **Webhook Trigger:** p95 < 200ms
- **Auth Login:** p95 < 300ms
- **WebSocket Connect:** p95 < 500ms

## Interpreting Results

### Successful Test Run
```
✓ workflow_api completed successfully
✓ execution_api completed successfully
✓ webhook_trigger completed successfully
✓ websocket_connections completed successfully
✓ auth_endpoints completed successfully

Test Summary:
  Total Tests:  5
  Passed:       5
  Failed:       0
```

### Key Metrics to Monitor

**HTTP Performance:**
- `http_reqs`: Total requests made
- `http_req_duration`: Request latency (avg, p95, p99)
- `http_req_failed`: Error rate

**Custom Metrics:**
- `*_duration`: Operation-specific latencies
- `*_throughput`: Operations per second
- `*_success_rate`: Operation success rates
- `*_errors`: Error counts

### What to Look For

**Good Performance:**
- ✓ p95 latencies within thresholds
- ✓ Error rates < 1%
- ✓ Success rates > 95%
- ✓ Stable throughput
- ✓ No increasing trend in latencies

**Performance Issues:**
- ✗ p95 latencies exceeding thresholds
- ✗ Error rates > 1%
- ✗ Success rates < 95%
- ✗ Declining throughput
- ✗ Increasing latencies over time (memory leak)

### Common Issues

**High Latency:**
- Database query optimization needed
- Insufficient server resources
- Network bottlenecks

**High Error Rate:**
- Application crashes under load
- Database connection pool exhausted
- Rate limiting triggered

**Memory Leaks (Soak Test):**
- Increasing latency over time
- Eventually OOM errors
- Fix: Review goroutine leaks, unclosed connections

## Results and Reporting

### Output Files

Tests generate multiple output files in the `results/` directory:

```
results/
├── workflow_api_20231220_143022.json          # Raw k6 metrics
├── workflow_api_20231220_143022_summary.txt   # Text summary
├── execution_api_20231220_143022.json
├── webhook_trigger_20231220_143022.json
├── websocket_connections_20231220_143022.json
├── auth_endpoints_20231220_143022.json
└── combined_report_20231220_143022.html       # HTML report (all tests)
```

### Viewing Results

**Text Summary:**
```bash
cat results/*_summary.txt
```

**JSON Analysis:**
```bash
# Use jq to query metrics
jq '.metrics.http_req_duration' results/workflow_api_*.json

# Get p95 latency
jq '.metrics.http_req_duration.values."p(95)"' results/workflow_api_*.json
```

**HTML Report:**
```bash
open results/combined_report_*.html
```

## Integration with CI/CD

### GitHub Actions Example

```yaml
name: Load Tests

on:
  schedule:
    - cron: '0 2 * * *'  # Daily at 2 AM
  workflow_dispatch:

jobs:
  load-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Setup k6
        run: |
          sudo gpg -k
          sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
          echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
          sudo apt-get update
          sudo apt-get install k6

      - name: Start Gorax
        run: |
          docker-compose up -d
          sleep 30

      - name: Run Load Tests
        env:
          BASE_URL: http://localhost:8080
          SCENARIO: load
        run: |
          cd tests/load
          ./run_tests.sh --scenario smoke

      - name: Upload Results
        uses: actions/upload-artifact@v3
        with:
          name: load-test-results
          path: tests/load/results/
```

## Prometheus/Grafana Integration

### Export Metrics to Prometheus

1. **Install k6 Prometheus exporter:**
```bash
docker run -d -p 9090:9090 prom/prometheus
```

2. **Configure k6 to export metrics:**
```bash
k6 run --out prometheus workflow_api.js
```

3. **Create Grafana dashboard** to visualize:
   - Request rates
   - Latency percentiles
   - Error rates
   - Throughput trends

### Sample Prometheus Queries

```promql
# p95 latency
histogram_quantile(0.95, rate(http_req_duration_bucket[5m]))

# Error rate
rate(http_req_failed[5m])

# Throughput
rate(http_reqs[5m])
```

## Performance Baselines

### Expected Performance (Local Development)

| Metric | Target | Acceptable | Critical |
|--------|--------|------------|----------|
| Workflow Create | < 500ms | < 1s | > 2s |
| Workflow Execute | < 1s | < 2s | > 5s |
| Webhook Ingestion | < 100ms | < 200ms | > 500ms |
| Auth Login | < 200ms | < 300ms | > 1s |
| WebSocket Connect | < 500ms | < 1s | > 2s |

### Expected Performance (Production)

| Metric | Target | Acceptable | Critical |
|--------|--------|------------|----------|
| Workflow Create | < 300ms | < 500ms | > 1s |
| Workflow Execute | < 500ms | < 1s | > 3s |
| Webhook Ingestion | < 50ms | < 100ms | > 200ms |
| Auth Login | < 100ms | < 200ms | > 500ms |
| WebSocket Connect | < 300ms | < 500ms | > 1s |

### Capacity Planning

**Single Instance:**
- Concurrent users: 50-100
- Workflow executions/min: 1,000-2,000
- Webhook ingestion/s: 100-200
- WebSocket connections: 1,000-2,000

**Scaled (3 instances):**
- Concurrent users: 150-300
- Workflow executions/min: 3,000-6,000
- Webhook ingestion/s: 300-600
- WebSocket connections: 3,000-6,000

## Troubleshooting

### Tests Failing to Connect

**Issue:** Connection refused or timeout
**Solution:**
```bash
# Verify server is running
curl http://localhost:8080/health

# Check if port is in use
lsof -i :8080

# Review server logs
docker logs gorax-api
```

### Authentication Failures

**Issue:** 401 Unauthorized errors
**Solution:**
```bash
# Verify test user exists
# Create user via API or admin interface

# Test login manually
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"loadtest@example.com","password":"loadtest123"}'
```

### Rate Limiting

**Issue:** 429 Too Many Requests
**Solution:**
```bash
# Adjust rate limits in server config
# Or reduce test VUs
./run_tests.sh --scenario smoke  # Uses fewer VUs
```

### Out of Memory

**Issue:** Server crashes during soak test
**Solution:**
- Increase server memory allocation
- Review for memory leaks
- Check database connection pooling

## Best Practices

1. **Run smoke tests first** before running longer tests
2. **Baseline performance** on a clean system before major changes
3. **Run tests regularly** (daily/weekly) to catch regressions
4. **Compare results** over time to identify trends
5. **Test in production-like environment** for accurate results
6. **Monitor server resources** (CPU, memory, disk) during tests
7. **Document baselines** for your specific infrastructure
8. **Gradually increase load** to find breaking points
9. **Fix issues before optimization** (correctness > performance)
10. **Test after major changes** (new features, refactoring, dependency updates)

## Contributing

When adding new tests:

1. Follow existing test structure
2. Add appropriate custom metrics
3. Set reasonable thresholds
4. Document the test in this README
5. Update `run_tests.sh` to include new test
6. Test locally before committing

## References

- [k6 Documentation](https://k6.io/docs/)
- [k6 Best Practices](https://k6.io/docs/testing-guides/running-large-tests/)
- [HTTP Load Testing](https://k6.io/docs/using-k6/http-requests/)
- [WebSocket Testing](https://k6.io/docs/using-k6/protocols/websockets/)
- [Performance Metrics](https://k6.io/docs/using-k6/metrics/)
