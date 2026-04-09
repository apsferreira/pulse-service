# PRD — Pulse Service v1.0
**Observabilidade, Feature Management e Analytics para o Ecossistema IIT**

**Status:** Implementação em progresso (P0: flags + events ✅ | P1: CSAT/surveys | P2: WebSocket dashboard)
**Última atualização:** 09 de abril de 2026
**Proprietário:** @pm (Antonio) | Líder técnico: @tl

---

## Executive Summary

**Pulse** é a camada de observabilidade, feature management e analytics centralizada do ecossistema Instituto Itinerante. Ele conecta todos os produtos (Libri, Nitro, Brio, SocialMake) em torno de um núcleo compartilhado de:

1. **Feature Flags** — controle em tempo real de features, A/B testing, rollout percentual
2. **Event Tracking** — produto analytics, comportamento do usuário, diagnóstico
3. **CSAT/NPS** — feedback estruturado de clientes
4. **Cost Tracking** — monitoramento de custos e overages (para SocialMake, etc)

**Valor para o negócio:**
- Reduz time-to-market: deployments sem downtime via flags
- Habilita data-driven decisions: analytics em tempo real
- Melhora retenção: CSAT/NPS contínuo
- Controla custos: tracking de API calls, GPT tokens, etc

**Modelo de negócio:** Freemium (Free/Growth/Pro), auto-faturamento via Asaas.

---

## 1. Contexto e Problema

### Estado Atual (Antes de Pulse)
- Cada produto (Libri, Nitro, Brio) mantém seu próprio sistema de analytics
- Deploy de novas features requer rebuild + downtime
- Sem A/B testing estruturado — decisões baseadas em intuição
- Feedback de clientes (CSAT/NPS) é manual, esporádico
- Custo de APIs externas (OpenAI, etc) é invisível — sem rastreamento por tenant

### Problema
1. **Fragmentação:** múltiplas sources of truth para analytics
2. **Velocidade:** deploys acoplados = time-to-market lento
3. **Risco:** mudanças sem teste — crashes em produção
4. **Falta de dados:** decisões de produto sem suporte em dados
5. **Custo opaco:** SocialMake não sabe quanto custa rodar Aria por cliente

### Oportunidade
Centralizar observabilidade, flags e analytics numa plataforma simples, integrada nativamente em todos os serviços IIT — diferencial competitivo contra concorrentes (Stripe, Shopify) que usam PostHog/Amplitude caros.

---

## 2. Objetivos (OKRs)

### Q2 2026 (Abril-Junho)
- **O1:** Feature flags + A/B testing estáveis em produção
  - KR1.1: 100% dos 5 serviços principais (auth, libri, nitro, brio, socialmake) integrados com Pulse
  - KR1.2: 0 downtime deploys via flags rollout percentual
  - KR1.3: <100ms p99 latência flag evaluation (cache Redis)

- **O2:** Event tracking confiável
  - KR2.1: 99.9% uptime para ingestão de eventos
  - KR2.2: <5s latência entre evento enviado e visível no dashboard

- **O3:** Validar pricing e Go-to-Market
  - KR3.1: 3+ clientes externos em Growth/Pro (beta)
  - KR3.2: NPS >= 50 entre early adopters

### Q3 2026 (Julho-Setembro)
- CSAT/NPS surveys integradas
- Retention analysis (cohort, churn rate)
- Dashboard WebSocket em tempo real

### Q4 2026+ (Roadmap)
- Self-hosted deployment
- Custom integrations (Slack, Segment)
- Enterprise SLA (99.99%, dedicated manager)

---

## 3. Personas e User Stories

### Persona 1: Dev/Tech Lead (Usuário Técnico)

**Identidade:** Líder técnico num serviço IIT (ex: my-library, focus-hub)
**Contexto:** Precisa fazer deploy de novo algoritmo de recomendação com segurança
**Goal:** Ativar feature apenas para 10% dos usuários, monitorar, expandir gradualmente

