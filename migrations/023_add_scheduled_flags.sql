-- BKL-1040: Adiciona campos starts_at e ends_at para scheduled flags
-- Flags com starts_at/ends_at são ativadas/desativadas automaticamente pelo scheduler.
-- DOWN: ALTER TABLE feature_flags DROP COLUMN IF EXISTS starts_at; ALTER TABLE feature_flags DROP COLUMN IF EXISTS ends_at;

ALTER TABLE feature_flags ADD COLUMN IF NOT EXISTS starts_at TIMESTAMPTZ;
ALTER TABLE feature_flags ADD COLUMN IF NOT EXISTS ends_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_feature_flags_scheduled
    ON feature_flags(starts_at, ends_at)
    WHERE starts_at IS NOT NULL OR ends_at IS NOT NULL;

COMMENT ON COLUMN feature_flags.starts_at IS 'BKL-1040: data/hora de ativação automática (NULL = sem agendamento)';
COMMENT ON COLUMN feature_flags.ends_at   IS 'BKL-1040: data/hora de desativação automática (NULL = sem expiração)';
