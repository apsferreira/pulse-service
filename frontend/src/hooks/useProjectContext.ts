// BKL-1076: Contexto de projeto selecionado — persiste em localStorage.
import { useState, useCallback } from 'react'

const STORAGE_KEY = 'pulse:selected_project_id'

export interface ProjectSelection {
  projectId: string | null
  setProjectId: (id: string | null) => void
  clearProject: () => void
}

export function useProjectContext(): ProjectSelection {
  const [projectId, setProjectIdState] = useState<string | null>(() => {
    if (typeof window === 'undefined') return null
    return localStorage.getItem(STORAGE_KEY)
  })

  const setProjectId = useCallback((id: string | null) => {
    setProjectIdState(id)
    if (id) {
      localStorage.setItem(STORAGE_KEY, id)
    } else {
      localStorage.removeItem(STORAGE_KEY)
    }
  }, [])

  const clearProject = useCallback(() => setProjectId(null), [setProjectId])

  return { projectId, setProjectId, clearProject }
}
