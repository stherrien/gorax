-- +goose Up
-- SSO Providers table
CREATE TABLE IF NOT EXISTS sso_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    provider_type VARCHAR(50) NOT NULL CHECK (provider_type IN ('saml', 'oidc')),
    enabled BOOLEAN NOT NULL DEFAULT true,
    enforce_sso BOOLEAN NOT NULL DEFAULT false,
    config JSONB NOT NULL,
    domains TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID,
    updated_by UUID,
    CONSTRAINT fk_sso_provider_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT fk_sso_provider_created_by FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT fk_sso_provider_updated_by FOREIGN KEY (updated_by) REFERENCES users(id) ON DELETE SET NULL
);

-- SSO Connections table (maps users to SSO providers)
CREATE TABLE IF NOT EXISTS sso_connections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    sso_provider_id UUID NOT NULL,
    external_id VARCHAR(255) NOT NULL,
    attributes JSONB,
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_sso_connection_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_sso_connection_provider FOREIGN KEY (sso_provider_id) REFERENCES sso_providers(id) ON DELETE CASCADE,
    CONSTRAINT uq_sso_connection_external UNIQUE (sso_provider_id, external_id)
);

-- SSO Login Events table (audit trail)
CREATE TABLE IF NOT EXISTS sso_login_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sso_provider_id UUID NOT NULL,
    user_id UUID,
    external_id VARCHAR(255),
    status VARCHAR(50) NOT NULL CHECK (status IN ('success', 'failure', 'error')),
    error_message TEXT,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_sso_event_provider FOREIGN KEY (sso_provider_id) REFERENCES sso_providers(id) ON DELETE CASCADE,
    CONSTRAINT fk_sso_event_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

-- Indexes for performance
CREATE INDEX idx_sso_providers_tenant ON sso_providers(tenant_id);
CREATE INDEX idx_sso_providers_enabled ON sso_providers(enabled) WHERE enabled = true;
CREATE INDEX idx_sso_providers_domains ON sso_providers USING gin(domains);
CREATE INDEX idx_sso_connections_user ON sso_connections(user_id);
CREATE INDEX idx_sso_connections_provider ON sso_connections(sso_provider_id);
CREATE INDEX idx_sso_login_events_provider ON sso_login_events(sso_provider_id);
CREATE INDEX idx_sso_login_events_created ON sso_login_events(created_at DESC);

-- Trigger for updated_at
CREATE TRIGGER update_sso_providers_updated_at
    BEFORE UPDATE ON sso_providers
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_sso_connections_updated_at
    BEFORE UPDATE ON sso_connections
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- +goose Down
DROP TABLE IF EXISTS sso_login_events CASCADE;
DROP TABLE IF EXISTS sso_connections CASCADE;
DROP TABLE IF EXISTS sso_providers CASCADE;
