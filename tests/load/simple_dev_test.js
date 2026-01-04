// Simple Development Load Test for Gorax
// Tests basic API operations using DevAuth (X-Tenant-ID header)
//
// Usage:
//   k6 run simple_dev_test.js
//   k6 run -e BASE_URL=http://localhost:8181 simple_dev_test.js

import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Trend, Rate, Counter } from 'k6/metrics';

// Custom metrics
const workflowCreateDuration = new Trend('workflow_create_duration');
const workflowListDuration = new Trend('workflow_list_duration');
const workflowGetDuration = new Trend('workflow_get_duration');
const workflowDeleteDuration = new Trend('workflow_delete_duration');
const successRate = new Rate('success_rate');
const errorCount = new Counter('errors');

// Configuration
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8181';
const TENANT_ID = __ENV.TENANT_ID || '00000000-0000-0000-0000-000000000001';

// Test options
export const options = {
  stages: [
    { duration: '30s', target: 5 },   // Ramp up to 5 VUs
    { duration: '1m', target: 5 },    // Stay at 5 VUs
    { duration: '30s', target: 0 },   // Ramp down
  ],
  thresholds: {
    'workflow_create_duration': ['p(95)<1000'],  // 95% < 1s
    'workflow_list_duration': ['p(95)<500'],     // 95% < 500ms
    'workflow_get_duration': ['p(95)<500'],      // 95% < 500ms
    'workflow_delete_duration': ['p(95)<500'],   // 95% < 500ms
    'success_rate': ['rate>0.95'],               // 95% success rate
    'http_req_duration': ['p(95)<1000'],         // 95% < 1s
    'http_req_failed': ['rate<0.01'],            // < 1% errors
  },
};

// Helper function to get auth headers
function getHeaders() {
  return {
    'Content-Type': 'application/json',
    'X-Tenant-ID': TENANT_ID,
  };
}

// Setup function
export function setup() {
  console.log('Testing Gorax API at:', BASE_URL);
  console.log('Using Tenant ID:', TENANT_ID);

  // Verify health endpoint
  const health = http.get(`${BASE_URL}/health`);
  if (health.status !== 200) {
    throw new Error(`Health check failed: ${health.status}`);
  }

  console.log('API is healthy, starting load test...');
  return { baseUrl: BASE_URL };
}

