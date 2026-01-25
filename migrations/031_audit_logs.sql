-- Migration: 031_audit_logs.sql
-- Description: Create comprehensive audit logging system for compliance and security monitoring

-- Audit events table with partitioning by date
CREATE TABLE IF NOT EXISTS audit_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id VARCHAR(255),
    user_email VARCHAR(255),
    category VARCHAR(50) NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    action VARCHAR(255) NOT NULL,
    resource_type VARCHAR(100),
    resource_id VARCHAR(255),
    resource_name VARCHAR(255),
    ip_address INET,
    user_agent TEXT,
    severity VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL,
    error_message TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT valid_category CHECK (category IN (
        'authentication', 'authorization', 'data_access', 'configuration',
        'workflow', 'integration', 'credential', 'user_management', 'system'
    )),
    CONSTRAINT valid_event_type CHECK (event_type IN (
        'create', 'read', 'update', 'delete', 'execute', 'login', 'logout',
        'permission_change', 'export', 'import', 'access', 'configure'
    )),
    CONSTRAINT valid_severity CHECK (severity IN ('info', 'warning', 'error', 'critical')),
    CONSTRAINT valid_status CHECK (status IN ('success', 'failure', 'partial'))
);

-- Audit log retention policies table
CREATE TABLE IF NOT EXISTS audit_retention_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    hot_retention_days INTEGER NOT NULL DEFAULT 90,
    warm_retention_days INTEGER NOT NULL DEFAULT 365,
    cold_retention_days INTEGER NOT NULL DEFAULT 2555,  -- 7 years
    archive_enabled BOOLEAN DEFAULT TRUE,
    archive_bucket VARCHAR(255),
    archive_path VARCHAR(500),
    purge_enabled BOOLEAN DEFAULT TRUE,
    last_archive_at TIMESTAMP,
    last_purge_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_tenant_policy UNIQUE (tenant_id),
    CONSTRAINT valid_retention_days CHECK (
        hot_retention_days > 0 AND
        warm_retention_days >= hot_retention_days AND
        cold_retention_days >= warm_retention_days
    )
);

-- Audit log integrity table (for tamper detection)
CREATE TABLE IF NOT EXISTS audit_log_integrity (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    event_count INTEGER NOT NULL,
    hash VARCHAR(64) NOT NULL,  -- SHA256 hash of all events for the date
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_tenant_date_integrity UNIQUE (tenant_id, date)
);

-- Create indexes for efficient querying
CREATE INDEX idx_audit_events_tenant ON audit_events(tenant_id);
CREATE INDEX idx_audit_events_user ON audit_events(user_id);
CREATE INDEX idx_audit_events_user_email ON audit_events(user_email);
CREATE INDEX idx_audit_events_category ON audit_events(category);
CREATE INDEX idx_audit_events_event_type ON audit_events(event_type);
CREATE INDEX idx_audit_events_action ON audit_events(action);
CREATE INDEX idx_audit_events_resource ON audit_events(resource_type, resource_id);
CREATE INDEX idx_audit_events_severity ON audit_events(severity);
CREATE INDEX idx_audit_events_status ON audit_events(status);
CREATE INDEX idx_audit_events_created_at ON audit_events(created_at DESC);
CREATE INDEX idx_audit_events_tenant_created ON audit_events(tenant_id, created_at DESC);
CREATE INDEX idx_audit_events_tenant_category ON audit_events(tenant_id, category);
CREATE INDEX idx_audit_events_tenant_user ON audit_events(tenant_id, user_id);
CREATE INDEX idx_audit_events_ip_address ON audit_events(ip_address);

-- GIN index for metadata JSON queries
CREATE INDEX idx_audit_events_metadata ON audit_events USING GIN(metadata);

-- Composite index for common query patterns
CREATE INDEX idx_audit_events_tenant_cat_created ON audit_events(tenant_id, category, created_at DESC);
CREATE INDEX idx_audit_events_tenant_severity_created ON audit_events(tenant_id, severity, created_at DESC);
CREATE INDEX idx_audit_events_critical ON audit_events(tenant_id, created_at DESC) WHERE severity = 'critical';
CREATE INDEX idx_audit_events_failures ON audit_events(tenant_id, created_at DESC) WHERE status = 'failure';

-- Indexes for retention policy table
CREATE INDEX idx_audit_retention_policies_tenant ON audit_retention_policies(tenant_id);
CREATE INDEX idx_audit_retention_policies_archive ON audit_retention_policies(last_archive_at) WHERE archive_enabled = TRUE;
CREATE INDEX idx_audit_retention_policies_purge ON audit_retention_policies(last_purge_at) WHERE purge_enabled = TRUE;

-- Indexes for integrity table
CREATE INDEX idx_audit_log_integrity_tenant_date ON audit_log_integrity(tenant_id, date DESC);
CREATE INDEX idx_audit_log_integrity_date ON audit_log_integrity(date DESC);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_audit_retention_policy_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to automatically update updated_at
CREATE TRIGGER audit_retention_policies_updated_at
    BEFORE UPDATE ON audit_retention_policies
    FOR EACH ROW
    EXECUTE FUNCTION update_audit_retention_policy_updated_at();

-- Insert default retention policy for existing tenants
INSERT INTO audit_retention_policies (tenant_id, hot_retention_days, warm_retention_days, cold_retention_days)
SELECT id, 90, 365, 2555
FROM tenants
ON CONFLICT (tenant_id) DO NOTHING;

-- Table comments
COMMENT ON TABLE audit_events IS 'Comprehensive audit log for all user actions and system events';
COMMENT ON TABLE audit_retention_policies IS 'Configurable retention policies for audit logs per tenant';
COMMENT ON TABLE audit_log_integrity IS 'Daily integrity hashes for tamper detection';

-- Column comments
COMMENT ON COLUMN audit_events.category IS 'Event category: authentication, authorization, data_access, configuration, workflow, integration, credential, user_management, system';
COMMENT ON COLUMN audit_events.event_type IS 'Type of event: create, read, update, delete, execute, login, logout, permission_change, export, import, access, configure';
COMMENT ON COLUMN audit_events.action IS 'Specific action taken, e.g., workflow.executed, credential.accessed';
COMMENT ON COLUMN audit_events.severity IS 'Event severity: info, warning, error, critical';
COMMENT ON COLUMN audit_events.status IS 'Event outcome: success, failure, partial';
COMMENT ON COLUMN audit_events.metadata IS 'Additional context as JSON (execution_id, duration, changes, etc.)';
COMMENT ON COLUMN audit_retention_policies.hot_retention_days IS 'Days to keep in main database (fast queries)';
COMMENT ON COLUMN audit_retention_policies.warm_retention_days IS 'Days to keep in compressed form (slower queries)';
COMMENT ON COLUMN audit_retention_policies.cold_retention_days IS 'Days to keep in archive storage (slowest queries)';
