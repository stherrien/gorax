// k6 Load Test: Workflow CRUD Operations
// Tests workflow creation, retrieval, update, and deletion under load

import http from 'k6/http';
import { check, group, sleep } from 'k6';
import { Counter, Rate, Trend } from 'k6/metrics';
import { config, getScenario, generateTestWorkflow } from './config.js';
import { setupAuth, getAuthHeaders } from './lib/auth.js';

// Custom metrics
const workflowCreateDuration = new Trend('workflow_create_duration', true);
const workflowReadDuration = new Trend('workflow_read_duration', true);
const workflowUpdateDuration = new Trend('workflow_update_duration', true);
const workflowDeleteDuration = new Trend('workflow_delete_duration', true);
const workflowListDuration = new Trend('workflow_list_duration', true);
const workflowErrors = new Counter('workflow_errors');
const workflowSuccessRate = new Rate('workflow_success_rate');

// Test configuration
const scenario = getScenario();
export const options = {
  stages: scenario.stages || [{ duration: scenario.duration, target: scenario.vus }],
  thresholds: config.thresholds,
  tags: config.options.tags,
};

// Setup: Authenticate based on environment
export function setup() {
  return setupAuth(config);
}

// Main test function
export default function (data) {
  const headers = {
    'Content-Type': 'application/json',
    ...getAuthHeaders(data),
  };

  const workflowId = `${__VU}-${__ITER}`;
  let createdWorkflowId;

  // Test 1: Create Workflow
  group('Create Workflow', () => {
    const workflow = generateTestWorkflow(workflowId);
    const createStart = Date.now();

    const createRes = http.post(
      `${config.baseUrl}/api/v1/workflows`,
      JSON.stringify(workflow),
      {
        headers,
        tags: { endpoint: 'workflow_create' },
      }
    );

    const createDuration = Date.now() - createStart;
    workflowCreateDuration.add(createDuration);

    const createSuccess = check(createRes, {
      'workflow created successfully': (r) => r.status === 201,
      'workflow has valid ID': (r) => r.json('id') !== undefined,
      'workflow name matches': (r) => r.json('name') === workflow.name,
      'response time < 1s': () => createDuration < 1000,
    });

    workflowSuccessRate.add(createSuccess);
    if (!createSuccess) {
      workflowErrors.add(1);
      console.error(`Create failed: ${createRes.status} - ${createRes.body}`);
      return; // Skip remaining tests if create failed
    }

    createdWorkflowId = createRes.json('id');
  });

  sleep(0.5);

  // Test 2: Get Workflow by ID
  group('Get Workflow by ID', () => {
    const readStart = Date.now();

    const getRes = http.get(
      `${config.baseUrl}/api/v1/workflows/${createdWorkflowId}`,
      {
        headers,
        tags: { endpoint: 'workflow_get' },
      }
    );

    const readDuration = Date.now() - readStart;
    workflowReadDuration.add(readDuration);

    const getSuccess = check(getRes, {
      'workflow retrieved successfully': (r) => r.status === 200,
      'workflow ID matches': (r) => r.json('id') === createdWorkflowId,
      'response time < 500ms': () => readDuration < 500,
    });

    workflowSuccessRate.add(getSuccess);
    if (!getSuccess) {
      workflowErrors.add(1);
      console.error(`Get failed: ${getRes.status} - ${getRes.body}`);
    }
  });

  sleep(0.5);

  // Test 3: List Workflows
  group('List Workflows', () => {
    const listStart = Date.now();

    const listRes = http.get(
      `${config.baseUrl}/api/v1/workflows?limit=20&offset=0`,
      {
        headers,
        tags: { endpoint: 'workflow_list' },
      }
    );

    const listDuration = Date.now() - listStart;
    workflowListDuration.add(listDuration);

    const listSuccess = check(listRes, {
      'workflows listed successfully': (r) => r.status === 200,
      'response is array': (r) => Array.isArray(r.json('workflows')),
      'response time < 800ms': () => listDuration < 800,
    });

    workflowSuccessRate.add(listSuccess);
    if (!listSuccess) {
      workflowErrors.add(1);
      console.error(`List failed: ${listRes.status} - ${listRes.body}`);
    }
  });

  sleep(0.5);

  // Test 4: Update Workflow
  group('Update Workflow', () => {
    const updateData = {
      name: `Updated-${workflowId}`,
      description: 'Updated by load test',
    };

    const updateStart = Date.now();

    const updateRes = http.patch(
      `${config.baseUrl}/api/v1/workflows/${createdWorkflowId}`,
      JSON.stringify(updateData),
      {
        headers,
        tags: { endpoint: 'workflow_update' },
      }
    );

    const updateDuration = Date.now() - updateStart;
    workflowUpdateDuration.add(updateDuration);

    const updateSuccess = check(updateRes, {
      'workflow updated successfully': (r) => r.status === 200,
      'name was updated': (r) => r.json('name') === updateData.name,
      'response time < 1s': () => updateDuration < 1000,
    });

    workflowSuccessRate.add(updateSuccess);
    if (!updateSuccess) {
      workflowErrors.add(1);
      console.error(`Update failed: ${updateRes.status} - ${updateRes.body}`);
    }
  });

  sleep(0.5);

  // Test 5: Delete Workflow
  group('Delete Workflow', () => {
    const deleteStart = Date.now();

    const deleteRes = http.del(
      `${config.baseUrl}/api/v1/workflows/${createdWorkflowId}`,
      null,
      {
        headers,
        tags: { endpoint: 'workflow_delete' },
      }
    );

    const deleteDuration = Date.now() - deleteStart;
    workflowDeleteDuration.add(deleteDuration);

    const deleteSuccess = check(deleteRes, {
      'workflow deleted successfully': (r) => r.status === 204 || r.status === 200,
      'response time < 500ms': () => deleteDuration < 500,
    });

    workflowSuccessRate.add(deleteSuccess);
    if (!deleteSuccess) {
      workflowErrors.add(1);
      console.error(`Delete failed: ${deleteRes.status} - ${deleteRes.body}`);
    }
  });

  sleep(1);

  // Test 6: Verify Deletion
  group('Verify Deletion', () => {
    const verifyRes = http.get(
      `${config.baseUrl}/api/v1/workflows/${createdWorkflowId}`,
      {
        headers,
        tags: { endpoint: 'workflow_get' },
      }
    );

    check(verifyRes, {
      'workflow not found after deletion': (r) => r.status === 404,
    });
  });

  sleep(1);
}

