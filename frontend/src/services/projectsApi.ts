import { api } from './api'

// ── Types ─────────────────────────────────────────────────────────────────────

export interface Project {
  id: string
  name: string
  slug: string
  api_key: string
  plan: string
  mau_current?: number
  is_active: boolean
  sdk_connected_at: string | null
  created_at: string
  updated_at?: string
}

export interface Environment {
  id: string
  project_id: string
  name: string
  slug: string
  api_key: string
  created_at: string
}

export interface ProjectListResponse {
  projects: Project[]
  total: number
}

export interface EnvironmentListResponse {
  environments: Environment[]
  total: number
}

export interface CreateProjectPayload {
  name: string
  slug: string
  plan?: string
}

export interface UpdateProjectPayload {
  name?: string
  plan?: string
  is_active?: boolean
}

export interface CreateEnvironmentPayload {
  name: string
  slug: string
}

// ── API calls ─────────────────────────────────────────────────────────────────

export const projectsApi = {
  list: () =>
    api.get<ProjectListResponse>('/projects').then((r) => r.data),

  create: (payload: CreateProjectPayload) =>
    api.post<Project>('/projects', payload).then((r) => r.data),

  get: (slug: string) =>
    api.get<Project>(`/projects/${slug}`).then((r) => r.data),

  update: (slug: string, payload: UpdateProjectPayload) =>
    api.put<Project>(`/projects/${slug}`, payload).then((r) => r.data),

  regenerateKey: (slug: string) =>
    api.post<{ api_key: string }>(`/projects/${slug}/regenerate-key`).then((r) => r.data),

  listEnvironments: (slug: string) =>
    api.get<EnvironmentListResponse>(`/projects/${slug}/environments`).then((r) => r.data),

  createEnvironment: (slug: string, payload: CreateEnvironmentPayload) =>
    api.post<Environment>(`/projects/${slug}/environments`, payload).then((r) => r.data),

  regenerateEnvKey: (slug: string, envId: string) =>
    api
      .post<{ api_key: string }>(`/projects/${slug}/environments/${envId}/regenerate-key`)
      .then((r) => r.data),
}
