-- =============================================================================
-- Pulse — Migration 049
-- Auditoria de mudanças de plano por projeto
-- Registra toda troca de plano, origem (checkout_order_id) e responsável
-- DOWN: comentado abaixo
-- =============================================================================

CREATE TABLE IF NOT EXISTS plan_changes (
    id                 UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id         UUID         NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    from_plan          VARCHAR(20)  NOT NULL,
    to_plan            VARCHAR(20)  NOT NULL,
    changed_by         VARCHAR(255),
    reason             VARCHAR(500),
    checkout_order_id  VARCHAR(255),
    created_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_plan_changes_project ON plan_changes(project_id);
CREATE INDEX idx_plan_changes_created ON plan_changes(created_at);
