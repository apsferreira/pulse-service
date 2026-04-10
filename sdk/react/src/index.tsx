/**
 * @pulse/react — SDK React para Pulse
 * Feature Flags + A/B Testing + Event Tracking
 *
 * Uso:
 *   import { PulseProvider, useFlag, useTrack, useExperiment } from '@pulse/react'
 *
 *   function App() {
 *     return (
 *       <PulseProvider projectKey="YOUR_KEY" apiUrl="https://pulse-api.institutoitinerante.com.br">
 *         <MyApp />
 *       </PulseProvider>
 *     )
 *   }
 *
 *   function MyComponent() {
 *     const newCheckout = useFlag('new-checkout')
 *     const track = useTrack()
 *     const { variant, trackExposure } = useExperiment('pricing-test')
 *
 *     useEffect(() => { trackExposure() }, [trackExposure])
 *
 *     return (
 *       <button onClick={() => track('button_clicked', { name: 'signup' })}>
 *         {newCheckout ? 'New Checkout' : 'Old Checkout'}
 *       </button>
 *     )
 *   }
 */

import React, {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useRef,
  useState,
} from 'react'

// ─── Types ─────────────────────────────────────────────────────────────────

export interface PulseConfig {
  /** Chave do projeto no Pulse */
  projectKey: string
  /** URL base do pulse-service */
  apiUrl: string
  /** ID do usuário autenticado (opcional — pode ser definido via identify()) */
  userId?: string
  /** Traits do usuário */
  traits?: Record<string, unknown>
  /** Habilita logs de debug no console */
  debug?: boolean
  /** Timeout de requests em ms (default: 10000) */
  timeout?: number
}

export interface PulseUser {
  userId: string
  traits?: Record<string, unknown>
}

export interface ExperimentResult {
  /** Variante atribuída ao usuário (ex: 'control', 'challenger') */
  variant: string
  /** Registra que o usuário viu a variante ($experiment_started) */
  trackExposure: () => void
}

interface FlagsMap {
  [key: string]: boolean
}

interface VariantsMap {
  [key: string]: string
}

// ─── Context ───────────────────────────────────────────────────────────────

interface PulseContextValue {
  flags: FlagsMap | null
  variants: VariantsMap | null
  userId: string | null
  anonymousId: string
  ready: boolean
  identify: (userId: string, traits?: Record<string, unknown>) => void
  track: (eventName: string, properties?: Record<string, unknown>) => void
  reloadFlags: () => Promise<void>
}

const PulseContext = createContext<PulseContextValue | null>(null)

// ─── Internal helpers ──────────────────────────────────────────────────────

function uuid(): string {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0
    return (c === 'x' ? r : (r & 0x3) | 0x8).toString(16)
  })
}

function getOrCreateAnonymousId(): string {
  try {
    const stored = localStorage.getItem('_pulse_anon_id')
    if (stored) return stored
    const id = 'anon_' + uuid()
    localStorage.setItem('_pulse_anon_id', id)
    return id
  } catch {
    return 'anon_' + uuid()
  }
}

