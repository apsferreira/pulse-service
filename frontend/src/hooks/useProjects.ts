// BKL-1076: Hook para carregar lista de projetos do Pulse (admin).
import { useState, useEffect, useCallback } from 'react'
import { projectsApi } from '../services/projectsApi'
import type { Project } from '../services/projectsApi'

export type { Project }

export function useProjects() {
  const [projects, setProjects] = useState<Project[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetchProjects = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const data = await projectsApi.list()
      setProjects(data.projects ?? [])
    } catch (err) {
      const e = err as { response?: { data?: { error?: string } }; message?: string }
      setError(e?.response?.data?.error ?? e?.message ?? 'Erro ao carregar projetos')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void fetchProjects()
  }, [fetchProjects])

  return { projects, loading, error, refetch: fetchProjects }
}
