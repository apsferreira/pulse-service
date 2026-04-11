-- BKL-1075: Ambientes por projeto (prod/staging/dev).
-- Cada projeto pode ter múltiplos ambientes independentes com API keys próprias.
-- Flags e eventos são scoped por environment, permitindo rollout progressivo.

CREATE TABLE IF NOT EXISTS environments (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name       VARCHAR(50) NOT NULL CHECK (name IN ('production', 'staging', 'development', 'test')),
    api_key    VARCHAR(80) NOT NULL UNIQUE,
    is_active  BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, name)
);

CREATE INDEX IF NOT EXISTS idx_environments_project_id ON environments(project_id);
CREATE INDEX IF NOT EXISTS idx_environments_api_key    ON environments(api_key);
CREATE INDEX IF NOT EXISTS idx_environments_active     ON environments(project_id, is_active);

COMMENT ON TABLE environments IS 'Ambientes por projeto — BKL-1075. Cada projeto tem prod/staging/dev com API keys independentes.';

-- Adiciona coluna environment_id às flags (nullable para retrocompatibilidade)
ALTER TABLE feature_flags
    ADD COLUMN IF NOT EXISTS environment_id UUID REFERENCES environments(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_flags_environment_id ON feature_flags(environment_id) WHERE environment_id IS NOT NULL;

-- Seed: cria ambiente 'production' para todos os projetos existentes
-- A API key do ambiente production é a mesma API key do projeto (retrocompatibilidade)
INSERT INTO environments (id, project_id, name, api_key, created_at, updated_at)
SELECT
    gen_random_uuid(),
    p.id,
    'production',
    'env_prod_' || encode(gen_random_bytes(24), 'hex'),
    p.created_at,
    NOW()
FROM projects p
ON CONFLICT (project_id, name) DO NOTHING;
