-- BKL-1074: Audit log de avaliações de feature flags.
-- Registra cada avaliação: quando, quem, qual flag, qual resultado, qual targeting se aplicou.

CREATE TABLE IF NOT EXISTS flag_evaluations (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    flag_key     VARCHAR(100) NOT NULL,
    project_id   UUID REFERENCES projects(id) ON DELETE SET NULL,
    user_id      VARCHAR(255),
    tenant_id    VARCHAR(255),
    result       BOOLEAN NOT NULL,
    variant      VARCHAR(100),
    reason       VARCHAR(50) NOT NULL DEFAULT 'default',
    -- reason: 'kill_switch'|'targeting_user'|'targeting_tenant'|'rollout'|'default'|'disabled'
    evaluated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_flag_evals_flag_key       ON flag_evaluations(flag_key, evaluated_at DESC);
CREATE INDEX IF NOT EXISTS idx_flag_evals_project_id     ON flag_evaluations(project_id, evaluated_at DESC);
CREATE INDEX IF NOT EXISTS idx_flag_evals_user_id        ON flag_evaluations(user_id) WHERE user_id IS NOT NULL;

COMMENT ON TABLE flag_evaluations IS 'Audit log de avaliações de flags — BKL-1074.';
COMMENT ON COLUMN flag_evaluations.reason IS 'Motivo da decisão: kill_switch, targeting_user, targeting_tenant, rollout, default, disabled.';
