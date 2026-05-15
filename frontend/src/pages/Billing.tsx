import { useState, useEffect, useCallback } from 'react'
import {
  CreditCard,
  CheckCircle,
  RefreshCw,
  TrendingUp,
  AlertTriangle,
  ExternalLink,
  Receipt,
  BarChart3,
  Flag,
  FolderOpen,
  Server,
  MessageCircle,
} from 'lucide-react'
import { useProject } from '../components/Layout'
import { useProjects } from '../hooks/useProjects'
import {
  billingApi,
  PLAN_DEFINITIONS,
  buildUpgradeURL,
} from '../services/billingApi'
import type { UpgradeRecommendation, PlanSlug } from '../services/billingApi'

// ── Tipos e constantes de billing cycle ───────────────────────────────────────

type BillingCycle = 'monthly' | 'yearly'

const PRICES: Record<BillingCycle, Record<Exclude<PlanSlug, 'free'>, number>> = {
  monthly: { starter: 399, growth: 1099, pro: 2399 },
  yearly: {
    starter: Math.round(383000 / 100 / 12),   // 319
    growth:  Math.round(1055000 / 100 / 12),   // 879
    pro:     Math.round(2303000 / 100 / 12),   // 1919
  },
} as const

// ── Helpers ───────────────────────────────────────────────────────────────────

const PLAN_RANK: Record<PlanSlug, number> = { free: 0, starter: 1, growth: 2, pro: 3 }

function planRank(slug: string): number {
  return PLAN_RANK[slug as PlanSlug] ?? 0
}

function quotaColor(pct: number): string {
  if (pct >= 90) return 'bg-red-500'
  if (pct >= 70) return 'bg-amber-400'
  return 'bg-violet-500'
}

function quotaTextColor(pct: number): string {
  if (pct >= 90) return 'text-red-400'
  if (pct >= 70) return 'text-amber-400'
  return 'text-violet-400'
}

// ── Seção A — Plano atual ─────────────────────────────────────────────────────

interface CurrentPlanSectionProps {
  recommendation: UpgradeRecommendation
  projectName: string
}

function CurrentPlanSection({ recommendation, projectName }: CurrentPlanSectionProps) {
  const currentPlan = recommendation.current_plan
  const quota = recommendation.quota_usage
  const usagePct = Math.min(Math.round(quota?.events_pct ?? 0), 100)
  const eventsUsed = quota?.events_last_30d ?? 0
  const eventsLimit = quota?.events_limit ?? 10000
  const activeFlagsCount = quota?.flags_used ?? 0
  const projectsCount = quota?.projects_used ?? 0

  const planDef = PLAN_DEFINITIONS.find((p) => p.slug === currentPlan)

  return (
    <section
      className="rounded-2xl border border-white/10 p-6 space-y-5"
      style={{ background: '#110f1a' }}
      aria-label="Plano atual"
    >
      <div className="flex items-center gap-3">
        <div
          className="w-9 h-9 rounded-lg flex items-center justify-center flex-shrink-0"
          style={{ background: '#8B5CF620' }}
        >
          <CreditCard className="w-4.5 h-4.5" style={{ color: '#8B5CF6' }} aria-hidden="true" />
        </div>
        <div>
          <h2 className="text-base font-semibold text-slate-100">Plano atual</h2>
          <p className="text-xs text-slate-400">{projectName}</p>
        </div>

        <span
          className="ml-auto inline-flex items-center px-3 py-1 rounded-full text-xs font-semibold"
          style={{ background: '#8B5CF620', color: '#a78bfa' }}
        >
          {planDef?.label ?? currentPlan}
        </span>
      </div>

      {/* Quota de eventos */}
      <div>
        <div className="flex items-center justify-between mb-1.5">
          <span className="text-xs font-medium text-slate-400">Eventos usados este mês</span>
          <span className={`text-xs font-semibold tabular-nums ${quotaTextColor(usagePct)}`}>
            {usagePct}%
          </span>
        </div>
        <div className="h-2 w-full rounded-full bg-white/10 overflow-hidden">
          <div
            className={`h-full rounded-full transition-all ${quotaColor(usagePct)}`}
            style={{ width: `${usagePct}%` }}
            role="progressbar"
            aria-valuenow={usagePct}
            aria-valuemin={0}
            aria-valuemax={100}
            aria-label={`${usagePct}% da quota de eventos utilizada`}
          />
        </div>
        <p className="mt-1 text-xs text-slate-500">
          {eventsUsed.toLocaleString('pt-BR')} /{' '}
          {eventsLimit > 0 ? eventsLimit.toLocaleString('pt-BR') : '∞'} eventos
        </p>
      </div>

      {/* KPIs */}
      <div className="grid grid-cols-2 gap-3">
        <div
          className="rounded-xl p-3 flex items-center gap-3"
          style={{ background: '#09080f' }}
        >
          <Flag className="w-4 h-4 text-violet-400 flex-shrink-0" aria-hidden="true" />
          <div>
            <p className="text-xs text-slate-500">Flags ativas</p>
            <p className="text-lg font-bold text-slate-100 tabular-nums">{activeFlagsCount}</p>
          </div>
        </div>
        <div
          className="rounded-xl p-3 flex items-center gap-3"
          style={{ background: '#09080f' }}
        >
          <FolderOpen className="w-4 h-4 text-violet-400 flex-shrink-0" aria-hidden="true" />
          <div>
            <p className="text-xs text-slate-500">Projetos</p>
            <p className="text-lg font-bold text-slate-100 tabular-nums">{projectsCount}</p>
          </div>
        </div>
      </div>
    </section>
  )
}

