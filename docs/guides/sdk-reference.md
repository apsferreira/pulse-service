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

## Identificação anônima e merge após login

### Como funciona

Todo visitante sem login recebe um `anonymousId` gerado automaticamente e persistido em `localStorage`. Isso garante que:

1. O mesmo visitante sempre cai no **mesmo bucket** de rollout entre sessões
2. Todos os eventos pré-login ficam associados a um ID consistente
3. Após o login, um único `identify()` retroativamente une os eventos ao usuário real

```
Visita 1 → anonymousId: "anon_x7k2p9"
  track('page_view')     → user_id: "anon_x7k2p9"
  track('cta_click')     → user_id: "anon_x7k2p9"

Login → userId: "user_123"
  identify({ anonymous_id: "anon_x7k2p9", user_id: "user_123" })
  → UPDATE events SET user_id='user_123' WHERE user_id='anon:anon_x7k2p9'
  → Todos os eventos anteriores agora pertencem a "user_123"

Eventos seguintes → user_id: "user_123" (direto, sem anonymousId)
```

### Gerar e persistir o anonymousId (JS/TS)

O SDK expõe um helper para isso. Guarde o ID retornado e passe-o nos eventos enquanto o usuário não estiver logado:

```typescript
// utils/identity.ts
const ANON_KEY = 'pulse_anonymous_id'

export function getAnonymousId(): string {
  const existing = localStorage.getItem(ANON_KEY)
  if (existing) return existing
  const id = 'anon_' + crypto.randomUUID()
  localStorage.setItem(ANON_KEY, id)
  return id
}

export function clearAnonymousId(): void {
  localStorage.removeItem(ANON_KEY)
}
```

### identify() — merge após login

```typescript
import { getAnonymousId, clearAnonymousId } from './utils/identity'

// Após login bem-sucedido:
async function onLoginSuccess(userId: string) {
  const anonymousId = getAnonymousId()

  await pulse.identify({
    anonymous_id: anonymousId,
    user_id: userId,
  })
  // → { merged_events: 12, anonymous_id: "anon_...", user_id: "user_123" }

  // Limpa o anonymousId — a partir daqui usar user_id diretamente
  clearAnonymousId()
}
```

### Rollout determinístico com usuários anônimos

O bucket de rollout é calculado por `sha256(userId + ":" + flagKey) % 100`. Com o `anonymousId` fixo em `localStorage`, o visitante sempre cai no mesmo grupo — mesmo voltando dias depois. Após o `identify()`, o mesmo usuário mantém consistência porque o `user_id` real passa a ser o identificador.

> **Sites sem login** (ex: landing pages): use o `anonymousId` como `user_id` nos eventos. Não há merge posterior, mas o rastreamento é consistente por dispositivo.

> **Apps com login** (Librinho, Nitro, SocialMake): chame `identify()` no callback de login, antes de redirecionar para a tela autenticada.

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

### Forçar grupo por URL (QA / demos)

Para testar os dois lados de uma flag sem alterar o backend, adicione `pulse_override` na URL:

```
# Forçar ON
https://seuapp.com/checkout?pulse_override[nome-da-flag]=on

# Forçar OFF
https://seuapp.com/checkout?pulse_override[nome-da-flag]=off

# Forçar uma variante específica de A/B test
https://seuapp.com/checkout?pulse_override[nome-da-flag]=challenger
```

O SDK JS lê esse parâmetro **automaticamente** a cada `isEnabled()` — nenhuma configuração necessária. Funciona em qualquer framework (React, Vue, Astro, vanilla).

> **Onde gerar o link:** abra a flag no dashboard → clique em **Editar** → seção **Testar avaliação** → digite um `user_id` → o link de override é gerado e pode ser copiado com um clique.

---

## Flutter / Dart

Adicione o SDK ao `pubspec.yaml`:

```yaml
dependencies:
  pulse_flutter:
    path: ../pulse-service/sdk/flutter  # dev local
    # ou via git:
    # git:
    #   url: https://github.com/apsferreira/pulse-service
    #   path: sdk/flutter
```

### Inicialização

```dart
// main.dart
import 'package:flutter/foundation.dart';
import 'package:pulse_flutter/pulse_flutter.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await Pulse.init(
    projectKey: 'librinho',
    debug: kDebugMode,
  );
  runApp(const MyApp());
}
```

### Rastrear eventos

```dart
Pulse.instance.track('book_scanned', properties: {'isbn': '978...'});
```

### Feature flags

```dart
final enabled = await Pulse.instance.isEnabled('new-reader-ui');
```

### Identificar usuário após login

```dart
Pulse.instance.identify(userId);
```

### Flush ao entrar em background

```dart
// No AppLifecycleListener ou em didChangeAppLifecycleState:
Pulse.instance.flush();
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
| Scale | 10.000.000 | HTTP 429 + header `X-Quota-Remaining` |
| Self-Hosted | Ilimitado | — |

Self-hosted: quotas são desabilitadas por padrão (variável `SELF_HOSTED=true`).

---

## Tradeoffs — Quando usar cada abordagem

| Cenário | Recomendação |
|---------|-------------|
| SPA (React, Vue, Svelte) | `@pulse/js` + hooks — cache automático, fail-open nativo |
| SSR (Next.js, Remix) | `@pulse/js` no client-side apenas (SSR guard obrigatório) |
| Backend Go | `sdk/go` — avaliação server-side, sem latência de rede extra |
| Backend Node | `@pulse/node` — ideal para microsserviços JS |
| Mobile Flutter (Android/iOS) | `pulse_flutter` em `sdk/flutter/` — persistência offline nativa |
| Feature crítica de segurança | Avalie no backend — nunca confie em avaliação client-side |
| Alta frequência de eventos (>100/s) | Use batch + `flushInterval` curto; monitore Redis |

---

## Suporte

- **Documentação:** `docs/` no repositório
- **Issues:** abra uma issue no GitHub com label `sdk`
