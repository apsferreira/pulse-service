import { api } from './api'

// ── Types ─────────────────────────────────────────────────────────────────────

export type PlanSlug = 'free' | 'starter' | 'growth' | 'pro' | 'scale' | 'self-hosted'

export interface QuotaInfo {
  flags_used: number
  flags_limit: number
  flags_pct: number
  events_last_30d: number
  events_limit: number
  events_pct: number
  projects_used: number
  projects_limit: number
  projects_pct: number
}

export interface UpgradeRecommendation {
  current_plan: string
  recommended_plan: string
  should_upgrade: boolean
  triggers: string[]
  quota_usage: QuotaInfo
  call_to_action: string
  upgrade_url: string
}

// ── Resposta do checkout-service ───────────────────────────────────────────────

export interface PlanResponse {
  slug: string
  name: string
  price_monthly_cents: number
  currency: string
  features: string[]
  highlighted: boolean
}

// ── Slugs canônicos (sem preço — use para upgrade gates e URLs) ───────────────

export const PULSE_PLAN_SLUGS = {
  STARTER: 'pulse-starter',
  GROWTH: 'pulse-growth',
  SCALE: 'pulse-scale',
  SELF_HOSTED: 'pulse-self-hosted',
} as const

// ── Fetch de planos via checkout-service ──────────────────────────────────────

const CHECKOUT_BASE = 'https://checkout-api.institutoitinerante.com.br'

export async function fetchPulsePlans(): Promise<PlanResponse[]> {
  const res = await fetch(`${CHECKOUT_BASE}/api/v1/plans?product=pulse`)
  if (!res.ok) throw new Error('Falha ao buscar planos do checkout-service')
  const data = (await res.json()) as { plans: PlanResponse[] }
  return data.plans
}

export function buildUpgradeURL(planSlug: PlanSlug, projectId: string): string {
  return `${CHECKOUT_BASE}/checkout?plan=pulse-${planSlug}&project_id=${projectId}`
}

// ── API calls ─────────────────────────────────────────────────────────────────

export const billingApi = {
  getRecommendation: (projectId: string) =>
    api
      .get<UpgradeRecommendation>('/upgrade/recommend', { params: { project_id: projectId } })
      .then((r) => r.data),
}