// ── Toggle mensal/anual ───────────────────────────────────────────────────────

interface BillingToggleProps {
  billing: BillingCycle
  onChange: (v: BillingCycle) => void
}

function BillingToggle({ billing, onChange }: BillingToggleProps) {
  return (
    <div
      className="inline-flex items-center rounded-xl border border-white/10 overflow-hidden"
      role="group"
      aria-label="Ciclo de cobrança"
      style={{ background: '#09080f' }}
    >
      <button
        onClick={() => onChange('monthly')}
        className={[
          'px-4 py-1.5 text-sm font-medium transition-colors',
          billing === 'monthly'
            ? 'text-white'
            : 'text-slate-400 hover:text-slate-200',
        ].join(' ')}
        style={billing === 'monthly' ? { background: '#7c3aed' } : {}}
        aria-pressed={billing === 'monthly'}
      >
        Mensal
      </button>
      <button
        onClick={() => onChange('yearly')}
        className={[
          'px-4 py-1.5 text-sm font-medium transition-colors flex items-center gap-1.5',
          billing === 'yearly'
            ? 'text-white'
            : 'text-slate-400 hover:text-slate-200',
        ].join(' ')}
        style={billing === 'yearly' ? { background: '#7c3aed' } : {}}
        aria-pressed={billing === 'yearly'}
      >
        Anual
        <span
          className="text-[10px] font-bold px-1.5 py-0.5 rounded-full"
          style={{ background: '#10b98120', color: '#34d399' }}
        >
          -20%
        </span>
      </button>
    </div>
  )
}

// ── Seção Self-Hosted ─────────────────────────────────────────────────────────

function SelfHostedCard() {
  const contactUrl =
    'https://wa.me/5571999999999?text=Olá%2C%20quero%20saber%20mais%20sobre%20o%20Pulse%20Self-Hosted'

  return (
    <div
      className="rounded-2xl border border-amber-700/40 p-6 flex flex-col gap-5"
      style={{ background: '#1a1510' }}
      aria-label="Plano Self-Hosted"
    >
      <div className="flex items-start gap-3">
        <div
          className="w-9 h-9 rounded-lg flex items-center justify-center flex-shrink-0"
          style={{ background: '#92400e20' }}
        >
          <Server className="w-4.5 h-4.5 text-amber-400" aria-hidden="true" />
        </div>
        <div className="flex-1">
          <div className="flex items-center gap-2 flex-wrap">
            <p className="text-sm font-semibold text-slate-100">Self-Hosted</p>
            <span
              className="text-[10px] font-bold px-2 py-0.5 rounded-full"
              style={{ background: '#92400e40', color: '#fbbf24' }}
            >
              Licença única
            </span>
          </div>
          <p className="text-2xl font-bold text-amber-300 mt-1">
            R$ 50.000
            <span className="text-xs font-normal text-slate-400"> /ano</span>
          </p>
        </div>
      </div>

      <p className="text-xs text-slate-400 leading-relaxed">
        Seus dados na sua infraestrutura. Deploy via Docker Compose ou K3s — sem
        dependência de cloud externa. Controle total sobre retenção e conformidade.
      </p>

      <ul className="space-y-1.5">
        {[
          'Projetos e eventos ilimitados',
          'Docker Compose ou K3s — sua escolha',
          'Suporte dedicado de onboarding',
          'SLA 99,9% garantido em contrato',
          'SSO / SAML incluído',
        ].map((f) => (
          <li key={f} className="flex items-center gap-2 text-xs text-slate-300">
            <CheckCircle
              className="w-3.5 h-3.5 flex-shrink-0 text-amber-400"
              aria-hidden="true"
            />
            {f}
          </li>
        ))}
      </ul>

      <a
        href={contactUrl}
        target="_blank"
        rel="noopener noreferrer"
        className="w-full flex items-center justify-center gap-2 py-2.5 rounded-xl text-sm font-semibold transition-colors"
        style={{ background: '#92400e', color: '#fef3c7' }}
        aria-label="Falar com o time sobre Self-Hosted"
      >
        <MessageCircle className="w-4 h-4" aria-hidden="true" />
        Falar com o time
        <ExternalLink className="w-3 h-3" aria-hidden="true" />
      </a>

      <p className="text-[11px] text-slate-500 text-center leading-relaxed">
        Após o pagamento, você recebe uma{' '}
        <code className="text-amber-400/80 font-mono">LICENSE_KEY</code> por email
        para ativar sua instância.
      </p>
    </div>
  )
}

