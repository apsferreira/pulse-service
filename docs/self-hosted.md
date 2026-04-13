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

| SDK | Ambiente | Caminho no repositório |
|---|---|---|
| JavaScript / CDN | Qualquer site via `<script>` | `sdk/pulse.js` |
| `@pulse/react` | React + TypeScript | `sdk/react/` |
| `@pulse-iit/react-native` | Expo / React Native | `sdk/react-native/` |
| Go | Backend Go | `sdk/go/` |
| Node.js | Backend Node | `sdk/node/` |
| Python | Backend Python | `sdk/python/` |

---

## Pré-requisitos

| Requisito | Versão mínima |
|---|---|
| Docker | 24+ |
| Docker Compose | v2.x (`docker compose`) |
| PostgreSQL | 15+ |
| Redis | 7+ |
| Memória RAM | 512 MB para o backend + infra |
| Disco | 1 GB para dados iniciais (cresce conforme volume de eventos) |

**Domínio próprio** é opcional, mas fortemente recomendado para produção. Sem ele, os SDKs precisarão apontar para um IP diretamente.

---

## Instalação Rápida (Docker Compose)

### Passo 1 — Criar o arquivo de variáveis de ambiente

```bash
cp .env.example .env   # se existir; ou crie manualmente
```

Conteúdo mínimo do `.env`:

```dotenv
# Obrigatório — token que protege todos os endpoints admin
SERVICE_TOKEN=seu-token-secreto-aqui

# Senha do PostgreSQL (gerada automaticamente no compose)
POSTGRES_PASSWORD=troque-esta-senha

# URL pública da sua instância (usada pelo frontend)
PULSE_API_URL=https://pulse-api.seudominio.com.br

# E-mail para alertas internos (opcional)
ADMIN_EMAIL=ops@seudominio.com.br

# Resend (opcional — necessário para e-mails de churn alert e onboarding)
RESEND_API_KEY=re_xxxxxxxx
RESEND_FROM=noreply@seudominio.com.br
```

> Gere um `SERVICE_TOKEN` seguro com: `openssl rand -hex 32`

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

O Pulse usa migrations SQL manuais. Aplique na ordem:

```bash
# Variáveis para o psql
export PGPASSWORD=troque-esta-senha

for f in $(ls migrations/*.sql | sort); do
  echo "Aplicando $f..."
  docker exec -i pulse-postgres psql -U pulse -d pulse_db < "$f"
done
```

As migrations são numeradas sequencialmente (`001_create_feature_flags.sql` → `041_...`). Nunca pule nenhuma.

---

## Docker Compose Completo

O arquivo `docker-compose.selfhosted.yml` na raiz do repositório já está pronto para uso. Abaixo o conteúdo para referência:

```yaml
version: "3.9"

services:
  pulse-backend:
    image: ghcr.io/institutoitinerante/pulse-service:latest
    container_name: pulse-backend
    restart: unless-stopped
    ports:
      - "3015:3015"
    environment:
      DATABASE_URL: postgres://pulse:${POSTGRES_PASSWORD:-pulse_secret}@pulse-postgres:5432/pulse_db?sslmode=disable
      REDIS_URL: redis://pulse-redis:6379
      PORT: "3015"
      SERVICE_TOKEN: ${SERVICE_TOKEN:?SERVICE_TOKEN is required}
      ADMIN_EMAIL: ${ADMIN_EMAIL:-}
      RESEND_API_KEY: ${RESEND_API_KEY:-}
      RESEND_FROM: ${RESEND_FROM:-}
      LOG_LEVEL: ${LOG_LEVEL:-info}
      SELF_HOSTED: "true"
    depends_on:
      pulse-postgres:
        condition: service_healthy
      pulse-redis:
        condition: service_healthy
    networks:
      - pulse-net
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:3015/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  pulse-frontend:
    image: ghcr.io/institutoitinerante/pulse-frontend:latest
    container_name: pulse-frontend
    restart: unless-stopped
    ports:
      - "3016:80"
    environment:
      VITE_API_URL: ${PULSE_API_URL:-http://localhost:3015}
    depends_on:
      - pulse-backend
    networks:
      - pulse-net

  pulse-postgres:
    image: postgres:15-alpine
    container_name: pulse-postgres
    restart: unless-stopped
    environment:
      POSTGRES_DB: pulse_db
      POSTGRES_USER: pulse
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:-pulse_secret}
    volumes:
      - pulse-postgres-data:/var/lib/postgresql/data
    networks:
      - pulse-net
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U pulse -d pulse_db"]
      interval: 10s
      timeout: 5s
      retries: 5

  pulse-redis:
    image: redis:7-alpine
    container_name: pulse-redis
    restart: unless-stopped
    command: redis-server --save 60 1 --loglevel warning
    volumes:
      - pulse-redis-data:/data
    networks:
      - pulse-net
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  pulse-postgres-data:
  pulse-redis-data:

networks:
  pulse-net:
    driver: bridge
```

