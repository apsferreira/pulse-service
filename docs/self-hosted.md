# Pulse — Guia de Self-Hosted

> Versão: 2026.Q2 | Stack: Go 1.24 / Fiber + PostgreSQL 15 + Redis 7

---

## Visão Geral

O Pulse é uma plataforma de **product analytics e feature flags** que você instala e opera na sua própria infraestrutura. Não há chamadas para servidores externos — todos os dados ficam no seu ambiente.

### O que o Pulse oferece

| Categoria | Funcionalidade |
|---|---|
| Feature Flags | Rollout percentual, targeting por `user_id`/`tenant_id`, múltiplos ambientes, agendamento com `starts_at`/`ends_at`, workflow de aprovação |
| A/B Testing | Análise frequentista (chi-squared, p-value) e Bayesiana (distribuição Beta + Monte Carlo) |
| Event Tracking | Ingest de eventos com propriedades livres, identificação de usuário, batch de até 500 eventos por chamada |
| Analytics | Funis de conversão, cohorts de retenção, session replay, SQL Explorer |
| Monitoramento | Uptime monitors, anomaly detection, weekly insights narrativos |
| LLM Observability | Métricas por modelo de IA, latência P50/P99, error rate — útil se seu produto usa LLMs |

### Arquitetura

```
┌──────────────────────────────────────────────────────┐
│                   pulse-backend (Go/Fiber)            │
│  REST API  ·  SSE (flag stream)  ·  WebSocket (events)│
│  Porta padrão: 3015                                   │
└──────────┬─────────────────────────┬─────────────────┘
           │                         │
    ┌──────▼──────┐           ┌──────▼──────┐
    │ PostgreSQL  │           │   Redis 7   │
    │ (dados perm)│           │ (cache flags│
    │             │           │  rate-limit)│
    └─────────────┘           └─────────────┘
```

O backend expõe:
- **API REST** para gerenciamento (autenticada via `SERVICE_TOKEN`) e para SDKs (endpoints públicos com `X-Project-Key`)
- **SSE** (`/api/v1/projects/:project_id/flags/stream`) para flag updates em tempo real
- **WebSocket** (`/ws`) para eventos em tempo real no dashboard
- **SDK CDN** servido estaticamente em `/sdk/pulse.js`

### SDKs disponíveis

| SDK | Ambiente | Instalação |
|---|---|---|
| JavaScript / CDN | Qualquer site via `<script>` | `<script src="https://seu-pulse.exemplo.com/sdk/pulse.js">` |
| `@pulse/react` | React + TypeScript | `npm install @pulse/react` |
| `@pulse/react-native` | Expo / React Native | `yarn add @pulse/react-native` |
| Go | Backend Go | `go get github.com/seu-org/pulse-service/sdk/go` |
| Node.js | Backend Node | `npm install @pulse/node` |
| Android | Android (Kotlin) | `sdk/android/` |
| iOS | iOS (Swift) | `sdk/ios/` |

> Todos os SDKs exigem que você forneça a `apiUrl`/`baseURL` apontando para sua instância. Não há URL padrão embutida.

---

## Pré-requisitos

| Requisito | Versão mínima |
|---|---|
| Docker | 24+ |
| Docker Compose | v2.x (`docker compose`) |
| Memória RAM | 512 MB para o backend + infra |
| Disco | 1 GB para dados iniciais (cresce conforme volume de eventos) |

**Domínio próprio** é opcional, mas fortemente recomendado para produção. Sem ele, os SDKs precisarão apontar para um IP diretamente.

---

## Instalação Rápida (Docker Compose)

### Passo 1 — Criar o arquivo de variáveis de ambiente

```bash
cp .env.example .env   # se existir; ou crie manualmente
```

Conteúdo mínimo do `.env` para **modo auth local** (recomendado para self-hosted):

