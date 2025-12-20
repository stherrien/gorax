# Gorax API Reference

## Overview

The Gorax API is a RESTful API that provides programmatic access to the Gorax workflow automation platform. This API enables you to create, manage, and execute workflows, configure webhooks, manage schedules, and monitor execution history.

**Base URL:** `http://localhost:8080/api/v1` (development)

**API Version:** 1.0

**Interactive Documentation:** `/api/docs/` (Swagger UI)

## Authentication

### Development Mode

In development mode, authentication is simplified using custom headers:

```bash
curl -X GET http://localhost:8080/api/v1/workflows \
  -H "X-User-ID: user_123" \
  -H "X-Tenant-ID: tenant_abc"
```

**Required Headers:**
- `X-User-ID`: User identifier
- `X-Tenant-ID`: Tenant identifier for multi-tenant isolation

### Production Mode

In production, Gorax uses [Ory Kratos](https://www.ory.sh/docs/kratos/) for authentication. After authenticating with Kratos, include the session cookie in your requests:

```bash
curl -X GET http://localhost:8080/api/v1/workflows \
  -H "X-Tenant-ID: tenant_abc" \
  --cookie "ory_kratos_session=<session_token>"
```

**Required Headers:**
- `X-Tenant-ID`: Tenant identifier
- Cookie: `ory_kratos_session` (set automatically by Kratos)

## Rate Limiting

API requests are rate-limited based on tenant quotas. When approaching rate limits, responses include these headers:

```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 995
X-RateLimit-Reset: 1640000000
```

If you exceed rate limits, you'll receive a `429 Too Many Requests` response:

```json
{
  "error": "rate limit exceeded",
  "retry_after": 60
}
```

## Error Handling

All errors follow a consistent JSON format:

```json
{
  "error": "descriptive error message"
}
```

### HTTP Status Codes

| Status Code | Description |
|-------------|-------------|
| 200 | OK - Request succeeded |
| 201 | Created - Resource created successfully |
| 202 | Accepted - Request accepted for processing |
| 204 | No Content - Request succeeded with no response body |
| 400 | Bad Request - Invalid request parameters or body |
| 401 | Unauthorized - Authentication required |
| 403 | Forbidden - Insufficient permissions |
| 404 | Not Found - Resource not found |
| 409 | Conflict - Resource conflict (e.g., duplicate) |
| 429 | Too Many Requests - Rate limit exceeded |
| 500 | Internal Server Error - Server error occurred |
| 503 | Service Unavailable - Service temporarily unavailable |

## Pagination

List endpoints support pagination using query parameters:

```bash
GET /api/v1/workflows?limit=20&offset=0
```

**Parameters:**
- `limit` (integer): Maximum number of results (default: 20, max: 100)
- `offset` (integer): Number of results to skip (default: 0)

**Response:**
```json
{
  "data": [...],
  "limit": 20,
  "offset": 0,
  "total": 150
}
```

---

## API Endpoints

### Health & Monitoring

#### Health Check
```http
GET /health
```

Returns basic health status of the API.

**Response 200:**
```json
{
  "status": "ok",
  "timestamp": "2024-01-20T10:00:00Z"
}
```

**Example:**
```bash
curl http://localhost:8080/health
```

---

#### Readiness Check
```http
GET /ready
```

Returns readiness status including dependency health checks (database, Redis).

**Response 200 (All healthy):**
```json
{
  "status": "ok",
  "timestamp": "2024-01-20T10:00:00Z",
  "checks": {
    "database": "healthy",
    "redis": "healthy"
  }
}
```

**Response 503 (Degraded):**
```json
{
  "status": "degraded",
  "timestamp": "2024-01-20T10:00:00Z",
  "checks": {
    "database": "healthy",
    "redis": "unhealthy: connection timeout"
  }
}
```

---

### Workflows

#### List Workflows
```http
GET /api/v1/workflows
```

Returns a paginated list of workflows for the authenticated tenant.

**Query Parameters:**
- `limit` (integer, optional): Maximum results (default: 20)
- `offset` (integer, optional): Pagination offset (default: 0)

**Response 200:**
```json
{
  "data": [
    {
      "id": "wf_abc123",
      "tenant_id": "tenant_xyz",
      "name": "Customer Onboarding",
      "description": "Automated customer onboarding workflow",
      "enabled": true,
      "version": 3,
      "created_at": "2024-01-15T10:00:00Z",
      "updated_at": "2024-01-20T15:30:00Z",
      "definition": { ... }
    }
  ],
  "limit": 20,
  "offset": 0
}
```

**Example:**
```bash
curl -X GET http://localhost:8080/api/v1/workflows \
  -H "X-User-ID: user_123" \
  -H "X-Tenant-ID: tenant_xyz"
```

---

#### Create Workflow
```http
POST /api/v1/workflows
```

Creates a new workflow with the provided configuration.

**Request Body:**
```json
{
  "name": "Order Processing",
  "description": "Automated order processing workflow",
  "enabled": true,
  "definition": {
    "trigger": {
      "type": "webhook",
      "config": {}
    },
    "actions": [
      {
        "id": "validate_order",
        "type": "http",
        "config": {
          "url": "https://api.example.com/validate",
          "method": "POST"
        }
      }
    ]
  }
}
```

**Response 201:**
```json
{
  "data": {
    "id": "wf_new123",
    "tenant_id": "tenant_xyz",
    "name": "Order Processing",
    "enabled": true,
    "version": 1,
    "created_at": "2024-01-20T16:00:00Z"
  }
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/api/v1/workflows \
  -H "X-User-ID: user_123" \
  -H "X-Tenant-ID: tenant_xyz" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Order Processing",
    "description": "Automated order processing",
    "enabled": true,
    "definition": {...}
  }'
```

---

#### Get Workflow
```http
GET /api/v1/workflows/{workflowID}
```

Retrieves a specific workflow by ID.

**Path Parameters:**
- `workflowID` (string, required): Workflow identifier

**Response 200:**
```json
{
  "data": {
    "id": "wf_abc123",
    "tenant_id": "tenant_xyz",
    "name": "Customer Onboarding",
    "description": "Automated customer onboarding",
    "enabled": true,
    "version": 3,
    "created_at": "2024-01-15T10:00:00Z",
    "updated_at": "2024-01-20T15:30:00Z",
    "definition": {...}
  }
}
```

**Example:**
```bash
curl -X GET http://localhost:8080/api/v1/workflows/wf_abc123 \
  -H "X-User-ID: user_123" \
  -H "X-Tenant-ID: tenant_xyz"
```

---

#### Update Workflow
```http
PUT /api/v1/workflows/{workflowID}
```

Updates an existing workflow. Creates a new version.

**Path Parameters:**
- `workflowID` (string, required): Workflow identifier

**Request Body:**
```json
{
  "name": "Customer Onboarding v2",
  "description": "Updated workflow",
  "enabled": true,
  "definition": {...}
}
```

**Response 200:**
```json
{
  "data": {
    "id": "wf_abc123",
    "version": 4,
    "updated_at": "2024-01-20T16:30:00Z"
  }
}
```

---

#### Delete Workflow
```http
DELETE /api/v1/workflows/{workflowID}
```

Deletes a workflow. This is a soft delete; execution history is preserved.

**Path Parameters:**
- `workflowID` (string, required): Workflow identifier

**Response 204:** No content

**Example:**
```bash
curl -X DELETE http://localhost:8080/api/v1/workflows/wf_abc123 \
  -H "X-User-ID: user_123" \
  -H "X-Tenant-ID: tenant_xyz"
```

---

#### Execute Workflow
```http
POST /api/v1/workflows/{workflowID}/execute
```

Triggers a manual execution of a workflow.

**Path Parameters:**
- `workflowID` (string, required): Workflow identifier

**Request Body (optional):**
```json
{
  "customer_id": "cust_123",
  "order_total": 99.99
}
```

**Response 202:**
```json
{
  "data": {
    "execution_id": "exec_xyz789",
    "workflow_id": "wf_abc123",
    "status": "running",
    "started_at": "2024-01-20T17:00:00Z"
  }
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/api/v1/workflows/wf_abc123/execute \
  -H "X-User-ID: user_123" \
  -H "X-Tenant-ID: tenant_xyz" \
  -H "Content-Type: application/json" \
  -d '{"customer_id": "cust_123"}'
```

---

#### Dry-Run Workflow
```http
POST /api/v1/workflows/{workflowID}/dry-run
```

Validates a workflow without executing it. Useful for testing.

**Path Parameters:**
- `workflowID` (string, required): Workflow identifier

**Request Body:**
```json
{
  "test_data": {
    "customer_id": "test_123",
    "order_total": 50.00
  }
}
```

**Response 200:**
```json
{
  "data": {
    "valid": true,
    "steps": [
      {
        "action_id": "validate_order",
        "result": "success",
        "output": {...}
      }
    ],
    "execution_time_ms": 125
  }
}
```

---

#### List Workflow Versions
```http
GET /api/v1/workflows/{workflowID}/versions
```

Retrieves all versions of a workflow.

**Path Parameters:**
- `workflowID` (string, required): Workflow identifier

**Response 200:**
```json
{
  "data": [
    {
      "version": 3,
      "created_at": "2024-01-20T15:30:00Z",
      "created_by": "user_123",
      "changes": "Updated validation step"
    },
    {
      "version": 2,
      "created_at": "2024-01-18T10:00:00Z",
      "created_by": "user_456"
    }
  ]
}
```

---

#### Get Workflow Version
```http
GET /api/v1/workflows/{workflowID}/versions/{version}
```

Retrieves a specific version of a workflow.

**Path Parameters:**
- `workflowID` (string, required): Workflow identifier
- `version` (integer, required): Version number

**Response 200:**
```json
{
  "data": {
    "version": 2,
    "definition": {...},
    "created_at": "2024-01-18T10:00:00Z"
  }
}
```

---

#### Restore Workflow Version
```http
POST /api/v1/workflows/{workflowID}/versions/{version}/restore
```

Restores a workflow to a previous version. Creates a new version with the restored configuration.

**Path Parameters:**
- `workflowID` (string, required): Workflow identifier
- `version` (integer, required): Version number to restore

**Response 200:**
```json
{
  "data": {
    "id": "wf_abc123",
    "version": 4,
    "restored_from": 2,
    "updated_at": "2024-01-20T18:00:00Z"
  }
}
```

---

### Webhooks

#### List Webhooks
```http
GET /api/v1/webhooks
```

Returns a paginated list of webhooks for the authenticated tenant.

**Query Parameters:**
- `limit` (integer, optional): Maximum results (default: 20)
- `offset` (integer, optional): Pagination offset (default: 0)

**Response 200:**
```json
{
  "data": [
    {
      "id": "wh_abc123",
      "tenant_id": "tenant_xyz",
      "workflow_id": "wf_abc123",
      "name": "Order Webhook",
      "path": "/orders",
      "auth_type": "signature",
      "enabled": true,
      "priority": 1,
      "created_at": "2024-01-15T10:00:00Z"
    }
  ],
  "total": 5,
  "limit": 20,
  "offset": 0
}
```

---

#### Create Webhook
```http
POST /api/v1/webhooks
```

Creates a new webhook endpoint for a workflow.

**Request Body:**
```json
{
  "name": "Payment Webhook",
  "workflowId": "wf_abc123",
  "path": "/payments",
  "authType": "signature",
  "description": "Receives payment notifications",
  "priority": 1
}
```

**Auth Types:**
- `none`: No authentication
- `signature`: HMAC signature verification
- `basic`: HTTP Basic authentication
- `api_key`: API key in header

**Priority Levels:**
- `0`: Low priority
- `1`: Normal priority (default)
- `2`: High priority
- `3`: Critical priority

**Response 201:**
```json
{
  "data": {
    "id": "wh_new123",
    "url": "https://your-domain.com/webhooks/wf_abc123/wh_new123",
    "secret": "whsec_...",
    "enabled": true
  }
}
```

---

#### Get Webhook
```http
GET /api/v1/webhooks/{id}
```

Retrieves a specific webhook by ID.

**Path Parameters:**
- `id` (string, required): Webhook identifier

**Response 200:**
```json
{
  "data": {
    "id": "wh_abc123",
    "tenant_id": "tenant_xyz",
    "workflow_id": "wf_abc123",
    "name": "Order Webhook",
    "url": "https://your-domain.com/webhooks/wf_abc123/wh_abc123",
    "auth_type": "signature",
    "enabled": true,
    "created_at": "2024-01-15T10:00:00Z"
  }
}
```

---

#### Update Webhook
```http
PUT /api/v1/webhooks/{id}
```

Updates an existing webhook.

**Path Parameters:**
- `id` (string, required): Webhook identifier

**Request Body:**
```json
{
  "name": "Updated Order Webhook",
  "authType": "signature",
  "description": "Updated description",
  "priority": 2,
  "enabled": true
}
```

**Response 200:**
```json
{
  "data": {
    "id": "wh_abc123",
    "updated_at": "2024-01-20T16:00:00Z"
  }
}
```

---

#### Delete Webhook
```http
DELETE /api/v1/webhooks/{id}
```

Deletes a webhook endpoint.

**Path Parameters:**
- `id` (string, required): Webhook identifier

**Response 204:** No content

---

#### Regenerate Webhook Secret
```http
POST /api/v1/webhooks/{id}/regenerate-secret
```

Regenerates the secret key for webhook signature verification.

**Path Parameters:**
- `id` (string, required): Webhook identifier

**Response 200:**
```json
{
  "data": {
    "id": "wh_abc123",
    "secret": "whsec_new...",
    "regenerated_at": "2024-01-20T16:30:00Z"
  }
}
```

---

#### Test Webhook
```http
POST /api/v1/webhooks/{id}/test
```

Tests a webhook with a sample payload.

**Path Parameters:**
- `id` (string, required): Webhook identifier

**Request Body:**
```json
{
  "method": "POST",
  "headers": {
    "Content-Type": "application/json"
  },
  "body": {
    "test": true,
    "order_id": "test_123"
  }
}
```

**Response 200:**
```json
{
  "success": true,
  "execution_id": "exec_test123",
  "response_time_ms": 250,
  "workflow_triggered": true
}
```

---

#### Get Webhook Event History
```http
GET /api/v1/webhooks/{id}/events
```

Retrieves the event history for a webhook.

**Path Parameters:**
- `id` (string, required): Webhook identifier

**Query Parameters:**
- `limit` (integer, optional): Maximum results (default: 20)
- `offset` (integer, optional): Pagination offset (default: 0)

**Response 200:**
```json
{
  "data": [
    {
      "id": "evt_abc123",
      "webhook_id": "wh_abc123",
      "received_at": "2024-01-20T17:00:00Z",
      "method": "POST",
      "status": "processed",
      "execution_id": "exec_xyz789",
      "response_time_ms": 150
    }
  ],
  "total": 45,
  "limit": 20,
  "offset": 0
}
```

---

#### Replay Webhook Event
```http
POST /api/v1/events/{eventID}/replay
```

Replays a previously received webhook event.

**Path Parameters:**
- `eventID` (string, required): Event identifier

**Response 202:**
```json
{
  "execution_id": "exec_replay123",
  "original_event_id": "evt_abc123",
  "status": "queued"
}
```

---

### Executions

#### List Executions
```http
GET /api/v1/executions
```

Returns a paginated list of workflow executions with advanced filtering.

**Query Parameters:**
- `limit` (integer, optional): Maximum results (default: 20)
- `offset` (integer, optional): Pagination offset (default: 0)
- `workflow_id` (string, optional): Filter by workflow ID
- `status` (string, optional): Filter by status (running, completed, failed)
- `trigger_type` (string, optional): Filter by trigger type (manual, webhook, schedule)
- `from` (string, optional): Start date (ISO 8601)
- `to` (string, optional): End date (ISO 8601)

**Response 200:**
```json
{
  "data": [
    {
      "id": "exec_xyz789",
      "workflow_id": "wf_abc123",
      "workflow_name": "Order Processing",
      "status": "completed",
      "trigger_type": "webhook",
      "started_at": "2024-01-20T17:00:00Z",
      "completed_at": "2024-01-20T17:00:15Z",
      "duration_ms": 15000,
      "steps_completed": 5,
      "steps_total": 5
    }
  ],
  "limit": 20,
  "offset": 0
}
```

**Example:**
```bash
curl -X GET 'http://localhost:8080/api/v1/executions?status=failed&from=2024-01-01' \
  -H "X-User-ID: user_123" \
  -H "X-Tenant-ID: tenant_xyz"
```

---

#### Get Execution
```http
GET /api/v1/executions/{executionID}
```

Retrieves details of a specific execution.

**Path Parameters:**
- `executionID` (string, required): Execution identifier

**Response 200:**
```json
{
  "data": {
    "id": "exec_xyz789",
    "workflow_id": "wf_abc123",
    "status": "completed",
    "trigger_type": "webhook",
    "trigger_data": {...},
    "started_at": "2024-01-20T17:00:00Z",
    "completed_at": "2024-01-20T17:00:15Z",
    "duration_ms": 15000,
    "error": null
  }
}
```

---

#### Get Execution Steps
```http
GET /api/v1/executions/{executionID}/steps
```

Retrieves detailed step-by-step execution information.

**Path Parameters:**
- `executionID` (string, required): Execution identifier

**Response 200:**
```json
{
  "data": {
    "execution_id": "exec_xyz789",
    "steps": [
      {
        "id": "step_1",
        "action_id": "validate_order",
        "status": "completed",
        "started_at": "2024-01-20T17:00:00Z",
        "completed_at": "2024-01-20T17:00:05Z",
        "duration_ms": 5000,
        "input": {...},
        "output": {...},
        "error": null
      }
    ]
  }
}
```

---

#### Get Execution Statistics
```http
GET /api/v1/executions/stats
```

Retrieves aggregated execution statistics.

**Query Parameters:**
- `workflow_id` (string, optional): Filter by workflow ID
- `from` (string, optional): Start date
- `to` (string, optional): End date

**Response 200:**
```json
{
  "total_executions": 1250,
  "completed": 1100,
  "failed": 125,
  "running": 25,
  "success_rate": 0.88,
  "average_duration_ms": 8500,
  "executions_by_day": [...]
}
```

---

### Schedules

#### List All Schedules
```http
GET /api/v1/schedules
```

Returns all schedules across all workflows for the tenant.

**Response 200:**
```json
{
  "data": [
    {
      "id": "sch_abc123",
      "workflow_id": "wf_abc123",
      "workflow_name": "Daily Report",
      "cron_expression": "0 9 * * *",
      "enabled": true,
      "next_run": "2024-01-21T09:00:00Z",
      "last_run": "2024-01-20T09:00:00Z"
    }
  ]
}
```

---

#### Create Schedule
```http
POST /api/v1/workflows/{workflowID}/schedules
```

Creates a new schedule for a workflow.

**Path Parameters:**
- `workflowID` (string, required): Workflow identifier

**Request Body:**
```json
{
  "cron_expression": "0 9 * * *",
  "timezone": "America/New_York",
  "enabled": true,
  "payload": {
    "report_type": "daily"
  }
}
```

**Response 201:**
```json
{
  "data": {
    "id": "sch_new123",
    "workflow_id": "wf_abc123",
    "cron_expression": "0 9 * * *",
    "next_run": "2024-01-21T09:00:00Z",
    "enabled": true
  }
}
```

---

#### Parse Cron Expression
```http
POST /api/v1/schedules/parse-cron
```

Validates and parses a cron expression, returning human-readable description and next execution times.

**Request Body:**
```json
{
  "cron_expression": "0 9 * * *",
  "timezone": "America/New_York"
}
```

**Response 200:**
```json
{
  "valid": true,
  "description": "At 09:00 AM every day",
  "next_5_runs": [
    "2024-01-21T09:00:00-05:00",
    "2024-01-22T09:00:00-05:00",
    "2024-01-23T09:00:00-05:00",
    "2024-01-24T09:00:00-05:00",
    "2024-01-25T09:00:00-05:00"
  ]
}
```

---

### Credentials

#### List Credentials
```http
GET /api/v1/credentials
```

Returns all credentials for the tenant (values are not included).

**Response 200:**
```json
{
  "data": [
    {
      "id": "cred_abc123",
      "name": "Stripe API Key",
      "type": "api_key",
      "description": "Production Stripe key",
      "created_at": "2024-01-15T10:00:00Z",
      "last_rotated": "2024-01-15T10:00:00Z",
      "expires_at": null
    }
  ]
}
```

---

#### Create Credential
```http
POST /api/v1/credentials
```

Creates a new encrypted credential.

**Request Body:**
```json
{
  "name": "AWS Access Key",
  "type": "aws",
  "description": "Production AWS credentials",
  "value": {
    "access_key_id": "AKIA...",
    "secret_access_key": "..."
  },
  "expires_at": "2025-01-01T00:00:00Z"
}
```

**Response 201:**
```json
{
  "data": {
    "id": "cred_new123",
    "name": "AWS Access Key",
    "type": "aws",
    "created_at": "2024-01-20T16:00:00Z"
  }
}
```

---

#### Get Credential Value
```http
GET /api/v1/credentials/{credentialID}/value
```

Retrieves the decrypted credential value. **This endpoint is audited.**

**Path Parameters:**
- `credentialID` (string, required): Credential identifier

**Response 200:**
```json
{
  "data": {
    "id": "cred_abc123",
    "value": {
      "access_key_id": "AKIA...",
      "secret_access_key": "..."
    }
  }
}
```

---

#### Rotate Credential
```http
POST /api/v1/credentials/{credentialID}/rotate
```

Rotates a credential to a new value, preserving the old value for a grace period.

**Path Parameters:**
- `credentialID` (string, required): Credential identifier

**Request Body:**
```json
{
  "new_value": {
    "access_key_id": "AKIA...",
    "secret_access_key": "..."
  }
}
```

**Response 200:**
```json
{
  "data": {
    "id": "cred_abc123",
    "version": 2,
    "rotated_at": "2024-01-20T16:30:00Z"
  }
}
```

---

### Metrics

#### Get Execution Trends
```http
GET /api/v1/metrics/trends
```

Returns execution trends over time.

**Query Parameters:**
- `from` (string, optional): Start date
- `to` (string, optional): End date
- `granularity` (string, optional): hour, day, week, month

**Response 200:**
```json
{
  "data": [
    {
      "timestamp": "2024-01-20T00:00:00Z",
      "total": 150,
      "completed": 135,
      "failed": 15
    }
  ]
}
```

---

#### Get Duration Statistics
```http
GET /api/v1/metrics/duration
```

Returns execution duration statistics by workflow.

**Response 200:**
```json
{
  "data": [
    {
      "workflow_id": "wf_abc123",
      "workflow_name": "Order Processing",
      "avg_duration_ms": 8500,
      "min_duration_ms": 1200,
      "max_duration_ms": 45000,
      "p50_duration_ms": 7800,
      "p95_duration_ms": 18000,
      "p99_duration_ms": 35000
    }
  ]
}
```

---

### Admin (Tenant Management)

These endpoints require admin role and do not use tenant context.

#### List Tenants
```http
GET /api/v1/admin/tenants
```

Returns all tenants (admin only).

**Response 200:**
```json
{
  "data": [
    {
      "id": "tenant_abc",
      "name": "Acme Corp",
      "status": "active",
      "created_at": "2024-01-01T00:00:00Z",
      "quotas": {
        "max_workflows": 100,
        "max_executions_per_day": 10000
      }
    }
  ]
}
```

---

#### Create Tenant
```http
POST /api/v1/admin/tenants
```

Creates a new tenant (admin only).

**Request Body:**
```json
{
  "name": "New Corp",
  "admin_email": "admin@newcorp.com",
  "quotas": {
    "max_workflows": 50,
    "max_executions_per_day": 5000
  }
}
```

**Response 201:**
```json
{
  "data": {
    "id": "tenant_new123",
    "name": "New Corp",
    "created_at": "2024-01-20T16:00:00Z"
  }
}
```

---

#### Update Tenant Quotas
```http
PUT /api/v1/admin/tenants/{tenantID}/quotas
```

Updates tenant quotas (admin only).

**Path Parameters:**
- `tenantID` (string, required): Tenant identifier

**Request Body:**
```json
{
  "max_workflows": 200,
  "max_executions_per_day": 20000,
  "max_api_requests_per_minute": 1000
}
```

**Response 200:**
```json
{
  "data": {
    "id": "tenant_abc",
    "quotas": {
      "max_workflows": 200,
      "max_executions_per_day": 20000,
      "max_api_requests_per_minute": 1000
    },
    "updated_at": "2024-01-20T16:30:00Z"
  }
}
```

---

### WebSocket

#### Connect to Execution Stream
```http
GET /api/v1/ws/executions/{executionID}
```

Establishes a WebSocket connection to receive real-time execution updates.

**Protocol:** WebSocket

**Path Parameters:**
- `executionID` (string, required): Execution identifier

**Message Format:**
```json
{
  "type": "execution_update",
  "execution_id": "exec_xyz789",
  "status": "running",
  "step": {
    "id": "step_2",
    "status": "completed",
    "output": {...}
  },
  "timestamp": "2024-01-20T17:00:10Z"
}
```

**Example (JavaScript):**
```javascript
const ws = new WebSocket('ws://localhost:8080/api/v1/ws/executions/exec_xyz789');

ws.onmessage = (event) => {
  const update = JSON.parse(event.data);
  console.log('Execution update:', update);
};
```

---

## Webhook Signature Verification

When using webhook `authType: "signature"`, incoming webhook requests include an HMAC signature for verification.

### Verifying Signatures

**Header:** `X-Webhook-Signature`

**Format:** `sha256=<hex_encoded_signature>`

**Example (Node.js):**
```javascript
const crypto = require('crypto');

function verifyWebhookSignature(payload, signature, secret) {
  const hmac = crypto.createHmac('sha256', secret);
  hmac.update(payload);
  const expectedSignature = 'sha256=' + hmac.digest('hex');

  return crypto.timingSafeEqual(
    Buffer.from(signature),
    Buffer.from(expectedSignature)
  );
}

// Usage
const isValid = verifyWebhookSignature(
  req.body,
  req.headers['x-webhook-signature'],
  webhookSecret
);
```

**Example (Python):**
```python
import hmac
import hashlib

def verify_webhook_signature(payload, signature, secret):
    expected_signature = 'sha256=' + hmac.new(
        secret.encode(),
        payload.encode(),
        hashlib.sha256
    ).hexdigest()

    return hmac.compare_digest(signature, expected_signature)
```

---

## Common Patterns

### Polling for Execution Completion

```bash
# Start execution
EXECUTION_ID=$(curl -X POST http://localhost:8080/api/v1/workflows/wf_abc123/execute \
  -H "X-Tenant-ID: tenant_xyz" | jq -r '.data.execution_id')

# Poll for completion
while true; do
  STATUS=$(curl -s http://localhost:8080/api/v1/executions/$EXECUTION_ID \
    -H "X-Tenant-ID: tenant_xyz" | jq -r '.data.status')

  if [ "$STATUS" == "completed" ] || [ "$STATUS" == "failed" ]; then
    echo "Execution $STATUS"
    break
  fi

  sleep 2
done
```

### Batch Operations

For bulk operations, use the bulk endpoints:

```bash
# Bulk delete workflows
curl -X POST http://localhost:8080/api/v1/workflows/bulk-delete \
  -H "X-Tenant-ID: tenant_xyz" \
  -H "Content-Type: application/json" \
  -d '{
    "workflow_ids": ["wf_1", "wf_2", "wf_3"]
  }'
```

---

## SDKs and Tools

### Official SDKs

- **Go SDK**: `go get github.com/gorax/gorax-go`
- **Node.js SDK**: `npm install @gorax/client`
- **Python SDK**: `pip install gorax`

### Postman Collection

Import the Postman collection from `/docs/api/gorax.postman_collection.json` for easy API exploration.

### OpenAPI Spec

Download the OpenAPI specification:
- JSON: `/docs/api/swagger.json`
- YAML: `/docs/api/swagger.yaml`

---

## Support

- **Documentation**: https://docs.gorax.io
- **GitHub**: https://github.com/gorax/gorax
- **Issues**: https://github.com/gorax/gorax/issues
- **Email**: support@gorax.io

---

**Last Updated:** 2024-01-20
**API Version:** 1.0
