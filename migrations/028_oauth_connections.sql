-- migrations/028_oauth_connections.sql
-- OAuth 2.0 connections for third-party service integrations

-- OAuth provider configurations
CREATE TABLE IF NOT EXISTS oauth_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_key VARCHAR(50) NOT NULL UNIQUE, -- github, google, slack, microsoft
    name VARCHAR(100) NOT NULL,
    description TEXT,
    auth_url TEXT NOT NULL,
    token_url TEXT NOT NULL,
    user_info_url TEXT,
    default_scopes TEXT[], -- Default scopes for this provider
    client_id VARCHAR(255), -- Can be NULL for tenant-specific configs
    client_secret_encrypted BYTEA, -- Encrypted client secret
    client_secret_nonce BYTEA,
    client_secret_auth_tag BYTEA,
    client_secret_encrypted_dek BYTEA,
    client_secret_kms_key_id VARCHAR(255),
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- active, inactive
    config JSONB DEFAULT '{}', -- Provider-specific configuration
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT oauth_providers_status_check CHECK (status IN ('active', 'inactive'))
);

-- Index for provider lookups
CREATE INDEX IF NOT EXISTS idx_oauth_providers_key ON oauth_providers(provider_key);
CREATE INDEX IF NOT EXISTS idx_oauth_providers_status ON oauth_providers(status);

-- OAuth connections per user per tenant
CREATE TABLE IF NOT EXISTS oauth_connections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    provider_key VARCHAR(50) NOT NULL,
    provider_user_id VARCHAR(255), -- User ID from OAuth provider
    provider_username VARCHAR(255), -- Username/email from provider
    provider_email VARCHAR(255), -- Email from provider

    -- Encrypted tokens using envelope encryption
    access_token_encrypted BYTEA NOT NULL,
    access_token_nonce BYTEA NOT NULL,
    access_token_auth_tag BYTEA NOT NULL,
    access_token_encrypted_dek BYTEA NOT NULL,
    access_token_kms_key_id VARCHAR(255) NOT NULL,

    refresh_token_encrypted BYTEA,
    refresh_token_nonce BYTEA,
    refresh_token_auth_tag BYTEA,
    refresh_token_encrypted_dek BYTEA,
    refresh_token_kms_key_id VARCHAR(255),

    token_expiry TIMESTAMPTZ, -- When access token expires
    scopes TEXT[] NOT NULL, -- Granted scopes

    status VARCHAR(20) NOT NULL DEFAULT 'active', -- active, revoked, expired

    -- Audit fields
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMPTZ, -- Last time this connection was used
    last_refresh_at TIMESTAMPTZ, -- Last time token was refreshed

    -- Metadata
    raw_token_response JSONB, -- Store raw OAuth response for debugging
    metadata JSONB DEFAULT '{}', -- Additional provider-specific data

    CONSTRAINT oauth_connections_status_check CHECK (status IN ('active', 'revoked', 'expired')),
    -- One active connection per user per tenant per provider
    CONSTRAINT oauth_connections_unique_active UNIQUE (user_id, tenant_id, provider_key)
);

-- Indexes for connection lookups
CREATE INDEX IF NOT EXISTS idx_oauth_connections_user_tenant ON oauth_connections(user_id, tenant_id);
CREATE INDEX IF NOT EXISTS idx_oauth_connections_provider ON oauth_connections(provider_key);
CREATE INDEX IF NOT EXISTS idx_oauth_connections_status ON oauth_connections(status);
CREATE INDEX IF NOT EXISTS idx_oauth_connections_tenant ON oauth_connections(tenant_id);
CREATE INDEX IF NOT EXISTS idx_oauth_connections_expiry ON oauth_connections(token_expiry) WHERE status = 'active';

-- OAuth state tracking for CSRF protection
CREATE TABLE IF NOT EXISTS oauth_states (
    state VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    provider_key VARCHAR(50) NOT NULL,
    redirect_uri TEXT,
    code_verifier VARCHAR(128), -- For PKCE
    scopes TEXT[],
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    used BOOLEAN NOT NULL DEFAULT FALSE
);

