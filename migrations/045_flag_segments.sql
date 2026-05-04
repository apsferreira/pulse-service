-- 045_flag_segments.sql
-- BKL-1153: Segments Avançados — targeting por atributo composto para feature flags

CREATE TABLE IF NOT EXISTS flag_segments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    rules JSONB NOT NULL DEFAULT '[]',  -- array de RuleGroup (OR entre grupos; AND dentro de cada grupo)
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_flag_segments_project_id ON flag_segments(project_id);

-- Tabela de associação flag <-> segment
CREATE TABLE IF NOT EXISTS flag_segment_assignments (
    flag_id UUID NOT NULL REFERENCES feature_flags(id) ON DELETE CASCADE,
    segment_id UUID NOT NULL REFERENCES flag_segments(id) ON DELETE CASCADE,
    rollout_percentage INT NOT NULL DEFAULT 100 CHECK (rollout_percentage BETWEEN 0 AND 100),
    PRIMARY KEY (flag_id, segment_id)
);

-- DOWN:
-- DROP TABLE IF EXISTS flag_segment_assignments;
-- DROP TABLE IF EXISTS flag_segments;