---

## Variáveis de Ambiente

### Backend (`pulse-backend`)

| Variável | Obrigatória | Padrão | Descrição |
|---|---|---|---|
| `DATABASE_URL` | Sim | — | Connection string PostgreSQL. Formato: `postgres://user:pass@host:5432/dbname?sslmode=disable` |
| `SERVICE_TOKEN` | Sim | — | Token de acesso a todos os endpoints admin (`/api/v1/*`). Mínimo 32 caracteres. |
| `REDIS_URL` | Não | `redis://localhost:6379/6` | URL do Redis. O DB 6 é usado por padrão para cache de flags. |
| `PORT` | Não | `3050` | Porta em que o servidor escuta. No compose self-hosted usamos `3015`. |
| `LOG_LEVEL` | Não | `info` | Nível de log: `debug`, `info`, `warn`, `error`. |
| `PULSE_ALLOWED_ORIGINS` | Não | `*` | Origens CORS permitidas (CSV). Ex: `https://app.seudominio.com.br,https://admin.seudominio.com.br`. O padrão `*` é mantido porque o SDK é embeddable em qualquer site. |
| `ADMIN_EMAIL` | Não | — | E-mail para alertas de churn crítico e notificações operacionais. |
| `RESEND_API_KEY` | Não | — | API key do Resend para envio de e-mails transacionais (alertas de churn, onboarding). |
| `RESEND_FROM` | Não | — | Endereço de remetente para os e-mails. Ex: `noreply@seudominio.com.br`. |
| `SELF_HOSTED` | Não | — | Quando `"true"`, desabilita métricas de quota e billing. |

### Frontend (`pulse-frontend`)

| Variável | Obrigatória | Padrão | Descrição |
|---|---|---|---|
| `VITE_API_URL` | Sim | `http://localhost:3015` | URL base da API. Deve ser a URL pública acessível pelo browser. |

---

## Primeiro Acesso

### 1. Criar o primeiro projeto

Com os containers rodando, crie um projeto via API:

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

Guarde o `api_key` — esse é o `X-Project-Key` que vai nos SDKs.

### 2. Acessar o dashboard

Abra `http://localhost:3016` no browser. Use o `SERVICE_TOKEN` definido no `.env` para autenticar.

### 3. Criar a primeira feature flag

Via dashboard ou via API:

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

---

## Configuração dos SDKs

### JavaScript / CDN

Não requer bundler nem instalação. Adicione no `<head>` do seu site:

```html
<script src="https://pulse-api.seudominio.com.br/sdk/pulse.js"></script>
<script>
  pulse.init({
    projectKey: 'pk_live_xxxxxxxxxxxxxxxxxx',
    apiUrl: 'https://pulse-api.seudominio.com.br',
    userId: 'user-123',       // opcional
    debug: false,
  })

  // Verificar flag
  if (pulse.isEnabled('nova-homepage')) {
    document.getElementById('hero').className = 'hero-v2'
  }

  // Rastrear evento
  pulse.track('page_viewed', { page: 'home', referrer: document.referrer })
</script>
```

