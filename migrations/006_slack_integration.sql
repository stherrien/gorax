-- Slack Integration Migration
-- Adds support for OAuth2-based integrations starting with Slack

-- Integration credentials table (supports Slack, GitHub, Jira, etc.)
CREATE TABLE IF NOT EXISTS integration_credentials (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    integration_type VARCHAR(50) NOT NULL, -- 'slack', 'github', 'jira', etc.
    name VARCHAR(255) NOT NULL,

    -- OAuth2 tokens (encrypted at rest)
    access_token TEXT NOT NULL,
    refresh_token TEXT,
    token_type VARCHAR(50) DEFAULT 'Bearer',
    expires_at TIMESTAMPTZ,

    -- Integration-specific metadata (stored as JSONB)
    metadata JSONB NOT NULL DEFAULT '{}',

    -- Common fields that might be in metadata:
    -- For Slack: team_id, team_name, user_id, bot_user_id, scope
    -- For GitHub: installation_id, account_login, account_type
    -- For Jira: cloud_id, site_url, user_email

    -- Audit fields
    created_by UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMPTZ,

    -- Constraints
    CONSTRAINT unique_integration_name_per_tenant UNIQUE (tenant_id, integration_type, name),
    CONSTRAINT valid_integration_type CHECK (integration_type IN ('slack', 'github', 'jira', 'google', 'stripe', 'twilio', 'sendgrid', 'aws'))
);

-- Indexes for performance
CREATE INDEX idx_integration_credentials_tenant ON integration_credentials(tenant_id);
CREATE INDEX idx_integration_credentials_type ON integration_credentials(integration_type);
CREATE INDEX idx_integration_credentials_last_used ON integration_credentials(last_used_at DESC NULLS LAST);
CREATE INDEX idx_integration_credentials_metadata ON integration_credentials USING gin(metadata);

-- Enable Row Level Security
ALTER TABLE integration_credentials ENABLE ROW LEVEL SECURITY;

-- RLS Policy for tenant isolation
CREATE POLICY tenant_isolation_integration_credentials ON integration_credentials
    USING (tenant_id = current_setting('app.current_tenant_id', true)::UUID);

-- Trigger for updated_at
CREATE TRIGGER update_integration_credentials_updated_at
    BEFORE UPDATE ON integration_credentials
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Integration usage tracking (for analytics and rate limiting)
CREATE TABLE IF NOT EXISTS integration_usage_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    credential_id UUID NOT NULL REFERENCES integration_credentials(id) ON DELETE CASCADE,

    -- Request details
    action_type VARCHAR(100) NOT NULL, -- 'slack:send_message', 'github:create_issue', etc.
    execution_id UUID, -- Optional: link to workflow execution

    -- Response details
    success BOOLEAN NOT NULL DEFAULT true,
    status_code INTEGER,
    error_message TEXT,
    response_time_ms INTEGER,

    -- Metadata
    metadata JSONB DEFAULT '{}',

    -- Timestamp
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for analytics queries
CREATE INDEX idx_integration_usage_log_tenant ON integration_usage_log(tenant_id);
CREATE INDEX idx_integration_usage_log_credential ON integration_usage_log(credential_id);
CREATE INDEX idx_integration_usage_log_action ON integration_usage_log(action_type);
CREATE INDEX idx_integration_usage_log_created ON integration_usage_log(created_at DESC);
CREATE INDEX idx_integration_usage_log_execution ON integration_usage_log(execution_id) WHERE execution_id IS NOT NULL;

-- Partitioning by month for better performance (optional, can be added later)
-- CREATE TABLE integration_usage_log_2025_12 PARTITION OF integration_usage_log
--     FOR VALUES FROM ('2025-12-01') TO ('2026-01-01');

-- Enable RLS
ALTER TABLE integration_usage_log ENABLE ROW LEVEL SECURITY;

-- RLS Policy
CREATE POLICY tenant_isolation_integration_usage_log ON integration_usage_log
    USING (tenant_id = current_setting('app.current_tenant_id', true)::UUID);

-- OAuth2 state tracking (for CSRF protection during OAuth flow)
CREATE TABLE IF NOT EXISTS oauth_states (
    state VARCHAR(255) PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    integration_type VARCHAR(50) NOT NULL,
    redirect_uri TEXT NOT NULL,

    -- Metadata for the OAuth flow
    metadata JSONB DEFAULT '{}',

    -- Expiration
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '10 minutes'
);

-- Index for cleanup
CREATE INDEX idx_oauth_states_expires ON oauth_states(expires_at);

-- Comments for documentation
COMMENT ON TABLE integration_credentials IS 'Stores OAuth2 credentials for third-party integrations (Slack, GitHub, etc.)';
COMMENT ON COLUMN integration_credentials.access_token IS 'OAuth2 access token (encrypted at rest)';
COMMENT ON COLUMN integration_credentials.metadata IS 'Integration-specific data (team_id, scopes, etc.) stored as JSON';
COMMENT ON TABLE integration_usage_log IS 'Tracks all integration API calls for analytics and rate limiting';
COMMENT ON TABLE oauth_states IS 'Temporary storage for OAuth2 state tokens (CSRF protection)';

-- Grant permissions (adjust based on your user roles)
-- GRANT SELECT, INSERT, UPDATE, DELETE ON integration_credentials TO rflow_api;
-- GRANT SELECT, INSERT ON integration_usage_log TO rflow_api;
-- GRANT SELECT, INSERT, DELETE ON oauth_states TO rflow_api;
