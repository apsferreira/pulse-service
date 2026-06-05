# Pulse-Service — Análise SDD 2026-06-05
## Lentes 06 (Qualidade) e 09 (Dados)

---

## Lente 06 — Qualidade | Score: 3/10

### Cobertura Estimada

11 arquivos de teste para um serviço com 46 handlers de produção (excluindo vendor).
Ratio bruto: ~24% de cobertura de handlers. Cobertura real é substancialmente menor porque:

- `event_service_test.go` testa lógica de domínio inline (if/else reimplementados no próprio teste) sem instanciar o service real — os testes passam mesmo que a implementação quebre, porque replicam a lógica em vez de chamá-la.
- Os testes de handler cobrem os happy paths e 2–3 casos de erro dos handlers mais simples (event, cost, flag). Todos os demais 40+ handlers têm cobertura zero.
- `isolation_test.go` e `evaluate_flags_projectid_test.go` cobrem casos de multi-tenant, mas apenas no layer de flag evaluation.

Estimativa calibrada de cobertura de branches críticos: **12–18%**. Muito abaixo da meta de 80% para services críticos (R4).

### Paths Críticos sem Teste

| Handler / Feature | Risco | Observação |
|---|---|---|
| `mab_handler.go` — SelectVariant, RecordEvent, RecalculateWeights | CRÍTICO | Algoritmo epsilon-greedy sem um único teste. Erro silencioso corrompe todos os experimentos A/B. |
| `checkout_webhook_handler.go` — HandleCheckoutConfirmation | CRÍTICO | Atualiza plano de projeto (faturamento). Zero testes. Payload malformado não é tratado com idempotência testada. |
| `churn_model_handler.go` — Train, UserChurnRisk | ALTO | Gradient descent implementado em Go puro. Sem teste de convergência, overfitting ou dados insuficientes. |
| `flag_variants` (008_add_flag_variants.sql) | ALTO | `flag_handler_test.go` existe mas cobre apenas CRUD básico — não testa seleção de variante por peso, rollout percentual, prerequisite flags (migration 025). |
| `north_star_handler.go` — brioResolutionRate | ALTO | Divisão inteira `(resolved * 100) / total` perde precisão e retorna 0 quando `resolved < total/100`. Sem teste de boundary. |
| `webhook_handler.go` (033_create_webhooks, 044_webhook_types_and_delivery_history) | ALTO | Entrega de webhooks externos sem teste de retry, timeout ou HMAC signature. |
| `anomaly_handler.go`, `retention_handler.go`, `funnel_handler.go` | MÉDIO | Lógica analítica complexa, zero cobertura. |
| `replay_handler.go` (041_create_session_replays) | MÉDIO | Session replay sem teste de sanitização de PII. |
| `llm_observability_handler.go` (035_add_llm_observability) | MÉDIO | Sem teste de captura de latência, tokens e erros de LLM. |

### Problema Estrutural nos Testes de Service

`event_service_test.go` contém um anti-pattern grave: os testes reimplementam a lógica de negócio inline no corpo do teste em vez de chamar o service. Exemplo:

```go
// No teste — replica a lógica em vez de chamar Ingest()
env := req.Environment
if env == "" {
    env = "production"
}
if env != "production" { t.Errorf(...) }
```

Esses testes têm zero valor de detecção de regressão: se alguém apagar ou inverter a condição no service real, os testes continuam passando. Mutation Score estimado para este arquivo: **~0%**.

### CI Gaps

- Nenhuma evidência de `gosec` ou `govulncheck` no pipeline (obrigatórios por stack IIT).
- Nenhum coverage gate configurado — `go test -cover` não bloqueia o build.
- Nenhum teste de integração com banco real (testcontainers-go) para validar queries das migrations mais recentes (040–060).
- `trufflehog` ausente no CI (o repositório tem 60 migrations com seeds — risco de credential leak em seeds SQL como `042_seed_iit_project_rotate_api_key.sql`).

### Top 5 Movimentos de Qualidade

**Q1 — Reescrever `event_service_test.go` com stub real (impacto imediato).**
Instanciar `EventService` com `stubEventRepo` e chamar `svc.Ingest(...)` de verdade. Os testes atuais não protegem nada. Esforço: 1 dia. Blocker para qualquer deploy crítico.

