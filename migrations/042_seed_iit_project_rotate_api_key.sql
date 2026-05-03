-- Migration 042: Seed do projeto IIT e rotação da API key hardcoded (P0 security fix)
-- Contexto: migration 009 continha API key com valor fixo/placeholder sem entropia real.
-- Esta migration garante que ambientes novos e produção usem uma key gerada aleatoriamente.

-- Caso 1: Produção — rotaciona a key placeholder para um valor aleatório seguro
UPDATE projects
SET
  api_key    = 'pulse_' || encode(gen_random_bytes(32), 'hex'),
  updated_at = NOW()
WHERE slug = 'iit'
  AND api_key = 'pulse_iit_default_key_00000000000000000000000000000000';

-- Caso 2: Ambientes novos (sem a row do projeto IIT) — insere com key aleatória
INSERT INTO projects (id, name, slug, api_key, description)
SELECT
  'a0000000-0000-0000-0000-000000000001',
  'Instituto Itinerante',
  'iit',
  'pulse_' || encode(gen_random_bytes(32), 'hex'),
  'Projeto padrão IIT — flags globais'
WHERE NOT EXISTS (SELECT 1 FROM projects WHERE slug = 'iit');

-- Após aplicar esta migration, recupere a nova API key com:
--   SELECT api_key FROM projects WHERE slug = 'iit';
-- e atualize a variável de ambiente PULSE_IIT_API_KEY nos deployments.

-- DOWN (não reversível — não restaurar key insegura):
-- UPDATE projects SET api_key = 'REDACTED', updated_at = NOW() WHERE slug = 'iit';
