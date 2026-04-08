-- Opcao A: rollout percentual (0-100) para liberação gradual
-- Opcao B: targeting por lista de user_ids ou tenant_ids
ALTER TABLE feature_flags ADD COLUMN IF NOT EXISTS rollout_percentage INT NOT NULL DEFAULT 0;
ALTER TABLE feature_flags ADD COLUMN IF NOT EXISTS targeting JSONB NOT NULL DEFAULT '{}';

-- Exemplos de targeting:
-- {"user_ids": ["uuid1", "uuid2"]}           → só esses usuários veem
-- {"tenant_ids": ["tenant-slug-1"]}           → só esses tenants veem
-- {} (vazio)                                   → avalia pelo rollout_percentage
--
-- Lógica de avaliação (em ordem):
-- 1. Se targeting tem user_ids/tenant_ids e o contexto bate → flag = true
-- 2. Se rollout_percentage > 0 → hash(user_id + flag_key) % 100 < rollout_percentage
-- 3. Se enabled = true e rollout_percentage = 0 e targeting vazio → flag = true (comportamento legado)
