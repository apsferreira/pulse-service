---
title: "Instalar o SDK JavaScript"
description: "Como instalar e configurar o @pulse/js em React, Next.js, Astro e projetos vanilla TypeScript."
---

# Instalar o SDK JavaScript

O `@pulse/js` é o SDK oficial do Pulse para aplicações JavaScript e TypeScript. Suporta React, Next.js (SSR/SSG), Astro e qualquer bundler moderno.

---

## Instalação

```bash
npm install @pulse/js
# ou
yarn add @pulse/js
# ou
pnpm add @pulse/js
```

Requisitos: Node.js 18+ ou ambiente de browser moderno (Chrome 90+, Firefox 88+, Safari 14+).

---

## Inicialização básica

```typescript
import { PulseSDK } from '@pulse/js'

const pulse = new PulseSDK({
  apiUrl: 'https://pulse-api.institutoitinerante.com.br',
  product: 'minha-app',                        // slug do projeto no dashboard
  serviceToken: process.env.PULSE_SERVICE_TOKEN,
  userId: 'user-123',                          // recomendado para rollout consistente
  tenantId: 'tenant-456',                      // opcional, para multi-tenant
  cacheTime: 60_000,                           // cache local em ms (padrão: 60s)
})

// Carregar flags (faz uma chamada à API e armazena em cache)
await pulse.getFlags()

// Verificar flag
if (pulse.isEnabled('nova_homepage')) {
  console.log('Exibindo nova homepage')
}

// Rastrear evento
await pulse.track('page_viewed', { path: '/home' })
```

---

## Variáveis de ambiente

Nunca coloque o `SERVICE_TOKEN` no código-fonte. Use variáveis de ambiente:

**React / Vite:**
```
# .env.local
VITE_PULSE_TOKEN=seu-token-aqui
VITE_PULSE_PRODUCT=minha-app
```

```typescript
const pulse = new PulseSDK({
  apiUrl: 'https://pulse-api.institutoitinerante.com.br',
  product: import.meta.env.VITE_PULSE_PRODUCT,
  serviceToken: import.meta.env.VITE_PULSE_TOKEN,
})
```

**Next.js:**
```
# .env.local
NEXT_PUBLIC_PULSE_TOKEN=seu-token-aqui
NEXT_PUBLIC_PULSE_PRODUCT=minha-app
```

```typescript
const pulse = new PulseSDK({
  apiUrl: 'https://pulse-api.institutoitinerante.com.br',
  product: process.env.NEXT_PUBLIC_PULSE_PRODUCT!,
  serviceToken: process.env.NEXT_PUBLIC_PULSE_TOKEN!,
})
```

> Tokens com prefixo `NEXT_PUBLIC_` ou `VITE_` ficam expostos no bundle do cliente. Para produção com alta segurança, prefira fazer a avaliação de flags no servidor e passar os resultados via props ou contexto.

---

## React — hook `useFeatureFlag`

O SDK exporta um hook pronto para uso:

```typescript
// src/hooks/usePulse.ts
import { createPulseContext } from '@pulse/js/react'

export const { PulseProvider, useFeatureFlag, usePulse } = createPulseContext()
```

```typescript
// src/main.tsx
import { PulseProvider } from './hooks/usePulse'

const pulseConfig = {
  apiUrl: 'https://pulse-api.institutoitinerante.com.br',
  product: import.meta.env.VITE_PULSE_PRODUCT,
  serviceToken: import.meta.env.VITE_PULSE_TOKEN,
  userId: currentUser?.id,
}

root.render(
  <PulseProvider config={pulseConfig}>
    <App />
  </PulseProvider>
)
```

```typescript
// src/components/Checkout.tsx
import { useFeatureFlag } from '../hooks/usePulse'

export function Checkout() {
  const novoCheckout = useFeatureFlag('novo_checkout_v2')

  if (novoCheckout === undefined) {
    return <CheckoutSkeleton /> // carregando
  }

  return novoCheckout ? <CheckoutV2 /> : <CheckoutV1 />
}
```

