# Pulse SDK — Referência Completa

> Versão 1.1 | Atualizado: 2026-05

---

## Instalação

### npm / pnpm / yarn

```bash
npm install @pulse/js
# ou
pnpm add @pulse/js
# ou
yarn add @pulse/js
```

### CDN (sem bundler)

```html
<!-- Servido diretamente pela sua instância Pulse -->
<script src="https://seu-pulse.seudominio.com/sdk/pulse.js"></script>
<script>
  PulseSDK.init({
    projectKey: 'pulse_xxxx',
    apiUrl: 'https://seu-pulse.seudominio.com',
  })
  PulseSDK.track('page_view', { page: '/home' })
</script>
```

> O arquivo `pulse.js` é servido pelo próprio backend Pulse em `/sdk/pulse.js`. Não há CDN externo — sempre use a URL da sua instância.

---

## Inicialização

### JavaScript / TypeScript

```typescript
import { PulseSDK } from '@pulse/js'

PulseSDK.init({
  projectKey: 'pulse_xxxx',              // Obrigatório: API key do projeto no dashboard
  apiUrl: 'https://seu-pulse.seudominio.com', // Obrigatório: URL da sua instância
  debug: false,                          // Opcional: loga eventos no console
  flushInterval: 5000,                   // Opcional: flush a cada 5s (default: 5000ms)
  maxBatchSize: 25,                      // Opcional: flush ao atingir N eventos
})
```

### Next.js (App Router)

```typescript
// app/providers.tsx
'use client'
import { useEffect } from 'react'
import { PulseSDK } from '@pulse/js'

export function PulseProvider({ children }: { children: React.ReactNode }) {
  useEffect(() => {
    if (typeof window !== 'undefined') {
      PulseSDK.init({
        projectKey: process.env.NEXT_PUBLIC_PULSE_KEY!,
        apiUrl: process.env.NEXT_PUBLIC_PULSE_URL!,
      })
    }
  }, [])
  return <>{children}</>
}
```

### Astro SSR

```typescript
// src/scripts/pulse.ts — apenas client-side
if (typeof window !== 'undefined') {
  const { PulseSDK } = await import('@pulse/js')
  PulseSDK.init({
    projectKey: import.meta.env.PUBLIC_PULSE_KEY,
    apiUrl: import.meta.env.PUBLIC_PULSE_URL,
  })
}
```

---

## track() — Rastrear Eventos

```typescript
// Evento simples
PulseSDK.track('button_clicked')

// Com propriedades
PulseSDK.track('checkout_started', {
  plan: 'growth',
  amount_cents: 9900,
  currency: 'BRL',
})

// Com user_id explícito
PulseSDK.track('feature_used', { feature: 'export_pdf' }, { userId: 'user_123' })

// Modo anônimo (antes do login)
PulseSDK.track('page_view', { page: '/pricing' }, { anonymousId: 'anon_abc123' })
```

---

## identify() — Vincular Usuário Anônimo

Após login, associe `anonymousId` ao `userId` real para unificar a jornada:

```typescript
PulseSDK.identify({
  userId: 'user_123',
  anonymousId: 'anon_abc123', // opcional
  traits: {
    email: 'user@empresa.com',
    plan: 'growth',
  },
})
```

---

## Feature Flags

### Hook React

```typescript
import { useFeatureFlag, useVariant } from '@pulse/js/react'

function CheckoutPage() {
  const isNewUX = useFeatureFlag('new_checkout_ux')
  return isNewUX ? <CheckoutV2 /> : <CheckoutV1 />
}

function PricingPage() {
  const variant = useVariant('pricing_experiment') // 'control' | 'treatment' | null
  return variant === 'treatment' ? <NewPricing /> : <OriginalPricing />
}
```

### Imperativo (qualquer framework)

```typescript
// Verificar flag
const isEnabled = await PulseSDK.getFeatureFlag('nova_feature', false)
//                                                               ^^^
//                               Valor padrão se API estiver inacessível (fail-open)

// Obter variante A/B
const variant = await PulseSDK.getVariant('ab_test_key', 'control')
```

---

## React Native / Expo

