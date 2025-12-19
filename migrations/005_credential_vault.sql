-- Credential Vault Enhancement Migration
-- Adds envelope encryption support with AWS KMS integration

-- Drop existing credentials table to recreate with new schema
DROP TABLE IF EXISTS credentials CASCADE;

-- Enhanced credentials table with envelope encryption
CREATE TABLE credentials (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(100) NOT NULL, -- api_key, oauth2, basic_auth, custom
    description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- active, inactive, revoked

    -- Envelope encryption fields (separate for proper AES-GCM support)
    encrypted_dek BYTEA NOT NULL, -- Data Encryption Key encrypted by KMS
    ciphertext BYTEA NOT NULL, -- Credential data encrypted with DEK using AES-256-GCM
    nonce BYTEA NOT NULL, -- Nonce for AES-GCM (12 bytes)
    auth_tag BYTEA NOT NULL, -- Authentication tag for AES-GCM (16 bytes)
    kms_key_id VARCHAR(255) NOT NULL, -- KMS key ID or ARN used for DEK encryption

    -- Metadata
    metadata JSONB NOT NULL DEFAULT '{}', -- Additional metadata (tags, source, etc.)
    expires_at TIMESTAMPTZ, -- Optional expiration time

    -- Audit fields
    created_by UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMPTZ,

    -- Constraints
    CONSTRAINT unique_credential_name_per_tenant UNIQUE (tenant_id, name)
);

-- Indexes for performance
CREATE INDEX idx_credentials_tenant_id ON credentials(tenant_id);
CREATE INDEX idx_credentials_type ON credentials(type);
CREATE INDEX idx_credentials_kms_key_id ON credentials(kms_key_id);
CREATE INDEX idx_credentials_last_used_at ON credentials(last_used_at DESC);
CREATE INDEX idx_credentials_created_at ON credentials(created_at DESC);
CREATE INDEX idx_credentials_expires_at ON credentials(expires_at) WHERE expires_at IS NOT NULL;

-- Enable Row Level Security
ALTER TABLE credentials ENABLE ROW LEVEL SECURITY;

-- RLS Policy for tenant isolation
CREATE POLICY tenant_isolation_credentials ON credentials
    USING (tenant_id = current_setting('app.current_tenant_id', true)::UUID);

-- Trigger for updated_at
CREATE TRIGGER update_credentials_updated_at BEFORE UPDATE ON credentials
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Credential rotation history table
CREATE TABLE credential_rotations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    credential_id UUID NOT NULL REFERENCES credentials(id) ON DELETE CASCADE,
    previous_key_id VARCHAR(255) NOT NULL,
    new_key_id VARCHAR(255) NOT NULL,
    rotated_by UUID NOT NULL,
    rotated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reason VARCHAR(255)
);

CREATE INDEX idx_credential_rotations_credential_id ON credential_rotations(credential_id);
CREATE INDEX idx_credential_rotations_rotated_at ON credential_rotations(rotated_at DESC);

-- Credential access audit log
CREATE TABLE credential_access_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    credential_id UUID NOT NULL REFERENCES credentials(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    accessed_by UUID NOT NULL,
    access_type VARCHAR(50) NOT NULL, -- read, update, rotate, delete
    ip_address VARCHAR(45),
    user_agent TEXT,
    success BOOLEAN NOT NULL DEFAULT true,
    error_message TEXT,
    accessed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_credential_access_log_credential_id ON credential_access_log(credential_id);
CREATE INDEX idx_credential_access_log_tenant_id ON credential_access_log(tenant_id);
CREATE INDEX idx_credential_access_log_accessed_at ON credential_access_log(accessed_at DESC);

-- Enable RLS on audit tables
ALTER TABLE credential_rotations ENABLE ROW LEVEL SECURITY;
ALTER TABLE credential_access_log ENABLE ROW LEVEL SECURITY;

-- Note: Rotation and access log policies would need to check credential's tenant_id
-- For simplicity, we'll enforce tenant isolation at the application layer for these tables

-- Comments for documentation
COMMENT ON TABLE credentials IS 'Stores encrypted credentials using envelope encryption with AWS KMS';
COMMENT ON COLUMN credentials.encrypted_dek IS 'Data Encryption Key (DEK) encrypted with AWS KMS master key';
COMMENT ON COLUMN credentials.ciphertext IS 'Credential data encrypted with AES-256-GCM using the DEK';
COMMENT ON COLUMN credentials.nonce IS 'Nonce for AES-GCM encryption (12 bytes)';
COMMENT ON COLUMN credentials.auth_tag IS 'Authentication tag for AES-GCM (16 bytes)';
COMMENT ON COLUMN credentials.kms_key_id IS 'AWS KMS key ID or ARN used to encrypt the DEK';
COMMENT ON COLUMN credentials.metadata IS 'Additional metadata (tags, source, etc.) stored as JSON';
