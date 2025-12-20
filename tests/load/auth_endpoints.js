// k6 Load Test: Authentication Endpoints
// Tests authentication performance under load

import http from 'k6/http';
import { check, group, sleep } from 'k6';
import { Counter, Rate, Trend } from 'k6/metrics';
import { config, getScenario } from './config.js';

// Custom metrics
const loginDuration = new Trend('login_duration', true);
const logoutDuration = new Trend('logout_duration', true);
const tokenRefreshDuration = new Trend('token_refresh_duration', true);
const tokenValidationDuration = new Trend('token_validation_duration', true);
const authThroughput = new Counter('auth_throughput');
const authErrors = new Counter('auth_errors');
const authSuccessRate = new Rate('auth_success_rate');
const invalidCredentialsAttempts = new Counter('invalid_credentials_attempts');

// Test configuration
const scenario = getScenario();
export const options = {
  stages: scenario.stages || [{ duration: scenario.duration, target: scenario.vus }],
  thresholds: {
    ...config.thresholds,
    'login_duration': ['p(95)<300', 'p(99)<500'],
    'token_refresh_duration': ['p(95)<200', 'p(99)<400'],
    'auth_success_rate': ['rate>0.95'],
  },
  tags: config.options.tags,
};

// Setup: Create test users if needed
export function setup() {
  // For this test, we'll use the existing test user
  return {
    testUser: config.testUser,
  };
}

