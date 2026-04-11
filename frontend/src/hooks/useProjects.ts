// BKL-1076: Hook para carregar lista de projetos do Pulse (admin).
import { useState, useEffect } from 'react'
import { api } from '../services/api'

export interface Project {
  id: string
  name: string
  slug: string
  api_key: string
  plan: string
  is_active: boolean
  created_at: string
}

interface ProjectListResponse {
  projects: Project[]
  total: number
}

export function useProjects() {
  const [projects, setProjects] = useState<Project[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    api.get<ProjectListResponse>('/projects')
      .then((res) => {
        setProjects(res.data.projects ?? [])
      })
      .catch((err) => {
        setError(err?.response?.data?.error ?? err.message)
      })
      .finally(() => setLoading(false))
  }, [])

  return { projects, loading, error }
}
