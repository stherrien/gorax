-- Webhook events and filtering system
-- Adds webhook event tracking, event type registry, and filter rules

-- Alter existing webhooks table to add new columns
ALTER TABLE webhooks ADD COLUMN IF NOT EXISTS name VARCHAR(255);
ALTER TABLE webhooks ADD COLUMN IF NOT EXISTS description TEXT;
ALTER TABLE webhooks ADD COLUMN IF NOT EXISTS priority INTEGER NOT NULL DEFAULT 1;
ALTER TABLE webhooks ADD COLUMN IF NOT EXISTS last_triggered_at TIMESTAMPTZ;
ALTER TABLE webhooks ADD COLUMN IF NOT EXISTS trigger_count INTEGER NOT NULL DEFAULT 0;

-- Create index for priority-based queries
CREATE INDEX IF NOT EXISTS idx_webhooks_priority ON webhooks(priority DESC);

-- Event types registry table
CREATE TABLE IF NOT EXISTS event_types (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    schema JSONB NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_event_types_name ON event_types(name);

-- Webhook events audit log table
CREATE TABLE IF NOT EXISTS webhook_events (
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
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_webhook_events_tenant_id ON webhook_events(tenant_id);
CREATE INDEX idx_webhook_events_webhook_id ON webhook_events(webhook_id);
CREATE INDEX idx_webhook_events_execution_id ON webhook_events(execution_id);
CREATE INDEX idx_webhook_events_status ON webhook_events(status);
CREATE INDEX idx_webhook_events_created_at ON webhook_events(created_at DESC);

-- Composite index for common queries
CREATE INDEX idx_webhook_events_tenant_status_created ON webhook_events(tenant_id, status, created_at DESC);

-- Webhook filters table
CREATE TABLE IF NOT EXISTS webhook_filters (
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

-- Composite index for filter evaluation
CREATE INDEX idx_webhook_filters_webhook_enabled ON webhook_filters(webhook_id, enabled);

-- Enable Row Level Security on new tables
ALTER TABLE webhook_events ENABLE ROW LEVEL SECURITY;
ALTER TABLE webhook_filters ENABLE ROW LEVEL SECURITY;

-- Create RLS policies for tenant isolation
CREATE POLICY tenant_isolation_webhook_events ON webhook_events
    USING (tenant_id = current_setting('app.current_tenant_id', true)::UUID);

CREATE POLICY tenant_isolation_webhook_filters ON webhook_filters
    USING (
        webhook_id IN (
            SELECT id FROM webhooks
            WHERE tenant_id = current_setting('app.current_tenant_id', true)::UUID
        )
    );

-- Create trigger for event_types updated_at
CREATE TRIGGER update_event_types_updated_at BEFORE UPDATE ON event_types
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Function to increment webhook trigger count
CREATE OR REPLACE FUNCTION increment_webhook_trigger_count()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE webhooks
    SET
        trigger_count = trigger_count + 1,
        last_triggered_at = NOW()
    WHERE id = NEW.webhook_id
    AND NEW.status = 'processed';
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger to auto-update webhook trigger stats
CREATE TRIGGER update_webhook_trigger_stats
    AFTER INSERT ON webhook_events
    FOR EACH ROW
    EXECUTE FUNCTION increment_webhook_trigger_count();

-- Function to validate webhook filter operators
CREATE OR REPLACE FUNCTION validate_webhook_filter_operator()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.operator NOT IN ('equals', 'not_equals', 'contains', 'not_contains', 'regex', 'gt', 'gte', 'lt', 'lte', 'in', 'not_in', 'exists', 'not_exists') THEN
        RAISE EXCEPTION 'Invalid operator: %. Allowed operators: equals, not_equals, contains, not_contains, regex, gt, gte, lt, lte, in, not_in, exists, not_exists', NEW.operator;
    END IF;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger to validate filter operators
CREATE TRIGGER validate_webhook_filter_operator_trigger
    BEFORE INSERT OR UPDATE ON webhook_filters
    FOR EACH ROW
    EXECUTE FUNCTION validate_webhook_filter_operator();

-- Seed default event types
INSERT INTO event_types (name, description, schema, version) VALUES
(
    'webhook.received',
    'Webhook request received',
    '{
        "type": "object",
        "properties": {
            "method": {"type": "string"},
            "path": {"type": "string"},
            "headers": {"type": "object"},
            "body": {"type": "object"},
            "timestamp": {"type": "string", "format": "date-time"}
        },
        "required": ["method", "path", "timestamp"]
    }'::JSONB,
    1
),
(
    'webhook.processed',
    'Webhook successfully processed and triggered workflow',
    '{
        "type": "object",
        "properties": {
            "webhook_id": {"type": "string", "format": "uuid"},
            "workflow_id": {"type": "string", "format": "uuid"},
            "execution_id": {"type": "string", "format": "uuid"},
            "processing_time_ms": {"type": "integer"},
            "timestamp": {"type": "string", "format": "date-time"}
        },
        "required": ["webhook_id", "workflow_id", "execution_id", "timestamp"]
    }'::JSONB,
    1
),
(
    'webhook.filtered',
    'Webhook filtered out by filter rules',
    '{
        "type": "object",
        "properties": {
            "webhook_id": {"type": "string", "format": "uuid"},
            "filter_id": {"type": "string", "format": "uuid"},
            "reason": {"type": "string"},
            "timestamp": {"type": "string", "format": "date-time"}
        },
        "required": ["webhook_id", "reason", "timestamp"]
    }'::JSONB,
    1
),
(
    'webhook.failed',
    'Webhook processing failed',
    '{
        "type": "object",
        "properties": {
            "webhook_id": {"type": "string", "format": "uuid"},
            "error": {"type": "string"},
            "error_code": {"type": "string"},
            "timestamp": {"type": "string", "format": "date-time"}
        },
        "required": ["webhook_id", "error", "timestamp"]
    }'::JSONB,
    1
),
(
    'execution.started',
    'Workflow execution started',
    '{
        "type": "object",
        "properties": {
            "execution_id": {"type": "string", "format": "uuid"},
            "workflow_id": {"type": "string", "format": "uuid"},
            "workflow_version": {"type": "integer"},
            "trigger_type": {"type": "string"},
            "timestamp": {"type": "string", "format": "date-time"}
        },
        "required": ["execution_id", "workflow_id", "trigger_type", "timestamp"]
    }'::JSONB,
    1
),
(
    'execution.completed',
    'Workflow execution completed successfully',
    '{
        "type": "object",
        "properties": {
            "execution_id": {"type": "string", "format": "uuid"},
            "workflow_id": {"type": "string", "format": "uuid"},
            "duration_ms": {"type": "integer"},
            "steps_executed": {"type": "integer"},
            "timestamp": {"type": "string", "format": "date-time"}
        },
        "required": ["execution_id", "workflow_id", "timestamp"]
    }'::JSONB,
    1
),
(
    'execution.failed',
    'Workflow execution failed',
    '{
        "type": "object",
        "properties": {
            "execution_id": {"type": "string", "format": "uuid"},
            "workflow_id": {"type": "string", "format": "uuid"},
            "error": {"type": "string"},
            "failed_step": {"type": "string"},
            "timestamp": {"type": "string", "format": "date-time"}
        },
        "required": ["execution_id", "workflow_id", "error", "timestamp"]
    }'::JSONB,
    1
),
(
    'execution.cancelled',
    'Workflow execution cancelled',
    '{
        "type": "object",
        "properties": {
            "execution_id": {"type": "string", "format": "uuid"},
            "workflow_id": {"type": "string", "format": "uuid"},
            "cancelled_by": {"type": "string", "format": "uuid"},
            "reason": {"type": "string"},
            "timestamp": {"type": "string", "format": "date-time"}
        },
        "required": ["execution_id", "workflow_id", "timestamp"]
    }'::JSONB,
    1
),
(
    'step.started',
    'Workflow step started',
    '{
        "type": "object",
        "properties": {
            "execution_id": {"type": "string", "format": "uuid"},
            "step_id": {"type": "string"},
            "step_type": {"type": "string"},
            "timestamp": {"type": "string", "format": "date-time"}
        },
        "required": ["execution_id", "step_id", "step_type", "timestamp"]
    }'::JSONB,
    1
),
(
    'step.completed',
    'Workflow step completed',
    '{
        "type": "object",
        "properties": {
            "execution_id": {"type": "string", "format": "uuid"},
            "step_id": {"type": "string"},
            "duration_ms": {"type": "integer"},
            "timestamp": {"type": "string", "format": "date-time"}
        },
        "required": ["execution_id", "step_id", "timestamp"]
    }'::JSONB,
    1
),
(
    'step.failed',
    'Workflow step failed',
    '{
        "type": "object",
        "properties": {
            "execution_id": {"type": "string", "format": "uuid"},
            "step_id": {"type": "string"},
            "error": {"type": "string"},
            "retry_count": {"type": "integer"},
            "timestamp": {"type": "string", "format": "date-time"}
        },
        "required": ["execution_id", "step_id", "error", "timestamp"]
    }'::JSONB,
    1
),
(
    'step.retrying',
    'Workflow step retrying after failure',
    '{
        "type": "object",
        "properties": {
            "execution_id": {"type": "string", "format": "uuid"},
            "step_id": {"type": "string"},
            "retry_count": {"type": "integer"},
            "max_retries": {"type": "integer"},
            "timestamp": {"type": "string", "format": "date-time"}
        },
        "required": ["execution_id", "step_id", "retry_count", "timestamp"]
    }'::JSONB,
    1
)
ON CONFLICT (name) DO NOTHING;

