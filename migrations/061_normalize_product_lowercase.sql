-- =============================================================================
-- Pulse — Migration 061
-- fix(#64): Normalização de `product` em dados históricos
--
-- Filtros SQL consultam `WHERE product = 'nitro'` (lowercase) mas SDKs podem
-- ter persistido 'Nitro', ' nitro ', etc. — métricas e sparklines retornavam
-- zero silenciosamente. A partir de fix(#64) a normalização ocorre na
-- ingestão; esta migration reconcilia o histórico.
--
-- DOWN: não há rollback — a perda de capitalização original é irreversível.
--       Caso necessário restaurar, usar backup pré-migration.
-- =============================================================================

BEGIN;

-- events: campo `product` é a chave de filtragem em todas as queries de métrica
UPDATE events
SET product = LOWER(TRIM(product))
WHERE product IS NOT NULL
  AND product <> LOWER(TRIM(product));

-- page_views: mesmo problema, mesma normalização
UPDATE page_views
SET product = LOWER(TRIM(product))
WHERE product IS NOT NULL
  AND product <> LOWER(TRIM(product));

COMMIT;
