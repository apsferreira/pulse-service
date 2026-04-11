-- BKL-1063: LLM Observability expandida — métricas por modelo, latência, error rate.
-- Diferencial único do Pulse vs concorrentes (Amplitude, PostHog não têm isso).

-- Adiciona colunas de latência e erro na tabela existente api_costs
ALTER TABLE api_costs
    ADD COLUMN IF NOT EXISTS latency_ms INT DEFAULT NULL,        -- latência da chamada LLM em ms
    ADD COLUMN IF NOT EXISTS is_error BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS error_type VARCHAR(100) DEFAULT NULL, -- ex: "rate_limit", "context_length", "timeout"
    ADD COLUMN IF NOT EXISTS feature VARCHAR(255) DEFAULT NULL;    -- feature que usou o LLM (ex: "caption_generator")

-- Índice para queries de observability por modelo/provider/período
CREATE INDEX IF NOT EXISTS idx_api_costs_provider_model ON api_costs(provider, model, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_api_costs_feature ON api_costs(feature, created_at DESC) WHERE feature IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_api_costs_error ON api_costs(is_error, created_at DESC) WHERE is_error = true;
