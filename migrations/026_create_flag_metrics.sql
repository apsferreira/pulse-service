-- BKL-1044: Vinculação de flags a métricas de negócio.
-- Permite calcular lift real entre variantes/controle via JOIN com a tabela events.
-- DOWN:
--   DROP TABLE IF EXISTS flag_metrics;

CREATE TABLE IF NOT EXISTS flag_metrics (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    flag_key      TEXT        NOT NULL REFERENCES feature_flags(key) ON DELETE CASCADE ON UPDATE CASCADE,
    project_id    UUID        NOT NULL,
    metric_name   TEXT        NOT NULL,                                -- nome do evento de negócio (ex: checkout_completed)
    description   TEXT        NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(flag_key, project_id, metric_name)
);

CREATE INDEX IF NOT EXISTS idx_flag_metrics_flag_key       ON flag_metrics(flag_key);
CREATE INDEX IF NOT EXISTS idx_flag_metrics_project_id     ON flag_metrics(project_id);
CREATE INDEX IF NOT EXISTS idx_flag_metrics_metric_name    ON flag_metrics(metric_name);

COMMENT ON TABLE  flag_metrics                    IS 'BKL-1044: Vincula uma feature flag a eventos de negócio para cálculo de lift';
COMMENT ON COLUMN flag_metrics.flag_key           IS 'Chave da flag (FK para feature_flags.key)';
COMMENT ON COLUMN flag_metrics.metric_name        IS 'Nome do evento de negócio (deve existir em events.event_type)';
COMMENT ON COLUMN flag_metrics.project_id         IS 'Escopo por projeto (tenant isolation)';
