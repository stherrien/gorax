# Quick Fix Checklist for k6 Load Tests

This checklist provides step-by-step instructions to fix the k6 load test suite so it works with Gorax's DevAuth authentication.

## Prerequisites
- ✅ k6 installed (`brew install k6`)
- ✅ Gorax API running on localhost:8181
- ✅ Basic understanding of JavaScript and k6

## Step-by-Step Fix

### Step 1: Create Authentication Helper (30 min)

**File:** `tests/load/auth.js` (NEW)

```javascript
import http from 'k6/http';
import { config } from './config.js';

// Setup authentication based on environment
export function setupAuth() {
  const mode = config.auth.mode;

  if (mode === 'dev') {
    console.log('Using DevAuth mode');
    return {
      mode: 'dev',
      tenantID: config.auth.devTenantID,
      userID: config.auth.devUserID,
    };
  } else if (mode === 'kratos') {
    console.log('Using Kratos auth mode');
    // TODO: Implement Kratos authentication
    throw new Error('Kratos auth not yet implemented');
  }

  throw new Error(`Unknown auth mode: ${mode}`);
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
```

### Step 2: Update Configuration (15 min)

**File:** `tests/load/config.js`

Add this section after line 13:

```javascript
  // Authentication configuration
  auth: {
    mode: __ENV.AUTH_MODE || 'dev', // 'dev' or 'kratos'
    devTenantID: __ENV.DEV_TENANT_ID || '00000000-0000-0000-0000-000000000001',
    devUserID: __ENV.DEV_USER_ID || '00000000-0000-0000-0000-000000000002',
    kratosPublicURL: __ENV.KRATOS_PUBLIC_URL || 'http://localhost:4433',
  },
```

Update baseUrl default (line 6):
```javascript
  baseUrl: __ENV.BASE_URL || 'http://localhost:8181', // Changed from 8080
```

Update wsUrl default (line 7):
```javascript
  wsUrl: __ENV.WS_URL || 'ws://localhost:8181', // Changed from 8080
```

### Step 3: Update workflow_api.js (10 min)

**File:** `tests/load/workflow_api.js`

**Find and replace:**

OLD (lines 1-40):
```javascript
import http from 'k6/http';
import { check, group, sleep } from 'k6';
import { Counter, Rate, Trend } from 'k6/metrics';
import { config, getScenario, generateTestWorkflow } from './config.js';
```

NEW:
```javascript
import http from 'k6/http';
import { check, group, sleep } from 'k6';
import { Counter, Rate, Trend } from 'k6/metrics';
import { config, getScenario, generateTestWorkflow } from './config.js';
import { setupAuth, getAuthHeaders } from './auth.js';
```

**Find and replace:**

