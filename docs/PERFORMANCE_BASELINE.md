# Gorax Performance Baseline

**Date:** January 2, 2026
**Version:** dev (commit: dc519b0)
**Test Environment:** Local Development (macOS arm64)
**Backend Version:** Running on port 8181
**Test Tool:** k6 v1.4.2

## Executive Summary

Initial performance baseline attempt identified that the existing k6 load test suite requires updates to work with the current Gorax authentication architecture. The platform uses Ory Kratos for production authentication but implements DevAuth middleware for development environments, which expects tenant identification via headers rather than traditional login endpoints.

## Test Environment

### Hardware
- **Platform:** macOS (Darwin 25.2.0)
- **Architecture:** arm64 (Apple Silicon)
- **Location:** Local development machine

### Software Stack
- **Go API:** Running on localhost:8181
- **Database:** PostgreSQL (localhost:5432)
- **Redis:** localhost:6379
- **Web Server:** Vite dev server (localhost:5173)
- **Authentication:** DevAuth mode (X-Tenant-ID header)

### Configuration
```bash
APP_ENV=development
SERVER_ADDRESS=:8181
DB_HOST=localhost
DB_PORT=5432
REDIS_ADDRESS=localhost:6379
```

## Test Execution Results

### Smoke Test Run (January 2, 2026 - 17:38:56)

Attempted to run full smoke test suite with the following command:
```bash
cd tests/load
BASE_URL=http://localhost:8181 WS_URL=ws://localhost:8181 ./run_tests.sh --scenario smoke
```

#### Results Summary
| Test Suite | Status | Reason |
|------------|--------|--------|
| workflow_api | FAILED | Authentication setup expects `/api/v1/auth/login` endpoint |
| execution_api | FAILED | Authentication setup expects `/api/v1/auth/login` endpoint |
| webhook_trigger | FAILED | Authentication setup expects `/api/v1/auth/login` endpoint |
| websocket_connections | FAILED | Authentication setup incompatible |
| auth_endpoints | FAILED | Login endpoint does not exist (404) |

**Total Tests:** 5
**Passed:** 0
**Failed:** 5

### Root Cause Analysis

The k6 test suite was designed for a traditional authentication flow:
```javascript
// Current test setup (workflow_api.js:26-39)
export function setup() {
  const loginRes = http.post(`${config.baseUrl}/api/v1/auth/login`, JSON.stringify({
    email: config.testUser.email,
    password: config.testUser.password,
  }), {
    headers: { 'Content-Type': 'application/json' },
  });

  if (loginRes.status !== 200) {
    throw new Error(`Setup failed: Unable to authenticate. Status: ${loginRes.status}`);
  }

  const authToken = loginRes.json('token');
  return { authToken };
}
```

**However**, Gorax uses Ory Kratos for production authentication and DevAuth for development:

```go
// internal/api/app.go:474-481
// Authentication middleware
if a.config.Server.Env == "development" {
    // Use development auth that bypasses Kratos
    r.Use(apiMiddleware.DevAuth())
} else {
    // Use production Kratos auth
    r.Use(apiMiddleware.KratosAuth(a.config.Kratos))
}
```

In development mode, requests only need the `X-Tenant-ID` header:
```go
// internal/api/middleware/dev_auth.go:21-25
tenantID := r.Header.Get("X-Tenant-ID")
if tenantID == "" {
    // Default to test tenant for convenience
    tenantID = "00000000-0000-0000-0000-000000000001"
}
```

### Verification Tests

To confirm the API is operational, manual tests were performed:

#### Test 1: List Workflows
```bash
curl -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001" \
  http://localhost:8181/api/v1/workflows
```
**Result:** ✅ Success - `{"data":null,"limit":20,"offset":0}`

#### Test 2: Health Check
```bash
curl http://localhost:8181/health
```
**Result:** ✅ Success - `{"status":"ok","timestamp":"2026-01-02T22:37:59Z"}`

**Conclusion:** The API is functional and responding correctly. The issue is solely with the test suite's authentication approach.

## Expected Performance Targets

Based on the test suite configuration (`tests/load/README.md`):

### Local Development Targets
| Metric | Target | Acceptable | Critical |
|--------|--------|------------|----------|
| Workflow Create | < 500ms | < 1s | > 2s |
| Workflow Execute | < 1s | < 2s | > 5s |
| Webhook Ingestion | < 100ms | < 200ms | > 500ms |
| Auth/Validation | < 200ms | < 300ms | > 1s |
| WebSocket Connect | < 500ms | < 1s | > 2s |

### HTTP Thresholds
```javascript
http_req_duration: ['p(95)<500', 'p(99)<1000']  // 95% < 500ms, 99% < 1s
http_req_failed: ['rate<0.01']                   // Error rate < 1%
iteration_duration: ['p(95)<5000']               // Iteration < 5s
```