### React

```bash
npm install @pulse/react
# ou
yarn add @pulse/react
```

```tsx
// main.tsx
import { PulseProvider } from '@pulse/react'

ReactDOM.createRoot(document.getElementById('root')!).render(
  <PulseProvider
    projectKey="pk_live_xxxxxxxxxxxxxxxxxx"
    apiUrl="https://pulse-api.seudominio.com.br"
    userId={currentUser?.id}
    defaults={{ 'nova-homepage': false }}
    failOpen={true}
  >
    <App />
  </PulseProvider>
)

// Em qualquer componente
import { usePulseFlag, useTrack, usePulseVariant } from '@pulse/react'

function Homepage() {
  const novaHomepage = usePulseFlag('nova-homepage', false)
  const variant = usePulseVariant('pricing-test', 'control')
  const track = useTrack()

  useEffect(() => {
    track('page_viewed', { page: 'home' })
  }, [])

  return novaHomepage ? <HomepageV2 /> : <HomepageV1 />
}
```

**Opções do `PulseProvider`:**

| Prop | Tipo | Descrição |
|---|---|---|
| `projectKey` | `string` | API key do projeto (`pk_live_...`) |
| `apiUrl` | `string` | URL base do backend Pulse |
| `userId` | `string?` | ID do usuário autenticado |
| `defaults` | `Record<string, boolean>?` | Valores padrão por flag (usados no fail-open) |
| `failOpen` | `boolean?` | Retorna defaults em vez de erro quando backend indisponível (padrão: `true`) |
| `timeout` | `number?` | Timeout de requests em ms (padrão: `2000`) |
| `debug` | `boolean?` | Logs detalhados no console |

### React Native / Expo

```bash
yarn add @pulse-iit/react-native @react-native-async-storage/async-storage
```

```tsx
// App.tsx
import { PulseProvider, useFeatureFlag, useTrack } from '@pulse-iit/react-native'

export default function App() {
  return (
    <PulseProvider
      config={{
        apiUrl: 'https://pulse-api.seudominio.com.br',
        projectKey: 'pk_live_xxxxxxxxxxxxxxxxxx',
      }}
    >
      <RootNavigator />
    </PulseProvider>
  )
}

// Num screen qualquer
function OnboardingScreen() {
  const { value: novoOnboarding, loading } = useFeatureFlag('novo-onboarding')
  const track = useTrack()

  if (loading) return <LoadingSpinner />

  return novoOnboarding ? <OnboardingV2 /> : <OnboardingV1 />
}
```

### Go

```bash
go get github.com/institutoitinerante/pulse-service/sdk/go
```

```go
import pulse "github.com/institutoitinerante/pulse-service/sdk/go"

func main() {
    client, err := pulse.NewClient("https://pulse-api.seudominio.com.br")
    if err != nil {
        log.Fatal(err)
    }

    // Verificar flag
    enabled, err := client.IsEnabled(ctx, "nova-homepage")
    if err != nil {
        // fail-open: use o valor padrão
        enabled = false
    }

    // Rastrear evento
    _, err = client.TrackEvent(ctx, pulse.TrackEventRequest{
        EventName:   "user_signup",
        UserID:      &userID,
        Environment: "production",
        Properties: map[string]interface{}{
            "plan": "free",
        },
    })
}
```

---

## Feature Flags — Como Funciona

### Criação e configuração

Uma flag é criada com uma chave única (ex: `nova-homepage`), associada a um projeto e pode ter:

- `enabled`: liga/desliga global
- `rollout_percentage`: 0–100, determina a fatia de usuários que vê a flag como `true`
- `targeting`: regras de targeting — force `true` ou `false` para `user_id`s ou `tenant_id`s específicos
- `variants`: configuração de A/B testing (ex: `["control", "challenger"]`)
- `starts_at` / `ends_at`: ativação e desativação automática agendadas
- `require_approval`: quando `true`, toda mudança na flag entra num workflow de aprovação antes de ser aplicada

