-- BKL-763: mapeamento cross-produto de identidade de usuário
-- UP
CREATE TABLE IF NOT EXISTS user_identities (
    id                SERIAL PRIMARY KEY,
    canonical_user_id UUID        NOT NULL,  -- UUID do auth-service
    product           VARCHAR(50) NOT NULL,  -- 'nitro', 'libri', 'brio', 'socialmake'
    product_user_id   TEXT        NOT NULL,  -- como o produto identifica o user
    email             TEXT,                  -- opcional, para correlação futura
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (product, product_user_id)
);

CREATE INDEX IF NOT EXISTS idx_user_identities_canonical ON user_identities (canonical_user_id);
CREATE INDEX IF NOT EXISTS idx_user_identities_product ON user_identities (product, product_user_id);

COMMENT ON TABLE user_identities IS 'BKL-763: mapeamento cross-produto de identidade de usuário';

-- DOWN
-- DROP INDEX IF EXISTS idx_user_identities_product;
-- DROP INDEX IF EXISTS idx_user_identities_canonical;
-- DROP TABLE IF EXISTS user_identities;
