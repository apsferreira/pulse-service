# brioResolutionRate — North Star do Brio

## Definição Formal

**Nome:** Taxa de Conversas Resolvidas pelo Agente (brioResolutionRate)

**Fórmula:**

```
brioResolutionRate = (conversas resolvidas pelo agente / total de conversas iniciadas) × 100
```

Onde:

- **Numerador:** `COUNT(*) WHERE product = 'brio' AND event_type = 'conversation.resolved_by_agent'`
- **Denominador:** `COUNT(*) WHERE product = 'brio' AND event_type IN ('conversation.started', 'session.initiated')`

**Unidade:** Percentual inteiro (0–100)

## Por que essa métrica?

O Brio é um agente de IA que atende clientes via conversas. Esta métrica mede a
efetividade do agente em resolver conversas sem escalonamento humano.

- **100%** = toda conversa iniciada foi resolvida pelo agente
- **0%** = nenhuma conversa foi resolvida pelo agente

## O Problema Corrigido (BKL-1015 / Issue #63)

O denominador originalmente contava **QUALQUER evento** do produto 'brio'
— incluindo `page_view`, `button_click`, etc. — em vez de apenas eventos de
início de conversa.

**Impacto:** Uma conversa resolvida em meio a 100 page_views resultava em
aproximadamente 1% de taxa de resolução, quando a taxa real era 100%.

**Correção (2026-06-08):** Adicionado filtro `AND event_type = ANY($n::text[])`
com `['conversation.started', 'session.initiated']` no denominador.

## Implementação

**Arquivo:** `internal/handler/north_star_handler.go`
**Função:** `brioResolutionRate()`

```go
conversationStarts := []string{"conversation.started", "session.initiated"}
// Query do denominador:
WHERE product = 'brio' AND event_type = ANY($n::text[])
  AND created_at >= $m AND created_at < $p
```

## Migração

O cálculo é feito on-the-fly (não armazenado), portanto nenhuma migração de
dados é necessária. Uma view de auditoria `metrics.brio_resolution_rate_raw`
foi criada em `migrations/006_fix_brio_resolution_rate_denominator.sql` para
facilitar depuração.

## Sparkline

A série histórica compreende as 8 semanas mais recentes, cada semana
representada como um ponto percentual inteiro. O índice 0 do sparkline é a
semana mais antiga; o índice 7 é a semana mais recente.

## Testes

Os testes unitários em `internal/handler/north_star_handler_test.go` validam:

1. **Denominador correto** — 3 conversas, 2 resolvidas → 66%
2. **Eventos mistos** — page_views não entram no denominador
3. **Zero eventos** — retorna 0 (não panic)
4. **Sparkline completo** — 8 semanas consistentes
