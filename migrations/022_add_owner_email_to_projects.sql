-- BKL-1082: Adiciona owner_email a projects para email sequence de onboarding.
-- Email do responsável pelo projeto — usado como destinatário dos emails de onboarding.

ALTER TABLE projects
    ADD COLUMN IF NOT EXISTS owner_email VARCHAR(255);

CREATE INDEX IF NOT EXISTS idx_projects_owner_email ON projects(owner_email) WHERE owner_email IS NOT NULL;

COMMENT ON COLUMN projects.owner_email IS 'Email do responsável pelo projeto — destinatário das emails de onboarding. BKL-1082.';
