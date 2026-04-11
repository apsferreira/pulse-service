-- BKL-1077: Adiciona coluna plan aos projetos para quota enforcement por plano.
-- Planos válidos: free (10k/mês), starter (500k/mês), growth (2M/mês), pro (ilimitado).

ALTER TABLE projects
    ADD COLUMN IF NOT EXISTS plan VARCHAR(20) NOT NULL DEFAULT 'free'
        CHECK (plan IN ('free', 'starter', 'growth', 'pro'));

CREATE INDEX IF NOT EXISTS idx_projects_plan ON projects(plan);

COMMENT ON COLUMN projects.plan IS 'Plano de assinatura: free=10k/mês, starter=500k/mês, growth=2M/mês, pro=ilimitado. BKL-1077.';
