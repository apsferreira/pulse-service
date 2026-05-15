-- BKL-1116: user_id na tabela events era UUID, mas o SDK envia strings livres
-- como "anon_x7k2p9..." (visitantes sem login) ou IDs de outros sistemas.
-- UUID rejeita qualquer valor não-UUID com erro 400 silencioso para o cliente.
-- Solução: converter para TEXT — sem perda de dados, índices recriados.

-- Remove índices que dependem do tipo UUID
DROP INDEX IF EXISTS idx_events_user_id;
DROP INDEX IF EXISTS idx_events_user_id_created_at;
DROP INDEX IF EXISTS idx_events_product_user_id_created_at;

-- Altera o tipo de UUID para TEXT (cast implícito UUID→text)
ALTER TABLE events ALTER COLUMN user_id TYPE TEXT USING user_id::text;

-- Recria índices com o novo tipo
CREATE INDEX IF NOT EXISTS idx_events_user_id
    ON events(user_id)
    WHERE user_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_events_user_id_created_at
    ON events(user_id, created_at DESC)
    WHERE user_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_events_product_user_id_created_at
    ON events(product, user_id, created_at DESC)
    WHERE user_id IS NOT NULL;
