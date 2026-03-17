-- Events ingestion table
CREATE TABLE IF NOT EXISTS events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product VARCHAR(50) NOT NULL,       -- 'libri', 'nitro', 'brio', etc.
    event_type VARCHAR(100) NOT NULL,   -- 'page_view', 'signup', 'login', 'error'
    user_id UUID,
    session_id VARCHAR(100),
    properties JSONB DEFAULT '{}',
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_events_product ON events(product);
CREATE INDEX IF NOT EXISTS idx_events_event_type ON events(event_type);
CREATE INDEX IF NOT EXISTS idx_events_created_at ON events(created_at DESC);
