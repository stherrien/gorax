-- Create webhook_endpoints table for dynamic workflow webhook endpoints
-- These are temporary endpoints created by webhook actions within workflow executions

CREATE TABLE IF NOT EXISTS webhook_endpoints (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    execution_id UUID NOT NULL REFERENCES executions(id) ON DELETE CASCADE,
    step_id VARCHAR(100) NOT NULL,

    -- Endpoint identification
    endpoint_token VARCHAR(64) NOT NULL UNIQUE, -- Cryptographically secure token for URL

    -- Configuration
    config JSONB NOT NULL DEFAULT '{}'::JSONB, -- Expected payload schema, timeout settings

    -- State
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    triggered_at TIMESTAMPTZ, -- When webhook was received
    payload JSONB, -- Received payload data

    -- Metadata
    source_ip VARCHAR(45), -- IPv4 or IPv6 address
    user_agent TEXT,
    content_type VARCHAR(255),

    -- Expiration and cleanup
    expires_at TIMESTAMPTZ NOT NULL,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for efficient queries
CREATE INDEX idx_webhook_endpoints_tenant ON webhook_endpoints(tenant_id);
CREATE INDEX idx_webhook_endpoints_execution ON webhook_endpoints(execution_id);
CREATE INDEX idx_webhook_endpoints_token ON webhook_endpoints(endpoint_token);
CREATE INDEX idx_webhook_endpoints_active ON webhook_endpoints(is_active) WHERE is_active = TRUE;
CREATE INDEX idx_webhook_endpoints_expires ON webhook_endpoints(expires_at) WHERE is_active = TRUE;
CREATE INDEX idx_webhook_endpoints_created_at ON webhook_endpoints(created_at DESC);

-- GIN index for config queries
CREATE INDEX idx_webhook_endpoints_config ON webhook_endpoints USING GIN(config);

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_webhook_endpoints_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_webhook_endpoints_updated_at
    BEFORE UPDATE ON webhook_endpoints
    FOR EACH ROW
    EXECUTE FUNCTION update_webhook_endpoints_updated_at();

-- Comments
COMMENT ON TABLE webhook_endpoints IS 'Dynamic webhook endpoints created by workflow webhook actions';
COMMENT ON COLUMN webhook_endpoints.endpoint_token IS 'Cryptographically secure token used in the webhook URL path';
COMMENT ON COLUMN webhook_endpoints.step_id IS 'The workflow step/node ID that created this endpoint';
COMMENT ON COLUMN webhook_endpoints.config IS 'JSON configuration including expected payload schema and validation rules';
COMMENT ON COLUMN webhook_endpoints.is_active IS 'Whether this endpoint is still accepting webhooks';
COMMENT ON COLUMN webhook_endpoints.triggered_at IS 'Timestamp when the webhook was triggered (null if not yet triggered)';
COMMENT ON COLUMN webhook_endpoints.payload IS 'The received webhook payload data';
COMMENT ON COLUMN webhook_endpoints.expires_at IS 'Timestamp when this endpoint expires and should be cleaned up';