### Capacity Expectations
**Single Instance:**
- Concurrent users: 50-100
- Workflow executions/min: 1,000-2,000
- Webhook ingestion/s: 100-200
- WebSocket connections: 1,000-2,000

## Recommendations

### 1. Update k6 Test Suite (HIGH PRIORITY)

The load test suite needs to be updated to support both development and production authentication modes:

#### Option A: Detect Environment and Adapt
```javascript
export function setup() {
  const env = __ENV.APP_ENV || 'development';

  if (env === 'development') {
    // Use DevAuth - no token needed, just return tenant context
    return {
      tenantID: __ENV.TENANT_ID || '00000000-0000-0000-0000-000000000001',
      devMode: true
    };
  } else {
    // Use Kratos auth for production
    const loginRes = http.post(`${config.kratosUrl}/sessions/whoami`, ...);
    // ... Kratos authentication flow
  }
}

export default function (data) {
  const headers = data.devMode
    ? {
        'Content-Type': 'application/json',
        'X-Tenant-ID': data.tenantID,
      }
    : {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${data.authToken}`,
      };

  // ... rest of test
}
```

#### Option B: Separate Test Suites
- `tests/load/dev/` - Tests using DevAuth (X-Tenant-ID header)
- `tests/load/prod/` - Tests using Kratos authentication

### 2. Create Development-Specific Tests (IMMEDIATE)

Create simplified load tests that work with current development setup:

**File: `tests/load/dev_workflow_api.js`**
```javascript
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend, Rate } from 'k6/metrics';

const createDuration = new Trend('workflow_create_duration');
const successRate = new Rate('workflow_success_rate');

export const options = {
  vus: 10,
  duration: '1m',
};

export default function () {
  const headers = {
    'Content-Type': 'application/json',
    'X-Tenant-ID': '00000000-0000-0000-0000-000000000001',
  };

  // Test workflow creation
  const workflow = {
    name: `test-workflow-${__VU}-${__ITER}`,
    description: 'Load test workflow',
    definition: {
      nodes: [],
      edges: []
    }
  };

  const createRes = http.post(
    'http://localhost:8181/api/v1/workflows',
    JSON.stringify(workflow),
    { headers }
  );

  const success = check(createRes, {
    'status is 201': (r) => r.status === 201,
    'response time < 500ms': (r) => r.timings.duration < 500,
  });

  createDuration.add(createRes.timings.duration);
  successRate.add(success);

  sleep(1);
}
```

### 3. Document Authentication Architecture

Update `tests/load/README.md` to document:
- Development vs. Production authentication modes
- How to run tests in each environment
- Header requirements for DevAuth mode
- Kratos integration for production testing

### 4. CI/CD Integration Considerations

When integrating load tests into CI/CD:
- Use development mode tests for PR validation
- Run production-mode tests against staging environment
- Set up test tenant in database before running tests
- Configure appropriate thresholds per environment

### 5. Go Benchmarks (ALTERNATIVE APPROACH)

Until k6 tests are updated, use Go benchmarks to establish performance baselines:

```bash
# Benchmark critical paths
go test -bench=. -benchmem ./internal/executor/
go test -bench=. -benchmem ./internal/workflow/
go test -bench=. -benchmem ./internal/webhook/