```dotenv
# Obrigatório — token que protege todos os endpoints admin
SERVICE_TOKEN=seu-token-secreto-aqui   # openssl rand -hex 32

# Senha do PostgreSQL
POSTGRES_PASSWORD=troque-esta-senha

# URL pública da sua instância (usada pelo frontend para chamar a API)
PULSE_API_URL=https://pulse-api.seudominio.com

# Modo de autenticação (local = email+senha self-contained, sem dependências externas)
PULSE_AUTH_MODE=local

# E-mail para alertas internos (opcional)
ADMIN_EMAIL=ops@seudominio.com

# Resend (opcional — para e-mails de alertas operacionais)
RESEND_API_KEY=re_xxxxxxxx
RESEND_FROM=noreply@seudominio.com
```

> Gere tokens seguros com: `openssl rand -hex 32`

### Passo 2 — Subir os serviços

```bash
docker compose -f docker-compose.selfhosted.yml up -d
```

Isso sobe quatro containers:

| Container | Porta exposta | Descrição |
|---|---|---|
| `pulse-backend` | `3015` | API REST + SSE + WebSocket |
| `pulse-frontend` | `3016` | Dashboard web (React) |
| `pulse-postgres` | nenhuma | PostgreSQL (isolado na rede interna) |
| `pulse-redis` | nenhuma | Redis (isolado na rede interna) |

### Passo 3 — Verificar o health check

```bash
curl http://localhost:3015/health
# {"status":"ok","service":"pulse-service"}
```

### Passo 4 — Aplicar as migrations

```bash
export PGPASSWORD=troque-esta-senha

for f in $(ls migrations/*.sql | sort); do
  echo "Aplicando $f..."
  docker exec -i pulse-postgres psql -U pulse -d pulse_db < "$f"
done
```

### Passo 5 — Criar o primeiro usuário admin (modo local)

Se `PULSE_AUTH_MODE=local`, crie o primeiro usuário via API — **este endpoint só funciona enquanto não há nenhum usuário cadastrado**:

```bash
curl -X POST http://localhost:3015/api/v1/auth/setup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@suaempresa.com",
    "password": "sua-senha-segura",
    "name": "Admin"
  }'
```

Resposta:

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "...",
    "email": "admin@suaempresa.com",
    "name": "Admin",
    "role": "admin"
  }
}
```

Após criar o primeiro usuário, o endpoint `/api/v1/auth/setup` retorna 403 para novos acessos.

### Passo 6 — Acessar o dashboard

Abra `http://localhost:3016` no browser. Faça login com o e-mail e senha criados no passo anterior.

---

## Modos de Autenticação

O Pulse suporta dois modos de autenticação para o dashboard. Configure via `PULSE_AUTH_MODE`.

### Opção A — Auth Local (recomendado para self-hosted)

Self-contained: email + senha armazenados no próprio banco. Não requer nenhum serviço externo.

```dotenv
PULSE_AUTH_MODE=local
PULSE_JWT_SECRET=seu-secret-jwt-aqui   # openssl rand -hex 32 (opcional — usa SERVICE_TOKEN como fallback)
```

Fluxo:
1. `POST /api/v1/auth/setup` → cria primeiro admin (só funciona se não há usuários)
2. `POST /api/v1/auth/login` → retorna JWT (válido por 24h)
3. `GET /api/v1/auth/me` → valida token e retorna dados do usuário
4. `POST /api/v1/auth/logout` → stateless (cliente descarta o token)

### Opção B — Auth Externo (OAuth via auth-service)

Integração com um auth-service OAuth externo. Adequado para organizações que já têm SSO ou um provedor de identidade próprio.

```dotenv
PULSE_AUTH_MODE=external
PULSE_AUTH_URL=https://auth.suaempresa.com
```

Fluxo: o frontend redireciona para `PULSE_AUTH_URL` que completa o OAuth e emite um token aceito pelo Pulse.

---

## Variáveis de Ambiente

### Backend (`pulse-backend`)

