# Pulse — Getting Started

Bem-vindo ao Pulse! Este guia vai te ajudar a configurar feature flags e event tracking na sua aplicação em minutos.

---

## Pré-requisitos

- Uma conta Pulse (Free, Growth ou Pro)
- Acesso ao seu `SERVICE_TOKEN` (Admin → Settings → API)
- Node.js 16+ (para SDK JavaScript) ou Go 1.24+ (para SDK Go)

---

## 1. Configurar Feature Flags via API

### Criar uma Flag

```bash
# Variáveis de ambiente
export PULSE_API=https://pulse-api.institutoitinerante.com.br
export SERVICE_TOKEN=seu-token-aqui
export PRODUCT=libri  # ou nitro, brio, etc

# Criar flag
curl -X POST $PULSE_API/api/v1/flags \
  -H "X-Service-Token: $SERVICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "amazon_buy_button",
    "product": "libri",
    "description": "Novo botão de compra na Amazon",
    "enabled": true,
    "rollout_percentage": 50,
    "tags": ["experimental"]
  }'
```

**Resposta:**
```json
{
  "id": "flag_abc123",
  "name": "amazon_buy_button",
  "product": "libri",
  "enabled": true,
  "rollout_percentage": 50,
  "created_at": "2026-04-09T18:00:00Z",
  "updated_at": "2026-04-09T18:00:00Z"
}
```

### Listar Flags

```bash
curl -X GET "$PULSE_API/api/v1/flags?product=libri" \
  -H "X-Service-Token: $SERVICE_TOKEN"
```

---

## 2. Integrar SDK JavaScript (Frontend)

### Instalação

```bash
# Copie o SDK para seu projeto
cp ../sdk/js/pulse.js src/lib/

# Ou use npm (quando publicado)
npm install @institutoitinerante/pulse-sdk-js
```

### Uso Básico

```typescript
import { PulseSDK } from './lib/pulse'

// Inicializar SDK
const pulse = new PulseSDK({
  apiUrl: 'https://pulse-api.institutoitinerante.com.br',
  product: 'libri',
  userId: 'user-123',  // opcional
  tenantId: 'tenant-456',  // opcional
  cacheTime: 60000  // cache 60s
})

// Obter todas as flags
const flags = await pulse.getFlags()
console.log(flags)

// Verificar uma flag específica
if (pulse.isEnabled('amazon_buy_button')) {
  // Render novo botão de compra
  return <AmazonBuyButton />
}

// Enviar evento customizado (em Growth/Pro)
await pulse.track('product_viewed', {
  productId: 'prod_123',
  price: 49.90,
  category: 'livro'
})
```

### Integração em React

```typescript
import { useEffect, useState } from 'react'
import { PulseSDK } from './lib/pulse'

export function AmazonCheckout() {
  const [flags, setFlags] = useState(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const pulse = new PulseSDK({
      apiUrl: 'https://pulse-api.institutoitinerante.com.br',
      product: 'libri',
      userId: user.id
    })

    pulse.getFlags()
      .then(f => {
        setFlags(f)
        setLoading(false)
      })
      .catch(err => {
        console.error('Pulse error:', err)
        setLoading(false)
      })
  }, [])

  if (loading) return <div>Carregando...</div>

  return (
    <>
      {flags?.amazon_buy_button && <AmazonBuyButton />}
      {flags?.stripe_checkout && <StripeCheckout />}
    </>
  )
}
```

---

## 3. Integrar SDK Go (Backend)

### Instalação

```bash
go get github.com/apsferreira/pulse-service/sdk/go
```

### Uso Básico

```go
package main

import (
  "github.com/apsferreira/pulse-service/sdk/go/pulse"
)

func main() {
  // Inicializar cliente
  client := pulse.NewClient(
    "https://pulse-api.institutoitinerante.com.br",
    "libri",  // product
  )

  // Avaliar flags para usuário
  flags, err := client.EvaluateFlags("user-123", "tenant-456")
  if err != nil {
    log.Fatal(err)
  }

  if flags["amazon_buy_button"] {
    // Ativa rota de compra na Amazon
  }

  // Enviar evento
  err = client.Track("user-123", "product_viewed", map[string]interface{}{
    "productId": "prod_123",
    "price": 49.90,
  })
}
```

---

## 4. Usar Feature Flags no Código

### Padrão: Flag Guard

```typescript
// Antes: sempre ativa
function renderCheckout() {
  return <StripeCheckout />
}

// Depois: protegida por flag
function renderCheckout(pulse: PulseSDK) {
  if (pulse.isEnabled('stripe_checkout_v2')) {
    return <StripeCheckoutV2 />
  }
  return <StripeCheckoutV1 />
}
```

### Padrão: Rollout Percentual

```typescript
// Apenas 30% dos usuários veem o novo design
if (pulse.isEnabled('new_dashboard_design')) {
  return <DashboardV2 />
}
return <DashboardV1 />
```

A flag com `rollout_percentage: 30` garante que o mesmo usuário sempre vê a mesma experiência (hash consistente por userID).

### Padrão: Override por Usuário

```bash
# Forçar usuário específico a ver nova feature
curl -X POST $PULSE_API/api/v1/flags/flag_abc123/overrides \
  -H "X-Service-Token: $SERVICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-123",
    "enabled": true
  }'
```

---

## 5. Monitorar Flags no Dashboard

1. Acesse `https://pulse.institutoitinerante.com.br` (login via auth-service)
2. Selecione seu projeto
3. Veja todas as flags ativas, rollout % e última mudança
4. Edite em tempo real (mudanças propagam em <10s via Redis)

---

## 6. Próximos Passos

- [A/B Testing Guide](./ab-testing.md) — criar e analisar testes
- [API Reference](../API.md) — endpoint completo
- [SDK Documentation](../sdk/README.md) — detalhes de cada SDK
- [Best Practices](./best-practices.md) — padrões recomendados

---

## Troubleshooting

### SDK não conecta na API

```
Erro: "Failed to fetch flags"
Solução: Verifique CORS (apiUrl deve estar acessível do browser)
        ou confira se SERVICE_TOKEN está correto
```

### Flags não mudam em tempo real

```
Erro: "Flag alterada mas não reflete no app"
Solução: Cache JavaScript padrão é 60s. Reduza cacheTime ao inicializar:
        new PulseSDK({ ..., cacheTime: 5000 })
```

### Limite de 10k eventos (Free)

```
Erro: "Event rejected: quota exceeded"
Solução: Faça upgrade para Growth (500k/mês) ou Pro (ilimitado)
        Veja docs/PRICING.md
```

---

## Suporte

- **Docs:** https://github.com/apsferreira/pulse-service/tree/main/docs
- **Issues:** https://github.com/apsferreira/pulse-service/issues
- **Email:** support@institutoitinerante.com.br
- **Discord:** [IIT Community](https://discord.gg/institutoitinerante)
