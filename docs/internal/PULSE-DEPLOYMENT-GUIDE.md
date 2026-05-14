# Pulse — Guia de Deployment

**Atualizado:** Abril 2026
**Versão Pulse:** 2.0 (migrations 001–036)

Este documento cobre as duas modalidades de uso do Pulse e como evoluir o serviço para suportar as duas bem.

---

## Modalidade A — Cloud Managed (IIT gerencia)

### O que é

O Pulse roda no K3s do IIT. O cliente (produto IIT ou cliente externo) só precisa de uma API key e do SDK. A infraestrutura, banco, atualizações e backups são responsabilidade do IIT.

### O que o cliente configura

Apenas dois valores:

| Item | Exemplo |
|------|---------|
| API key (projectKey) | `pulse_<hex32>` — gerada na migration 009 ou via `POST /api/v1/projects` |
| URL da API | `https://pulse-api.institutoitinerante.com.br` (hoje interno, sem CF Tunnel público) |

Para produtos com frontend web, adicionar o script tag no HTML:

```html
<script src="https://pulse-api.institutoitinerante.com.br/sdk/pulse.js"></script>
<script>
  pulse.init({
    projectKey: 'pulse_SEU_KEY_AQUI',
    apiUrl: 'https://pulse-api.institutoitinerante.com.br'
  })
  pulse.identify('user_uuid', { plan: 'growth', name: 'João' })
</script>
```

Para backends Go, usar o client em `sdk/go/`:

```go
import "github.com/institutoitinerante/pulse-service/sdk/go"

client := pulse.NewClient("http://pulse-service:3015", "pulse_SEU_KEY_AQUI")
enabled := client.IsEnabled(ctx, "nova-feature")
client.TrackEvent(ctx, pulse.Event{Name: "tarefa_criada", UserID: userID})
```

### O que o IIT mantém

- K3s deployment (`deploy/k8s/`)
- PostgreSQL no VM 192.168.30.121 (banco `pulse_db` dedicado — ver gap abaixo)
- Redis no shared-infra (DB 1)
- Aplicação das migrations sequencialmente a cada release
- Secret K8s `pulse-service-secrets` com as variáveis abaixo
- Imagem Docker no GHCR: `ghcr.io/institutoitinerante/pulse-service:latest`

### Variáveis de Ambiente (K3s Secret)

| Variável | Obrigatório | Descrição |
|----------|-------------|-----------|
| `DATABASE_URL` | Sim | `postgres://postgres:senha@192.168.30.121:5432/pulse_db` |
| `REDIS_URL` | Sim | `redis://redis:6379` |
| `SERVICE_TOKEN` | Sim | Token admin para endpoints `/api/v1/*` |
| `RESEND_API_KEY` | Sim | Churn alerts por email |
| `RESEND_FROM` | Sim | `pulse@institutoitinerante.com.br` |
| `ADMIN_EMAIL` | Sim | Destinatário dos alertas diários |
| `PORT` | Nao (default 3015) | Porta HTTP |
| `LOG_LEVEL` | Nao (default info) | `info` ou `debug` |
| `SELF_HOSTED` | Nao | `false` |

### Como o multi-tenant funciona

Cada cliente é um `project` na tabela `projects` (migration 009). O isolamento é lógico via `project_id`:

```
projects
  id (UUID)                ← discriminador
  name                     ← "Nitro", "Libri", "Cliente X"
  slug                     ← "nitro", "libri", "cliente-x"
  api_key (pulse_<hex32>)  ← enviado no header X-Pulse-Key
  plan                     ← free / growth / pro
```

O SDK envia `X-Pulse-Key` em todo request. O middleware resolve o `project_id` a partir da key e injeta no contexto. Todos os handlers filtram por `project_id`.

**Gap atual:** a migration 017 adicionou `project_id` nas tabelas de eventos, mas o middleware de autenticação por API key ainda não está implementado para rotas externas. Hoje o acesso é via `SERVICE_TOKEN` (admin) ou sem autenticação nas rotas públicas de flags. Ver roadmap.

### Diagrama — Cloud Managed

