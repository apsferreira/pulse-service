-- 056_accounts_plan.sql
-- Move o campo plan de projects para uma tabela accounts por owner_email.
-- Todos os projetos de uma conta compartilham o mesmo plano.
--
-- DOWN (comentado):
-- ALTER TABLE projects DROP CONSTRAINT IF EXISTS fk_projects_account;
-- ALTER TABLE projects ALTER COLUMN owner_email DROP NOT NULL;
-- DROP INDEX IF EXISTS idx_accounts_plan;
-- DROP TABLE IF EXISTS accounts;

-- Cria tabela accounts
CREATE TABLE accounts (
    email VARCHAR(255) PRIMARY KEY,
    plan VARCHAR(20) NOT NULL DEFAULT 'free'
        CHECK (plan IN ('free','starter','growth','pro','enterprise','internal')),
    data_retention_days INT NOT NULL DEFAULT 90,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Índice útil para lookup de plan
CREATE INDEX idx_accounts_plan ON accounts(plan);

-- Popula accounts a partir dos projetos existentes que têm owner_email.
-- Usa o plano de maior rank de cada email (não rebaixa quem tem projetos mistos).
INSERT INTO accounts (email, plan, data_retention_days)
SELECT
    owner_email,
    (ARRAY['free','starter','growth','pro','enterprise','internal'])[
        MAX(
            CASE plan
                WHEN 'free'       THEN 1
                WHEN 'starter'    THEN 2
                WHEN 'growth'     THEN 3
                WHEN 'pro'        THEN 4
                WHEN 'enterprise' THEN 5
                WHEN 'internal'   THEN 6
                ELSE 1
            END
        )
    ],
    MAX(COALESCE(data_retention_days, 90))
FROM projects
WHERE owner_email IS NOT NULL
GROUP BY owner_email
ON CONFLICT (email) DO NOTHING;

-- Garante que apsf88@gmail.com tem plano enterprise (dono de todos os projetos IIT).
INSERT INTO accounts (email, plan, data_retention_days)
VALUES ('apsf88@gmail.com', 'enterprise', 365)
ON CONFLICT (email) DO UPDATE SET plan = 'enterprise', data_retention_days = 365, updated_at = NOW();

-- Atribui owner_email = 'apsf88@gmail.com' nos projetos que ainda não têm.
UPDATE projects SET owner_email = 'apsf88@gmail.com' WHERE owner_email IS NULL;

-- Adiciona FK de projects → accounts e torna owner_email NOT NULL.
ALTER TABLE projects
    ALTER COLUMN owner_email SET NOT NULL,
    ADD CONSTRAINT fk_projects_account
        FOREIGN KEY (owner_email) REFERENCES accounts(email) ON UPDATE CASCADE;
