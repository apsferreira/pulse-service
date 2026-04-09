# Pulse — A/B Testing Guide

A/B testing no Pulse permite que você compare duas ou mais variantes de uma feature e meça qual gera melhor resultado. Este guia cobre desde criação até análise de testes.

---

## O que é A/B Testing no Pulse?

Um A/B test é uma flag controlada com múltiplas variantes que você mede durante um período:

- **Variante A (Control):** experiência atual (baseline)
- **Variante B (Treatment):** nova experiência que você quer testar
- **Métrica:** qual KPI você mede (ex: conversão, tempo na página, retention)
- **Duração:** geralmente 1-4 semanas

**Exemplo:** "Botão de checkout azul vs vermelho — qual gera mais conversões?"

---

## 1. Criar um A/B Test via API

### Passo 1: Preparar Variantes

Primeiro, crie a flag com dois designs ou fluxos:

```typescript
// Seu código (Frontend)
const checkoutButtonColor = pulse.getVariant('checkout_button_color')

if (checkoutButtonColor === 'blue') {
  return <CheckoutButton color="blue" />
}
if (checkoutButtonColor === 'red') {
  return <CheckoutButton color="red" />
}
return <CheckoutButton color="blue" />  // fallback
```

### Passo 2: Criar a Flag com Variantes

```bash
export PULSE_API=https://pulse-api.institutoitinerante.com.br
export SERVICE_TOKEN=seu-token

curl -X POST $PULSE_API/api/v1/flags \
  -H "X-Service-Token: $SERVICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "checkout_button_color",
    "product": "libri",
    "description": "A/B test: azul vs vermelho",
    "enabled": true,
    "type": "multivariate",
    "variants": [
      {
        "id": "blue",
        "weight": 50,
        "config": {
          "color": "blue",
          "text": "Pagar com Stripe"
        }
      },
      {
        "id": "red",
        "weight": 50,
        "config": {
          "color": "red",
          "text": "Finalizar Compra"
        }
      }
    ],
    "start_date": "2026-04-09T00:00:00Z",
    "end_date": "2026-04-23T23:59:59Z",
    "tracking_metrics": [
      "checkout_button_clicked",
      "checkout_completed",
      "checkout_abandoned"
    ]
  }'
```

**Resposta:**
```json
{
  "id": "flag_multivar_abc",
  "name": "checkout_button_color",
  "type": "multivariate",
  "variants": [
    {"id": "blue", "weight": 50},
    {"id": "red", "weight": 50}
  ],
  "start_date": "2026-04-09T00:00:00Z",
  "end_date": "2026-04-23T23:59:59Z"
}
```

---

## 2. Integrar no Frontend via SDK JS

### Obter Variante Designada

```typescript
import { PulseSDK } from './lib/pulse'
import { useEffect, useState } from 'react'

export function CheckoutButton({ userId }) {
  const [variant, setVariant] = useState<'blue' | 'red'>('blue')

  useEffect(() => {
    const pulse = new PulseSDK({
      apiUrl: 'https://pulse-api.institutoitinerante.com.br',
      product: 'libri',
      userId
    })

    // Obter variante (determinística por userID)
    const v = pulse.getVariant('checkout_button_color')
    setVariant(v || 'blue')

    // Registrar que usuário viu uma variante
    pulse.track('ab_test_exposed', {
      flagId: 'checkout_button_color',
      variant: v,
      timestamp: new Date().toISOString()
    })
  }, [userId])

  return (
    <button
      className={`btn btn-${variant}`}
      onClick={() => {
        // Usuário clicou — enviar evento
        pulse.track('checkout_button_clicked', {
          variant,
          timestamp: new Date().toISOString()
        })
      }}
    >
      {variant === 'blue' ? 'Pagar com Stripe' : 'Finalizar Compra'}
    </button>
  )
}
```

### Registrar Conversão

```typescript
export function CheckoutSuccess({ userId, orderId, variant }) {
  useEffect(() => {
    const pulse = new PulseSDK({
      apiUrl: 'https://pulse-api.institutoitinerante.com.br',
      product: 'libri',
      userId
    })

    // Checkout completado
    pulse.track('checkout_completed', {
      flagId: 'checkout_button_color',
      variant,
      orderId,
      revenue: 99.90,
      timestamp: new Date().toISOString()
    })
  }, [])

  return <div>Pedido {orderId} confirmado!</div>
}
```

---

## 3. Integrar Backend (Go) para Rastrear

```go
import (
  "github.com/apsferreira/pulse-service/sdk/go/pulse"
)

func HandleCheckout(w http.ResponseWriter, r *http.Request) {
  userId := r.Header.Get("X-User-ID")
  orderId := r.FormValue("order_id")
  revenue := r.FormValue("revenue")

  // Obter variante do A/B test
  variant, err := pulse.GetVariant(userId, "checkout_button_color")
  if err != nil {
    variant = "blue"  // fallback
  }

  // Processar checkout...
  order, err := processOrder(orderId, revenue)
  if err != nil {
    // Rastrear abandono
    pulse.Track(userId, "checkout_abandoned", map[string]interface{}{
      "variant": variant,
      "reason": "payment_failed",
      "error": err.Error(),
    })
    return
  }

  // Rastrear sucesso
  pulse.Track(userId, "checkout_completed", map[string]interface{}{
    "flagId": "checkout_button_color",
    "variant": variant,
    "orderId": order.ID,
    "revenue": order.Total,
  })

  w.WriteHeader(200)
  json.NewEncoder(w).Encode(order)
}
```

---

## 4. Monitorar Testes no Dashboard

### Acessar Relatório de A/B Test

