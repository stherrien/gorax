// k6 Load Test: Webhook Ingestion Rate
// Tests webhook trigger performance and ingestion throughput

import http from 'k6/http';
import { check, group, sleep } from 'k6';
import { Counter, Rate, Trend } from 'k6/metrics';
import { config, getScenario } from './config.js';

// Custom metrics
const webhookIngestionDuration = new Trend('webhook_ingestion_duration', true);
const webhookProcessingDuration = new Trend('webhook_processing_duration', true);
const webhookThroughput = new Counter('webhook_throughput');
const webhookErrors = new Counter('webhook_errors');
const webhookSuccessRate = new Rate('webhook_success_rate');
const webhookValidationErrors = new Counter('webhook_validation_errors');
const webhookAuthErrors = new Counter('webhook_auth_errors');

// Test configuration
const scenario = getScenario();
export const options = {
  stages: scenario.stages || [{ duration: scenario.duration, target: scenario.vus }],
  thresholds: {
    ...config.thresholds,
    'webhook_ingestion_duration': ['p(95)<200', 'p(99)<500'],
    'webhook_success_rate': ['rate>0.99'],
  },
  tags: config.options.tags,
};

// Setup: Create webhook-triggered workflows
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

  // Create webhook-triggered workflows
  const webhookWorkflows = [];
  for (let i = 0; i < 5; i++) {
    const workflow = {
      name: `webhook-loadtest-${i}`,
      description: 'Load test webhook workflow',
      trigger: {
        type: 'webhook',
        config: {
          path: `/loadtest/webhook-${i}`,
          method: 'POST',
          validation: {
            required_headers: ['x-test-id'],
          },
        },
      },
      steps: [
        {
          id: 'validate',
          type: 'transform',
          config: {
            formula: '{{ trigger.body }}',
          },
        },
      ],
    };

    const createRes = http.post(
      `${config.baseUrl}/api/v1/workflows`,
      JSON.stringify(workflow),
      { headers }
    );

    if (createRes.status === 201) {
      const workflowData = createRes.json();
      webhookWorkflows.push({
        id: workflowData.id,
        webhookPath: workflow.trigger.config.path,
      });
    }
  }

  if (webhookWorkflows.length === 0) {
    throw new Error('Setup failed: Unable to create webhook workflows');
  }

  return { authToken, webhookWorkflows };
}

