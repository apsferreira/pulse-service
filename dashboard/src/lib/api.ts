const BASE_URL = (import.meta.env?.VITE_PULSE_API_URL as string) || ''
const SERVICE_TOKEN = (import.meta.env?.VITE_SERVICE_TOKEN as string) || ''

function headers(): Record<string, string> {
  const h: Record<string, string> = { 'Content-Type': 'application/json' }
  if (SERVICE_TOKEN) {
    h['X-Service-Token'] = SERVICE_TOKEN
  }
  return h
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE_URL}${path}`, {
    ...init,
    headers: { ...headers(), ...init?.headers },
  })
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error((body as { error?: string }).error || `HTTP ${res.status}`)
  }
  if (res.status === 204) return undefined as unknown as T
  return res.json() as Promise<T>
}

// ── Types ──────────────────────────────────────────────────────────

export interface FeatureFlag {
  id: string
  key: string
  description: string
  enabled: boolean
  created_at: string
  updated_at: string
}

export interface CreateFlagPayload {
  key: string
  description: string
  enabled: boolean
}

export interface UpdateFlagPayload {
  description?: string
  enabled?: boolean
}

export interface EventSummaryItem {
  product: string
  event_type: string
  count: number
}

export interface EventsSummaryResponse {
  items: EventSummaryItem[]
  total_last_24h: number
}

export interface EventCountItem {
  event_name: string
  count: number
}

export interface EventCountResponse {
  items: EventCountItem[]
  total: number
}

export interface Event {
  id: string
  product: string
  event_type: string
  user_id?: string
  session_id?: string
  properties: Record<string, unknown>
  ip_address?: string
  user_agent?: string
  created_at: string
}

export interface EventListResponse {
  items: Event[]
  total: number
  limit: number
  offset: number
}

// ── Flag API ──────────────────────────────────────────────────────

export const flagsApi = {
  list: () => request<FeatureFlag[]>('/api/v1/flags'),

  create: (payload: CreateFlagPayload) =>
    request<FeatureFlag>('/api/v1/flags', {
      method: 'POST',
      body: JSON.stringify(payload),
    }),

  update: (key: string, payload: UpdateFlagPayload) =>
    request<FeatureFlag>(`/api/v1/flags/${encodeURIComponent(key)}`, {
      method: 'PUT',
      body: JSON.stringify(payload),
    }),

  remove: (key: string) =>
    request<void>(`/api/v1/flags/${encodeURIComponent(key)}`, {
      method: 'DELETE',
    }),
}

// ── Events API ────────────────────────────────────────────────────

export const eventsApi = {
  summary: () => request<EventsSummaryResponse>('/api/v1/events/summary'),

  count: (params?: { event_name?: string; start_date?: string; end_date?: string }) => {
    const q = new URLSearchParams()
    if (params?.event_name) q.set('event_name', params.event_name)
    if (params?.start_date) q.set('start_date', params.start_date)
    if (params?.end_date) q.set('end_date', params.end_date)
    const qs = q.toString()
    return request<EventCountResponse>(`/api/v1/events/count${qs ? `?${qs}` : ''}`)
  },

  list: (params?: { event_name?: string; user_id?: string; limit?: number; offset?: number }) => {
    const q = new URLSearchParams()
    if (params?.event_name) q.set('event_name', params.event_name)
    if (params?.user_id) q.set('user_id', params.user_id)
    if (params?.limit) q.set('limit', String(params.limit))
    if (params?.offset) q.set('offset', String(params.offset))
    const qs = q.toString()
    return request<EventListResponse>(`/api/v1/events${qs ? `?${qs}` : ''}`)
  },
}