**User Story:**
```
Como dev de my-library
Quero criar uma flag "new_recommendation_v2" e ativar para 10% dos usuários
Para fazer deploy sem risco — se houver crash, desativo e volta ao v1 em <1s

Critérios de Aceite:
- Dado que criei a flag via API
- Quando seto rollout_percentage: 10
- Então 10% dos users veem v2, 90% veem v1 (hash por userID = consistência)
- E quando desativo, 100% veem v1 novamente (em <10s)

Hipótese de Valor:
Se implementarmos flags, esperamos reduzir MTTR (mean time to recovery) de crashes de 30min (rebuild + deploy) para 30s (disable flag).
```

### Persona 2: Product Manager / Data Analyst

**Identidade:** PM do SocialMake, precisa validar se Aria está aumentando conversion
**Contexto:** Rodou A/B test: Aria vs baseline em 50% dos usuários
**Goal:** Medir qual variante melhor, decidir qual versão vai para todos

**User Story:**
```
Como PM do SocialMake
Quero criar um A/B test "aria_vs_baseline" com 50/50 split
E ver em tempo real: conversion, cost (GPT tokens), time-to-post de cada variante
Para decidir qual versão rollout para 100%

Critérios de Aceite:
- Dado que criei test com 2 variantes
- Quando envio eventos "post_created" com metadata {variant, cost, time}
- Então dashboard mostra: conversion rate, cost médio, p50/p95 time de cada variante
- E CI/confidence interval do teste
- E quando atinjo 1000 amostras por variante, sugerir winner

Hipótese de Valor:
Se implementarmos A/B testing estruturado, esperamos aumentar conversion de Aria de 30% para 40% (based on early data).
```

### Persona 3: Customer Success / Support

**Identidade:** CS do IIT monitorando saúde geral do SocialMake
**Contexto:** Precisa detectar rapidamente quando algo quebrou em produção
**Goal:** Dashboard de CSAT/NPS, alertas quando satisfação cai

**User Story:**
```
Como CS do SocialMake
Quero ver NPS médio dos tenants mensalmente
E receber alerta quando NPS cai >20% vs mês anterior
Para intervir antes de churn

Critérios de Aceite:
- Dado que envio evento "survey_response" com nps_score: 8
- Quando no final do mês agrego por tenant
- Então dashboard mostra: NPS por tenant, NPS médio geral, trend linha temporal
- E quando NPS cai >20% vs baseline, gera alert

Hipótese de Valor:
Se rastrearmos NPS contínuamente, esperamos reduzir churn de 5% para 2% (target Libri).
```

---

## 4. Features Mapeadas (por Fase)

### P0: MVP (Feature Flags + Event Tracking)
**Status:** ✅ Implementado

- [x] **Feature Flags — CRUD**
  - Create, read, update, delete flags via API (`/api/v1/flags`)
  - Metadados: name, description, tags, enabled, rollout_percentage
  - Persistência em PostgreSQL + cache Redis

- [x] **Feature Flags — Evaluation**
  - Avaliar flag para (userId, tenantId) via SDK
  - Hash consistente (xxhash userId + salt) para rollout percentual
  - Overrides por usuário/tenant (forçar enabled/disabled)

- [x] **Feature Flags — Batch**
  - Endpoint POST `/api/v1/flags/evaluate` para avaliar múltiplas flags de uma vez
  - Usado por frontends para cachear todas as flags numa única chamada

- [x] **A/B Testing — Variants**
  - Flag com `type: "ab_test"` pode ter múltiplas variantes (ex: {control, variant_a, variant_b})
  - Distribuição percentual de usuários entre variantes
  - SDK retorna `{flag_name: "control"}` ou `variant_a`, etc

