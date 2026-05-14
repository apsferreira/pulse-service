# Pulse

Feature flags e event tracking self-hostável. Mude comportamento de qualquer produto sem redeploy.

## Documentação

| Documento | O que cobre |
|-----------|-------------|
| [docs/self-hosted.md](docs/self-hosted.md) | Instalação completa — Docker Compose + Kubernetes + env vars |
| [docs/guides/getting-started.md](docs/guides/getting-started.md) | Primeiros passos: criar projeto, flag e integrar SDK |
| [docs/guides/sdk-reference.md](docs/guides/sdk-reference.md) | Referência completa de todos os SDKs (JS, React, RN, Go, Node) |
| [docs/guides/validation-checklist.md](docs/guides/validation-checklist.md) | Checklist pós-instalação — validar que tudo funciona |
| [docs/guides/ab-testing.md](docs/guides/ab-testing.md) | Criar e analisar testes A/B |
| [docs/benchmarks.md](docs/benchmarks.md) | Benchmarks de performance (P99, RPS, sizing) |

**Documentação orientada a produto (repositório do ecossistema):**

| Documento | O que cobre |
|-----------|-------------|
| [instituto-itinerante-ecosystem/docs/pulse/ab-testing.md](../instituto-itinerante-ecosystem/docs/pulse/ab-testing.md) | A/B testing — perspectiva de produto, exemplos reais IIT (SocialMake, Catraca, Librinho) |
| [instituto-itinerante-ecosystem/docs/pulse/eventos.md](../instituto-itinerante-ecosystem/docs/pulse/eventos.md) | Sistema de eventos — ingest via API e SDK, casos de uso por produto |

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

```bash
# 1. Infraestrutura compartilhada
cd ../shared-infra && docker compose up -d

# 2. Migrations
psql $DATABASE_URL -f migrations/001_initial_schema.sql

# 3. Backend
DATABASE_URL=postgres://postgres:postgres@localhost:5432/pulse_db \
REDIS_URL=redis://localhost:6379 \
SERVICE_TOKEN=dev-token \
go run ./cmd/server

# 4. Frontend (outra aba)
cd frontend && npm install && npm run dev
```

## Comportamento quando o Pulse fica fora do ar

O Pulse foi projetado para ser **fail-open**: se o serviço ficar indisponível, os produtos que o integram continuam funcionando normalmente.

| Camada | O que acontece |
|--------|---------------|
| SDK com cache em memória | Usa estado salvo dos últimos 30 segundos |
| SDK com cache localStorage | Usa último estado salvo (TTL 5 minutos) |
| Sem cache | Retorna o `defaultValue` de cada `getFeatureFlag('key', default)` |
| Rastreamento de eventos | Falha silenciosamente — sem erro visível ao usuário |
| Dashboard | Indisponível, mas os produtos **não percebem** |

```typescript
// Se Pulse estiver fora, retorna `false` (defaultValue) sem erro
const isEnabled = await PulseSDK.getFeatureFlag('nova_feature', false)
```

K3s garante reinício automático do pod em caso de falha (liveness probe a cada 10s).

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

Estimativa de custo self-hosted para 1M MAU: **~R$ 1.250/mês** (vs ~R$ 65.000/mês no LaunchDarkly).

## Licença

[Business Source License 1.1 (BSL 1.1)](LICENSE)

- ✅ Uso interno em qualquer empresa — gratuito
- ✅ Modificar e usar internamente — gratuito
- ❌ Oferecer como SaaS cloud para terceiros — requer licença comercial
- ✅ Converte automaticamente para Apache 2.0 após 4 anos

## Uso em produção e validação

Para validar o Pulse em produção antes de apresentar para terceiros, e para detalhes sobre modelo de contratação e precificação:

→ [`docs/pm/pulse-apresentacao-comercial.md`](../instituto-itinerante-ecosystem/docs/pm/pulse-apresentacao-comercial.md) (no repositório do ecossistema)
