-- 006_fix_brio_resolution_rate_denominator.sql
-- Recalcula historicamente a taxa de resolução do Brio corrigindo o denominador.
--
-- Problema: brioResolutionRate contava QUALQUER evento do produto 'brio' no
-- denominador (page_view, button_click, etc.) em vez de apenas eventos de
-- conversa iniciada. O numerador já filtrava por 'conversation.resolved_by_agent'.
--
-- Correção: o denominador agora conta apenas eventos onde event_type IN
-- ('conversation.started', 'session.initiated').
--
-- Como o cálculo da taxa é feito em tempo real (on-the-fly) no
-- north_star_handler.go (não é armazenado), nenhum UPDATE em tabelas é
-- necessário. Este migration documenta a correção e (opcionalmente) cria
-- uma view materializada de auditoria para depuração.
--
-- BKL-1015

-- View de auditoria: expõe as contagens brutas usadas no cálculo,
-- útil para debug e verificação de consistência.
CREATE OR REPLACE VIEW metrics.brio_resolution_rate_raw AS
WITH weekly_buckets AS (
    SELECT
        date_trunc('week', created_at)::date AS week_start,
        COUNT(*) FILTER (WHERE event_type IN ('conversation.started', 'session.initiated'))
            AS total_conversations,
        COUNT(*) FILTER (WHERE event_type = 'conversation.resolved_by_agent')
            AS total_resolved
    FROM events
    WHERE product = 'brio'
      AND created_at >= date_trunc('week', NOW()) - INTERVAL '8 weeks'
    GROUP BY date_trunc('week', created_at)::date
)
SELECT
    week_start,
    total_conversations,
    total_resolved,
    CASE
        WHEN total_conversations > 0
        THEN (total_resolved * 100) / total_conversations
        ELSE 0
    END AS resolution_rate_pct
FROM weekly_buckets
ORDER BY week_start DESC;

COMMENT ON VIEW metrics.brio_resolution_rate_raw IS
    'Auditoria do cálculo de brioResolutionRate (BKL-1015). Mostra contagens brutas por semana.';
