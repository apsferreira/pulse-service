# Pulse — Modos de Deployment

**Data:** Abril 2026
**Versão:** 1.0

Este documento descreve os três modos de deployment do Pulse e o roadmap de paridade com PostHog/Mixpanel.

---

## 1. Modo IIT — Gerenciado (Atual)

O Pulse roda no K3s do IIT como serviço interno. É o modo de referência — o que está em produção hoje.

### Arquitetura

```
Internet → Cloudflare Tunnel → Traefik (K3s) → pulse-service (Go/Fiber)
                                                      ↓
                                              PostgreSQL (VM 192.168.30.121)
                                              Redis (shared-infra)
```

O frontend (`pulse.institutoitinerante.com.br`) não está exposto via CF Tunnel por política — é acesso interno/admin.

### Variáveis de Ambiente (K3s Secret: `pulse-service-secrets`)

| Variável | Obrigatório | Exemplo | Descrição |
|----------|-------------|---------|-----------|
| `DATABASE_URL` | Sim | `postgres://postgres:senha@192.168.30.121:5432/pulse_db` | PostgreSQL connection string |
| `REDIS_URL` | Sim | `redis://redis:6379` | Redis para cache de flags |
| `SERVICE_TOKEN` | Sim | `tok_...` | Token admin para endpoints `/api/v1/*` |
| `PORT` | Não (default 3015) | `3015` | Porta HTTP do backend |
| `RESEND_API_KEY` | Sim (churn alerts) | `re_...` | Churn alert emails |
| `RESEND_FROM` | Sim (churn alerts) | `pulse@institutoitinerante.com.br` | Remetente |
| `ADMIN_EMAIL` | Sim (churn alerts) | `antonio@institutoitinerante.com.br` | Destinatário do alerta diário |
| `LOG_LEVEL` | Não (default info) | `info` | Nível de log |
| `SELF_HOSTED` | Não | `false` | Distingue modo managed de self-hosted |

### Migrations

16 migrations em `migrations/`. Aplicar em ordem sequencial:

```bash
for f in migrations/*.sql; do
  psql "$DATABASE_URL" -f "$f"
done
```

### O que falta para produção plena no modo IIT

1. **Exposição controlada do frontend:** hoje não está no CF Tunnel. Adicionar rota Traefik com Cloudflare Access (OTP) para acesso admin externo.
2. **Banco dedicado:** o Pulse ainda usa `shared-infra`. Criar banco `pulse_db` dedicado no VM 121 para evitar contaminação com outros produtos.
3. **Cron de health score:** `StartHealthScoreJob` é iniciado no startup — confirmar que está funcionando em produção (checar logs `[health-score]`).
4. **Seed de projeto IIT:** a migration 009 cria o projeto `iit` com API key padrão. Substituir pela key real dos produtos (`pulse_iit_default_key_...` é placeholder).

---

## 2. Modo SaaS — Multi-Tenant

Permite que clientes externos criem contas e usem o Pulse como serviço. Requer mudanças arquiteturais além do estado atual.

### Arquitetura Necessária

```
Cliente → Pulse SaaS (cloud) → PostgreSQL (por tenant ou shared com RLS)
                             → Redis (namespace por tenant: pulse:{tenant_id}:flags)
                             → checkout-service (billing via Asaas)
                             → auth-service (SSO ou auth próprio)
```

### Modelo de Dados — Isolamento por Projeto

A tabela `projects` (migration 009) é a base do multi-tenancy. Cada cliente é um `project` com `api_key` única. O isolamento atual é lógico (filtro por `project_id`), não físico.

**Isolamento atual (lógico — suficiente para tier Growth):**
- Feature flags filtradas por `project_id`
- Eventos filtrados por `product` (ainda não por `project_id` — gap)
- Redis: namespace por tenant via prefixo na chave (`pulse:{project_id}:flags`)

**Isolamento necessário para tier Enterprise (físico):**
- Schema separado por tenant no PostgreSQL (PostgreSQL schemas, não databases)
- Ou: database separado por tenant (mais pesado, mas mais isolado)
- Ou: Row-Level Security (RLS) nativo do PostgreSQL — menos overhead operacional

**Recomendação:** implementar RLS para os primeiros 50 clientes. Migrar para schema-per-tenant apenas se um cliente exigir contratualmente.

### Mudanças Necessárias

**Backend:**
- [ ] Adicionar `project_id` à tabela `events` (gap atual — eventos não são scoped por projeto)
- [ ] Adicionar `project_id` à tabela `nps_responses` e `csat_responses`
- [ ] Adicionar `project_id` à tabela `tenant_health_scores`
- [ ] Endpoint de criação de projeto com geração automática de API key (`pulse_<hex32>`)
- [ ] Autenticação SDK via `X-Pulse-Key` (já existe na migration 009, falta middleware)
- [ ] Rate limiting por projeto (Redis token bucket: `pulse:{project_id}:ratelimit`)
- [ ] Middleware que resolve `project_id` a partir da API key em cada request do SDK

