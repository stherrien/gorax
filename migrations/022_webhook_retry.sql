-- Add webhook retry functionality
-- This migration adds retry tracking fields to webhook_events table

-- Add retry-related columns to webhook_events table
ALTER TABLE webhook_events ADD COLUMN IF NOT EXISTS retry_count INTEGER NOT NULL DEFAULT 0;
ALTER TABLE webhook_events ADD COLUMN IF NOT EXISTS max_retries INTEGER NOT NULL DEFAULT 3;
ALTER TABLE webhook_events ADD COLUMN IF NOT EXISTS next_retry_at TIMESTAMPTZ;
ALTER TABLE webhook_events ADD COLUMN IF NOT EXISTS last_retry_at TIMESTAMPTZ;
ALTER TABLE webhook_events ADD COLUMN IF NOT EXISTS retry_error TEXT;
ALTER TABLE webhook_events ADD COLUMN IF NOT EXISTS permanently_failed BOOLEAN NOT NULL DEFAULT false;

-- Create index for retry processing queries
CREATE INDEX IF NOT EXISTS idx_webhook_events_next_retry ON webhook_events(next_retry_at)
    WHERE next_retry_at IS NOT NULL AND permanently_failed = false AND status = 'failed';

-- Create index for failed events
CREATE INDEX IF NOT EXISTS idx_webhook_events_permanently_failed ON webhook_events(permanently_failed, status);

-- Create index for retry count queries
CREATE INDEX IF NOT EXISTS idx_webhook_events_retry_count ON webhook_events(retry_count);

-- Add comments for documentation
COMMENT ON COLUMN webhook_events.retry_count IS 'Number of retry attempts made for this webhook delivery';
COMMENT ON COLUMN webhook_events.max_retries IS 'Maximum number of retry attempts allowed (default: 3)';
COMMENT ON COLUMN webhook_events.next_retry_at IS 'Timestamp when the next retry should be attempted';
COMMENT ON COLUMN webhook_events.last_retry_at IS 'Timestamp of the last retry attempt';
COMMENT ON COLUMN webhook_events.retry_error IS 'Error message from the last retry attempt';
COMMENT ON COLUMN webhook_events.permanently_failed IS 'True if max retries exceeded or non-retryable error occurred';

-- Create function to calculate next retry time with exponential backoff
CREATE OR REPLACE FUNCTION calculate_next_retry_time(
    attempt INTEGER,
    base_delay_seconds INTEGER DEFAULT 1,
    max_delay_seconds INTEGER DEFAULT 30,
    multiplier NUMERIC DEFAULT 2.0
) RETURNS TIMESTAMPTZ AS $$
DECLARE
    delay_seconds NUMERIC;
BEGIN
    -- Exponential backoff: base_delay * multiplier^attempt
    delay_seconds := base_delay_seconds * POWER(multiplier, attempt);

    -- Cap at max_delay
    IF delay_seconds > max_delay_seconds THEN
        delay_seconds := max_delay_seconds;
    END IF;

    RETURN NOW() + (delay_seconds || ' seconds')::INTERVAL;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Create function to mark event for retry
CREATE OR REPLACE FUNCTION mark_webhook_event_for_retry(
    event_id UUID,
    error_msg TEXT
) RETURNS void AS $$
DECLARE
    current_retry_count INTEGER;
    current_max_retries INTEGER;
    next_retry TIMESTAMPTZ;
BEGIN
    -- Get current retry info
    SELECT retry_count, max_retries INTO current_retry_count, current_max_retries
    FROM webhook_events
    WHERE id = event_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Webhook event not found: %', event_id;
    END IF;

    -- Increment retry count
    current_retry_count := current_retry_count + 1;

    -- Check if max retries exceeded
    IF current_retry_count >= current_max_retries THEN
        -- Mark as permanently failed
        UPDATE webhook_events
        SET
            retry_count = current_retry_count,
            last_retry_at = NOW(),
            retry_error = error_msg,
            permanently_failed = true,
            next_retry_at = NULL
        WHERE id = event_id;
    ELSE
        -- Calculate next retry time
        next_retry := calculate_next_retry_time(current_retry_count);

        -- Update for next retry
        UPDATE webhook_events
        SET
            retry_count = current_retry_count,
            last_retry_at = NOW(),
            retry_error = error_msg,
            next_retry_at = next_retry,
            status = 'failed'
        WHERE id = event_id;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Create function to get events ready for retry
CREATE OR REPLACE FUNCTION get_webhook_events_for_retry(
    batch_size INTEGER DEFAULT 100
) RETURNS SETOF webhook_events AS $$
BEGIN
    RETURN QUERY
    SELECT *
    FROM webhook_events
    WHERE
        status = 'failed'
        AND permanently_failed = false
        AND next_retry_at IS NOT NULL
        AND next_retry_at <= NOW()
        AND retry_count < max_retries
    ORDER BY next_retry_at ASC
    LIMIT batch_size
    FOR UPDATE SKIP LOCKED;
END;
$$ LANGUAGE plpgsql;

-- Create view for retry statistics
CREATE OR REPLACE VIEW webhook_retry_stats AS
SELECT
    webhook_id,
    COUNT(*) FILTER (WHERE retry_count > 0) as total_retried_events,
    COUNT(*) FILTER (WHERE permanently_failed = true) as permanently_failed_events,
    COUNT(*) FILTER (WHERE next_retry_at IS NOT NULL AND permanently_failed = false) as pending_retries,
    AVG(retry_count) FILTER (WHERE retry_count > 0) as avg_retry_count,
    MAX(retry_count) as max_retry_count
FROM webhook_events
WHERE status = 'failed' OR permanently_failed = true
GROUP BY webhook_id;

-- Add retry stats to webhook health view
DROP VIEW IF EXISTS webhook_health;
CREATE OR REPLACE VIEW webhook_health AS
SELECT
    w.id as webhook_id,
    w.tenant_id,
    w.name,
    w.enabled,
    w.trigger_count,
    w.last_triggered_at,
    COALESCE(s.total_events, 0) as total_events,
    COALESCE(s.failed_count, 0) as failed_count,
    COALESCE(s.filtered_count, 0) as filtered_count,
    COALESCE(r.total_retried_events, 0) as total_retried_events,
    COALESCE(r.permanently_failed_events, 0) as permanently_failed_events,
    COALESCE(r.pending_retries, 0) as pending_retries,
    CASE
        WHEN w.enabled = false THEN 'disabled'
        WHEN s.total_events = 0 THEN 'unused'
        WHEN s.failed_count::float / NULLIF(s.total_events, 0) > 0.5 THEN 'unhealthy'
        WHEN s.failed_count::float / NULLIF(s.total_events, 0) > 0.1 THEN 'degraded'
        ELSE 'healthy'
    END as health_status,
    COALESCE(s.avg_processing_time_ms, 0) as avg_processing_time_ms
FROM webhooks w
LEFT JOIN webhook_event_stats s ON w.id = s.webhook_id
LEFT JOIN webhook_retry_stats r ON w.id = r.webhook_id;

-- Add comments for new views
COMMENT ON VIEW webhook_retry_stats IS 'Statistics about webhook event retries per webhook';
COMMENT ON VIEW webhook_health IS 'Overall health status of webhooks including retry statistics';
