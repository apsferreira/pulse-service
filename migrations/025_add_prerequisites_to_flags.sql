-- BKL-1039: Adiciona campo prerequisites às feature flags.
-- Uma flag com prerequisites só é avaliada como ativa se todas as flags
-- da lista estiverem ativas (e com a variante correta, se especificada).
-- DOWN: ALTER TABLE feature_flags DROP COLUMN IF EXISTS prerequisites;

ALTER TABLE feature_flags ADD COLUMN IF NOT EXISTS prerequisites JSONB NOT NULL DEFAULT '[]'::jsonb;

COMMENT ON COLUMN feature_flags.prerequisites IS 'BKL-1039: array de [{flag_key, required_variant}] — dependências entre flags';
