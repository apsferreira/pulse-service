-- Additional indices for event tracking queries
-- Composite index for count queries filtered by event_type + date range
CREATE INDEX IF NOT EXISTS idx_events_event_type_created_at ON events(event_type, created_at DESC);

-- Index for user_id filtering
CREATE INDEX IF NOT EXISTS idx_events_user_id ON events(user_id) WHERE user_id IS NOT NULL;
