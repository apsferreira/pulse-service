import api from './api'

// ── Monitor types ─────────────────────────────────────────────────────────────

export interface Monitor {
  id: string
  name: string
  url: string
  method: string
  interval_seconds: number
  timeout_seconds: number
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface MonitorCheck {
  id: string
  monitor_id: string
  status: 'up' | 'down' | 'timeout'
  status_code?: number
  response_time_ms?: number
  error_message?: string
  checked_at: string
}

export interface MonitorStatus {
  monitor: Monitor
  last_check?: MonitorCheck
  uptime_24h: number
  avg_response_ms: number
}

export interface CreateMonitorPayload {
  name: string
  url: string
  method?: string
  interval_seconds?: number
  timeout_seconds?: number
  is_active?: boolean
}

export interface UpdateMonitorPayload {
  name?: string
  url?: string
  method?: string
  interval_seconds?: number
  timeout_seconds?: number
  is_active?: boolean
}

// ── Events types ──────────────────────────────────────────────────────────────

export interface EventSummaryItem {
  product: string
  event_type: string
  count: number
}

export interface EventsSummaryResponse {
  items: EventSummaryItem[]
  total_last_24h: number
}

// ── Monitor API ───────────────────────────────────────────────────────────────

export const monitorsApi = {
  list: () => api.get<MonitorStatus[]>('/monitors'),

  create: (payload: CreateMonitorPayload) =>
    api.post<Monitor>('/monitors', payload),

  update: (id: string, payload: UpdateMonitorPayload) =>
    api.put<Monitor>(`/monitors/${id}`, payload),

  remove: (id: string) => api.delete(`/monitors/${id}`),

  history: (id: string) => api.get<MonitorCheck[]>(`/monitors/${id}/history`),
}

// ── Events API ────────────────────────────────────────────────────────────────

export const eventsApi = {
  summary: () => api.get<EventsSummaryResponse>('/events/summary'),
}
