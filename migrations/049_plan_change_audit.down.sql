-- =============================================================================
-- Pulse — Migration 049 DOWN
-- Remove tabela de auditoria de mudanças de plano
-- =============================================================================

-- DROP INDEX IF EXISTS idx_plan_changes_created;
-- DROP INDEX IF EXISTS idx_plan_changes_project;
-- DROP TABLE IF EXISTS plan_changes;
