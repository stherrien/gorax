# Webhook & Events System Requirements

## Current State Summary

### Already Implemented ✓
- **Webhook Receiver**: HMAC-SHA256 signature validation, multiple auth types (none, signature, basic, API key)
- **Retry Logic**: Exponential backoff with jitter (1s initial, 30s max, 2.0 multiplier)
- **Event Queue**: AWS SQS integration with batch support, long polling
- **Dead Letter Queue**: Auto-routing for failed messages, metrics tracking
- **Context Injection**: Trigger data (headers, query, body) passed to workflows
- **Execution History**: Stored in database with full trigger_data JSONB
- **WebSocket Events**: Real-time execution/step broadcasting
- **Circuit Breaker**: Failure threshold with auto-recovery

### Missing Features ✗

---

## 1. Event Type Registry

**Priority**: High

**Description**: Centralized registry of supported event types with schema definitions.

**Requirements**:
- [ ] Create `event_types` database table (id, name, schema, version, description)
- [ ] Define core event types:
  - `webhook.received`
  - `workflow.triggered`
  - `execution.started`
  - `execution.completed`
  - `execution.failed`
  - `step.started`
  - `step.completed`
  - `step.failed`
- [ ] JSON Schema validation for event payloads
- [ ] Event type versioning for backward compatibility
- [ ] API endpoints: GET /api/v1/event-types

---

## 2. Webhook Management UI

**Priority**: High

**Description**: Dashboard for managing webhook configurations.

**Requirements**:
- [ ] Webhook list page with table view
  - Columns: Name, Path, Workflow, Auth Type, Status, Created, Last Triggered
  - Actions: Enable/Disable, Edit, Delete, Copy URL
- [ ] Webhook detail/edit page
  - Path configuration
  - Auth type selector (None, Signature, Basic, API Key)
  - Secret display with copy button
  - Secret regeneration with confirmation
  - Associated workflow selector
- [ ] Webhook URL generation and display
  - Format: `{base_url}/api/v1/webhooks/{path}`
  - Copy-to-clipboard functionality
- [ ] Status indicators (active, disabled, error rate)

---

## 3. Webhook Testing Interface

**Priority**: High

**Description**: Built-in interface to test webhooks with sample payloads.

**Requirements**:
- [ ] Test payload editor (JSON)
- [ ] HTTP method selector
- [ ] Custom headers input
- [ ] Send test request button
- [ ] Response viewer (status, headers, body)
- [ ] Sample payload templates per event type
- [ ] Request/response history (last 10 tests)
- [ ] Backend endpoint: POST /api/v1/webhooks/{id}/test

---

## 4. Event Filtering Rules

**Priority**: Medium

**Description**: Pre-trigger filtering to conditionally execute workflows.

**Requirements**:
- [ ] Filter rule model (webhook_id, field, operator, value, enabled)
- [ ] Supported operators:
  - `equals`, `not_equals`
  - `contains`, `not_contains`
  - `starts_with`, `ends_with`
  - `regex_match`
  - `greater_than`, `less_than`
  - `in`, `not_in`
- [ ] JSON path support for nested fields (e.g., `$.data.status`)
- [ ] Multiple conditions with AND/OR logic
- [ ] Visual rule builder UI
- [ ] Rule testing with sample payloads
- [ ] Filter evaluation before workflow trigger

---

## 5. Event History Viewer

**Priority**: Medium

**Description**: UI for viewing webhook delivery history and payloads.

**Requirements**:
- [ ] Event list with filtering:
  - Date range
  - Webhook/workflow
  - Status (success, failed, filtered)
  - Search by payload content
- [ ] Event detail view:
  - Request headers
  - Request body (formatted JSON)
  - Response status
  - Processing time
  - Triggered execution (link)
- [ ] Pagination and export (CSV/JSON)
- [ ] Retention policy configuration

---

## 6. Event Replay Capability

**Priority**: Medium

**Description**: Re-trigger workflows from stored webhook payloads.

**Requirements**:
- [ ] Replay button on event history items
- [ ] Batch replay selection
- [ ] Replay with modified payload option
- [ ] Replay audit trail
- [ ] API endpoint: POST /api/v1/events/{id}/replay
- [ ] Prevent infinite replay loops (max replay count)

---

## 7. Priority Queue Handling

**Priority**: Low

**Description**: Prioritized event processing for critical webhooks.

**Requirements**:
- [ ] Priority levels: Low (0), Normal (1), High (2), Critical (3)
- [ ] Priority configuration per webhook
- [ ] Separate high-priority queue or priority attribute in SQS
- [ ] Priority-based worker allocation
- [ ] SLA monitoring for high-priority events

