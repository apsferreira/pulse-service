-- 010_create_api_costs.sql
-- Rastreamento de custos de API (LLM, imagens, etc.) por produto/operacao
-- UP
CREATE TABLE IF NOT EXISTS api_costs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product     VARCHAR(100) NOT NULL,
    operation   VARCHAR(100) NOT NULL,
    provider    VARCHAR(50)  NOT NULL,
    model       VARCHAR(100),
    tokens_in   INT NOT NULL DEFAULT 0,
    tokens_out  INT NOT NULL DEFAULT 0,
    images      INT NOT NULL DEFAULT 0,
    cost_usd    NUMERIC(12, 8) NOT NULL DEFAULT 0,
    metadata    JSONB DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_api_costs_product   ON api_costs (product);
CREATE INDEX IF NOT EXISTS idx_api_costs_created   ON api_costs (created_at);
CREATE INDEX IF NOT EXISTS idx_api_costs_operation ON api_costs (operation);

-- DOWN
-- DROP INDEX IF EXISTS idx_api_costs_operation;
-- DROP INDEX IF EXISTS idx_api_costs_created;
-- DROP INDEX IF EXISTS idx_api_costs_product;
-- DROP TABLE IF EXISTS api_costs;
