# Gorax Integration and E2E Test Suite

Comprehensive test suite for enterprise features including OAuth, SSO, Marketplace, Audit Logging, Database Connectors, and Error Handling.

## Table of Contents

- [Overview](#overview)
- [Test Structure](#test-structure)
- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Integration Tests](#integration-tests)
- [E2E Tests](#e2e-tests)
- [Mock Services](#mock-services)
- [CI/CD Integration](#cicd-integration)
- [Writing New Tests](#writing-new-tests)
- [Troubleshooting](#troubleshooting)

## Overview

This test suite provides:

- **Integration Tests**: Backend API and service layer tests with real database
- **E2E Tests**: Full user flow tests using Playwright
- **Mock Services**: OAuth IdP, SAML IdP, and external service mocks
- **Test Fixtures**: Reusable test data generators
- **Performance Tests**: Load and performance benchmarks

### Coverage Goals

- Integration Tests: 80%+ coverage of new features
- E2E Tests: All critical user flows
- Performance Tests: Baseline metrics established

## Test Structure

```
tests/
├── integration/           # Go integration tests
│   ├── setup.go          # Test server setup
│   ├── helpers.go        # HTTP helpers
│   ├── marketplace_full_test.go
│   ├── audit_test.go
│   ├── oauth_test.go
│   ├── sso_test.go
│   ├── database_test.go
│   └── error_handling_test.go
├── e2e/                   # Playwright E2E tests
│   ├── setup.ts          # E2E test setup
│   ├── marketplace.spec.ts
│   ├── oauth.spec.ts
│   ├── sso.spec.ts
│   ├── audit.spec.ts
│   └── workflow_error_handling.spec.ts
├── mocks/                 # Mock services
│   ├── oauth_idp.go      # Mock OAuth provider
│   └── saml/             # SAML IdP config
├── fixtures/              # Test data generators
│   └── generators.go
├── performance/           # Performance tests
│   └── k6/               # k6 load tests
├── docker-compose.test.yml
└── README.md
```

## Prerequisites

### For Integration Tests

- Go 1.23+
- Docker & Docker Compose
- PostgreSQL 16+ (via Docker)
- Redis 7+ (via Docker)

### For E2E Tests

- Node.js 20+
- Playwright browsers
- Running backend server

### For All Tests

```bash
# Install Go dependencies
go mod download

# Install Playwright (for E2E tests)
cd ../web
npm install
npx playwright install

# Start test dependencies
cd ../tests
docker-compose -f docker-compose.test.yml up -d
```

## Quick Start

### Run All Tests

```bash
# From project root
make test-all
```

### Run Integration Tests Only

```bash
# Start test dependencies
docker-compose -f tests/docker-compose.test.yml up -d

# Wait for services to be ready
./scripts/wait-for-services.sh

# Run tests
go test -v -tags=integration ./tests/integration/...

# With coverage
go test -v -tags=integration -coverprofile=coverage.out ./tests/integration/...
```

### Run E2E Tests Only

```bash
# Start backend server in test mode
ENV=test go run cmd/api/main.go &

# Run Playwright tests
cd web
npm run test:e2e

# With UI mode (interactive)
npm run test:e2e:ui
```

### Run Specific Test Suite

```bash
# Integration test for marketplace
go test -v -tags=integration ./tests/integration/ -run TestMarketplace

# Integration test for audit logging
go test -v -tags=integration ./tests/integration/ -run TestAudit

# E2E test for marketplace
cd web
npx playwright test marketplace.spec.ts
```

## Integration Tests

Integration tests verify backend functionality with real database connections.

### Test Server Setup

The `TestServer` provides a fully configured environment:

```go
func TestMyFeature(t *testing.T) {
    ts := SetupTestServer(t)
    defer ts.Cleanup()

    // Create test data
    tenantID := ts.CreateTestTenant(t, "Test Tenant")
    userID := ts.CreateTestUser(t, tenantID, "user@test.com", "user")

    // Make API requests
    headers := DefaultTestHeaders(tenantID)
    resp := ts.MakeRequest(t, "GET", "/api/v1/workflows", nil, headers)
    AssertStatusCode(t, resp, http.StatusOK)
}
```

### Available Test Suites

#### Marketplace Tests
- Full workflow (publish → search → rate → install)
- Review management (create, update, delete)
- Search and filtering
- Popular and trending templates

#### OAuth Tests
- Provider listing
- Authorization flow with PKCE
- Token exchange
- Refresh token flow
- Connection revocation
- Auto token refresh

#### SSO Tests
- SAML provider configuration
- OIDC provider configuration
- SSO login flow
- JIT user provisioning
- Domain-based discovery

#### Audit Logging Tests
- Event logging (sync and async)
- Query and filtering
- Statistics aggregation
- Retention policies
- Performance benchmarks

#### Database Connector Tests
- PostgreSQL queries
- MySQL queries
- MongoDB queries
- SQL injection prevention
- SSRF protection
- Connection pooling

#### Error Handling Tests
- Try/catch workflow execution
- Retry with exponential backoff
- Circuit breaker
- Error classification
- Finally blocks

### Running with Different Databases

```bash
# Use specific database
DATABASE_URL=postgres://user:pass@localhost:5432/testdb go test ./tests/integration/...

# Use Docker services
docker-compose -f tests/docker-compose.test.yml up postgres-test redis-test
go test ./tests/integration/...
```

## E2E Tests

End-to-end tests verify complete user workflows using Playwright.

### Test Structure

```typescript
import { test, expect } from '@playwright/test';

test('marketplace workflow', async ({ page }) => {
  // Login
  await page.goto('/login');
  await page.fill('[name="email"]', 'test@example.com');
  await page.click('button[type="submit"]');

  // Navigate to marketplace
  await page.goto('/marketplace');

  // Test user flow
  // ...
});
```

### Available E2E Test Suites

- **marketplace.spec.ts**: Browse, search, install templates
- **oauth.spec.ts**: Connect OAuth providers
- **sso.spec.ts**: SSO login and configuration
- **audit.spec.ts**: View and filter audit logs
- **workflow_error_handling.spec.ts**: Error handling in workflows

### Running E2E Tests

```bash
# All E2E tests
npm run test:e2e

# Specific file
npx playwright test marketplace.spec.ts

# With UI mode (debug)
npx playwright test --ui

# Headed mode (see browser)
npx playwright test --headed

# Specific browser
npx playwright test --project=chromium
```

### E2E Test Configuration

Edit `playwright.config.ts`:

```typescript
export default defineConfig({
  testDir: './tests/e2e',
  timeout: 30000,
  retries: 2,
  use: {
    baseURL: 'http://localhost:3000',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
  },
  projects: [
    { name: 'chromium', use: { ...devices['Desktop Chrome'] } },
    { name: 'firefox', use: { ...devices['Desktop Firefox'] } },
    { name: 'webkit', use: { ...devices['Desktop Safari'] } },
  ],
});
```

## Mock Services

### Mock OAuth Provider

Located in `tests/mocks/oauth_idp.go`, provides:

- OAuth 2.0 authorization flow
- Token exchange
- Refresh token flow
- User info endpoint
- OIDC discovery

**Usage in tests:**

```go
import "github.com/gorax/gorax/tests/mocks"

func TestOAuth(t *testing.T) {
    // Create mock provider
    mockProvider := mocks.NewMockOAuthProvider("github", "client-id", "client-secret")
    defer mockProvider.Close()

    // Add test user
    mockProvider.AddUser(&mocks.MockUser{
        ID:       "user-123",
        Email:    "test@example.com",
        Name:     "Test User",
        Username: "testuser",
    })

    // Use in tests
    authURL := mockProvider.AuthURL()
    // ... test OAuth flow
}
```

### Mock SAML IdP

Uses `kristophjunge/test-saml-idp` Docker image:

```bash
docker-compose -f tests/docker-compose.test.yml up mock-saml-idp
```

Access at `http://localhost:8091/simplesaml/`

### LocalStack (AWS Services)

Mock AWS services (S3, SQS, KMS):

```bash
docker-compose -f tests/docker-compose.test.yml up localstack-test
```

Configure tests:

```bash
AWS_ENDPOINT=http://localhost:4566
AWS_ACCESS_KEY_ID=test
AWS_SECRET_ACCESS_KEY=test
```

## Test Fixtures

Test data generators in `tests/fixtures/generators.go`:

```go
import "github.com/gorax/gorax/tests/fixtures"

// Generate workflow
workflow := fixtures.GenerateWorkflowDefinition(fixtures.WorkflowDefinition{
    Name: "My Workflow",
    NodeTypes: []string{"trigger", "action", "condition"},
    EdgeCount: 2,
})

// Generate user
user := fixtures.GenerateUser()

// Generate tenant
tenant := fixtures.GenerateTenant()

// Generate template
template := fixtures.GenerateWorkflowTemplate()

// Generate credential
cred := fixtures.GenerateCredential("api_key")

// Generate OAuth connection
conn := fixtures.GenerateOAuthConnection("github")
```

## CI/CD Integration

### GitHub Actions

Tests run automatically on PR and push:

```yaml
# .github/workflows/ci.yml
jobs:
  integration-tests:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16
        # ...
    steps:
      - name: Run integration tests
        run: go test -v -tags=integration ./tests/integration/...

  e2e-tests:
    runs-on: ubuntu-latest
    steps:
      - name: Install Playwright
        run: npx playwright install --with-deps

      - name: Run E2E tests
        run: npm run test:e2e
```

### Running Tests Locally (Same as CI)

```bash
# Simulate CI environment
docker-compose -f tests/docker-compose.test.yml up -d
export DATABASE_URL=postgres://gorax:gorax_test@localhost:5433/gorax_test?sslmode=disable
export REDIS_URL=localhost:6380

# Run migrations
go run cmd/migrate/main.go up

# Run tests
go test -v -tags=integration -race ./tests/integration/...
cd web && npm run test:e2e
```

## Performance Tests

### k6 Load Tests

Located in `tests/performance/k6/`:

```bash
# Install k6
brew install k6  # macOS
# or download from https://k6.io

# Run load test
k6 run tests/performance/k6/marketplace.js

# With custom options
k6 run --vus 50 --duration 30s tests/performance/k6/oauth.js
```

### Performance Benchmarks

```bash
# Go benchmarks
go test -bench=. -benchmem ./tests/integration/...

# Specific benchmark
go test -bench=BenchmarkAuditLogInsertion ./tests/integration/audit_test.go
```

## Writing New Tests

### Integration Test Template

```go
package integration

import (
    "net/http"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestMyNewFeature(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }

    ts := SetupTestServer(t)
    defer ts.Cleanup()

    // Setup
    tenantID := ts.CreateTestTenant(t, "Test Tenant")
    userID := ts.CreateTestUser(t, tenantID, "user@test.com", "user")
    headers := DefaultTestHeaders(tenantID)

    t.Run("SubTest1", func(t *testing.T) {
        // Test logic
        resp := ts.MakeRequest(t, http.MethodGet, "/api/v1/endpoint", nil, headers)
        AssertStatusCode(t, resp, http.StatusOK)

        var result map[string]interface{}
        ParseJSONResponse(t, resp, &result)
        assert.NotEmpty(t, result)
    })
}
```

### E2E Test Template

```typescript
import { test, expect } from '@playwright/test';

test.describe('My Feature', () => {
  test.beforeEach(async ({ page }) => {
    // Setup: login, navigate, etc.
    await page.goto('/');
  });

  test('user can perform action', async ({ page }) => {
    // Test steps
    await page.click('[data-testid="action-button"]');

    // Assertions
    await expect(page.locator('[data-testid="result"]')).toBeVisible();
  });
});
```

## Troubleshooting

### Common Issues

#### Tests fail with "connection refused"

**Solution**: Ensure Docker services are running:

```bash
docker-compose -f tests/docker-compose.test.yml ps
docker-compose -f tests/docker-compose.test.yml up -d
```

#### Database migration errors

**Solution**: Reset test database:

```bash
docker-compose -f tests/docker-compose.test.yml down -v
docker-compose -f tests/docker-compose.test.yml up -d
sleep 5
go run cmd/migrate/main.go up
```

#### E2E tests timeout

**Solution**: Increase timeout in `playwright.config.ts`:

```typescript
timeout: 60000, // 60 seconds
```

#### Flaky tests

**Solution**: Add retries and wait conditions:

```typescript
// E2E
test.describe.configure({ retries: 2 });
await page.waitForLoadState('networkidle');

// Integration
WaitForCondition(t, 5*time.Second, func() bool {
    return condition()
})
```

### Debug Mode

#### Integration Tests

```bash
# Verbose output
go test -v ./tests/integration/...

# Run single test
go test -v -run TestMarketplace_FullWorkflow ./tests/integration/...

# With race detector
go test -race ./tests/integration/...
```

#### E2E Tests

```bash
# UI mode (interactive)
npx playwright test --ui

# Debug mode
npx playwright test --debug

# Headed mode (see browser)
npx playwright test --headed

# Generate trace
npx playwright test --trace on
npx playwright show-trace trace.zip
```

### Viewing Logs

```bash
# Docker service logs
docker-compose -f tests/docker-compose.test.yml logs -f postgres-test

# Application logs (if running via Docker)
docker logs gorax-api-test

# Test output
go test -v ./tests/integration/... 2>&1 | tee test.log
```

## Best Practices

1. **Isolation**: Each test should be independent
2. **Cleanup**: Always use `t.Cleanup()` or `defer`
3. **Descriptive Names**: Use clear test and subtest names
4. **Fast Tests**: Keep tests fast (< 5s per test)
5. **Deterministic**: Avoid time-dependent logic
6. **Assertions**: Use meaningful assertion messages
7. **Test Data**: Use fixtures for consistent data

## Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Testify Documentation](https://github.com/stretchr/testify)
- [Playwright Documentation](https://playwright.dev/)
- [k6 Documentation](https://k6.io/docs/)

## Support

For issues or questions:
- Open an issue in the repository
- Check existing test examples
- Review CI/CD logs for failures
