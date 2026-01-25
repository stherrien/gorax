-- Migration: 038_default_tenant.sql
-- Purpose: Seed default tenant for single-tenant deployments
-- This migration ensures backwards compatibility with existing single-tenant installations

-- Insert default tenant if it doesn't exist
-- Using DO block to handle idempotent insertion
DO $$
DECLARE
    default_tenant_id UUID := 'a0000000-0000-0000-0000-000000000001';
    default_quotas JSONB := '{
        "max_workflows": -1,
        "max_executions_per_day": -1,
        "max_concurrent_executions": 100,
        "max_storage_bytes": -1,
        "max_api_calls_per_minute": 1000,
        "execution_history_retention_days": 365
    }';
    default_settings JSONB := '{
        "default_timezone": "UTC",
        "webhook_secret": ""
    }';
BEGIN
    -- Check if default tenant already exists
    IF NOT EXISTS (SELECT 1 FROM tenants WHERE subdomain = 'default') THEN
        INSERT INTO tenants (
            id,
            name,
            subdomain,
            status,
            tier,
            settings,
            quotas,
            created_at,
            updated_at
        ) VALUES (
            default_tenant_id,
            'Default Tenant',
            'default',
            'active',
            'enterprise',
            default_settings,
            default_quotas,
            NOW(),
            NOW()
        );

        RAISE NOTICE 'Created default tenant with ID: %', default_tenant_id;
    ELSE
        RAISE NOTICE 'Default tenant already exists, skipping creation';
    END IF;
END $$;

-- Add composite indexes on (tenant_id, id) for better tenant-scoped queries
-- These indexes improve performance for queries that filter by tenant_id first

-- Only create indexes if they don't already exist
DO $$
BEGIN
    -- Workflows composite index
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_workflows_tenant_id_id') THEN
        CREATE INDEX idx_workflows_tenant_id_id ON workflows(tenant_id, id);
        RAISE NOTICE 'Created index idx_workflows_tenant_id_id';
    END IF;

    -- Executions composite index
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_executions_tenant_id_id') THEN
        CREATE INDEX idx_executions_tenant_id_id ON executions(tenant_id, id);
        RAISE NOTICE 'Created index idx_executions_tenant_id_id';
    END IF;

    -- Credentials composite index
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_credentials_tenant_id_id') THEN
        CREATE INDEX idx_credentials_tenant_id_id ON credentials(tenant_id, id);
        RAISE NOTICE 'Created index idx_credentials_tenant_id_id';
    END IF;

    -- Webhooks composite index
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_webhooks_tenant_id_id') THEN
        CREATE INDEX idx_webhooks_tenant_id_id ON webhooks(tenant_id, id);
        RAISE NOTICE 'Created index idx_webhooks_tenant_id_id';
    END IF;

    -- Users composite index
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_users_tenant_id_id') THEN
        CREATE INDEX idx_users_tenant_id_id ON users(tenant_id, id);
        RAISE NOTICE 'Created index idx_users_tenant_id_id';
    END IF;

    -- API keys composite index
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_api_keys_tenant_id_id') THEN
        CREATE INDEX idx_api_keys_tenant_id_id ON api_keys(tenant_id, id);
        RAISE NOTICE 'Created index idx_api_keys_tenant_id_id';
    END IF;

    -- Audit logs composite index
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_audit_logs_tenant_id_id') THEN
        CREATE INDEX idx_audit_logs_tenant_id_id ON audit_logs(tenant_id, id);
        RAISE NOTICE 'Created index idx_audit_logs_tenant_id_id';
    END IF;
END $$;

-- Add suspended status to tenants check constraint if it doesn't exist
-- This ensures the database enforces valid tenant statuses
DO $$
BEGIN
    -- Check if the constraint already exists
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE constraint_name = 'tenants_status_check'
        AND table_name = 'tenants'
    ) THEN
        ALTER TABLE tenants ADD CONSTRAINT tenants_status_check
        CHECK (status IN ('active', 'inactive', 'suspended', 'deleted'));
        RAISE NOTICE 'Added tenants_status_check constraint';
    END IF;

    -- Check if tier constraint exists
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE constraint_name = 'tenants_tier_check'
        AND table_name = 'tenants'
    ) THEN
        ALTER TABLE tenants ADD CONSTRAINT tenants_tier_check
        CHECK (tier IN ('free', 'professional', 'enterprise'));
        RAISE NOTICE 'Added tenants_tier_check constraint';
    END IF;
END $$;

-- Add comment for documentation
COMMENT ON TABLE tenants IS 'Multi-tenant organization management. Each tenant represents an isolated organization with its own workflows, credentials, and users.';
COMMENT ON COLUMN tenants.subdomain IS 'Unique subdomain identifier for the tenant. Used for tenant resolution in multi-tenant mode.';
COMMENT ON COLUMN tenants.status IS 'Tenant status: active (operational), inactive (disabled), suspended (billing/policy issues), deleted (soft deleted)';
COMMENT ON COLUMN tenants.tier IS 'Pricing tier: free (limited), professional (extended), enterprise (unlimited)';