-- Create view for webhook event statistics
CREATE OR REPLACE VIEW webhook_event_stats AS
SELECT
    webhook_id,
    COUNT(*) as total_events,
    COUNT(*) FILTER (WHERE status = 'received') as received_count,
    COUNT(*) FILTER (WHERE status = 'processed') as processed_count,
    COUNT(*) FILTER (WHERE status = 'filtered') as filtered_count,
    COUNT(*) FILTER (WHERE status = 'failed') as failed_count,
    AVG(processing_time_ms) FILTER (WHERE processing_time_ms IS NOT NULL) as avg_processing_time_ms,
    MAX(created_at) as last_event_at
FROM webhook_events
GROUP BY webhook_id;

-- Create view for webhook health status
CREATE OR REPLACE VIEW webhook_health AS
SELECT
    w.id as webhook_id,
    w.tenant_id,
    w.name,
    w.enabled,
    w.trigger_count,
    w.last_triggered_at,
    COALESCE(s.total_events, 0) as total_events,
    COALESCE(s.failed_count, 0) as failed_count,
    COALESCE(s.filtered_count, 0) as filtered_count,
    CASE
        WHEN w.enabled = false THEN 'disabled'
        WHEN s.total_events = 0 THEN 'unused'
        WHEN s.failed_count::float / NULLIF(s.total_events, 0) > 0.5 THEN 'unhealthy'
        WHEN s.failed_count::float / NULLIF(s.total_events, 0) > 0.1 THEN 'degraded'
        ELSE 'healthy'
    END as health_status,
    COALESCE(s.avg_processing_time_ms, 0) as avg_processing_time_ms
FROM webhooks w
LEFT JOIN webhook_event_stats s ON w.id = s.webhook_id;

-- Add comments for documentation
COMMENT ON TABLE webhook_events IS 'Audit log for all webhook deliveries and their processing results';
COMMENT ON TABLE webhook_filters IS 'Filter rules for conditional webhook processing';
COMMENT ON TABLE event_types IS 'Registry of event types with their JSON schemas';
COMMENT ON COLUMN webhook_events.status IS 'Event status: received, processed, filtered, failed';
COMMENT ON COLUMN webhook_filters.operator IS 'Filter operator: equals, not_equals, contains, not_contains, regex, gt, gte, lt, lte, in, not_in, exists, not_exists';
COMMENT ON COLUMN webhook_filters.logic_group IS 'Group number for AND/OR logic - filters with same group are ORed, different groups are ANDed';
COMMENT ON COLUMN webhooks.priority IS 'Webhook priority (1=lowest, higher numbers = higher priority)';
