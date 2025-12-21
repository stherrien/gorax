-- Migration: 018_smart_suggestions
-- Description: Add smart suggestions for execution error analysis
-- Date: 2025-12-20

-- Execution suggestions table
CREATE TABLE IF NOT EXISTS execution_suggestions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    execution_id UUID NOT NULL,
    node_id VARCHAR(255) NOT NULL,

    -- Suggestion classification
    category VARCHAR(50) NOT NULL,
    type VARCHAR(50) NOT NULL,
    confidence VARCHAR(20) NOT NULL,

    -- Suggestion content
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    details TEXT,

    -- Actionable fix data (JSON)
    fix JSONB,

    -- Metadata
    source VARCHAR(20) NOT NULL, -- 'pattern' or 'llm'
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, applied, dismissed

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    applied_at TIMESTAMPTZ,
    dismissed_at TIMESTAMPTZ,

    -- Constraints
    CONSTRAINT valid_category CHECK (category IN ('network', 'auth', 'data', 'rate_limit', 'timeout', 'config', 'external_service', 'unknown')),
    CONSTRAINT valid_type CHECK (type IN ('retry', 'config_change', 'credential_update', 'data_fix', 'workflow_modification', 'manual_intervention')),
    CONSTRAINT valid_confidence CHECK (confidence IN ('high', 'medium', 'low')),
    CONSTRAINT valid_source CHECK (source IN ('pattern', 'llm')),
    CONSTRAINT valid_status CHECK (status IN ('pending', 'applied', 'dismissed'))
);

-- Indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_suggestions_tenant_id ON execution_suggestions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_suggestions_execution_id ON execution_suggestions(execution_id);
CREATE INDEX IF NOT EXISTS idx_suggestions_tenant_execution ON execution_suggestions(tenant_id, execution_id);
CREATE INDEX IF NOT EXISTS idx_suggestions_status ON execution_suggestions(status);
CREATE INDEX IF NOT EXISTS idx_suggestions_category ON execution_suggestions(category);
CREATE INDEX IF NOT EXISTS idx_suggestions_confidence ON execution_suggestions(confidence);
CREATE INDEX IF NOT EXISTS idx_suggestions_created_at ON execution_suggestions(created_at DESC);

-- Custom error patterns table (for tenant-specific patterns)
CREATE TABLE IF NOT EXISTS error_patterns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID, -- NULL for global patterns

    -- Pattern identification
    name VARCHAR(100) NOT NULL,
    description TEXT,

    -- Matching rules
    category VARCHAR(50) NOT NULL,
    patterns JSONB NOT NULL, -- Array of regex patterns
    http_codes JSONB, -- Array of HTTP status codes
    node_types JSONB, -- Array of node types this applies to

    -- Suggestion template
    suggestion_type VARCHAR(50) NOT NULL,
    suggestion_title VARCHAR(255) NOT NULL,
    suggestion_description TEXT NOT NULL,
    suggestion_confidence VARCHAR(20) NOT NULL,
    fix_template JSONB,

    -- Metadata
    priority INT NOT NULL DEFAULT 0, -- Higher priority patterns are checked first
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT valid_pattern_category CHECK (category IN ('network', 'auth', 'data', 'rate_limit', 'timeout', 'config', 'external_service', 'unknown')),
    CONSTRAINT valid_pattern_type CHECK (suggestion_type IN ('retry', 'config_change', 'credential_update', 'data_fix', 'workflow_modification', 'manual_intervention')),
    CONSTRAINT valid_pattern_confidence CHECK (suggestion_confidence IN ('high', 'medium', 'low'))
);

-- Indexes for error patterns
CREATE INDEX IF NOT EXISTS idx_patterns_tenant_id ON error_patterns(tenant_id);
CREATE INDEX IF NOT EXISTS idx_patterns_active ON error_patterns(is_active) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_patterns_priority ON error_patterns(priority DESC);
CREATE INDEX IF NOT EXISTS idx_patterns_category ON error_patterns(category);

