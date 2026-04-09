-- UP
ALTER TABLE feature_flags ADD COLUMN IF NOT EXISTS project_id UUID;
CREATE INDEX IF NOT EXISTS idx_feature_flags_project ON feature_flags (project_id);

-- DOWN
DROP INDEX IF EXISTS idx_feature_flags_project;
ALTER TABLE feature_flags DROP COLUMN IF EXISTS project_id;