- [x] **Event Tracking — Ingestão**
  - POST `/api/v1/events` para registrar evento de usuário
  - Payload: {userId, tenantId, eventName, timestamp, metadata}
  - Validação de quotas (Free: 10k/mês, Growth: 500k/mês, Pro: ilimitado)
  - Fila Kafka/Redis para ingestão assíncrona

- [x] **Event Tracking — Retenção**
  - Histórico: Free 7d, Growth 30d, Pro 90d
  - Cleanup automático de eventos expirados

- [x] **SDKs**
  - [x] SDK JavaScript (frontend + Node backend)
  - [x] SDK Go (backend)
  - [ ] SDK Python (roadmap)

---

### P1: Analytics + CSAT (Q3 2026)
**Status:** ⏳ Planejado

- [ ] **CSAT/NPS Surveys**
  - Widget embeddável: pergunta "Como foi sua experiência?" (1-10)
  - Endpoint POST `/api/v1/surveys` para registrar resposta
  - Dashboard: distribuição scores, trend temporal

- [ ] **Cost Tracking**
  - Evento especial: `{eventName: "cost_incurred", metadata: {service, cost_usd, unit}}`
  - Agregação: custo total por tenant/serviço/mês
  - Dashboard: custo por tenant, trending (ex: Aria cost/1000 posts)

- [ ] **Retention Cohort Analysis**
  - Usuários grouped por data de signup
  - Métrica: %usuarios que voltaram em semana 1, 2, 4, 8, etc
  - Visualização: heatmap cohort

- [ ] **Event Explorer**
  - Busca livre em eventos: `eventName LIKE 'post_%' AND metadata.network = 'linkedin'`
  - Aggregations: count, histogram, top values (distinct userIds, etc)

---

### P2: Real-time Dashboard + Enterprise (Q4 2026+)
**Status:** ⏳ Planejado

- [ ] **WebSocket Dashboard**
  - Real-time updates: evento registrado → visível em <1s no dashboard
  - Canais por project_id, filtrável por eventName/tags

- [ ] **Custom Webhooks**
  - Trigger: "quando evento com {tag: 'critical'}, chama webhook X"
  - Integração Slack: alerta instant quando NPS cai

- [ ] **Project Scoping**
  - Isolamento: projeto A não vê dados de projeto B
  - RBAC: owner pode convidar teammates com role-based access

- [ ] **Self-hosted Deployment**
  - Docker Compose para deploy privado
  - PostgreSQL + Redis externo

---

## 5. Integrações Obrigatórias (MVP)

| Serviço | Feature | Status | Deadline |
|---------|---------|--------|----------|
| **auth-service** | Flags: `force_2fa`, `sso_google` | ✅ Integrado | 09 mar |
| **my-library (Libri)** | Flags: `amazon_button`, `audiobook_v2` | ✅ Integrado | 09 mar |
| **focus-hub (Nitro)** | Flags: `pomodoro_v2`, `notification_redesign` | ✅ Integrado | 09 mar |
| **attend-service (Brio)** | Events: `booking_created`, `whatsapp_message_sent` | ⏳ Em progresso | 15 mai |
| **SocialMake** | Flags: `aria_enabled` | ✅ Integrado |  |
| | Events: `post_created`, `cost_incurred` | ✅ Integrado | 09 abr |
| **checkout-service** | Events: `checkout_started`, `payment_completed` | ⏳ Planejado | 30 mai |
| **contract-service** | Events: `contract_signed`, `renewal_due` | ⏳ Planejado | 30 jun |

**Dependências críticas:**
- Pulse assume que `auth-service` está UP (autentica SERVICE_TOKEN)
- Redis deve estar acessível em todos os serviços (cache flags)

---

## 6. Preços e Planos

Veja `/docs/PRICING.md` para detalhes. Resumo:

