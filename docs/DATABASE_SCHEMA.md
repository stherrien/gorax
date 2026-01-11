# Database Schema Documentation

## Overview

Gorax uses **PostgreSQL 15+** as its primary database with a robust multi-tenant architecture. This document provides comprehensive documentation of the database schema, relationships, and best practices.

### Technology Stack
- **Database**: PostgreSQL 15+ with extensions (uuid-ossp)
- **Migration Tool**: Custom SQL migrations (numbered)
- **ORM/Query Layer**: sqlx with manual queries
- **Connection Pool**: pgx

### Schema Design Principles

1. **Multi-Tenancy First**: All tenant-scoped tables include `tenant_id` with Row-Level Security (RLS)
2. **UUID Primary Keys**: All tables use UUIDs for globally unique identifiers
3. **Timezone-Aware Timestamps**: All timestamps use `TIMESTAMPTZ` (UTC)
4. **JSONB for Flexibility**: Workflow definitions, metadata, and configurations stored as JSONB
5. **Audit Trail**: Comprehensive audit logging with `created_at`, `updated_at`, and audit tables
6. **Cascade Deletes**: Foreign keys configured for proper cascade behavior
7. **Performance**: Strategic indexes including composite, partial, and GIN indexes

---

## Multi-Tenancy Model

### Tenant Isolation Strategy

Gorax implements **strict tenant isolation** using PostgreSQL Row-Level Security (RLS) policies. Every query is automatically scoped to the current tenant context.

```sql
-- Set tenant context before queries
SELECT set_config('app.current_tenant_id', $1, false);

-- All queries are automatically filtered by tenant_id through RLS policies
```

### Tenant Context Flow

```
┌─────────────────┐
│   HTTP Request  │
│   (Tenant ID)   │
└────────┬────────┘
         │
         ↓
┌─────────────────┐
│   Middleware    │
│ Extract Tenant  │
└────────┬────────┘
         │
         ↓
┌─────────────────┐
│   Database Hook │
│ set_config()    │
└────────┬────────┘
         │
         ↓
┌─────────────────┐
│  RLS Policies   │
│  Filter Rows    │
└─────────────────┘
```

**Implementation**: See `/Users/shawntherrien/Projects/gorax/internal/database/tenant_hooks.go`

---

## Core Tables

### 1. tenants

The root table for multi-tenant architecture. Each tenant represents an organization using Gorax.

```sql
CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    subdomain VARCHAR(63) UNIQUE NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    tier VARCHAR(50) NOT NULL DEFAULT 'free',
    settings JSONB NOT NULL DEFAULT '{}',
    quotas JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

**Columns:**
- `id`: Unique tenant identifier
- `name`: Display name for the tenant
- `subdomain`: Unique subdomain for tenant isolation (e.g., `acme.gorax.io`)
- `status`: Tenant status (`active`, `suspended`, `cancelled`)
- `tier`: Subscription tier (`free`, `starter`, `professional`, `enterprise`)
- `settings`: JSONB configuration including retention policies, feature flags
- `quotas`: JSONB quota limits (workflows, executions, storage)
- `created_at`, `updated_at`: Standard audit timestamps

**Settings JSONB Structure:**
```json
{
  "retention_days": 90,
  "retention_enabled": true,
  "features": {
    "ai_enabled": true,
    "marketplace_enabled": true
  }
}
```

**Indexes:**
- `subdomain` - UNIQUE index for subdomain lookups

**Example Queries:**
```sql
-- Get tenant by subdomain
SELECT * FROM tenants WHERE subdomain = 'acme';

-- Update tenant tier
UPDATE tenants SET tier = 'professional' WHERE id = $1;

-- Get active tenants with AI enabled
SELECT * FROM tenants
WHERE status = 'active'
  AND settings->>'ai_enabled' = 'true';
```

---

### 2. users

User accounts linked to Ory Kratos identities. Users belong to a single tenant.

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    kratos_identity_id VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'member',
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_tenant_id ON users(tenant_id);
CREATE INDEX idx_users_kratos_identity_id ON users(kratos_identity_id);
```

**Columns:**
- `id`: Internal user identifier
- `tenant_id`: Foreign key to tenants table (CASCADE delete)
- `kratos_identity_id`: Ory Kratos identity UUID (external auth system)
- `email`: User email address
- `role`: User role (`admin`, `member`, `viewer`) - basic role, extended by RBAC
- `status`: Account status (`active`, `suspended`, `deleted`)

**Relationships:**
- `tenant_id` → `tenants.id` (CASCADE)
- One-to-many: `users` ← `user_roles` (RBAC)

**RLS Policy:**
```sql
CREATE POLICY tenant_isolation_users ON users
    USING (tenant_id = current_setting('app.current_tenant_id', true)::UUID);
```

---

### 3. workflows

Workflow definitions stored as JSONB node graphs. The heart of the workflow automation system.

```sql
CREATE TABLE workflows (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    definition JSONB NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    version INTEGER NOT NULL DEFAULT 1,
    created_by UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_workflow_name_per_tenant UNIQUE (tenant_id, name)
);

CREATE INDEX idx_workflows_tenant_id ON workflows(tenant_id);
CREATE INDEX idx_workflows_status ON workflows(status);
CREATE INDEX idx_workflows_created_by ON workflows(created_by);
```

**Columns:**
- `id`: Unique workflow identifier
- `tenant_id`: Owner tenant
- `name`: Workflow name (unique per tenant)
- `description`: Optional description
- `definition`: JSONB workflow graph (nodes, edges, configuration)
- `status`: Workflow status (`draft`, `active`, `inactive`, `archived`)
- `version`: Current version number (incremented on publish)
- `created_by`: User ID who created the workflow

**Definition JSONB Structure:**
```json
{
  "nodes": [
    {
      "id": "trigger-1",
      "type": "webhook_trigger",
      "position": { "x": 100, "y": 100 },
      "config": {
        "path": "/webhooks/orders",
        "method": "POST"
      }
    },
    {
      "id": "action-1",
      "type": "http_request",
      "position": { "x": 300, "y": 100 },
      "config": {
        "url": "https://api.example.com/notify",
        "method": "POST",
        "body": "{{trigger.body}}"
      }
    }
  ],
  "edges": [
    {
      "id": "edge-1",
      "source": "trigger-1",
      "target": "action-1"
    }
  ]
}
```

**Indexes:**
- Composite: `(tenant_id, status)` - for filtering workflows by status
- Single: `created_by` - for user-specific queries

**Example Queries:**
```sql
-- Get active workflows for tenant
SELECT * FROM workflows
WHERE status = 'active'
ORDER BY updated_at DESC;

-- Search workflows by name (uses RLS)
SELECT id, name, status, updated_at
FROM workflows
WHERE name ILIKE '%order%';

-- Get workflow with specific node type
SELECT * FROM workflows
WHERE definition @> '{"nodes": [{"type": "webhook_trigger"}]}';
```

---

### 4. workflow_versions

Version history for workflows. Enables rollback and audit trail.

```sql
CREATE TABLE workflow_versions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    version INTEGER NOT NULL,
    definition JSONB NOT NULL,
    created_by UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_workflow_version UNIQUE (workflow_id, version)
);

CREATE INDEX idx_workflow_versions_workflow_id ON workflow_versions(workflow_id);
```

**Columns:**
- `id`: Version record identifier
- `workflow_id`: Parent workflow
- `version`: Version number (unique per workflow)
- `definition`: Snapshot of workflow definition at this version
- `created_by`: User who created this version

**Relationships:**
- `workflow_id` → `workflows.id` (CASCADE)

**Example Queries:**
```sql
-- Get version history for workflow
SELECT version, created_by, created_at
FROM workflow_versions
WHERE workflow_id = $1
ORDER BY version DESC;

-- Rollback to previous version
SELECT definition FROM workflow_versions
WHERE workflow_id = $1 AND version = $2;
```

---

### 5. executions

Execution records for workflow runs. Partitionable by date for scale.

```sql
CREATE TABLE executions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    workflow_version INTEGER NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    trigger_type VARCHAR(50) NOT NULL,
    trigger_data JSONB,
    output_data JSONB,
    error_message TEXT,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    retention_until TIMESTAMPTZ,
    parent_execution_id UUID REFERENCES executions(id) ON DELETE CASCADE,
    execution_depth INTEGER NOT NULL DEFAULT 0
);
```