function log(debug: boolean, ...args: unknown[]) {
  if (debug && typeof console !== 'undefined') {
    console.log('[Pulse]', ...args)
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
  const payload = JSON.stringify(body)

  // Beacon API: fire-and-forget para eventos
  if (
    path.includes('/events') &&
    typeof navigator !== 'undefined' &&
    navigator.sendBeacon
  ) {
    try {
      const blob = new Blob([payload], { type: 'application/json' })
      navigator.sendBeacon(url, blob)
      return
    } catch {
      // fallback to fetch
    }
  }

  const controller = new AbortController()
  const timer = setTimeout(() => controller.abort(), timeout)
  try {
    await fetch(url, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-Project-Key': projectKey,
      },
      body: payload,
      signal: controller.signal,
    })
  } catch {
    // silently fail — events are best-effort
  } finally {
    clearTimeout(timer)
  }
}

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
  const [ready, setReady] = useState(false)

  const anonymousId = useRef(getOrCreateAnonymousId()).current
  const traitsRef = useRef<Record<string, unknown>>(initialTraits ?? {})

  const effectiveUserId = userId || anonymousId

  const loadFlags = useCallback(async () => {
    const uid = userId || anonymousId
    const qs = `?project_id=${encodeURIComponent(projectKey)}&user_id=${encodeURIComponent(uid)}`
    const data = await fetchGet(apiUrl, projectKey, `/flags${qs}`, timeout) as { flags?: FlagsMap } | null
    if (data?.flags) {
      setFlags(data.flags)
      log(debug, 'flags loaded', data.flags)
    }
  }, [apiUrl, projectKey, userId, anonymousId, timeout, debug])

  const loadVariants = useCallback(async () => {
    const uid = userId || anonymousId
    const qs = `?user_id=${encodeURIComponent(uid)}&project_id=${encodeURIComponent(projectKey)}`
    const data = await fetchGet(apiUrl, projectKey, `/flags/variants${qs}`, timeout) as { variants?: VariantsMap } | null
    if (data?.variants) {
      setVariants(data.variants)
      log(debug, 'variants loaded', data.variants)
    }
  }, [apiUrl, projectKey, userId, anonymousId, timeout, debug])

  // Init: carrega flags e variantes
  useEffect(() => {
    Promise.all([loadFlags(), loadVariants()]).then(() => setReady(true))
  }, [loadFlags, loadVariants])

  const identify = useCallback(
    (newUserId: string, newTraits?: Record<string, unknown>) => {
      setUserId(newUserId)
      traitsRef.current = { ...traitsRef.current, ...(newTraits ?? {}) }
      try {
        localStorage.setItem('_pulse_user_id', newUserId)
      } catch {
        // ignore
      }
      log(debug, 'identify', newUserId, newTraits)
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
      // Re-carrega flags com contexto do usuário
      Promise.all([loadFlags(), loadVariants()])
    },
    [apiUrl, projectKey, anonymousId, timeout, debug, loadFlags, loadVariants],
  )

  const track = useCallback(
    (eventName: string, properties?: Record<string, unknown>) => {
      if (!eventName) return
      log(debug, 'track', eventName, properties)
      fetchPost(
        apiUrl,
        projectKey,
        '/api/v1/events/track',
        {
          event_name: eventName,
          user_id: effectiveUserId,
          project_key: projectKey,
          url: typeof window !== 'undefined' ? window.location.href : undefined,
          referrer: typeof document !== 'undefined' ? document.referrer || undefined : undefined,
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

// ─── usePulse ──────────────────────────────────────────────────────────────

function usePulse(): PulseContextValue {
  const ctx = useContext(PulseContext)
  if (!ctx) {
    throw new Error('usePulse must be used inside <PulseProvider>')
  }
  return ctx
}

// ─── useFlag ───────────────────────────────────────────────────────────────

/**
 * Retorna se uma feature flag está habilitada para o usuário atual.
 *
 * @param flagKey - chave da flag (ex: 'new-checkout')
 * @param defaultValue - valor padrão enquanto carrega (default: false)
 *
 * @example
 * const newCheckout = useFlag('new-checkout')
 * if (newCheckout) showNewCheckout()
 */
export function useFlag(flagKey: string, defaultValue = false): boolean {
  const { flags } = usePulse()
  if (flags === null) return defaultValue
  return flags[flagKey] === true
}

// ─── useTrack ──────────────────────────────────────────────────────────────

/**
 * Retorna função para registrar eventos de comportamento.
 *
 * @example
 * const track = useTrack()
 * track('button_clicked', { button: 'signup' })
 */
export function useTrack(): (eventName: string, properties?: Record<string, unknown>) => void {
  const { track } = usePulse()
  return track
}

// ─── useIdentify ───────────────────────────────────────────────────────────

/**
 * Retorna função para identificar o usuário autenticado.
 *
 * @example
 * const identify = useIdentify()
 * identify('user_123', { name: 'João', plan: 'growth' })
 */
export function useIdentify(): (userId: string, traits?: Record<string, unknown>) => void {
  const { identify } = usePulse()
  return identify
}

// ─── useExperiment ─────────────────────────────────────────────────────────

/**
 * Retorna a variante de um experimento A/B e função para registrar exposição.
 *
 * @param flagKey - chave da flag com variantes (ex: 'pricing-test')
 * @param defaultVariant - variante padrão se não há atribuição (default: 'control')
 *
 * @example
 * const { variant, trackExposure } = useExperiment('pricing-test')
 * useEffect(() => { trackExposure() }, [trackExposure])
 * return variant === 'challenger' ? <NewPricing /> : <OldPricing />
 */
export function useExperiment(
  flagKey: string,
  defaultVariant = 'control',
): ExperimentResult {
  const { variants, track } = usePulse()
  const variant = variants?.[flagKey] ?? defaultVariant

  const trackExposure = useCallback(() => {
    track('$experiment_started', {
      experiment_name: flagKey,
      variant,
    })
  }, [track, flagKey, variant])

  return { variant, trackExposure }
}

// ─── usePulseReady ─────────────────────────────────────────────────────────

/**
 * Retorna true quando o SDK terminou de carregar flags e variantes do servidor.
 * Útil para evitar flash de conteúdo durante o carregamento inicial.
 *
 * @example
 * const ready = usePulseReady()
 * if (!ready) return <Skeleton />
 */
export function usePulseReady(): boolean {
  const { ready } = usePulse()
  return ready
}

// ─── useReloadFlags ────────────────────────────────────────────────────────

/**
 * Retorna função que recarrega flags do servidor (útil após mudança de contexto).
 */
export function useReloadFlags(): () => Promise<void> {
  const { reloadFlags } = usePulse()
  return reloadFlags
}

// ─── Re-exports para conveniência ──────────────────────────────────────────

export { usePulse }
