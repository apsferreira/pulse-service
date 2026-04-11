-- BKL-1037: Adiciona campo data_retention_days por projeto para política de retenção de dados.
-- Default por plano:
--   free:    90 dias
--   starter: 180 dias
--   growth:  180 dias
--   pro:     365 dias
-- DOWN: ALTER TABLE projects DROP COLUMN IF EXISTS data_retention_days;

ALTER TABLE projects ADD COLUMN IF NOT EXISTS data_retention_days INTEGER NOT NULL DEFAULT 90;

-- Atualiza projetos existentes baseado no plano
UPDATE projects SET data_retention_days = 180 WHERE COALESCE(plan, 'free') IN ('starter', 'growth');
UPDATE projects SET data_retention_days = 365 WHERE COALESCE(plan, 'free') = 'pro';

COMMENT ON COLUMN projects.data_retention_days IS 'BKL-1037: dias de retenção de eventos (90=free, 180=starter/growth, 365=pro)';
