---
title: "Self-hosted — como funciona"
description: "Como rodar o Pulse na sua própria infraestrutura usando Docker Compose, variáveis de ambiente necessárias e diferenças em relação ao plano managed."
---

# Self-hosted — como funciona

O modo self-hosted permite rodar o Pulse inteiramente na sua infraestrutura. Os dados nunca saem do seu ambiente.

> Para um guia completo com todos os detalhes, consulte [docs/self-hosted.md](../self-hosted.md).

---

## Quando faz sentido usar self-hosted

| Situação | Managed | Self-hosted |
|----------|---------|-------------|
| Startup / equipe pequena | Recomendado | Overhead desnecessário |
| Volume alto de eventos (>5M/mês) | Caro | Custo de infra pode ser menor |
| Requisitos de soberania de dados (LGPD, saúde, financeiro) | Depende | Recomendado |
| Dados nunca podem sair do datacenter próprio | Inadequado | Obrigatório |
| Equipe com experiência em DevOps / K8s | Qualquer um | OK |
| Sem DevOps dedicado | Recomendado | Não recomendado |

---

## Pré-requisitos

- Docker 24+ e Docker Compose v2
- PostgreSQL 15+
- Redis 7+
- 512 MB RAM mínimo para o backend Pulse
- Acesso HTTPS à instância (para o SDK funcionar a partir do browser)

---

## Instalação rápida

### 1. Configurar variáveis de ambiente

Crie um arquivo `.env` no mesmo diretório do compose (nunca commite este arquivo):

```dotenv
# Obrigatório
SERVICE_TOKEN=seu-token-secreto-aqui   # openssl rand -hex 32
POSTGRES_PASSWORD=troque-esta-senha

# URL pública da instância (usada pelo frontend para chamar a API)
PULSE_API_URL=https://pulse-api.seudominio.com

# Modo de autenticação: "local" (self-contained) ou "external" (OAuth)
PULSE_AUTH_MODE=local

# Opcional
ADMIN_EMAIL=ops@seudominio.com
RESEND_API_KEY=re_xxxxxxxx
RESEND_FROM=noreply@seudominio.com
```

### 2. Subir os serviços

```bash
docker compose -f docker-compose.selfhosted.yml up -d
```

### 3. Aplicar migrations

```bash
export PGPASSWORD=troque-esta-senha
for f in $(ls migrations/*.sql | sort); do
  docker exec -i pulse-postgres psql -U pulse -d pulse_db < "$f"
done
```

### 4. Criar primeiro usuário admin (modo local)

```bash
curl -X POST http://localhost:3015/api/v1/auth/setup \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@suaempresa.com", "password": "sua-senha-segura", "name": "Admin"}'
```

### 5. Verificar saúde

```bash
curl http://localhost:3015/health
# {"status":"ok","service":"pulse-service"}
```

---

## Modos de autenticação

### Modo local (recomendado para self-hosted)

Email + senha armazenados no próprio banco. Não requer serviços externos.

```dotenv
PULSE_AUTH_MODE=local
PULSE_JWT_SECRET=seu-secret-jwt   # openssl rand -hex 32 (opcional)
```

### Modo externo (OAuth)

Integração com auth-service OAuth externo.

```dotenv
PULSE_AUTH_MODE=external
PULSE_AUTH_URL=https://auth.suaempresa.com
```

---

## Configurar o SDK para apontar para sua instância

```typescript
// JavaScript / TypeScript
const pulse = new PulseClient({
  apiUrl: 'https://pulse-api.seudominio.com',
  projectKey: 'pk_live_xxxxxxxxxxxxxxxxxx',
})
```

```go
// Go
client, _ := pulse.NewClient("https://pulse-api.seudominio.com")
```

> Todos os SDKs exigem `apiUrl`/`baseURL` explícito. Não há URL padrão embutida.

---

## Deploy em K3s / Kubernetes

```bash
# Manifests no repositório
ls k8s/
# deployment.yaml  service.yaml  ingress.yaml  configmap.yaml  secret.yaml

kubectl apply -f k8s/
```

---

## Diferenças entre managed e self-hosted

| Funcionalidade | Managed | Self-hosted |
|---------------|---------|-------------|
| Updates automáticos | Sim | Manual (você controla) |
| Backups automáticos | Sim | Sua responsabilidade |
| Dados fora da sua infra | Sim | Não (100% local) |
| Customizações | Não | Fork + build próprio |
| Custo de infra | Incluso | Você paga o servidor |

---

## Backup e recuperação

```bash
# Backup do banco
docker exec pulse-postgres pg_dump -U pulse pulse_db \
  | gzip > backup-$(date +%Y%m%d_%H%M%S).sql.gz

# Restaurar
gunzip -c backup-20260411_120000.sql.gz \
  | docker exec -i pulse-postgres psql -U pulse pulse_db
```

---

## Próximos passos

- [self-hosted.md](../self-hosted.md) — guia completo com todas as variáveis e troubleshooting
- [05-qual-plano-escolher.md](./05-qual-plano-escolher.md) — verificar se self-hosted é o plano certo