**Columns:**
- `id`: Unique execution identifier
- `tenant_id`: Owner tenant (for RLS)
- `workflow_id`: Executed workflow
- `workflow_version`: Workflow version executed (immutable)
- `status`: Execution status (`pending`, `running`, `completed`, `failed`, `cancelled`)
- `trigger_type`: How workflow was triggered (`webhook`, `schedule`, `manual`, `api`)
- `trigger_data`: JSONB input data from trigger
- `output_data`: JSONB final output data
- `error_message`: Error message if failed
- `started_at`: When execution started
- `completed_at`: When execution finished
- `retention_until`: Scheduled deletion date (retention policy)
- `parent_execution_id`: Parent execution if this is a sub-workflow
- `execution_depth`: Nesting depth (0 = root, 1+ = sub-workflow)

**Indexes:**
```sql
CREATE INDEX idx_executions_tenant_id ON executions(tenant_id);
CREATE INDEX idx_executions_workflow_id ON executions(workflow_id);
CREATE INDEX idx_executions_status ON executions(status);
CREATE INDEX idx_executions_created_at ON executions(created_at DESC);

-- Composite indexes for filtering
CREATE INDEX idx_executions_cursor_pagination ON executions(tenant_id, created_at DESC, id);
CREATE INDEX idx_executions_tenant_status ON executions(tenant_id, status, created_at DESC);
CREATE INDEX idx_executions_tenant_workflow ON executions(tenant_id, workflow_id, created_at DESC);
CREATE INDEX idx_executions_tenant_trigger_type ON executions(tenant_id, trigger_type, created_at DESC);

-- Partial index for retention policy
CREATE INDEX idx_executions_retention_until ON executions(retention_until)
WHERE retention_until IS NOT NULL;

-- Sub-workflow indexes
CREATE INDEX idx_executions_parent_id ON executions(parent_execution_id);
CREATE INDEX idx_executions_depth ON executions(execution_depth);
```

**Relationships:**
- `tenant_id` → `tenants.id` (CASCADE)
- `workflow_id` → `workflows.id` (CASCADE)
- `parent_execution_id` → `executions.id` (CASCADE) - self-referencing

**Partitioning Strategy:**

For high-volume tenants, partition by month:

```sql
-- Example monthly partitioning
CREATE TABLE executions_2024_01 PARTITION OF executions
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');
```

**Example Queries:**
```sql
-- Get recent executions with pagination
SELECT id, workflow_id, status, created_at
FROM executions
ORDER BY created_at DESC
LIMIT 50;

-- Cursor-based pagination (for infinite scroll)
SELECT id, workflow_id, status, created_at
FROM executions
WHERE created_at < $1
ORDER BY created_at DESC
LIMIT 50;

-- Get executions by status
SELECT * FROM executions
WHERE status = 'failed'
  AND created_at >= NOW() - INTERVAL '7 days';

-- Get sub-workflow executions
SELECT * FROM executions
WHERE parent_execution_id = $1
ORDER BY created_at;

-- Get execution tree (recursive)
WITH RECURSIVE execution_tree AS (
    SELECT id, workflow_id, parent_execution_id, execution_depth, status
    FROM executions
    WHERE id = $1
    UNION ALL
    SELECT e.id, e.workflow_id, e.parent_execution_id, e.execution_depth, e.status
    FROM executions e
    INNER JOIN execution_tree et ON e.parent_execution_id = et.id
)
SELECT * FROM execution_tree;
```

---

### 6. step_executions

Individual step execution records within a workflow execution.

```sql
CREATE TABLE step_executions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    execution_id UUID NOT NULL REFERENCES executions(id) ON DELETE CASCADE,
    node_id VARCHAR(255) NOT NULL,
    node_type VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    input_data JSONB,
    output_data JSONB,
    error_message TEXT,
    retry_count INTEGER NOT NULL DEFAULT 0,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    duration_ms INTEGER
);

CREATE INDEX idx_step_executions_execution_id ON step_executions(execution_id);
CREATE INDEX idx_step_executions_node_id ON step_executions(node_id);
```

**Columns:**
- `id`: Unique step execution identifier
- `execution_id`: Parent execution
- `node_id`: Node identifier from workflow definition
- `node_type`: Type of node (`http_request`, `transform`, `condition`, etc.)
- `status`: Step status (`pending`, `running`, `completed`, `failed`, `skipped`)
- `input_data`: JSONB input to the step
- `output_data`: JSONB output from the step
- `error_message`: Error message if failed
- `retry_count`: Number of retries attempted
- `started_at`, `completed_at`: Execution timestamps
- `duration_ms`: Execution duration in milliseconds

**Relationships:**
- `execution_id` → `executions.id` (CASCADE)

**Example Queries:**
```sql
-- Get all steps for an execution
SELECT node_id, node_type, status, duration_ms
FROM step_executions
WHERE execution_id = $1
ORDER BY started_at;

-- Find slowest steps
SELECT node_id, node_type, AVG(duration_ms) as avg_duration
FROM step_executions
WHERE status = 'completed'
GROUP BY node_id, node_type
ORDER BY avg_duration DESC
LIMIT 10;

-- Get failed steps with retries
SELECT * FROM step_executions
WHERE status = 'failed' AND retry_count > 0
ORDER BY completed_at DESC;
```

---

### 7. credentials

Encrypted credentials using envelope encryption (AES-256-GCM + AWS KMS).

```sql
CREATE TABLE credentials (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(100) NOT NULL,
    description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'active',

    -- Envelope encryption fields
    encrypted_dek BYTEA NOT NULL,
    ciphertext BYTEA NOT NULL,
    nonce BYTEA NOT NULL,
    auth_tag BYTEA NOT NULL,
    kms_key_id VARCHAR(255) NOT NULL,

    metadata JSONB NOT NULL DEFAULT '{}',
    expires_at TIMESTAMPTZ,

    created_by UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMPTZ,

    CONSTRAINT unique_credential_name_per_tenant UNIQUE (tenant_id, name)
);

CREATE INDEX idx_credentials_tenant_id ON credentials(tenant_id);
CREATE INDEX idx_credentials_type ON credentials(type);
CREATE INDEX idx_credentials_kms_key_id ON credentials(kms_key_id);
CREATE INDEX idx_credentials_last_used_at ON credentials(last_used_at DESC);
CREATE INDEX idx_credentials_expires_at ON credentials(expires_at)
WHERE expires_at IS NOT NULL;
```

**Columns:**
- `id`: Credential identifier
- `tenant_id`: Owner tenant
- `name`: Credential name (unique per tenant)
- `type`: Credential type (`api_key`, `oauth2`, `basic_auth`, `custom`)
- `status`: Status (`active`, `inactive`, `revoked`)
- `encrypted_dek`: Data Encryption Key encrypted by KMS
- `ciphertext`: Credential data encrypted with DEK using AES-256-GCM
- `nonce`: 12-byte nonce for AES-GCM
- `auth_tag`: 16-byte authentication tag for AES-GCM
- `kms_key_id`: AWS KMS key ID/ARN used for DEK encryption
- `metadata`: Additional metadata (tags, source)
- `expires_at`: Optional expiration timestamp
- `last_used_at`: Last usage timestamp (updated on access)

**Encryption Flow:**
```
1. Generate random DEK (Data Encryption Key)
2. Encrypt credential data with DEK using AES-256-GCM
3. Encrypt DEK with AWS KMS master key
4. Store encrypted_dek, ciphertext, nonce, auth_tag
```

**Credential Types:**
- `api_key`: Simple API key
- `oauth2`: OAuth 2.0 tokens (access_token, refresh_token, expires_at)
- `basic_auth`: Username/password
- `custom`: Custom credential format

**Example Queries:**
```sql
-- List credentials for tenant
SELECT id, name, type, status, last_used_at
FROM credentials
ORDER BY name;

-- Find expired credentials
SELECT * FROM credentials
WHERE expires_at < NOW() AND status = 'active';

-- Get unused credentials (potential cleanup)
SELECT * FROM credentials
WHERE last_used_at < NOW() - INTERVAL '90 days'
   OR last_used_at IS NULL;
```

---

### 8. credential_rotations

Audit log for credential key rotations.

```sql
CREATE TABLE credential_rotations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    credential_id UUID NOT NULL REFERENCES credentials(id) ON DELETE CASCADE,
    previous_key_id VARCHAR(255) NOT NULL,
    new_key_id VARCHAR(255) NOT NULL,
    rotated_by UUID NOT NULL,
    rotated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reason VARCHAR(255)
);

CREATE INDEX idx_credential_rotations_credential_id ON credential_rotations(credential_id);
CREATE INDEX idx_credential_rotations_rotated_at ON credential_rotations(rotated_at DESC);
```