**Q2 — Testar MAB SelectVariant com distribuição estatística.**
Rodar SelectVariant 1000 vezes e verificar que a frequência de seleção converge para os pesos configurados (tolerância ±5%). Sem isso qualquer bug no epsilon-greedy é invisível. Esforço: 2 dias.

**Q3 — Testar checkout webhook com idempotência e payload inválido.**
Simular: evento desconhecido, project_id inexistente, plan_slug sem mapeamento, requisição duplicada. O handler atual atualiza plano de billing sem idempotência testada. Esforço: 1 dia.

**Q4 — Adicionar coverage gate no CI (threshold 60% inicial, meta 80%).**
`go test ./... -coverprofile=cover.out && go tool cover -func=cover.out | grep -E "^total" | awk '{if ($3+0 < 60) exit 1}'`. Gate atual: nenhum. Esforço: 2 horas.

**Q5 — Corrigir `brioResolutionRate` e adicionar teste de boundary.**
A fórmula `(resolved * 100) / total` com int64 retorna 0 quando resolved < 1% do total. Converter para float64 antes de multiplicar. Adicionar teste com resolved=1, total=200 (esperado: 0% com bug, 0.5% correto). Esforço: 2 horas.

---

## Lente 09 — Dados | Score: 4/10

### Schema de Eventos (tabela `events`)

Campos coletados após todas as migrations aplicadas (004 + 013 + 017 + 052):

| Campo | Tipo | Observação |
|---|---|---|
| `id` | UUID PK | gen_random_uuid() |
| `product` | VARCHAR(50) NOT NULL | 'libri', 'nitro', 'brio', etc. — string livre, sem FK |
| `event_type` | VARCHAR(100) NOT NULL | string livre, sem enum/validação |
| `user_id` | TEXT | migration 052 converteu de UUID para TEXT (anon: prefix) |
| `session_id` | VARCHAR(100) | nullable |
| `properties` | JSONB DEFAULT '{}' | schema livre — sem event_schemas enforcement em runtime |
| `ip_address` | INET | nullable |
| `user_agent` | TEXT | nullable |
| `environment` | VARCHAR(20) DEFAULT 'production' | adicionado em 013 |
| `project_id` | UUID FK → projects | adicionado em 017, nullable (dados históricos = NULL) |
| `created_at` | TIMESTAMPTZ DEFAULT NOW() | |
| `anonymous_id` | ausente | não existe como coluna separada — merged no user_id com prefixo 'anon:' |

**Indexação atual:**
- `idx_events_product` (product)
- `idx_events_event_type` (event_type)
- `idx_events_created_at` (created_at DESC)
- `idx_events_event_type_created_at` (event_type, created_at DESC) — migration 005
- `idx_events_user_id` WHERE user_id IS NOT NULL — migration 005
- `idx_events_environment` (environment) — migration 013
- `idx_events_project_id` (project_id) — migration 017

**Índice crítico ausente:** não existe índice composto `(product, event_type, created_at)` — que é exatamente o padrão de query do `north_star_handler.go` (countEvents usa WHERE product=$1 AND event_type=$2 AND created_at BETWEEN). Em tabela de eventos com alto volume, essa query faz seq scan ou usa o índice parcial de event_type+created_at sem filtrar product, o que é ineficiente.

### Retenção de Dados

- `monitor_checks`: cleanup automático via função `cleanup_monitor_checks()` (90 dias) — migration 059. Mas não há cron job ou pg_cron chamando essa função: a função existe no schema mas nunca é invocada automaticamente.
- `events`: sem política de retenção configurada. A migration 024 adiciona `data_retention` ao projeto mas não há job que execute o purge baseado nessa configuração.
- `session_replays`: sem retenção explícita (migration 041). Pode crescer indefinidamente.
- `api_costs`: sem retenção. Tabela de custo acumula para sempre.

### North Star Metrics — Implementadas vs Esperadas

| Produto | Métrica Implementada | Problema |
|---|---|---|
| SocialMake | `post.published` count/semana | OK — evento direto |
| Nitro | `session.completed` count/semana | OK — evento direto |
| Libri | `book.status_changed` + `{"to_status":"read"}` count/mês | OK — JSONB filter correto |
| MeuIngresso | `ticket.purchased` count/semana | Produto não existe no ecossistema ativo — dados sempre zero |
| Brio | `conversation.resolved_by_agent` / total brio events | FALHO — denominador conta TODOS os eventos de brio, não só conversas. Uma `page_view` no painel admin dilui a taxa. |
| Pulse | `COUNT(DISTINCT project_id)` WHERE `event_type='flag_evaluation'` | OK — mas depende de eventos chegando via SDK; projetos que usam só REST não geram esse evento |

