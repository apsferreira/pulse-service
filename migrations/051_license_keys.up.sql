CREATE TABLE IF NOT EXISTS license_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key VARCHAR(64) UNIQUE NOT NULL,
    plan VARCHAR(20) NOT NULL DEFAULT 'self-hosted',
    customer_email VARCHAR(255) NOT NULL,
    customer_name VARCHAR(255),
    order_id VARCHAR(255),
    valid_from TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    valid_until TIMESTAMPTZ NOT NULL,
    activated_at TIMESTAMPTZ,
    activation_hostname VARCHAR(255),
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_license_keys_key ON license_keys(key);
CREATE INDEX idx_license_keys_order ON license_keys(order_id);
