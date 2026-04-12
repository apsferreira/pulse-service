-- BKL-998: Health score configurável por projeto.
-- Adiciona coluna JSONB em projects para armazenar pesos e thresholds do health score.
-- DOWN:
--   ALTER TABLE projects DROP COLUMN IF EXISTS health_config;

ALTER TABLE projects
    ADD COLUMN IF NOT EXISTS health_config JSONB NOT NULL DEFAULT '{
        "weights": {
            "flags_with_audit": 0.3,
            "experiments_running": 0.3,
            "events_tracked_24h": 0.2,
            "error_rate": 0.2
        },
        "thresholds": {
            "green": 80,
            "yellow": 60
        }
    }'::jsonb;

COMMENT ON COLUMN projects.health_config IS 'BKL-998: Configuração de pesos e thresholds para cálculo de health score por projeto';
