/**
 * @pulse/react — SDK React para Pulse
 * Feature Flags + A/B Testing + Event Tracking
 *
 * BKL-307: usePulseFlag, usePulseVariant e PulseProvider com serviceToken/projectId.
 * BKL-258: integra com fail-open do SDK JS — defaults são respeitados nos hooks.
 *
 * Uso (API original):
 *   import { PulseProvider, useFlag, useTrack, useExperiment } from '@pulse/react'
 *
 *   <PulseProvider projectKey="YOUR_KEY" apiUrl="https://pulse-api.institutoitinerante.com.br">
 *     <MyApp />
 *   </PulseProvider>
 *
 * Uso (BKL-307 — API com serviceToken/projectId):
 *   import { PulseProvider, usePulseFlag, usePulseVariant } from '@pulse/react'
 *
 *   <PulseProvider serviceToken="svc_xxx" projectId="proj_yyy"
 *                  apiUrl="https://pulse-api.institutoitinerante.com.br"
 *                  defaults={{ 'nova-feature': false }}
 *                  failOpen={true}>
 *     <MyApp />
 *   </PulseProvider>
 *
 *   function MyComponent() {
 *     const newFeature = usePulseFlag('nova-feature', false)
 *     const variant = usePulseVariant('pricing-test', 'control')
 *     return newFeature ? <NewFeature /> : <OldFeature />
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

/** BKL-258: mapa de defaults por flag — usados quando fail-open ativa */
export type DefaultsMap = Record<string, boolean>

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
  /**
   * Timeout de requests em ms (default: 2000).
   * BKL-258: 2s para não bloquear a aplicação.
   */
  timeout?: number
  /**
   * BKL-258: defaults por flag — usados quando fail-open ativa.
   * Ex: { 'nova-feature': false }
   */
  defaults?: DefaultsMap
  /**
   * BKL-258: fail-open ativo (default: true).
   * Retorna defaults em vez de lançar erro quando backend indisponível.
   */
  failOpen?: boolean
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
  /** BKL-258: defaults por flag configurados no Provider */
  defaults: DefaultsMap
  /** BKL-258: se fail-open está ativo */
  failOpen: boolean
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

interface FetchGetOptions {
  apiUrl: string
  projectKey: string
  path: string
  timeout: number
  /** BKL-307: service token para autenticação */
  serviceToken?: string
}

/**
 * BKL-258: fetchGet com timeout de 2s e fail-open (retorna null em vez de throw).
 * BKL-307: suporta X-Service-Token para autenticação server-side.
 */
async function fetchGet({
  apiUrl,
  projectKey,
  path,
  timeout,
  serviceToken,
}: FetchGetOptions): Promise<unknown> {
  const url = apiUrl.replace(/\/+$/, '') + path
  const controller = new AbortController()
  const timer = setTimeout(() => controller.abort(), timeout)
  try {
    const headers: Record<string, string> = {
      'X-Project-Key': projectKey,
      Accept: 'application/json',
    }
    if (serviceToken) {
      headers['X-Service-Token'] = serviceToken
    }
    const res = await fetch(url, {
      method: 'GET',
      headers,
      signal: controller.signal,
    })
    if (!res.ok) return null
    return await res.json()
  } catch {
    // BKL-258: fail-open — retorna null em vez de throw (defaults usados pelo caller)
    return null
  } finally {
    clearTimeout(timer)
  }
}

interface FetchPostOptions {
  apiUrl: string
  projectKey: string
  path: string
  body: unknown
  timeout: number
  serviceToken?: string
}

