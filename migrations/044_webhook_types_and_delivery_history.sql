-- Migration 044: Tipos de webhook (slack/datadog/custom) e colunas de histórico de entregas.
-- Estende BKL-1078 (migration 033) para suportar formatação específica por destino
-- e métricas de latência no log de deliveries.
-- Nota: o número 043 está reservado a um placeholder (ver 043_renumbered_to_044.sql);
-- esta migration foi originalmente rascunhada como "043" e renumerada para 044.

-- Tipo de webhook — controla formatação do payload no dispatch
ALTER TABLE webhooks
    ADD COLUMN IF NOT EXISTS webhook_type VARCHAR(50) NOT NULL DEFAULT 'custom';

-- Índice para filtrar por tipo (útil para dashboards/stats)
CREATE INDEX IF NOT EXISTS idx_webhooks_type ON webhooks(webhook_type);

-- Colunas de telemetria nas deliveries
ALTER TABLE webhook_deliveries
    ADD COLUMN IF NOT EXISTS success BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS duration_ms INT;

-- Retroativamente marca como sucesso as entregas com status_code 2xx
UPDATE webhook_deliveries
SET success = true
WHERE status_code >= 200 AND status_code < 300;

-- Índice para consulta rápida de histórico por webhook + recente
CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_success ON webhook_deliveries(webhook_id, success, delivered_at DESC);

-- DOWN:
-- ALTER TABLE webhooks DROP COLUMN IF EXISTS webhook_type;
-- DROP INDEX IF EXISTS idx_webhooks_type;
-- ALTER TABLE webhook_deliveries DROP COLUMN IF EXISTS success;
-- ALTER TABLE webhook_deliveries DROP COLUMN IF EXISTS duration_ms;
-- DROP INDEX IF EXISTS idx_webhook_deliveries_success;
