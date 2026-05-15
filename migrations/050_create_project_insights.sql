-- BKL-insights: Tabela de insights acionáveis gerados pelo agente autônomo de análise.
-- Armazena sugestões sobre feature flags (dormentes, 100% rollout, spike/drop, etc.)
-- com suporte a dismiss por usuário e expiração automática em 7 dias.

-- up
CREATE TABLE IF NOT EXISTS project_insights (
    id           UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id   UUID         NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    insight_type VARCHAR(50)  NOT NULL,     -- dormant_flag | full_rollout | spike | drop | negative_rate | ab_suggestion
    flag_key     VARCHAR(255),
    message      TEXT         NOT NULL,
    severity     VARCHAR(10)  NOT NULL DEFAULT 'info', -- info | warning | critical
    metadata     JSONB,
    is_dismissed BOOLEAN      NOT NULL DEFAULT FALSE,
    dismissed_at TIMESTAMPTZ,
    dismissed_by VARCHAR(255),
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    expires_at   TIMESTAMPTZ
);

CREATE INDEX idx_project_insights_project ON project_insights(project_id, is_dismissed);
CREATE INDEX idx_project_insights_created ON project_insights(created_at);

COMMENT ON TABLE project_insights IS 'BKL-insights: Insights acionáveis gerados pelo agente autônomo de análise de feature flags e eventos';
COMMENT ON COLUMN project_insights.insight_type IS 'dormant_flag | full_rollout | spike | drop | negative_rate | ab_suggestion';
COMMENT ON COLUMN project_insights.severity IS 'info | warning | critical';

-- down
-- DROP TABLE IF EXISTS project_insights;
