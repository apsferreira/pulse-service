-- =============================================================================
-- Pulse — Migration 062
-- ADR-052 (SocialMake, Opção B): índice para filtro por properties->>'attribution_ref'
-- em POST /api/v1/projects/:id/query (query_explorer_handler.go).
--
-- O ADR-052 aponta "não existe índice GIN em events.properties" como pré-requisito
-- de performance. Um GIN genérico sobre a coluna properties (jsonb_ops) NÃO acelera
-- o padrão de acesso real usado aqui — filtro de IGUALDADE sobre um valor extraído
-- via ->>'key' (ex: properties->>'attribution_ref' = $1) não é um dos operadores
-- que GIN(jsonb) indexa (@>, ?, ?&, ?|). O que acelera esse padrão é um índice de
-- expressão B-tree sobre o valor já extraído — por isso a escolha abaixo diverge do
-- texto do ADR na ferramenta (B-tree de expressão, não GIN), mantendo a mesma
-- intenção (pré-requisito de performance do Opção B).
--
-- Nota de volume: hoje project_id (idx_events_project_id, migration 017) e
-- event_type (idx_events_event_type_created_at, migration 005) já restringem bem
-- antes de qualquer filtro sobre properties — o volume por loja do StoreMake não
-- exige isso ainda. O índice abaixo é para quando isso deixar de ser verdade.
--
-- DOWN: comentado abaixo
-- =============================================================================

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_events_properties_attribution_ref
  ON events ((properties ->> 'attribution_ref'))
  WHERE properties ->> 'attribution_ref' IS NOT NULL;

-- DOWN:
-- DROP INDEX CONCURRENTLY IF EXISTS idx_events_properties_attribution_ref;
