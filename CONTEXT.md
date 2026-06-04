# pulse-service — Context

## Propósito
Plataforma de **feature flags** e **event tracking** self-hostável do ecossistema Instituto Itinerante de Tecnologia (IIT). Permite ativar/desativar funcionalidades sem redeploy, segmentar usuários por rollout determinístico, rastrear eventos de produto e gerar insights analíticos. Serve múltiplos produtos (SocialMake, Librinho, Nitro, Site IIT) via isolamento por `Project`.

## Stack
- **Linguagem/framework:** Go 1.25 + Fiber v2
- **Banco:** PostgreSQL 15 (pgxpool / pgx v5)
- **Cache:** Redis 7 (go-redis/v9, TTL 30s para avaliação de flags)
- **Frontend (embutido):** React 19 + TypeScript + Vite + TailwindCSS
- **Observabilidade:** OpenTelemetry (OTLP/gRPC) + Prometheus (fiberprometheus)
- **Infra:** Docker Compose (dev) / K3s + ArgoCD (prod)

## Linguagem ubíqua

| Termo | Definição |
|---|---|
| **Project** | Unidade de isolamento; cada produto IIT tem seu próprio projeto com `api_key` e `plan` |
| **FeatureFlag** | Flag on/off com rollout percentual, targeting e variantes A/B; identificada por `key` |
| **RolloutPercentage** | % de usuários que veem a flag ativa; determinístico via `hash(user_id) mod 100` |
| **Targeting** | Lista explícita de `user_ids` / `tenant_ids` que recebem a flag independente do rollout |
| **Variant** | Alternativa nomeada de uma flag A/B (ex: `control`, `challenger`); pesos somam 100 |
| **Segment** | Conjunto de regras compostas sobre atributos do usuário para targeting avançado |
| **Event** | Ação rastreada de produto (`event_type`, `user_id`, `properties`, `environment`) |
| **Environment** | Escopo de execução: `production` \| `staging` \| `development` |
| **HealthScore** | Pontuação 0–100 de "saúde" de um projeto, composta por 4 dimensões configuráveis |
| **MAB** | Multi-Armed Bandit epsilon-greedy para otimização automática de variantes A/B |
| **FlagEvaluation** | Registro de auditoria de cada avaliação de flag (user_id, resultado, timestamp) |
| **InsightsAgent** | Agente autônomo que analisa flags e eventos e gera insights narrativos por projeto |

## Limites (bounded context)

- **Possui:** feature flags, rollout/targeting, variantes A/B, eventos de produto, projetos/contas IIT, ambientes, segments, health scores, session replays, LLM observability, ai-configs, webhooks de produto, CSAT/NPS, retenção de dados, licenças self-hosted
- **NÃO possui:** autenticação de usuários finais (auth-service), billing/checkout (checkout-service), catálogo de produtos (catalog-service), contratos (contract-service)
- **Upstream:** `auth-service` (enriquecimento de identidade via `authclient`), `checkout-service` (webhook de confirmação de pagamento → atualiza plano)
- **Downstream:** SDKs JS/React/React Native/Go/Node (consomem `GET /flags` e `POST /events`)

## Mapa de módulos

| Dir / arquivo-chave | O que faz |
|---|---|
| `cmd/server/main.go` | Ponto de entrada; wiring de todas as camadas (repository → service → handler) e registro de rotas |
| `internal/domain/` | Tipos canônicos: `FeatureFlag`, `Project`, `Event`, `Segment`, `MAB`, etc. |
| `internal/handler/` | Handlers Fiber por recurso; autenticação via `X-Project-Key` (SDK) ou `X-Service-Token` (admin) |
| `internal/service/flag_service.go` | Avaliação determinística de flags, cache Redis, approval workflow |
| `internal/repository/` | Acesso a PostgreSQL (pgxpool); um arquivo por entidade |
| `internal/agent/` | `InsightsAgent` — loop periódico que gera insights narrativos com LLM |
| `internal/pkg/flagscheduler/` | Job 60s: ativa/desativa flags agendadas (`starts_at` / `ends_at`) |
| `internal/pkg/retentionpurge/` | Job diário: purga eventos mais antigos que `data_retention_days` do projeto |
| `migrations/` | 59 migrations SQL numeradas; fonte da verdade do schema |
| `sdk/` | SDKs oficiais distribuídos junto ao serviço |

## Pontos de entrada

- **Servidor HTTP:** `go run ./cmd/server` — porta `3050` (configurável via `PORT`)
- **SDK público:** `GET /flags` + `POST /api/v1/events` (auth: `X-Project-Key`)
- **Admin API:** `GET|POST|PUT|DELETE /api/v1/*` (auth: `X-Service-Token`)
- **WebSocket:** `GET /ws` — push de atualizações em tempo real para o dashboard
- **SSE:** `GET /api/v1/flags/stream` — stream de mudanças de flags para o frontend
- **Jobs background:** monitor de uptime, health score diário, churn alert, flag scheduler, retention purge, insights agent