| Variável | Obrigatória | Padrão | Descrição |
|---|---|---|---|
| `SERVICE_TOKEN` | **Sim** | — | Token de acesso a todos os endpoints admin. Mínimo 32 caracteres. Gere com `openssl rand -hex 32`. |
| `DATABASE_URL` | **Sim** | — | Connection string PostgreSQL. Formato: `postgres://user:pass@host:5432/dbname?sslmode=disable` |
| `REDIS_URL` | Não | `redis://localhost:6379/6` | URL do Redis. DB 6 é usado para cache de flags. |
| `PORT` | Não | `3050` | Porta em que o servidor escuta. No compose self-hosted: `3015`. |
| `PULSE_AUTH_MODE` | Não | `external` | Modo de autenticação do dashboard: `local` ou `external`. |
| `PULSE_JWT_SECRET` | Não | *(usa `SERVICE_TOKEN`)* | Secret para assinar JWTs no modo local. Se omitido, usa `SERVICE_TOKEN`. |
| `PULSE_AUTH_URL` | Não | — | URL do auth-service externo. Necessário apenas no modo `external`. |
| `PULSE_BILLING_URL` | Não | — | URL do portal de upgrade exibida quando a quota mensal é atingida (HTTP 429). Deixe vazio para self-hosted sem billing. |
| `PULSE_ALLOWED_ORIGINS` | Não | `*` | Origens CORS permitidas (CSV). Padrão `*` é mantido porque o SDK é embeddable em qualquer site. |
| `LOG_LEVEL` | Não | `info` | Nível de log: `debug`, `info`, `warn`, `error`. |
| `ADMIN_EMAIL` | Não | — | E-mail para alertas de churn crítico e notificações operacionais. |
| `RESEND_API_KEY` | Não | — | API key do Resend para e-mails transacionais. |
| `RESEND_FROM` | Não | — | Endereço de remetente. Ex: `noreply@seudominio.com`. |

### Frontend (`pulse-frontend`)

| Variável | Obrigatória | Padrão | Descrição |
|---|---|---|---|
| `VITE_API_URL` | Não | `/api/v1` (relativo) | URL base da API. Deve ser a URL pública acessível pelo browser. Ex: `https://pulse-api.seudominio.com` |
| `VITE_AUTH_MODE` | Não | `external` | Espelha `PULSE_AUTH_MODE`. Define qual UI de login exibir. |
| `VITE_AUTH_SERVICE_URL` | Não | — | URL do auth-service externo. Necessário apenas no modo `external`. |

---

## Configuração dos SDKs

> Todos os SDKs exigem a URL da sua instância Pulse. Não há fallback para nenhuma URL externa.

### JavaScript / CDN

```html
<script src="https://pulse-api.seudominio.com/sdk/pulse.js"></script>
<script>
  pulse.init({
    projectKey: 'pk_live_xxxxxxxxxxxxxxxxxx',
    apiUrl: 'https://pulse-api.seudominio.com',
    userId: 'user-123',
    debug: false,
  })

  if (pulse.isEnabled('nova-feature')) {
    // mostra nova feature
  }

  pulse.track('page_viewed', { page: 'home' })
</script>
```

### React

```bash
npm install @pulse/react
```

```tsx
import { PulseProvider, usePulseFlag, useTrack } from '@pulse/react'

// main.tsx
<PulseProvider
  projectKey="pk_live_xxxxxxxxxxxxxxxxxx"
  apiUrl="https://pulse-api.seudominio.com"
  userId={currentUser?.id}
  defaults={{ 'nova-feature': false }}
  failOpen={true}
>
  <App />
</PulseProvider>

// Componente
function MyComponent() {
  const novaFeature = usePulseFlag('nova-feature', false)
  const track = useTrack()
  return novaFeature ? <V2 /> : <V1 />
}
```

### React Native / Expo

```bash
yarn add @pulse/react-native @react-native-async-storage/async-storage
```

```tsx
import { PulseProvider, useFeatureFlag } from '@pulse/react-native'

<PulseProvider
  config={{
    apiUrl: 'https://pulse-api.seudominio.com',
    projectKey: 'pk_live_xxxxxxxxxxxxxxxxxx',
  }}
>
  <RootNavigator />
</PulseProvider>
```

