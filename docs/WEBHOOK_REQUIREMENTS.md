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

### Recently Implemented ✓ (Phase 1-2)

---

## 1. Event Type Registry ✅

**Priority**: High | **Status**: Complete

**Description**: Centralized registry of supported event types with schema definitions.

**Implemented**:
- [x] Create `event_types` database table (id, name, schema, version, description)
- [x] Define core event types:
  - `webhook.received`, `webhook.processed`, `webhook.filtered`, `webhook.failed`
  - `execution.started`, `execution.completed`, `execution.failed`, `execution.cancelled`
  - `step.started`, `step.completed`, `step.failed`, `step.retrying`
- [x] JSON Schema validation for event payloads
- [x] Event type versioning for backward compatibility
- [x] API endpoints: GET /api/v1/event-types
- [x] Repository (`internal/eventtypes/repository.go`)
- [x] Service (`internal/eventtypes/service.go`)
- [x] Handler (`internal/api/handlers/event_types_handler.go`)

---

## 2. Webhook Management UI ✅

**Priority**: High | **Status**: Complete

**Description**: Dashboard for managing webhook configurations.

**Implemented**:
- [x] Webhook list page with table view (`web/src/pages/WebhookList.tsx`)
  - Columns: Name, Path, Workflow, Auth Type, Status, Created, Last Triggered
  - Actions: Enable/Disable, Edit, Delete, Copy URL
- [x] Webhook detail/edit page
  - Path configuration
  - Auth type selector (None, Signature, Basic, API Key)
  - Associated workflow selector
- [x] Webhook URL generation and display
  - Copy-to-clipboard functionality
- [x] Status indicators (active, disabled, error rate)
- [x] API client (`web/src/api/webhooks.ts`)
- [x] React hooks (`web/src/hooks/useWebhooks.ts`)

---

## 3. Webhook Testing Interface ✅

**Priority**: High | **Status**: Complete

**Description**: Built-in interface to test webhooks with sample payloads.

**Implemented**:
- [x] Test payload editor (JSON)
- [x] HTTP method selector
- [x] Custom headers input
- [x] Send test request button
- [x] Response viewer (status, headers, body)
- [x] Backend endpoint: POST /api/v1/webhooks/{id}/test

---

## 4. Event Filtering Rules ✅

**Priority**: Medium | **Status**: Complete

**Description**: Pre-trigger filtering to conditionally execute workflows.

**Implemented**:
- [x] Filter rule model (`webhook_filters` table)
- [x] Supported operators: equals, not_equals, contains, not_contains, regex, gt, gte, lt, lte, in, not_in, exists, not_exists
- [x] JSON path support for nested fields
- [x] Multiple conditions with AND/OR logic (logic_group)
- [x] Filter evaluation (`internal/webhook/filter.go`)
- [x] Filter tests (`internal/webhook/filter_test.go`)

---

## 5. Event History Viewer ✅

**Priority**: Medium | **Status**: Complete

**Description**: UI for viewing webhook delivery history and payloads.

**Implemented**:
- [x] Event list with filtering (`WebhookEventHistory.tsx`)
  - Date range
  - Webhook/workflow
  - Status (success, failed, filtered)
- [x] Event detail view:
  - Request headers
  - Request body (formatted JSON)
  - Response status
  - Processing time
  - Triggered execution (link)
- [x] Pagination and export (CSV)
- [x] CSV export utility (`web/src/utils/csvExport.ts`)

---

## 6. Event Replay Capability ✅

**Priority**: Medium | **Status**: Complete

**Description**: Re-trigger workflows from stored webhook payloads.

**Implemented**:
- [x] Replay button on event history items
- [x] Replay service (`internal/webhook/replay.go`)
- [x] Replay tests (`internal/webhook/replay_test.go`)
- [x] API endpoint: POST /api/v1/events/{id}/replay
- [x] Max replay count protection

---

## 7. Priority Queue Handling

**Priority**: Low | **Status**: Partial

**Description**: Prioritized event processing for critical webhooks.

**Implemented**:
- [x] Priority column added to webhooks table (INTEGER DEFAULT 1)
- [x] Priority index for queries

**Remaining**:
- [ ] Priority levels UI: Low (0), Normal (1), High (2), Critical (3)
- [ ] Separate high-priority queue or priority attribute in SQS
- [ ] Priority-based worker allocation
- [ ] SLA monitoring for high-priority events

---

## 8. Frontend Webhook API Integration ✅

**Priority**: High | **Status**: Complete

**Description**: API client methods for webhook management.

**Implemented**:
- [x] `web/src/api/webhooks.ts` with full CRUD operations
- [x] Webhook interface with all required fields
- [x] list, get, create, update, delete methods
- [x] regenerateSecret method
- [x] test endpoint integration
- [x] Event history API methods

---

## 9. Webhook Delivery Log Table ✅

**Priority**: Medium | **Status**: Complete

**Description**: Persistent audit trail for webhook deliveries.

**Implemented**:
- [x] `webhook_events` table with all fields (migrations/002_webhook_events.sql)
- [x] Indexes on tenant_id, webhook_id, created_at, status
- [x] Composite index for common queries
- [x] Webhook event statistics view
- [x] Webhook health status view

**Remaining**:
- [ ] Retention policy configuration UI
- [ ] Automatic cleanup job (cron)

---

## 10. Event Metadata Enrichment

**Priority**: Low | **Status**: Not Started

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

## Implementation Status Summary

### Completed ✅
- Event Type Registry (Section 1)
- Webhook Management UI (Section 2)
- Webhook Testing Interface (Section 3)
- Event Filtering Rules (Section 4)
- Event History Viewer (Section 5)
- Event Replay Capability (Section 6)
- Frontend Webhook API (Section 8)
- Webhook Delivery Log Table (Section 9)

### Partial/In Progress 🔄
- Priority Queue Handling (Section 7) - database ready, UI/workers pending

### Not Started ❌
- Event Metadata Enrichment (Section 10)
- Retention policy cleanup job

---

## Original Implementation Priority Order

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