### Rollout percentual

O rollout é **determinístico por `user_id`**: o mesmo usuário sempre cai no mesmo balde. O algoritmo usa SHA-256 do `user_id` para derivar um número de 0–100.

Isso garante que um usuário que vê `nova-homepage=true` continuará vendo mesmo que a página seja recarregada ou que ele troque de dispositivo.

```
hash = SHA256(user_id + flag_key)
bucket = hash % 100
enabled = bucket < rollout_percentage
```

### Avaliação de flags via API pública

```bash
# Avalia todas as flags para um usuário específico
GET /flags?user_id=user-123&project_id=550e8400-...

# Resposta
{"flags": {"nova-homepage": true, "dark-mode": false}}
```

```bash
# Avalia um lote de flags específicas (SDKs usam isso no init)
POST /flags/evaluate-batch
{
  "keys": ["nova-homepage", "dark-mode"],
  "user_id": "user-123",
  "tenant_id": "acme-corp",
  "project_id": "550e8400-..."
}
```

### Atualizações em tempo real via SSE

O Pulse expõe um endpoint SSE que notifica clientes imediatamente quando uma flag é atualizada — sem polling.

```
GET /api/v1/projects/:project_id/flags/stream
Header: X-Service-Token: seu-token-secreto-aqui
```

Formato dos eventos:

```
: connected project_id=550e8400-...

data: {"flag_key":"nova-homepage","enabled":true,"updated_at":"2026-01-01T12:00:00Z"}

data: {"type":"heartbeat"}
```

O heartbeat é enviado a cada 30 segundos para manter a conexão viva através de proxies e load balancers.

**Uso com JavaScript:**

```javascript
const evtSource = new EventSource(
  `https://pulse-api.seudominio.com.br/api/v1/projects/${projectId}/flags/stream`,
  { headers: { 'X-Service-Token': serviceToken } }
)

evtSource.onmessage = (e) => {
  const update = JSON.parse(e.data)
  if (update.type === 'heartbeat') return
  console.log(`Flag ${update.flag_key} agora é ${update.enabled}`)
  // Atualiza estado local sem recarregar a página
}
```

### Variants A/B

```bash
GET /flags/variants?user_id=user-123&project_id=550e8400-...

# Resposta
{"variants": {"pricing-test": "challenger", "onboarding-flow": "control"}}
```

A atribuição de variante também é determinística por `user_id`.

---

## Integração com seu Sistema

### Enviar eventos

O endpoint de events é **público** — não requer `SERVICE_TOKEN`. Use o `X-Project-Key` do projeto.

```bash
# Evento simples
POST /api/v1/events
Content-Type: application/json
X-Project-Key: pk_live_xxxxxxxxxxxxxxxxxx

{
  "event_name": "checkout_completed",
  "user_id": "user-123",
  "session_id": "sess-abc",
  "environment": "production",
  "properties": {
    "plan": "growth",
    "amount_cents": 79900,
    "coupon": "LAUNCH25"
  }
}
```

```bash
# Batch de eventos (flush de fila do SDK — até 500 por chamada)
POST /api/v1/events/batch
Content-Type: application/json

{
  "events": [
    {"event_name": "page_viewed", "user_id": "user-123", ...},
    {"event_name": "button_clicked", "user_id": "user-123", ...}
  ]
}
```

```bash
# Identificar usuário (merge anonymous_id → user_id após login)
POST /api/v1/events/identify

{
  "anonymous_id": "anon-xyz",
  "user_id": "user-123",
  "traits": {
    "name": "Maria Silva",
    "email": "maria@empresa.com",
    "plan": "free"
  }
}
```

### Avaliar flags no backend

Para aplicações server-side que precisam avaliar flags sem usar os SDKs:

```bash
GET /flags?user_id=user-123&project_id=550e8400-e29b-41d4-a716-446655440000

