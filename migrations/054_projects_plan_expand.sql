-- BKL-billing: expandir planos aceitos + atualizar IIT para enterprise
-- O check constraint original só tinha free|starter|growth|pro.
-- Adicionamos enterprise e internal para suportar contratos especiais e projetos internos.

ALTER TABLE projects
  DROP CONSTRAINT IF EXISTS projects_plan_check;

ALTER TABLE projects
  ADD CONSTRAINT projects_plan_check
    CHECK (plan IN ('free', 'starter', 'growth', 'pro', 'enterprise', 'internal'));

-- Instituto Itinerante Site → enterprise (acesso irrestrito, projeto interno do IIT)
UPDATE projects
  SET plan = 'enterprise', data_retention_days = 365
  WHERE slug = 'iit-site';
