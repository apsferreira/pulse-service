-- Indices for retention metrics queries (DAU/WAU/MAU + cohorts)

-- Composite index for user_id + created_at (retention lookups by user over time)
CREATE INDEX IF NOT EXISTS idx_events_user_id_created_at
    ON events(user_id, created_at DESC)
    WHERE user_id IS NOT NULL;

-- Composite index for product + created_at (product-filtered date range queries)
CREATE INDEX IF NOT EXISTS idx_events_product_created_at
    ON events(product, created_at DESC);

-- Composite index for product + user_id + created_at (filtered retention per product)
CREATE INDEX IF NOT EXISTS idx_events_product_user_id_created_at
    ON events(product, user_id, created_at DESC)
    WHERE user_id IS NOT NULL;
