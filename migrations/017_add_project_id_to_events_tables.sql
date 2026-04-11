-- BKL-1038: adiciona project_id nas tabelas de eventos para isolamento multi-tenant
-- Sem project_id, eventos de diferentes tenants ficam misturados — cross-tenant data leak.
-- Backfill: NULL para registros existentes (dados históricos sem projeto associado).
-- DOWN:
--   ALTER TABLE events DROP COLUMN IF EXISTS project_id;
--   ALTER TABLE api_costs DROP COLUMN IF EXISTS project_id;
--   ALTER TABLE nps_responses DROP COLUMN IF EXISTS project_id;
--   ALTER TABLE monitors DROP COLUMN IF EXISTS project_id;

-- events
ALTER TABLE events ADD COLUMN IF NOT EXISTS project_id UUID REFERENCES projects(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_events_project_id ON events(project_id);

-- api_costs
ALTER TABLE api_costs ADD COLUMN IF NOT EXISTS project_id UUID REFERENCES projects(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_api_costs_project_id ON api_costs(project_id);

-- nps_responses
ALTER TABLE nps_responses ADD COLUMN IF NOT EXISTS project_id UUID REFERENCES projects(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_nps_responses_project_id ON nps_responses(project_id);

-- monitors
ALTER TABLE monitors ADD COLUMN IF NOT EXISTS project_id UUID REFERENCES projects(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_monitors_project_id ON monitors(project_id);
