// k6 Load Test: WebSocket Connection Scaling
// Tests WebSocket connection handling and real-time updates under load

import ws from 'k6/ws';
import http from 'k6/http';
import { check, group, sleep } from 'k6';
import { Counter, Rate, Trend } from 'k6/metrics';
import { config, getScenario } from './config.js';
import { setupAuth, getAuthHeaders } from './lib/auth.js';

// Custom metrics
const wsConnectionDuration = new Trend('ws_connection_duration', true);
const wsMessageLatency = new Trend('ws_message_latency', true);
const wsMessagesReceived = new Counter('ws_messages_received');
const wsMessagesSent = new Counter('ws_messages_sent');
const wsConnectionErrors = new Counter('ws_connection_errors');
const wsConnectionSuccessRate = new Rate('ws_connection_success_rate');
const wsActiveConnections = new Counter('ws_active_connections');

// Test configuration
const scenario = getScenario();
export const options = {
  stages: scenario.stages || [{ duration: scenario.duration, target: scenario.vus }],
  thresholds: {
    ...config.thresholds,
    'ws_connection_duration': ['p(95)<1000', 'p(99)<2000'],
    'ws_message_latency': ['p(95)<100', 'p(99)<200'],
    'ws_connection_success_rate': ['rate>0.95'],
  },
  tags: config.options.tags,
};

// Setup: Authenticate
export function setup() {
  return setupAuth(config);
}

// Main test function
export default function (data) {
  // Build WebSocket URL with auth parameters
  let wsUrl = `${config.wsUrl}/ws`;

  // Add auth parameters based on mode
  if (data.mode === 'dev') {
    wsUrl += `?tenant_id=${data.tenantID}&user_id=${data.userID}`;
  } else if (data.mode === 'kratos') {
    wsUrl += `?token=${data.sessionToken}`;
  }

  let messagesReceived = 0;
  let messageLatencies = [];

  // Test 1: WebSocket Connection
  group('WebSocket Connection', () => {
    const connectStart = Date.now();

    const res = ws.connect(wsUrl, {
      tags: { endpoint: 'websocket' },
    }, function (socket) {
      const connectDuration = Date.now() - connectStart;
      wsConnectionDuration.add(connectDuration);

      const connectSuccess = check(connectDuration, {
        'connected successfully': () => true,
        'connection time < 1s': () => connectDuration < 1000,
      });

      wsConnectionSuccessRate.add(connectSuccess);
      if (!connectSuccess) {
        wsConnectionErrors.add(1);
        return;
      }

      wsActiveConnections.add(1);

      // Handle incoming messages
      socket.on('open', () => {
        console.log(`VU ${__VU}: WebSocket connected`);

        // Send initial subscription message
        const subscribeMsg = JSON.stringify({
          type: 'subscribe',
          channels: ['executions', 'workflows'],
          vu: __VU,
          iteration: __ITER,
        });

        socket.send(subscribeMsg);
        wsMessagesSent.add(1);
      });

      socket.on('message', (msg) => {
        const receiveTime = Date.now();
        messagesReceived++;
        wsMessagesReceived.add(1);

        try {
          const data = JSON.parse(msg);

          // Calculate latency if message has timestamp
          if (data.timestamp) {
            const sentTime = new Date(data.timestamp).getTime();
            const latency = receiveTime - sentTime;
            wsMessageLatency.add(latency);
            messageLatencies.push(latency);
          }

          check(data, {
            'message has type': () => data.type !== undefined,
            'message is valid JSON': () => true,
          });

          // Echo messages for some scenarios
          if (data.type === 'execution.started') {
            socket.send(JSON.stringify({
              type: 'ack',
              messageId: data.id,
              timestamp: new Date().toISOString(),
            }));
            wsMessagesSent.add(1);
          }
        } catch (e) {
          console.error(`Failed to parse message: ${e.message}`);
          wsConnectionErrors.add(1);
        }
      });

      socket.on('error', (e) => {
        console.error(`VU ${__VU}: WebSocket error: ${e.error()}`);
        wsConnectionErrors.add(1);
      });

      socket.on('close', () => {
        console.log(`VU ${__VU}: WebSocket closed. Received ${messagesReceived} messages`);
        wsActiveConnections.add(-1);
      });

      // Test 2: Send Messages at Intervals
      const messagingDuration = 30; // seconds
      const messageInterval = 2; // seconds
      const iterations = messagingDuration / messageInterval;

      for (let i = 0; i < iterations; i++) {
        socket.send(JSON.stringify({
          type: 'ping',
          timestamp: new Date().toISOString(),
          vu: __VU,
          iteration: i,
        }));
        wsMessagesSent.add(1);

        // Wait for message interval
        socket.setTimeout(() => {}, messageInterval * 1000);
      }

      // Test 3: Burst Messages
      group('Message Burst', () => {
        const burstSize = 10;
        const burstStart = Date.now();

        for (let i = 0; i < burstSize; i++) {
          socket.send(JSON.stringify({
            type: 'test.burst',
            index: i,
            timestamp: new Date().toISOString(),
          }));
          wsMessagesSent.add(1);
        }

        const burstDuration = Date.now() - burstStart;
        check(burstDuration, {
          'burst sent successfully': () => true,
          'burst time < 500ms': () => burstDuration < 500,
        });
      });

      // Keep connection alive for test duration
      socket.setTimeout(() => {
        socket.close();
      }, 35000);
    });

    if (res !== undefined) {
      wsConnectionErrors.add(1);
      wsConnectionSuccessRate.add(false);
      console.error(`Connection failed: ${res}`);
    }
  });

  sleep(1);

  // Test 4: Connection Reliability Check
  group('Connection Reliability', () => {
    check(messagesReceived, {
      'received messages': () => messagesReceived > 0,
      'message latency acceptable': () => {
        if (messageLatencies.length === 0) return true;
        const avgLatency = messageLatencies.reduce((a, b) => a + b, 0) / messageLatencies.length;
        return avgLatency < 200;
      },
    });
  });

  sleep(2);
}

