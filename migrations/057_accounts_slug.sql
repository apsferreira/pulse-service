-- 057_accounts_slug.sql
-- Adiciona slug único às contas para URLs públicas de status page.
--
-- DOWN (comentado):
-- DROP INDEX IF EXISTS idx_accounts_slug;
-- ALTER TABLE accounts DROP COLUMN IF EXISTS slug;

ALTER TABLE accounts ADD COLUMN IF NOT EXISTS slug VARCHAR(80) UNIQUE;

-- Gera slug a partir do email (parte antes do @), sanitizado:
-- apsf88@gmail.com → 'apsf88'
-- antonio.pedro@example.com → 'antonio-pedro'
UPDATE accounts
SET slug = LOWER(REGEXP_REPLACE(SPLIT_PART(email, '@', 1), '[^a-z0-9]', '-', 'g'))
WHERE slug IS NULL;

-- Índice para lookup eficiente por slug (já é UNIQUE, mas explícito para clareza)
CREATE UNIQUE INDEX IF NOT EXISTS idx_accounts_slug ON accounts(slug);
