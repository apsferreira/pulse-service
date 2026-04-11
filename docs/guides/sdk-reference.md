# Pulse SDK — Referência Completa

> BKL-1060 | Documentação pública do SDK | Versão 1.0 | 2026-04-10

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
<!-- IIFE — funciona direto no browser sem bundler -->
<script src="https://pulse-api.institutoitinerante.com.br/sdk/pulse.js"></script>
<script>
  PulseSDK.init({ projectKey: 'pulse_xxxx' })
  PulseSDK.track('page_view', { page: '/home' })
</script>
```

---

## Inicialização

### JavaScript / TypeScript

```typescript
import { PulseSDK } from '@pulse/js'

PulseSDK.init({
  projectKey: 'pulse_xxxx',   // Obrigatório: API key do seu projeto
  apiUrl: 'https://pulse-api.institutoitinerante.com.br', // Opcional: default é a URL pública
  debug: false,               // Opcional: loga eventos no console (default: false)
  flushInterval: 5000,        // Opcional: flush a cada 5s (default: 5000ms)
  maxBatchSize: 25,           // Opcional: flush ao atingir 25 eventos (default: 25)
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
    // SSR guard: window só existe no browser
    if (typeof window !== 'undefined') {
      PulseSDK.init({ projectKey: process.env.NEXT_PUBLIC_PULSE_KEY! })
    }
  }, [])
  return <>{children}</>
}
```

### Astro SSR

```typescript
// src/scripts/pulse.ts
// Importar apenas em scripts client-side (não no servidor)
if (typeof window !== 'undefined') {
  const { PulseSDK } = await import('@pulse/js')
  PulseSDK.init({ projectKey: import.meta.env.PUBLIC_PULSE_KEY })
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

// Com user_id
PulseSDK.track('feature_used', {
  feature: 'export_pdf',
}, {
  userId: 'user_123',
})

// Modo anônimo (antes do login)
PulseSDK.track('page_view', {
  page: '/pricing',
}, {
  anonymousId: 'anon_abc123',
})
```

---

## identify() — Vincular Usuário Anônimo

Após o login, associe o `anonymousId` ao `userId` real para unificar a jornada:

```typescript
// Após login bem-sucedido
PulseSDK.identify({
  userId: 'user_123',
  anonymousId: 'anon_abc123', // Opcional: se rastreou antes do login
  traits: {
    email: 'user@empresa.com',
    plan: 'growth',
    company: 'Empresa LTDA',
  },
})
```

---

## useFeatureFlag() — Feature Flags em React

```typescript
import { useFeatureFlag } from '@pulse/js/react'

function MyComponent() {
  const isEnabled = useFeatureFlag('nova_checkout_ux')

  if (!isEnabled) return <CheckoutV1 />
  return <CheckoutV2 />
}
```

### Com variante A/B

```typescript
import { useVariant } from '@pulse/js/react'

function PricingPage() {
  const variant = useVariant('pricing_experiment') // 'control' | 'treatment' | null

  if (variant === 'treatment') return <NewPricingLayout />
  return <OriginalPricingLayout />
}
```

---

## React Native / Expo

```typescript
import { PulseSDK } from '@pulse/react-native'

// Inicializar uma vez no App.tsx
PulseSDK.init({
  projectKey: 'pulse_xxxx',
  // AsyncStorage é usado automaticamente para persistência de userId
})

// Rastrear evento
PulseSDK.track('screen_view', { screen: 'Home' })

// Hook (Expo + React Native)
import { useFeatureFlag } from '@pulse/react-native'

function HomeScreen() {
  const showBanner = useFeatureFlag('home_banner_v2')
  // ...
}
```

---

## Configuração de Ambientes (BKL-1075)

Use API keys diferentes por ambiente para isolar dados de staging/prod:

```typescript
const isProd = process.env.NODE_ENV === 'production'

PulseSDK.init({
  projectKey: isProd
    ? process.env.PULSE_KEY_PROD       // env_prod_xxxx
    : process.env.PULSE_KEY_STAGING,   // env_staging_xxxx
})
```

---

## Tratamento de Erros e Fallbacks

O SDK é **fail-open**: erros de rede nunca lançam exceções para o seu código.

```typescript
// getFeatureFlag() com fallback explícito
const isEnabled = await PulseSDK.getFeatureFlag('minha_flag', false)
//                                                               ^^^
//                              Valor padrão se o servidor estiver inacessível

// O SDK armazena o último valor válido em memória (30s) e localStorage (5min)
// como fallback automático durante downtime.
```

---

## Flush Manual

O SDK envia eventos em batch automaticamente. Para forçar envio imediato:

```typescript
// Útil antes de navegação ou logout
await PulseSDK.flush()
```

---

## Opções Completas de Inicialização

```typescript
PulseSDK.init({
  projectKey: string,           // Obrigatório
  apiUrl?: string,              // Default: 'https://pulse-api.institutoitinerante.com.br'
  debug?: boolean,              // Default: false
  flushInterval?: number,       // Default: 5000 (ms)
  maxBatchSize?: number,        // Default: 25 eventos
  sessionTimeout?: number,      // Default: 30 (minutos)
  disableAutoPageview?: boolean, // Default: false — desabilita page_view automático
})
```

---

## Exemplos Completos

### React SPA com login

```typescript
// App.tsx
import { useEffect } from 'react'
import { PulseSDK } from '@pulse/js'

export default function App() {
  useEffect(() => {
    PulseSDK.init({ projectKey: import.meta.env.VITE_PULSE_KEY })
  }, [])

  const handleLogin = async (user: User) => {
    await loginAPI(user)
    PulseSDK.identify({ userId: user.id, traits: { plan: user.plan } })
    PulseSDK.track('user_logged_in', { method: 'email' })
  }

  return <Router />
}
```

---

## Rate Limits e Quotas

| Plano | Eventos/mês | Resposta ao atingir limite |
|-------|------------|--------------------------|
| Free | 10.000 | HTTP 429 — SDK armazena localmente e tenta re-enviar no próximo mês |
| Starter | 500.000 | HTTP 429 + header X-Quota-Remaining |
| Growth | 2.000.000 | HTTP 429 + header X-Quota-Remaining |
| Pro | Ilimitado | — |

O SDK respeita o header `X-Quota-Remaining` e ajusta a frequência de envio automaticamente.

---

## Suporte

- **Docs:** pulse.institutoitinerante.com.br/docs
- **Email:** suporte@institutoitinerante.com.br
- **Issues:** github.com/institutoitinerante/pulse-service/issues