```
                  ┌─────────────────────────────────┐
                  │           PRODUTOS IIT            │
                  │                                   │
                  │  focus-hub ──► pulse SDK Go       │
                  │  my-library ──► pulse SDK Go      │
                  │  browser ──► pulse.js (script tag)│
                  └──────────────────┬────────────────┘
                                     │ HTTP (inter-service)
                                     │ X-Pulse-Key: pulse_xxx
                                     ▼
                  ┌─────────────────────────────────────────────┐
                  │            K3s — VLAN 30 (INTERNAL)          │
                  │                                              │
                  │  ┌─────────────────────┐                    │
                  │  │   pulse-service      │                    │
                  │  │   Go/Fiber :3015     │                    │
                  │  │                      │                    │
                  │  │  handlers/           │                    │
                  │  │   event_handler      │                    │
                  │  │   flag_handler       │                    │
                  │  │   health_score       │                    │
                  │  │   nps / csat         │                    │
                  │  │   funnel             │                    │
                  │  │   experiment         │                    │
                  │  └──────────┬───────────┘                    │
                  │             │                                │
                  │      ┌──────┴──────┐                        │
                  │      ▼             ▼                        │
                  │  PostgreSQL     Redis :6379                  │
                  │  VM 121         shared-infra                 │
                  │  pulse_db       flag cache                   │
                  └─────────────────────────────────────────────┘
                                     │
                                     │ Resend API (churn alerts)
                                     ▼
                              notification-service
                              (futuro: via RabbitMQ)
```

### Checklist — Configuração de Novo Cliente (Cloud Managed)

```
[ ] 1. Criar projeto no Pulse
       POST /api/v1/projects
       Authorization: Bearer $SERVICE_TOKEN
       Body: { "name": "Nome do Produto", "slug": "slug-unico", "plan": "free" }
       Resposta: { "api_key": "pulse_xxxxx" }

[ ] 2. Guardar a api_key em secret do K8s (ou .env do serviço cliente)
       kubectl create secret generic <servico>-secrets \
         --from-literal=PULSE_API_KEY=pulse_xxxxx \
         --from-literal=PULSE_URL=http://pulse-service:3015

[ ] 3. Para backend Go — adicionar client
       client := pulse.NewClient(os.Getenv("PULSE_URL"), os.Getenv("PULSE_API_KEY"))
       // injetar no router/handler via DI

[ ] 4. Para frontend web — adicionar script tag
       <script src="https://pulse-api.institutoitinerante.com.br/sdk/pulse.js"></script>
       pulse.init({ projectKey: os.env.VITE_PULSE_KEY, apiUrl: ... })

[ ] 5. Verificar que eventos chegam
       GET /api/v1/projects/{slug}/events?limit=10
       Authorization: Bearer $SERVICE_TOKEN

[ ] 6. Criar pelo menos um monitor de saúde
       POST /api/v1/monitors
       { "project_id": "...", "name": "API Health", "url": "https://api.seuservico.com/health" }

[ ] 7. Configurar envio de NPS (opcional)
       POST /api/v1/projects/{id}/nps-config
       { "trigger": "after_30_days", "survey_id": "..." }
```

---

## Modalidade B — Self-Hosted

### O que o cliente precisa ter

- Docker 24+ e Docker Compose v2
- Porta 3015 (backend) e 3016 (frontend) disponíveis
- Mínimo 512MB RAM para os containers
- Acesso à internet para pull da imagem GHCR

Opcional mas recomendado para produção:
- PostgreSQL e Redis externos (ao invés dos containers bundled)
- Reverse proxy (nginx, Caddy) na frente para TLS

### Como o cliente configura

```bash
# 1. Clonar o repositório
git clone https://github.com/institutoitinerante/pulse-service.git
cd pulse-service

# 2. Criar arquivo de configuração
cat > .env.selfhosted << 'EOF'
SERVICE_TOKEN=gere-um-token-seguro-aqui
ADMIN_EMAIL=seu@email.com
RESEND_API_KEY=re_...          # opcional: churn alerts por email
RESEND_FROM=pulse@seudominio.com
# DATABASE_URL e REDIS_URL são configurados automaticamente pelo compose
# defina POSTGRES_PASSWORD para senha não-padrão
POSTGRES_PASSWORD=senha-segura-aqui
EOF

# 3. Subir os containers
docker compose -f docker-compose.selfhosted.yml up -d

# 4. Aguardar PostgreSQL ficar healthy e aplicar migrations
docker compose -f docker-compose.selfhosted.yml exec pulse-backend \
  sh -c 'for f in /app/migrations/*.sql; do psql "$DATABASE_URL" -f "$f" 2>/dev/null; done'

# 5. Verificar
curl http://localhost:3015/health
# Esperado: { "status": "ok" }
```

