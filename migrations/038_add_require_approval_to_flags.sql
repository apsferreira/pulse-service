-- BKL-990: require_approval já foi adicionado em 027 como requires_approval.
-- Esta migration adiciona o campo environment e expande o fluxo de change_requests
-- com suporte a old_value/new_value distintos do payload genérico da 027.
-- Renomeia requires_approval → require_approval para consistência com a API.
-- DOWN:
--   ALTER TABLE feature_flags RENAME COLUMN require_approval TO requires_approval;

ALTER TABLE feature_flags
    RENAME COLUMN requires_approval TO require_approval;

COMMENT ON COLUMN feature_flags.require_approval IS 'BKL-990: Quando true, toda mudança via PUT /flags/:key gera um change_request ao invés de aplicar direto';
