# Error Handling Examples

This document provides practical examples of using Gorax error handling features in real-world scenarios.

## Example 1: HTTP API with Retry and Fallback

This example shows how to retry a primary API endpoint and fall back to a secondary endpoint if all retries fail.

```json
{
  "nodes": [
    {
      "id": "trigger-1",
      "type": "trigger:webhook",
      "data": {
        "label": "Webhook Trigger",
        "config": {
          "path": "/api/webhook"
        }
      }
    },
    {
      "id": "try-1",
      "type": "control:try",
      "data": {
        "label": "Try Primary API",
        "config": {
          "tryNodes": ["primary-api"],
          "catchNodes": ["fallback-api"],
          "finallyNodes": ["log-result"],
          "errorBinding": "apiError"
        }
      }
    },
    {
      "id": "primary-api",
      "type": "control:retry",
      "data": {
        "label": "Primary API (Retry)",
        "config": {
          "strategy": "exponential_jitter",
          "maxAttempts": 3,
          "initialDelayMs": 1000,
          "maxDelayMs": 10000,
          "multiplier": 2.0,
          "jitter": true,
          "retryableStatusCodes": [429, 500, 502, 503, 504]
        }
      }
    },
    {
      "id": "primary-http",
      "type": "action:http",
      "data": {
        "label": "Call Primary API",
        "config": {
          "method": "POST",
          "url": "https://primary-api.example.com/process",
          "headers": {
            "Content-Type": "application/json"
          },
          "body": "{{trigger.data}}",
          "timeout": 5000
        }
      }
    },
    {
      "id": "fallback-api",
      "type": "action:http",
      "data": {
        "label": "Fallback API",
        "config": {
          "method": "POST",
          "url": "https://backup-api.example.com/process",
          "headers": {
            "Content-Type": "application/json"
          },
          "body": "{{trigger.data}}",
          "timeout": 10000
        }
      }
    },
    {
      "id": "log-result",
      "type": "action:slack_send_message",
      "data": {
        "label": "Log Result",
        "config": {
          "channel": "#api-logs",
          "message": "API call completed. Used fallback: {{apiError != null}}"
        }
      }
    }
  ],
  "edges": [
    { "source": "trigger-1", "target": "try-1" },
    { "source": "try-1", "target": "primary-api", "sourceHandle": "try-1-try" },
    { "source": "primary-api", "target": "primary-http" },
    { "source": "try-1", "target": "fallback-api", "sourceHandle": "try-1-catch" },
    { "source": "try-1", "target": "log-result", "sourceHandle": "try-1-finally" }
  ]
}
```

**Execution Flow:**
1. Webhook triggers the workflow
2. Try block executes primary API with retry
3. If primary API fails after 3 retries, catch block executes fallback API
4. Finally block always logs the result (success or fallback used)

## Example 2: Database Operation with Circuit Breaker

Protect your database from overload using a circuit breaker.

```json
{
  "nodes": [
    {
      "id": "schedule-1",
      "type": "trigger:schedule",
      "data": {
        "label": "Every 5 Minutes",
        "config": {
          "cron": "*/5 * * * *"
        }
      }
    },
    {
      "id": "breaker-1",
      "type": "control:circuit_breaker",
      "data": {
        "label": "Database Circuit Breaker",
        "config": {
          "enabled": true,
          "name": "postgres-breaker",
          "maxFailures": 5,
          "timeoutMs": 60000,
          "failureThreshold": 0.6,
          "slidingWindowSize": 20
        }
      }
    },
    {
      "id": "db-query",
      "type": "action:http",
      "data": {
        "label": "Query Database",
        "config": {
          "method": "POST",
          "url": "{{env.DB_API_URL}}/query",
          "body": {
            "query": "SELECT * FROM users WHERE active = true"
          }
        }
      }
    },
    {
      "id": "process-data",
      "type": "action:transform",
      "data": {
        "label": "Process Results",
        "config": {
          "expression": "steps['db-query'].output.rows"
        }
      }
    }
  ],
  "edges": [
    { "source": "schedule-1", "target": "breaker-1" },
    { "source": "breaker-1", "target": "db-query" },
    { "source": "db-query", "target": "process-data" }
  ]
}
```

**Behavior:**
- After 5 consecutive failures, circuit opens
- While open, all requests fail immediately (no database calls)
- After 60 seconds, circuit enters half-open state
- If requests succeed, circuit closes and normal operation resumes

## Example 3: Multi-Service Orchestration with Error Handling

Complex workflow that orchestrates multiple services with comprehensive error handling.