Após o setup:
- Backend API: `http://localhost:3015`
- Dashboard: `http://localhost:3016`

### O que o IIT precisa fornecer

| Item | Estado atual |
|------|-------------|
| Imagem Docker no GHCR | `ghcr.io/institutoitinerante/pulse-service:latest` — disponivel |
| `docker-compose.selfhosted.yml` | Disponivel na raiz do repo |
| Migrations em `/app/migrations/` | Embutidas na imagem |
| Documentação de setup | Este documento |
| Changelog de breaking changes | Pendente — criar `CHANGELOG.md` |
| `.env.selfhosted.example` | Pendente — criar |

### Atualização (self-hosted)

```bash
git pull origin main
docker compose -f docker-compose.selfhosted.yml pull
docker compose -f docker-compose.selfhosted.yml up -d
# Aplicar novas migrations (idempotente — IF NOT EXISTS em todas)
docker compose -f docker-compose.selfhosted.yml exec pulse-backend \
  sh -c 'for f in /app/migrations/*.sql; do psql "$DATABASE_URL" -f "$f" 2>/dev/null; done'
```

### Diagrama — Self-Hosted

```
             INFRAESTRUTURA DO CLIENTE
  ┌────────────────────────────────────────────────────┐
  │                                                    │
  │  ┌────────────────────┐   ┌──────────────────────┐ │
  │  │  pulse-backend     │   │   pulse-frontend     │ │
  │  │  :3015             │   │   :3016 (nginx)      │ │
  │  │  ghcr.io/iit/...   │   │   ghcr.io/iit/...   │ │
  │  └────────┬───────────┘   └──────────────────────┘ │
  │           │                                        │
  │    ┌──────┴──────┐                                 │
  │    ▼             ▼                                 │
  │ pulse-postgres  pulse-redis                        │
  │ :5432 (interno) :6379 (interno)                    │
  │ volume persistente                                 │
  │                                                    │
  └────────────────────────────────────────────────────┘
           │ SDK calls do app do cliente
           ▼
      App do cliente (browser ou backend)
      X-Pulse-Key: pulse_xxx
      POST http://localhost:3015/api/v1/events/ingest
```

### Limitações vs Cloud Managed

| Aspecto | Cloud Managed | Self-Hosted |
|---------|--------------|-------------|
| Configuração inicial | Receber api_key, pronto | Clone + compose + migrations |
| Atualizações | Automaticas (IIT faz deploy) | Manual (`docker compose pull + up`) |
| Backups | IIT responsavel | Cliente responsavel |
| Monitoramento | Pulse monitora a si mesmo | Cliente monitora |
| Churn alerts por email | Sim (Resend configurado) | Requer RESEND_API_KEY próprio |
| Suporte a novas features | Imediato | Depende de atualizar a imagem |
| Isolamento de dados | Logico (project_id) | Fisico (banco proprio) |
| Custo infra | Incluso no plano | Cliente arca com infra |
| Dashboard | Acesso via IP interno IIT | Acesso local em :3016 |
| Compliance / dados na propria infra | Nao | Sim |
| SLA | Conforme plano contratado | Sem SLA (cliente gerencia) |

---

## Tabela Comparativa — Cloud vs Self-Hosted

