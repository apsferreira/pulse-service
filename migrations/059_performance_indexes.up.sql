-- =============================================================================
-- Pulse — Migration 059
-- Índices de performance para monitor_checks e monitors
-- Aplicados diretamente em produção em 2026-05-21. Migration criada para documentação.
-- DOWN: comentado abaixo
-- =============================================================================

-- Índice composto principal para histórico por monitor
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_monitor_checks_monitor_checked
  ON monitor_checks (monitor_id, checked_at DESC);

-- Índices em monitors para queries de monitoramento ativo
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_monitors_active
  ON monitors (is_active) WHERE is_active = true;

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_monitors_project_active
  ON monitors (project_id, is_active);

-- Função de limpeza automática (retenção 90 dias)
CREATE OR REPLACE FUNCTION cleanup_monitor_checks() RETURNS void AS $$
BEGIN
  DELETE FROM monitor_checks
  WHERE checked_at < NOW() - INTERVAL '90 days';
END;
$$ LANGUAGE plpgsql;

-- DOWN: comentado abaixo
-- DROP INDEX CONCURRENTLY IF EXISTS idx_monitor_checks_monitor_checked;
-- DROP INDEX CONCURRENTLY IF EXISTS idx_monitors_active;
-- DROP INDEX CONCURRENTLY IF EXISTS idx_monitors_project_active;
-- DROP FUNCTION IF EXISTS cleanup_monitor_checks();