// Main test function
export default function (data) {
  const headers = getHeaders();
  let workflowID = null;

  // Test 1: Create Workflow
  group('Create Workflow', () => {
    const workflow = {
      name: `loadtest-wf-${__VU}-${__ITER}`,
      description: `Load test workflow from VU ${__VU} iteration ${__ITER}`,
      definition: {
        nodes: [
          {
            id: 'start',
            type: 'trigger',
            position: { x: 0, y: 0 },
            data: {
              label: 'Start',
              type: 'manual',
              config: {}
            }
          },
          {
            id: 'action1',
            type: 'action',
            position: { x: 200, y: 0 },
            data: {
              label: 'Log Message',
              type: 'log',
              config: {
                message: 'Hello from load test'
              }
            }
          }
        ],
        edges: [
          {
            id: 'e1',
            source: 'start',
            target: 'action1'
          }
        ]
      }
    };

    const createRes = http.post(
      `${data.baseUrl}/api/v1/workflows`,
      JSON.stringify(workflow),
      { headers }
    );

    const createSuccess = check(createRes, {
      'create: status is 201': (r) => r.status === 201,
      'create: has workflow ID': (r) => {
        if (r.status === 201) {
          try {
            const body = JSON.parse(r.body);
            return body.id !== undefined;
          } catch (e) {
            return false;
          }
        }
        return false;
      },
      'create: response time < 1s': (r) => r.timings.duration < 1000,
    });

    workflowCreateDuration.add(createRes.timings.duration);
    successRate.add(createSuccess);

    if (!createSuccess) {
      errorCount.add(1);
      console.error(`Create failed: ${createRes.status} - ${createRes.body}`);
      return; // Skip remaining tests if create failed
    }

    // Extract workflow ID for subsequent tests
    try {
      const body = JSON.parse(createRes.body);
      workflowID = body.id;
    } catch (e) {
      console.error('Failed to parse create response:', e);
      errorCount.add(1);
      return;
    }
  });

  if (!workflowID) {
    return; // Skip if we don't have a workflow ID
  }

  sleep(0.5); // Small delay between operations

  // Test 2: List Workflows
  group('List Workflows', () => {
    const listRes = http.get(
      `${data.baseUrl}/api/v1/workflows`,
      { headers }
    );

    const listSuccess = check(listRes, {
      'list: status is 200': (r) => r.status === 200,
      'list: returns array': (r) => {
        try {
          const body = JSON.parse(r.body);
          return body.data !== undefined;
        } catch (e) {
          return false;
        }
      },
      'list: response time < 500ms': (r) => r.timings.duration < 500,
    });

    workflowListDuration.add(listRes.timings.duration);
    successRate.add(listSuccess);

    if (!listSuccess) {
      errorCount.add(1);
      console.error(`List failed: ${listRes.status} - ${listRes.body}`);
    }
  });

  sleep(0.5);

  // Test 3: Get Specific Workflow
  group('Get Workflow', () => {
    const getRes = http.get(
      `${data.baseUrl}/api/v1/workflows/${workflowID}`,
      { headers }
    );

    const getSuccess = check(getRes, {
      'get: status is 200': (r) => r.status === 200,
      'get: returns workflow': (r) => {
        try {
          const body = JSON.parse(r.body);
          return body.id === workflowID;
        } catch (e) {
          return false;
        }
      },
      'get: response time < 500ms': (r) => r.timings.duration < 500,
    });

    workflowGetDuration.add(getRes.timings.duration);
    successRate.add(getSuccess);

    if (!getSuccess) {
      errorCount.add(1);
      console.error(`Get failed: ${getRes.status} - ${getRes.body}`);
    }
  });

  sleep(0.5);

  // Test 4: Delete Workflow
  group('Delete Workflow', () => {
    const deleteRes = http.del(
      `${data.baseUrl}/api/v1/workflows/${workflowID}`,
      null,
      { headers }
    );

    const deleteSuccess = check(deleteRes, {
      'delete: status is 204 or 200': (r) => r.status === 204 || r.status === 200,
      'delete: response time < 500ms': (r) => r.timings.duration < 500,
    });

    workflowDeleteDuration.add(deleteRes.timings.duration);
    successRate.add(deleteSuccess);

    if (!deleteSuccess) {
      errorCount.add(1);
      console.error(`Delete failed: ${deleteRes.status} - ${deleteRes.body}`);
    }
  });

  sleep(1); // Think time between iterations
}

// Teardown function
export function teardown(data) {
  console.log('\n=== Load Test Complete ===');
  console.log(`Tested against: ${data.baseUrl}`);
}

// Handle summary for display
export function handleSummary(data) {
  const createP95 = data.metrics.workflow_create_duration?.values['p(95)'] || 0;
  const listP95 = data.metrics.workflow_list_duration?.values['p(95)'] || 0;
  const getP95 = data.metrics.workflow_get_duration?.values['p(95)'] || 0;
  const deleteP95 = data.metrics.workflow_delete_duration?.values['p(95)'] || 0;
  const successPct = (data.metrics.success_rate?.values.rate || 0) * 100;
  const totalErrors = data.metrics.errors?.values.count || 0;

  console.log('\n╔══════════════════════════════════════════════════════════╗');
  console.log('║         Gorax Development Load Test Summary             ║');
  console.log('╚══════════════════════════════════════════════════════════╝');
  console.log('\nPerformance Metrics:');
  console.log(`  Workflow Create p95:  ${createP95.toFixed(2)}ms`);
  console.log(`  Workflow List p95:    ${listP95.toFixed(2)}ms`);
  console.log(`  Workflow Get p95:     ${getP95.toFixed(2)}ms`);
  console.log(`  Workflow Delete p95:  ${deleteP95.toFixed(2)}ms`);
  console.log(`\nReliability:`);
  console.log(`  Success Rate:         ${successPct.toFixed(2)}%`);
  console.log(`  Total Errors:         ${totalErrors}`);
  console.log('\n');

  return {
    'stdout': textSummary(data, { indent: ' ', enableColors: true }),
  };
}

// Helper to generate text summary
function textSummary(data, opts) {
  return JSON.stringify(data, null, 2);
}
