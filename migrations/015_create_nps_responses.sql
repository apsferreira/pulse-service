-- BKL-811: Tabela para armazenar respostas de NPS coletadas pelos produtos.
-- Escopo: SocialMake no primeiro lançamento, extensível para outros produtos.

CREATE TABLE IF NOT EXISTS nps_responses (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   TEXT NOT NULL,
    product     TEXT NOT NULL,          -- 'socialmake', 'nitro', 'libri', etc.
    score       INT NOT NULL CHECK (score >= 0 AND score <= 10),
    comment     TEXT,
    triggered   TEXT NOT NULL DEFAULT 'after_30_days', -- trigger que gerou o survey
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_nps_responses_tenant ON nps_responses(tenant_id);
CREATE INDEX idx_nps_responses_product ON nps_responses(product);
CREATE INDEX idx_nps_responses_score ON nps_responses(score);
CREATE INDEX idx_nps_responses_created_at ON nps_responses(created_at);

COMMENT ON TABLE nps_responses IS 'Respostas de NPS (0-10) coletadas pelos produtos IIT. BKL-811.';
