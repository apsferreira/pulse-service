# Pulse

Feature flags e event tracking self-hostável. Mude comportamento de qualquer produto sem redeploy.

## Arquitetura

```mermaid
graph TD
    subgraph Produtos["Produtos / Clientes"]
        A["Site IIT (Astro)"]
        B["SocialMake (React)"]
        C["Librinho (React Native)"]
        D["Nitro (React Native)"]
    end

    subgraph nginx["nginx — proxy site IIT"]
        PF[/pulse-flags]
        PI[/pulse-ingest]
    end

    subgraph K3s["K3s — namespace production"]
        PS["pulse-service (Go/Fiber :3050)"]
        RD[("Redis — cache flags")]
        DB[("PostgreSQL — pulse_db")]
        FE["pulse-service-frontend (React :80)"]
    end

    A -->|GET /pulse-flags| PF
    A -->|POST /pulse-ingest| PI
    PF -->|proxy interno| PS
    PI -->|proxy interno| PS
    B -->|X-Project-Key| PS
    C -->|X-Project-Key| PS
    D -->|X-Project-Key| PS
    PS <-->|cache 30s| RD
    PS <--> DB
    FE -->|X-Service-Token| PS
```

## Fluxo de avaliação de flag

```mermaid
sequenceDiagram
    participant SDK as SDK / Browser
    participant PS as pulse-service
    participant RD as Redis
    participant DB as PostgreSQL

    SDK->>PS: GET /flags?key=minha_flag&user_id=u-123
    PS->>RD: GET cache:flag:projeto:minha_flag
    alt cache hit (TTL 30s)
        RD-->>PS: {rollout: 50, enabled: true}
        PS->>PS: hash(u-123) mod 100 < 50?
        PS-->>SDK: {"enabled": true}
    else cache miss
        RD-->>PS: nil
        PS->>DB: SELECT * FROM feature_flags WHERE key = 'minha_flag'
        DB-->>PS: {rollout: 50, ...}
        PS->>RD: SET cache:flag:... EX 30
        PS->>PS: hash(u-123) mod 100 < 50?
        PS-->>SDK: {"enabled": true}
    end
    PS->>DB: INSERT INTO flag_evaluations (async)
```

## Rollout determinístico

```mermaid
graph LR
    subgraph "hash(user_id) mod 100 — rollout 50%"
        U1["user-01 → 3"] --> K1[controle]
        U2["user-12 → 47"] --> T1[tratamento]
        U3["user-33 → 81"] --> K2[controle]
        U4["user-77 → 12"] --> T2[tratamento]
    end
    style T1 fill:#6ee7b7,color:#000
    style T2 fill:#6ee7b7,color:#000
    style K1 fill:#e2e8f0,color:#000
    style K2 fill:#e2e8f0,color:#000
```

O mesmo `user_id` **sempre** cai no mesmo grupo — sem guardar estado, sem sessão, sem banco. Só hash.

## Ciclo de vida de um projeto

```mermaid
stateDiagram-v2
    [*] --> AguardandoSDK: Projeto criado no Dashboard
    AguardandoSDK --> Ativo: "Primeira requisição autenticada (ActivateByAPIKey)"
    Ativo --> Ativo: Flags avaliadas / eventos recebidos
```

## Documentação

| Documento | O que cobre |
|-----------|-------------|
| [docs/self-hosted.md](docs/self-hosted.md) | Instalação completa — Docker Compose + Kubernetes + env vars |
| [docs/guides/getting-started.md](docs/guides/getting-started.md) | Primeiros passos: criar projeto, flag e integrar SDK |
| [docs/guides/sdk-reference.md](docs/guides/sdk-reference.md) | Referência completa de todos os SDKs (JS, React, RN, Go, Node) |
| [docs/guides/validation-checklist.md](docs/guides/validation-checklist.md) | Checklist pós-instalação — validar que tudo funciona |
| [docs/guides/ab-testing.md](docs/guides/ab-testing.md) | Criar e analisar testes A/B (técnico) |
| [docs/guides/ab-testing-produto.md](docs/guides/ab-testing-produto.md) | A/B testing — perspectiva de produto, exemplos reais IIT |
| [docs/guides/eventos-produto.md](docs/guides/eventos-produto.md) | Sistema de eventos — ingest via API e SDK, casos de uso por produto |
| [docs/benchmarks.md](docs/benchmarks.md) | Benchmarks de performance (P99, RPS, sizing) |

## Início rápido (self-hosted)

```bash
git clone <este-repo>
cp .env.example .env          # preencher DATABASE_URL, REDIS_URL, SERVICE_TOKEN
docker compose -f docker-compose.selfhosted.yml up -d
```

Guia completo com migrations, auth e troubleshooting: [`docs/self-hosted.md`](docs/self-hosted.md)

## Conceito central

Cada produto usa um **projeto** isolado com sua própria `api_key`. As feature flags vivem dentro do projeto — não há flags compartilhadas entre produtos.

```
Pulse
├── projeto: librinho    (api_key: pulse_libri_...)
│   ├── flag: isbn_camera_scan
│   └── flag: reading_insights_v2
├── projeto: nitro       (api_key: pulse_nitro_...)
│   └── flag: ai_study_plan
└── projeto: socialmake  (api_key: pulse_sm_...)
    └── flag: new_ai_copy_engine
```

## Endpoints principais