// ── Seção B — Cards de planos ─────────────────────────────────────────────────

interface PlanCardsProps {
  currentPlan: string
  projectId: string
  billing: BillingCycle
  onBillingChange: (v: BillingCycle) => void
}

function PlanCards({ currentPlan, projectId, billing, onBillingChange }: PlanCardsProps) {
  const currentRank = planRank(currentPlan)

  const resolvedPrice = (slug: PlanSlug): string => {
    if (slug === 'free') return 'R$ 0'
    const price = PRICES[billing][slug]
    return `R$ ${price.toLocaleString('pt-BR')}`
  }

  const upgradeUrl = (slug: PlanSlug): string => {
    const base = buildUpgradeURL(slug, projectId)
    return billing === 'yearly' ? `${base.replace(`pulse-${slug}`, `pulse-${slug}-annual`)}&billing=yearly` : base
  }

  return (
    <section aria-label="Planos disponíveis">
      <div className="flex items-center justify-between gap-3 mb-5 flex-wrap">
        <div className="flex items-center gap-2">
          <TrendingUp className="w-4 h-4 text-violet-400" aria-hidden="true" />
          <h2 className="text-base font-semibold text-slate-100">Planos</h2>
        </div>
        <BillingToggle billing={billing} onChange={onBillingChange} />
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        {PLAN_DEFINITIONS.map((plan) => {
          const isCurrent = plan.slug === currentPlan
          const isDowngrade = planRank(plan.slug) < currentRank
          const isUpgrade = planRank(plan.slug) > currentRank

          return (
            <div
              key={plan.slug}
              className={[
                'relative rounded-2xl border p-5 flex flex-col gap-4 transition-all',
                plan.highlighted && !isCurrent
                  ? 'border-violet-500/60 shadow-lg shadow-violet-900/20'
                  : 'border-white/10',
              ].join(' ')}
              style={{ background: '#110f1a' }}
            >
              {plan.highlighted && (
                <div
                  className="absolute -top-2.5 left-1/2 -translate-x-1/2 px-3 py-0.5 rounded-full text-xs font-semibold"
                  style={{ background: '#8B5CF6', color: '#fff' }}
                >
                  Mais popular
                </div>
              )}

              {isCurrent && (
                <div
                  className="absolute -top-2.5 left-1/2 -translate-x-1/2 px-3 py-0.5 rounded-full text-xs font-semibold"
                  style={{ background: '#1e293b', color: '#94a3b8', border: '1px solid #334155' }}
                >
                  Plano atual
                </div>
              )}

              {/* Header */}
              <div>
                <p className="text-sm font-semibold text-slate-100">{plan.label}</p>
                <p className="text-2xl font-bold text-slate-100 mt-1">
                  {resolvedPrice(plan.slug)}
                  {plan.priceMonthly !== null && plan.priceMonthly > 0 && (
                    <span className="text-xs font-normal text-slate-400"> /mês</span>
                  )}
                </p>
                {plan.slug !== 'free' && billing === 'yearly' && (
                  <p className="text-[11px] text-emerald-400 mt-0.5">
                    Cobrado anualmente — economize 20%
                  </p>
                )}
              </div>

              {/* Limites */}
              <ul className="space-y-1.5 flex-1">
                <li className="flex items-center gap-2 text-xs text-slate-300">
                  <BarChart3 className="w-3.5 h-3.5 text-violet-400 flex-shrink-0" aria-hidden="true" />
                  {plan.eventsLabel}
                </li>
                <li className="flex items-center gap-2 text-xs text-slate-300">
                  <Flag className="w-3.5 h-3.5 text-violet-400 flex-shrink-0" aria-hidden="true" />
                  {plan.flagsLabel}
                </li>
                <li className="flex items-center gap-2 text-xs text-slate-300">
                  <FolderOpen className="w-3.5 h-3.5 text-violet-400 flex-shrink-0" aria-hidden="true" />
                  {plan.projectsLabel}
                </li>

                <li className="pt-2 border-t border-white/5">
                  <ul className="space-y-1">
                    {plan.features.map((f) => (
                      <li key={f} className="flex items-start gap-1.5 text-xs text-slate-400">
                        <CheckCircle
                          className="w-3.5 h-3.5 text-emerald-500 flex-shrink-0 mt-px"
                          aria-hidden="true"
                        />
                        {f}
                      </li>
                    ))}
                  </ul>
                </li>
              </ul>

              {/* CTA */}
              {isCurrent ? (
                <div
                  className="w-full text-center py-2 rounded-lg text-xs font-medium text-slate-400"
                  style={{ background: '#09080f' }}
                >
                  Plano atual
                </div>
              ) : isUpgrade ? (
                <a
                  href={upgradeUrl(plan.slug)}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="w-full flex items-center justify-center gap-1.5 py-2 rounded-lg text-xs font-semibold transition-colors"
                  style={{ background: '#8B5CF6', color: '#fff' }}
                  aria-label={`Fazer upgrade para o plano ${plan.label}`}
                >
                  Fazer upgrade
                  <ExternalLink className="w-3 h-3" aria-hidden="true" />
                </a>
              ) : isDowngrade ? (
                <div
                  className="w-full text-center py-2 rounded-lg text-xs font-medium text-slate-500"
                  style={{ background: '#09080f' }}
                >
                  Plano inferior
                </div>
              ) : null}
            </div>
          )
        })}
      </div>

      {/* Self-Hosted */}
      <SelfHostedCard />
    </section>
  )
}

