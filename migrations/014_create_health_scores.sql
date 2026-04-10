-- Migration 014: tabela de health scores por tenant (BKL-801)
-- Score 0-100 calculado diariamente com sinais de engajamento.

CREATE TABLE IF NOT EXISTS tenant_health_scores (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id        VARCHAR(255) NOT NULL UNIQUE,
    score            INTEGER NOT NULL DEFAULT 0 CHECK (score >= 0 AND score <= 100),
    login_score      INTEGER NOT NULL DEFAULT 0 CHECK (login_score >= 0 AND login_score <= 30),
    content_score    INTEGER NOT NULL DEFAULT 0 CHECK (content_score >= 0 AND content_score <= 40),
    adoption_score   INTEGER NOT NULL DEFAULT 0 CHECK (adoption_score >= 0 AND adoption_score <= 20),
    subscription_score INTEGER NOT NULL DEFAULT 0 CHECK (subscription_score >= 0 AND subscription_score <= 10),
    last_login_at    TIMESTAMPTZ,
    last_content_at  TIMESTAMPTZ,
    calculated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_health_scores_tenant ON tenant_health_scores(tenant_id);
CREATE INDEX IF NOT EXISTS idx_health_scores_score ON tenant_health_scores(score);

COMMENT ON TABLE tenant_health_scores IS 'Health score 0-100 por tenant para churn prevention (BKL-801)';
COMMENT ON COLUMN tenant_health_scores.login_score IS 'Frequência de login: até 30 pontos';
COMMENT ON COLUMN tenant_health_scores.content_score IS 'Criação de conteúdo: até 40 pontos';
COMMENT ON COLUMN tenant_health_scores.adoption_score IS 'Adoção de features: até 20 pontos';
COMMENT ON COLUMN tenant_health_scores.subscription_score IS 'Status da assinatura: até 10 pontos';
