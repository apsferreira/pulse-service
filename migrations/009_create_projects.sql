-- Migration 009: Tabela de projetos para scoping de flags e eventos (BKL-099)
-- Cada projeto tem uma API key única para autenticação do SDK.

CREATE TABLE IF NOT EXISTS projects (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(200) NOT NULL,
  slug VARCHAR(100) NOT NULL UNIQUE,
  api_key VARCHAR(64) NOT NULL UNIQUE,  -- gerada automaticamente (hex 32 bytes)
  description TEXT,
  is_active BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_projects_api_key ON projects(api_key);
CREATE INDEX idx_projects_slug ON projects(slug);

-- Adiciona project_id opcional em feature_flags (NULL = flag global, sem scoping)
ALTER TABLE feature_flags ADD COLUMN IF NOT EXISTS project_id UUID REFERENCES projects(id) ON DELETE SET NULL;
CREATE INDEX idx_feature_flags_project_id ON feature_flags(project_id);

COMMENT ON TABLE projects IS 'Projetos que consomem o Pulse SDK. Cada projeto tem uma API key única.';
COMMENT ON COLUMN projects.api_key IS 'API key do SDK: pulse_<hex32>. Enviada no header X-Pulse-Key.';
COMMENT ON COLUMN feature_flags.project_id IS 'NULL = flag global (sem scoping). Quando definido, flag pertence ao projeto.';

-- Seed do projeto padrão "iit" movido para migration 042
-- (API key gerada dinamicamente via gen_random_bytes — sem valor hardcoded)
