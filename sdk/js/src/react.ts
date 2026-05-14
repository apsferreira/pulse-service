/**
 * React hooks para o Pulse SDK.
 *
 * @example
 * ```tsx
 * import { usePulseFlag, usePulseVariant } from '@iit/pulse-sdk/react'
 * import { PulseClient } from '@iit/pulse-sdk'
 *
 * const pulse = new PulseClient({ baseURL: 'https://seu-pulse.exemplo.com' })
 *
 * function MyComponent() {
 *   const newDashboard = usePulseFlag(pulse, 'new-dashboard')
 *   const variant = usePulseVariant(pulse, 'pricing-page', 'control')
 *
 *   if (newDashboard) return <NewDashboard />
 *   return variant === 'challenger' ? <PricingB /> : <PricingA />
 * }
 * ```
 *
 * React >= 16.8 como peer dependency.
 */

import { useState, useEffect } from 'react'
import type { PulseClient } from './index'

/**
 * usePulseFlag — retorna true/false para uma feature flag.
 * Retorna false enquanto carrega (fail-open: nunca bloqueia render).
 * Recarrega quando flagKey ou userId mudam.
 */
export function usePulseFlag(
  client: PulseClient,
  flagKey: string,
  defaultValue = false,
  userId?: string
): boolean {
  const [enabled, setEnabled] = useState<boolean>(defaultValue)

  useEffect(() => {
    let cancelled = false
    client.isEnabled(flagKey, defaultValue).then((val) => {
      if (!cancelled) setEnabled(val)
    })
    return () => { cancelled = true }
  }, [flagKey, defaultValue, userId]) // eslint-disable-line react-hooks/exhaustive-deps

  return enabled
}

/**
 * usePulseVariant — retorna a variante A/B para uma flag.
 * Retorna defaultValue enquanto carrega.
 */
export function usePulseVariant(
  client: PulseClient,
  flagKey: string,
  defaultValue: string,
  userId?: string
): string {
  const [variant, setVariant] = useState<string>(defaultValue)

  useEffect(() => {
    let cancelled = false
    client.isEnabled(flagKey, false).then((enabled) => {
      if (!cancelled) setVariant(enabled ? flagKey : defaultValue)
    })
    return () => { cancelled = true }
  }, [flagKey, defaultValue, userId]) // eslint-disable-line react-hooks/exhaustive-deps

  return variant
}