```json
{
  "nodes": [
    {
      "id": "trigger-1",
      "type": "trigger:webhook",
      "data": {
        "label": "Order Webhook",
        "config": {
          "path": "/orders"
        }
      }
    },
    {
      "id": "try-main",
      "type": "control:try",
      "data": {
        "label": "Process Order",
        "config": {
          "tryNodes": ["validate-order", "process-payment", "create-shipment"],
          "catchNodes": ["handle-error", "notify-admin"],
          "finallyNodes": ["update-metrics"],
          "errorBinding": "orderError"
        }
      }
    },
    {
      "id": "validate-order",
      "type": "action:http",
      "data": {
        "label": "Validate Order",
        "config": {
          "method": "POST",
          "url": "{{env.VALIDATION_SERVICE}}/validate",
          "body": "{{trigger.data}}"
        }
      }
    },
    {
      "id": "process-payment",
      "type": "control:retry",
      "data": {
        "label": "Process Payment (Retry)",
        "config": {
          "strategy": "exponential",
          "maxAttempts": 3,
          "initialDelayMs": 2000,
          "maxDelayMs": 10000,
          "retryableErrors": ["timeout", "gateway.*timeout"],
          "nonRetryableErrors": ["insufficient.*funds", "card.*declined"]
        }
      }
    },
    {
      "id": "payment-api",
      "type": "action:http",
      "data": {
        "label": "Payment Gateway",
        "config": {
          "method": "POST",
          "url": "{{env.PAYMENT_GATEWAY}}/charge",
          "body": {
            "order_id": "{{trigger.data.order_id}}",
            "amount": "{{trigger.data.amount}}",
            "card": "{{trigger.data.card}}"
          }
        }
      }
    },
    {
      "id": "create-shipment",
      "type": "action:http",
      "data": {
        "label": "Create Shipment",
        "config": {
          "method": "POST",
          "url": "{{env.SHIPPING_SERVICE}}/shipments",
          "body": {
            "order_id": "{{trigger.data.order_id}}",
            "address": "{{trigger.data.shipping_address}}"
          }
        }
      }
    },
    {
      "id": "handle-error",
      "type": "action:transform",
      "data": {
        "label": "Handle Error",
        "config": {
          "expression": "{ error: orderError, order: trigger.data }"
        }
      }
    },
    {
      "id": "notify-admin",
      "type": "action:slack_send_message",
      "data": {
        "label": "Notify Admin",
        "config": {
          "channel": "#orders-failed",
          "message": "Order {{trigger.data.order_id}} failed: {{orderError.errorMessage}}"
        }
      }
    },
    {
      "id": "update-metrics",
      "type": "action:http",
      "data": {
        "label": "Update Metrics",
        "config": {
          "method": "POST",
          "url": "{{env.METRICS_SERVICE}}/track",
          "body": {
            "event": "order_processed",
            "success": "{{orderError == null}}",
            "duration_ms": "{{execution.duration}}"
          }
        }
      }
    }
  ],
  "edges": [
    { "source": "trigger-1", "target": "try-main" },
    { "source": "try-main", "target": "validate-order", "sourceHandle": "try-main-try" },
    { "source": "validate-order", "target": "process-payment" },
    { "source": "process-payment", "target": "payment-api" },
    { "source": "payment-api", "target": "create-shipment" },
    { "source": "try-main", "target": "handle-error", "sourceHandle": "try-main-catch" },
    { "source": "handle-error", "target": "notify-admin" },
    { "source": "try-main", "target": "update-metrics", "sourceHandle": "try-main-finally" }
  ]
}
```

**Features:**
- Validation before payment
- Retryable payment processing with smart error filtering
- Rollback notification on failure
- Metrics tracking in finally block (always runs)

## Example 4: Batch Processing with Error Tolerance

Process a batch of items with graceful error handling for individual items.

```json
{
  "nodes": [
    {
      "id": "trigger-1",
      "type": "trigger:schedule",
      "data": {
        "label": "Every Hour",
        "config": {
          "cron": "0 * * * *"
        }
      }
    },
    {
      "id": "fetch-items",
      "type": "action:http",
      "data": {
        "label": "Fetch Pending Items",
        "config": {
          "method": "GET",
          "url": "{{env.API_URL}}/items/pending"
        }
      }
    },
    {
      "id": "loop-1",
      "type": "control:loop",
      "data": {
        "label": "Process Each Item",
        "config": {
          "source": "{{steps['fetch-items'].output.items}}",
          "itemVariable": "item",
          "indexVariable": "index",
          "maxIterations": 1000,
          "onError": "continue"
        }
      }
    },
    {
      "id": "try-item",
      "type": "control:try",
      "data": {
        "label": "Try Process Item",
        "config": {
          "tryNodes": ["process-item"],
          "catchNodes": ["log-error"],
          "errorBinding": "itemError"
        }
      }
    },
    {
      "id": "process-item",
      "type": "control:retry",
      "data": {
        "label": "Process with Retry",
        "config": {
          "strategy": "exponential_jitter",
          "maxAttempts": 2,
          "initialDelayMs": 500,
          "maxDelayMs": 5000
        }
      }
    },
    {
      "id": "http-process",
      "type": "action:http",
      "data": {
        "label": "Process Item",
        "config": {
          "method": "POST",
          "url": "{{env.PROCESSOR_URL}}/process",
          "body": "{{item}}"
        }
      }
    },
    {
      "id": "log-error",
      "type": "action:transform",
      "data": {
        "label": "Log Failed Item",
        "config": {
          "expression": "logError({ item: item, error: itemError })"
        }
      }
    }
  ],
  "edges": [
    { "source": "trigger-1", "target": "fetch-items" },
    { "source": "fetch-items", "target": "loop-1" },
    { "source": "loop-1", "target": "try-item" },
    { "source": "try-item", "target": "process-item", "sourceHandle": "try-item-try" },
    { "source": "process-item", "target": "http-process" },
    { "source": "try-item", "target": "log-error", "sourceHandle": "try-item-catch" }
  ]
}
```

