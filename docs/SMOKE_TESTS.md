# Smoke Tests Guide

## Overview

Smoke tests are quick, automated tests that verify critical system functionality. They're designed to catch major issues immediately after deployment or code changes.

## Quick Start

```bash
# Run all smoke tests
make smoke-tests

# Run specific test suites
make smoke-tests-api      # API endpoints only
make smoke-tests-db       # Database only
make smoke-tests-quick    # Skip Go tests (faster)
```

## What Are Smoke Tests?

Smoke tests are:
- ✅ **Fast**: Complete in < 2 minutes
- ✅ **Critical**: Test essential functionality only
- ✅ **Automated**: Run on every commit via CI/CD
- ✅ **Simple**: Clear pass/fail results

Smoke tests are NOT:
- ❌ Comprehensive unit tests
- ❌ Full integration test suites
- ❌ Load or performance tests
- ❌ Security audits

## Test Coverage

### 1. API Smoke Tests
**File**: `tests/smoke/api-smoke.sh`

Tests critical API endpoints:
- Health check (`/health`)
- Ready check (`/ready`)
- Marketplace endpoints
- OAuth provider listing
- Frontend app loading

**Example**:
```bash
./tests/smoke/api-smoke.sh
```

### 2. Database Smoke Tests
**File**: `tests/smoke/db-smoke.sh`

Verifies database connectivity and structure:
- PostgreSQL connection
- All critical tables exist
- Indices are present
- Foreign key constraints
- Query performance

**Critical Tables Tested**:
- `tenants`, `users`, `workflows`, `executions`
- `credentials`, `webhooks`, `schedules`
- `marketplace_*`, `oauth_*`, `audit_events`

**Example**:
```bash
./tests/smoke/db-smoke.sh
```

### 3. Service Dependency Tests
**File**: `tests/smoke/service-smoke.sh`

Tests external service dependencies:
- Redis connectivity
- Redis read/write operations
- Docker containers (if applicable)
- LocalStack (AWS mock)
- Kratos (if configured)

**Example**:
```bash
./tests/smoke/service-smoke.sh
```

### 4. Performance Tests
**File**: `tests/smoke/perf-smoke.sh`

Validates response time thresholds:
- Health endpoints: < 100ms
- API endpoints: < 500ms
- Database queries: < 1000ms
- Redis operations: < 50ms

**Example**:
```bash
./tests/smoke/perf-smoke.sh
```

### 5. Go Workflow Tests
**File**: `tests/smoke/go/workflow_smoke_test.go`

End-to-end workflow execution tests:
- Workflow creation
- Execution tracking
- Database operations
- Table verification

**Example**:
```bash
cd tests/smoke/go && go test -v -tags=smoke .
```

## Configuration

### Environment Variables

```bash
# API Configuration
export BASE_URL=http://localhost:8080
export TEST_TENANT_ID=test-tenant-1

# Database Configuration
export DATABASE_URL=postgres://user:pass@localhost:5433/gorax?sslmode=disable

# Redis Configuration
export REDIS_HOST=localhost
export REDIS_PORT=6379

# Skip Specific Tests
export SKIP_API=true        # Skip API tests
export SKIP_DB=true         # Skip database tests
export SKIP_SERVICES=true   # Skip service tests
export SKIP_PERF=true       # Skip performance tests
export SKIP_GO=true         # Skip Go tests

# Wait for Services
export WAIT_FOR_SERVICES=true
```

## Running Locally

### Option 1: Quick Run (Development)

```bash
# 1. Start services
make dev-simple

# 2. Start API
make run-api-dev

# 3. Run smoke tests
make smoke-tests
```

### Option 2: Docker Environment

```bash
# 1. Start all services with Docker
docker-compose up -d

# 2. Wait for services
./scripts/wait-for-services.sh

# 3. Run smoke tests
make smoke-tests
```

### Option 3: Individual Test Suites

```bash
# Run each suite separately
make smoke-tests-api
make smoke-tests-db
make smoke-tests-services
make smoke-tests-perf
```

## CI/CD Integration

### GitHub Actions

Smoke tests run automatically on:
- Every pull request to `main` or `dev`
- Every push to `main` or `dev`
- Manual workflow dispatch

**Workflow File**: `.github/workflows/smoke-tests.yml`

**View Results**:
1. Go to "Actions" tab in GitHub
2. Select "Smoke Tests" workflow
3. View detailed test results

### Local CI Simulation

Run tests exactly as CI does:

```bash
# 1. Start test services
docker-compose -f tests/docker-compose.test.yml up -d

# 2. Set environment
export DATABASE_URL=postgres://postgres:postgres@localhost:5433/gorax?sslmode=disable
export BASE_URL=http://localhost:8080

# 3. Run migrations
for f in migrations/*.sql; do
    psql "$DATABASE_URL" -f "$f"
done

# 4. Build and start API
make build-api
./bin/gorax-api &
API_PID=$!

# 5. Wait for services
./scripts/wait-for-services.sh

# 6. Run tests
make smoke-tests

# 7. Cleanup
kill $API_PID
docker-compose -f tests/docker-compose.test.yml down
```

## Interpreting Results

### Success

```
=========================================
   📊 Final Summary
=========================================
Total Suites:  5
Passed:        5
Failed:        0

✓ ALL SMOKE TESTS PASSED

All critical paths are working correctly!
```