**Purpose:** Track KMS key rotations for credentials (compliance requirement).

---

### 9. credential_access_log

Audit log for credential access (GDPR compliance).

```sql
CREATE TABLE credential_access_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    credential_id UUID NOT NULL REFERENCES credentials(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    accessed_by UUID NOT NULL,
    access_type VARCHAR(50) NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    success BOOLEAN NOT NULL DEFAULT true,
    error_message TEXT,
    accessed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_credential_access_log_credential_id ON credential_access_log(credential_id);
CREATE INDEX idx_credential_access_log_tenant_id ON credential_access_log(tenant_id);
CREATE INDEX idx_credential_access_log_accessed_at ON credential_access_log(accessed_at DESC);
```

**Access Types:**
- `read`: Credential decrypted and read
- `update`: Credential updated
- `rotate`: Key rotated
- `delete`: Credential deleted

**Example Queries:**
```sql
-- Get access log for credential
SELECT accessed_by, access_type, accessed_at, success
FROM credential_access_log
WHERE credential_id = $1
ORDER BY accessed_at DESC;

-- Failed access attempts (security audit)
SELECT credential_id, accessed_by, ip_address, error_message
FROM credential_access_log
WHERE success = false
  AND accessed_at >= NOW() - INTERVAL '24 hours';
```

---

### 10. webhooks

Webhook endpoints that trigger workflows.

```sql
CREATE TABLE webhooks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    node_id VARCHAR(255) NOT NULL,
    name VARCHAR(255),
    description TEXT,
    path VARCHAR(255) NOT NULL,
    secret VARCHAR(255),
    auth_type VARCHAR(50) NOT NULL DEFAULT 'none',
    enabled BOOLEAN NOT NULL DEFAULT true,
    priority INTEGER NOT NULL DEFAULT 1,
    last_triggered_at TIMESTAMPTZ,
    trigger_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_webhook_path UNIQUE (path)
);

CREATE INDEX idx_webhooks_tenant_id ON webhooks(tenant_id);
CREATE INDEX idx_webhooks_workflow_id ON webhooks(workflow_id);
CREATE INDEX idx_webhooks_path ON webhooks(path);
CREATE INDEX idx_webhooks_priority ON webhooks(priority DESC);
```

**Columns:**
- `id`: Webhook identifier
- `tenant_id`: Owner tenant
- `workflow_id`: Target workflow to execute
- `node_id`: Node ID in workflow definition
- `name`, `description`: Display information
- `path`: URL path (e.g., `/webhooks/orders`)
- `secret`: Optional HMAC secret for signature verification
- `auth_type`: Authentication type (`none`, `secret`, `basic`, `bearer`)
- `enabled`: Whether webhook is active
- `priority`: Priority for processing (1=lowest, higher=more important)
- `last_triggered_at`: Last successful trigger timestamp
- `trigger_count`: Total number of successful triggers

**Webhook URL Format:**
```
https://api.gorax.io/webhooks/{path}
```

**Example Queries:**
```sql
-- Get webhook by path
SELECT * FROM webhooks WHERE path = $1 AND enabled = true;

-- List webhooks for workflow
SELECT id, name, path, enabled, trigger_count
FROM webhooks
WHERE workflow_id = $1;

-- Get high-priority webhooks
SELECT * FROM webhooks
WHERE priority >= 5 AND enabled = true
ORDER BY priority DESC;
```

---

### 11. webhook_events

Audit log for all webhook deliveries.

```sql
CREATE TABLE webhook_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    webhook_id UUID NOT NULL REFERENCES webhooks(id) ON DELETE CASCADE,
    execution_id UUID REFERENCES executions(id) ON DELETE SET NULL,
    request_method VARCHAR(10) NOT NULL,
    request_headers JSONB NOT NULL DEFAULT '{}',
    request_body JSONB NOT NULL DEFAULT '{}',
    response_status INTEGER,
    processing_time_ms INTEGER,
    status VARCHAR(20) NOT NULL DEFAULT 'received',
    error_message TEXT,
    filtered_reason TEXT,
    replay_count INTEGER NOT NULL DEFAULT 0,
    source_event_id UUID REFERENCES webhook_events(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_webhook_events_tenant_id ON webhook_events(tenant_id);
CREATE INDEX idx_webhook_events_webhook_id ON webhook_events(webhook_id);
CREATE INDEX idx_webhook_events_execution_id ON webhook_events(execution_id);
CREATE INDEX idx_webhook_events_status ON webhook_events(status);
CREATE INDEX idx_webhook_events_created_at ON webhook_events(created_at DESC);
CREATE INDEX idx_webhook_events_tenant_status_created
    ON webhook_events(tenant_id, status, created_at DESC);
```

**Columns:**
- `id`: Event identifier
- `tenant_id`: Owner tenant
- `webhook_id`: Target webhook
- `execution_id`: Created execution (if processed)
- `request_method`: HTTP method (POST, PUT, etc.)
- `request_headers`: JSONB request headers
- `request_body`: JSONB request body
- `response_status`: HTTP response status
- `processing_time_ms`: Processing duration
- `status`: Event status (`received`, `processed`, `filtered`, `failed`)
- `error_message`: Error message if failed
- `filtered_reason`: Reason if filtered out
- `replay_count`: Number of times replayed
- `source_event_id`: Original event if this is a replay

**Status Values:**
- `received`: Event received
- `processed`: Successfully triggered workflow
- `filtered`: Filtered out by filter rules
- `failed`: Processing failed

**Example Queries:**
```sql
-- Get recent webhook events
SELECT id, webhook_id, status, created_at
FROM webhook_events
ORDER BY created_at DESC
LIMIT 100;

-- Get failed events for replay
SELECT * FROM webhook_events
WHERE status = 'failed'
  AND replay_count < 3
  AND created_at >= NOW() - INTERVAL '24 hours';

-- Webhook health check
SELECT
    webhook_id,
    COUNT(*) as total,
    COUNT(*) FILTER (WHERE status = 'processed') as processed,
    COUNT(*) FILTER (WHERE status = 'failed') as failed,
    AVG(processing_time_ms) as avg_processing_time
FROM webhook_events
WHERE created_at >= NOW() - INTERVAL '7 days'
GROUP BY webhook_id;
```

---

### 12. webhook_filters

JSONPath-based filtering rules for webhooks.

```sql
CREATE TABLE webhook_filters (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    webhook_id UUID NOT NULL REFERENCES webhooks(id) ON DELETE CASCADE,
    field_path VARCHAR(255) NOT NULL,
    operator VARCHAR(20) NOT NULL,
    value JSONB NOT NULL,
    logic_group INTEGER NOT NULL DEFAULT 0,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_webhook_filters_webhook_id ON webhook_filters(webhook_id);
CREATE INDEX idx_webhook_filters_enabled ON webhook_filters(enabled);
CREATE INDEX idx_webhook_filters_webhook_enabled ON webhook_filters(webhook_id, enabled);
```

**Columns:**
- `id`: Filter identifier
- `webhook_id`: Parent webhook
- `field_path`: JSONPath to field (e.g., `$.order.status`)
- `operator`: Filter operator (see below)
- `value`: JSONB value to compare
- `logic_group`: Group number for AND/OR logic
- `enabled`: Whether filter is active

**Operators:**
- `equals`, `not_equals`
- `contains`, `not_contains`
- `regex`
- `gt`, `gte`, `lt`, `lte`
- `in`, `not_in`
- `exists`, `not_exists`

**Logic Groups:**
Filters with the same `logic_group` are ORed together. Different groups are ANDed.

```
(group_0_filter_1 OR group_0_filter_2)
AND
(group_1_filter_1 OR group_1_filter_2)
```

**Example Filters:**
```sql
-- Filter: order.status = 'completed'
INSERT INTO webhook_filters (webhook_id, field_path, operator, value, logic_group)
VALUES ($1, '$.order.status', 'equals', '"completed"', 0);

-- Filter: order.total >= 100
INSERT INTO webhook_filters (webhook_id, field_path, operator, value, logic_group)
VALUES ($1, '$.order.total', 'gte', '100', 0);

-- Complex: (status=completed OR status=shipped) AND total>=100
INSERT INTO webhook_filters (webhook_id, field_path, operator, value, logic_group)
VALUES
    ($1, '$.order.status', 'equals', '"completed"', 0),
    ($1, '$.order.status', 'equals', '"shipped"', 0),
    ($1, '$.order.total', 'gte', '100', 1);
```

