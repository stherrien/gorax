# Load Test Suite Updates Required

## Issue Summary

The k6 load test suite currently fails because it expects traditional JWT-based authentication with `/api/v1/auth/login` endpoints. However, Gorax uses:
- **Production:** Ory Kratos for authentication
- **Development:** DevAuth middleware with `X-Tenant-ID` header

## Current Failures

All 5 test suites fail during setup:
```
Error: Setup failed: Unable to authenticate. Status: 404
```

The tests attempt to authenticate via:
```javascript
const loginRes = http.post(`${config.baseUrl}/api/v1/auth/login`, JSON.stringify({
  email: config.testUser.email,
  password: config.testUser.password,
}));
```

This endpoint does not exist in the current architecture.

## Required Changes

### 1. Update `config.js`

Add authentication mode detection:

```javascript
export const config = {
  // ... existing config ...

  // Authentication configuration
  auth: {
    mode: __ENV.AUTH_MODE || 'dev', // 'dev' or 'kratos'
    devTenantID: __ENV.DEV_TENANT_ID || '00000000-0000-0000-0000-000000000001',
    devUserID: __ENV.DEV_USER_ID || '00000000-0000-0000-0000-000000000002',
    kratosPublicURL: __ENV.KRATOS_PUBLIC_URL || 'http://localhost:4433',
  },
};
```

### 2. Create Authentication Helper

**New File:** `tests/load/auth.js`

```javascript
import http from 'k6/http';
import { config } from './config.js';

// Setup authentication based on environment
export function setupAuth() {
  if (config.auth.mode === 'dev') {
    return {
      mode: 'dev',
      tenantID: config.auth.devTenantID,
      userID: config.auth.devUserID,
    };
  } else if (config.auth.mode === 'kratos') {
    // Kratos authentication flow
    // This would integrate with Ory Kratos session management
    return {
      mode: 'kratos',
      sessionToken: authenticateWithKratos(),
    };
  }

  throw new Error(`Unknown auth mode: ${config.auth.mode}`);
}

// Get headers for authenticated requests
export function getAuthHeaders(authData) {
  if (authData.mode === 'dev') {
    return {
      'X-Tenant-ID': authData.tenantID,
      'X-User-ID': authData.userID,
    };
  } else if (authData.mode === 'kratos') {
    return {
      'Authorization': `Bearer ${authData.sessionToken}`,
    };
  }

  return {};
}

function authenticateWithKratos() {
  // TODO: Implement Kratos authentication
  // This would involve:
  // 1. POST to Kratos login flow initialization
  // 2. Submit credentials
  // 3. Extract session token
  throw new Error('Kratos authentication not yet implemented');
}
```

### 3. Update Each Test File

#### Before (workflow_api.js):
```javascript
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

export default function (data) {
  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${data.authToken}`,
  };
  // ... rest of test
}
```

#### After (workflow_api.js):
```javascript
import { setupAuth, getAuthHeaders } from './auth.js';

export function setup() {
  return setupAuth();
}

export default function (data) {
  const headers = {
    'Content-Type': 'application/json',
    ...getAuthHeaders(data),
  };
  // ... rest of test
}
```

### 4. Update `run_tests.sh`

Add environment variable for auth mode:

```bash
# Configuration
BASE_URL="${BASE_URL:-http://localhost:8181}"  # Updated default port
WS_URL="${WS_URL:-ws://localhost:8181}"        # Updated default port
AUTH_MODE="${AUTH_MODE:-dev}"                   # NEW: Auth mode
DEV_TENANT_ID="${DEV_TENANT_ID:-00000000-0000-0000-0000-000000000001}"  # NEW
TEST_USER_EMAIL="${TEST_USER_EMAIL:-loadtest@example.com}"
TEST_USER_PASSWORD="${TEST_USER_PASSWORD:-loadtest123}"
SCENARIO="${SCENARIO:-load}"
OUTPUT_DIR="${OUTPUT_DIR:-./results}"

