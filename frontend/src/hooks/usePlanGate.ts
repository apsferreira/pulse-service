import { useProject } from '../components/Layout'

// Ordem de precedência dos planos
const PLAN_RANK: Record<string, number> = {
  free:       0,
  starter:    1,
  growth:     2,
  pro:        3,
  enterprise: 4,
  internal:   4,
}

// canUse retorna true se o plano atual atinge o plano mínimo requerido.
export function canUsePlan(currentPlan: string, requiredPlan: string): boolean {
  const current = PLAN_RANK[currentPlan] ?? 0
  const required = PLAN_RANK[requiredPlan] ?? 0
  return current >= required
}

// Hook que retorna se o projeto atual pode usar um recurso de determinado plano mínimo.
export function usePlanGate(requiredPlan: string): boolean {
  const { projectPlan } = useProject()
  return canUsePlan(projectPlan, requiredPlan)
}
