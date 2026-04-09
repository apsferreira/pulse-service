-- Migration 013: adiciona coluna environment na tabela events para separar dados de dev/prod
-- DOWN: ALTER TABLE events DROP COLUMN IF EXISTS environment;

ALTER TABLE events ADD COLUMN IF NOT EXISTS environment VARCHAR(20) NOT NULL DEFAULT 'production';
CREATE INDEX IF NOT EXISTS idx_events_environment ON events(environment);