# Export environment variables for k6
export BASE_URL
export WS_URL
export AUTH_MODE          # NEW
export DEV_TENANT_ID      # NEW
export TEST_USER_EMAIL
export TEST_USER_PASSWORD
export SCENARIO
```

### 5. Update Test Files

Files that need updates:
- ✅ `config.js` - Add auth configuration
- ✅ Create `auth.js` - Authentication helper
- ✅ `workflow_api.js` - Use new auth helper
- ✅ `execution_api.js` - Use new auth helper
- ✅ `webhook_trigger.js` - Use new auth helper
- ✅ `websocket_connections.js` - Use new auth helper
- ✅ `auth_endpoints.js` - Rewrite for Kratos (or skip in dev mode)
- ✅ `run_tests.sh` - Add auth mode configuration

## Implementation Priority

### Phase 1: Development Mode Support (IMMEDIATE)
1. Create `auth.js` helper with DevAuth support
2. Update all test files to use new helper
3. Update `run_tests.sh` with new defaults
4. Test and verify all tests pass in development

### Phase 2: Kratos Support (FUTURE)
1. Implement Kratos authentication in `auth.js`
2. Document Kratos setup requirements
3. Test against staging environment with Kratos
4. Update CI/CD pipelines

### Phase 3: Enhancement (OPTIONAL)
1. Add automatic mode detection (check for /health endpoint metadata)
2. Create separate test suites for different environments
3. Add performance comparison between dev and production auth overhead

## Quick Fix for Immediate Testing

For immediate performance testing, create a simplified test:

**File:** `tests/load/simple_workflow_test.js`

```javascript
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend, Rate } from 'k6/metrics';

const createDuration = new Trend('workflow_create_duration');
const successRate = new Rate('workflow_success_rate');

export const options = {
  stages: [
    { duration: '30s', target: 10 },
    { duration: '1m', target: 10 },
    { duration: '30s', target: 0 },
  ],
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8181';
const TENANT_ID = __ENV.TENANT_ID || '00000000-0000-0000-0000-000000000001';

export default function () {
  const headers = {
    'Content-Type': 'application/json',
    'X-Tenant-ID': TENANT_ID,
  };

  const workflow = {
    name: `loadtest-${__VU}-${__ITER}`,
    description: 'Load test workflow',
    definition: {
      nodes: [
        {
          id: 'start',
          type: 'trigger',
          config: { type: 'manual' }
        }
      ],
      edges: []
    }
  };

  const res = http.post(
    `${BASE_URL}/api/v1/workflows`,
    JSON.stringify(workflow),
    { headers }
  );

  const success = check(res, {
    'status is 201': (r) => r.status === 201,
    'response time < 1s': (r) => r.timings.duration < 1000,
  });

  createDuration.add(res.timings.duration);
  successRate.add(success);

  sleep(1);
}
```

Run with:
```bash
k6 run -e BASE_URL=http://localhost:8181 simple_workflow_test.js
```

## Testing After Updates

After implementing changes, verify with:

```bash
# Development mode (default)
cd tests/load
./run_tests.sh --scenario smoke

# Production mode (when Kratos is available)
AUTH_MODE=kratos ./run_tests.sh --scenario smoke --url https://staging.gorax.io
```

## Documentation Updates

After implementing changes, update:
1. `tests/load/README.md` - Document auth modes
2. `tests/load/QUICK_START.md` - Update examples
3. `docs/PERFORMANCE_BASELINE.md` - Record baseline with working tests
4. CI/CD documentation - Add auth mode configuration

## Related Files

- `/Users/shawntherrien/Projects/gorax/internal/api/middleware/dev_auth.go` - DevAuth implementation
- `/Users/shawntherrien/Projects/gorax/internal/api/middleware/auth.go` - Kratos auth implementation
- `/Users/shawntherrien/Projects/gorax/internal/api/app.go:474-481` - Auth middleware selection

## Estimated Effort

- **Phase 1 (Dev Mode):** 2-4 hours
- **Phase 2 (Kratos):** 4-8 hours
- **Phase 3 (Enhancement):** 2-4 hours
- **Total:** 8-16 hours

## Success Criteria

- ✅ All 5 test suites pass in development mode
- ✅ Performance baselines established for local development
- ✅ Tests run successfully in CI/CD
- ✅ Documentation updated and comprehensive
- ✅ Kratos mode works in staging environment

## Questions for Team

1. Is Kratos integration a near-term priority, or should we focus on dev mode only?
2. Should we maintain separate test suites for dev/prod, or use dynamic auth mode?
3. What are the performance SLAs we need to test against?
4. Should load tests run in CI/CD on every PR, or scheduled?

---

**Status:** 🔴 BLOCKED - Tests cannot run until auth updates are implemented

**Owner:** TBD

**Target Completion:** TBD
