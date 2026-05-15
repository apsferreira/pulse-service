-- BKL-1117: separar "projeto habilitado" de "SDK conectado"
-- Antes: is_active=false até o SDK chamar — projetos recém-criados rejeitavam chamadas.
-- Depois: is_active controla se o projeto está habilitado (sempre true ao criar);
--         sdk_connected_at registra a primeira chamada real do SDK (null = ainda não conectou).

ALTER TABLE projects
  ADD COLUMN IF NOT EXISTS sdk_connected_at TIMESTAMPTZ;

-- Projetos já ativos que nunca receberam chamada ficam com sdk_connected_at NULL.
-- Projetos que já tinham is_active=true antes desta migration são considerados conectados
-- (foram ativados pelo middleware na primeira chamada do SDK).
UPDATE projects
  SET sdk_connected_at = updated_at
  WHERE is_active = true AND sdk_connected_at IS NULL;

-- Garante que todos os projetos existentes fiquem habilitados.
UPDATE projects SET is_active = true WHERE is_active = false;
