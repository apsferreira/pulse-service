-- Migration 008: Adiciona suporte a variantes para A/B testing (BKL-094)
-- variants é um array JSON de objetos {key, weight} onde weight é percentual (0-100).
-- Exemplo: [{"key": "control", "weight": 50}, {"key": "challenger", "weight": 50}]
-- Se variants for vazio ou null, a flag é tratada como on/off normal.

ALTER TABLE feature_flags ADD COLUMN IF NOT EXISTS variants JSONB DEFAULT '[]'::jsonb;

COMMENT ON COLUMN feature_flags.variants IS 'Variantes A/B: [{key, weight}]. Vazio = comportamento on/off normal.';