---

### 13. schedules

Scheduled workflow executions (cron-based).

```sql
CREATE TABLE schedules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    cron_expression VARCHAR(255) NOT NULL,
    timezone VARCHAR(100) NOT NULL DEFAULT 'UTC',
    enabled BOOLEAN NOT NULL DEFAULT true,
    next_run_at TIMESTAMPTZ,
    last_run_at TIMESTAMPTZ,
    last_execution_id UUID,
    created_by UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_schedule_name_per_tenant UNIQUE (tenant_id, workflow_id, name)
);

CREATE INDEX idx_schedules_tenant_id ON schedules(tenant_id);
CREATE INDEX idx_schedules_workflow_id ON schedules(workflow_id);
CREATE INDEX idx_schedules_enabled ON schedules(enabled);
CREATE INDEX idx_schedules_next_run_at ON schedules(next_run_at) WHERE enabled = true;
```

**Columns:**
- `id`: Schedule identifier
- `tenant_id`: Owner tenant
- `workflow_id`: Workflow to execute
- `name`: Schedule name
- `cron_expression`: Cron expression (e.g., `0 9 * * *` for daily at 9am)
- `timezone`: Timezone for cron evaluation (IANA format)
- `enabled`: Whether schedule is active
- `next_run_at`: Next scheduled execution time
- `last_run_at`: Last execution time
- `last_execution_id`: Last execution record

**Cron Expression Examples:**
```
0 9 * * *       - Daily at 9:00 AM
0 */6 * * *     - Every 6 hours
0 0 * * 0       - Weekly on Sunday at midnight
0 0 1 * *       - Monthly on the 1st at midnight
*/15 * * * *    - Every 15 minutes
```

**Example Queries:**
```sql
-- Get schedules ready to run
SELECT * FROM schedules
WHERE enabled = true
  AND next_run_at <= NOW()
ORDER BY next_run_at;

-- Update next run time after execution
UPDATE schedules
SET last_run_at = NOW(),
    last_execution_id = $2,
    next_run_at = $3
WHERE id = $1;
```

---

### 14. workflow_templates

Reusable workflow templates (private to tenant).

```sql
CREATE TABLE workflow_templates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100),
    definition JSONB NOT NULL,
    tags TEXT[],
    is_public BOOLEAN NOT NULL DEFAULT FALSE,
    created_by UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_template_name_per_tenant UNIQUE (tenant_id, name)
);

CREATE INDEX idx_templates_tenant ON workflow_templates(tenant_id);
CREATE INDEX idx_templates_category ON workflow_templates(category);
CREATE INDEX idx_templates_tags ON workflow_templates USING GIN(tags);
CREATE INDEX idx_templates_is_public ON workflow_templates(is_public);
```

**Columns:**
- `id`: Template identifier
- `tenant_id`: Owner tenant (NULL for global templates)
- `name`: Template name
- `description`: Template description
- `category`: Category (e.g., `data_processing`, `notifications`, `integrations`)
- `definition`: JSONB workflow definition
- `tags`: Array of tags for search
- `is_public`: Whether template is public (shareable)
- `created_by`: Creator user ID

**RLS Policy:**
```sql
-- Users can see their tenant's templates + public templates
CREATE POLICY tenant_isolation_templates ON workflow_templates
    USING (
        tenant_id = current_setting('app.current_tenant_id', true)::UUID
        OR is_public = true
    );
```

---

### 15. marketplace_templates

Public marketplace templates (shared across tenants).

```sql
CREATE TABLE marketplace_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT NOT NULL,
    category VARCHAR(100) NOT NULL,
    definition JSONB NOT NULL,
    tags TEXT[] DEFAULT '{}',
    author_id VARCHAR(255) NOT NULL,
    author_name VARCHAR(255) NOT NULL,
    version VARCHAR(50) NOT NULL,
    download_count INTEGER DEFAULT 0,
    average_rating DECIMAL(3,2) DEFAULT 0.0,
    total_ratings INTEGER DEFAULT 0,
    is_verified BOOLEAN DEFAULT FALSE,
    source_tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL,
    source_template_id UUID,
    published_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT valid_rating CHECK (average_rating >= 0 AND average_rating <= 5)
);

CREATE INDEX idx_marketplace_templates_category ON marketplace_templates(category);
CREATE INDEX idx_marketplace_templates_download_count ON marketplace_templates(download_count DESC);
CREATE INDEX idx_marketplace_templates_rating ON marketplace_templates(average_rating DESC);
CREATE INDEX idx_marketplace_templates_tags ON marketplace_templates USING GIN(tags);
```

**Columns:**
- `id`: Template identifier
- `name`: Template name (unique across marketplace)
- `description`: Detailed description
- `category`: Category for filtering
- `definition`: JSONB workflow definition
- `tags`: Array of search tags
- `author_id`, `author_name`: Template creator
- `version`: Semantic version (e.g., `1.0.0`)
- `download_count`: Installation count
- `average_rating`: Average user rating (0-5)
- `total_ratings`: Number of ratings
- `is_verified`: Whether verified by Gorax team
- `source_tenant_id`: Original tenant (for attribution)

**Related Tables:**
- `marketplace_installations`: Track installations per tenant
- `marketplace_reviews`: User reviews and ratings
- `marketplace_template_versions`: Version history

---

### 16. marketplace_installations

Track template installations by tenants.

```sql
CREATE TABLE marketplace_installations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id UUID NOT NULL REFERENCES marketplace_templates(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL,
    workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    installed_version VARCHAR(50) NOT NULL,
    installed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_tenant_template UNIQUE (tenant_id, template_id, workflow_id)
);

CREATE INDEX idx_marketplace_installations_template ON marketplace_installations(template_id);
CREATE INDEX idx_marketplace_installations_tenant ON marketplace_installations(tenant_id);
```

---

### 17. marketplace_reviews

User reviews for marketplace templates.

```sql
CREATE TABLE marketplace_reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id UUID NOT NULL REFERENCES marketplace_templates(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL,
    user_name VARCHAR(255) NOT NULL,
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    comment TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_user_template_review UNIQUE (tenant_id, user_id, template_id)
);

CREATE INDEX idx_marketplace_reviews_template ON marketplace_reviews(template_id);
CREATE INDEX idx_marketplace_reviews_tenant_user ON marketplace_reviews(tenant_id, user_id);
```

---

### 18. human_tasks

Human approval tasks within workflow executions.

```sql
CREATE TABLE human_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    execution_id UUID NOT NULL REFERENCES executions(id) ON DELETE CASCADE,
    step_id VARCHAR(100) NOT NULL,
    task_type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    assignees JSONB NOT NULL DEFAULT '[]'::JSONB,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    due_date TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    completed_by UUID REFERENCES users(id),
    response_data JSONB,
    config JSONB NOT NULL DEFAULT '{}'::JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_task_type CHECK (task_type IN ('approval', 'input', 'review')),
    CONSTRAINT chk_status CHECK (status IN ('pending', 'approved', 'rejected', 'expired', 'cancelled'))
);

CREATE INDEX idx_human_tasks_tenant ON human_tasks(tenant_id);
CREATE INDEX idx_human_tasks_execution ON human_tasks(execution_id);
CREATE INDEX idx_human_tasks_status ON human_tasks(status);
CREATE INDEX idx_human_tasks_assignee ON human_tasks USING GIN(assignees);
CREATE INDEX idx_human_tasks_due_date ON human_tasks(due_date) WHERE status = 'pending';
```

**Columns:**
- `id`: Task identifier
- `tenant_id`: Owner tenant
- `execution_id`: Parent execution (workflow paused)
- `step_id`: Step identifier in workflow
- `task_type`: Type of task (`approval`, `input`, `review`)
- `title`, `description`: Task display information
- `assignees`: JSONB array of user IDs or roles
- `status`: Task status (`pending`, `approved`, `rejected`, `expired`, `cancelled`)
- `due_date`: Optional deadline
- `completed_at`: Completion timestamp
- `completed_by`: User who completed the task
- `response_data`: JSONB response (approval reason, form data, etc.)
- `config`: Task configuration (form fields, escalation settings)