| Plano | Preço | Eventos/mês | Flags | Projetos | A/B Test | SLA |
|-------|-------|-------------|-------|----------|----------|-----|
| **Free** | R$ 0 | 10k | 5 | 1 | ❌ | Best effort |
| **Growth** | R$ 99 | 500k | ∞ | 5 | ✅ | 99.0% |
| **Pro** | R$ 297 | ∞ | ∞ | ∞ | ✅ + CSAT | 99.9% |
| **Enterprise** | Custom | ∞ | ∞ | ∞ | ✅ + RAG | 99.99% |

**Diferenciador:**
- Preço em R$, calibrado para startups brasileiras
- All-in: flags + events + CSAT num único plano (vs PostHog/Amplitude que cobram por feature)
- Sem "taxas ocultas" — limites de quota claros

---

## 7. Arquitetura Técnica

### Stack
- **Backend:** Go 1.24 + Fiber framework
- **DB:** PostgreSQL 15 (flags, projetos, metadados) + Qdrant (vector search para CSAT)
- **Cache:** Redis 7 (flags cache, quota tracking)
- **Queue:** Kafka ou RabbitMQ (ingestão assíncrona de eventos — a definir)
- **Frontend:** React + TypeScript + Vite + TailwindCSS
- **Hosting:** K3s (prod) / Docker Compose (dev)

### Componentes

**Backend (`github.com/apsferreira/pulse-service/backend`):**
- `cmd/api` — Fiber app (routes, handlers)
- `internal/flags` — flag CRUD, evaluation logic, cache invalidation
- `internal/events` — ingestion, quota validation, retention cleanup
- `internal/surveys` — CSAT/NPS, widget embed code gen
- `internal/auth` — JWT validation via auth-service

**Frontend (`github.com/apsferreira/pulse-service/frontend`):**
- `/pages/Flags.tsx` — CRUD flags, edit rollout%, preview
- `/pages/ABTest.tsx` — create/analyze A/B tests
- `/pages/Events.tsx` — explorer, aggregations
- `/pages/Dashboard.tsx` — overview (event rate, flag changes, top events)

**SDKs:**
- `/sdk/go` — `pulse.NewClient()`, `EvaluateFlags()`, `Track()`
- `/sdk/js` — `new PulseSDK()`, isEnabled(), track()

### Fluxos Principais

**Flag Evaluation (P0):**
```
Client calls pulse.isEnabled('amazon_button', userId: 'user-123')
  → SDK requests GET /api/v1/flags/{flag_id}/evaluate?userId=user-123
  → Backend: checks Redis cache (flags + user overrides)
    If cache miss: query PostgreSQL, populate cache (TTL 60s)
  → Return: {enabled: true/false}
  → (Alternative: Frontend caches all flags, no per-flag request)
```

**Event Ingestão (P0):**
```
Client calls pulse.track('post_created', {cost: 0.5, network: 'linkedin'})
  → SDK batches events (100 events ou 5s timeout)
  → POST /api/v1/events (batch) com SERVICE_TOKEN
  → Backend: validate quota, check plan limits
    If quota OK: put evento no Kafka/Redis queue para processamento async
    Return: {accepted: true}
  → Worker processes queue: insert em PostgreSQL events table + update aggregations (Redis)
  → Events visível no dashboard em <5s
```

**Flag Rollout Percentual (P0):**
```
Flag {name: "new_algo", rollout_percentage: 30, salt: "flag-seed-123"}
Client evaluation para userId "user-123":
  → hash = xxhash(userId + salt) % 100
  → if hash < 30: enabled = true
  → Determinístico: mesmo usuário sempre vê a mesma variante
```

---

## 8. Roadmap Detalhado

### Sprint 1 (09-20 abr 2026) — P0 Hardening
- [x] Feature flags: CRUD, eval, cache
- [x] Event ingestão: batch, quota, async
- [x] SDK JS + Go
- [x] Dashboard básica
- [x] Auth integration (SERVICE_TOKEN)
- [ ] Load test: 1M events/min, 100ms p99 flag eval
- [ ] Alert em caso de quota exceed

