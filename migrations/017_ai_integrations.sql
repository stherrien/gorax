-- AI Integrations Migration
-- Adds support for AI/LLM providers and usage tracking

-- AI usage log table for tracking token consumption and costs
CREATE TABLE IF NOT EXISTS ai_usage_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    credential_id UUID,

    -- Request details
    provider VARCHAR(50) NOT NULL,
    model VARCHAR(100) NOT NULL,
    action_type VARCHAR(100) NOT NULL,
    execution_id UUID,
    workflow_id UUID,

    -- Token usage
    prompt_tokens INTEGER NOT NULL DEFAULT 0,
    completion_tokens INTEGER NOT NULL DEFAULT 0,
    total_tokens INTEGER NOT NULL DEFAULT 0,

    -- Cost tracking (in USD cents for precision)
    estimated_cost_cents INTEGER NOT NULL DEFAULT 0,

    -- Response details
    success BOOLEAN NOT NULL DEFAULT true,
    error_code VARCHAR(100),
    error_message TEXT,
    latency_ms INTEGER,

    -- Request metadata (for debugging and analytics)
    request_metadata JSONB DEFAULT '{}',

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for common queries
CREATE INDEX IF NOT EXISTS idx_ai_usage_log_tenant ON ai_usage_log(tenant_id);
CREATE INDEX IF NOT EXISTS idx_ai_usage_log_tenant_created ON ai_usage_log(tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_ai_usage_log_credential ON ai_usage_log(credential_id);
CREATE INDEX IF NOT EXISTS idx_ai_usage_log_provider_model ON ai_usage_log(provider, model);
CREATE INDEX IF NOT EXISTS idx_ai_usage_log_execution ON ai_usage_log(execution_id) WHERE execution_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_ai_usage_log_workflow ON ai_usage_log(workflow_id) WHERE workflow_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_ai_usage_log_created ON ai_usage_log(created_at DESC);

-- AI model pricing table for cost estimation
CREATE TABLE IF NOT EXISTS ai_model_pricing (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider VARCHAR(50) NOT NULL,
    model VARCHAR(100) NOT NULL,

    -- Pricing per 1 million tokens (in USD cents for precision)
    input_cost_per_million INTEGER NOT NULL,
    output_cost_per_million INTEGER NOT NULL,

    -- Model capabilities
    context_window INTEGER NOT NULL,
    max_output_tokens INTEGER,
    supports_vision BOOLEAN NOT NULL DEFAULT false,
    supports_function_calling BOOLEAN NOT NULL DEFAULT false,
    supports_json_mode BOOLEAN NOT NULL DEFAULT false,

    -- Lifecycle
    is_active BOOLEAN NOT NULL DEFAULT true,
    deprecated_at TIMESTAMPTZ,
    effective_from TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    effective_until TIMESTAMPTZ,

    -- Metadata
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT unique_model_pricing UNIQUE (provider, model, effective_from)
);

-- Index for pricing lookups
CREATE INDEX IF NOT EXISTS idx_ai_model_pricing_provider_model ON ai_model_pricing(provider, model);
CREATE INDEX IF NOT EXISTS idx_ai_model_pricing_active ON ai_model_pricing(is_active) WHERE is_active = true;

-- Seed initial pricing data (as of December 2024)
INSERT INTO ai_model_pricing (provider, model, input_cost_per_million, output_cost_per_million, context_window, max_output_tokens, supports_vision, supports_function_calling, supports_json_mode) VALUES
    -- OpenAI models
    ('openai', 'gpt-4-turbo', 1000, 3000, 128000, 4096, true, true, true),
    ('openai', 'gpt-4-turbo-preview', 1000, 3000, 128000, 4096, false, true, true),
    ('openai', 'gpt-4o', 500, 1500, 128000, 4096, true, true, true),
    ('openai', 'gpt-4o-mini', 15, 60, 128000, 16384, true, true, true),
    ('openai', 'gpt-4', 3000, 6000, 8192, 4096, false, true, true),
    ('openai', 'gpt-3.5-turbo', 50, 150, 16385, 4096, false, true, true),
    ('openai', 'text-embedding-3-small', 2, 0, 8191, 0, false, false, false),
    ('openai', 'text-embedding-3-large', 13, 0, 8191, 0, false, false, false),

    -- Anthropic models
    ('anthropic', 'claude-3-opus-20240229', 1500, 7500, 200000, 4096, true, true, false),
    ('anthropic', 'claude-3-sonnet-20240229', 300, 1500, 200000, 4096, true, true, false),
    ('anthropic', 'claude-3-haiku-20240307', 25, 125, 200000, 4096, true, true, false),
    ('anthropic', 'claude-3-5-sonnet-20241022', 300, 1500, 200000, 8192, true, true, false),

    -- AWS Bedrock (Anthropic models via Bedrock)
    ('bedrock', 'anthropic.claude-3-opus-20240229-v1:0', 1500, 7500, 200000, 4096, true, true, false),
    ('bedrock', 'anthropic.claude-3-sonnet-20240229-v1:0', 300, 1500, 200000, 4096, true, true, false),
    ('bedrock', 'anthropic.claude-3-haiku-20240307-v1:0', 25, 125, 200000, 4096, true, true, false),
    ('bedrock', 'anthropic.claude-3-5-sonnet-20241022-v2:0', 300, 1500, 200000, 8192, true, true, false),

    -- AWS Bedrock (Titan models)
    ('bedrock', 'amazon.titan-text-express-v1', 80, 240, 8000, 8000, false, false, false),
    ('bedrock', 'amazon.titan-text-lite-v1', 15, 20, 4000, 4000, false, false, false),
    ('bedrock', 'amazon.titan-embed-text-v2:0', 2, 0, 8192, 0, false, false, false)
ON CONFLICT (provider, model, effective_from) DO UPDATE SET
    input_cost_per_million = EXCLUDED.input_cost_per_million,
    output_cost_per_million = EXCLUDED.output_cost_per_million,
    context_window = EXCLUDED.context_window,
    max_output_tokens = EXCLUDED.max_output_tokens,
    supports_vision = EXCLUDED.supports_vision,
    supports_function_calling = EXCLUDED.supports_function_calling,
    supports_json_mode = EXCLUDED.supports_json_mode,
    updated_at = NOW();

-- AI tenant quotas table (optional - for per-tenant limits)
CREATE TABLE IF NOT EXISTS ai_tenant_quotas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL UNIQUE,

    -- Token limits (per month, 0 = unlimited)
    monthly_token_limit INTEGER NOT NULL DEFAULT 0,
    monthly_cost_limit_cents INTEGER NOT NULL DEFAULT 0,

    -- Rate limits (per minute)
    requests_per_minute INTEGER NOT NULL DEFAULT 60,
    tokens_per_minute INTEGER NOT NULL DEFAULT 100000,

    -- Current usage (reset monthly)
    current_month_tokens INTEGER NOT NULL DEFAULT 0,
    current_month_cost_cents INTEGER NOT NULL DEFAULT 0,
    usage_reset_at TIMESTAMPTZ NOT NULL DEFAULT DATE_TRUNC('month', NOW()) + INTERVAL '1 month',

    -- Lifecycle
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ai_tenant_quotas_tenant ON ai_tenant_quotas(tenant_id);

-- Function to get current model pricing
CREATE OR REPLACE FUNCTION get_ai_model_pricing(p_provider VARCHAR, p_model VARCHAR)
RETURNS TABLE (
    input_cost_per_million INTEGER,
    output_cost_per_million INTEGER,
    context_window INTEGER,
    max_output_tokens INTEGER
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        amp.input_cost_per_million,
        amp.output_cost_per_million,
        amp.context_window,
        amp.max_output_tokens
    FROM ai_model_pricing amp
    WHERE amp.provider = p_provider
      AND amp.model = p_model
      AND amp.is_active = true
      AND (amp.effective_until IS NULL OR amp.effective_until > NOW())
    ORDER BY amp.effective_from DESC
    LIMIT 1;
END;
$$ LANGUAGE plpgsql;

-- Function to estimate cost for a request
CREATE OR REPLACE FUNCTION estimate_ai_cost(
    p_provider VARCHAR,
    p_model VARCHAR,
    p_prompt_tokens INTEGER,
    p_completion_tokens INTEGER
)
RETURNS INTEGER AS $$
DECLARE
    v_input_cost INTEGER;
    v_output_cost INTEGER;
    v_total_cost INTEGER;
BEGIN
    SELECT input_cost_per_million, output_cost_per_million
    INTO v_input_cost, v_output_cost
    FROM get_ai_model_pricing(p_provider, p_model);

    IF v_input_cost IS NULL THEN
        RETURN 0;
    END IF;

    -- Calculate cost: (tokens / 1,000,000) * cost_per_million
    -- Multiply first then divide to avoid precision loss
    v_total_cost := (p_prompt_tokens * v_input_cost + p_completion_tokens * v_output_cost) / 1000000;

    RETURN v_total_cost;
END;
$$ LANGUAGE plpgsql;

-- View for monthly AI usage summary by tenant
CREATE OR REPLACE VIEW ai_usage_monthly_summary AS
SELECT
    tenant_id,
    DATE_TRUNC('month', created_at) AS month,
    provider,
    model,
    COUNT(*) AS request_count,
    SUM(prompt_tokens) AS total_prompt_tokens,
    SUM(completion_tokens) AS total_completion_tokens,
    SUM(total_tokens) AS total_tokens,
    SUM(estimated_cost_cents) AS total_cost_cents,
    AVG(latency_ms)::INTEGER AS avg_latency_ms,
    COUNT(*) FILTER (WHERE success = true) AS successful_requests,
    COUNT(*) FILTER (WHERE success = false) AS failed_requests
FROM ai_usage_log
GROUP BY tenant_id, DATE_TRUNC('month', created_at), provider, model;

-- Comments for documentation
COMMENT ON TABLE ai_usage_log IS 'Tracks all AI/LLM API calls for billing and analytics';
COMMENT ON TABLE ai_model_pricing IS 'Stores pricing information for AI models';
COMMENT ON TABLE ai_tenant_quotas IS 'Optional per-tenant AI usage quotas and limits';
COMMENT ON FUNCTION get_ai_model_pricing IS 'Returns current pricing for a given provider and model';
COMMENT ON FUNCTION estimate_ai_cost IS 'Estimates cost in USD cents for a given token count';