**What this means**:
- ✅ All services are running
- ✅ API is responding correctly
- ✅ Database is accessible
- ✅ Performance is within thresholds

### Failure

```
=========================================
   📊 Final Summary
=========================================
Total Suites:  5
Passed:        4
Failed:        1

Failed test suites:
  - Database Smoke Tests

✗ SMOKE TESTS FAILED
```

**What to do**:
1. Review failed test output
2. Check service logs
3. Verify configuration
4. Fix issue and re-run

## Troubleshooting

### API Connection Failed

**Symptom**: `connection refused` errors

**Solutions**:
```bash
# Check if API is running
curl http://localhost:8080/health

# Start API if needed
make run-api-dev

# Check port availability
lsof -i :8080
```

### Database Connection Failed

**Symptom**: `could not connect to server`

**Solutions**:
```bash
# Check if PostgreSQL is running
docker ps | grep postgres

# Test connection manually
psql postgres://postgres:postgres@localhost:5433/gorax -c "SELECT 1"

# Start database if needed
docker-compose up -d postgres
```

### Redis Connection Failed

**Symptom**: `Could not connect to Redis`

**Solutions**:
```bash
# Check if Redis is running
docker ps | grep redis

# Test connection
redis-cli -h localhost -p 6379 ping

# Start Redis if needed
docker-compose up -d redis
```

### Performance Tests Failing

**Symptom**: Response times exceeding thresholds

**Causes**:
- Services still warming up
- System under load
- Network latency

**Solutions**:
```bash
# Give services time to warm up
sleep 10
make smoke-tests

# Skip performance tests temporarily
export SKIP_PERF=true
make smoke-tests

# Adjust thresholds (in perf-smoke.sh if needed)
```

### Go Tests Failing

**Symptom**: Module or dependency errors

**Solutions**:
```bash
# Install dependencies
cd tests/smoke/go
go mod download
go mod tidy

# Run with verbose output
go test -v -tags=smoke .

# Check environment
export DATABASE_URL=postgres://postgres:postgres@localhost:5433/gorax?sslmode=disable
```

## Pre-Deployment Checklist

Before deploying to production:

- [ ] All smoke tests pass locally
- [ ] All smoke tests pass in CI/CD
- [ ] No skipped tests (all suites run)
- [ ] Performance within acceptable ranges
- [ ] No database connectivity issues
- [ ] All services responding correctly

**Command**:
```bash
make smoke-tests
```

**Expected Output**: `✓ ALL SMOKE TESTS PASSED`

## Development Workflow

### Adding New Features

When adding a new feature:

1. **Write smoke test** for critical path:
```bash
# Add to appropriate test file
vim tests/smoke/api-smoke.sh
```

2. **Test locally**:
```bash
make smoke-tests
```

3. **Commit with tests**:
```bash
git add tests/smoke/
git commit -m "feat: add new feature with smoke tests"
```

4. **CI runs automatically**:
- Tests run on PR
- Must pass before merge

### Adding New Smoke Tests

#### Bash Tests

Add to existing scripts:

```bash
# tests/smoke/api-smoke.sh
test_endpoint "My new endpoint" "GET" "/api/v1/my-feature" 200
```

#### Go Tests

Add to `workflow_smoke_test.go`:

```go
func TestMyNewFeature(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping smoke test in short mode")
    }

    // Test implementation
    ctx := context.Background()
    // ... test logic
}
```

## Best Practices

### DO ✅

- Keep tests fast (< 2 seconds each)
- Test critical paths only
- Use descriptive test names
- Clean up test data
- Make tests idempotent
- Handle failures gracefully

### DON'T ❌

- Test every edge case (use unit tests)
- Make external API calls
- Test implementation details
- Leave test data behind
- Skip cleanup on failure
- Make tests dependent on order

## Performance Benchmarks

Target response times:

| Service | Target | Threshold | Test |
|---------|--------|-----------|------|
| Health endpoint | < 50ms | < 100ms | ✓ |
| API endpoints | < 200ms | < 500ms | ✓ |
| Database queries | < 100ms | < 1s | ✓ |
| Redis operations | < 10ms | < 50ms | ✓ |

## Related Documentation

- [Testing Strategy](TESTING_STRATEGY.md)
- [Integration Tests](../tests/integration/README.md)
- [E2E Tests](../tests/e2e/README.md)
- [Load Tests](../tests/load/README.md)
- [CI/CD Setup](CI-CD-SETUP-SUMMARY.md)

## FAQ

### Q: How long should smoke tests take?
**A**: < 2 minutes total. Individual tests should be < 2 seconds.

### Q: Can I run smoke tests in production?
**A**: Yes, but use read-only tests and dedicated test tenant.

### Q: What if smoke tests are flaky?
**A**: Add retries, increase timeouts, or add wait conditions. Report flaky tests as bugs.

### Q: Should I run smoke tests before every commit?
**A**: Recommended, but not required. Always run before:
- Creating a pull request
- Deploying to any environment
- After major refactoring

### Q: Can I skip smoke tests?
**A**: In development, yes (use `SKIP_*` flags). In CI/CD, no.

### Q: How do I add a new smoke test?
**A**: See "Adding New Smoke Tests" section above.

## Support

For issues:
- Check this documentation
- Review test output
- Check CI/CD logs
- Open an issue in the repository

## Version History

- **v1.0** (2026-01-02): Initial smoke test suite
  - API, Database, Service, Performance, Go tests
  - CI/CD integration
  - Comprehensive documentation
