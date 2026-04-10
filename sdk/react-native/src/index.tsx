/**
 * @pulse/react-native — SDK React Native / Expo para Pulse
 * Feature Flags + A/B Testing + Event Tracking
 *
 * Requisito: @react-native-async-storage/async-storage instalado
 *
 * Uso:
 *   import { PulseProvider, useFlag, useTrack, useExperiment } from '@pulse/react-native'
 *
 *   export default function App() {
 *     return (
 *       <PulseProvider projectKey="YOUR_KEY" apiUrl="https://pulse-api.institutoitinerante.com.br">
 *         <MainNavigator />
 *       </PulseProvider>
 *     )
 *   }
 *
 *   function HomeScreen() {
 *     const newOnboarding = useFlag('new-onboarding')
 *     const track = useTrack()
 *     const { variant, trackExposure } = useExperiment('onboarding-test')
 *
 *     useEffect(() => { trackExposure() }, [trackExposure])
 *
 *     return <Button onPress={() => track('home_cta_pressed')} title="Começar" />
 *   }
 */

import AsyncStorage from '@react-native-async-storage/async-storage'
import React, {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useRef,
  useState,
} from 'react'

// ─── Types ─────────────────────────────────────────────────────────────────

export interface ExperimentResult {
  variant: string
  trackExposure: () => void
}

interface FlagsMap {
  [key: string]: boolean
}

interface VariantsMap {
  [key: string]: string
}

// ─── Internal helpers ──────────────────────────────────────────────────────

function uuid(): string {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0
    return (c === 'x' ? r : (r & 0x3) | 0x8).toString(16)
  })
}

async function getOrCreateAnonymousId(): Promise<string> {
  try {
    const stored = await AsyncStorage.getItem('_pulse_anon_id')
    if (stored) return stored
    const id = 'anon_' + uuid()
    await AsyncStorage.setItem('_pulse_anon_id', id)
    return id
  } catch {
    return 'anon_' + uuid()
  }
}

// ─── Core fetch helpers ────────────────────────────────────────────────────

async function fetchGet(
  apiUrl: string,
  projectKey: string,
  path: string,
  timeout: number,
): Promise<unknown> {
  const url = apiUrl.replace(/\/+$/, '') + path
  const controller = new AbortController()
  const timer = setTimeout(() => controller.abort(), timeout)
  try {
    const res = await fetch(url, {
      method: 'GET',
      headers: { 'X-Project-Key': projectKey, Accept: 'application/json' },
      signal: controller.signal,
    })
    if (!res.ok) return null
    return await res.json()
  } catch {
    return null
  } finally {
    clearTimeout(timer)
  }
}

async function fetchPost(
  apiUrl: string,
  projectKey: string,
  path: string,
  body: unknown,
  timeout: number,
): Promise<void> {
  const url = apiUrl.replace(/\/+$/, '') + path
  const controller = new AbortController()
  const timer = setTimeout(() => controller.abort(), timeout)
  try {
    await fetch(url, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-Project-Key': projectKey,
      },
      body: JSON.stringify(body),
      signal: controller.signal,
    })
  } catch {
    // eventos são best-effort
  } finally {
    clearTimeout(timer)
  }
}

// ─── Context ───────────────────────────────────────────────────────────────

interface PulseContextValue {
  flags: FlagsMap | null
  variants: VariantsMap | null
  userId: string | null
  anonymousId: string | null
  ready: boolean
  identify: (userId: string, traits?: Record<string, unknown>) => void
  track: (eventName: string, properties?: Record<string, unknown>) => void
  reloadFlags: () => Promise<void>
}

const PulseContext = createContext<PulseContextValue | null>(null)

// ─── PulseProvider ─────────────────────────────────────────────────────────

export interface PulseProviderProps {
  projectKey: string
  apiUrl: string
  userId?: string
  traits?: Record<string, unknown>
  debug?: boolean
  timeout?: number
  children: React.ReactNode
}