// Teardown: Clean up any remaining test data
export function teardown(data) {
  console.log('Load test completed. Check metrics for results.');
}

// Handle test summary
export function handleSummary(data) {
  return {
    'stdout': textSummary(data, { indent: ' ', enableColors: true }),
    'workflow_api_results.json': JSON.stringify(data, null, 2),
  };
}

function textSummary(data, options) {
  const indent = options?.indent || '';
  const enableColors = options?.enableColors || false;

  let summary = '\n';
  summary += `${indent}Workflow API Load Test Summary\n`;
  summary += `${indent}================================\n\n`;

  // Test metadata
  summary += `${indent}Test Duration: ${formatDuration(data.state.testRunDurationMs)}\n`;
  summary += `${indent}VUs: ${data.metrics.vus?.values?.max || 'N/A'}\n`;
  summary += `${indent}Iterations: ${data.metrics.iterations?.values?.count || 0}\n\n`;

  // HTTP metrics
  summary += `${indent}HTTP Performance:\n`;
  summary += `${indent}  Requests: ${data.metrics.http_reqs?.values?.count || 0}\n`;
  summary += `${indent}  Failed: ${data.metrics.http_req_failed?.values?.rate * 100 || 0}%\n`;
  summary += `${indent}  Avg Duration: ${data.metrics.http_req_duration?.values?.avg?.toFixed(2) || 0}ms\n`;
  summary += `${indent}  p95 Duration: ${data.metrics.http_req_duration?.values['p(95)']?.toFixed(2) || 0}ms\n`;
  summary += `${indent}  p99 Duration: ${data.metrics.http_req_duration?.values['p(99)']?.toFixed(2) || 0}ms\n\n`;

  // Custom metrics
  summary += `${indent}Workflow Operations:\n`;
  summary += `${indent}  Create p95: ${data.metrics.workflow_create_duration?.values['p(95)']?.toFixed(2) || 0}ms\n`;
  summary += `${indent}  Read p95: ${data.metrics.workflow_read_duration?.values['p(95)']?.toFixed(2) || 0}ms\n`;
  summary += `${indent}  Update p95: ${data.metrics.workflow_update_duration?.values['p(95)']?.toFixed(2) || 0}ms\n`;
  summary += `${indent}  Delete p95: ${data.metrics.workflow_delete_duration?.values['p(95)']?.toFixed(2) || 0}ms\n`;
  summary += `${indent}  List p95: ${data.metrics.workflow_list_duration?.values['p(95)']?.toFixed(2) || 0}ms\n`;
  summary += `${indent}  Success Rate: ${(data.metrics.workflow_success_rate?.values?.rate * 100 || 0).toFixed(2)}%\n`;
  summary += `${indent}  Errors: ${data.metrics.workflow_errors?.values?.count || 0}\n\n`;

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
