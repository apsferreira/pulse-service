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

```mermaid
graph TB
    subgraph Clientes
        SDK[SDK JS/Flutter/Go/Node]
        Browser[Dashboard Browser]
    end

    subgraph pulse-backend Go/Fiber porta 3015
        API[REST API\nadmin SERVICE_TOKEN\npublico X-Project-Key]
        SSE[SSE /flags/stream\nflag updates tempo real]
        WS[WebSocket /ws\neventos tempo real]
        CDN[CDN /sdk/pulse.js]
    end

    subgraph Storage
        PG[(PostgreSQL\ndados persistentes)]
        Redis[(Redis DB6\ncache de flags\nrate-limit)]
    end

    Browser -->|HTTPS| API
    Browser -->|EventSource| SSE
    Browser -->|ws| WS
    SDK -->|X-Project-Key| API
    SDK -->|script tag| CDN
    API --> PG
    API --> Redis
    SSE --> Redis
```

O backend expõe:
- **API REST** para gerenciamento (autenticada via `SERVICE_TOKEN`) e para SDKs (endpoints públicos com `X-Project-Key`)
- **SSE** (`/api/v1/projects/:project_id/flags/stream`) para flag updates em tempo real
- **WebSocket** (`/ws`) para eventos em tempo real no dashboard
- **SDK CDN** servido estaticamente em `/sdk/pulse.js`

### SDKs disponíveis

| SDK | Ambiente | Como usar |
|---|---|---|
| JavaScript / CDN | Qualquer site via `<script>` | `<script src="https://seu-pulse.seudominio.com/sdk/pulse.js">` |
| `@pulse/js` | React + TypeScript + Vite | `npm install @pulse/js` |
| Flutter/Dart | Apps Flutter (Android/iOS) | `sdk/flutter/` no monorepo |
| Android (Kotlin) | Android nativo | `sdk/android/` no monorepo |
| iOS (Swift) | iOS nativo | `sdk/ios/` no monorepo |
| Go | Backend Go | `sdk/go/` no monorepo |
| Node.js | Backend Node | `sdk/node/` no monorepo |

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

# Licença self-hosted (obrigatório para instâncias não-IIT)
# Obtida em: https://pulse.institutoitinerante.com.br/self-hosted
PULSE_LICENSE_KEY=pulse_lic_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

> Gere tokens seguros com: `openssl rand -hex 32`

> **Trial de 14 dias:** sem uma licença válida, o Pulse funciona normalmente por 14 dias. Após esse período, o dashboard entra em modo somente leitura — nenhum dado é perdido, mas criar/modificar flags e projetos fica bloqueado até ativar uma licença.

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

```mermaid
sequenceDiagram
    participant U as Usuário
    participant F as pulse-frontend
    participant B as pulse-backend
    participant DB as PostgreSQL

    Note over B,DB: Primeiro acesso — tabela pulse_users vazia
    U->>F: Acessa o dashboard
    F->>B: POST /api/v1/auth/setup
    B->>DB: INSERT pulse_users (admin)
    B-->>F: JWT token
    F-->>U: Redireciona para dashboard

    Note over U,DB: Acessos subsequentes
    U->>F: Login email + senha
    F->>B: POST /api/v1/auth/login
    B->>DB: Valida credenciais
    B-->>F: JWT (24h)
    F-->>U: Dashboard carregado

    U->>F: Requisição autenticada
    F->>B: GET /api/v1/... Bearer JWT
    B->>B: Valida JWT com PULSE_JWT_SECRET
    B-->>F: Dados
```

### Opção B — Auth Externo (OAuth via auth-service)

Integração com um auth-service OAuth externo. Adequado para organizações que já têm SSO ou um provedor de identidade próprio.

```dotenv
PULSE_AUTH_MODE=external
PULSE_AUTH_URL=https://auth.suaempresa.com
```

```mermaid
sequenceDiagram
    participant U as Usuário
    participant F as pulse-frontend
    participant A as auth-service externo
    participant B as pulse-backend

    U->>F: Acessa o dashboard
    F->>A: Redireciona para PULSE_AUTH_URL
    U->>A: Login SSO/OAuth
    A-->>F: Redirect com token OAuth
    F->>B: Requisições com token OAuth
    B->>A: Valida token (opcional)
    B-->>F: Dados autorizados
```

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
| `PULSE_LICENSE_KEY` | Não* | — | Chave de licença self-hosted. *Obrigatória após os 14 dias de trial para instâncias não-IIT. Obtida em `https://pulse.institutoitinerante.com.br/self-hosted`. |

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

### Flutter / Dart

```yaml
# pubspec.yaml
dependencies:
  pulse_flutter:
    path: ../pulse-service/sdk/flutter  # dev local
```

```dart
import 'package:flutter/foundation.dart';
import 'package:pulse_flutter/pulse_flutter.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await Pulse.init(
    projectKey: 'pk_live_xxxxxxxxxxxxxxxxxx',
    debug: kDebugMode,
  );
  runApp(const MyApp());
}
```

Consulte [guides/sdk-reference.md](./guides/sdk-reference.md#flutter--dart) para a referência completa.

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

```mermaid
sequenceDiagram
    participant SDK as SDK/Cliente
    participant B as pulse-backend
    participant Redis as Redis cache
    participant DB as PostgreSQL

    SDK->>B: GET /flags?user_id=X&project_id=Y
    B->>Redis: GET pulse:flags:project_id:map
    alt Cache hit TTL 5min
        Redis-->>B: flags em cache
    else Cache miss
        B->>DB: SELECT * FROM feature_flags WHERE project_id=Y
        DB-->>B: flags do projeto
        B->>Redis: SET pulse:flags:project_id:map TTL 300s
    end
    B->>B: Avalia cada flag para user_id X
    Note over B: rollout hash, targeting, prerequisites
    B-->>SDK: {flags: {flag-a: true, flag-b: false}}
    B->>DB: INSERT flag_evaluations (audit log async)

    Note over SDK,Redis: Flag atualizada no dashboard
    SDK->>B: GET /flags/stream (SSE)
    B-->>SDK: event: flag_updated data: {key, enabled}
    SDK->>SDK: Atualiza estado local sem novo request
```

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

```mermaid
flowchart TD
    A[Verificar CHANGELOG\nnova versão disponível] --> B[Fazer backup do banco]
    B --> C[docker pull nova imagem]
    C --> D{Há novas migrations?}
    D -->|Sim| E[Aplicar migrations\nem ordem numérica]
    D -->|Não| F[docker compose up -d]
    E --> F
    F --> G[Verificar GET /health]
    G --> H{Retornou status ok?}
    H -->|Sim| I[Atualização concluída]
    H -->|Não| J[docker logs pulse-backend\nDiagnóstico]
    J --> K[Rollback:\ndocker compose up -d com tag anterior]
```

```bash
# 1. Backup antes de tudo
docker exec pulse-postgres pg_dump -U pulse pulse_db > backup-$(date +%Y%m%d).sql

# 2. Baixar nova imagem
docker pull ghcr.io/seu-org/pulse-service:latest
docker pull ghcr.io/seu-org/pulse-frontend:latest

# 3. Aplicar novas migrations (verifique o CHANGELOG)
for f in migrations/NEW_*.sql; do
  docker exec -i pulse-postgres psql -U pulse -d pulse_db < "$f"
done

# 4. Reiniciar
docker compose -f docker-compose.selfhosted.yml up -d

# 5. Validar
curl http://localhost:3015/health
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