// Teardown
export function teardown(data) {
  console.log('WebSocket load test completed.');
}

// Handle test summary
export function handleSummary(data) {
  return {
    'stdout': textSummary(data, { indent: ' ', enableColors: true }),
    'websocket_connections_results.json': JSON.stringify(data, null, 2),
  };
}

function textSummary(data, options) {
  const indent = options?.indent || '';

  let summary = '\n';
  summary += `${indent}WebSocket Connection Load Test Summary\n`;
  summary += `${indent}======================================\n\n`;

  // Test metadata
  summary += `${indent}Test Duration: ${formatDuration(data.state.testRunDurationMs)}\n`;
  summary += `${indent}VUs: ${data.metrics.vus?.values?.max || 'N/A'}\n`;
  summary += `${indent}Iterations: ${data.metrics.iterations?.values?.count || 0}\n\n`;

  // WebSocket connection metrics
  summary += `${indent}Connection Performance:\n`;
  summary += `${indent}  Connection p95: ${data.metrics.ws_connection_duration?.values['p(95)']?.toFixed(2) || 0}ms\n`;
  summary += `${indent}  Connection p99: ${data.metrics.ws_connection_duration?.values['p(99)']?.toFixed(2) || 0}ms\n`;
  summary += `${indent}  Success Rate: ${(data.metrics.ws_connection_success_rate?.values?.rate * 100 || 0).toFixed(2)}%\n`;
  summary += `${indent}  Connection Errors: ${data.metrics.ws_connection_errors?.values?.count || 0}\n`;
  summary += `${indent}  Peak Concurrent: ${data.metrics.ws_active_connections?.values?.max || 0}\n\n`;

  // WebSocket message metrics
  summary += `${indent}Message Performance:\n`;
  summary += `${indent}  Messages Sent: ${data.metrics.ws_messages_sent?.values?.count || 0}\n`;
  summary += `${indent}  Messages Received: ${data.metrics.ws_messages_received?.values?.count || 0}\n`;
  summary += `${indent}  Message Latency p95: ${data.metrics.ws_message_latency?.values['p(95)']?.toFixed(2) || 0}ms\n`;
  summary += `${indent}  Message Latency p99: ${data.metrics.ws_message_latency?.values['p(99)']?.toFixed(2) || 0}ms\n`;
  summary += `${indent}  Avg Message Latency: ${data.metrics.ws_message_latency?.values?.avg?.toFixed(2) || 0}ms\n\n`;

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