### Go

```bash
go get github.com/seu-org/pulse-service/sdk/go
```

```go
import pulse "github.com/seu-org/pulse-service/sdk/go"

client, _ := pulse.NewClient("https://pulse-api.seudominio.com")

enabled, err := client.IsEnabled(ctx, "nova-feature")
if err != nil {
    enabled = false // fail-open
}

client.TrackEvent(ctx, pulse.TrackEventRequest{
    EventName: "user_signup",
    UserID:    &userID,
    Properties: map[string]interface{}{"plan": "free"},
})
```

### Node.js

```bash
npm install @pulse/node
```

```js
const { PulseClient } = require('@pulse/node')

const pulse = new PulseClient({
  baseURL: 'https://pulse-api.seudominio.com',
  projectKey: 'pk_live_xxxxxxxxxxxxxxxxxx',
})

const enabled = await pulse.isEnabled('nova-feature', { userId: 'user-123' })
```

---

## Criando o Primeiro Projeto

Após subir e criar o primeiro usuário admin, crie um projeto:

```bash
curl -X POST http://localhost:3015/api/v1/projects \
  -H "Content-Type: application/json" \
  -H "X-Service-Token: seu-token-secreto-aqui" \
  -d '{
    "name": "Meu Produto",
    "slug": "meu-produto",
    "plan": "free"
  }'
```

Resposta:

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Meu Produto",
  "slug": "meu-produto",
  "api_key": "pk_live_xxxxxxxxxxxxxxxxxx",
  "plan": "free",
  "created_at": "2026-01-01T00:00:00Z"
}
```

Guarde o `api_key` — esse é o `X-Project-Key` usado nos SDKs e nos endpoints de ingest de eventos.

### Estrutura de dados

```
Usuário (admin)
  └── Projeto A (api_key: pk_live_aaa...)
        ├── Feature Flag: nova-homepage
        ├── Feature Flag: dark-mode
        └── Feature Flag: checkout-v2
  └── Projeto B (api_key: pk_live_bbb...)
        └── Feature Flag: ...
```

Cada projeto tem sua própria `api_key` e seus dados são isolados.

---

## Feature Flags

### Criar uma flag

```bash
curl -X POST http://localhost:3015/api/v1/flags \
  -H "Content-Type: application/json" \
  -H "X-Service-Token: seu-token-secreto-aqui" \
  -d '{
    "key": "nova-homepage",
    "description": "Nova versão da homepage",
    "enabled": false,
    "project_id": "550e8400-e29b-41d4-a716-446655440000",
    "rollout_percentage": 0
  }'
```

### Avaliar flags via API pública

```bash
# Avalia todas as flags para um usuário (sem autenticação, usa X-Project-Key)
GET /flags?user_id=user-123&project_id=550e8400-...
X-Project-Key: pk_live_xxxxxxxxxxxxxxxxxx
```

### Atualizações em tempo real (SSE)

```
GET /api/v1/projects/:project_id/flags/stream
X-Service-Token: seu-token-secreto-aqui
```

O Pulse notifica clientes imediatamente quando uma flag muda, sem polling. Heartbeats a cada 30s mantêm a conexão viva.

---

## Segurança

### Tokens

| Token | Header | Escopo |
|---|---|---|
| `SERVICE_TOKEN` | `X-Service-Token` | Endpoints admin: CRUD de flags, projetos, dashboard |
| `api_key` do projeto | `X-Project-Key` | Endpoints públicos do SDK: avaliação de flags, ingest de eventos |

### Recomendações para produção

1. **Use HTTPS obrigatoriamente.** Configure um reverse proxy (Nginx, Traefik, Caddy) com certificado SSL.
2. **Nunca exponha o `SERVICE_TOKEN` no frontend.** Apenas em variáveis server-side ou secrets do Kubernetes.
3. **Restrinja CORS em produção:**
   ```dotenv
   PULSE_ALLOWED_ORIGINS=https://app.seudominio.com,https://www.seudominio.com
   ```
4. **Não exponha PostgreSQL e Redis externamente.** As portas estão comentadas no compose intencionalmente.
5. **Rotacione o `SERVICE_TOKEN` periodicamente** e atualize a variável de ambiente.

### Configuração Nginx para SSE

```nginx
server {
    listen 443 ssl;
    server_name pulse-api.seudominio.com;

    location / {
        proxy_pass http://localhost:3015;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_buffering off;
        proxy_cache off;
        proxy_read_timeout 3600s;
        add_header X-Accel-Buffering no;
    }
}
```

---

## Atualização

```bash
# Baixar nova imagem
docker pull ghcr.io/seu-org/pulse-service:latest
docker pull ghcr.io/seu-org/pulse-frontend:latest