| Method | Path | Auth | Descrição |
|--------|------|------|-----------|
| `GET` | `/health` | — | Health check |
| `GET` | `/flags` | `X-Project-Key` | Avaliar flags para um usuário (SDK) |
| `GET` | `/api/v1/flags` | `X-Service-Token` | Listar flags (admin) |
| `POST` | `/api/v1/flags` | `X-Service-Token` | Criar flag |
| `PUT` | `/api/v1/flags/:key` | `X-Service-Token` | Atualizar flag |
| `DELETE` | `/api/v1/flags/:key` | `X-Service-Token` | Deletar flag |
| `GET` | `/api/v1/projects` | `X-Service-Token` | Listar projetos |
| `POST` | `/api/v1/projects` | `X-Service-Token` | Criar projeto |
| `POST` | `/api/v1/events` | `X-Project-Key` | Rastrear evento |
| `GET` | `/api/v1/events/summary` | `X-Service-Token` | Resumo de eventos por projeto |

Endpoint de avaliação (SDK):
```
GET /flags?user_id=user-123&project_id=<uuid>
X-Project-Key: pulse_xxxx
```

## SDK — exemplo mínimo

```typescript
import { PulseSDK } from '@pulse/js'

PulseSDK.init({
  projectKey: process.env.VITE_PULSE_KEY,
  apiUrl: process.env.VITE_PULSE_URL,
})

// Verificar flag
const isEnabled = await PulseSDK.getFeatureFlag('nova_feature', false)

// Rastrear evento
PulseSDK.track('checkout_started', { plan: 'growth' })
```

Referência completa com Go, Node, React Native e tratamento de erros: [`docs/guides/sdk-reference.md`](docs/guides/sdk-reference.md)

## Stack

- **Backend:** Go 1.24 + Fiber v2
- **Banco:** PostgreSQL 15 (pgxpool)
- **Cache:** Redis 7 (go-redis/v9)
- **Frontend:** React 19 + TypeScript + Vite + TailwindCSS
- **Infra:** Docker Compose (dev) / K3s (prod) / ArgoCD (GitOps)

## Desenvolvimento local

Você precisa de PostgreSQL 15+ e Redis 7+ rodando localmente. A forma mais simples é via Docker:

```bash
# PostgreSQL + Redis via Docker (se não tiver rodando)
docker run -d --name pulse-postgres -e POSTGRES_PASSWORD=postgres -p 5432:5432 postgres:15-alpine
docker run -d --name pulse-redis -p 6379:6379 redis:7-alpine
```

Ou use qualquer outra forma que preferir — Homebrew, Postgres.app, etc.

```bash
# 1. Variáveis de ambiente
cp .env.example .env
# Edite .env com suas credenciais locais

# 2. Criar banco e rodar migrations
createdb pulse_db   # ou: psql -U postgres -c "CREATE DATABASE pulse_db"
psql $DATABASE_URL -f migrations/001_initial_schema.sql

# 3. Backend
go run ./cmd/server

# 4. Frontend (outra aba)
cd frontend && npm install && npm run dev
```

**`.env` mínimo para desenvolvimento:**

```env
DATABASE_URL=postgres://postgres:postgres@localhost:5432/pulse_db
REDIS_URL=redis://localhost:6379
SERVICE_TOKEN=dev-token
PORT=3050
```

## Comportamento quando o Pulse fica fora do ar

O Pulse foi projetado para ser **fail-open**: se o serviço ficar indisponível, os produtos que o integram continuam funcionando normalmente.

| Camada | O que acontece |
|--------|---------------|
| SDK JS/React — cache em memória | Usa o último estado retornado pela API (durante a vida da instância) |
| SDK JS/React — fallback localStorage | Usa último estado salvo (TTL 5 minutos) |
| Sem chamada prévia | Retorna o `defaultValue` configurado por flag |
| Rastreamento de eventos | Falha silenciosamente — sem erro visível ao usuário |
| Dashboard | Indisponível, mas os produtos **não percebem** |

```typescript
const client = new PulseClient({
  baseURL: 'https://seu-pulse.exemplo.com',
  failOpen: true,           // padrão: true
  defaults: {
    nova_feature: false,    // retornado se Pulse estiver fora e não houver cache
  },
})

const flags = await client.getFlags()
// Se falhar: usa localStorage (5min) → defaults → não lança erro
```

## Performance

P99 < 100ms a 5.000+ RPS (2 vCPU / 2 GB):

| Endpoint | Req/s | P99 |
|---|---|---|
| `GET /flags` (Redis cache) | 8.200 | 18ms |
| `POST /api/v1/events` | 3.800 | 82ms |

Benchmark completo e sizing recomendado: [`docs/benchmarks.md`](docs/benchmarks.md)

## Escalabilidade

Para 1 milhão de usuários ativos:

| Componente | Config mínima | Config para 1M MAU |
|------------|--------------|-------------------|
| Réplicas backend | 1 | 3-5 (HPA CPU > 70%) |
| Redis | Standalone | Cluster (3 nós) |
| PostgreSQL | Shared | Dedicado com read replicas |
| CDN `/flags` | — | Cloudflare Cache (TTL 30s) |

Para dimensionamento e licenciamento entre em contato: contato@institutoitinerante.com.br

## Licença e uso comercial

O Pulse é uma plataforma comercial desenvolvida pelo Instituto Itinerante de Tecnologia.

- O código-fonte está disponível para auditoria e self-hosting avaliativo
- Uso em produção (próprio ou para terceiros) requer licença comercial
- Para planos e preços: [institutoitinerante.com.br/pulse](https://institutoitinerante.com.br/pulse)
- Contato comercial: contato@institutoitinerante.com.br