**Assignees Format:**
```json
[
  {"type": "user", "id": "uuid-123"},
  {"type": "role", "name": "approvers"}
]
```

**Example Queries:**
```sql
-- Get pending tasks for user
SELECT * FROM human_tasks
WHERE status = 'pending'
  AND assignees @> '[{"type": "user", "id": "user-uuid"}]'
ORDER BY due_date NULLS LAST;

-- Get overdue tasks
SELECT * FROM human_tasks
WHERE status = 'pending'
  AND due_date < NOW()
ORDER BY due_date;
```

---

### 19. roles (RBAC)

Tenant-specific roles for fine-grained access control.

```sql
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_system BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT roles_tenant_name_unique UNIQUE(tenant_id, name)
);

CREATE INDEX idx_roles_tenant_id ON roles(tenant_id);
CREATE INDEX idx_roles_is_system ON roles(is_system);
```

**System Roles:**
- `admin`: Full access
- `editor`: Create/edit workflows
- `viewer`: Read-only access

---

### 20. permissions (RBAC)

Available permissions in the system.

```sql
CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resource VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT permissions_resource_action_unique UNIQUE(resource, action)
);

CREATE INDEX idx_permissions_resource ON permissions(resource);
```

**Permission Format:** `resource:action`

Examples:
- `workflow:create`
- `workflow:read`
- `workflow:update`
- `workflow:delete`
- `workflow:execute`
- `execution:read`
- `execution:cancel`
- `credential:create`
- `credential:read`

---

### 21. role_permissions (RBAC)

Maps permissions to roles.

```sql
CREATE TABLE role_permissions (
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (role_id, permission_id)
);

CREATE INDEX idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission_id ON role_permissions(permission_id);
```

---

### 22. user_roles (RBAC)

Maps users to roles.

```sql
CREATE TABLE user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    granted_by UUID REFERENCES users(id),
    granted_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id)
);

CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);
```

**Example RBAC Query:**
```sql
-- Check if user has permission
SELECT EXISTS (
    SELECT 1
    FROM user_roles ur
    JOIN role_permissions rp ON ur.role_id = rp.role_id
    JOIN permissions p ON rp.permission_id = p.id
    WHERE ur.user_id = $1
      AND p.resource = $2
      AND p.action = $3
) AS has_permission;
```

---

### 23. ai_usage_log

Tracks AI/LLM API usage for billing and analytics.

```sql
CREATE TABLE ai_usage_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    credential_id UUID,
    provider VARCHAR(50) NOT NULL,
    model VARCHAR(100) NOT NULL,
    action_type VARCHAR(100) NOT NULL,
    execution_id UUID,
    workflow_id UUID,
    prompt_tokens INTEGER NOT NULL DEFAULT 0,
    completion_tokens INTEGER NOT NULL DEFAULT 0,
    total_tokens INTEGER NOT NULL DEFAULT 0,
    estimated_cost_cents INTEGER NOT NULL DEFAULT 0,
    success BOOLEAN NOT NULL DEFAULT true,
    error_code VARCHAR(100),
    error_message TEXT,
    latency_ms INTEGER,
    request_metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ai_usage_log_tenant ON ai_usage_log(tenant_id);
CREATE INDEX idx_ai_usage_log_tenant_created ON ai_usage_log(tenant_id, created_at DESC);
CREATE INDEX idx_ai_usage_log_provider_model ON ai_usage_log(provider, model);
CREATE INDEX idx_ai_usage_log_execution ON ai_usage_log(execution_id) WHERE execution_id IS NOT NULL;
```

**Columns:**
- `id`: Log entry identifier
- `tenant_id`: Owner tenant
- `credential_id`: Credential used (if any)
- `provider`: AI provider (`openai`, `anthropic`, `bedrock`)
- `model`: Model name (e.g., `gpt-4o`, `claude-3-5-sonnet`)
- `action_type`: Action type (`chat_completion`, `embedding`, `entity_extraction`)
- `execution_id`, `workflow_id`: Context
- `prompt_tokens`, `completion_tokens`, `total_tokens`: Token usage
- `estimated_cost_cents`: Cost in USD cents
- `success`: Whether request succeeded
- `error_code`, `error_message`: Error information
- `latency_ms`: Request latency

**Example Queries:**
```sql
-- Monthly AI costs by tenant
SELECT
    tenant_id,
    DATE_TRUNC('month', created_at) as month,
    SUM(estimated_cost_cents)::FLOAT / 100 as total_cost_usd,
    SUM(total_tokens) as total_tokens
FROM ai_usage_log
GROUP BY tenant_id, month
ORDER BY month DESC;

-- Most expensive models
SELECT
    provider,
    model,
    COUNT(*) as requests,
    SUM(estimated_cost_cents)::FLOAT / 100 as total_cost
FROM ai_usage_log
WHERE created_at >= NOW() - INTERVAL '30 days'
GROUP BY provider, model
ORDER BY total_cost DESC;
```

---

### 24. ai_model_pricing

Pricing information for AI models.

```sql
CREATE TABLE ai_model_pricing (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider VARCHAR(50) NOT NULL,
    model VARCHAR(100) NOT NULL,
    input_cost_per_million INTEGER NOT NULL,
    output_cost_per_million INTEGER NOT NULL,
    context_window INTEGER NOT NULL,
    max_output_tokens INTEGER,
    supports_vision BOOLEAN NOT NULL DEFAULT false,
    supports_function_calling BOOLEAN NOT NULL DEFAULT false,
    supports_json_mode BOOLEAN NOT NULL DEFAULT false,
    is_active BOOLEAN NOT NULL DEFAULT true,
    deprecated_at TIMESTAMPTZ,
    effective_from TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    effective_until TIMESTAMPTZ,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_model_pricing UNIQUE (provider, model, effective_from)
);

CREATE INDEX idx_ai_model_pricing_provider_model ON ai_model_pricing(provider, model);
CREATE INDEX idx_ai_model_pricing_active ON ai_model_pricing(is_active) WHERE is_active = true;
```

**Cost Calculation:**
```sql
-- Cost in USD cents = (tokens / 1,000,000) * cost_per_million
CREATE OR REPLACE FUNCTION estimate_ai_cost(
    p_provider VARCHAR,
    p_model VARCHAR,
    p_prompt_tokens INTEGER,
    p_completion_tokens INTEGER
)
RETURNS INTEGER AS $$
DECLARE
    v_input_cost INTEGER;
    v_output_cost INTEGER;
BEGIN
    SELECT input_cost_per_million, output_cost_per_million
    INTO v_input_cost, v_output_cost
    FROM ai_model_pricing
    WHERE provider = p_provider
      AND model = p_model
      AND is_active = true
    ORDER BY effective_from DESC
    LIMIT 1;

    RETURN (p_prompt_tokens * v_input_cost + p_completion_tokens * v_output_cost) / 1000000;
END;
$$ LANGUAGE plpgsql;
```

---

### 25. notifications

In-app notifications for users.

```sql
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    user_id TEXT NOT NULL,
    title TEXT NOT NULL,
    message TEXT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('info', 'warning', 'error', 'success')),
    link TEXT,
    metadata JSONB DEFAULT '{}'::jsonb,
    is_read BOOLEAN NOT NULL DEFAULT false,
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

CREATE INDEX idx_notifications_tenant_user ON notifications(tenant_id, user_id);
CREATE INDEX idx_notifications_user_unread ON notifications(user_id, is_read) WHERE is_read = false;
CREATE INDEX idx_notifications_created_at ON notifications(created_at DESC);
```

**RLS Policies:**
```sql
-- Tenant isolation
CREATE POLICY tenant_isolation_notifications ON notifications
    FOR ALL
    USING (tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID);

-- User isolation
CREATE POLICY user_isolation_notifications ON notifications
    FOR ALL
    USING (user_id = current_setting('app.current_user_id', TRUE));
```

---

### 26. audit_logs

Global audit log for security and compliance.

```sql
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    user_id UUID,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100) NOT NULL,
    resource_id UUID,
    metadata JSONB NOT NULL DEFAULT '{}',
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_tenant_id ON audit_logs(tenant_id);
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
```

**Actions:**
- `workflow.created`, `workflow.updated`, `workflow.deleted`
- `execution.started`, `execution.completed`, `execution.failed`
- `credential.created`, `credential.accessed`, `credential.deleted`
- `user.login`, `user.logout`, `user.invited`

---

### 27. retention_cleanup_logs

Audit log for retention policy cleanup operations.