async function fetchPost(
  apiUrl: string,
  projectKey: string,
  path: string,
  body: unknown,
  timeout: number,
  serviceToken?: string,
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
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      'X-Project-Key': projectKey,
    }
    if (serviceToken) {
      headers['X-Service-Token'] = serviceToken
    }
    await fetch(url, {
      method: 'POST',
      headers,
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
  /**
   * Chave do projeto (API legada).
   * BKL-307: use `projectId` para a nova API.
   */
  projectKey?: string
  /**
   * BKL-307: ID do projeto — usado com `serviceToken`.
   * Equivale a `projectKey` na API legada.
   */
  projectId?: string
  /**
   * BKL-307: service token para autenticação server-side.
   * Passado como header `X-Service-Token` em requests.
   */
  serviceToken?: string
  /** URL base do pulse-service */
  apiUrl: string
  userId?: string
  traits?: Record<string, unknown>
  debug?: boolean
  /**
   * Timeout em ms para fetch de flags (default: 2000).
   * BKL-258: 2s para não bloquear a aplicação.
   */
  timeout?: number
  /**
   * BKL-258: defaults por flag — usados quando fail-open ativa.
   * Ex: { 'nova-feature': false }
   */
  defaults?: DefaultsMap
  /**
   * BKL-258: fail-open ativo (default: true).
   * Retorna defaults em vez de lançar erro quando backend indisponível.
   */
  failOpen?: boolean
  children: React.ReactNode
}

export function PulseProvider({
  projectKey: projectKeyProp,
  projectId,
  serviceToken,
  apiUrl,
  userId: initialUserId,
  traits: initialTraits,
  debug = false,
  timeout = 2000,
  defaults = {},
  failOpen = true,
  children,
}: PulseProviderProps) {
  // BKL-307: suporta tanto projectKey (legado) quanto projectId (novo)
  const projectKey = projectKeyProp ?? projectId ?? ''

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
    // BKL-258: fetchGet retorna null em caso de timeout/erro (fail-open no caller)
    const data = await fetchGet({
      apiUrl,
      projectKey,
      path: `/flags${qs}`,
      timeout,
      serviceToken,
    }) as { flags?: FlagsMap } | null
    if (data?.flags) {
      setFlags(data.flags)
      log(debug, 'flags loaded', data.flags)
    } else if (failOpen && Object.keys(defaults).length > 0) {
      // BKL-258: backend indisponível — aplica defaults como flags iniciais
      setFlags({ ...defaults })
      log(debug, 'flags fail-open: usando defaults', defaults)
    }
  }, [apiUrl, projectKey, serviceToken, userId, anonymousId, timeout, debug, failOpen, defaults])

  const loadVariants = useCallback(async () => {
    const uid = userId || anonymousId
    const qs = `?user_id=${encodeURIComponent(uid)}&project_id=${encodeURIComponent(projectKey)}`
    const data = await fetchGet({
      apiUrl,
      projectKey,
      path: `/flags/variants${qs}`,
      timeout,
      serviceToken,
    }) as { variants?: VariantsMap } | null
    if (data?.variants) {
      setVariants(data.variants)
      log(debug, 'variants loaded', data.variants)
    }
    // BKL-258: variantes não têm defaults — se falhar, mantém null (hooks usam defaultVariant)
  }, [apiUrl, projectKey, serviceToken, userId, anonymousId, timeout, debug])

  // Init: carrega flags e variantes; setReady=true mesmo em fail-open
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
        serviceToken,
      )
      // Re-carrega flags com contexto do usuário
      Promise.all([loadFlags(), loadVariants()])
    },
    [apiUrl, projectKey, serviceToken, anonymousId, timeout, debug, loadFlags, loadVariants],
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
        serviceToken,
      )
    },
    [apiUrl, projectKey, serviceToken, effectiveUserId, timeout, debug],
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
        defaults,
        failOpen,
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
 * BKL-258: usa defaults do Provider quando flags não carregadas (fail-open).
 *
 * @param flagKey - chave da flag (ex: 'new-checkout')
 * @param defaultValue - valor padrão; se omitido, usa o default do Provider
 *
 * @example
 * const newCheckout = useFlag('new-checkout')
 * if (newCheckout) showNewCheckout()
 */
export function useFlag(flagKey: string, defaultValue?: boolean): boolean {
  const { flags, defaults } = usePulse()
  // BKL-258: resolução de prioridade:
  //   1. flags carregadas do servidor
  //   2. default explícito do caller
  //   3. default configurado no Provider
  //   4. false
  if (flags === null) return defaultValue ?? defaults[flagKey] ?? false
  if (!(flagKey in flags)) return defaultValue ?? defaults[flagKey] ?? false
  return flags[flagKey] === true
}

// ─── usePulseFlag ──────────────────────────────────────────────────────────

/**
 * BKL-307: alias semântico de useFlag com assinatura explícita de defaultValue.
 * Retorna o valor da flag; usa defaultValue se SDK não inicializado ou fail-open.
 *
 * @param flagKey - chave da flag (ex: 'nova-feature')
 * @param defaultValue - valor padrão (default: false)
 *
 * @example
 * const novaFeature = usePulseFlag('nova-feature', false)
 * return novaFeature ? <NovaFeature /> : <FeatureAntiga />
 */
export function usePulseFlag(flagKey: string, defaultValue = false): boolean {
  return useFlag(flagKey, defaultValue)
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

// ─── usePulseVariant ───────────────────────────────────────────────────────

/**
 * BKL-307: retorna a variante de um experimento A/B.
 * Usa defaultVariant como fallback quando SDK não inicializado ou fail-open.
 *
 * Diferença de useExperiment: retorna apenas a string da variante (sem trackExposure).
 * Use useExperiment quando precisar registrar a exposição.
 *
 * @param flagKey - chave da flag com variantes (ex: 'pricing-test')
 * @param defaultVariant - variante padrão (default: 'control')
 *
 * @example
 * const variant = usePulseVariant('pricing-test', 'control')
 * return variant === 'challenger' ? <NewPricing /> : <OldPricing />
 */
export function usePulseVariant(flagKey: string, defaultVariant = 'control'): string {
  const { variants } = usePulse()
  return variants?.[flagKey] ?? defaultVariant
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
