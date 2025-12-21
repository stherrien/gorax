// k6 Load Test: Workflow Execution Throughput
// Tests workflow execution performance and throughput under load

import http from 'k6/http';
import { check, group, sleep } from 'k6';
import { Counter, Rate, Trend } from 'k6/metrics';
import { config, getScenario, generateTestWorkflow } from './config.js';

// Custom metrics
const executionStartDuration = new Trend('execution_start_duration', true);
const executionCompleteDuration = new Trend('execution_complete_duration', true);
const executionStatusCheckDuration = new Trend('execution_status_check_duration', true);
const executionThroughput = new Counter('execution_throughput');
const executionErrors = new Counter('execution_errors');
const executionSuccessRate = new Rate('execution_success_rate');
const executionFailureRate = new Rate('execution_failure_rate');

// Test configuration
const scenario = getScenario();
export const options = {
  stages: scenario.stages || [{ duration: scenario.duration, target: scenario.vus }],
  thresholds: {
    ...config.thresholds,
    'execution_start_duration': ['p(95)<2000', 'p(99)<3000'],
    'execution_complete_duration': ['p(95)<5000', 'p(99)<10000'],
    'execution_success_rate': ['rate>0.95'],
  },
  tags: config.options.tags,
};

// Setup: Create test workflows and authenticate
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
  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${authToken}`,
  };

  // Create test workflows for execution
  const workflows = [];
  for (let i = 0; i < 5; i++) {
    const workflow = generateTestWorkflow(`exec-test-${i}`);
    const createRes = http.post(
      `${config.baseUrl}/api/v1/workflows`,
      JSON.stringify(workflow),
      { headers }
    );

    if (createRes.status === 201) {
      workflows.push(createRes.json());
    }
  }

  if (workflows.length === 0) {
    throw new Error('Setup failed: Unable to create test workflows');
  }

  return { authToken, workflows };
}

// Main test function
export default function (data) {
  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${data.authToken}`,
  };

  // Select a random workflow for this iteration
  const workflow = data.workflows[Math.floor(Math.random() * data.workflows.length)];
  const executionInput = {
    input: {
      value: Math.floor(Math.random() * 100),
      timestamp: new Date().toISOString(),
      iteration: __ITER,
      vu: __VU,
    },
  };

  let executionId;

  // Test 1: Start Workflow Execution
  group('Start Execution', () => {
    const startTime = Date.now();

    const execRes = http.post(
      `${config.baseUrl}/api/v1/workflows/${workflow.id}/execute`,
      JSON.stringify(executionInput),
      {
        headers,
        tags: { endpoint: 'workflow_execute' },
      }
    );

    const startDuration = Date.now() - startTime;
    executionStartDuration.add(startDuration);

    const startSuccess = check(execRes, {
      'execution started successfully': (r) => r.status === 202 || r.status === 200,
      'execution has valid ID': (r) => r.json('execution_id') !== undefined,
      'start time < 2s': () => startDuration < 2000,
    });

    executionSuccessRate.add(startSuccess);
    if (!startSuccess) {
      executionErrors.add(1);
      executionFailureRate.add(1);
      console.error(`Execution start failed: ${execRes.status} - ${execRes.body}`);
      return;
    }

    executionId = execRes.json('execution_id');
    executionThroughput.add(1);
  });

  sleep(0.5);

  // Test 2: Poll Execution Status
  let completed = false;
  let attempts = 0;
  const maxAttempts = 20; // Max 10 seconds of polling

  group('Poll Execution Status', () => {
    while (!completed && attempts < maxAttempts) {
      const statusStart = Date.now();

      const statusRes = http.get(
        `${config.baseUrl}/api/v1/executions/${executionId}`,
        {
          headers,
          tags: { endpoint: 'execution_status' },
        }
      );

      const statusDuration = Date.now() - statusStart;
      executionStatusCheckDuration.add(statusDuration);

      const statusCheck = check(statusRes, {
        'status retrieved successfully': (r) => r.status === 200,
        'status check time < 500ms': () => statusDuration < 500,
      });

      if (!statusCheck) {
        executionErrors.add(1);
        break;
      }

      const status = statusRes.json('status');
      if (status === 'completed' || status === 'failed' || status === 'error') {
        completed = true;

        const executionCheck = check(statusRes, {
          'execution completed successfully': (r) => r.json('status') === 'completed',
          'execution has result': (r) => r.json('result') !== undefined,
        });

        executionSuccessRate.add(executionCheck);
        if (!executionCheck) {
          executionFailureRate.add(1);
          console.error(`Execution failed: ${statusRes.json('error')}`);
        }
      }

      attempts++;
      sleep(0.5);
    }

    if (!completed) {
      executionErrors.add(1);
      executionFailureRate.add(1);
      console.error(`Execution timeout after ${attempts} attempts`);
    }
  });

  // Test 3: Get Execution History
  group('Get Execution History', () => {
    const historyRes = http.get(
      `${config.baseUrl}/api/v1/workflows/${workflow.id}/executions?limit=10`,
      {
        headers,
        tags: { endpoint: 'execution_history' },
      }
    );

    check(historyRes, {
      'history retrieved successfully': (r) => r.status === 200,
      'history contains executions': (r) => Array.isArray(r.json('executions')) && r.json('executions').length > 0,
    });
  });

  sleep(1);

  // Test 4: Concurrent Executions
  group('Concurrent Executions', () => {
    const concurrentCount = 3;
    const requests = [];

    for (let i = 0; i < concurrentCount; i++) {
      requests.push({
        method: 'POST',
        url: `${config.baseUrl}/api/v1/workflows/${workflow.id}/execute`,
        body: JSON.stringify({
          input: {
            value: i,
            concurrent: true,
          },
        }),
        params: {
          headers,
          tags: { endpoint: 'workflow_execute', concurrent: 'true' },
        },
      });
    }

    const responses = http.batch(requests);

    const concurrentSuccess = responses.every(r => r.status === 202 || r.status === 200);
    check(responses, {
      'all concurrent executions started': () => concurrentSuccess,
    });

    if (concurrentSuccess) {
      executionThroughput.add(concurrentCount);
    } else {
      executionErrors.add(concurrentCount - responses.filter(r => r.status === 202 || r.status === 200).length);
    }
  });

  sleep(2);
}

