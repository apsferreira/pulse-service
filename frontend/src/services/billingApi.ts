import { api } from './api'

// ── Types ─────────────────────────────────────────────────────────────────────

export type PlanSlug = 'free' | 'starter' | 'growth' | 'pro'

export interface UpgradeRecommendation {
  currentPlan: PlanSlug
  quotaUsage: number        // 0-100 (percentual de eventos usados no mês)
  shouldUpgrade: boolean
  recommendedPlan: PlanSlug | null
  upgradeURL: string | null
  callToAction: string | null
  activeFlagsCount: number
  projectsCount: number
  eventsUsed: number
  eventsLimit: number
}

export interface PlanDefinition {
  slug: PlanSlug
  label: string
  priceLabel: string
  priceMonthly: number | null  // null = gratuito
  eventsLabel: string
  flagsLabel: string
  projectsLabel: string
  features: string[]
  highlighted: boolean
}

// ── Planos estáticos (fonte de verdade no frontend enquanto não há endpoint) ──

export const PLAN_DEFINITIONS: PlanDefinition[] = [
  {
    slug: 'free',
    label: 'Free',
    priceLabel: 'R$ 0',
    priceMonthly: 0,
    eventsLabel: '10k eventos/mês',
    flagsLabel: '5 feature flags',
    projectsLabel: '1 projeto',
    features: ['SDK JS/Go/Python', 'Rollout percentual', 'Suporte via comunidade'],
    highlighted: false,
  },
  {
    slug: 'starter',
    label: 'Starter',
    priceLabel: 'R$ 399/mês',
    priceMonthly: 399,
    eventsLabel: '500k eventos/mês',
    flagsLabel: 'Flags ilimitadas',
    projectsLabel: '3 projetos',
    features: [
      'SDK JS/Go/Python',
      'Rollout percentual',
      'Targeting por user/tenant',
      'Suporte por email',
    ],
    highlighted: false,
  },
  {
    slug: 'growth',
    label: 'Growth',
    priceLabel: 'R$ 1.099/mês',
    priceMonthly: 1099,
    eventsLabel: '5M eventos/mês',
    flagsLabel: 'Flags ilimitadas',
    projectsLabel: '10 projetos',
    features: [
      'Tudo do Starter',
      'A/B Testing avançado',
      'Health Score por feature',
      'Data lake export',
      'Suporte prioritário',
    ],
    highlighted: true,
  },
  {
    slug: 'pro',
    label: 'Pro / Scale',
    priceLabel: 'R$ 2.399/mês',
    priceMonthly: 2399,
    eventsLabel: 'Ilimitado',
    flagsLabel: 'Flags ilimitadas',
    projectsLabel: 'Projetos ilimitados',
    features: [
      'Tudo do Growth',
      'SLA 99,9%',
      'Onboarding dedicado',
      'SSO / SAML',
      'Auditoria e compliance',
    ],
    highlighted: false,
  },
]

const CHECKOUT_BASE = 'https://checkout-api.institutoitinerante.com.br'

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