| Feature | Cloud (Free) | Cloud (Growth) | Cloud (Pro) | Self-Hosted |
|---------|-------------|----------------|-------------|-------------|
| Feature flags | Ate 10 | Ilimitadas | Ilimitadas | Ilimitadas |
| Eventos/mes | 10.000 | 500.000 | 5.000.000 | Sem limite (infra do cliente) |
| Projetos | 1 | 5 | Ilimitados | Ilimitados |
| A/B testing | Nao | Sim | Sim | Sim |
| Health score | Nao | Sim | Sim | Sim |
| NPS / CSAT | Nao | Sim | Sim | Sim |
| Webhooks | Nao | Sim | Sim | Sim |
| Funnels | Nao | Nao | Sim | Sim |
| LLM Observability | Nao | Nao | Sim | Sim |
| Retencao de dados | 30 dias | 90 dias | 1 ano | Definida pelo cliente |
| Suporte | Community | Email | Dedicado | Sem SLA |
| Preco | Gratis | R$99/mes | R$197/mes | Gratis (open source) |

---

## Como o Pulse Precisa Evoluir para Suportar os Dois Modelos Bem

### Gaps Criticos (bloqueiam clientes externos)

**1. Middleware de autenticacao por API key (Tier 1 — implementar antes de qualquer cliente externo)**

Hoje os endpoints do SDK (`POST /api/v1/events/ingest`, `GET /flags`) nao validam a API key. Qualquer um que conhecer a URL consegue ingerir eventos em qualquer projeto.

```go
// O que falta: middleware que resolve project_id a partir do header
func APIKeyMiddleware(db *pgx.Pool) fiber.Handler {
    return func(c *fiber.Ctx) error {
        key := c.Get("X-Pulse-Key")
        if key == "" {
            return c.Status(401).JSON(fiber.Map{"error": "X-Pulse-Key ausente"})
        }
        project, err := repo.GetProjectByAPIKey(c.Context(), key)
        if err != nil || !project.IsActive {
            return c.Status(401).JSON(fiber.Map{"error": "API key invalida"})
        }
        c.Locals("project_id", project.ID)
        c.Locals("project_plan", project.Plan)
        return c.Next()
    }
}
```

**2. `project_id` ausente em events para projetos legados**

Migration 017 adicionou a coluna, mas eventos sem `project_id` ainda existem. O SDK precisa enviar o project_id no corpo do evento. O handler precisa recusar eventos sem project_id quando o middleware de API key estiver ativo.

**3. Endpoint de criacao de projeto via API**

Hoje so e possivel criar projetos via migration SQL. Precisa de `POST /api/v1/projects` protegido pelo SERVICE_TOKEN para que o IIT provisione clientes programaticamente.

### Gaps Operacionais (self-hosted)

**4. `.env.selfhosted.example` ausente**

O `docker-compose.selfhosted.yml` existe mas o `.env.selfhosted.example` nao. Cliente nao sabe quais variaveis sao obrigatorias.

```bash
# Criar em: .env.selfhosted.example
SERVICE_TOKEN=                    # OBRIGATORIO: token secreto para endpoints admin
ADMIN_EMAIL=                      # OBRIGATORIO: email para receber alertas de churn
POSTGRES_PASSWORD=pulse_secret    # Recomendado: mudar do padrao em producao
RESEND_API_KEY=                   # OPCIONAL: para churn alerts por email
RESEND_FROM=pulse@seudominio.com  # OPCIONAL: remetente dos emails
LOG_LEVEL=info                    # OPCIONAL: info ou debug
```

**5. CHANGELOG.md**

Clients self-hosted precisam saber se uma atualizacao tem breaking changes em migrations ou endpoints. Criar `CHANGELOG.md` com formato `## v2.x.x — breaking: ...`.

**6. Verificacao de migrations aplicadas**

O loop de migrations atual e ingênuo — reaplica tudo e suprime erros com `2>/dev/null`. Implementar tabela `schema_migrations` com controle de versao aplicada (padrao golang-migrate).

### Roadmap de Paridade por Prioridade

**Curto prazo (antes do primeiro cliente externo):**

```
[ ] Middleware X-Pulse-Key (autenticacao por projeto)
[ ] POST /api/v1/projects (criacao via API)
[ ] .env.selfhosted.example
[ ] Imagem GHCR do frontend (hoje so backend tem imagem publicada)
[ ] CHANGELOG.md
```

**Medio prazo (primeiros 5 clientes):**

```
[ ] Rate limiting por projeto via Redis token bucket
[ ] Quota enforcement (bloquear ingestao apos limite do plano)
[ ] Dashboard multi-projeto (seletor de projeto no frontend)
[ ] Retencao de dados configuravel por projeto/plano
[ ] golang-migrate para controle de migrations
```

