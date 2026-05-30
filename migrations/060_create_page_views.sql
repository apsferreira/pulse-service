-- =============================================================================
-- Pulse — Migration 060
-- PULSE-PV-02: Tabela de page views para rastreamento automático de páginas
-- DOWN: comentado abaixo
-- =============================================================================

CREATE TABLE IF NOT EXISTS page_views (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  session_id      VARCHAR(100),
  user_id         UUID,
  product         VARCHAR(50),
  page_title      VARCHAR(255),
  page_path       VARCHAR(500),
  time_on_page_ms INT,
  referrer        VARCHAR(500),
  created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_page_views_user_product
  ON page_views (user_id, product, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_page_views_session
  ON page_views (session_id, created_at);

CREATE INDEX IF NOT EXISTS idx_page_views_product_path
  ON page_views (product, page_path, created_at DESC);

-- DOWN:
-- DROP INDEX IF EXISTS idx_page_views_product_path;
-- DROP INDEX IF EXISTS idx_page_views_session;
-- DROP INDEX IF EXISTS idx_page_views_user_product;
-- DROP TABLE IF EXISTS page_views;
