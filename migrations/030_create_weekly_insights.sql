-- BKL-864: Tabela de insights narrativos semanais gerados via Claude Haiku.
-- Armazena o texto acionável + snapshot de métricas para histórico e auditoria.
CREATE TABLE IF NOT EXISTS weekly_insights (
    id          SERIAL PRIMARY KEY,
    week_start  DATE        NOT NULL,
    week_end    DATE        NOT NULL,
    narrative   TEXT        NOT NULL,
    metrics     JSONB       NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (week_start)
);

CREATE INDEX IF NOT EXISTS idx_weekly_insights_week_start ON weekly_insights (week_start DESC);

COMMENT ON TABLE weekly_insights IS 'BKL-864: Insights narrativos semanais gerados por Claude Haiku';