// Teardown: Clean up test workflows
export function teardown(data) {
  const headers = {
    'Authorization': `Bearer ${data.authToken}`,
  };

  // Delete test workflows
  data.workflows.forEach(workflow => {
    http.del(`${config.baseUrl}/api/v1/workflows/${workflow.id}`, null, { headers });
  });

  console.log('Execution load test completed.');
}

// Handle test summary
export function handleSummary(data) {
  return {
    'stdout': textSummary(data, { indent: ' ', enableColors: true }),
    'execution_api_results.json': JSON.stringify(data, null, 2),
  };
}

function textSummary(data, options) {
  const indent = options?.indent || '';

  let summary = '\n';
  summary += `${indent}Workflow Execution Load Test Summary\n`;
  summary += `${indent}=====================================\n\n`;

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

  // Execution metrics
  summary += `${indent}Execution Performance:\n`;
  summary += `${indent}  Total Executions: ${data.metrics.execution_throughput?.values?.count || 0}\n`;
  summary += `${indent}  Throughput: ${(data.metrics.execution_throughput?.values?.rate || 0).toFixed(2)} exec/s\n`;
  summary += `${indent}  Start p95: ${data.metrics.execution_start_duration?.values['p(95)']?.toFixed(2) || 0}ms\n`;
  summary += `${indent}  Start p99: ${data.metrics.execution_start_duration?.values['p(99)']?.toFixed(2) || 0}ms\n`;
  summary += `${indent}  Status Check p95: ${data.metrics.execution_status_check_duration?.values['p(95)']?.toFixed(2) || 0}ms\n`;
  summary += `${indent}  Success Rate: ${(data.metrics.execution_success_rate?.values?.rate * 100 || 0).toFixed(2)}%\n`;
  summary += `${indent}  Failure Rate: ${(data.metrics.execution_failure_rate?.values?.rate * 100 || 0).toFixed(2)}%\n`;
  summary += `${indent}  Errors: ${data.metrics.execution_errors?.values?.count || 0}\n\n`;

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
