-- BKL-999: Tabela para armazenar modelos de churn treinados por projeto.
-- Cada projeto pode ter um modelo de regressão logística treinado com seus próprios dados.
-- Os pesos são armazenados como JSONB para suportar vetor de features arbitrário.

CREATE TABLE IF NOT EXISTS project_churn_models (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id  UUID NOT NULL UNIQUE REFERENCES projects(id) ON DELETE CASCADE,
    weights     JSONB NOT NULL DEFAULT '[]',   -- vetor de pesos [bias, w1, ..., wn]
    feature_names JSONB NOT NULL DEFAULT '[]', -- nomes das features na ordem dos pesos
    trained_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    accuracy    FLOAT NOT NULL DEFAULT 0.0,    -- acurácia no conjunto de treino (0.0-1.0)
    sample_size INT NOT NULL DEFAULT 0,        -- número de usuários usados no treino
    churned_count INT NOT NULL DEFAULT 0,      -- usuários com label churn=1
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_project_churn_models_project_id ON project_churn_models(project_id);

-- DOWN:
-- DROP INDEX IF EXISTS idx_project_churn_models_project_id;
-- DROP TABLE IF EXISTS project_churn_models;