// Main test function
export default function (data) {
  // Select a random webhook for this iteration
  const workflow = data.webhookWorkflows[Math.floor(Math.random() * data.webhookWorkflows.length)];

  // Generate webhook payload
  const payload = {
    event: 'test.event',
    timestamp: new Date().toISOString(),
    data: {
      id: `event-${__VU}-${__ITER}`,
      value: Math.floor(Math.random() * 1000),
      vu: __VU,
      iteration: __ITER,
    },
  };

  const headers = {
    'Content-Type': 'application/json',
    'x-test-id': `test-${__VU}-${__ITER}`,
  };

  // Test 1: Single Webhook Trigger
  group('Single Webhook Trigger', () => {
    const startTime = Date.now();

    const webhookRes = http.post(
      `${config.baseUrl}${workflow.webhookPath}`,
      JSON.stringify(payload),
      {
        headers,
        tags: { endpoint: 'webhook_trigger' },
      }
    );

    const ingestionDuration = Date.now() - startTime;
    webhookIngestionDuration.add(ingestionDuration);

    const success = check(webhookRes, {
      'webhook accepted': (r) => r.status === 202 || r.status === 200,
      'webhook has execution ID': (r) => r.json('execution_id') !== undefined || r.json('message') !== undefined,
      'ingestion time < 200ms': () => ingestionDuration < 200,
    });

    webhookSuccessRate.add(success);
    if (success) {
      webhookThroughput.add(1);
    } else {
      webhookErrors.add(1);
      if (webhookRes.status === 400) {
        webhookValidationErrors.add(1);
      } else if (webhookRes.status === 401 || webhookRes.status === 403) {
        webhookAuthErrors.add(1);
      }
      console.error(`Webhook failed: ${webhookRes.status} - ${webhookRes.body}`);
    }
  });

  sleep(0.5);

  // Test 2: Batch Webhook Triggers
  group('Batch Webhook Triggers', () => {
    const batchSize = 5;
    const requests = [];

    for (let i = 0; i < batchSize; i++) {
      const batchPayload = {
        ...payload,
        data: {
          ...payload.data,
          batchIndex: i,
        },
      };

      requests.push({
        method: 'POST',
        url: `${config.baseUrl}${workflow.webhookPath}`,
        body: JSON.stringify(batchPayload),
        params: {
          headers,
          tags: { endpoint: 'webhook_trigger', batch: 'true' },
        },
      });
    }

    const batchStart = Date.now();
    const responses = http.batch(requests);
    const batchDuration = Date.now() - batchStart;

    const successCount = responses.filter(r => r.status === 202 || r.status === 200).length;
    const batchSuccess = successCount === batchSize;

    check(responses, {
      'all webhooks accepted': () => batchSuccess,
      'batch time < 1s': () => batchDuration < 1000,
    });

    webhookSuccessRate.add(batchSuccess);
    webhookThroughput.add(successCount);
    webhookErrors.add(batchSize - successCount);
  });

  sleep(0.5);

  // Test 3: Webhook with Different Content Types
  group('Different Content Types', () => {
    // JSON payload
    const jsonRes = http.post(
      `${config.baseUrl}${workflow.webhookPath}`,
      JSON.stringify(payload),
      {
        headers: { ...headers, 'Content-Type': 'application/json' },
        tags: { endpoint: 'webhook_trigger', content_type: 'json' },
      }
    );

    check(jsonRes, {
      'JSON webhook accepted': (r) => r.status === 202 || r.status === 200,
    });

    // Form data
    const formData = `event=${payload.event}&value=${payload.data.value}`;
    const formRes = http.post(
      `${config.baseUrl}${workflow.webhookPath}`,
      formData,
      {
        headers: { ...headers, 'Content-Type': 'application/x-www-form-urlencoded' },
        tags: { endpoint: 'webhook_trigger', content_type: 'form' },
      }
    );

    check(formRes, {
      'Form webhook accepted': (r) => r.status === 202 || r.status === 200,
    });
  });

  sleep(0.5);

  // Test 4: Webhook with Large Payload
  group('Large Payload', () => {
    const largePayload = {
      ...payload,
      data: {
        ...payload.data,
        largeArray: Array.from({ length: 100 }, (_, i) => ({
          index: i,
          value: Math.random(),
          timestamp: new Date().toISOString(),
        })),
      },
    };

    const largeStart = Date.now();
    const largeRes = http.post(
      `${config.baseUrl}${workflow.webhookPath}`,
      JSON.stringify(largePayload),
      {
        headers,
        tags: { endpoint: 'webhook_trigger', payload_size: 'large' },
      }
    );

    const largeDuration = Date.now() - largeStart;

    check(largeRes, {
      'large webhook accepted': (r) => r.status === 202 || r.status === 200,
      'large payload time < 500ms': () => largeDuration < 500,
    });
  });

  sleep(0.5);

  // Test 5: Missing Required Headers (Error Case)
  group('Validation - Missing Headers', () => {
    const invalidRes = http.post(
      `${config.baseUrl}${workflow.webhookPath}`,
      JSON.stringify(payload),
      {
        headers: { 'Content-Type': 'application/json' }, // Missing x-test-id
        tags: { endpoint: 'webhook_trigger', validation: 'missing_header' },
      }
    );

    check(invalidRes, {
      'missing header rejected': (r) => r.status === 400 || r.status === 422,
    });
  });

  sleep(0.5);

  // Test 6: High-Frequency Bursts
  group('High-Frequency Burst', () => {
    const burstSize = 10;
    const burstStart = Date.now();

    for (let i = 0; i < burstSize; i++) {
      http.post(
        `${config.baseUrl}${workflow.webhookPath}`,
        JSON.stringify({ ...payload, burstIndex: i }),
        {
          headers,
          tags: { endpoint: 'webhook_trigger', burst: 'true' },
        }
      );
    }

    const burstDuration = Date.now() - burstStart;
    const throughputRate = (burstSize / burstDuration) * 1000; // requests per second

    check(burstDuration, {
      'burst completed': () => true,
      'burst throughput > 50 req/s': () => throughputRate > 50,
    });

    webhookThroughput.add(burstSize);
  });

  sleep(1);
}

// Teardown: Clean up test workflows
export function teardown(data) {
  const headers = {
    'Authorization': `Bearer ${data.authToken}`,
  };

  // Delete test workflows
  data.webhookWorkflows.forEach(workflow => {
    http.del(`${config.baseUrl}/api/v1/workflows/${workflow.id}`, null, { headers });
  });

  console.log('Webhook trigger load test completed.');
}

// Handle test summary
export function handleSummary(data) {
  return {
    'stdout': textSummary(data, { indent: ' ', enableColors: true }),
    'webhook_trigger_results.json': JSON.stringify(data, null, 2),
  };
}

function textSummary(data, options) {
  const indent = options?.indent || '';

  let summary = '\n';
  summary += `${indent}Webhook Trigger Load Test Summary\n`;
  summary += `${indent}==================================\n\n`;

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

  // Webhook metrics
  summary += `${indent}Webhook Performance:\n`;
  summary += `${indent}  Total Webhooks: ${data.metrics.webhook_throughput?.values?.count || 0}\n`;
  summary += `${indent}  Throughput: ${(data.metrics.webhook_throughput?.values?.rate || 0).toFixed(2)} webhook/s\n`;
  summary += `${indent}  Ingestion p95: ${data.metrics.webhook_ingestion_duration?.values['p(95)']?.toFixed(2) || 0}ms\n`;
  summary += `${indent}  Ingestion p99: ${data.metrics.webhook_ingestion_duration?.values['p(99)']?.toFixed(2) || 0}ms\n`;
  summary += `${indent}  Success Rate: ${(data.metrics.webhook_success_rate?.values?.rate * 100 || 0).toFixed(2)}%\n`;
  summary += `${indent}  Errors: ${data.metrics.webhook_errors?.values?.count || 0}\n`;
  summary += `${indent}  Validation Errors: ${data.metrics.webhook_validation_errors?.values?.count || 0}\n`;
  summary += `${indent}  Auth Errors: ${data.metrics.webhook_auth_errors?.values?.count || 0}\n\n`;

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
