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

export type Product = 'libri' | 'nitro' | 'brio' | 'global'

export interface FeatureFlag {
  id: string
  name: string
  description: string
  product: Product | null
  is_enabled: boolean
  rollout_pct: number
  created_at: string
  updated_at: string
}

export interface CreateFlagPayload {
  name: string
  description: string
  is_enabled: boolean
  product: Product | null
  rollout_pct: number
}

export interface UpdateFlagPayload {
  description?: string
  is_enabled?: boolean
  rollout_pct?: number
}

// ── Flag API calls ───────────────────────────────────────────────────────────

export const flagsApi = {
  list: () => api.get<FeatureFlag[]>('/flags'),

  create: (payload: CreateFlagPayload) => api.post<FeatureFlag>('/flags', payload),

  update: (id: string, payload: UpdateFlagPayload) =>
    api.put<FeatureFlag>(`/flags/${id}`, payload),

  remove: (id: string) => api.delete(`/flags/${id}`),

  evaluate: (product: Product, userId: string) =>
    api.get<Record<string, boolean>>('/flags/evaluate', {
      params: { product, user_id: userId },
    }),
}

export default api