// Main test function
export default function (data) {
  const headers = {
    'Content-Type': 'application/json',
  };

  let authToken;
  let refreshToken;

  // Test 1: User Login
  group('Login', () => {
    const loginStart = Date.now();

    const loginRes = http.post(
      `${config.baseUrl}/api/v1/auth/login`,
      JSON.stringify({
        email: data.testUser.email,
        password: data.testUser.password,
      }),
      {
        headers,
        tags: { endpoint: 'auth_login' },
      }
    );

    const loginDur = Date.now() - loginStart;
    loginDuration.add(loginDur);

    const loginSuccess = check(loginRes, {
      'login successful': (r) => r.status === 200,
      'token received': (r) => r.json('token') !== undefined,
      'login time < 300ms': () => loginDur < 300,
    });

    authSuccessRate.add(loginSuccess);
    if (loginSuccess) {
      authToken = loginRes.json('token');
      refreshToken = loginRes.json('refresh_token');
      authThroughput.add(1);
    } else {
      authErrors.add(1);
      console.error(`Login failed: ${loginRes.status} - ${loginRes.body}`);
      return; // Skip remaining tests if login failed
    }
  });

  sleep(0.5);

  // Test 2: Validate Token
  group('Token Validation', () => {
    const validateStart = Date.now();

    const validateRes = http.get(
      `${config.baseUrl}/api/v1/auth/validate`,
      {
        headers: {
          ...headers,
          'Authorization': `Bearer ${authToken}`,
        },
        tags: { endpoint: 'auth_validate' },
      }
    );

    const validateDur = Date.now() - validateStart;
    tokenValidationDuration.add(validateDur);

    const validateSuccess = check(validateRes, {
      'token valid': (r) => r.status === 200,
      'validation time < 100ms': () => validateDur < 100,
    });

    authSuccessRate.add(validateSuccess);
    if (!validateSuccess) {
      authErrors.add(1);
    }
  });

  sleep(0.5);

  // Test 3: Access Protected Endpoint
  group('Protected Endpoint Access', () => {
    const protectedRes = http.get(
      `${config.baseUrl}/api/v1/workflows`,
      {
        headers: {
          ...headers,
          'Authorization': `Bearer ${authToken}`,
        },
        tags: { endpoint: 'protected_resource' },
      }
    );

    const accessSuccess = check(protectedRes, {
      'protected endpoint accessible': (r) => r.status === 200,
      'response is valid': (r) => r.json() !== undefined,
    });

    authSuccessRate.add(accessSuccess);
    if (!accessSuccess) {
      authErrors.add(1);
    }
  });

  sleep(0.5);

  // Test 4: Token Refresh
  if (refreshToken) {
    group('Token Refresh', () => {
      const refreshStart = Date.now();

      const refreshRes = http.post(
        `${config.baseUrl}/api/v1/auth/refresh`,
        JSON.stringify({
          refresh_token: refreshToken,
        }),
        {
          headers,
          tags: { endpoint: 'auth_refresh' },
        }
      );

      const refreshDur = Date.now() - refreshStart;
      tokenRefreshDuration.add(refreshDur);

      const refreshSuccess = check(refreshRes, {
        'token refreshed': (r) => r.status === 200,
        'new token received': (r) => r.json('token') !== undefined,
        'refresh time < 200ms': () => refreshDur < 200,
      });

      authSuccessRate.add(refreshSuccess);
      if (refreshSuccess) {
        authToken = refreshRes.json('token');
      } else {
        authErrors.add(1);
      }
    });

    sleep(0.5);
  }

  // Test 5: Invalid Credentials (Error Case)
  group('Invalid Credentials', () => {
    const invalidRes = http.post(
      `${config.baseUrl}/api/v1/auth/login`,
      JSON.stringify({
        email: data.testUser.email,
        password: 'wrongpassword',
      }),
      {
        headers,
        tags: { endpoint: 'auth_login', test: 'invalid' },
      }
    );

    const invalidCheck = check(invalidRes, {
      'invalid credentials rejected': (r) => r.status === 401,
      'error message present': (r) => r.json('error') !== undefined,
    });

    invalidCredentialsAttempts.add(1);
    if (!invalidCheck) {
      authErrors.add(1);
    }
  });

  sleep(0.5);

  // Test 6: Expired/Invalid Token (Error Case)
  group('Invalid Token', () => {
    const invalidTokenRes = http.get(
      `${config.baseUrl}/api/v1/workflows`,
      {
        headers: {
          ...headers,
          'Authorization': 'Bearer invalid_token_12345',
        },
        tags: { endpoint: 'protected_resource', test: 'invalid_token' },
      }
    );

    check(invalidTokenRes, {
      'invalid token rejected': (r) => r.status === 401,
    });
  });

  sleep(0.5);

  // Test 7: Missing Token (Error Case)
  group('Missing Token', () => {
    const noTokenRes = http.get(
      `${config.baseUrl}/api/v1/workflows`,
      {
        headers,
        tags: { endpoint: 'protected_resource', test: 'no_token' },
      }
    );

    check(noTokenRes, {
      'missing token rejected': (r) => r.status === 401,
    });
  });

  sleep(0.5);

  // Test 8: Logout
  group('Logout', () => {
    const logoutStart = Date.now();

    const logoutRes = http.post(
      `${config.baseUrl}/api/v1/auth/logout`,
      null,
      {
        headers: {
          ...headers,
          'Authorization': `Bearer ${authToken}`,
        },
        tags: { endpoint: 'auth_logout' },
      }
    );

    const logoutDur = Date.now() - logoutStart;
    logoutDuration.add(logoutDur);

    const logoutSuccess = check(logoutRes, {
      'logout successful': (r) => r.status === 200 || r.status === 204,
      'logout time < 200ms': () => logoutDur < 200,
    });

    authSuccessRate.add(logoutSuccess);
    if (!logoutSuccess) {
      authErrors.add(1);
    }
  });

  sleep(0.5);

  // Test 9: Token Invalid After Logout
  group('Post-Logout Validation', () => {
    const postLogoutRes = http.get(
      `${config.baseUrl}/api/v1/workflows`,
      {
        headers: {
          ...headers,
          'Authorization': `Bearer ${authToken}`,
        },
        tags: { endpoint: 'protected_resource', test: 'post_logout' },
      }
    );

    check(postLogoutRes, {
      'token invalid after logout': (r) => r.status === 401,
    });
  });

  sleep(1);

  // Test 10: Concurrent Login Attempts
  group('Concurrent Logins', () => {
    const concurrentCount = 3;
    const requests = [];

    for (let i = 0; i < concurrentCount; i++) {
      requests.push({
        method: 'POST',
        url: `${config.baseUrl}/api/v1/auth/login`,
        body: JSON.stringify({
          email: data.testUser.email,
          password: data.testUser.password,
        }),
        params: {
          headers,
          tags: { endpoint: 'auth_login', concurrent: 'true' },
        },
      });
    }

    const responses = http.batch(requests);
    const successCount = responses.filter(r => r.status === 200).length;

    check(responses, {
      'all concurrent logins successful': () => successCount === concurrentCount,
    });

    authThroughput.add(successCount);
  });

  sleep(1);
}

