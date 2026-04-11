-- BKL-1071: Tabela de overage para cobrança mensal de eventos acima da quota.
-- Ao final de cada mês, o billing job varre esta tabela e gera cobranças via checkout-service/Asaas.
-- Preço: R$0,02 por 1.000 eventos acima da quota do plano.

CREATE TABLE IF NOT EXISTS overage_billing (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL,
    period_year     INT  NOT NULL,   -- ex: 2026
    period_month    INT  NOT NULL,   -- ex: 4 (abril)
    plan            VARCHAR(20) NOT NULL,
    quota_limit     BIGINT NOT NULL, -- limite do plano em eventos
    events_consumed BIGINT NOT NULL, -- total de eventos no mês
    overage_events  BIGINT NOT NULL, -- eventos acima da quota (events_consumed - quota_limit)
    price_per_1k    NUMERIC(10,4) NOT NULL DEFAULT 0.02, -- R$ por 1.000 eventos
    amount_brl      NUMERIC(10,2) NOT NULL,              -- total a cobrar
    status          VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending | invoiced | paid | waived
    external_charge_id VARCHAR(100),  -- ID da cobrança no Asaas
    invoiced_at     TIMESTAMPTZ,
    paid_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT overage_billing_project_period_uniq UNIQUE (project_id, period_year, period_month)
);

CREATE INDEX IF NOT EXISTS idx_overage_billing_status ON overage_billing(status);
CREATE INDEX IF NOT EXISTS idx_overage_billing_project ON overage_billing(project_id);
CREATE INDEX IF NOT EXISTS idx_overage_billing_period  ON overage_billing(period_year, period_month);
