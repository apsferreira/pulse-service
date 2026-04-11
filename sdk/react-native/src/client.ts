/**
 * PulseClient — singleton com estado de configuração e métodos de SDK.
 * Não usa dependências externas além do fetch nativo do React Native.
 */

import type { FlagsMap, PulseConfig, VariantsMap } from './types'

// ─── Cache de flags em memória (TTL 5 min) ────────────────────────────────

interface CacheEntry<T> {
  value: T
  expiresAt: number
}

const FLAG_CACHE_TTL_MS = 5 * 60 * 1000 // 5 minutos

const _memCache: {
  flags: CacheEntry<FlagsMap> | null
  variants: CacheEntry<VariantsMap> | null
} = {
  flags: null,
  variants: null,
}

function cacheGet<T>(entry: CacheEntry<T> | null): T | null {
  if (!entry) return null
  if (Date.now() > entry.expiresAt) return null
  return entry.value
}

function cacheSet<T>(value: T): CacheEntry<T> {
  return { value, expiresAt: Date.now() + FLAG_CACHE_TTL_MS }
}

// ─── Helpers internos ─────────────────────────────────────────────────────

function buildHeaders(config: PulseConfig): Record<string, string> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    'X-Project-Key': config.projectKey,
    Accept: 'application/json',
  }
  if (config.serviceToken) {
    headers['X-Service-Token'] = config.serviceToken
  }
  return headers
}

async function fetchGet(
  config: PulseConfig,
  path: string,
): Promise<unknown> {
  const url = config.apiUrl.replace(/\/+$/, '') + path
  const controller = new AbortController()
  const timer = setTimeout(() => controller.abort(), config.timeout ?? 10000)
  try {
    const res = await fetch(url, {
      method: 'GET',
      headers: buildHeaders(config),
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
  config: PulseConfig,
  path: string,
  body: unknown,
): Promise<void> {
  const url = config.apiUrl.replace(/\/+$/, '') + path
  const controller = new AbortController()
  const timer = setTimeout(() => controller.abort(), config.timeout ?? 10000)
  try {
    await fetch(url, {
      method: 'POST',
      headers: buildHeaders(config),
      body: JSON.stringify(body),
      signal: controller.signal,
    })
  } catch {
    // eventos são best-effort
  } finally {
    clearTimeout(timer)
  }
}

function uuid(): string {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0
    return (c === 'x' ? r : (r & 0x3) | 0x8).toString(16)
  })
}

// ─── PulseClient ──────────────────────────────────────────────────────────

export class PulseClient {
  private config: PulseConfig | null = null
  private _userId: string | null = null
  private _anonymousId: string | null = null
  private _traits: Record<string, unknown> = {}

  // AsyncStorage é injetado para evitar import direto (peer dep opcional em SSR)
  private _storage: {
    getItem(key: string): Promise<string | null>
    setItem(key: string, value: string): Promise<void>
  } | null = null

  init(
    config: PulseConfig,
    storage?: {
      getItem(key: string): Promise<string | null>
      setItem(key: string, value: string): Promise<void>
    },
    initialUserId?: string,
    initialTraits?: Record<string, unknown>,
  ): void {
    this.config = config
    this._storage = storage ?? null
    if (initialUserId) this._userId = initialUserId
    if (initialTraits) this._traits = initialTraits
    // Limpa cache ao reinicializar
    _memCache.flags = null
    _memCache.variants = null
  }

  private ensureInit(): PulseConfig {
    if (!this.config) throw new Error('[Pulse] Chame init() antes de usar o cliente.')
    return this.config
  }

  get userId(): string | null {
    return this._userId
  }

  get anonymousId(): string | null {
    return this._anonymousId
  }

  get effectiveUserId(): string | null {
    return this._userId ?? this._anonymousId
  }

  async loadAnonymousId(): Promise<string> {
    if (this._anonymousId) return this._anonymousId
    try {
      const stored = await this._storage?.getItem('_pulse_anon_id')
      if (stored) {
        this._anonymousId = stored
        return stored
      }
    } catch {
      // ignora erros de storage
    }
    const id = 'anon_' + uuid()
    this._anonymousId = id
    try {
      await this._storage?.setItem('_pulse_anon_id', id)
    } catch {
      // ignora
    }
    return id
  }

  async identify(userId: string, properties?: Record<string, unknown>): Promise<void> {
    const config = this.ensureInit()
    this._userId = userId
    this._traits = { ...this._traits, ...(properties ?? {}) }
    try {
      await this._storage?.setItem('_pulse_user_id', userId)
    } catch {
      // ignora
    }
    if (config.debug) console.log('[Pulse] identify', userId, properties)
    await fetchPost(config, '/api/v1/events/identify', {
      user_id: userId,
      anonymous_id: this._anonymousId,
      traits: this._traits,
      project_key: config.projectKey,
    })
    // Recarrega flags com novo userId
    _memCache.flags = null
    _memCache.variants = null
  }

  async track(eventType: string, properties?: Record<string, unknown>): Promise<void> {
    const config = this.ensureInit()
    const effectiveId = this.effectiveUserId
    if (!eventType || !effectiveId) return
    if (config.debug) console.log('[Pulse] track', eventType, properties)
    await fetchPost(config, '/api/v1/events/track', {
      event_name: eventType,
      user_id: effectiveId,
      project_key: config.projectKey,
      platform: 'react-native',
      timestamp: new Date().toISOString(),
      properties: { ...this._traits, ...(properties ?? {}) },
    })
  }

  async getFlag(key: string, defaultValue = false): Promise<boolean> {
    const config = this.ensureInit()
    const flags = await this.getFlags()
    if (flags === null) return defaultValue
    return flags[key] === true
  }

  async getFlags(): Promise<FlagsMap | null> {
    const config = this.ensureInit()
    const cached = cacheGet(_memCache.flags)
    if (cached) return cached

    const effectiveId = this.effectiveUserId
    if (!effectiveId) return null

    const qs = `?project_id=${encodeURIComponent(config.projectKey)}&user_id=${encodeURIComponent(effectiveId)}`
    const data = await fetchGet(config, `/flags${qs}`) as { flags?: FlagsMap } | null
    if (data?.flags) {
      _memCache.flags = cacheSet(data.flags)
      if (config.debug) console.log('[Pulse] flags loaded', data.flags)
      return data.flags
    }
    return null
  }

  async getVariants(): Promise<VariantsMap | null> {
    const config = this.ensureInit()
    const cached = cacheGet(_memCache.variants)
    if (cached) return cached

    const effectiveId = this.effectiveUserId
    if (!effectiveId) return null

    const qs = `?user_id=${encodeURIComponent(effectiveId)}&project_id=${encodeURIComponent(config.projectKey)}`
    const data = await fetchGet(config, `/flags/variants${qs}`) as { variants?: VariantsMap } | null
    if (data?.variants) {
      _memCache.variants = cacheSet(data.variants)
      if (config.debug) console.log('[Pulse] variants loaded', data.variants)
      return data.variants
    }
    return null
  }

  invalidateCache(): void {
    _memCache.flags = null
    _memCache.variants = null
  }
}

/** Instância singleton global do PulseClient. */
export const pulseClient = new PulseClient()

/** Atalhos funcionais para uso fora de hooks React. */
export const track = (
  eventType: string,
  properties?: Record<string, unknown>,
): Promise<void> => pulseClient.track(eventType, properties)

export const identify = (
  userId: string,
  properties?: Record<string, unknown>,
): Promise<void> => pulseClient.identify(userId, properties)
