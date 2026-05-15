---
title: "Limite de eventos — o que fazer"
description: "Como agir quando você recebe HTTP 429, como ler o cabeçalho X-Quota-Remaining, otimizar o tracking e decidir quando fazer upgrade."
---

# Limite de eventos — o que fazer

Quando o seu plano atinge a cota mensal de eventos, a API passa a retornar `HTTP 429 Too Many Requests` para chamadas de tracking. Feature flags continuam funcionando — apenas o rastreamento de eventos é bloqueado.

---

## Identificar o problema

### HTTP 429 na resposta

```json
{
  "error": "quota_exceeded",
  "message": "Monthly event quota reached. Upgrade your plan or wait for reset.",
  "quota_limit": 10000,
  "quota_used": 10000,
  "reset_at": "2026-05-01T00:00:00Z"
}
```

### Cabeçalhos de quota

Toda resposta da API inclui os cabeçalhos de quota, mesmo antes de atingir o limite:

```
X-Quota-Limit: 10000
X-Quota-Remaining: 423
X-Quota-Reset: 2026-05-01T00:00:00Z
X-Quota-Used: 9577
```

Leia esses cabeçalhos para implementar alertas proativos antes de atingir o limite.

### Verificar quota via API

```bash
curl -s "https://pulse.institutoitinerante.com.br/api/v1/quota" \
  -H "X-Service-Token: $SERVICE_TOKEN" | python3 -m json.tool
```

```json
{
  "plan": "free",
  "quota_limit": 10000,
  "quota_used": 9577,
  "quota_remaining": 423,
  "overage_events": 0,
  "overage_cost_brl": 0.00,
  "reset_at": "2026-05-01T00:00:00Z",
  "percentage_used": 95.77
}
```

---

## Tratar 429 no SDK

O SDK JavaScript faz retry automático com backoff exponencial para erros de rede, mas **não** para 429 (quota). Implemente tratamento explícito:

```typescript
// Implementar handler de quota no SDK
const pulse = new PulseSDK({
  apiUrl: 'https://pulse.institutoitinerante.com.br',
  product: 'minha-app',
  serviceToken: process.env.PULSE_SERVICE_TOKEN!,
  onError: (error) => {
    if (error.status === 429) {
      // Silenciar silenciosamente — não logar como erro crítico
      // Feature flags continuam funcionando
      console.warn('Pulse: quota de eventos atingida, tracking pausado até reset mensal')
    } else {
      console.error('Pulse error:', error)
    }
  },
})
```

Em Go, o retorno de `Track()` inclui o erro:

```go
err := client.Track(userID, "event_name", props)
if err != nil {
  if errors.Is(err, pulse.ErrQuotaExceeded) {
    // Logar warning, não erro crítico
    logger.Warn("pulse quota exceeded, event dropped", "event", "event_name")
    return  // não fazer panic, não retornar 500 para o usuário
  }
  return err
}
```

---

## Otimizar o tracking — reduzir eventos desnecessários

Antes de fazer upgrade, analise quais eventos realmente importam para decisões de produto.

### Auditoria de eventos

```bash
# Ver os eventos mais frequentes dos últimos 7 dias
curl "https://pulse.institutoitinerante.com.br/api/v1/events/summary?product=minha-app&days=7" \
  -H "X-Service-Token: $SERVICE_TOKEN"
```

```json
{
  "top_events": [
    { "name": "page_viewed",        "count": 4821, "percentage": 48.2 },
    { "name": "button_hovered",     "count": 2103, "percentage": 21.0 },
    { "name": "scroll_depth_50pct", "count": 1840, "percentage": 18.4 },
    { "name": "checkout_completed", "count":  412, "percentage":  4.1 },
    { "name": "signup_completed",   "count":   88, "percentage":  0.9 }
  ]
}
```

No exemplo acima, `button_hovered` e `scroll_depth_50pct` consomem 39% da quota sem gerar insights acionáveis. Removê-los libera quase 4.000 eventos/mês.

### Filtrar eventos de baixo valor

```typescript
// Antes — rastreia tudo
document.querySelectorAll('button').forEach(btn => {
  btn.addEventListener('mouseover', () => {
    pulse.track('button_hovered', { id: btn.id })  // ruído, não rastrear
  })
})

// Depois — rastrear apenas cliques com intenção real
document.querySelectorAll('[data-track]').forEach(btn => {
  btn.addEventListener('click', () => {
    pulse.track('button_clicked', { id: btn.dataset.track })
  })
})
```

### Usar sampling para eventos de alto volume

Para eventos muito frequentes (ex: `page_viewed`), amostre 1 em cada N:

```typescript
function trackWithSampling(eventName: string, props: object, rate = 0.1) {
  // Amostrar 10% dos eventos de page_viewed
  if (Math.random() > rate) return
  pulse.track(eventName, { ...props, sampled: true, sample_rate: rate })
}

// 90% dos page_viewed são descartados no cliente
trackWithSampling('page_viewed', { path: location.pathname }, 0.1)

// Eventos de conversão sempre rastreados (100%)
pulse.track('checkout_completed', { orderId, value })
```

### Debounce para eventos de scroll/resize

```typescript
function debounce(fn: Function, ms: number) {
  let timer: ReturnType<typeof setTimeout>
  return (...args: unknown[]) => {
    clearTimeout(timer)
    timer = setTimeout(() => fn(...args), ms)
  }
}

const trackScroll = debounce((depth: number) => {
  pulse.track('scroll_depth', { depth })
}, 2000)  // máximo 1 evento a cada 2s por scroll contínuo

window.addEventListener('scroll', () => {
  const depth = Math.round((window.scrollY / document.body.scrollHeight) * 100)
  if (depth === 25 || depth === 50 || depth === 75 || depth === 100) {
    trackScroll(depth)
  }
})
```

---

## Overage pricing

Nos planos **Starter** e **Growth**, eventos acima da cota mensal são cobrados como overage ao invés de bloquear:

| Plano | Overage por 1.000 eventos |
|-------|--------------------------|
| Free | Bloqueado (HTTP 429) |
| Starter | R$ 4,00 |
| Growth | R$ 2,50 |
| Pro | Ilimitado (incluso) |

O overage é cobrado automaticamente no cartão cadastrado ao final do ciclo. Você pode definir um teto de overage em **Settings → Quota → Limite de Overage** — ao atingir o teto, o comportamento volta a ser HTTP 429.

---

## Quando fazer upgrade

Faça upgrade quando qualquer uma dessas condições for verdade:

1. **Você atinge a cota antes do dia 20 do mês** com frequência.
2. **Você está filtrando eventos que queria rastrear** só para caber na cota.
3. **O custo de overage mensal** ultrapassa a diferença de preço para o plano superior.
4. **Você precisa de retenção maior** (audit, compliance, análise histórica).

Para calcular o ponto exato de quando upgrade compensa: [05-qual-plano-escolher.md](./05-qual-plano-escolher.md).

---

## Próximos passos

- [05-qual-plano-escolher.md](./05-qual-plano-escolher.md) — comparar planos e calcular ROI do upgrade
- [07-exportar-dados-testes.md](./07-exportar-dados-testes.md) — exportar eventos para análise antes de otimizar