```sql
CREATE TABLE retention_cleanup_logs (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    executions_deleted INTEGER NOT NULL DEFAULT 0,
    step_executions_deleted INTEGER NOT NULL DEFAULT 0,
    retention_days INTEGER NOT NULL,
    cutoff_date TIMESTAMPTZ NOT NULL,
    duration_ms INTEGER NOT NULL,
    status VARCHAR(50) NOT NULL,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_retention_cleanup_logs_tenant_id ON retention_cleanup_logs(tenant_id);
CREATE INDEX idx_retention_cleanup_logs_created_at ON retention_cleanup_logs(created_at DESC);
```

---

## Entity Relationship Diagrams

### Workflow Domain

```
┌──────────────┐
│   tenants    │
└──────┬───────┘
       │
       │ 1:N
       │
┌──────▼───────┐          ┌──────────────────┐
│  workflows   │◄─────────│ workflow_versions│
└──────┬───────┘   1:N    └──────────────────┘
       │
       │ 1:N
       │
┌──────▼───────┐          ┌──────────────────┐
│  executions  │◄─────────│ step_executions  │
│              │   1:N    │                  │
│ parent_id ───┼──┐       └──────────────────┘
└──────────────┘  │
       ▲          │
       │          │
       └──────────┘ self-referencing (sub-workflows)
```

### Credential Domain

```
┌──────────────┐
│   tenants    │
└──────┬───────┘
       │
       │ 1:N
       │
┌──────▼───────────────┐
│    credentials       │
│  (encrypted_dek,     │
│   ciphertext, nonce, │
│   auth_tag)          │
└──────┬───────────────┘
       │
       ├──────────────────────┐
       │                      │
       │ 1:N                  │ 1:N
       │                      │
┌──────▼──────────────┐  ┌───▼──────────────────┐
│ credential_rotations│  │ credential_access_log│
└─────────────────────┘  └──────────────────────┘
```

### Webhook Domain

```
┌──────────────┐          ┌──────────────┐
│   tenants    │          │  workflows   │
└──────┬───────┘          └──────┬───────┘
       │                         │
       │ 1:N                     │ 1:N
       │                         │
┌──────▼─────────────────────────▼──┐
│           webhooks                │
└──────┬────────────────────────────┘
       │
       │ 1:N
       │
┌──────▼───────────┐          ┌──────────────────┐
│  webhook_events  │          │ webhook_filters  │
│                  │          │ (JSONPath rules) │
│  source_event◄───┼──┐       └──────────────────┘
└──────────────────┘  │                ▲
       ▲              │                │
       │              │                │ 1:N
       │              └────────────────┘
       └───────────────┘ self-referencing (replays)
```

### Marketplace Domain

```
┌──────────────────────┐
│marketplace_templates │
└──────┬───────────────┘
       │
       ├─────────────────────┬──────────────────┐
       │ 1:N                 │ 1:N              │ 1:N
       │                     │                  │
┌──────▼─────────────┐  ┌───▼───────────┐  ┌──▼──────────────────┐
│marketplace_        │  │marketplace_   │  │marketplace_template_│
│installations       │  │reviews        │  │versions             │
└────────────────────┘  └───────────────┘  └─────────────────────┘
```

### RBAC Domain

```
┌──────────────┐          ┌──────────────┐
│   tenants    │          │   users      │
└──────┬───────┘          └──────┬───────┘
       │                         │
       │ 1:N                     │ N:M
       │                         │
┌──────▼───────┐          ┌──────▼──────┐
│    roles     │◄─────────│ user_roles  │
└──────┬───────┘   N:M    └─────────────┘
       │
       │ N:M
       │
┌──────▼──────────┐       ┌──────────────┐
│role_permissions │──────►│ permissions  │
└─────────────────┘  N:M  └──────────────┘
```

### Analytics Domain (In-Memory)

Analytics data is computed on-the-fly from `executions` and `step_executions` tables. No separate analytics tables exist - all metrics are aggregated queries.

---

### Collaboration Domain (In-Memory)

Collaboration sessions are stored in-memory using WebSocket connections. No database tables for real-time collaboration state.

**In-Memory Structure:**
```
Hub (singleton)
  └── Sessions (map[workflowID]*EditSession)
        └── EditSession
              ├── Users (map[userID]*UserPresence)
              └── Locks (map[elementID]*EditLock)
```

---

## Indexes and Performance

### Index Strategy

1. **Primary Keys**: All tables use UUID primary keys with default index
2. **Foreign Keys**: Indexed for JOIN performance
3. **Tenant Isolation**: All tenant-scoped queries include `tenant_id` in composite indexes
4. **Time-Based**: DESC indexes on timestamp columns for recent-first queries
5. **JSONB**: GIN indexes for JSONB containment queries (`@>`, `@?`)
6. **Partial Indexes**: Filter index for specific conditions (e.g., `enabled = true`)

### Composite Index Patterns

```sql
-- Cursor-based pagination
CREATE INDEX idx_executions_cursor_pagination
    ON executions(tenant_id, created_at DESC, id);

-- Filtering + sorting
CREATE INDEX idx_executions_tenant_status
    ON executions(tenant_id, status, created_at DESC);

-- Multi-column filters
CREATE INDEX idx_webhook_events_tenant_status_created
    ON webhook_events(tenant_id, status, created_at DESC);
```

### Partial Indexes

Used when most queries filter on a specific value:

```sql
-- Only index enabled schedules
CREATE INDEX idx_schedules_next_run_at ON schedules(next_run_at)
WHERE enabled = true;

-- Only index non-expired credentials
CREATE INDEX idx_credentials_expires_at ON credentials(expires_at)
WHERE expires_at IS NOT NULL;

-- Only index unread notifications
CREATE INDEX idx_notifications_user_unread ON notifications(user_id, is_read)
WHERE is_read = false;
```

### GIN Indexes

For JSONB and array columns:

```sql
-- Array containment
CREATE INDEX idx_workflow_templates_tags
    ON workflow_templates USING GIN(tags);

-- JSONB containment
CREATE INDEX idx_human_tasks_assignee
    ON human_tasks USING GIN(assignees);
```

### Index Maintenance

```sql
-- Check index usage
SELECT
    schemaname, tablename, indexname,
    idx_scan, idx_tup_read, idx_tup_fetch
FROM pg_stat_user_indexes
WHERE schemaname = 'public'
ORDER BY idx_scan;

-- Find unused indexes
SELECT
    schemaname, tablename, indexname
FROM pg_stat_user_indexes
WHERE idx_scan = 0
  AND indexname NOT LIKE '%pkey';

-- Rebuild bloated indexes
REINDEX INDEX CONCURRENTLY idx_executions_created_at;
```

---

## Query Optimization Examples

### Avoiding N+1 Queries

**Bad:**
```go
// N+1: Queries in loop
workflows := getWorkflows()
for _, w := range workflows {
    executions := getExecutionsForWorkflow(w.ID) // N queries
}
```

**Good:**
```go
// Single query with JOIN
SELECT
    w.id, w.name,
    e.id as execution_id, e.status
FROM workflows w
LEFT JOIN executions e ON w.id = e.workflow_id
WHERE e.created_at >= NOW() - INTERVAL '7 days'
```

### Efficient Pagination

**Bad (OFFSET):**
```sql
-- Slow for large offsets
SELECT * FROM executions
ORDER BY created_at DESC
LIMIT 50 OFFSET 10000;
```

**Good (Cursor-based):**
```sql
-- Fast using index
SELECT * FROM executions
WHERE created_at < $1  -- cursor from previous page
ORDER BY created_at DESC
LIMIT 50;
```

### JSONB Query Optimization

```sql
-- Efficient JSONB containment query
EXPLAIN ANALYZE
SELECT * FROM workflows
WHERE definition @> '{"nodes": [{"type": "webhook_trigger"}]}'::jsonb;

-- Uses GIN index if available
-- Without index: Seq Scan (slow)
-- With GIN index: Bitmap Index Scan (fast)
```

### Aggregate Queries

```sql
-- Efficient aggregation with indexes
SELECT
    workflow_id,
    status,
    COUNT(*) as count,
    AVG(EXTRACT(EPOCH FROM (completed_at - started_at))) * 1000 as avg_duration_ms
FROM executions
WHERE created_at >= NOW() - INTERVAL '30 days'
GROUP BY workflow_id, status;

-- Uses: idx_executions_tenant_workflow
```

---

