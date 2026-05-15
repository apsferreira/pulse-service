-- 058_seed_environments_missing.sql
-- Cria ambiente 'production' para projetos que não têm nenhum ambiente cadastrado.
-- Projetos criados após a migration 019 não receberam o seed inicial.

INSERT INTO environments (id, project_id, name, api_key, created_at, updated_at)
SELECT
    gen_random_uuid(),
    p.id,
    'production',
    'env_prod_' || encode(gen_random_bytes(24), 'hex'),
    p.created_at,
    NOW()
FROM projects p
WHERE p.id NOT IN (SELECT DISTINCT project_id FROM environments)
ON CONFLICT (project_id, name) DO NOTHING;
