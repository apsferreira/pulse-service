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

## 4. Roadmap de Paridade — PostHog/Mixpanel Core

Lista priorizada do que construir para atingir paridade nas funcionalidades core dos concorrentes mais relevantes. Ordenado por impacto/esforço.

### Tier 1 — Bloqueadores de adoção (build em 90 dias)

| Item | Esforço | Impacto | Justificativa |
|------|---------|---------|---------------|
| **Funis** — SQL com window functions + UI | Alto | Muito alto | Funcionalidade #1 pedida em product analytics. Sem funis, não é analytics sério. |
| **SDK React Native** — wrapper do JS SDK + AsyncStorage | Médio | Muito alto | Libri e Nitro são mobile. Hoje o Pulse não captura dados desses apps. |
| **project_id em events** — adicionar migration + filtros | Baixo | Alto | Gap crítico para multi-tenancy: eventos não estão scoped por projeto ainda. |
| **Audit log de flags** — tabela flag_changes + UI | Médio | Alto | Quem mudou qual flag, quando. Requisito mínimo para usar flags em produção com segurança. |
| **Middleware API key** — resolver project_id no SDK | Baixo | Alto | Pré-requisito para multi-tenant real. X-Pulse-Key já existe na migration 009, falta enforcement. |

### Tier 2 — Paridade funcional (build em 180 dias)

| Item | Esforço | Impacto | Justificativa |
|------|---------|---------|---------------|
| **Flag stale detection** — cron que detecta flags não avaliadas em 30d | Baixo | Médio | Boa prática — LaunchDarkly tem, evita acúmulo de flags mortas. |
| **Webhooks de eventos** — POST para URL configurada pelo cliente | Médio | Médio | Permite que clientes conectem o Pulse a seus sistemas (Slack, n8n, etc.). |
| **Environments por projeto** — staging/prod com flags independentes | Médio | Alto | Migration 013 adiciona `environment` em events, mas falta em flags. LaunchDarkly core feature. |
| **SDK Go maduro** — publish em `pkg.go.dev`, readme, exemplos | Baixo | Médio | SDK Go existe em `sdk/go/` mas precisa de documentação e versionamento. |
| **Rate limiting por projeto** — Redis token bucket | Baixo | Alto | Sem rate limiting, um cliente pode afetar todos os outros no modo SaaS. |
| **Quota enforcement** — bloquear eventos acima do limite do plano | Médio | Alto | Pré-requisito para lançar plano Free público sem risco de abuso. |
| **Dashboard de projeto** — seletor + overview por projeto | Médio | Alto | Hoje o dashboard é hardcoded para o IIT. Para SaaS, cada cliente precisa ver só seus dados. |

### Tier 3 — Diferenciação (build em 12 meses)

| Item | Esforço | Impacto | Justificativa |
|------|---------|---------|---------------|
| **Health score configurável** — pesos e thresholds via UI | Médio | Alto | Hoje hardcoded no Go. Clientes SaaS precisam de health score customizável para seu produto. |
| **Funis de feature flags** — correlação entre flag ativa e conversão | Alto | Alto | Funcionalidade única — ver se ativar uma flag aumentou conversão no funil. Nenhum concorrente tem nativo. |
| **NPS triggered** — disparar survey após evento customizável | Médio | Médio | Hoje trigger é `after_30_days` hardcoded. Clientes precisam triggerar por evento (após primeira compra, etc.). |
| **Approval flow para flags** — request de review antes de ativar em prod | Médio | Médio | LaunchDarkly killer feature para times maiores. Relevante para plano Pro. |
| **API key rotation** — gerar nova key sem downtime | Baixo | Médio | Requisito de segurança para self-hosted e SaaS. |
| **Export de dados** — CSV ou JSON dos eventos brutos | Baixo | Médio | Clientes querem levar dados para fora (migração, BI próprio). |

### Não priorizar

- Session recording / heatmaps — infra pesada, ROI baixo para o estágio atual
- Data warehouse connectors (BigQuery, Snowflake) — segmento enterprise, fora do foco BR
- Tracking plan governance (Avo-style) — útil apenas com 10+ clientes com times de data separados
- Edge SDK (Cloudflare Workers) — não é demanda do mercado BR early-stage
