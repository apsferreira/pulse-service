/**
 * Hooks públicos do @pulse-iit/react-native
 */

import { useCallback } from 'react'
import { usePulseContext } from './context'
import type {
  ExperimentResult,
  FeatureFlagResult,
  FeatureFlagsResult,
  FlagsMap,
} from './types'

// ─── usePulse ─────────────────────────────────────────────────────────────

/**
 * Acessa o PulseClient diretamente via contexto.
 *
 * @example
 * const pulse = usePulse()
 * pulse.track('button_pressed', { button: 'signup' })
 */
export { usePulseContext as usePulse }

// ─── useFeatureFlag ───────────────────────────────────────────────────────

/**
 * Retorna o valor de uma feature flag com indicador de loading.
 *
 * @example
 * const { value: newOnboarding, loading } = useFeatureFlag('new-onboarding-flow')
 * if (loading) return <Skeleton />
 * return newOnboarding ? <NewOnboarding /> : <OldOnboarding />
 */
export function useFeatureFlag(
  key: string,
  defaultValue = false,
): FeatureFlagResult {
  const { flags, ready } = usePulseContext()
  const value = flags !== null ? (flags[key] === true) : defaultValue
  return { value, loading: !ready }
}

// ─── useFeatureFlags ──────────────────────────────────────────────────────

/**
 * Retorna todas as feature flags com indicador de loading.
 *
 * @example
 * const { flags, loading } = useFeatureFlags()
 */
export function useFeatureFlags(): FeatureFlagsResult {
  const { flags, ready } = usePulseContext()
  return {
    flags: flags ?? ({} as FlagsMap),
    loading: !ready,
  }
}

// ─── useFlag (alias simples) ──────────────────────────────────────────────

/**
 * Retorna diretamente o boolean da flag (sem loading).
 * Use useFeatureFlag() quando precisar do estado de loading.
 *
 * @example
 * const newOnboarding = useFlag('new-onboarding')
 */
export function useFlag(flagKey: string, defaultValue = false): boolean {
  const { flags } = usePulseContext()
  if (flags === null) return defaultValue
  return flags[flagKey] === true
}

// ─── useTrack ─────────────────────────────────────────────────────────────

/**
 * Retorna função estável para registrar eventos de comportamento.
 *
 * @example
 * const track = useTrack()
 * track('screen_viewed', { screen: 'Home' })
 */
export function useTrack(): (
  eventName: string,
  properties?: Record<string, unknown>,
) => void {
  const { client } = usePulseContext()
  return useCallback(
    (eventName: string, properties?: Record<string, unknown>) =>
      void client.track(eventName, properties),
    [client],
  )
}

// ─── useIdentify ──────────────────────────────────────────────────────────

/**
 * Retorna função para identificar o usuário autenticado.
 *
 * @example
 * const identify = useIdentify()
 * identify('user_123', { name: 'João', plan: 'growth' })
 */
export function useIdentify(): (
  userId: string,
  traits?: Record<string, unknown>,
) => void {
  const { client } = usePulseContext()
  return useCallback(
    (userId: string, traits?: Record<string, unknown>) =>
      void client.identify(userId, traits),
    [client],
  )
}

// ─── useExperiment ────────────────────────────────────────────────────────

/**
 * Retorna variante de um experimento A/B + função para registrar exposição.
 *
 * @example
 * const { variant, trackExposure } = useExperiment('onboarding-test')
 * useEffect(() => { trackExposure() }, [trackExposure])
 */
export function useExperiment(
  flagKey: string,
  defaultVariant = 'control',
): ExperimentResult {
  const { variants, client } = usePulseContext()
  const variant = variants?.[flagKey] ?? defaultVariant

  const trackExposure = useCallback(() => {
    void client.track('$experiment_started', {
      experiment_name: flagKey,
      variant,
    })
  }, [client, flagKey, variant])

  return { variant, trackExposure }
}

// ─── usePulseReady ────────────────────────────────────────────────────────

/**
 * Retorna true quando o SDK terminou de carregar flags do servidor.
 */
export function usePulseReady(): boolean {
  return usePulseContext().ready
}

// ─── useReloadFlags ───────────────────────────────────────────────────────

/**
 * Retorna função que invalida o cache e recarrega todas as flags do servidor.
 */
export function useReloadFlags(): () => Promise<void> {
  return usePulseContext().reloadFlags
}