**Longo prazo (escala):**

```
[ ] RLS no PostgreSQL por project_id (hoje so isolamento logico em Go)
[ ] Streaming SSE para flag updates (eliminar polling do SDK)
[ ] @pulse/react e @pulse/react-native no NPM
[ ] Export de dados (CSV/JSON) — requisito de confianca
```

---

## Checklist de Integracao por Servico IIT

### Nitro (focus-hub) — JA INTEGRADO

```
[x] PULSE_URL em env vars
[x] PULSE_PROJECT_KEY em env vars
[x] Client Go em internal/pulse/client.go
[x] Middleware de tracking em internal/pulse/middleware.go
[ ] Adicionar X-Pulse-Key no client (quando middleware de autenticacao estiver pronto)
[ ] Adicionar project_id nos eventos trackados
[ ] Frontend Nitro: adicionar pulse.js + pulse.identify apos login
```

### Libri (my-library) — JA INTEGRADO

```
[x] Client Go em pkg/pulse/client.go
[x] Middleware em internal/middleware/pulse.go
[ ] Adicionar X-Pulse-Key no client
[ ] Frontend Libri: adicionar pulse.js + pulse.identify
[ ] Eventos especificos: book_scanned, reading_session_started, goal_completed
```

### SocialMake — NAO INTEGRADO

```
[ ] Adicionar PULSE_URL e PULSE_PROJECT_KEY em env vars
[ ] Importar sdk/go/ ou criar client interno
[ ] Trackear: post_scheduled, post_published, social_account_connected, tenant_created
[ ] Frontend: pulse.js + identify com tenant_id
[ ] Feature flag: "nova-ia-caption" para controle de rollout
```

### auth-service — NAO INTEGRADO

```
[ ] Trackear: user_registered, login_success, login_failed, otp_sent
[ ] Nao integrar pulse.js no frontend de auth (nao tem frontend proprio)
[ ] Configurar monitor de health: GET /health do auth-service
```

### checkout-service — NAO INTEGRADO

```
[ ] Trackear: checkout_started, payment_confirmed, payment_failed, subscription_created
[ ] Esses eventos sao criticos para funis de conversao — alta prioridade
[ ] Monitor: GET /health
```

### attend-service (Brio) — NAO INTEGRADO

```
[ ] Trackear: conversation_started, lead_qualified, handoff_to_human
[ ] Feature flag: "nova-ia-modelo" para testar modelos diferentes
```

---

## Variaveis de Ambiente — Referencia Completa

### Modo Cloud Managed (K8s Secret)

```yaml
# pulse-service-secrets
DATABASE_URL: postgres://postgres:senha@192.168.30.121:5432/pulse_db?sslmode=disable
REDIS_URL: redis://redis:6379
SERVICE_TOKEN: tok_gere_aqui_com_openssl_rand_hex_32
RESEND_API_KEY: re_...
RESEND_FROM: pulse@institutoitinerante.com.br
ADMIN_EMAIL: antonio@institutoitinerante.com.br
PORT: "3015"
LOG_LEVEL: info
SELF_HOSTED: "false"
```

### Modo Self-Hosted (docker-compose.selfhosted.yml)

```bash
# .env.selfhosted
SERVICE_TOKEN=              # OBRIGATORIO
ADMIN_EMAIL=                # OBRIGATORIO
POSTGRES_PASSWORD=          # recomendado mudar do padrao
RESEND_API_KEY=             # opcional
RESEND_FROM=                # opcional
LOG_LEVEL=info
SELF_HOSTED=true
# DATABASE_URL e REDIS_URL sao injetados automaticamente pelo compose
# para usar banco externo, defina manualmente e remova os services
# pulse-postgres e pulse-redis do compose
```

### Variaveis adicionais para modo SaaS (futuro)

```bash
CHECKOUT_SERVICE_URL=https://checkout-api.institutoitinerante.com.br
CHECKOUT_SERVICE_TOKEN=tok_...
QUOTA_ENFORCEMENT=true
MAX_EVENTS_FREE=10000
MAX_EVENTS_GROWTH=500000
MAX_EVENTS_PRO=5000000
```
