-- BKL-1154: Multi-Armed Bandits — tabelas de estado MAB por experiment

CREATE TABLE IF NOT EXISTS mab_experiments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL,
    flag_key VARCHAR(255) NOT NULL,
    epsilon NUMERIC(4,3) NOT NULL DEFAULT 0.1,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(project_id, flag_key)
);

CREATE TABLE IF NOT EXISTS mab_variant_stats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    experiment_id UUID NOT NULL REFERENCES mab_experiments(id) ON DELETE CASCADE,
    variant_key VARCHAR(255) NOT NULL,
    impressions BIGINT NOT NULL DEFAULT 0,
    conversions BIGINT NOT NULL DEFAULT 0,
    weight NUMERIC(8,6) NOT NULL DEFAULT 0.5,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(experiment_id, variant_key)
);

CREATE INDEX IF NOT EXISTS idx_mab_experiments_project_flag ON mab_experiments(project_id, flag_key);
CREATE INDEX IF NOT EXISTS idx_mab_variant_stats_experiment ON mab_variant_stats(experiment_id);

-- DOWN
-- DROP INDEX IF EXISTS idx_mab_variant_stats_experiment;
-- DROP INDEX IF EXISTS idx_mab_experiments_project_flag;
-- DROP TABLE IF EXISTS mab_variant_stats;
-- DROP TABLE IF EXISTS mab_experiments;
