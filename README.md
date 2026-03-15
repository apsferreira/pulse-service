# pulse-service

Feature flags and event tracking service for Instituto Itinerante products (libri, nitro, brio).

## Fase 1 — Feature Flags MVP

- CRUD de feature flags com suporte a produtos e flags globais
- Avaliação por usuário com cache Redis (TTL 60s)
- Rollout percentual com hash consistente por userID (crc32)
- Overrides por usuário ou tenant
- SDK JavaScript mínimo (`sdk/pulse.js`)
- Proteção de endpoints admin via `X-Service-Token`

## Stack

- Go 1.24 + Fiber v2
- PostgreSQL 15 (pgxpool)
- Redis 7 (go-redis/v9)
- K3s + Traefik + ArgoCD

## Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | /health | — | Health check |
| GET | /api/v1/flags | X-Service-Token | List all flags |
| POST | /api/v1/flags | X-Service-Token | Create flag |
| PUT | /api/v1/flags/:id | X-Service-Token | Update flag |
| DELETE | /api/v1/flags/:id | X-Service-Token | Delete flag |
| GET | /api/v1/flags/evaluate | — | Evaluate flags for user |

### Evaluate query params

- `product` — produto (libri, nitro, brio) ou vazio para flags globais
- `user_id` — ID do usuário (opcional)
- `tenant_id` — ID do tenant (opcional)

## Configuração

| Env var | Default | Description |
|---------|---------|-------------|
| `DATABASE_URL` | — | PostgreSQL connection string (obrigatório) |
| `REDIS_URL` | `redis://localhost:6379` | Redis connection URL |
| `SERVICE_TOKEN` | — | Token para endpoints admin (obrigatório) |
| `PORT` | `3015` | Porta HTTP |
| `LOG_LEVEL` | `info` | Nível de log |

## Desenvolvimento local

```bash
# Suba shared-infra primeiro
cd ../shared-infra && docker compose up -d

# Aplique a migration
psql $DATABASE_URL -f migrations/001_create_feature_flags.sql

# Execute o servidor
DATABASE_URL=... SERVICE_TOKEN=dev-token go run ./cmd/server
```

## SDK JavaScript

```js
import { PulseSDK } from './sdk/pulse.js'

const pulse = new PulseSDK({
  apiUrl: 'https://pulse-api.institutoitinerante.com.br',
  product: 'libri',
  userId: 'user-123',
})

const flags = await pulse.getFlags()
if (pulse.isEnabled('amazon_buy_button')) {
  // mostra botão
}
```

## Produção

- **URL:** https://pulse-api.institutoitinerante.com.br
- **Namespace K8s:** production
- **Secret:** `pulse-service-secrets` (database-url, redis-url, service-token)
- **Deploy:** ArgoCD auto-sync em `deploy/k8s/`