OLD (lines 25-40):
```javascript
// Setup: Create test user and authenticate
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

NEW:
```javascript
// Setup: Authenticate based on environment
export function setup() {
  return setupAuth();
}
```

**Find and replace:**

OLD (line 43-47):
```javascript
export default function (data) {
  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${data.authToken}`,
  };
```

NEW:
```javascript
export default function (data) {
  const headers = {
    'Content-Type': 'application/json',
    ...getAuthHeaders(data),
  };
```

### Step 4: Update execution_api.js (10 min)

**File:** `tests/load/execution_api.js`

Apply the same changes as Step 3:
1. Add `import { setupAuth, getAuthHeaders } from './auth.js';`
2. Replace setup() function
3. Update headers in default function

### Step 5: Update webhook_trigger.js (10 min)

**File:** `tests/load/webhook_trigger.js`

Apply the same changes as Step 3:
1. Add import statement
2. Replace setup() function
3. Update headers in default function

### Step 6: Update websocket_connections.js (10 min)

**File:** `tests/load/websocket_connections.js`

Apply the same changes as Step 3:
1. Add import statement
2. Replace setup() function
3. Update headers in default function

### Step 7: Update auth_endpoints.js (15 min)

**File:** `tests/load/auth_endpoints.js`

**Option A: Skip in dev mode (Quick)**
```javascript
import { setupAuth } from './auth.js';

export function setup() {
  const authData = setupAuth();
  if (authData.mode === 'dev') {
    console.log('Auth endpoints test skipped in dev mode');
    return { skip: true };
  }
  return authData;
}

export default function (data) {
  if (data.skip) {
    return; // Skip test in dev mode
  }
  // ... original test logic for Kratos
}
```

**Option B: Test DevAuth (Complete)**
- Rewrite to test X-Tenant-ID header validation
- Test authenticated vs unauthenticated requests
- Test invalid tenant IDs

### Step 8: Update run_tests.sh (10 min)

**File:** `tests/load/run_tests.sh`

Update configuration section (lines 16-22):

```bash
# Configuration
BASE_URL="${BASE_URL:-http://localhost:8181}"  # Changed from 8080
WS_URL="${WS_URL:-ws://localhost:8181}"        # Changed from 8080
AUTH_MODE="${AUTH_MODE:-dev}"                   # NEW
DEV_TENANT_ID="${DEV_TENANT_ID:-00000000-0000-0000-0000-000000000001}"  # NEW
TEST_USER_EMAIL="${TEST_USER_EMAIL:-loadtest@example.com}"
TEST_USER_PASSWORD="${TEST_USER_PASSWORD:-loadtest123}"
SCENARIO="${SCENARIO:-load}"
OUTPUT_DIR="${OUTPUT_DIR:-./results}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
```

Update export section (lines 54-60):

```bash
    # Export environment variables for k6
    export BASE_URL
    export WS_URL
    export AUTH_MODE          # NEW
    export DEV_TENANT_ID      # NEW
    export TEST_USER_EMAIL
    export TEST_USER_PASSWORD
    export SCENARIO
    export TEST_ENV="${TEST_ENV:-local}"
```

### Step 9: Test Your Changes (15 min)

```bash
cd /Users/shawntherrien/Projects/gorax/tests/load

# Test 1: Verify config loads
k6 run --no-setup --no-teardown --iterations 1 workflow_api.js

# Test 2: Run smoke test (single test)
./run_tests.sh workflow --scenario smoke

# Test 3: Run full smoke test (all tests)
./run_tests.sh --scenario smoke

# Test 4: If all pass, run load test
./run_tests.sh --scenario load
```

### Step 10: Verify Results (5 min)

Check that:
- ✅ No authentication errors (404 on login endpoint)
- ✅ Tests create workflows successfully
- ✅ Success rate > 95%
- ✅ Latencies within thresholds
- ✅ Results files generated in `./results/`

## Common Issues

### Issue: "workflow must have at least one trigger"
**Cause:** Workflow definition schema mismatch
**Fix:** Use proper trigger definition in `generateTestWorkflow()`:

```javascript
export function generateTestWorkflow(id) {
  return {
    name: `${config.testData.workflowNamePrefix}${id}`,
    description: `Load test workflow ${id}`,
    trigger: {
      type: 'manual', // or 'webhook', 'schedule'
      config: {}
    },
    definition: {
      nodes: [
        {
          id: 'trigger-1',
          type: 'trigger',
          data: { type: 'manual', config: {} }
        }
      ],
      edges: []
    }
  };
}
```

### Issue: "Connection refused" or "ECONNREFUSED"
**Cause:** API not running
**Fix:**
```bash
# In separate terminal
cd /Users/shawntherrien/Projects/gorax
make dev-start
# Or: make run-api-dev
```

### Issue: Tests timeout
**Cause:** API slow to respond or database connection issues
**Fix:**
1. Check API logs
2. Verify database is running
3. Check for connection pool exhaustion
4. Reduce VUs in test scenario

## Verification Checklist

After completing all steps:

- [ ] `auth.js` file created and working
- [ ] `config.js` updated with auth section
- [ ] `workflow_api.js` uses new auth helper
- [ ] `execution_api.js` uses new auth helper
- [ ] `webhook_trigger.js` uses new auth helper
- [ ] `websocket_connections.js` uses new auth helper
- [ ] `auth_endpoints.js` updated or skips in dev mode
- [ ] `run_tests.sh` exports AUTH_MODE and DEV_TENANT_ID
- [ ] Smoke test runs without authentication errors
- [ ] Load test completes successfully
- [ ] Results files generated with metrics
- [ ] Success rate > 95%
- [ ] Latencies within acceptable thresholds
- [ ] Documentation updated (README.md)

## Time Estimate

| Step | Time | Cumulative |
|------|------|------------|
| Step 1 (auth.js) | 30 min | 30 min |
| Step 2 (config.js) | 15 min | 45 min |
| Step 3 (workflow_api) | 10 min | 55 min |
| Step 4 (execution_api) | 10 min | 1h 5m |
| Step 5 (webhook_trigger) | 10 min | 1h 15m |
| Step 6 (websocket) | 10 min | 1h 25m |
| Step 7 (auth_endpoints) | 15 min | 1h 40m |
| Step 8 (run_tests.sh) | 10 min | 1h 50m |
| Step 9 (testing) | 15 min | 2h 5m |
| Step 10 (verification) | 5 min | 2h 10m |
| **Total** | | **~2 hours** |

## Success Criteria

When complete:
- ✅ All 5 tests pass in development mode
- ✅ No authentication failures
- ✅ Baseline metrics collected
- ✅ Results documented in `PERFORMANCE_BASELINE.md`
- ✅ Tests can run in CI/CD

## Next Steps After Fix

1. Run full load test suite
2. Document baseline metrics
3. Update PERFORMANCE_BASELINE.md
4. Set up CI/CD integration
5. Plan Kratos authentication implementation

## Need Help?

- Review: `/docs/LOAD_TEST_SUMMARY.md`
- Details: `/tests/load/LOAD_TEST_UPDATES_NEEDED.md`
- Example: `/tests/load/simple_dev_test.js`

---

**Estimated Time:** 2 hours
**Difficulty:** Easy
**Prerequisites:** JavaScript, k6 basics