**Problema crítico no Brio:** `brioResolutionRate` divide `conversation.resolved_by_agent` pelo total de eventos do produto "brio" (qualquer event_type). Se o app Brio registrar eventos de UI (`page_view`, `button_click`) além de conversas, a taxa de resolução aparece artificialmente baixa. A métrica correta seria: `resolved / total WHERE event_type LIKE 'conversation.%'`.

**Problema estrutural:** o campo `product` é string livre sem foreign key ou enum. O `north_star_handler` filtra por strings hardcoded ("socialmake", "nitro", "libri", "brio", "meuingresso") que precisam ser idênticas às enviadas pelos SDKs. Se um produto enviar "Nitro" (maiúscula) ou "focus-hub" em vez de "nitro", a métrica retorna zero silenciosamente. Não há validação ou normalização no ingestion path.

### Gaps de Dados

**G1 — Dados históricos pré-migration 017 sem project_id.**
Todos os eventos ingeridos antes da migration 017 têm `project_id = NULL`. As queries do north_star_handler não filtram por project_id (correto para o dashboard global), mas queries por projeto (handlers de analytics por projeto) podem misturar dados de diferentes tenants se `project_id IS NULL` não for tratado explicitamente.

**G2 — Ausência de `anonymous_id` como coluna separada.**
O design atual usa `user_id = 'anon:device-uuid'` para eventos anônimos. Isso impede queries de funil anon→identified sem string matching (`LIKE 'anon:%'`). O `user_identity` (migration 031) resolve o merge, mas o custo de join é alto.

**G3 — `event_schemas` (migration 028) não são aplicados em runtime.**
A tabela `event_schemas` define tipos e propriedades esperadas por evento. Mas o `event_handler.go` insere eventos sem validar contra esses schemas — qualquer payload JSONB é aceito. O schema existe apenas como documentação, não como contrato.

**G4 — Sem índice `(product, event_type, created_at)` para queries North Star.**
As 9 queries paralelas do `north_star_handler` (8 semanas de sparkline + current + prev por produto) executam sem o índice composto ideal. Em produção com volume alto, esse endpoint causa 54 queries sequenciais sem índice otimizado.

**G5 — `page_views` é tabela separada (migration 060) duplicando dados de eventos.**
A tabela `page_views` foi criada em 060 como entidade separada, mas `event_type='page_view'` também é ingerido na tabela `events`. Não está claro qual é a fonte de verdade para análise de page views. Risco de métricas divergentes.

### Top 5 Movimentos de Dados

**D1 — Criar índice composto `(product, event_type, created_at DESC)` na tabela events.**
Query do north_star_handler atinge esse padrão 54x por request. Sem o índice, degrada com volume. Adicionar migration 061 com `CREATE INDEX CONCURRENTLY`. Esforço: 30 minutos.

**D2 — Corrigir denominador do brioResolutionRate.**
Trocar `WHERE product='brio'` para `WHERE product='brio' AND event_type LIKE 'conversation.%'` no cálculo do denominador. Sem isso a North Star do Brio é matematicamente errada. Esforço: 2 horas + teste.

**D3 — Invocar `cleanup_monitor_checks()` via pg_cron ou cron job na VM.**
A função existe mas nunca é chamada. `monitor_checks` cresce indefinidamente. Adicionar `SELECT cron.schedule('cleanup-monitor-checks', '0 3 * * *', 'SELECT cleanup_monitor_checks()')` ou equivalente no deploy. Esforço: 1 hora.

**D4 — Normalizar `product` no ingestion path.**
Adicionar `strings.ToLower(strings.TrimSpace(req.Product))` no `EventService.Ingest` antes de persistir. Evita que "Nitro" e "nitro" gerem métricas separadas e North Stars zeradas. Esforço: 1 hora + teste.

**D5 — Definir fonte de verdade entre `events` e `page_views`.**
Decidir: page views são registradas só em `page_views` (migration 060) ou também em `events`? Documentar no ADR e, se duplicado, criar migration de limpeza ou remover o `page_view_handler` duplicado. Ambiguidade atual impossibilita análise de funil confiável. Esforço: decisão arquitetural (30min) + migration de consolidação (2h).
