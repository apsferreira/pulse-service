DROP INDEX CONCURRENTLY IF EXISTS idx_monitor_checks_monitor_checked;
DROP INDEX CONCURRENTLY IF EXISTS idx_monitors_active;
DROP INDEX CONCURRENTLY IF EXISTS idx_monitors_project_active;
DROP FUNCTION IF EXISTS cleanup_monitor_checks();
