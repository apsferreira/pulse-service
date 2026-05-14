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

export interface RawEvent {
  id: string
  product: string
  event_type: string
  properties: Record<string, unknown>
  ip_address: string
  user_agent: string
  environment: string
  created_at: string
}

export interface EventsListResponse {
  items: RawEvent[]
  total: number
  limit: number
  offset: number
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
  list: (params?: Record<string, string | number>) =>
    api.get<EventsListResponse>('/events', { params }),
}

// ── Session Replay types ──────────────────────────────────────────────────────

export interface ReplaySession {
  id: string
  session_id: string
  project_id: string
  user_id: string
  started_at: string
  ended_at: string
  duration_ms: number
  event_count: number
  page_url: string
  user_agent: string
  screen_width: number
  screen_height: number
}

export type ReplayEventType = 'dom_mutation' | 'click' | 'scroll' | 'viewport' | 'custom'

export interface ReplayEvent {
  id: string
  type: ReplayEventType
  timestamp_ms: number
  data: Record<string, unknown>
}

export interface ReplaySessionsResponse {
  sessions: ReplaySession[]
  total: number
}

// ── Replay API ────────────────────────────────────────────────────────────────

export const replayApi = {
  sessions: (params?: { url?: string; user_id?: string; page?: number; per_page?: number }) =>
    api.get<ReplaySessionsResponse>('/replay/sessions', { params }),

  events: (sessionId: string) =>
    api.get<ReplayEvent[]>(`/replay/sessions/${sessionId}/events`),
}

// ── Approval Workflow types ───────────────────────────────────────────────────

export type ApprovalStatus = 'pending' | 'approved' | 'rejected'

export interface Approval {
  id: string
  flag_key: string
  requested_by: string
  status: ApprovalStatus
  changes: Record<string, unknown>
  message?: string
  comment?: string
  created_at: string
  reviewed_at?: string
}

export interface CreateApprovalPayload {
  changes: Record<string, unknown>
  message: string
}

export interface ReviewApprovalPayload {
  action: 'approve' | 'reject'
  comment: string
}

// ── Approvals API ─────────────────────────────────────────────────────────────

export const approvalsApi = {
  listByFlag: (flagKey: string) =>
    api.get<Approval[]>(`/flags/${flagKey}/approvals`),

  create: (flagKey: string, payload: CreateApprovalPayload) =>
    api.post<Approval>(`/flags/${flagKey}/approvals`, payload),

  review: (approvalId: string, payload: ReviewApprovalPayload) =>
    api.post<Approval>(`/approvals/${approvalId}/review`, payload),

  listAll: (projectId?: string | null) =>
    projectId
      ? api.get<Approval[]>('/approvals/pending', { params: { project_id: projectId } })
      : Promise.resolve({ data: [] as Approval[], status: 200, statusText: 'OK', headers: {}, config: {} as never }),
}
