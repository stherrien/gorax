-- Database Connectors Migration
-- Adds support for managing database connections for workflow actions

-- Database connection types
CREATE TYPE database_type AS ENUM ('postgresql', 'mysql', 'sqlite', 'mongodb');

-- Database connections table
CREATE TABLE database_connections (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    database_type database_type NOT NULL,

    -- Reference to credential storing connection details
    credential_id UUID NOT NULL REFERENCES credentials(id) ON DELETE RESTRICT,

    -- Connection configuration (non-sensitive metadata)
    connection_config JSONB NOT NULL DEFAULT '{}', -- host, port, database, ssl_mode, read_only, max_connections, timeout_seconds

    -- Connection status
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- active, inactive, error
    last_tested_at TIMESTAMPTZ,
    last_test_result JSONB, -- { "success": true/false, "error": "...", "latency_ms": 123 }

    -- Usage tracking
    last_used_at TIMESTAMPTZ,
    usage_count BIGINT NOT NULL DEFAULT 0,

    -- Audit fields
    created_by UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT unique_database_connection_name_per_tenant UNIQUE (tenant_id, name),
    CONSTRAINT valid_status CHECK (status IN ('active', 'inactive', 'error'))
);

-- Indexes for performance
CREATE INDEX idx_database_connections_tenant_id ON database_connections(tenant_id);
CREATE INDEX idx_database_connections_database_type ON database_connections(database_type);
CREATE INDEX idx_database_connections_status ON database_connections(status);
CREATE INDEX idx_database_connections_credential_id ON database_connections(credential_id);
CREATE INDEX idx_database_connections_last_used_at ON database_connections(last_used_at DESC);
CREATE INDEX idx_database_connections_created_at ON database_connections(created_at DESC);

-- Enable Row Level Security
ALTER TABLE database_connections ENABLE ROW LEVEL SECURITY;

-- RLS Policy for tenant isolation
CREATE POLICY tenant_isolation_database_connections ON database_connections
    USING (tenant_id = current_setting('app.current_tenant_id', true)::UUID);

-- Trigger for updated_at
CREATE TRIGGER update_database_connections_updated_at BEFORE UPDATE ON database_connections
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Database connection usage log (for audit and analytics)
CREATE TABLE database_connection_queries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    connection_id UUID NOT NULL REFERENCES database_connections(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    workflow_id UUID REFERENCES workflows(id) ON DELETE SET NULL,
    execution_id UUID, -- Reference to execution if within workflow context

    -- Query details
    query_type VARCHAR(50) NOT NULL, -- select, insert, update, delete, find, aggregate
    query_hash VARCHAR(64), -- Hash of query for grouping similar queries
    rows_affected INTEGER,
    execution_time_ms INTEGER,

    -- Result
    success BOOLEAN NOT NULL DEFAULT true,
    error_message TEXT,

    -- Audit
    executed_by UUID NOT NULL,
    executed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for query log
CREATE INDEX idx_database_connection_queries_connection_id ON database_connection_queries(connection_id);
CREATE INDEX idx_database_connection_queries_tenant_id ON database_connection_queries(tenant_id);
CREATE INDEX idx_database_connection_queries_workflow_id ON database_connection_queries(workflow_id);
CREATE INDEX idx_database_connection_queries_execution_id ON database_connection_queries(execution_id);
CREATE INDEX idx_database_connection_queries_executed_at ON database_connection_queries(executed_at DESC);
CREATE INDEX idx_database_connection_queries_query_type ON database_connection_queries(query_type);
CREATE INDEX idx_database_connection_queries_query_hash ON database_connection_queries(query_hash);

-- Enable RLS on query log
ALTER TABLE database_connection_queries ENABLE ROW LEVEL SECURITY;

-- RLS Policy for tenant isolation on query log
CREATE POLICY tenant_isolation_database_connection_queries ON database_connection_queries
    USING (tenant_id = current_setting('app.current_tenant_id', true)::UUID);

-- Comments for documentation
COMMENT ON TABLE database_connections IS 'Stores database connection configurations for workflow actions';
COMMENT ON COLUMN database_connections.credential_id IS 'Reference to credential storing sensitive connection details (connection string, password, etc.)';
COMMENT ON COLUMN database_connections.connection_config IS 'Non-sensitive connection metadata (host, port, database name, SSL mode, read-only flag, connection limits)';
COMMENT ON COLUMN database_connections.last_test_result IS 'Result of last connection test including success status, error message, and latency';
COMMENT ON TABLE database_connection_queries IS 'Audit log of database queries executed through connections';
COMMENT ON COLUMN database_connection_queries.query_hash IS 'SHA-256 hash of normalized query for grouping similar queries';