// ── Seção C — Histórico de cobranças ─────────────────────────────────────────

function BillingHistory() {
  return (
    <section aria-label="Histórico de cobranças">
      <div className="flex items-center gap-2 mb-4">
        <Receipt className="w-4 h-4 text-violet-400" aria-hidden="true" />
        <h2 className="text-base font-semibold text-slate-100">Histórico de cobranças</h2>
      </div>

      <div
        className="rounded-2xl border border-white/10 overflow-hidden"
        style={{ background: '#110f1a' }}
      >
        <table className="w-full text-sm" role="table">
          <thead>
            <tr className="border-b border-white/10">
              <th
                scope="col"
                className="px-4 py-3 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider"
              >
                Data
              </th>
              <th
                scope="col"
                className="px-4 py-3 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider"
              >
                Plano
              </th>
              <th
                scope="col"
                className="px-4 py-3 text-right text-xs font-semibold text-slate-400 uppercase tracking-wider"
              >
                Valor
              </th>
              <th
                scope="col"
                className="px-4 py-3 text-right text-xs font-semibold text-slate-400 uppercase tracking-wider"
              >
                Status
              </th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td colSpan={4} className="px-4 py-10 text-center">
                <div className="flex flex-col items-center gap-2">
                  <Receipt
                    className="w-8 h-8 text-slate-700"
                    aria-hidden="true"
                  />
                  <p className="text-sm text-slate-500">Nenhuma cobrança encontrada</p>
                  <p className="text-xs text-slate-600">
                    As cobranças aparecerão aqui após o primeiro ciclo de faturamento.
                  </p>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  )
}

// ── Skeleton de carregamento ──────────────────────────────────────────────────

function BillingSkeleton() {
  return (
    <div className="space-y-8 animate-pulse">
      <div className="rounded-2xl border border-white/10 p-6 h-48" style={{ background: '#110f1a' }} />
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {[1, 2, 3, 4].map((i) => (
          <div
            key={i}
            className="rounded-2xl border border-white/10 p-5 h-64"
            style={{ background: '#110f1a' }}
          />
        ))}
      </div>
      <div className="rounded-2xl border border-white/10 p-6 h-32" style={{ background: '#110f1a' }} />
    </div>
  )
}