---

## 8. Frontend Webhook API Integration

**Priority**: High

**Description**: API client methods for webhook management.

**Requirements**:
- [ ] Add to `web/src/api/webhooks.ts`:
  ```typescript
  interface Webhook {
    id: string
    tenantId: string
    workflowId: string
    nodeId: string
    path: string
    authType: 'none' | 'signature' | 'basic' | 'api_key'
    enabled: boolean
    createdAt: string
    updatedAt: string
    lastTriggeredAt?: string
  }

  webhookAPI.list(params): Promise<WebhookListResponse>
  webhookAPI.get(id): Promise<Webhook>
  webhookAPI.create(input): Promise<Webhook>
  webhookAPI.update(id, input): Promise<Webhook>
  webhookAPI.delete(id): Promise<void>
  webhookAPI.regenerateSecret(id): Promise<{secret: string}>
  webhookAPI.test(id, payload): Promise<TestResponse>
  ```

---

## 9. Webhook Delivery Log Table

**Priority**: Medium

**Description**: Persistent audit trail for webhook deliveries.

**Requirements**:
- [ ] Create `webhook_events` table:
  ```sql
  CREATE TABLE webhook_events (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    webhook_id UUID NOT NULL,
    execution_id UUID,
    request_method VARCHAR(10),
    request_headers JSONB,
    request_body JSONB,
    response_status INTEGER,
    processing_time_ms INTEGER,
    status VARCHAR(20), -- received, processed, filtered, failed
    error_message TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
  );
  ```
- [ ] Indexes on tenant_id, webhook_id, created_at, status
- [ ] Retention policy (default 30 days)
- [ ] Automatic cleanup job

---

## 10. Event Metadata Enrichment

**Priority**: Low

**Description**: Automatic enrichment of webhook events with additional context.

**Requirements**:
- [ ] Capture metadata:
  - Source IP address
  - User agent parsing
  - Geolocation (optional)
  - Request timing
- [ ] Custom metadata fields per webhook
- [ ] Metadata available in workflow context
- [ ] Configurable enrichment rules

---

## Implementation Priority Order

1. **Phase 1 (Week 1-2)**: High Priority
   - Webhook API endpoints (#8)
   - Webhook Management UI (#2)
   - Webhook Testing Interface (#3)

2. **Phase 2 (Week 2-3)**: Medium Priority
   - Event Type Registry (#1)
   - Event History Viewer (#5)
   - Webhook Delivery Log (#9)

3. **Phase 3 (Week 3-4)**: Medium Priority
   - Event Filtering Rules (#4)
   - Event Replay (#6)

4. **Phase 4 (Optional)**: Low Priority
   - Priority Queue Handling (#7)
   - Event Metadata Enrichment (#10)

---

## Database Migrations Required

```sql
-- Migration: Add webhook_events table
CREATE TABLE webhook_events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id),
  webhook_id UUID NOT NULL REFERENCES webhooks(id),
  execution_id UUID REFERENCES executions(id),
  request_method VARCHAR(10) NOT NULL,
  request_headers JSONB,
  request_body JSONB,
  response_status INTEGER,
  processing_time_ms INTEGER,
  status VARCHAR(20) NOT NULL DEFAULT 'received',
  error_message TEXT,
  filtered_reason TEXT,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_webhook_events_tenant ON webhook_events(tenant_id);
CREATE INDEX idx_webhook_events_webhook ON webhook_events(webhook_id);
CREATE INDEX idx_webhook_events_created ON webhook_events(created_at);
CREATE INDEX idx_webhook_events_status ON webhook_events(status);

-- Migration: Add event_types table
CREATE TABLE event_types (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(100) NOT NULL UNIQUE,
  description TEXT,
  schema JSONB NOT NULL,
  version INTEGER DEFAULT 1,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Migration: Add webhook_filters table
CREATE TABLE webhook_filters (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  webhook_id UUID NOT NULL REFERENCES webhooks(id) ON DELETE CASCADE,
  field_path VARCHAR(255) NOT NULL,
  operator VARCHAR(20) NOT NULL,
  value JSONB NOT NULL,
  enabled BOOLEAN DEFAULT true,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_webhook_filters_webhook ON webhook_filters(webhook_id);

-- Migration: Add columns to webhooks table
ALTER TABLE webhooks ADD COLUMN priority INTEGER DEFAULT 1;
ALTER TABLE webhooks ADD COLUMN last_triggered_at TIMESTAMPTZ;
ALTER TABLE webhooks ADD COLUMN trigger_count INTEGER DEFAULT 0;
```