# Save baseline
go test -bench=. -benchmem ./internal/executor/ > baseline_executor.txt
go test -bench=. -benchmem ./internal/workflow/ > baseline_workflow.txt
```

Benefits:
- Tests business logic directly without HTTP overhead
- No authentication setup required
- Precise measurements of core operations
- Easy to run in CI/CD

## Current System Status

### API Endpoints Verified
- ✅ Health check: `GET /health` - Responding
- ✅ Workflow list: `GET /api/v1/workflows` - Responding (requires X-Tenant-ID header)
- ❌ Auth endpoints: No `/api/v1/auth/login` (uses Kratos in production)

### Services Running
- ✅ API Server (port 8181)
- ✅ PostgreSQL (port 5432)
- ✅ Redis (port 6379)
- ✅ Web Dev Server (port 5173)
- ❌ Kratos (not running in local dev)

## Next Steps

### Immediate (Before Next Performance Test)
1. **Update test suite** to support DevAuth mode
2. **Create dev-specific tests** that can run without Kratos
3. **Document the update** in `tests/load/README.md`
4. **Verify tests pass** in development environment

### Short Term (Next Sprint)
1. **Run full test suite** with updated tests
2. **Establish baseline metrics** for all endpoints
3. **Set up monitoring** for key performance indicators
4. **Integrate with CI/CD** for regression detection

### Long Term (Production Ready)
1. **Create production test suite** using Kratos authentication
2. **Set up test environment** in staging with production-like config
3. **Establish production baselines** with scaled infrastructure
4. **Implement automated alerting** for performance degradation

## Files Generated

Test run generated the following result files:
```
tests/load/results/
├── workflow_api_20260102_173856.json
├── workflow_api_20260102_173856_summary.txt
├── execution_api_20260102_173856.json
├── execution_api_20260102_173856_summary.txt
├── webhook_trigger_20260102_173856.json
├── webhook_trigger_20260102_173856_summary.txt
├── websocket_connections_20260102_173856.json
├── websocket_connections_20260102_173856_summary.txt
├── auth_endpoints_20260102_173856.json
├── auth_endpoints_20260102_173856_summary.txt
└── combined_report_20260102_173856.html
```

**Note:** These results show authentication failures and cannot be used as performance baselines.

## Test Infrastructure Issues Identified

During the baseline attempt, multiple test infrastructure issues were discovered:

### 1. Authentication Mismatch (HIGH PRIORITY)
The k6 test suite expects traditional JWT login endpoints (`/api/v1/auth/login`), but Gorax uses:
- Production: Ory Kratos authentication
- Development: DevAuth middleware with X-Tenant-ID header

**Impact:** All existing tests fail at the setup phase.

### 2. Workflow Definition Schema (MEDIUM PRIORITY)
The test workflow definitions don't match the current API schema. The API validation requires:
- At least one properly configured trigger
- Specific node type definitions
- Proper edge connections

**Impact:** Even with fixed authentication, workflow creation tests would fail.

### 3. Port Configuration (RESOLVED)
Tests default to port 8080, but API runs on port 8181.
**Resolution:** Updated test runs to use BASE_URL=http://localhost:8181

## Immediate Actions Taken

1. ✅ **Documented Authentication Issue** - Created `LOAD_TEST_UPDATES_NEEDED.md` with detailed fix instructions
2. ✅ **Created Simple Dev Test** - Built `simple_dev_test.js` using DevAuth headers
3. ✅ **Updated Documentation** - This baseline document captures current state
4. ✅ **Verified API Health** - Confirmed all services are running and responding

## Test Infrastructure Roadmap

### Phase 1: Fix Development Tests (2-4 hours)
- Update `config.js` to support DevAuth mode
- Create `auth.js` helper for authentication abstraction
- Fix workflow definitions to match current schema
- Update all 5 test files to use new auth helper
- Verify all tests pass in development mode

### Phase 2: Establish Baseline (1-2 hours)
- Run full smoke test suite
- Run load test suite
- Document baseline metrics
- Compare against performance targets

### Phase 3: Production Tests (4-8 hours)
- Implement Kratos authentication flow
- Test against staging environment
- Document production baselines
- Set up monitoring and alerting

## Alternative: Go Benchmarks

As an immediate alternative to k6 tests, Go benchmarks can establish performance baselines for core business logic:

```bash
# Benchmark key components
go test -bench=. -benchmem ./internal/executor/ > baseline_executor.txt
go test -bench=. -benchmem ./internal/workflow/ > baseline_workflow.txt
go test -bench=. -benchmem ./internal/webhook/ > baseline_webhook.txt
```

**Benefits:**
- No authentication setup required
- Tests business logic directly
- Precise measurements
- Easy to run in CI/CD
- No HTTP/network overhead

**Limitations:**
- Doesn't test HTTP layer
- Doesn't test end-to-end flows
- Doesn't test WebSocket connections
- Doesn't measure real-world API performance

## Conclusion

The Gorax platform is operational and healthy. The API responds correctly to authenticated requests using the DevAuth middleware. However, the existing load test suite requires updates to work with the current authentication architecture.

**Key Findings:**
1. ✅ API is healthy and responding on port 8181
2. ✅ DevAuth middleware works correctly
3. ❌ k6 tests incompatible with current auth architecture
4. ❌ Workflow definitions in tests don't match current schema
5. ⚠️ No performance baseline can be established until tests are fixed

**Recommended Actions:**
1. **IMMEDIATE:** Update k6 test suite to support DevAuth mode
2. **IMMEDIATE:** Fix workflow definitions in test suite
3. **SHORT-TERM:** Establish baseline with updated tests
4. **LONG-TERM:** Implement Kratos authentication tests for production

**Alternative Approach:**
Run Go benchmarks on core components to establish business logic performance baselines while k6 tests are being updated.

## References

- **Test Suite:** `/Users/shawntherrien/Projects/gorax/tests/load/`
- **DevAuth Implementation:** `/Users/shawntherrien/Projects/gorax/internal/api/middleware/dev_auth.go`
- **API Setup:** `/Users/shawntherrien/Projects/gorax/internal/api/app.go`
- **Test Documentation:** `/Users/shawntherrien/Projects/gorax/tests/load/README.md`
- **k6 Documentation:** https://k6.io/docs/

---

**Baseline Status:** ⚠️ INCOMPLETE - Test suite requires updates before baseline can be established

**Next Baseline Date:** TBD (after test suite updates)
