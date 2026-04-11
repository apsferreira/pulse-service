---
title: "Self-hosted — como funciona"
description: "Como rodar o Pulse na sua própria infraestrutura usando Docker Compose, variáveis de ambiente necessárias e diferenças em relação ao plano managed."
---

# Self-hosted — como funciona

O modo self-hosted permite rodar o Pulse inteiramente na sua infraestrutura. Os dados nunca saem do seu ambiente. Disponível no plano **Pro** mediante consulta.

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
- PostgreSQL 15+ (pode ser o mesmo que você já usa ou um novo banco)
- Redis 7+
- 2 vCPU / 2 GB RAM mínimo para o backend Pulse
- Acesso HTTPS à instância (para o SDK funcionar a partir do browser)

---

## Instalação com Docker Compose

### 1. Obter o arquivo de composição

O arquivo `docker-compose.selfhosted.yml` é fornecido ao ativar o plano Pro. Solicite via support@institutoitinerante.com.br com seu `project_id`.

Estrutura do arquivo:

```yaml
# docker-compose.selfhosted.yml
version: '3.9'

services:
  pulse-api:
    image: ghcr.io/apsferreira/pulse-service:latest
    restart: unless-stopped
    ports:
      - "8090:8090"
    environment:
      - DATABASE_URL=${DATABASE_URL}
      - REDIS_URL=${REDIS_URL}
      - SERVICE_TOKEN=${SERVICE_TOKEN}
      - JWT_SECRET=${JWT_SECRET}
      - PORT=8090
      - ENV=production
      - LOG_LEVEL=info
      - CORS_ORIGINS=${CORS_ORIGINS}
    depends_on:
      - pulse-migrate
    networks:
      - pulse-net

  pulse-migrate:
    image: ghcr.io/apsferreira/pulse-service:latest
    command: ["migrate", "up"]
    environment:
      - DATABASE_URL=${DATABASE_URL}
    networks:
      - pulse-net

  pulse-dashboard:
    image: ghcr.io/apsferreira/pulse-dashboard:latest
    restart: unless-stopped
    ports:
      - "8091:80"
    environment:
      - VITE_PULSE_API_URL=${PULSE_API_URL}
    networks:
      - pulse-net

networks:
  pulse-net:
    driver: bridge
```

### 2. Configurar variáveis de ambiente

Crie um arquivo `.env` no mesmo diretório (nunca commite este arquivo):

```bash
# .env — self-hosted Pulse

# Banco de dados PostgreSQL
DATABASE_URL=postgres://pulse_user:senha_segura@<REDACTED-INTERNAL-IP>:5432/pulse_db?sslmode=disable

# Redis
REDIS_URL=redis://:senha_redis@<REDACTED-INTERNAL-IP>:6379/5

# Token de acesso ao serviço (gerado na configuração inicial)
SERVICE_TOKEN=pulse_svc_seu_token_secreto_aqui

# Secret para JWT interno (mínimo 32 caracteres aleatórios)
JWT_SECRET=sua_chave_jwt_de_pelo_menos_32_caracteres

# URL pública da API (usada pelo dashboard)
PULSE_API_URL=https://pulse-api.seudominio.com.br

# Origens permitidas no CORS (lista separada por vírgula)
CORS_ORIGINS=https://pulse.seudominio.com.br,https://app.seudominio.com.br
```

Gere valores seguros com:

```bash
# SERVICE_TOKEN
openssl rand -hex 32

# JWT_SECRET
openssl rand -base64 48
```

### 3. Criar o banco e rodar migrations

```bash
# Criar banco (se necessário)
psql -h <REDACTED-INTERNAL-IP> -U postgres -c "CREATE DATABASE pulse_db;"
psql -h <REDACTED-INTERNAL-IP> -U postgres -c "CREATE USER pulse_user WITH PASSWORD 'senha_segura';"
psql -h <REDACTED-INTERNAL-IP> -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE pulse_db TO pulse_user;"

# Rodar migrations via Docker Compose
docker compose -f docker-compose.selfhosted.yml run --rm pulse-migrate
```

### 4. Subir os serviços

```bash
docker compose -f docker-compose.selfhosted.yml up -d

# Verificar logs
docker compose -f docker-compose.selfhosted.yml logs -f pulse-api
```

### 5. Verificar saúde

```bash
curl https://pulse-api.seudominio.com.br/health
```

```json
{
  "status": "ok",
  "version": "1.4.2",
  "db": "ok",
  "redis": "ok",
  "uptime_seconds": 42
}
```

---

## Configurar o SDK para apontar para sua instância

```typescript
const pulse = new PulseSDK({
  apiUrl: 'https://pulse-api.seudominio.com.br',  // sua instância, não a managed
  product: 'minha-app',
  serviceToken: process.env.PULSE_SERVICE_TOKEN!,
})
```

---

## Atualizações

Para atualizar para uma nova versão:

```bash
# Puxar nova imagem
docker compose -f docker-compose.selfhosted.yml pull

# Rodar migrations da nova versão (sem downtime se backward-compatible)
docker compose -f docker-compose.selfhosted.yml run --rm pulse-migrate

# Reiniciar o serviço
docker compose -f docker-compose.selfhosted.yml up -d pulse-api
```

Consulte o [CHANGELOG](https://github.com/apsferreira/pulse-service/blob/main/CHANGELOG.md) antes de cada update para verificar se há breaking changes nas migrations.

---

## Deploy em K3s / Kubernetes

Se você já usa K3s (como o ambiente do IIT), prefira o manifesto Kubernetes em vez do Docker Compose:

```bash
# Manifests disponíveis no repositório
ls /Users/antonio.ferreira/projects/pulse-service/k8s/
# deployment.yaml  service.yaml  ingress.yaml  configmap.yaml  secret.yaml
```

Aplique com:

```bash
KUBECONFIG=~/.kube/k3s-homelab kubectl apply -f k8s/
```

---

## Diferenças entre managed e self-hosted

| Funcionalidade | Managed (Pro) | Self-hosted |
|---------------|---------------|-------------|
| Updates automáticos | Sim | Manual (você controla) |
| Backups automáticos | Sim | Sua responsabilidade |
| SLA 99,95% | Contratual | Depende da sua infra |
| Suporte prioritário | Incluído | Incluído (email/GitHub) |
| Compliance de dados | Dados no IIT | Dados 100% na sua infra |
| Customizações | Não | Fork + build próprio |
| Custo de infra | Incluso | Você paga o servidor |

---

## Backup e recuperação

```bash
# Backup do banco (rodar via cron)
pg_dump -h <REDACTED-INTERNAL-IP> -U pulse_user pulse_db \
  | gzip > /backups/pulse_$(date +%Y%m%d_%H%M%S).sql.gz

# Restaurar
gunzip -c /backups/pulse_20260411_120000.sql.gz \
  | psql -h <REDACTED-INTERNAL-IP> -U pulse_user pulse_db
```

---

## Próximos passos

- [05-qual-plano-escolher.md](./05-qual-plano-escolher.md) — verificar se self-hosted é o plano certo para você
- [10-integrar-pulse-ci-cd.md](./10-integrar-pulse-ci-cd.md) — automatizar deploy da instância self-hosted em CI/CD
