-- BKL-1084: Histórico de health scores por tenant para detectar 7 dias consecutivos abaixo de 30.
-- Cada execução do cron diário de BKL-801 insere um registro nesta tabela.
CREATE TABLE IF NOT EXISTS tenant_health_score_history (
    id          SERIAL PRIMARY KEY,
    tenant_id   VARCHAR(255) NOT NULL,
    score       INTEGER      NOT NULL CHECK (score >= 0 AND score <= 100),
    recorded_at DATE         NOT NULL DEFAULT CURRENT_DATE,
    UNIQUE (tenant_id, recorded_at)
);

CREATE INDEX IF NOT EXISTS idx_ths_history_tenant_date ON tenant_health_score_history (tenant_id, recorded_at DESC);

COMMENT ON TABLE tenant_health_score_history IS 'BKL-1084: histórico diário de health scores para detecção de churn trend';