export function PulseProvider({
  projectKey,
  apiUrl,
  userId: initialUserId,
  traits: initialTraits,
  debug = false,
  timeout = 10000,
  children,
}: PulseProviderProps) {
  const [flags, setFlags] = useState<FlagsMap | null>(null)
  const [variants, setVariants] = useState<VariantsMap | null>(null)
  const [userId, setUserId] = useState<string | null>(initialUserId ?? null)
  const [anonymousId, setAnonymousId] = useState<string | null>(null)
  const [ready, setReady] = useState(false)

  const traitsRef = useRef<Record<string, unknown>>(initialTraits ?? {})

  // Inicializa anonymous ID via AsyncStorage (assíncrono no RN)
  useEffect(() => {
    getOrCreateAnonymousId().then(setAnonymousId)
  }, [])

  const effectiveUserId = userId || anonymousId

  const loadFlags = useCallback(async () => {
    if (!effectiveUserId) return
    const qs = `?project_id=${encodeURIComponent(projectKey)}&user_id=${encodeURIComponent(effectiveUserId)}`
    const data = await fetchGet(apiUrl, projectKey, `/flags${qs}`, timeout) as { flags?: FlagsMap } | null
    if (data?.flags) {
      setFlags(data.flags)
      if (debug) console.log('[Pulse] flags loaded', data.flags)
    }
  }, [apiUrl, projectKey, effectiveUserId, timeout, debug])

  const loadVariants = useCallback(async () => {
    if (!effectiveUserId) return
    const qs = `?user_id=${encodeURIComponent(effectiveUserId)}&project_id=${encodeURIComponent(projectKey)}`
    const data = await fetchGet(apiUrl, projectKey, `/flags/variants${qs}`, timeout) as { variants?: VariantsMap } | null
    if (data?.variants) {
      setVariants(data.variants)
      if (debug) console.log('[Pulse] variants loaded', data.variants)
    }
  }, [apiUrl, projectKey, effectiveUserId, timeout, debug])

  // Carrega quando anonymousId estiver pronto
  useEffect(() => {
    if (!anonymousId) return
    Promise.all([loadFlags(), loadVariants()]).then(() => setReady(true))
  }, [anonymousId, loadFlags, loadVariants])

  const identify = useCallback(
    (newUserId: string, newTraits?: Record<string, unknown>) => {
      setUserId(newUserId)
      traitsRef.current = { ...traitsRef.current, ...(newTraits ?? {}) }
      AsyncStorage.setItem('_pulse_user_id', newUserId).catch(() => {})
      if (debug) console.log('[Pulse] identify', newUserId, newTraits)
      fetchPost(
        apiUrl,
        projectKey,
        '/api/v1/events/identify',
        {
          user_id: newUserId,
          anonymous_id: anonymousId,
          traits: traitsRef.current,
          project_key: projectKey,
        },
        timeout,
      )
      Promise.all([loadFlags(), loadVariants()])
    },
    [apiUrl, projectKey, anonymousId, timeout, debug, loadFlags, loadVariants],
  )

  const track = useCallback(
    (eventName: string, properties?: Record<string, unknown>) => {
      if (!eventName || !effectiveUserId) return
      if (debug) console.log('[Pulse] track', eventName, properties)
      fetchPost(
        apiUrl,
        projectKey,
        '/api/v1/events/track',
        {
          event_name: eventName,
          user_id: effectiveUserId,
          project_key: projectKey,
          platform: 'react-native',
          timestamp: new Date().toISOString(),
          properties: { ...traitsRef.current, ...(properties ?? {}) },
        },
        timeout,
      )
    },
    [apiUrl, projectKey, effectiveUserId, timeout, debug],
  )

  const reloadFlags = useCallback(async () => {
    setFlags(null)
    setVariants(null)
    await Promise.all([loadFlags(), loadVariants()])
  }, [loadFlags, loadVariants])

  return (
    <PulseContext.Provider
      value={{
        flags,
        variants,
        userId,
        anonymousId,
        ready,
        identify,
        track,
        reloadFlags,
      }}
    >
      {children}
    </PulseContext.Provider>
  )
}

// ─── Hooks públicos ────────────────────────────────────────────────────────

function usePulse(): PulseContextValue {
  const ctx = useContext(PulseContext)
  if (!ctx) throw new Error('usePulse must be used inside <PulseProvider>')
  return ctx
}

/**
 * Retorna se uma feature flag está habilitada para o usuário atual.
 *
 * @example
 * const newOnboarding = useFlag('new-onboarding')
 */
export function useFlag(flagKey: string, defaultValue = false): boolean {
  const { flags } = usePulse()
  if (flags === null) return defaultValue
  return flags[flagKey] === true
}

/**
 * Retorna função para registrar eventos de comportamento.
 *
 * @example
 * const track = useTrack()
 * track('screen_viewed', { screen: 'Home' })
 */
export function useTrack(): (eventName: string, properties?: Record<string, unknown>) => void {
  return usePulse().track
}

/**
 * Retorna função para identificar o usuário autenticado.
 *
 * @example
 * const identify = useIdentify()
 * identify('user_123', { name: 'João', plan: 'growth' })
 */
export function useIdentify(): (userId: string, traits?: Record<string, unknown>) => void {
  return usePulse().identify
}

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
  const { variants, track } = usePulse()
  const variant = variants?.[flagKey] ?? defaultVariant

  const trackExposure = useCallback(() => {
    track('$experiment_started', { experiment_name: flagKey, variant })
  }, [track, flagKey, variant])

  return { variant, trackExposure }
}

/**
 * Retorna true quando o SDK terminou de carregar flags do servidor.
 */
export function usePulseReady(): boolean {
  return usePulse().ready
}

/**
 * Retorna função que recarrega flags do servidor.
 */
export function useReloadFlags(): () => Promise<void> {
  return usePulse().reloadFlags
}

export { usePulse }
