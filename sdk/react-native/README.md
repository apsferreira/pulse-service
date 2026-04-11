# @pulse-iit/react-native

SDK React Native / Expo para o Pulse — Feature Flags, A/B Testing e Event Tracking.

## Instalação

```bash
# No projeto Expo (Libri ou Nitro)
yarn add @pulse-iit/react-native
yarn add @react-native-async-storage/async-storage

# Ou via path local (monorepo / link direto)
yarn add ../../../pulse-service/sdk/react-native
```

## Uso básico

```tsx
// App.tsx
import { PulseProvider, useFeatureFlag } from '@pulse-iit/react-native'

function MyApp() {
  const { value: newOnboarding, loading } = useFeatureFlag('new-onboarding-flow')

  if (loading) return null // ou <Skeleton />

  return newOnboarding ? <NewOnboarding /> : <OldOnboarding />
}

export default function App() {
  return (
    <PulseProvider
      config={{
        apiUrl: 'https://pulse-api.institutoitinerante.com.br',
        projectKey: 'YOUR_PROJECT_KEY',
        serviceToken: 'YOUR_SERVICE_TOKEN',
      }}
    >
      <MyApp />
    </PulseProvider>
  )
}
```

## API

### `<PulseProvider config={...}>`

Inicializa o SDK e provê contexto para todos os hooks filhos.

```ts
interface PulseConfig {
  apiUrl: string        // URL base da API Pulse
  projectKey: string    // X-Project-Key do projeto
  serviceToken?: string // X-Service-Token (opcional)
  timeout?: number      // ms, default 10000
  debug?: boolean       // logs no console, default false
}
```

### Hooks

#### `useFeatureFlag(key, defaultValue?)`

```tsx
const { value, loading } = useFeatureFlag('new-onboarding-flow')
```

#### `useFeatureFlags()`

```tsx
const { flags, loading } = useFeatureFlags()
// flags = { 'feature-a': true, 'feature-b': false }
```

#### `useFlag(key, defaultValue?)` — sem loading

```tsx
const newOnboarding = useFlag('new-onboarding-flow')
```

#### `useTrack()`

```tsx
const track = useTrack()
track('screen_viewed', { screen: 'Home' })
```

#### `useIdentify()`

```tsx
const identify = useIdentify()
// Chamar após login do usuário
identify('user_123', { name: 'João', plan: 'growth' })
```

#### `useExperiment(key, defaultVariant?)`

```tsx
const { variant, trackExposure } = useExperiment('onboarding-test')
useEffect(() => { trackExposure() }, [trackExposure])
// variant = 'control' | 'treatment-a' | ...
```

#### `usePulseReady()`

```tsx
const ready = usePulseReady() // true quando flags foram carregadas
```

#### `useReloadFlags()`

```tsx
const reload = useReloadFlags()
await reload() // invalida cache e recarrega do servidor
```

### Funções fora de componentes

```ts
import { track, identify, pulseClient } from '@pulse-iit/react-native'

// Usar após pulseClient.init() ter sido chamado pelo PulseProvider
track('background_sync_completed', { records: 42 })
identify('user_456', { plan: 'pro' })
```

## Cache

- Flags são cacheadas em memória com TTL de 5 minutos
- `@react-native-async-storage/async-storage` persiste o `anonymousId` e `userId` entre sessões
- `useReloadFlags()` invalida o cache imediatamente

## Integração com Libri e Nitro

```tsx
// apps/libri/App.tsx ou apps/nitro/App.tsx
import { PulseProvider } from '@pulse-iit/react-native'
import { PULSE_API_URL, PULSE_PROJECT_KEY } from '@env' // react-native-dotenv

export default function App() {
  return (
    <PulseProvider
      config={{
        apiUrl: PULSE_API_URL,
        projectKey: PULSE_PROJECT_KEY,
      }}
    >
      <RootNavigator />
    </PulseProvider>
  )
}
```