**Billing:**
- [ ] Integrar com `checkout-service` para criação de subscription (planos Free/Growth/Pro)
- [ ] Webhook de `payment_confirmed` → ativar projeto / upgrade de plano
- [ ] Webhook de `subscription_cancelled` → downgrade para Free ou desativar
- [ ] Quota enforcement: bloquear ingestão de eventos quando evento/mês ultrapassar limite do plano
- [ ] Overage tracking: contagem de eventos por `project_id` por mês (tabela `project_quotas`)

**Frontend:**
- [ ] Onboarding: tela de criação de projeto + copiar API key
- [ ] Seletor de projeto no dashboard (hoje hardcoded para o IIT)
- [ ] Billing dashboard: consumo do mês, plano atual, link para upgrade

### Variáveis de Ambiente Adicionais (Modo SaaS)

| Variável | Exemplo | Descrição |
|----------|---------|-----------|
| `CHECKOUT_SERVICE_URL` | `https://checkout-api.institutoitinerante.com.br` | URL do checkout-service para billing |
| `CHECKOUT_SERVICE_TOKEN` | `tok_...` | Token de autenticação inter-service |
| `QUOTA_ENFORCEMENT` | `true` | Habilitar bloquei de ingestão por quota |
| `MAX_EVENTS_FREE` | `10000` | Limite de eventos/mês no plano Free |
| `MAX_EVENTS_GROWTH` | `500000` | Limite de eventos/mês no plano Growth |

---

## 3. Modo Self-Hosted

Para clientes que preferem rodar o Pulse em sua própria infra. O arquivo `docker-compose.selfhosted.yml` na raiz do repositório cobre este caso.

### Instalação em 5 Comandos

```bash
# 1. Clone o repositório
git clone https://github.com/institutoitinerante/pulse-service.git && cd pulse-service

# 2. Configure as variáveis de ambiente
cp .env.selfhosted.example .env.selfhosted

# 3. Edite o .env.selfhosted com suas configurações
$EDITOR .env.selfhosted

# 4. Suba os containers
docker compose -f docker-compose.selfhosted.yml up -d

# 5. Aplique as migrations
docker compose -f docker-compose.selfhosted.yml exec pulse-backend \
  sh -c 'for f in /app/migrations/*.sql; do psql "$DATABASE_URL" -f "$f"; done'
```

Após o passo 5, o Pulse estará disponível em:
- Backend API: `http://localhost:3015`
- Frontend: `http://localhost:3016`

### Variáveis Mínimas para Self-Hosted (`.env.selfhosted`)

```bash
SERVICE_TOKEN=seu-token-secreto-aqui
ADMIN_EMAIL=seu@email.com
RESEND_API_KEY=re_...          # opcional — apenas para churn alerts por email
RESEND_FROM=pulse@seudominio.com  # opcional
SELF_HOSTED=true
```

O `DATABASE_URL` e `REDIS_URL` são configurados automaticamente pelo Docker Compose (apontam para os containers `pulse-postgres` e `pulse-redis`). Não definir manualmente a menos que queira usar banco externo.

### Atualização

```bash
git pull origin main
docker compose -f docker-compose.selfhosted.yml pull
docker compose -f docker-compose.selfhosted.yml up -d
# aplicar novas migrations se houver
docker compose -f docker-compose.selfhosted.yml exec pulse-backend \
  sh -c 'for f in /app/migrations/*.sql; do psql "$DATABASE_URL" -f "$f" 2>/dev/null; done'
```

---

## 4. Roadmap de Paridade — LaunchDarkly / VWO / PostHog Core

> **Posicionamento revisado (v2.0):** O Pulse é concorrente direto do LaunchDarkly (feature flags), VWO (A/B testing) e PostHog (product analytics). Ver `COMPETITIVE-ANALYSIS.md` para análise completa.

Lista priorizada do que construir para atingir paridade com os concorrentes diretos. Ordenado por impacto/esforço.

### Tier 1 — MVP competitivo (build em 60 dias)

Mínimo para ser alternativa real ao LaunchDarkly starter para SaaS brasileiros.

