-- BKL-1064: AI Configs — Feature flags para prompts e parâmetros de LLM.
-- Permite deploy de prompt sem redeploy de código. Rollback instantâneo.
-- Diferencial único do Pulse para times com produtos de IA.
CREATE TABLE IF NOT EXISTS ai_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    key VARCHAR(255) NOT NULL,              -- ex: "checkout_ai.system_prompt"
    description TEXT,
    -- Parâmetros de modelo
    model VARCHAR(100) NOT NULL DEFAULT 'claude-haiku-4-5-20251001',
    temperature NUMERIC(3,2) NOT NULL DEFAULT 0.7 CHECK (temperature >= 0 AND temperature <= 2),
    max_tokens INT NOT NULL DEFAULT 1024,
    -- Prompt
    system_prompt TEXT,
    user_prompt_template TEXT,             -- template com variáveis {{var}}
    -- Controle
    is_active BOOLEAN NOT NULL DEFAULT true,
    version INT NOT NULL DEFAULT 1,
    created_by VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, key)
);

CREATE INDEX IF NOT EXISTS idx_ai_configs_project_key ON ai_configs(project_id, key) WHERE is_active = true;

-- Histórico de versões para rollback
CREATE TABLE IF NOT EXISTS ai_config_versions (
    id SERIAL PRIMARY KEY,
    ai_config_id UUID NOT NULL REFERENCES ai_configs(id) ON DELETE CASCADE,
    version INT NOT NULL,
    model VARCHAR(100) NOT NULL,
    temperature NUMERIC(3,2) NOT NULL,
    max_tokens INT NOT NULL,
    system_prompt TEXT,
    user_prompt_template TEXT,
    changed_by VARCHAR(255),
    changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ai_config_versions_config ON ai_config_versions(ai_config_id, version DESC);