O hook retorna `undefined` enquanto as flags estão carregando, `true` ou `false` após resolução. Trate o estado de loading explicitamente para evitar flash de conteúdo.

---

## Next.js — SSR guard

O SDK usa `fetch` e `localStorage`, que não existem no servidor. Use o guard `typeof window` ou a diretiva `'use client'`:

**App Router (Next.js 13+):**
```typescript
// src/components/FeatureGate.tsx
'use client'

import { useFeatureFlag } from '../hooks/usePulse'

export function NovaFeatureGate({ children }: { children: React.ReactNode }) {
  const enabled = useFeatureFlag('nova_feature')
  if (!enabled) return null
  return <>{children}</>
}
```

**Pages Router com getServerSideProps:**
```typescript
// pages/index.tsx
import type { GetServerSideProps } from 'next'

export const getServerSideProps: GetServerSideProps = async (ctx) => {
  // Avaliação server-side via API direta (sem SDK de browser)
  const res = await fetch(
    `https://pulse-api.institutoitinerante.com.br/api/v1/flags?product=minha-app`,
    { headers: { 'X-Service-Token': process.env.PULSE_SERVICE_TOKEN! } }
  )
  const flags = await res.json()

  return { props: { flags } }
}

export default function Home({ flags }: { flags: Record<string, boolean> }) {
  return flags.nova_homepage ? <HomepageV2 /> : <HomepageV1 />
}
```

**Guard manual (vanilla):**
```typescript
// Seguro para usar em qualquer contexto
const pulse =
  typeof window !== 'undefined'
    ? new PulseSDK({ ... })
    : null

if (pulse) {
  await pulse.getFlags()
}
```

---

## Astro

Em Astro, componentes `.astro` são renderizados no servidor por padrão. Use componentes React com `client:load` para flags que dependem do browser:

```astro
---
// src/pages/index.astro
// Avaliação server-side (sem SDK JS)
const res = await fetch(
  `https://pulse-api.institutoitinerante.com.br/api/v1/flags?product=minha-app`,
  { headers: { 'X-Service-Token': import.meta.env.PULSE_SERVICE_TOKEN } }
)
const flags = await res.json()
---

{flags.nova_hero ? <HeroV2 /> : <HeroV1 />}

<!-- Componente React com flag dinâmica (client-side) -->
<CheckoutButton client:load />
```

Para componentes React interativos no Astro, use o `PulseProvider` dentro do componente com `client:load`.

---

## Rastrear eventos

```typescript
// Evento simples
await pulse.track('button_clicked', { buttonId: 'cta-hero' })

// Evento de conversão (para A/B test)
await pulse.track('checkout_completed', {
  orderId: 'ord_123',
  value: 149.90,
  currency: 'BRL',
})

// Identificar usuário depois do login
pulse.identify('user-456', { plan: 'growth', company: 'Acme' })
```

---

## Referência de opções do construtor

| Opção | Tipo | Padrão | Descrição |
|-------|------|--------|-----------|
| `apiUrl` | `string` | — | URL base da API Pulse |
| `product` | `string` | — | Slug do projeto |
| `serviceToken` | `string` | — | Token de autenticação |
| `userId` | `string` | `undefined` | ID do usuário para rollout consistente |
| `tenantId` | `string` | `undefined` | ID do tenant (multi-tenant) |
| `cacheTime` | `number` | `60000` | TTL do cache local em ms |
| `timeout` | `number` | `5000` | Timeout das chamadas HTTP em ms |
| `onError` | `function` | `console.error` | Callback de erro personalizado |

---

## Próximos passos

- [03-debug-flag-nao-ativa.md](./03-debug-flag-nao-ativa.md) — flag habilitada mas não aparece
- [04-ab-test-interpretar-p-value.md](./04-ab-test-interpretar-p-value.md) — experimentos com análise estatística
- [sdk-reference.md](./sdk-reference.md) — referência completa da API do SDK