## Data Types and Conventions

### UUID vs BIGSERIAL

Gorax uses **UUIDs** for all primary keys:

**Advantages:**
- Globally unique (no collisions across tenants/shards)
- Can be generated client-side
- No auto-increment race conditions
- Better for distributed systems

**Disadvantages:**
- Larger storage (16 bytes vs 8 bytes)
- Slightly slower index lookups

### Timestamp Handling

All timestamps use `TIMESTAMPTZ` (timezone-aware):

```sql
-- Always stored in UTC
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()

-- Application converts to user's timezone for display
```

**Go Code:**
```go
// Always parse as UTC
time.Parse(time.RFC3339, "2024-01-15T10:30:00Z")

// Convert to user timezone for display
userTime := utcTime.In(userLocation)
```

### JSONB Storage

Used for flexible, schema-less data:

```sql
-- Workflow definitions
definition JSONB NOT NULL

-- Metadata
metadata JSONB NOT NULL DEFAULT '{}'

-- Settings
settings JSONB NOT NULL DEFAULT '{}'
```

**Querying JSONB:**
```sql
-- Exact match
WHERE settings->>'retention_days' = '90'

-- Numeric comparison
WHERE (settings->>'retention_days')::INTEGER > 30

-- Nested access
WHERE definition->'nodes'->0->>'type' = 'webhook_trigger'

-- Containment
WHERE definition @> '{"nodes": [{"type": "webhook_trigger"}]}'

-- Existence
WHERE definition ? 'nodes'
```

### Enum Types

Gorax uses **VARCHAR with CHECK constraints** instead of PostgreSQL ENUMs:

**Reason:** ENUMs are difficult to modify (require ALTER TYPE)

```sql
-- Flexible approach
status VARCHAR(50) NOT NULL DEFAULT 'active'

-- With constraint for validation
CONSTRAINT chk_status CHECK (status IN ('active', 'inactive', 'suspended'))
```

### TEXT vs VARCHAR

**Rule of thumb:**
- Use `VARCHAR(N)` when there's a meaningful length limit
- Use `TEXT` when length is unbounded

```sql
-- Limited length
name VARCHAR(255) NOT NULL
email VARCHAR(255) NOT NULL

-- Unbounded
description TEXT
error_message TEXT
```

---

## Partitioning Strategy

### When to Partition

Partition tables that:
1. Grow very large (>100M rows)
2. Have time-based access patterns
3. Need efficient data archival

**Candidate Tables:**
- `executions` (partition by month)
- `webhook_events` (partition by month)
- `audit_logs` (partition by month)
- `ai_usage_log` (partition by month)

### Partitioning Implementation

```sql
-- Convert executions to partitioned table
ALTER TABLE executions RENAME TO executions_old;

CREATE TABLE executions (
    id UUID,
    tenant_id UUID NOT NULL,
    workflow_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    -- ... other columns
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

-- Create monthly partitions
CREATE TABLE executions_2024_01 PARTITION OF executions
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

CREATE TABLE executions_2024_02 PARTITION OF executions
    FOR VALUES FROM ('2024-02-01') TO ('2024-03-01');

-- Migrate data
INSERT INTO executions SELECT * FROM executions_old;

-- Drop old table
DROP TABLE executions_old CASCADE;
```

### Automatic Partition Creation

```sql
-- Function to create next month's partition
CREATE OR REPLACE FUNCTION create_next_month_partition()
RETURNS void AS $$
DECLARE
    next_month DATE := DATE_TRUNC('month', NOW() + INTERVAL '1 month');
    partition_name TEXT := 'executions_' || TO_CHAR(next_month, 'YYYY_MM');
    start_date DATE := next_month;
    end_date DATE := next_month + INTERVAL '1 month';
BEGIN
    EXECUTE format(
        'CREATE TABLE IF NOT EXISTS %I PARTITION OF executions FOR VALUES FROM (%L) TO (%L)',
        partition_name, start_date, end_date
    );
END;
$$ LANGUAGE plpgsql;

-- Schedule via cron or pg_cron
SELECT cron.schedule('create-partitions', '0 0 1 * *', 'SELECT create_next_month_partition()');
```

### Archival Strategy

```sql
-- Detach old partition
ALTER TABLE executions DETACH PARTITION executions_2023_01;

-- Archive to S3 or move to cold storage
-- Then drop the table
DROP TABLE executions_2023_01;
```

---

## Backup and Recovery

### Backup Strategy

**1. Continuous Archiving (PITR)**

```bash
# Enable WAL archiving in postgresql.conf
wal_level = replica
archive_mode = on
archive_command = 'aws s3 cp %p s3://gorax-wal-archive/%f'
```

**2. Daily Base Backups**

```bash
# Full backup daily
pg_basebackup -D /backup/base-$(date +%Y%m%d) -Ft -z -P

# Or use pgBackRest
pgbackrest --stanza=gorax --type=full backup
```

**3. Logical Dumps**

```bash
# Per-tenant backup
pg_dump -h localhost -U gorax \
    --no-owner --no-acl \
    --table='executions' \
    --table='workflows' \
    --table='credentials' \
    > tenant-backup.sql

# Schema-only backup
pg_dump -h localhost -U gorax --schema-only > schema.sql
```

### Point-in-Time Recovery (PITR)

```bash
# 1. Restore base backup
tar -xzf base-backup.tar.gz -C /var/lib/postgresql/14/main

# 2. Create recovery.conf
cat > /var/lib/postgresql/14/main/recovery.conf <<EOF
restore_command = 'aws s3 cp s3://gorax-wal-archive/%f %p'
recovery_target_time = '2024-01-15 14:30:00 UTC'
EOF

# 3. Start PostgreSQL
pg_ctl start
```

### Disaster Recovery

**Recovery Time Objective (RTO):** < 1 hour
**Recovery Point Objective (RPO):** < 5 minutes

**DR Runbook:**

1. **Detect outage** (monitoring alerts)
2. **Assess scope** (database corruption, hardware failure, region outage)
3. **Failover to replica** (if available)
4. **Restore from backup** (if no replica)
5. **Verify data integrity**
6. **Resume operations**

---

## Common Queries

### Tenant Statistics

```sql
-- Tenant overview
SELECT
    t.id,
    t.name,
    t.tier,
    COUNT(DISTINCT w.id) as workflow_count,
    COUNT(DISTINCT u.id) as user_count,
    COUNT(e.id) FILTER (WHERE e.created_at >= NOW() - INTERVAL '30 days') as executions_30d
FROM tenants t
LEFT JOIN workflows w ON t.id = w.tenant_id
LEFT JOIN users u ON t.id = u.tenant_id
LEFT JOIN executions e ON t.id = e.tenant_id
WHERE t.status = 'active'
GROUP BY t.id, t.name, t.tier;
```

### Workflow Performance

```sql
-- Top 10 slowest workflows
SELECT
    w.id,
    w.name,
    COUNT(*) as execution_count,
    AVG(EXTRACT(EPOCH FROM (e.completed_at - e.started_at))) * 1000 as avg_duration_ms,
    MAX(EXTRACT(EPOCH FROM (e.completed_at - e.started_at))) * 1000 as max_duration_ms
FROM workflows w
JOIN executions e ON w.id = e.workflow_id
WHERE e.status = 'completed'
  AND e.created_at >= NOW() - INTERVAL '7 days'
GROUP BY w.id, w.name
HAVING COUNT(*) >= 10
ORDER BY avg_duration_ms DESC
LIMIT 10;
```

### Error Analysis

```sql
-- Most common errors
SELECT
    w.name as workflow_name,
    e.error_message,
    COUNT(*) as error_count,
    MAX(e.created_at) as last_occurrence
FROM executions e
JOIN workflows w ON e.workflow_id = w.id
WHERE e.status = 'failed'
  AND e.created_at >= NOW() - INTERVAL '7 days'
GROUP BY w.name, e.error_message
ORDER BY error_count DESC
LIMIT 20;
```

### Credential Audit

```sql
-- Credentials not used in 90 days
SELECT
    c.id,
    c.name,
    c.type,
    c.created_at,
    c.last_used_at,
    EXTRACT(DAY FROM NOW() - COALESCE(c.last_used_at, c.created_at)) as days_unused
FROM credentials c
WHERE c.status = 'active'
  AND (c.last_used_at < NOW() - INTERVAL '90 days' OR c.last_used_at IS NULL)
ORDER BY days_unused DESC;
```

### Webhook Health

