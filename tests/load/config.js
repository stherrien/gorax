// k6 Load Test Configuration for Gorax Platform
// Configure target URLs, virtual users, and test scenarios

export const config = {
  // Environment configuration
  baseUrl: __ENV.BASE_URL || 'http://localhost:8181',  // Updated default port
  wsUrl: __ENV.WS_URL || 'ws://localhost:8181',        // Updated default port

  // Authentication configuration
  auth: {
    mode: __ENV.AUTH_MODE || 'dev', // 'dev' or 'kratos'
    devTenantID: __ENV.TEST_TENANT_ID || '00000000-0000-0000-0000-000000000001',
    devUserID: __ENV.TEST_USER_ID || 'default-test-user',
    kratosPublicURL: __ENV.KRATOS_PUBLIC_URL || 'http://localhost:4433',
  },

  // Test user credentials (for Kratos mode)
  testUser: {
    email: __ENV.TEST_USER_EMAIL || 'loadtest@example.com',
    password: __ENV.TEST_USER_PASSWORD || 'loadtest123',
  },

  // Test data configuration
  testData: {
    workflowNamePrefix: 'loadtest-workflow-',
    webhookNamePrefix: 'loadtest-webhook-',
  },

  // Virtual user (VU) configurations for different test types
  scenarios: {
    smoke: {
      vus: 1,
      duration: '1m',
      description: 'Minimal load to verify system functionality',
    },
    load: {
      stages: [
        { duration: '2m', target: 10 },  // Ramp up to 10 users
        { duration: '5m', target: 10 },  // Stay at 10 users
        { duration: '2m', target: 0 },   // Ramp down
      ],
      description: 'Normal expected load',
    },
    stress: {
      stages: [
        { duration: '2m', target: 20 },   // Ramp up to 20
        { duration: '5m', target: 20 },   // Stay at 20
        { duration: '2m', target: 50 },   // Spike to 50
        { duration: '5m', target: 50 },   // Stay at 50
        { duration: '2m', target: 100 },  // Spike to 100
        { duration: '5m', target: 100 },  // Stay at 100
        { duration: '5m', target: 0 },    // Ramp down
      ],
      description: 'Push system beyond normal capacity',
    },
    spike: {
      stages: [
        { duration: '1m', target: 10 },   // Normal load
        { duration: '1m', target: 100 },  // Sudden spike
        { duration: '3m', target: 100 },  // Stay at spike
        { duration: '1m', target: 10 },   // Return to normal
        { duration: '2m', target: 0 },    // Ramp down
      ],
      description: 'Sudden traffic spike',
    },
    soak: {
      stages: [
        { duration: '5m', target: 20 },    // Ramp up
        { duration: '60m', target: 20 },   // Stay for 1 hour
        { duration: '5m', target: 0 },     // Ramp down
      ],
      description: 'Extended duration test for memory leaks',
    },
  },

  // Performance thresholds
  thresholds: {
    // HTTP metrics
    http_req_duration: ['p(95)<500', 'p(99)<1000'], // 95% < 500ms, 99% < 1s
    http_req_failed: ['rate<0.01'],                  // Error rate < 1%

    // Specific endpoint thresholds
    'http_req_duration{endpoint:workflow_create}': ['p(95)<1000'],
    'http_req_duration{endpoint:workflow_execute}': ['p(95)<2000'],
    'http_req_duration{endpoint:webhook_trigger}': ['p(95)<200'],
    'http_req_duration{endpoint:auth_login}': ['p(95)<300'],

    // WebSocket metrics
    'ws_connecting{endpoint:websocket}': ['p(95)<500'],
    'ws_session_duration{endpoint:websocket}': ['p(95)<60000'],

    // Iteration metrics
    iteration_duration: ['p(95)<5000'],
    iterations: ['rate>0.5'],                        // At least 0.5 iterations/s
  },

  // Rate limiting
  rps: __ENV.MAX_RPS || 100, // Max requests per second

  // Test options
  options: {
    discardResponseBodies: false,
    setupTimeout: '60s',
    teardownTimeout: '60s',
    noConnectionReuse: false,
    userAgent: 'k6-load-test/1.0',

    // Tags for all requests
    tags: {
      platform: 'gorax',
      environment: __ENV.TEST_ENV || 'local',
    },
  },
};

// Helper function to get scenario config
export function getScenario(name) {
  const scenario = config.scenarios[name || __ENV.SCENARIO || 'load'];
  if (!scenario) {
    throw new Error(`Unknown scenario: ${name}. Available: ${Object.keys(config.scenarios).join(', ')}`);
  }
  return scenario;
}

// Helper function to generate test workflow
// Updated to match current Gorax API schema (nodes/edges structure)
export function generateTestWorkflow(id) {
  return {
    name: `${config.testData.workflowNamePrefix}${id}`,
    description: `Load test workflow ${id}`,
    definition: {
      nodes: [
        {
          id: `trigger-${id}`,
          type: 'trigger:webhook',  // Must use full type for validation
          position: { x: 0, y: 0 },
          data: {
            name: 'Webhook Trigger',
            config: {
              path: `/webhook/test-${id}`,
              auth_type: 'none',  // No auth for test webhooks
            },
          },
        },
        {
          id: `transform-${id}`,
          type: 'action:transform',  // Use full type
          position: { x: 200, y: 0 },
          data: {
            name: 'Transform Data',
            config: {
              expression: '{{ trigger.body.value * 2 }}',
            },
          },
        },
        {
          id: `http-${id}`,
          type: 'action:http',  // Use full type
          position: { x: 400, y: 0 },
          data: {
            name: 'Send HTTP Request',
            config: {
              method: 'POST',
              url: 'https://httpbin.org/post',
              body: '{{ transform.result }}',
            },
          },
        },
      ],
      edges: [
        {
          id: `e1-${id}`,
          source: `trigger-${id}`,
          target: `transform-${id}`,
        },
        {
          id: `e2-${id}`,
          source: `transform-${id}`,
          target: `http-${id}`,
        },
      ],
    },
  };
}

// Helper function to generate test webhook
export function generateTestWebhook(id) {
  return {
    name: `${config.testData.webhookNamePrefix}${id}`,
    path: `/webhook/test-${id}`,
    method: 'POST',
    enabled: true,
  };
}

export default config;