-- Seed some default global error patterns
INSERT INTO error_patterns (tenant_id, name, category, patterns, http_codes, suggestion_type, suggestion_title, suggestion_description, suggestion_confidence, fix_template, priority)
VALUES
    -- Network errors
    (NULL, 'connection_refused', 'network',
     '["connection refused", "ECONNREFUSED", "dial tcp.*connection refused"]'::jsonb,
     NULL,
     'retry', 'Connection Refused',
     'The target service is not accepting connections. This may be a temporary issue.',
     'high',
     '{"action_type": "retry_with_backoff", "retry_config": {"max_retries": 5, "backoff_ms": 2000, "backoff_factor": 2.0}}'::jsonb,
     100),

    (NULL, 'dns_resolution', 'network',
     '["no such host", "DNS resolution failed", "getaddrinfo ENOTFOUND"]'::jsonb,
     NULL,
     'config_change', 'DNS Resolution Failed',
     'The hostname could not be resolved. Check if the URL is correct.',
     'high',
     '{"action_type": "config_change", "config_path": "url"}'::jsonb,
     100),

    -- Authentication errors
    (NULL, 'auth_401', 'auth',
     NULL,
     '[401]'::jsonb,
     'credential_update', 'Authentication Failed',
     'The credentials used for this action are invalid or expired. Please update your credentials.',
     'high',
     '{"action_type": "credential_update"}'::jsonb,
     100),

    (NULL, 'auth_403', 'auth',
     NULL,
     '[403]'::jsonb,
     'credential_update', 'Access Forbidden',
     'The credentials do not have permission for this operation. Check the credential permissions.',
     'high',
     '{"action_type": "credential_update"}'::jsonb,
     100),

    -- Rate limiting
    (NULL, 'rate_limit_429', 'rate_limit',
     '["rate limit", "too many requests", "throttle", "exceeded.*limit"]'::jsonb,
     '[429]'::jsonb,
     'config_change', 'Rate Limit Exceeded',
     'The API rate limit was exceeded. Consider adding delays between requests or reducing request frequency.',
     'high',
     '{"action_type": "config_change", "config_path": "rate_limit", "new_value": {"delay_ms": 1000, "max_concurrent": 1}}'::jsonb,
     100),

    -- Timeout errors
    (NULL, 'timeout', 'timeout',
     '["timeout", "timed out", "deadline exceeded", "context deadline exceeded"]'::jsonb,
     '[504, 408]'::jsonb,
     'config_change', 'Request Timeout',
     'The request took too long to complete. Consider increasing the timeout value.',
     'high',
     '{"action_type": "config_change", "config_path": "timeout", "new_value": 60}'::jsonb,
     100),

    -- Data/parsing errors
    (NULL, 'json_parse', 'data',
     '["invalid json", "json.*parse error", "unexpected token", "syntax error.*json", "invalid character"]'::jsonb,
     NULL,
     'data_fix', 'Invalid JSON Data',
     'The data format is invalid JSON. Check the input data structure and ensure it is valid JSON.',
     'medium',
     '{"action_type": "data_fix"}'::jsonb,
     90),

    (NULL, 'validation_error', 'data',
     '["validation.*failed", "required field", "invalid.*format", "must be.*type"]'::jsonb,
     '[400, 422]'::jsonb,
     'data_fix', 'Data Validation Failed',
     'The input data does not meet validation requirements. Check the data format and required fields.',
     'medium',
     '{"action_type": "data_fix"}'::jsonb,
     90),

    -- Server errors
    (NULL, 'server_error_500', 'external_service',
     NULL,
     '[500]'::jsonb,
     'retry', 'Internal Server Error',
     'The external service returned an internal error. This is usually a temporary issue.',
     'medium',
     '{"action_type": "retry_with_backoff", "retry_config": {"max_retries": 3, "backoff_ms": 5000, "backoff_factor": 2.0}}'::jsonb,
     80),

    (NULL, 'server_error_502_503', 'external_service',
     NULL,
     '[502, 503]'::jsonb,
     'retry', 'Service Unavailable',
     'The external service is temporarily unavailable. This is usually a temporary issue.',
     'high',
     '{"action_type": "retry_with_backoff", "retry_config": {"max_retries": 5, "backoff_ms": 3000, "backoff_factor": 2.0}}'::jsonb,
     80)

ON CONFLICT DO NOTHING;

-- Add comment for documentation
COMMENT ON TABLE execution_suggestions IS 'Smart suggestions for fixing workflow execution errors';
COMMENT ON TABLE error_patterns IS 'Pattern definitions for automatic error classification and suggestions';