1. Abra `https://pulse.institutoitinerante.com.br/ab-testing`
2. Selecione o teste: `checkout_button_color`
3. Veja em tempo real:
   - **Variante A:** 1200 usuários | 240 conversões (20%)
   - **Variante B:** 1195 usuários | 286 conversões (23.9%)
   - **Diferença:** +3.9% (p-value: 0.042 ✅ Significante)

### Gráficos Disponíveis

- **Distribuição:** % de usuários em cada variante
- **Conversão:** taxa por variante ao longo do tempo
- **Eventos:** click, abandono, tempo médio
- **Revenue:** renda por variante (se rastreado)

---

## 5. Analisar Resultados

### Estatística Básica

O Pulse calcula automaticamente:
- **Taxa de conversão** por variante
- **Intervalo de confiança** (95%)
- **P-value** para significância estatística
- **Lift:** quanto melhor B vs A em %

**Exemplo:**
```
Variante A (Control):
  - Amostra: 2000 usuários
  - Conversões: 400 (20%)
  - Intervalo: 18.8% - 21.2%

Variante B (Red button):
  - Amostra: 2000 usuários
  - Conversões: 460 (23%)
  - Intervalo: 21.7% - 24.3%

Resultado: Variante B é 15% melhor (p=0.013, significante ✅)
```

### Quando Encerrar o Teste?

1. **Tempo mínimo:** 1 semana (para capturar variações semanais)
2. **Amostra mínima:** 100-200 conversões por variante
3. **Significância:** p-value < 0.05
4. **Duração máxima:** 4 semanas (aproveitar tendências)

---

## 6. Promover Variante Vencedora

### Via API

```bash
# Após confirmar que Variante B é melhor
curl -X POST $PULSE_API/api/v1/flags/flag_multivar_abc/promote \
  -H "X-Service-Token: $SERVICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "winning_variant": "red",
    "rollout_percentage": 100
  }'
```

### No Dashboard

1. Clique em "Finalizar Teste"
2. Selecione a variante vencedora
3. Escolha rollout (100% = todos, ou gradual 10% → 25% → 50% → 100%)
4. Confirme

### Rollout Gradual (Recomendado)

```bash
# Dia 1: 10% em "red"
curl -X PUT $PULSE_API/api/v1/flags/flag_multivar_abc \
  -d '{"variants": [
    {"id": "blue", "weight": 90},
    {"id": "red", "weight": 10}
  ]}'

# Dia 2: 25%
# {"variants": [{"id": "blue", "weight": 75}, {"id": "red", "weight": 25}]}

# Dia 3: 50%
# {"variants": [{"id": "blue", "weight": 50}, {"id": "red", "weight": 50}]}

# Dia 4: 100%
# {"variants": [{"id": "red", "weight": 100}]}
```

Isso reduz risco: se houver bug, você já volta rápido.

---

## 7. Arquivar Teste Finalizado

```bash
curl -X DELETE $PULSE_API/api/v1/flags/flag_multivar_abc \
  -H "X-Service-Token: $SERVICE_TOKEN"
```

Dados do teste são mantidos no histórico (90 dias em Pro, 30 em Growth, 7 em Free).

---

## Melhores Práticas

### Teste Design

✅ **Bom:** Teste UNM variação (cor, texto, tamanho)
❌ **Ruim:** Teste múltiplas mudanças ao mesmo tempo

✅ **Bom:** 50/50 split inicial
❌ **Ruim:** 10/90 split (alta variância na minoria)

### Métrica

✅ **Bom:** "Checkout completado" (conversa clara)
❌ **Ruim:** "Usuário ativo" (muito vago)

✅ **Bom:** "Revenue por sessão"
❌ **Ruim:** "Cliques totais" (métrica de vaidade)

### Duração

✅ **Bom:** Pelo menos 1 semana (captura ciclos)
❌ **Ruim:** 1-2 dias (muito ruído)

✅ **Bom:** 4 semanas máximo
❌ **Ruim:** Deixar rodar 3 meses (oportunidade perdida)

---

## Exemplos de A/B Tests Reais

### 1. Experimento: Botão CTA

```
Nome: "signup_button_text"
Variantes:
  - A: "Comece Agora" (control)
  - B: "Teste Grátis"
Métrica: signup_completed
Resultado: B ganha +12% (p=0.018)
Ação: Promover B a 100%
```

### 2. Experimento: Email Retenção

```
Nome: "retention_email_frequency"
Variantes:
  - A: 1 email/semana (control)
  - B: 2 emails/semana
Métrica: subscription_renewed
Resultado: Sem diferença (p=0.42)
Ação: Manter A (economiza custo)
```

### 3. Experimento: Preço

```
Nome: "checkout_price_plan"
Variantes:
  - A: R$ 99/mês (control)
  - B: R$ 149/mês
Métrica: revenue_per_user
Resultado: B ganha +8% (p=0.006) mas conversão cai 5%
Ação: Promover B (revenue > volume)
```

---

## Troubleshooting

### Teste não coleta eventos

```
Problema: Variantes criadas mas sem eventos
Solução: Verifique que pulse.track() é chamado no frontend/backend
        Logs em https://pulse.institutoitinerante.com.br/events
```

### Amostra muito pequena

```
Problema: "Não há dados suficientes para significância"
Solução: Deixe rodar mais dias
        Ou aumente traffic para a variante (reduzir weight de control)
```

### Variantes desbalanceadas

```
Problema: 70/30 split mas deveria ser 50/50
Solução: Verifique weights na configuração da flag
        Pulse usa hash(userID) — determinístico, não aleatório
```

---

## Referências

- [Getting Started](./getting-started.md) — setup básico
- [API Reference](../API.md) — endpoints multivariate
- [Dashboard](https://pulse.institutoitinerante.com.br) — visuals
- [Asynchronous A/B Testing](https://blog.amplitude.com/ab-testing-guide) — teoria
