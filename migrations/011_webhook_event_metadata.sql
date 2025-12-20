-- Add metadata field to webhook_events table
-- This migration adds a JSONB column to store additional request metadata

-- Add metadata column to webhook_events table
ALTER TABLE webhook_events ADD COLUMN IF NOT EXISTS metadata JSONB;

-- Create GIN index on metadata for efficient querying
CREATE INDEX IF NOT EXISTS idx_webhook_events_metadata ON webhook_events USING GIN (metadata);

-- Add comment to document the metadata structure
COMMENT ON COLUMN webhook_events.metadata IS 'Additional request metadata including sourceIp, userAgent, receivedAt, contentType, and contentLength';

-- Sample metadata structure:
-- {
--   "sourceIp": "192.168.1.1",
--   "userAgent": "Mozilla/5.0 ...",
--   "receivedAt": "2024-01-15T10:30:00Z",
--   "contentType": "application/json",
--   "contentLength": 256
-- }
