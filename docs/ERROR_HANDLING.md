# Error Handling in Gorax Workflows

This document describes the comprehensive error handling capabilities in Gorax workflows, including try/catch/finally blocks, automatic retries, and circuit breakers.

## Table of Contents

1. [Overview](#overview)
2. [Try/Catch/Finally](#trycatchfinally)
3. [Retry Strategies](#retry-strategies)
4. [Circuit Breakers](#circuit-breakers)
5. [Error Classification](#error-classification)
6. [Best Practices](#best-practices)
7. [Examples](#examples)

## Overview

Gorax provides production-ready error handling mechanisms to make your workflows resilient and reliable:

- **Try/Catch/Finally**: Structured error handling with cleanup blocks
- **Retry Logic**: Automatic retries with multiple strategies (fixed, exponential, exponential with jitter)
- **Circuit Breakers**: Prevent cascading failures by failing fast when a service is down
- **Error Classification**: Automatic classification of errors as transient or permanent
- **Error Context**: Rich error metadata for debugging and monitoring

## Try/Catch/Finally

### Node Type: `control:try`

The Try node allows you to handle errors gracefully using try/catch/finally semantics familiar from programming languages.

### Configuration

```json
{
  "type": "control:try",
  "config": {
    "tryNodes": ["action-node-1", "action-node-2"],
    "catchNodes": ["error-handler-node"],
    "finallyNodes": ["cleanup-node"],
    "errorBinding": "error",
    "retryConfig": {
      "strategy": "exponential",
      "maxAttempts": 3,
      "initialDelayMs": 1000,
      "maxDelayMs": 60000,
      "multiplier": 2.0,
      "jitter": true
    }
  }
}
```

### Configuration Options

| Field | Type | Description |
|-------|------|-------------|
| `tryNodes` | `string[]` | Array of node IDs to execute in the try block (required) |
| `catchNodes` | `string[]` | Array of node IDs to execute if an error occurs (optional) |
| `finallyNodes` | `string[]` | Array of node IDs to always execute, even if try/catch fails (optional) |
| `errorBinding` | `string` | Variable name to bind error details (default: "error") |
| `retryConfig` | `object` | Optional retry configuration for the try block |

### Error Object Structure

When an error occurs, it's bound to the context with the following structure:

```json
{
  "errorType": "HTTPError",
  "errorMessage": "connection timeout",
  "errorStack": "...",
  "classification": "transient",
  "nodeID": "http-node-1",
  "nodeType": "action:http",
  "retryAttempt": 2,
  "maxRetries": 3,
  "timestamp": "2024-01-15T10:30:00Z",
  "context": {},
  "caughtBy": "catch-node-1",
  "recoveryAction": "handled",
  "httpStatusCode": 504
}
```

### Behavior

1. **Try Block**: Executes all nodes in `tryNodes` sequentially
2. **Catch Block**: If any try node fails, executes `catchNodes` with error bound to context
3. **Finally Block**: Always executes `finallyNodes`, even if try/catch fails
4. **Error Propagation**: If no catch block or catch block fails, error propagates to parent

## Retry Strategies

### Node Type: `control:retry`

The Retry node wraps another node with automatic retry logic using configurable strategies.

### Configuration

```json
{
  "type": "control:retry",
  "config": {
    "strategy": "exponential_jitter",
    "maxAttempts": 5,
    "initialDelayMs": 1000,
    "maxDelayMs": 60000,
    "multiplier": 2.0,
    "jitter": true,
    "retryableErrors": ["timeout", "5xx", "connection.*"],
    "nonRetryableErrors": ["4xx", "invalid.*"],
    "retryableStatusCodes": [408, 429, 500, 502, 503, 504]
  }
}
```

### Strategy Types

#### 1. Fixed Delay (`"fixed"`)

Retries with a constant delay between attempts.

```json
{
  "strategy": "fixed",
  "maxAttempts": 3,
  "initialDelayMs": 1000
}
```

**Delay Pattern**: 1s, 1s, 1s

**Use Cases**:
- Simple scenarios where exponential backoff isn't needed
- Testing and development
- When the service has consistent recovery time

#### 2. Exponential Backoff (`"exponential"`)

Delays increase exponentially with each retry.

```json
{
  "strategy": "exponential",
  "maxAttempts": 5,
  "initialDelayMs": 1000,
  "maxDelayMs": 30000,
  "multiplier": 2.0
}
```

**Delay Pattern**: 1s, 2s, 4s, 8s, 16s, 30s (capped at max)

**Use Cases**:
- API rate limiting
- Service recovery scenarios
- Database connection issues
- Network instability

#### 3. Exponential Backoff with Jitter (`"exponential_jitter"`)

Adds random variation to exponential delays to prevent thundering herd.

```json
{
  "strategy": "exponential_jitter",
  "maxAttempts": 5,
  "initialDelayMs": 1000,
  "maxDelayMs": 60000,
  "multiplier": 2.0,
  "jitter": true
}
```

**Delay Pattern**: ~1s ±25%, ~2s ±25%, ~4s ±25%, ...

**Use Cases**:
- High-traffic systems
- Multiple concurrent retries
- Distributed systems
- Preventing synchronized retry storms

### Configuration Options

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `strategy` | `string` | `exponential_jitter` | Retry strategy: `fixed`, `exponential`, `exponential_jitter` |
| `maxAttempts` | `number` | Required | Maximum number of retry attempts (0 = no retries) |
| `initialDelayMs` | `number` | Required | Initial delay in milliseconds |
| `maxDelayMs` | `number` | 60000 | Maximum delay in milliseconds (for exponential) |
| `multiplier` | `number` | 2.0 | Backoff multiplier (for exponential) |
| `jitter` | `boolean` | true | Add random jitter to delays |
| `retryableErrors` | `string[]` | `[]` | Error patterns to retry (regex supported) |
| `nonRetryableErrors` | `string[]` | `[]` | Error patterns to never retry (regex supported) |
| `retryableStatusCodes` | `number[]` | `[408, 429, 500, 502, 503, 504]` | HTTP status codes to retry |

### Error Filtering

Errors can be filtered using patterns (regex supported):

```json
{
  "retryableErrors": [
    "timeout",
    "connection.*refused",
    "temporary.*failure",
    "5\\d{2}"
  ],
  "nonRetryableErrors": [
    "authentication.*failed",
    "unauthorized",
    "4\\d{2}"
  ]
}
```

## Circuit Breakers

### Node Type: `control:circuit_breaker`

Circuit breakers prevent cascading failures by failing fast when a service is consistently failing.

### States

1. **Closed**: Normal operation, all requests pass through
2. **Open**: Service is failing, all requests fail immediately
3. **Half-Open**: Testing if service has recovered, limited requests allowed

### Configuration

```json
{
  "type": "control:circuit_breaker",
  "config": {
    "enabled": true,
    "maxFailures": 5,
    "timeoutMs": 60000,
    "maxRequests": 3,
    "failureThreshold": 0.5,
    "slidingWindowSize": 10,
    "name": "external-api-breaker"
  }
}
```

### Configuration Options

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | `boolean` | true | Enable circuit breaker |
| `maxFailures` | `number` | 5 | Consecutive failures to open circuit |
| `timeoutMs` | `number` | 60000 | Time to wait before half-open (ms) |
| `maxRequests` | `number` | 3 | Max requests in half-open state |
| `failureThreshold` | `number` | 0.5 | Failure ratio to open (0.0-1.0) |
| `slidingWindowSize` | `number` | 10 | Window size for failure tracking |
| `name` | `string` | node-id | Circuit breaker name (for monitoring) |

### Behavior

1. **Closed → Open**: Opens after `maxFailures` consecutive failures or when failure ratio exceeds `failureThreshold`
2. **Open → Half-Open**: Transitions after `timeoutMs` milliseconds
3. **Half-Open → Closed**: Closes if success ratio exceeds `1 - failureThreshold`
4. **Half-Open → Open**: Re-opens on any failure

## Error Classification

Gorax automatically classifies errors to determine retry behavior:

### Transient Errors (Retryable)

- Network timeouts
- Connection refused/reset
- DNS resolution failures (temporary)
- HTTP 408, 429, 500, 502, 503, 504
- Rate limiting
- Temporary service unavailability

### Permanent Errors (Non-Retryable)

- Invalid input/syntax
- Authentication failures
- Authorization failures (403)
- Not found (404)
- Bad request (400)
- Malformed data
- Unsupported operations

### Unknown Errors

- Errors that don't match known patterns
- Custom application errors
- Treated as non-retryable by default

## Best Practices

### 1. Choose the Right Retry Strategy

```
Fixed Delay → Simple, predictable, testing
Exponential → Production API calls, rate limiting
Exponential + Jitter → High-traffic, distributed systems
```

### 2. Set Appropriate Timeouts

```json
{
  "maxAttempts": 3,
  "initialDelayMs": 1000,
  "maxDelayMs": 30000
}
```

- Start with 1-2 second initial delay
- Cap max delay at 30-60 seconds
- Total retry time = sum of all delays

### 3. Use Circuit Breakers for External Services

```json
{
  "maxFailures": 5,
  "timeoutMs": 60000,
  "failureThreshold": 0.5
}
```

- Monitor external dependencies
- Fail fast when service is down
- Prevent resource exhaustion

### 4. Combine Try/Catch with Retries

```json
{
  "tryNodes": ["http-with-retry"],
  "catchNodes": ["fallback-action"],
  "finallyNodes": ["cleanup"]
}
```

- Retry within try block
- Fallback to alternative in catch
- Clean up resources in finally

### 5. Log Error Context

Always log error metadata for debugging:

```json
{
  "errorType": "HTTPError",
  "classification": "transient",
  "retryAttempt": 2,
  "httpStatusCode": 503
}
```

### 6. Set Realistic Retry Limits

```
Quick operations: 3 retries
Critical operations: 5-7 retries
Background jobs: 10+ retries
```

### 7. Use Error Patterns Wisely

```json
{
  "retryableErrors": ["timeout", "5\\d{2}"],
  "nonRetryableErrors": ["auth.*", "4\\d{2}"]
}
```

- Be specific with patterns
- Use regex for flexibility
- Test patterns thoroughly

## Examples

### Example 1: HTTP Request with Retry

```json
{
  "nodes": [
    {
      "id": "http-1",
      "type": "action:http",
      "config": {
        "method": "GET",
        "url": "https://api.example.com/data",
        "timeout": 5000
      }
    },
    {
      "id": "retry-1",
      "type": "control:retry",
      "config": {
        "strategy": "exponential_jitter",
        "maxAttempts": 3,
        "initialDelayMs": 1000,
        "maxDelayMs": 30000,
        "retryableStatusCodes": [429, 500, 502, 503, 504]
      }
    }
  ]
}
```

### Example 2: Try/Catch with Fallback

```json
{
  "nodes": [
    {
      "id": "try-1",
      "type": "control:try",
      "config": {
        "tryNodes": ["primary-api"],
        "catchNodes": ["fallback-api"],
        "finallyNodes": ["log-result"],
        "errorBinding": "error"
      }
    },
    {
      "id": "primary-api",
      "type": "action:http",
      "config": {
        "url": "https://primary-api.example.com"
      }
    },
    {
      "id": "fallback-api",
      "type": "action:http",
      "config": {
        "url": "https://fallback-api.example.com"
      }
    },
    {
      "id": "log-result",
      "type": "action:transform",
      "config": {
        "expression": "log(result)"
      }
    }
  ]
}
```

### Example 3: Circuit Breaker for External Service

```json
{
  "nodes": [
    {
      "id": "breaker-1",
      "type": "control:circuit_breaker",
      "config": {
        "name": "payment-gateway",
        "maxFailures": 5,
        "timeoutMs": 60000,
        "failureThreshold": 0.6
      }
    },
    {
      "id": "payment-api",
      "type": "action:http",
      "config": {
        "url": "https://payment.example.com/charge"
      }
    }
  ]
}
```

### Example 4: Complex Error Handling Pipeline

```json
{
  "nodes": [
    {
      "id": "try-1",
      "type": "control:try",
      "config": {
        "tryNodes": ["retry-http"],
        "catchNodes": ["handle-error"],
        "finallyNodes": ["cleanup"],
        "errorBinding": "error",
        "retryConfig": {
          "strategy": "exponential_jitter",
          "maxAttempts": 3,
          "initialDelayMs": 1000
        }
      }
    },
    {
      "id": "retry-http",
      "type": "control:retry",
      "config": {
        "strategy": "exponential_jitter",
        "maxAttempts": 5,
        "initialDelayMs": 1000,
        "maxDelayMs": 60000,
        "retryableErrors": ["timeout", "connection", "5\\d{2}"],
        "nonRetryableErrors": ["auth", "4\\d{2}"]
      }
    },
    {
      "id": "handle-error",
      "type": "action:slack_send_message",
      "config": {
        "channel": "#alerts",
        "message": "API call failed: {{error.errorMessage}}"
      }
    },
    {
      "id": "cleanup",
      "type": "action:transform",
      "config": {
        "expression": "clearCache()"
      }
    }
  ]
}
```

### Example 5: Database Operations with Retry

```json
{
  "nodes": [
    {
      "id": "db-insert",
      "type": "action:http",
      "config": {
        "method": "POST",
        "url": "{{env.DB_API_URL}}/records",
        "body": "{{trigger.data}}"
      }
    },
    {
      "id": "retry-db",
      "type": "control:retry",
      "config": {
        "strategy": "exponential",
        "maxAttempts": 7,
        "initialDelayMs": 500,
        "maxDelayMs": 30000,
        "multiplier": 2.0,
        "retryableErrors": [
          "connection pool exhausted",
          "deadlock detected",
          "timeout"
        ],
        "nonRetryableErrors": [
          "duplicate key",
          "foreign key constraint"
        ]
      }
    }
  ]
}
```

## Monitoring and Observability

Error handling nodes emit detailed metrics and logs:

### Metrics

- `retry_attempts_total`: Total number of retry attempts
- `retry_success_total`: Successful retries
- `retry_failure_total`: Failed retries (exhausted)
- `circuit_breaker_state`: Current circuit breaker state
- `circuit_breaker_failures`: Failure count per breaker
- `error_handling_recovery_actions`: Recovery actions taken (handled, retry, fallback, propagate)

### Logs

All error handling operations are logged with rich context:

```json
{
  "level": "error",
  "msg": "retry attempt failed",
  "node_id": "retry-1",
  "attempt": 2,
  "max_attempts": 5,
  "delay_ms": 2000,
  "error_type": "HTTPError",
  "classification": "transient",
  "http_status": 503
}
```

## Database Schema

Error handling history is stored in the `error_handling_history` table:

```sql
CREATE TABLE error_handling_history (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    execution_id UUID NOT NULL,
    node_id VARCHAR(255) NOT NULL,
    error_type VARCHAR(100) NOT NULL,
    error_message TEXT NOT NULL,
    error_classification VARCHAR(50) NOT NULL,
    retry_attempt INTEGER,
    max_retries INTEGER,
    retry_strategy VARCHAR(50),
    caught_by_node_id VARCHAR(255),
    recovery_action VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL
);
```

Query error patterns:

```sql
-- Most common errors
SELECT error_type, COUNT(*) as count
FROM error_handling_history
WHERE tenant_id = $1
GROUP BY error_type
ORDER BY count DESC;

-- Retry effectiveness
SELECT
    retry_strategy,
    AVG(retry_attempt) as avg_attempts,
    COUNT(CASE WHEN recovery_action = 'handled' THEN 1 END) as successful_recoveries
FROM error_handling_history
WHERE retry_strategy IS NOT NULL
GROUP BY retry_strategy;
```

## Troubleshooting

### Issue: Too Many Retries

**Symptom**: Executions taking too long, timeout errors

**Solution**:
- Reduce `maxAttempts` to 3-5
- Lower `maxDelayMs` to 30000
- Use `exponential_jitter` to prevent thundering herd
- Add `nonRetryableErrors` patterns for permanent failures

### Issue: Circuit Breaker Opens Too Quickly

**Symptom**: Circuit breaker opens on minor issues

**Solution**:
- Increase `maxFailures` to 8-10
- Raise `failureThreshold` to 0.6-0.7
- Increase `slidingWindowSize` to 20-30

### Issue: Errors Not Being Caught

**Symptom**: Errors propagate despite catch block

**Solution**:
- Check error patterns in `CatchConfig.errorPatterns`
- Verify error types in `CatchConfig.errorTypes`
- Use empty filters to catch all errors

### Issue: Finally Block Not Executing

**Symptom**: Cleanup code not running

**Solution**:
- Check for panics in try/catch blocks
- Verify `finallyNodes` configuration
- Ensure finally block doesn't depend on try output

## See Also

- [Workflow Execution Guide](./WORKFLOW_EXECUTION.md)
- [Node Types Reference](./NODE_TYPES.md)
- [Monitoring and Observability](./MONITORING.md)
- [Best Practices](./BEST_PRACTICES.md)
