# Gorax Smoke Test Suite

Quick smoke tests to verify all major features work together. These tests are designed to run fast (< 2 minutes) and verify critical paths are operational.

## 🎯 Purpose

Smoke tests provide:
- **Fast feedback**: Complete in under 2 minutes
- **Critical path verification**: Tests that core functionality works
- **Deployment confidence**: Run before/after deployments
- **CI/CD integration**: Automated testing on every commit

## 📋 What's Tested

### 1. API Smoke Tests (`api-smoke.sh`)
- Health and ready endpoints
- Marketplace API (public endpoints)
- OAuth provider listing
- Prometheus metrics
- Frontend app loading
- GraphQL endpoint

### 2. Database Smoke Tests (`db-smoke.sh`)
- PostgreSQL connectivity
- All critical tables exist:
  - `tenants`, `users`, `workflows`, `executions`
  - `credentials`, `webhooks`, `webhook_events`
  - `schedules`, `marketplace_*`, `oauth_*`
  - `audit_events`, `notifications`
- Database indices
- Foreign key constraints
- Query performance

### 3. Service Dependency Tests (`service-smoke.sh`)
- Redis connectivity and operations
- Docker container health (if applicable)
- LocalStack (AWS services mock)
- Kratos authentication service

### 4. Performance Tests (`perf-smoke.sh`)
- API response times (< 500ms)
- Health endpoints (< 100ms)
- Database queries (< 1000ms)
- Redis operations (< 50ms)

### 5. Go Workflow Tests (`go/workflow_smoke_test.go`)
- End-to-end workflow creation
- Execution tracking
- Database operations
- Critical table verification

## 🚀 Quick Start

### Prerequisites

```bash
# Required tools
- curl
- psql (PostgreSQL client)
- redis-cli (optional, for Redis tests)
- bc (for calculations)
- Go 1.23+ (for Go tests)
```

### Running All Tests

```bash
# From project root
cd tests/smoke
./run-all.sh
```

### Running Individual Test Suites

```bash
# API tests only
./api-smoke.sh

# Database tests only
./db-smoke.sh

# Service dependency tests
./service-smoke.sh

# Performance tests
./perf-smoke.sh

# Go workflow tests
cd go && go test -v -tags=smoke .
```

## ⚙️ Configuration

Set environment variables to customize behavior:

```bash
# API configuration
export BASE_URL=http://localhost:8080
export TEST_TENANT_ID=test-tenant-1

# Database configuration
export DATABASE_URL=postgres://user:pass@localhost:5433/gorax?sslmode=disable

# Redis configuration
export REDIS_HOST=localhost
export REDIS_PORT=6379

# Skip specific test suites
export SKIP_API=true        # Skip API tests
export SKIP_DB=true         # Skip database tests
export SKIP_SERVICES=true   # Skip service tests
export SKIP_PERF=true       # Skip performance tests
export SKIP_GO=true         # Skip Go tests

# Wait for services to start
export WAIT_FOR_SERVICES=true
```

## 📊 Example Output

```
=========================================
   🔥 Gorax Smoke Test Suite
=========================================

Base URL: http://localhost:8080
Database: postgres://postgres:postgres@localhost:5433/gorax
Redis: localhost:6379

=========================================
API Smoke Tests
=========================================
✓ Health check
✓ Ready check
✗ List marketplace templates (got 401, expected 200)
✓ Frontend app root

=========================================
Test Summary
=========================================
Total:  10
Passed: 9
Failed: 1

Failed tests:
  - List marketplace templates

✗ SMOKE TESTS FAILED
```

## 🔧 Troubleshooting

### Tests Fail: Connection Refused

**Problem**: API or services not running

**Solution**:
```bash
# Start services
docker-compose up -d postgres redis

# Start API
make run-api-dev

# Or use Docker
docker-compose up -d
```

### Database Connection Failed

**Problem**: PostgreSQL not accessible

**Solution**:
```bash
# Check PostgreSQL is running
docker ps | grep postgres

# Test connection manually
psql postgres://postgres:postgres@localhost:5433/gorax -c "SELECT 1"

# Ensure DATABASE_URL is correct
export DATABASE_URL=postgres://postgres:postgres@localhost:5433/gorax?sslmode=disable
```

### Redis Connection Failed

**Problem**: Redis not accessible

**Solution**:
```bash
# Check Redis is running
docker ps | grep redis

# Test connection manually
redis-cli -h localhost -p 6379 ping

# Ensure Redis config is correct
export REDIS_HOST=localhost
export REDIS_PORT=6379
```

### Performance Tests Failing

**Problem**: Response times too slow

**Causes**:
- Services warming up
- Resource constraints
- Network latency