-- Index for state validation and cleanup
CREATE INDEX IF NOT EXISTS idx_oauth_states_expires ON oauth_states(expires_at);
CREATE INDEX IF NOT EXISTS idx_oauth_states_user_tenant ON oauth_states(user_id, tenant_id);

-- OAuth connection usage logs
CREATE TABLE IF NOT EXISTS oauth_connection_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    connection_id UUID NOT NULL REFERENCES oauth_connections(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    action VARCHAR(50) NOT NULL, -- authorize, token_refresh, api_call, revoke
    success BOOLEAN NOT NULL,
    error_message TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for logs
CREATE INDEX IF NOT EXISTS idx_oauth_logs_connection ON oauth_connection_logs(connection_id);
CREATE INDEX IF NOT EXISTS idx_oauth_logs_tenant ON oauth_connection_logs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_oauth_logs_created ON oauth_connection_logs(created_at);

-- Insert default OAuth providers
INSERT INTO oauth_providers (provider_key, name, description, auth_url, token_url, user_info_url, default_scopes, status) VALUES
    ('github', 'GitHub', 'GitHub API integration', 'https://github.com/login/oauth/authorize', 'https://github.com/login/oauth/access_token', 'https://api.github.com/user', ARRAY['user', 'repo'], 'active'),
    ('google', 'Google', 'Google Workspace and APIs integration', 'https://accounts.google.com/o/oauth2/v2/auth', 'https://oauth2.googleapis.com/token', 'https://www.googleapis.com/oauth2/v1/userinfo', ARRAY['https://www.googleapis.com/auth/userinfo.email', 'https://www.googleapis.com/auth/userinfo.profile'], 'active'),
    ('slack', 'Slack', 'Slack workspace integration', 'https://slack.com/oauth/v2/authorize', 'https://slack.com/api/oauth.v2.access', 'https://slack.com/api/users.identity', ARRAY['chat:write', 'channels:read'], 'active'),
    ('microsoft', 'Microsoft', 'Microsoft 365 and Azure integration', 'https://login.microsoftonline.com/common/oauth2/v2.0/authorize', 'https://login.microsoftonline.com/common/oauth2/v2.0/token', 'https://graph.microsoft.com/v1.0/me', ARRAY['user.read', 'mail.read'], 'active')
ON CONFLICT (provider_key) DO NOTHING;

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_oauth_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Triggers for updated_at
CREATE TRIGGER oauth_providers_updated_at
    BEFORE UPDATE ON oauth_providers
    FOR EACH ROW
    EXECUTE FUNCTION update_oauth_updated_at();

CREATE TRIGGER oauth_connections_updated_at
    BEFORE UPDATE ON oauth_connections
    FOR EACH ROW
    EXECUTE FUNCTION update_oauth_updated_at();

-- Function to clean up expired OAuth states (older than 1 hour)
CREATE OR REPLACE FUNCTION cleanup_expired_oauth_states()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM oauth_states WHERE expires_at < NOW();
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Comments for documentation
COMMENT ON TABLE oauth_providers IS 'OAuth 2.0 provider configurations for third-party integrations';
COMMENT ON TABLE oauth_connections IS 'User OAuth connections with encrypted tokens per tenant';
COMMENT ON TABLE oauth_states IS 'Temporary OAuth state for CSRF protection and PKCE';
COMMENT ON TABLE oauth_connection_logs IS 'Audit log for OAuth connection operations';
COMMENT ON COLUMN oauth_connections.access_token_encrypted IS 'Encrypted access token using envelope encryption';
COMMENT ON COLUMN oauth_connections.refresh_token_encrypted IS 'Encrypted refresh token using envelope encryption';
COMMENT ON COLUMN oauth_states.code_verifier IS 'PKCE code verifier for enhanced security';