// Teardown
export function teardown(data) {
  console.log('Authentication load test completed.');
}

// Handle test summary
export function handleSummary(data) {
  return {
    'stdout': textSummary(data, { indent: ' ', enableColors: true }),
    'auth_endpoints_results.json': JSON.stringify(data, null, 2),
  };
}

function textSummary(data, options) {
  const indent = options?.indent || '';

  let summary = '\n';
  summary += `${indent}Authentication Load Test Summary\n`;
  summary += `${indent}=================================\n\n`;

  // Test metadata
  summary += `${indent}Test Duration: ${formatDuration(data.state.testRunDurationMs)}\n`;
  summary += `${indent}VUs: ${data.metrics.vus?.values?.max || 'N/A'}\n`;
  summary += `${indent}Iterations: ${data.metrics.iterations?.values?.count || 0}\n\n`;

  // HTTP metrics
  summary += `${indent}HTTP Performance:\n`;
  summary += `${indent}  Requests: ${data.metrics.http_reqs?.values?.count || 0}\n`;
  summary += `${indent}  Failed: ${(data.metrics.http_req_failed?.values?.rate * 100 || 0).toFixed(2)}%\n`;
  summary += `${indent}  Avg Duration: ${data.metrics.http_req_duration?.values?.avg?.toFixed(2) || 0}ms\n`;
  summary += `${indent}  p95 Duration: ${data.metrics.http_req_duration?.values['p(95)']?.toFixed(2) || 0}ms\n\n`;

  // Authentication metrics
  summary += `${indent}Authentication Performance:\n`;
  summary += `${indent}  Total Auth Operations: ${data.metrics.auth_throughput?.values?.count || 0}\n`;
  summary += `${indent}  Login p95: ${data.metrics.login_duration?.values['p(95)']?.toFixed(2) || 0}ms\n`;
  summary += `${indent}  Login p99: ${data.metrics.login_duration?.values['p(99)']?.toFixed(2) || 0}ms\n`;
  summary += `${indent}  Token Refresh p95: ${data.metrics.token_refresh_duration?.values['p(95)']?.toFixed(2) || 0}ms\n`;
  summary += `${indent}  Token Validation p95: ${data.metrics.token_validation_duration?.values['p(95)']?.toFixed(2) || 0}ms\n`;
  summary += `${indent}  Logout p95: ${data.metrics.logout_duration?.values['p(95)']?.toFixed(2) || 0}ms\n`;
  summary += `${indent}  Success Rate: ${(data.metrics.auth_success_rate?.values?.rate * 100 || 0).toFixed(2)}%\n`;
  summary += `${indent}  Errors: ${data.metrics.auth_errors?.values?.count || 0}\n`;
  summary += `${indent}  Invalid Credential Attempts: ${data.metrics.invalid_credentials_attempts?.values?.count || 0}\n\n`;

  return summary;
}

function formatDuration(ms) {
  const seconds = Math.floor(ms / 1000);
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);

  if (hours > 0) {
    return `${hours}h ${minutes % 60}m ${seconds % 60}s`;
  } else if (minutes > 0) {
    return `${minutes}m ${seconds % 60}s`;
  } else {
    return `${seconds}s`;
  }
}
