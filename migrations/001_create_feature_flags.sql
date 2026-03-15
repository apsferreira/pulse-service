-- Feature flags table
CREATE TABLE IF NOT EXISTS feature_flags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,  -- e.g. "new_book_search"
    description TEXT,
    is_enabled BOOLEAN NOT NULL DEFAULT false,
    product VARCHAR(50),         -- null = global; 'libri', 'nitro', 'brio'
    rollout_pct INT DEFAULT 100 CHECK (rollout_pct BETWEEN 0 AND 100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(name, product)
);

-- Flag overrides per user/tenant
CREATE TABLE IF NOT EXISTS flag_overrides (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    flag_id UUID NOT NULL REFERENCES feature_flags(id) ON DELETE CASCADE,
    user_id VARCHAR(100),
    tenant_id VARCHAR(100),
    is_enabled BOOLEAN NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(flag_id, user_id, tenant_id)
);

-- Seed some default flags
INSERT INTO feature_flags (name, description, is_enabled, product) VALUES
    ('amazon_buy_button',    'Show Amazon buy link on book pages',         true,  'libri'),
    ('public_profile',       'Enable public profile pages for users',      true,  'libri'),
    ('ai_suggestions',       'AI-powered task suggestions',                false, 'nitro'),
    ('gamification',         'XP and achievement system',                  true,  'nitro'),
    ('pulse_tracking',       'Enable pulse-service event tracking in SDKs', false, null)
ON CONFLICT DO NOTHING;