// ── Estado: sem projeto selecionado ──────────────────────────────────────────

function NoProjectSelected() {
  return (
    <div
      className="rounded-2xl border border-white/10 p-12 text-center"
      style={{ background: '#110f1a' }}
    >
      <FolderOpen className="w-10 h-10 text-slate-700 mx-auto mb-3" aria-hidden="true" />
      <p className="text-sm font-medium text-slate-400">Nenhum projeto selecionado</p>
      <p className="text-xs text-slate-600 mt-1">
        Selecione um projeto no seletor do topo para ver o faturamento.
      </p>
    </div>
  )
}

// ── Página principal ──────────────────────────────────────────────────────────

export default function Billing() {
  const { projectId } = useProject()
  const { projects } = useProjects()
  const [recommendation, setRecommendation] = useState<UpgradeRecommendation | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [billing, setBilling] = useState<BillingCycle>('monthly')

  const selectedProject = projects.find((p) => p.id === projectId)

  const fetchRecommendation = useCallback(async (id: string) => {
    setLoading(true)
    setError(null)
    try {
      const data = await billingApi.getRecommendation(id)
      setRecommendation(data)
    } catch (err) {
      const e = err as { response?: { data?: { error?: string } }; message?: string }
      // Quando a API ainda não existe, montar dados mínimos a partir do projeto local
      if ((err as { response?: { status?: number } }).response?.status === 404 || !e.response) {
        const fallback: UpgradeRecommendation = {
          current_plan: selectedProject?.plan ?? 'free',
          recommended_plan: selectedProject?.plan ?? 'free',
          should_upgrade: false,
          triggers: [],
          quota_usage: {
            flags_used: 0, flags_limit: 5, flags_pct: 0,
            events_last_30d: 0, events_limit: 10000, events_pct: 0,
            projects_used: projects.length, projects_limit: 1, projects_pct: 0,
          },
          call_to_action: '',
          upgrade_url: '',
        }
        setRecommendation(fallback)
      } else {
        setError(e?.response?.data?.error ?? e?.message ?? 'Falha ao carregar dados de faturamento')
      }
    } finally {
      setLoading(false)
    }
  }, [selectedProject, projects.length])

  useEffect(() => {
    if (projectId) {
      void fetchRecommendation(projectId)
    } else {
      setRecommendation(null)
    }
  }, [projectId, fetchRecommendation])

  return (
    <div style={{ background: '#09080f', minHeight: '100%' }} className="space-y-8 pb-8">
      {/* Cabeçalho */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div
            className="w-9 h-9 rounded-lg flex items-center justify-center"
            style={{ background: '#8B5CF620' }}
          >
            <CreditCard className="w-4.5 h-4.5" style={{ color: '#8B5CF6' }} aria-hidden="true" />
          </div>
          <div>
            <h1 className="text-xl font-bold text-slate-100">Faturamento</h1>
            <p className="text-sm text-slate-400 mt-0.5">
              Gerencie seu plano e acompanhe o consumo
            </p>
          </div>
        </div>

        {projectId && (
          <button
            onClick={() => void fetchRecommendation(projectId)}
            disabled={loading}
            className="p-2 rounded-lg text-slate-400 hover:bg-white/5 transition-colors disabled:opacity-50"
            aria-label="Recarregar dados de faturamento"
          >
            <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
          </button>
        )}
      </div>

      {/* Conteúdo */}
      {!projectId ? (
        <NoProjectSelected />
      ) : loading ? (
        <BillingSkeleton />
      ) : error ? (
        <div
          className="rounded-2xl border border-red-900/40 p-6 text-center"
          style={{ background: '#110f1a' }}
          role="alert"
        >
          <AlertTriangle className="w-8 h-8 text-red-500 mx-auto mb-2" aria-hidden="true" />
          <p className="text-sm text-red-400">{error}</p>
          <button
            onClick={() => void fetchRecommendation(projectId)}
            className="mt-3 text-xs text-red-300 underline"
          >
            Tentar novamente
          </button>
        </div>
      ) : recommendation ? (
        <>
          <CurrentPlanSection
            recommendation={recommendation}
            projectName={selectedProject?.name ?? projectId}
          />
          <PlanCards
            currentPlan={recommendation.current_plan}
            projectId={projectId}
            billing={billing}
            onBillingChange={setBilling}
          />
          <BillingHistory />
        </>
      ) : null}
    </div>
  )
}