### Sprint 2 (21 abr-04 mai 2026) — A/B Testing
- [ ] Flag type "ab_test" com múltiplas variantes
- [ ] Analytics por variante: conversion, cost, retention
- [ ] Statistical significance (CI, p-value)
- [ ] Dashboard winner recommendation
- [ ] Integração SocialMake (Aria A/B test)

### Sprint 3 (05-18 mai 2026) — CSAT/Surveys
- [ ] Survey widget embeddável
- [ ] Dashboard: NPS, CSAT distribuição, trends
- [ ] Integração Libri (survey ao finish book)
- [ ] RAG para respostas abertas (Qdrant + Aria)

### Sprint 4+ (mai-jun 2026) — Polish + Enterprise
- [ ] Custom webhooks + Slack integration
- [ ] Project scoping + RBAC
- [ ] Self-hosted deployment docs
- [ ] Enterprise pricing ($1990/mês)

---

## 9. Métricas de Sucesso

### Adoção
- **% serviços integrados:** (serviços com Pulse integrado) / 5 = 100% (target)
- **% flags em produção:** flags ativas em P0 vs total flags criadas
- **# tenants ativos:** clientes SaaS (crescimento mensal)

### Confiabilidade
- **Uptime:** 99.9% em Growth+, Best effort em Free
- **Latência flag eval:** p99 < 100ms (cached) / 500ms (cold)
- **Event loss rate:** < 0.1% (target zero)

### Negócio
- **MRR Pulse:** R$ (Growth users × 99 + Pro users × 297)
- **NPS Pulse:** >= 50 (early adopters)
- **Cost-per-event:** (<R$ 0,001 por evento) para manter margen

### Produto
- **Adoption KPI:** DAU (daily active flags) / total users (target: 30% de users veem flag)
- **A/B test velocity:** # tests launched by month (target: 4+ testes/mês)
- **CSAT score:** média >= 4.0 em escala 1-5

---

## 10. Requisitos Funcionais (PRD → Specs)

### RF-001: Feature Flag CRUD
**Descrição:** Admin e devs podem criar, listar, atualizar e deletar flags.
**Endpoints:**
- POST `/api/v1/flags` — Create flag
- GET `/api/v1/flags` — List (com paginação, filtro por product/tag)
- GET `/api/v1/flags/{id}` — Get flag details
- PUT `/api/v1/flags/{id}` — Update (enabled, rollout%, description)
- DELETE `/api/v1/flags/{id}` — Delete

**Validações:**
- name único por project
- rollout_percentage: 0-100
- Se enabled=false, rollout_percentage ignorado

---

### RF-002: Flag Evaluation
**Descrição:** SDKs podem avaliar se uma flag está ativa para um usuário.
**Endpoint:**
- POST `/api/v1/flags/evaluate` — Avaliar múltiplas flags ({userId, tenantId} + flags array)
- GET `/api/v1/flags/{id}/evaluate?userId=X&tenantId=Y` — Avaliar uma flag

**Lógica:**
- Se flag.enabled = false: retorna false para todos
- Se flag.enabled = true:
  - Checar override por (userId, tenantId)
  - Se override: retornar override value
  - Senão: hash userId com rollout_percentage
- Resultado: {flag_name: boolean}

---

### RF-003: Event Tracking
**Descrição:** Aplicações rastreiam eventos (pageview, purchase, etc) com metadata.
**Endpoint:**
- POST `/api/v1/events` — Ingerir batch de eventos

**Payload:**
```json
{
  "events": [
    {
      "userId": "user-123",
      "tenantId": "tenant-456",
      "eventName": "post_created",
      "timestamp": "2026-04-09T18:00:00Z",
      "metadata": {
        "network": "linkedin",
        "cost": 0.05
      }
    }
  ]
}
```

**Validações:**
- userId + tenantId obrigatórios
- eventName max 64 chars
- metadata max 2KB
- Verificar quota da conta

