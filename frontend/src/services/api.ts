import axios from 'axios'

const API_BASE_URL = (import.meta.env?.VITE_API_URL as string) ?? ''
const SERVICE_TOKEN = (import.meta.env?.VITE_SERVICE_TOKEN as string) ?? ''

export const api = axios.create({
  baseURL: `${API_BASE_URL}/api/v1`,
  timeout: 15000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Attach X-Service-Token on every request
api.interceptors.request.use(
  (config) => {
    if (SERVICE_TOKEN) {
      config.headers['X-Service-Token'] = SERVICE_TOKEN
    }
    if ((import.meta.env?.DEV as boolean) || (import.meta.env?.MODE as string) === 'development') {
      console.log(`[Pulse API] ${config.method?.toUpperCase()} ${config.url}`)
    }
    return config
  },
  (error) => {
    console.error('[Pulse API] Request error:', error)
    return Promise.reject(error)
  }
)

api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response) {
      console.error('[Pulse API] Response error:', error.response.status, error.response.data)
    } else if (error.request) {
      console.error('[Pulse API] No response received')
    } else {
      console.error('[Pulse API] Error:', error.message)
    }
    return Promise.reject(error)
  }
)

// ── Types ────────────────────────────────────────────────────────────────────

export interface FlagTargeting {
  user_ids?: string[]
  tenant_ids?: string[]
}

// BKL-124: Variante para A/B testing — key único + peso percentual (soma deve ser 100)
export interface FlagVariant {
  key: string    // ex: "control", "challenger", "variant-b"
  weight: number // 0-100, soma de todas as variantes deve ser 100
}

export interface FeatureFlag {
  id: string
  key: string
  description: string
  enabled: boolean
  rollout_percentage: number // BKL-123: 0-100
  targeting: FlagTargeting   // BKL-123: user_ids / tenant_ids
  variants: FlagVariant[]    // BKL-124: [] = on/off normal; [{key,weight}] = A/B
  created_at: string
  updated_at: string
}

export interface CreateFlagPayload {
  key: string
  description: string
  enabled: boolean
  rollout_percentage?: number
  targeting?: FlagTargeting
  variants?: FlagVariant[]   // BKL-124
}

export interface UpdateFlagPayload {
  description?: string
  enabled?: boolean
  rollout_percentage?: number
  targeting?: FlagTargeting
  variants?: FlagVariant[]   // BKL-124
}

// ── Flag API calls ───────────────────────────────────────────────────────────

export const flagsApi = {
  list: () => api.get<FeatureFlag[]>('/flags'),

  getByKey: (key: string) => api.get<FeatureFlag>(`/flags/${key}`),

  create: (payload: CreateFlagPayload) => api.post<FeatureFlag>('/flags', payload),

  update: (key: string, payload: UpdateFlagPayload) =>
    api.put<FeatureFlag>(`/flags/${key}`, payload),

  remove: (key: string) => api.delete(`/flags/${key}`),
}

// ── Retention / Cohort types (BKL-095) ──────────────────────────────────────

export interface ActiveUsersCountItem {
  date: string
  count: number
}

export interface ActiveUsersResponse {
  product: string
  dau: ActiveUsersCountItem[]
  wau: ActiveUsersCountItem[]
  mau: ActiveUsersCountItem[]
}

export interface CohortBucket {
  offset: number
  active_users: number
  retention_pct: number
}

export interface CohortItem {
  cohort_label: string
  cohort_size: number
  buckets: CohortBucket[]
}

export interface CohortsResponse {
  product: string
  granularity: string
  cohorts: CohortItem[]
}

// ── Metrics API calls (BKL-095) ─────────────────────────────────────────────

export const metricsApi = {
  activeUsers: (product?: string, startDate?: string, endDate?: string) => {
    const params = new URLSearchParams()
    if (product) params.set('product', product)
    if (startDate) params.set('start_date', startDate)
    if (endDate) params.set('end_date', endDate)
    return api.get<ActiveUsersResponse>(`/metrics/active-users?${params}`)
  },

  cohorts: (product?: string, granularity: 'week' | 'month' = 'week') => {
    const params = new URLSearchParams()
    if (product) params.set('product', product)
    params.set('granularity', granularity)
    return api.get<CohortsResponse>(`/metrics/cohorts?${params}`)
  },
}

export default api
