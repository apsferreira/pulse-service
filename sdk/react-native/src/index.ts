/**
 * @pulse-iit/react-native
 * SDK React Native / Expo para Pulse — Feature Flags + A/B Testing + Event Tracking
 *
 * Uso básico:
 *
 *   // App.tsx
 *   import { PulseProvider, useFeatureFlag } from '@pulse-iit/react-native'
 *
 *   function MyApp() {
 *     const { value: newOnboarding } = useFeatureFlag('new-onboarding-flow')
 *     return newOnboarding ? <NewOnboarding /> : <OldOnboarding />
 *   }
 *
 *   export default function App() {
 *     return (
 *       <PulseProvider config={{ apiUrl: '...', projectKey: '...', serviceToken: '...' }}>
 *         <MyApp />
 *       </PulseProvider>
 *     )
 *   }
 */

// Tipos
export type {
  ExperimentResult,
  FeatureFlag,
  FeatureFlagResult,
  FeatureFlagsResult,
  FlagsMap,
  PulseConfig,
  TrackEvent,
  VariantsMap,
} from './types'

// Cliente singleton + helpers funcionais
export { identify, pulseClient, PulseClient, track } from './client'

// Context + Provider
export type { PulseContextValue, PulseProviderProps } from './context'
export { PulseProvider } from './context'

// Hooks
export {
  useExperiment,
  useFeatureFlag,
  useFeatureFlags,
  useFlag,
  useIdentify,
  usePulse,
  usePulseReady,
  useReloadFlags,
  useTrack,
} from './hooks'