| Item | Esforço | Impacto | Justificativa |
|------|---------|---------|---------------|
| **`pulse.js` IIFE via CDN** — reescrever como script tag (não ES module), `track()` real | 3-5 dias | Muito alto | SDK atual usa `export` e tem `track()` no-op. Não funciona como `<script src>`. Bloqueia todo tracking de frontend. |
| **`project_id` em events** — migration + filtros nos handlers | 1 dia | Alto | Gap estrutural: eventos não scoped por projeto. Dados misturados entre projetos. |
| **Ambientes por projeto em flags** — migration + campo `environment` em `feature_flags` | 1 dia | Alto | LaunchDarkly core feature. Migration 013 já fez em events; replicar em flags. |
| **Audit log de flags** — tabela `flag_changes` + UI | 3 dias | Alto | Rastreabilidade de quem ativou qual flag. Requisito para usar flags em produção com segurança. |
| **Análise estatística de experimentos** — p-value + confidence interval (chi-squared em Go) | 5 dias | Alto | Variantes existem mas sem significância calculada = divisão de tráfego, não experimentação. |
| **Middleware de autenticação por API key** — resolver `project_id` via `X-Pulse-Key` | 2 dias | Alto | Migration 009 tem a estrutura; falta enforcement no middleware. Pré-requisito para clientes externos. |

### Tier 2 — Paridade funcional (build em 6 meses)

Paridade com LaunchDarkly core + VWO básico + PostHog analytics essencial.

| Item | Esforço | Impacto | Justificativa |
|------|---------|---------|---------------|
| **`@pulse/react`** — pacote NPM com hooks: `useFlag`, `useTrack`, `useExperiment` | 1 semana | Muito alto | React é o stack padrão do IIT e de 90% dos SaaS BR. DX ruim = baixa adoção. |
| **`@pulse/react-native`** — wrapper com AsyncStorage para Expo/RN | 1 semana | Muito alto | Libri e Nitro são mobile. O Pulse não captura nenhum dado desses apps hoje. |
| **Funis de conversão** — SQL com window functions + UI | 2 semanas | Alto | Funcionalidade #1 em product analytics. Sem funis, não há análise séria de onboarding. |
| **Metric assignment** — vincular experimento a evento de conversão | 1 semana | Alto | Transformar A/B testing em experimentação real de negócio (flag → taxa de conversão). |
| **Rate limiting por projeto** — Redis token bucket | 2 dias | Alto | Sem rate limiting, um cliente abusivo afeta todos no modo SaaS. |
| **Quota enforcement** — bloquear eventos acima do limite do plano | 3 dias | Alto | Pré-requisito para plano Free público sem risco de abuso. |
| **Dashboard por projeto** — seletor + overview isolado | 1 semana | Alto | Hoje hardcoded para o IIT. Para SaaS, cada cliente vê só seus dados. |
| **Webhooks de eventos** — POST para URL configurada pelo cliente | 3 dias | Médio | Integrações com Slack, n8n, sistemas internos do cliente. |
| **Flag stale detection** — cron que marca flags não avaliadas em 30d | 1 dia | Médio | Boa prática, fácil. Evita acúmulo de flags mortas. |
| **SDK Go maduro** — publish em `pkg.go.dev` + exemplos | 2 dias | Médio | SDK existe em `sdk/go/`, falta documentação e versionamento semântico. |

### Tier 3 — Diferenciação (build em 12 meses)

O que o Pulse pode ter que LaunchDarkly e VWO não têm.

| Item | Esforço | Impacto | Justificativa |
|------|---------|---------|---------------|
| **Funis de feature flags** — correlação flag ativa → conversão no funil | 2 semanas | Muito alto | Único no mercado. LD e VWO não têm nativo. Ver se ativar uma flag aumentou conversão. |
| **Health score configurável via UI** — pesos e thresholds por projeto | 1 semana | Alto | Hoje hardcoded em Go. Clientes SaaS precisam customizar para o perfil do seu produto. |
| **Churn prediction por ML leve** — regressão logística sobre health score + eventos | 3 semanas | Alto | Evoluir de threshold fixo para probabilidade calculada. Diferencial B2B. |
| **Schema registry leve** — validação opcional de schema no `POST /ingest` | 1 semana | Médio | Tracking plan governance leve (sem complexidade do Avo). Evita eventos inconsistentes em escala. |
| **Approval flow para flags** — review antes de ativar em prod | 2 semanas | Médio | Para times acima de 15 engenheiros. LaunchDarkly Pro feature. |
| **NPS triggered por evento** — disparar survey após evento customizável | 3 dias | Médio | Hoje trigger é `after_30_days` hardcoded. Clientes precisam triggerar por evento (primeira compra, etc.). |
| **Export de dados** — CSV/JSON de eventos brutos | 3 dias | Médio | Requisito de confiança: clientes querem sair com seus dados se precisar. |
| **Streaming SSE para flag updates** — eliminar polling no SDK | 1 semana | Médio | Reduz latência de propagação de ~30s para <1s. Útil para kill switches emergenciais. |

### Não priorizar

- Session recording / heatmaps — infra pesada (object storage DOM snapshots), ROI baixo
- Data warehouse connectors (BigQuery, Snowflake) — segmento enterprise, fora do foco BR early-stage
- Tracking plan governance completo (Avo-style com branch/PR) — útil apenas com 10+ clientes com times de data separados
- Edge SDK (Cloudflare Workers) — não é demanda do mercado BR early-stage