**Behavior:**
- Fetches batch of items
- Processes each item in a loop
- Retries failed items up to 2 times
- Logs errors but continues processing remaining items
- Perfect for data sync, notification sending, etc.

## Example 5: API Rate Limiting with Adaptive Backoff

Handle API rate limits intelligently with exponential backoff and jitter.

```json
{
  "nodes": [
    {
      "id": "trigger-1",
      "type": "trigger:webhook",
      "data": {
        "label": "API Request",
        "config": {
          "path": "/api/request"
        }
      }
    },
    {
      "id": "retry-1",
      "type": "control:retry",
      "data": {
        "label": "Rate Limited API",
        "config": {
          "strategy": "exponential_jitter",
          "maxAttempts": 5,
          "initialDelayMs": 1000,
          "maxDelayMs": 60000,
          "multiplier": 3.0,
          "jitter": true,
          "retryableStatusCodes": [429],
          "retryableErrors": ["rate.*limit", "too.*many.*requests"]
        }
      }
    },
    {
      "id": "api-call",
      "type": "action:http",
      "data": {
        "label": "Rate Limited API",
        "config": {
          "method": "GET",
          "url": "{{env.EXTERNAL_API}}/data",
          "headers": {
            "Authorization": "Bearer {{credentials.api_token}}"
          }
        }
      }
    }
  ],
  "edges": [
    { "source": "trigger-1", "target": "retry-1" },
    { "source": "retry-1", "target": "api-call" }
  ]
}
```

**Backoff Pattern with 3x multiplier:**
- Attempt 1: Immediate
- Attempt 2: ~1s ±25%
- Attempt 3: ~3s ±25%
- Attempt 4: ~9s ±25%
- Attempt 5: ~27s ±25%
- Attempt 6: 60s (capped at max)

**Jitter Benefits:**
- Prevents thundering herd
- Spreads out concurrent retries
- More API-friendly

## Best Practices Summary

### 1. Layered Error Handling

```
Circuit Breaker (outer)
  → Try/Catch (middle)
    → Retry (inner)
      → Action
```

### 2. Appropriate Retry Counts

- **Interactive requests**: 2-3 retries, short delays
- **Background jobs**: 5-7 retries, exponential backoff
- **Critical operations**: 10+ retries, long delays

### 3. Error Classification

Always specify retryable vs non-retryable errors:

```json
{
  "retryableErrors": ["timeout", "5\\d{2}", "connection"],
  "nonRetryableErrors": ["4\\d{2}", "auth", "invalid"]
}
```

### 4. Monitoring

Log error metadata in catch/finally blocks:

```json
{
  "id": "log-error",
  "type": "action:slack_send_message",
  "config": {
    "channel": "#errors",
    "message": "{{error.errorType}}: {{error.errorMessage}} (attempt {{error.retryAttempt}}/{{error.maxRetries}})"
  }
}
```

### 5. Resource Cleanup

Always use finally blocks for cleanup:

```json
{
  "finallyNodes": ["close-connection", "release-lock", "update-status"]
}
```

## Testing Error Handling

Use these patterns to test your error handling logic:

```bash
# Trigger with simulated errors
curl -X POST http://localhost:8080/api/webhook \
  -H "Content-Type: application/json" \
  -d '{"simulate_error": "timeout"}'

# Monitor retry attempts
curl http://localhost:8080/api/executions/{execution_id}/steps

# Check circuit breaker state
curl http://localhost:8080/api/metrics/circuit-breakers
```

## See Also

- [Error Handling Guide](./ERROR_HANDLING.md)
- [Node Types Reference](./NODE_TYPES.md)
- [Testing Guide](./TESTING.md)