**Solution**:
```bash
# Give services time to warm up
sleep 10

# Skip performance tests
export SKIP_PERF=true
./run-all.sh

# Adjust thresholds in perf-smoke.sh
MAX_API_TIME=1000  # Increase from 500ms
```

### Go Tests Failing

**Problem**: Module or dependency issues

**Solution**:
```bash
# Install dependencies
go mod download
go mod tidy

# Run with verbose output
cd go && go test -v -tags=smoke .

# Check test environment
export DATABASE_URL=postgres://postgres:postgres@localhost:5433/gorax?sslmode=disable
```

## 🚦 CI/CD Integration

### GitHub Actions

Smoke tests run automatically on:
- Every pull request
- Every push to `main` or `dev`
- Manual workflow dispatch

See `.github/workflows/smoke-tests.yml` for configuration.

### Running Locally (Like CI)

```bash
# Start services
docker-compose -f tests/docker-compose.test.yml up -d

# Set environment
export DATABASE_URL=postgres://postgres:postgres@localhost:5433/gorax?sslmode=disable
export BASE_URL=http://localhost:8080
export REDIS_HOST=localhost
export REDIS_PORT=6379

# Run migrations
for f in migrations/*.sql; do
    psql "$DATABASE_URL" -f "$f"
done

# Build and start API
make build-api
./bin/gorax-api &

# Wait for API
sleep 5

# Run smoke tests
cd tests/smoke
./run-all.sh

# Cleanup
killall gorax-api
docker-compose -f tests/docker-compose.test.yml down
```

## 📦 Pre-Deployment Checklist

Before deploying, run smoke tests:

```bash
# 1. Start services
make dev-simple

# 2. Start API
make run-api-dev

# 3. Run smoke tests
cd tests/smoke && ./run-all.sh

# 4. Verify all tests pass
# Expected: "✓ ALL SMOKE TESTS PASSED"
```

## 🔍 What Smoke Tests Don't Cover

Smoke tests are **not** a replacement for:
- **Unit tests**: Test individual functions/methods
- **Integration tests**: Test component interactions
- **E2E tests**: Test complete user workflows
- **Load tests**: Test system under load
- **Security tests**: Test for vulnerabilities

Use smoke tests to verify:
- ✅ Services are running
- ✅ Basic connectivity works
- ✅ Critical endpoints respond
- ✅ Database is accessible
- ✅ Performance is acceptable

## 📈 Performance Benchmarks

Expected response times:

| Endpoint | Target | Threshold |
|----------|--------|-----------|
| `/health` | < 50ms | < 100ms |
| `/ready` | < 50ms | < 100ms |
| API endpoints | < 200ms | < 500ms |
| Database queries | < 100ms | < 1000ms |
| Redis operations | < 10ms | < 50ms |

## 🛠️ Development

### Adding New Smoke Tests

1. **API Tests**: Add to `api-smoke.sh`
```bash
test_endpoint "My new endpoint" "GET" "/api/v1/my-endpoint" 200
```

2. **Database Tests**: Add to `db-smoke.sh`
```bash
test_database "My table" "SELECT COUNT(*) FROM my_table"
```

3. **Performance Tests**: Add to `perf-smoke.sh`
```bash
test_endpoint_with_timing "My endpoint" "GET" "/api/v1/endpoint" 200 500
```

4. **Go Tests**: Add to `go/workflow_smoke_test.go`
```go
func TestMyFeature(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping smoke test in short mode")
    }
    // Test implementation
}
```

### Test Library Functions

Available in `lib.sh`:

```bash
# Test HTTP endpoint
test_endpoint "name" "METHOD" "/path" expected_status

# Test with timing
test_endpoint_with_timing "name" "METHOD" "/path" expected_status max_ms

# Test database query
test_database "name" "SQL query"

# Test Redis
test_redis "name"

# Test Docker container
test_container "name" "container-name"

# Wait for service
wait_for_service "name" "http://url" max_attempts wait_seconds

# Print functions
print_status "PASS|FAIL|SKIP|INFO" "message"
print_header "Section Title"
print_summary  # Print final results
```

## 📚 Resources

- [Testing Strategy](../TESTING_STRATEGY.md)
- [Integration Tests](../integration/README.md)
- [E2E Tests](../e2e/README.md)
- [Load Tests](../load/README.md)

## 🤝 Contributing

When adding new features:

1. **Add smoke test** for critical path
2. **Keep tests fast** (< 2 seconds per test)
3. **Use descriptive names** for test cases
4. **Handle failures gracefully** (cleanup on failure)
5. **Document new tests** in this README

## 📝 Notes

- Smoke tests use build tag `smoke` for Go tests
- Tests run in parallel where possible
- Failed tests print detailed error messages
- All tests clean up after themselves
- Tests are idempotent (can run multiple times)

## 🔖 Version

Last Updated: 2026-01-02
Gorax Version: Compatible with v1.0+