---

### RF-004: A/B Testing
**Descrição:** Flags podem ter múltiplas variantes com distribuição percentual.
**Payload flag com A/B:**
```json
{
  "name": "aria_test",
  "type": "ab_test",
  "variants": [
    {"name": "control", "percentage": 50},
    {"name": "aria_v1", "percentage": 50}
  ]
}
```

**Evaluation retorna:**
```json
{
  "aria_test": "control"  // ou "aria_v1"
}
```

---

### RF-005: CSAT/NPS (P1)
**Descrição:** Widget embeddável coleta feedback 1-10.
**Endpoint:**
- POST `/api/v1/surveys` — Registrar resposta

**Payload:**
```json
{
  "userId": "user-123",
  "tenantId": "tenant-456",
  "score": 8,
  "comment": "Excelente experiência",
  "context": "post_published"
}
```

**Dashboard:** NPS médio, distribuição, trends por tenant.

---

## 11. Requisitos Não-Funcionais

- **Performance:** Flag evaluation < 100ms p99
- **Escalabilidade:** 1M events/min
- **Confiabilidade:** 99.9% uptime (Growth+)
- **Segurança:** JWT via auth-service, SERVICE_TOKEN para backend
- **Compliance:** LGPD (direito ao esquecimento de eventos), audit logs (Enterprise)

---

## 12. Dependências Externas

- **auth-service:** valida SERVICE_TOKEN
- **PostgreSQL:** flags, surveys, aggregations
- **Redis:** cache flags, quota tracking, session user-flags
- **Kafka/RabbitMQ:** fila de eventos (a definir — pode ser Redis no MVP)

**Risco de bloqueio:** Nenhum — Pulse não depende de nenhum serviço IIT para operar. Auth-service é via token, desacoplado.

---

## 13. Go-to-Market e Pricing

### Estratégia
1. **Beta Interno:** Integrar Libri, Nitro, Brio, SocialMake (gratuito, validar UX)
2. **Early Adopters:** 3-5 startups em Growth/Pro (R$ 99-297/mês)
3. **Public Launch:** Comunicar via GDG Salvador, LinkedIn, blog IIT

### Conversão Esperada
- Free users: 5% → Growth (monthly churn 10% = 4 users)
- Growth users: 20% → Pro

### Suporte
- Free: Discord + GitHub Issues
- Growth: Email (resposta < 48h)
- Pro: Email + Slack channel (< 2h)

---

## 14. Riscos e Mitigações

| Risco | Severidade | Mitigação |
|-------|-----------|----------|
| Spam de eventos | Medium | Quota por plan, rate limiting |
| Performance em escala | High | Load test 1M/min, Kafka, índices PG |
| Leak de dados entre tenants | Critical | RBAC, isolamento query by tenant_id |
| Churn de Free users | Low | NPS tracking, customer feedback loop |

---

## 15. Success Criteria

**MVP (09 abr):**
- ✅ Feature flags funcionando em 5+ serviços
- ✅ Event tracking em produção com 99.0% uptime
- ✅ Dashboard básica mostrando flags e event timeline

**Go-live (30 mai):**
- 3+ clientes em beta (Growth/Pro)
- NPS >= 50 entre early adopters
- A/B testing com análise de variantes

**1 ano (abr 2027):**
- 50+ clientes ativos
- MRR R$ 5k+ (mix Free/Growth/Pro)
- NPS >= 60

---

## Apêndice: Links e Referências

- **Getting Started:** `docs/guides/getting-started.md`
- **API Reference:** `docs/API.md` (a criar em specs)
- **SDK Docs:** `docs/sdk/README.md`
- **Pricing:** `docs/PRICING.md`
- **Architecture Decision Records:** `docs/adr/` (ADRs) — a estruturar
- **Competitive Analysis:** Pulse vs PostHog vs LaunchDarkly vs Amplitude (ver PRICING.md)