```typescript
import { PulseSDK } from '@pulse/react-native'

// App.tsx — uma vez
PulseSDK.init({
  projectKey: 'pulse_xxxx',
  apiUrl: 'https://seu-pulse.seudominio.com',
  // AsyncStorage é usado automaticamente para persistência
})

// Hook
import { useFeatureFlag } from '@pulse/react-native'

function HomeScreen() {
  const showBanner = useFeatureFlag('home_banner_v2')
  return showBanner ? <BannerV2 /> : <BannerV1 />
}
```

---

## SDK Go (backend)

```go
import pulse "github.com/seu-org/pulse-service/sdk/go"

client, err := pulse.NewClient("https://seu-pulse.seudominio.com")
if err != nil {
    log.Fatal(err)
}

// Feature flag
enabled, err := client.IsEnabled(ctx, "nova_feature")
if err != nil {
    enabled = false // fail-open
}

// Evento
client.TrackEvent(ctx, pulse.TrackEventRequest{
    EventName:  "user_signup",
    UserID:     &userID,
    Properties: map[string]interface{}{"plan": "free"},
})
```

---

## SDK Node.js (backend)

```typescript
import { PulseClient } from '@pulse/node'

const pulse = new PulseClient({
  baseURL: 'https://seu-pulse.seudominio.com',
  projectKey: process.env.PULSE_PROJECT_KEY!,
})

const enabled = await pulse.isEnabled('nova_feature', { userId: 'user_123' })
await pulse.track('order_placed', { amount: 150.00 })
```

---

## Configuração por Ambiente

Use `api_key` diferentes por ambiente para isolar dados de staging/prod:

```typescript
PulseSDK.init({
  projectKey: process.env.NODE_ENV === 'production'
    ? process.env.PULSE_KEY_PROD
    : process.env.PULSE_KEY_STAGING,
  apiUrl: 'https://seu-pulse.seudominio.com',
})
```

---

## Tratamento de Erros e Fallbacks

O SDK é **fail-open**: erros de rede nunca lançam exceções no seu código.

```typescript
// Fallback explícito — retorna `false` se API inacessível
const isEnabled = await PulseSDK.getFeatureFlag('minha_flag', false)

// O SDK armazena o último valor válido:
// - Em memória: 30 segundos
// - Em localStorage: 5 minutos
// Mudanças propagam em <10s via Redis após alterar no dashboard.
```

---

## Flush Manual

```typescript
// Útil antes de navegação, logout ou encerrar processo Node
await PulseSDK.flush()
```

---

## Referência Completa de Opções

```typescript
PulseSDK.init({
  projectKey: string,              // Obrigatório
  apiUrl: string,                  // Obrigatório — URL da sua instância Pulse
  debug?: boolean,                 // Default: false
  flushInterval?: number,          // Default: 5000 (ms)
  maxBatchSize?: number,           // Default: 25 eventos
  sessionTimeout?: number,         // Default: 30 (minutos)
  disableAutoPageview?: boolean,   // Default: false
})
```

---

## Rate Limits e Quotas

| Plano | Eventos/mês | Comportamento ao atingir limite |
|-------|------------|--------------------------------|
| Free | 10.000 | HTTP 429 — SDK armazena e tenta re-enviar no próximo ciclo |
| Starter | 500.000 | HTTP 429 + header `X-Quota-Remaining` |
| Growth | 2.000.000 | HTTP 429 + header `X-Quota-Remaining` |
| Pro | Ilimitado | — |

Self-hosted: quotas são desabilitadas por padrão (variável `SELF_HOSTED=true`).

---

## Tradeoffs — Quando usar cada abordagem

| Cenário | Recomendação |
|---------|-------------|
| SPA (React, Vue, Svelte) | `@pulse/js` + hooks — cache automático, fail-open nativo |
| SSR (Next.js, Remix) | `@pulse/js` no client-side apenas (SSR guard obrigatório) |
| Backend Go | `sdk/go` — avaliação server-side, sem latência de rede extra |
| Backend Node | `@pulse/node` — ideal para microsserviços JS |
| Mobile | `@pulse/react-native` — AsyncStorage para persistência offline |
| Feature crítica de segurança | Avalie no backend — nunca confie em avaliação client-side |
| Alta frequência de eventos (>100/s) | Use batch + `flushInterval` curto; monitore Redis |

---

## Suporte

- **Documentação:** `docs/` no repositório
- **Issues:** abra uma issue no GitHub com label `sdk`
