/**
 * PulseContext — Context React para prover o PulseClient para hooks filhos.
 */

import AsyncStorage from '@react-native-async-storage/async-storage'
import React, { createContext, useContext, useEffect, useRef, useState } from 'react'
import { pulseClient, PulseClient } from './client'
import type { FlagsMap, PulseConfig, VariantsMap } from './types'

// ─── Tipos do contexto ────────────────────────────────────────────────────

export interface PulseContextValue {
  client: PulseClient
  flags: FlagsMap | null
  variants: VariantsMap | null
  userId: string | null
  anonymousId: string | null
  ready: boolean
  reloadFlags: () => Promise<void>
}

export const PulseContext = createContext<PulseContextValue | null>(null)

// ─── PulseProvider ────────────────────────────────────────────────────────

export interface PulseProviderProps {
  config: PulseConfig
  /** userId já autenticado, se disponível no momento da montagem */
  userId?: string
  /** Traits iniciais do usuário */
  traits?: Record<string, unknown>
  children: React.ReactNode
}

export function PulseProvider({
  config,
  userId: initialUserId,
  traits: initialTraits,
  children,
}: PulseProviderProps) {
  const [flags, setFlags] = useState<FlagsMap | null>(null)
  const [variants, setVariants] = useState<VariantsMap | null>(null)
  const [anonymousId, setAnonymousId] = useState<string | null>(null)
  const [userId, setUserId] = useState<string | null>(initialUserId ?? null)
  const [ready, setReady] = useState(false)
  const initializedRef = useRef(false)

  useEffect(() => {
    if (initializedRef.current) return
    initializedRef.current = true

    pulseClient.init(config, AsyncStorage, initialUserId, initialTraits)

    // Carrega anonymous ID do AsyncStorage e então busca flags
    pulseClient.loadAnonymousId().then(async (anonId) => {
      setAnonymousId(anonId)
      if (initialUserId) setUserId(initialUserId)
      const [flagsData, variantsData] = await Promise.all([
        pulseClient.getFlags(),
        pulseClient.getVariants(),
      ])
      setFlags(flagsData)
      setVariants(variantsData)
      setReady(true)
    })
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  const reloadFlags = async () => {
    pulseClient.invalidateCache()
    setFlags(null)
    setVariants(null)
    const [flagsData, variantsData] = await Promise.all([
      pulseClient.getFlags(),
      pulseClient.getVariants(),
    ])
    setFlags(flagsData)
    setVariants(variantsData)
  }

  return (
    <PulseContext.Provider
      value={{
        client: pulseClient,
        flags,
        variants,
        userId,
        anonymousId,
        ready,
        reloadFlags,
      }}
    >
      {children}
    </PulseContext.Provider>
  )
}

// ─── usePulseContext ──────────────────────────────────────────────────────

export function usePulseContext(): PulseContextValue {
  const ctx = useContext(PulseContext)
  if (!ctx) throw new Error('[Pulse] usePulse* deve ser usado dentro de <PulseProvider>')
  return ctx
}