# Aplicar novas migrations (verifique o CHANGELOG)
for f in migrations/NEW_*.sql; do
  docker exec -i pulse-postgres psql -U pulse -d pulse_db < "$f"
done

# Reiniciar
docker compose -f docker-compose.selfhosted.yml up -d
```

Em produção, prefira tags semânticas em vez de `latest`:

```yaml
image: ghcr.io/seu-org/pulse-service:v2026.04
```

---

## Troubleshooting

### "SERVICE_TOKEN is required" ao subir

O `SERVICE_TOKEN` é obrigatório. Verifique se o `.env` existe e está no mesmo diretório do compose:

```bash
docker compose -f docker-compose.selfhosted.yml config | grep SERVICE_TOKEN
```

### Login local não funciona — "invalid credentials"

Verifique se `PULSE_AUTH_MODE=local` está definido no backend **e** `VITE_AUTH_MODE=local` no frontend. Se o frontend estiver em modo `external`, ele redireciona para o auth-service ao invés de exibir o formulário de login.

### Setup retorna 403 — "setup already completed"

O endpoint `/api/v1/auth/setup` só funciona enquanto a tabela `pulse_users` está vazia. Se você já criou um usuário, use `/api/v1/auth/login` normalmente.

Para resetar (ambiente de teste):

```bash
docker exec pulse-postgres psql -U pulse -d pulse_db -c "TRUNCATE pulse_users;"
```

### Migrations falham — "relation already exists"

Verifique qual foi a última migration aplicada e aplique apenas a partir da que está faltando:

```bash
docker exec pulse-postgres psql -U pulse -d pulse_db \
  -c "SELECT tablename FROM pg_tables WHERE schemaname='public' ORDER BY tablename;"
```

### CORS bloqueando requests do frontend

Para debug temporário: `PULSE_ALLOWED_ORIGINS=*`. Para produção, liste os domínios exatos. Reinicie após alterar:

```bash
docker compose -f docker-compose.selfhosted.yml restart pulse-backend
```

### SSE desconectando a cada 60s (proxy)

Configure o timeout do proxy: **Nginx** `proxy_read_timeout 3600s;`, **Traefik** ajuste middleware de timeout, **Cloudflare** (plano Free) tem limite de 100s — use `X-Accel-Buffering: no`.

### Logs detalhados

```bash
docker logs -f pulse-backend

# Ativar debug completo
LOG_LEVEL=debug docker compose -f docker-compose.selfhosted.yml up pulse-backend
```

### Backup dos dados

```bash
# Dump
docker exec pulse-postgres pg_dump -U pulse pulse_db > backup-$(date +%Y%m%d).sql

# Restaurar
docker exec -i pulse-postgres psql -U pulse -d pulse_db < backup-20260101.sql
```

---

## Suporte

- **Documentação completa:** `docs/` neste repositório
- **Issues:** abra uma issue no GitHub com o label `self-hosted`
- **Suporte comercial (SLA):** disponível nos planos Self-Hosted Enterprise

Para reportar um problema, inclua:
1. Versão da imagem em uso (`docker inspect pulse-backend | grep Image`)
2. Saída do `docker logs pulse-backend` (últimas 100 linhas)
3. Resultado do `GET /health`
4. Sistema operacional e versão do Docker