```sql
-- Webhook failure rate (last 24 hours)
SELECT
    w.id,
    w.name,
    COUNT(*) as total_events,
    COUNT(*) FILTER (WHERE we.status = 'processed') as successful,
    COUNT(*) FILTER (WHERE we.status = 'failed') as failed,
    (COUNT(*) FILTER (WHERE we.status = 'failed')::FLOAT / COUNT(*) * 100)::DECIMAL(5,2) as failure_rate_pct
FROM webhooks w
JOIN webhook_events we ON w.id = we.webhook_id
WHERE we.created_at >= NOW() - INTERVAL '24 hours'
GROUP BY w.id, w.name
HAVING COUNT(*) >= 10
ORDER BY failure_rate_pct DESC;
```

### Active Users

```sql
-- Active users by tenant (last 30 days)
SELECT
    t.name as tenant_name,
    COUNT(DISTINCT al.user_id) as active_users
FROM tenants t
JOIN audit_logs al ON t.id = al.tenant_id
WHERE al.created_at >= NOW() - INTERVAL '30 days'
GROUP BY t.name
ORDER BY active_users DESC;
```

---

## Migration Patterns

### Migration File Structure

```
migrations/
  001_initial_schema.sql          - Core tables
  002_webhook_events.sql          - Webhook system
  003_schedules.sql               - Scheduled triggers
  004_execution_history_enhancements.sql
  005_credential_vault.sql        - Encryption upgrade
  ...
  020_marketplace.sql             - Latest
```

### Migration Naming Convention

```
{number}_{description}.sql

- Number: 3-digit zero-padded (001, 002, ...)
- Description: snake_case, descriptive
```

### Up Migration Pattern

```sql
-- Add column with default
ALTER TABLE workflows
ADD COLUMN IF NOT EXISTS tags TEXT[] DEFAULT '{}';

-- Create index concurrently (no table lock)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_workflows_tags
ON workflows USING GIN(tags);

-- Add constraint
ALTER TABLE executions
ADD CONSTRAINT chk_status
CHECK (status IN ('pending', 'running', 'completed', 'failed', 'cancelled'));
```

### Down Migration Pattern

```sql
-- migrations/020_marketplace_rollback.sql

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS marketplace_reviews CASCADE;
DROP TABLE IF EXISTS marketplace_installations CASCADE;
DROP TABLE IF EXISTS marketplace_template_versions CASCADE;
DROP TABLE IF EXISTS marketplace_templates CASCADE;
```

### Data Migration Best Practices

```sql
-- Add column with nullable
ALTER TABLE workflows ADD COLUMN new_field TEXT;

-- Backfill data in batches (avoid table lock)
DO $$
DECLARE
    batch_size INT := 1000;
    processed INT := 0;
BEGIN
    LOOP
        WITH batch AS (
            SELECT id FROM workflows
            WHERE new_field IS NULL
            LIMIT batch_size
            FOR UPDATE SKIP LOCKED
        )
        UPDATE workflows w
        SET new_field = 'default_value'
        FROM batch
        WHERE w.id = batch.id;

        GET DIAGNOSTICS processed = ROW_COUNT;
        EXIT WHEN processed = 0;

        COMMIT;  -- Commit each batch
    END LOOP;
END $$;

-- Make column NOT NULL after backfill
ALTER TABLE workflows ALTER COLUMN new_field SET NOT NULL;
```

### Zero-Downtime Migrations

**Pattern 1: Add Column**
```sql
-- Step 1: Add nullable column
ALTER TABLE workflows ADD COLUMN new_field TEXT;

-- Step 2: Deploy code that writes to both old and new field
-- (Application deployment)

-- Step 3: Backfill old data
UPDATE workflows SET new_field = old_field WHERE new_field IS NULL;

-- Step 4: Make column NOT NULL
ALTER TABLE workflows ALTER COLUMN new_field SET NOT NULL;

-- Step 5: Deploy code that reads from new field
-- (Application deployment)

-- Step 6: Drop old column
ALTER TABLE workflows DROP COLUMN old_field;
```

**Pattern 2: Rename Column**
```sql
-- Use a view to map old name to new name during transition
CREATE VIEW workflows_compat AS
SELECT
    id,
    name,
    new_name AS old_name,  -- Alias new column as old name
    created_at
FROM workflows;

-- Update application code gradually
-- Drop view when migration complete
```

---

## Performance Tuning

### Connection Pooling

```go
// Recommended pool settings
db.SetMaxOpenConns(25)          // Max connections
db.SetMaxIdleConns(5)           // Idle connections
db.SetConnMaxLifetime(5 * time.Minute)  // Connection lifetime
db.SetConnMaxIdleTime(10 * time.Minute) // Idle timeout
```

### Query Timeout

```sql
-- Set statement timeout (prevent long-running queries)
SET statement_timeout = '30s';

-- Per-query timeout
SELECT * FROM executions
WHERE tenant_id = $1
LIMIT 1000
OPTION (statement_timeout = '10s');
```

### EXPLAIN ANALYZE

```sql
-- Analyze query performance
EXPLAIN (ANALYZE, BUFFERS, VERBOSE)
SELECT * FROM executions
WHERE tenant_id = $1
  AND status = 'running'
ORDER BY created_at DESC
LIMIT 50;

-- Look for:
-- - Seq Scan (bad) vs Index Scan (good)
-- - Actual time vs Estimated time
-- - Rows returned vs Rows estimated
```

### Table Bloat

```sql
-- Check table bloat
SELECT
    schemaname, tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename) - pg_relation_size(schemaname||'.'||tablename)) AS external_size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

-- Vacuum to reclaim space
VACUUM ANALYZE executions;

-- Full vacuum (requires table lock)
VACUUM FULL executions;
```

---

## Security Considerations

### Row-Level Security (RLS)

All tenant-scoped tables have RLS enabled:

```sql
ALTER TABLE workflows ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_workflows ON workflows
    USING (tenant_id = current_setting('app.current_tenant_id', true)::UUID);
```

**Best Practices:**
- ✅ Always set tenant context before queries
- ✅ Use RLS for defense-in-depth
- ✅ Test RLS policies thoroughly
- ❌ Don't rely solely on application-level filtering

### Credential Encryption

**Envelope Encryption:**
1. Generate random DEK (Data Encryption Key)
2. Encrypt credential with DEK using AES-256-GCM
3. Encrypt DEK with AWS KMS master key
4. Store: `encrypted_dek`, `ciphertext`, `nonce`, `auth_tag`

**Why AES-256-GCM?**
- Authenticated encryption (prevents tampering)
- Fast and secure
- Built-in integrity check (auth_tag)

### Parameterized Queries

**Always use parameterized queries** to prevent SQL injection:

```go
// ✅ Good: Parameterized
db.Query("SELECT * FROM workflows WHERE id = $1", workflowID)

// ❌ Bad: String concatenation
db.Query(fmt.Sprintf("SELECT * FROM workflows WHERE id = '%s'", workflowID))
```

### Least Privilege

```sql
-- Application user (limited permissions)
CREATE USER gorax_app WITH PASSWORD 'secure_password';

GRANT CONNECT ON DATABASE gorax TO gorax_app;
GRANT USAGE ON SCHEMA public TO gorax_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO gorax_app;
GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO gorax_app;

-- Read-only user (analytics)
CREATE USER gorax_readonly WITH PASSWORD 'secure_password';
GRANT CONNECT ON DATABASE gorax TO gorax_readonly;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO gorax_readonly;
```

---

## Maintenance Tasks

### Daily
- Monitor query performance (slow query log)
- Check replication lag (if using replicas)
- Verify backups completed

### Weekly
- `VACUUM ANALYZE` large tables
- Review table/index bloat
- Audit failed queries

### Monthly
- Review and update statistics
- Analyze index usage (drop unused indexes)
- Review retention policy logs
- Update AI model pricing

### Quarterly
- Full database performance audit
- Review and optimize slow queries
- Plan partitioning for growing tables
- Review disaster recovery procedures

---

## References

- PostgreSQL Documentation: https://www.postgresql.org/docs/
- pgBackRest: https://pgbackrest.org/
- Row-Level Security: https://www.postgresql.org/docs/current/ddl-rowsecurity.html
- JSONB Performance: https://www.postgresql.org/docs/current/datatype-json.html

---

**Document Version:** 1.0
**Last Updated:** 2026-01-01
**Maintained By:** Gorax Engineering Team
