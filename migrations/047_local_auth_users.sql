-- 047: usuários locais para auth sem auth-service externo (self-hosted)
-- PULSE_AUTH_MODE=local habilita este fluxo

CREATE TABLE IF NOT EXISTS pulse_users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         TEXT NOT NULL UNIQUE,
    name          TEXT NOT NULL DEFAULT '',
    password_hash TEXT NOT NULL,
    role          TEXT NOT NULL DEFAULT 'admin',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Índice para lookup por email no login
CREATE INDEX IF NOT EXISTS idx_pulse_users_email ON pulse_users(email);

-- DOWN
-- DROP INDEX IF EXISTS idx_pulse_users_email;
-- DROP TABLE IF EXISTS pulse_users;