# Ou para uma flag específica (endpoint admin)
GET /api/v1/flags/nova-homepage/evaluate?user_id=user-123
Header: X-Service-Token: seu-token-secreto-aqui
```

---

## Gerenciamento de Projetos

### Criar projeto

```bash
POST /api/v1/projects
X-Service-Token: seu-token-secreto-aqui

{
  "name": "Meu SaaS",
  "slug": "meu-saas",
  "plan": "free"
}
```

### Listar projetos

```bash
GET /api/v1/projects
X-Service-Token: seu-token-secreto-aqui
```

### Regenerar API key

Útil quando uma chave é comprometida:

```bash
POST /api/v1/projects/meu-saas/regenerate-key
X-Service-Token: seu-token-secreto-aqui
```

### Ambientes

O Pulse suporta múltiplos ambientes por projeto (ex: `staging`, `production`):

```bash
POST /api/v1/environments
X-Service-Token: seu-token-secreto-aqui

{
  "name": "Staging",
  "slug": "staging",
  "project_id": "550e8400-..."
}
```

Ao enviar eventos, passe `"environment": "staging"` no payload. Flags e métricas ficam isolados por ambiente.

---

## Segurança

### Modelo de autenticação

O Pulse usa dois tokens com escopos diferentes:

| Token | Header | Escopo |
|---|---|---|
| `SERVICE_TOKEN` | `X-Service-Token` | Endpoints admin: CRUD de flags, projetos, consultas de dados, SSE stream |
| `X-Project-Key` (`api_key` do projeto) | `X-Project-Key` | Endpoints públicos do SDK: avaliação de flags, ingest de eventos |

Os endpoints públicos (sem autenticação ou com apenas `X-Project-Key`) são:

- `GET /flags` — avaliação de flags
- `POST /flags/evaluate-batch` — batch evaluation
- `GET /flags/variants` — variants A/B
- `POST /api/v1/events` — ingest de eventos
- `POST /api/v1/events/batch` — batch ingest
- `POST /api/v1/events/track` — simplified tracking
- `POST /api/v1/events/identify` — merge de identidade
- `GET /health` — health check
- `GET /sdk/*` — arquivos do SDK via CDN

### Recomendações para produção

1. **Use HTTPS obrigatoriamente.** O `SERVICE_TOKEN` trafega em headers e deve ser protegido por TLS. Configure um reverse proxy (Nginx, Traefik, Caddy) com certificado SSL.

2. **Nunca exponha o `SERVICE_TOKEN` no frontend.** Ele fica apenas em ambientes server-side, variáveis de ambiente CI/CD, ou secrets do Kubernetes.

3. **Restrinja CORS em produção.** Defina `PULSE_ALLOWED_ORIGINS` com os domínios exatos do seu frontend:
   ```
   PULSE_ALLOWED_ORIGINS=https://app.seudominio.com.br,https://www.seudominio.com.br
   ```

4. **Rate limiting.** O Pulse usa Redis sliding window para rate limiting automático. Por padrão é configurado para tolerar tráfego normal de SDKs. Ajuste conforme seu volume.

5. **Não exponha o PostgreSQL e o Redis externamente.** No compose, as portas dessas dependências estão comentadas intencionalmente.

6. **Rotacione o `SERVICE_TOKEN` periodicamente.** Ao rotacionar, atualize a variável de ambiente e reinicie o container.

### Configuração de Nginx para produção

```nginx
server {
    listen 443 ssl;
    server_name pulse-api.seudominio.com.br;

    ssl_certificate /etc/ssl/certs/seudominio.crt;
    ssl_certificate_key /etc/ssl/private/seudominio.key;

    location / {
        proxy_pass http://localhost:3015;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;

        # Necessário para SSE (Server-Sent Events)
        proxy_buffering off;
        proxy_cache off;
        proxy_read_timeout 3600s;
        add_header X-Accel-Buffering no;
    }
}
```

---

## Atualização

### Atualizar para a versão mais recente

```bash
# Baixar nova imagem
docker pull ghcr.io/institutoitinerante/pulse-service:latest
docker pull ghcr.io/institutoitinerante/pulse-frontend:latest

# Aplicar novas migrations (se houver)
# Verifique o CHANGELOG ou as migrations adicionadas desde sua versão atual
for f in $(ls migrations/NEW_*.sql | sort); do
  docker exec -i pulse-postgres psql -U pulse -d pulse_db < "$f"
done

# Reiniciar os containers
docker compose -f docker-compose.selfhosted.yml up -d
```

### Verificar a versão atual

```bash
docker exec pulse-backend ./pulse-service --version 2>/dev/null || \
  curl -s http://localhost:3015/health | jq .
```

### Estratégia de versões

As imagens são publicadas com tags semânticas (`v2026.04`, `latest`). Em produção, prefira tags específicas em vez de `latest` para ter controle sobre qual versão está rodando.

```yaml
# docker-compose.selfhosted.yml — troque :latest por uma tag específica
image: ghcr.io/institutoitinerante/pulse-service:v2026.04
```

---

## Troubleshooting

### Serviço não inicia — "DATABASE_URL environment variable is required"

O container aborta na inicialização se `DATABASE_URL` não estiver definida. Verifique:

```bash
docker compose -f docker-compose.selfhosted.yml config | grep DATABASE_URL
```

Se o valor estiver vazio, verifique se o arquivo `.env` existe na mesma pasta do compose e se `SERVICE_TOKEN` também está preenchido (obrigatório).

### Serviço não inicia — "SERVICE_TOKEN environment variable is required"

Igual ao anterior — `SERVICE_TOKEN` também é obrigatório.

### Erro de conexão com o Redis — "failed to parse REDIS_URL"

O formato esperado é `redis://host:port/db` ou `redis://:password@host:port/db`. Evite espaços ou caracteres especiais sem encoding.

### Migrations falham — "relation already exists"

As migrations são idempotentes quando possível, mas se aplicadas parcialmente podem gerar conflito. Verifique qual foi a última migration aplicada:

```bash
docker exec pulse-postgres psql -U pulse -d pulse_db \
  -c "SELECT tablename FROM pg_tables WHERE schemaname='public' ORDER BY tablename;"
```

Compare com os arquivos em `migrations/` e aplique apenas a partir da que está faltando.

### CORS bloqueando requests do frontend

Se o browser mostra erros de CORS, verifique `PULSE_ALLOWED_ORIGINS`. Para debug temporário, defina `*`. Para produção, liste os domínios exatos:

```
PULSE_ALLOWED_ORIGINS=https://app.seudominio.com.br
```

Reinicie o container após alterar a variável:

```bash
docker compose -f docker-compose.selfhosted.yml restart pulse-backend
```

### SSE desconectando a cada 60 segundos (proxy)

O Pulse envia heartbeats a cada 30 segundos. Se sua conexão SSE ainda cai, seu proxy ou load balancer provavelmente tem timeout de idle connection. Configure:

- **Nginx:** `proxy_read_timeout 3600s;`
- **Traefik:** annotation `traefik.ingress.kubernetes.io/router.middlewares` com timeout ajustado
- **CloudFlare:** o plano Free tem timeout de 100 segundos para SSE — use o header `X-Accel-Buffering: no` (já configurado pelo Pulse)

### Logs detalhados

```bash
# Logs do backend em tempo real
docker logs -f pulse-backend

# Ativar debug completo
docker compose -f docker-compose.selfhosted.yml stop pulse-backend
LOG_LEVEL=debug docker compose -f docker-compose.selfhosted.yml up pulse-backend
```

### Backup dos dados

```bash
# Dump completo
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
