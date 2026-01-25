-- Credential Version Tracking Migration
-- Adds support for credential version history and proper key rotation

-- Credential versions table for tracking version history
DO $$ BEGIN
    CREATE TABLE IF NOT EXISTS credential_versions (
        id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
        credential_id UUID NOT NULL REFERENCES credentials(id) ON DELETE CASCADE,
        tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
        version INTEGER NOT NULL DEFAULT 1,

        -- Envelope encryption fields (copied from main credential on rotation)
        encrypted_dek BYTEA NOT NULL,
        ciphertext BYTEA NOT NULL,
        nonce BYTEA NOT NULL,
        auth_tag BYTEA NOT NULL,
        kms_key_id VARCHAR(255) NOT NULL,

        -- Version metadata
        is_active BOOLEAN NOT NULL DEFAULT true,
        created_by UUID NOT NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        deactivated_at TIMESTAMPTZ,
        deactivated_by UUID,
        rotation_reason VARCHAR(255),

        -- Ensure unique version per credential
        CONSTRAINT unique_credential_version UNIQUE (credential_id, version)
    );
EXCEPTION WHEN duplicate_table THEN
    -- Table already exists, do nothing
    NULL;
END $$;

-- Indexes for credential_versions
DO $$ BEGIN
    CREATE INDEX IF NOT EXISTS idx_credential_versions_credential_id
        ON credential_versions(credential_id);
EXCEPTION WHEN duplicate_table THEN NULL;
END $$;

DO $$ BEGIN
    CREATE INDEX IF NOT EXISTS idx_credential_versions_tenant_id
        ON credential_versions(tenant_id);
EXCEPTION WHEN duplicate_table THEN NULL;
END $$;

DO $$ BEGIN
    CREATE INDEX IF NOT EXISTS idx_credential_versions_created_at
        ON credential_versions(created_at DESC);
EXCEPTION WHEN duplicate_table THEN NULL;
END $$;

DO $$ BEGIN
    CREATE INDEX IF NOT EXISTS idx_credential_versions_active
        ON credential_versions(credential_id, is_active) WHERE is_active = true;
EXCEPTION WHEN duplicate_table THEN NULL;
END $$;

-- Enable Row Level Security on credential_versions
ALTER TABLE credential_versions ENABLE ROW LEVEL SECURITY;

-- RLS Policy for tenant isolation (drop existing and recreate)
DO $$ BEGIN
    DROP POLICY IF EXISTS tenant_isolation_credential_versions ON credential_versions;
    CREATE POLICY tenant_isolation_credential_versions ON credential_versions
        USING (tenant_id = current_setting('app.current_tenant_id', true)::UUID);
EXCEPTION WHEN undefined_object THEN
    CREATE POLICY tenant_isolation_credential_versions ON credential_versions
        USING (tenant_id = current_setting('app.current_tenant_id', true)::UUID);
END $$;

-- Add version column to credentials if it doesn't exist
DO $$ BEGIN
    ALTER TABLE credentials ADD COLUMN IF NOT EXISTS current_version INTEGER DEFAULT 1;
EXCEPTION WHEN duplicate_column THEN
    -- Column already exists
    NULL;
END $$;

-- Encryption key management table for master key rotation tracking
DO $$ BEGIN
    CREATE TABLE IF NOT EXISTS encryption_keys (
        id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
        key_id VARCHAR(255) NOT NULL UNIQUE, -- KMS Key ID or identifier
        key_type VARCHAR(50) NOT NULL DEFAULT 'kms', -- 'kms', 'simple', 'hsm'
        status VARCHAR(50) NOT NULL DEFAULT 'active', -- 'active', 'pending_rotation', 'retired'

        -- Key metadata
        algorithm VARCHAR(50) NOT NULL DEFAULT 'AES-256-GCM',
        key_spec VARCHAR(50) DEFAULT 'AES_256',

        -- Rotation tracking
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        activated_at TIMESTAMPTZ,
        last_used_at TIMESTAMPTZ,
        scheduled_rotation_at TIMESTAMPTZ,
        retired_at TIMESTAMPTZ,

        -- Metadata
        metadata JSONB NOT NULL DEFAULT '{}'
    );
EXCEPTION WHEN duplicate_table THEN
    -- Table already exists
    NULL;
END $$;

-- Index for encryption_keys
DO $$ BEGIN
    CREATE INDEX IF NOT EXISTS idx_encryption_keys_status
        ON encryption_keys(status);
EXCEPTION WHEN duplicate_table THEN NULL;
END $$;

DO $$ BEGIN
    CREATE INDEX IF NOT EXISTS idx_encryption_keys_key_id
        ON encryption_keys(key_id);
EXCEPTION WHEN duplicate_table THEN NULL;
END $$;

-- Key rotation history table
DO $$ BEGIN
    CREATE TABLE IF NOT EXISTS encryption_key_rotations (
        id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
        old_key_id VARCHAR(255) NOT NULL,
        new_key_id VARCHAR(255) NOT NULL,
        rotation_type VARCHAR(50) NOT NULL DEFAULT 'scheduled', -- 'scheduled', 'manual', 'emergency'
        status VARCHAR(50) NOT NULL DEFAULT 'pending', -- 'pending', 'in_progress', 'completed', 'failed'

        -- Progress tracking
        credentials_total INTEGER DEFAULT 0,
        credentials_rotated INTEGER DEFAULT 0,

        -- Timestamps
        started_at TIMESTAMPTZ,
        completed_at TIMESTAMPTZ,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

        -- Audit
        initiated_by UUID,
        error_message TEXT,

        CONSTRAINT fk_old_key FOREIGN KEY (old_key_id) REFERENCES encryption_keys(key_id),
        CONSTRAINT fk_new_key FOREIGN KEY (new_key_id) REFERENCES encryption_keys(key_id)
    );
EXCEPTION WHEN duplicate_table THEN
    -- Table already exists
    NULL;
END $$;

-- Index for key rotations
DO $$ BEGIN
    CREATE INDEX IF NOT EXISTS idx_encryption_key_rotations_status
        ON encryption_key_rotations(status);
EXCEPTION WHEN duplicate_table THEN NULL;
END $$;

-- Comments for documentation
COMMENT ON TABLE credential_versions IS 'Stores historical versions of credential values for rotation tracking';
COMMENT ON TABLE encryption_keys IS 'Tracks encryption master keys for rotation management';
COMMENT ON TABLE encryption_key_rotations IS 'Tracks key rotation operations and their progress';
COMMENT ON COLUMN credential_versions.is_active IS 'Whether this version is currently active for the credential';
COMMENT ON COLUMN credential_versions.version IS 'Sequential version number for this credential';
COMMENT ON COLUMN encryption_keys.status IS 'Current status of the key: active, pending_rotation, or retired';
