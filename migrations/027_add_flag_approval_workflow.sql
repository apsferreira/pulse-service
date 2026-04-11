-- BKL-1079: Approval workflow para feature flags.
-- Flags em produção podem requerer aprovação de um segundo revisor antes de ativar.
-- Fluxo: draft → pending_approval → approved/rejected
-- DOWN:
--   DROP TABLE IF EXISTS flag_approvals;
--   ALTER TABLE feature_flags DROP COLUMN IF EXISTS approval_status;
--   ALTER TABLE feature_flags DROP COLUMN IF EXISTS requires_approval;
--   DROP TYPE IF EXISTS flag_approval_status;

CREATE TYPE flag_approval_status AS ENUM ('draft', 'pending_approval', 'approved', 'rejected');

-- Campo de status de aprovação na flag
ALTER TABLE feature_flags
    ADD COLUMN IF NOT EXISTS requires_approval BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS approval_status   flag_approval_status NOT NULL DEFAULT 'approved';

-- Tabela de solicitações de aprovação
CREATE TABLE IF NOT EXISTS flag_approvals (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    flag_key     TEXT        NOT NULL REFERENCES feature_flags(key) ON DELETE CASCADE ON UPDATE CASCADE,
    project_id   UUID        NOT NULL,
    requested_by TEXT        NOT NULL,                       -- email ou user_id do solicitante
    reviewed_by  TEXT,                                       -- email ou user_id do revisor
    status       flag_approval_status NOT NULL DEFAULT 'pending_approval',
    notes        TEXT        NOT NULL DEFAULT '',             -- observações do revisor
    payload      JSONB       NOT NULL DEFAULT '{}',           -- snapshot do UpdateFlagRequest aprovado
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reviewed_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_flag_approvals_flag_key   ON flag_approvals(flag_key);
CREATE INDEX IF NOT EXISTS idx_flag_approvals_project_id ON flag_approvals(project_id);
CREATE INDEX IF NOT EXISTS idx_flag_approvals_status     ON flag_approvals(status) WHERE status = 'pending_approval';

COMMENT ON TABLE  flag_approvals IS 'BKL-1079: Workflow de aprovação para ativação de feature flags em produção';
COMMENT ON COLUMN flag_approvals.payload IS 'Snapshot do UpdateFlagRequest que será aplicado após aprovação';
