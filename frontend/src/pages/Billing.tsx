import { useState, useEffect, useCallback } from 'react'
import {
  CreditCard,
  CheckCircle,
  RefreshCw,
  TrendingUp,
  AlertTriangle,
  ExternalLink,
  Receipt,
  Flag,
  FolderOpen,
  Server,
  MessageCircle,
  ArrowRight,
} from 'lucide-react'
import api from '../services/api'
import { useProject } from '../components/Layout'
import { useProjects } from '../hooks/useProjects'
import {
  billingApi,
  fetchPulsePlans,
  buildUpgradeURL,
} from '../services/billingApi'
import type { UpgradeRecommendation, PlanSlug, PlanResponse } from '../services/billingApi'

// ── Tipos e constantes ────────────────────────────────────────────────────────

type BillingCycle = 'monthly' | 'yearly'

const PLAN_RANK: Record<PlanSlug, number> = {
  free: 0,
  starter: 1,
  growth: 2,
  pro: 3,
  scale: 4,
  'self-hosted': 5,
}
function planRank(slug: string): number { return PLAN_RANK[slug as PlanSlug] ?? 0 }

function resolveDisplayPrice(plan: PlanResponse, billing: BillingCycle): string {
  if (plan.price_monthly_cents === 0) return 'R$ 0'
  const monthly = plan.price_monthly_cents / 100
  const price = billing === 'yearly' ? Math.round(monthly * 0.8) : monthly
  return price.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL', maximumFractionDigits: 0 })
}

const PLAN_LABELS: Record<string, string> = {
  free: 'Free', starter: 'Starter', growth: 'Growth',
  pro: 'Pro', enterprise: 'Enterprise', internal: 'Internal',
}

function quotaBarColor(pct: number): string {
  if (pct >= 90) return 'bg-red-500'
  if (pct >= 70) return 'bg-amber-400'
  return 'bg-violet-500'
}

function quotaTextColor(pct: number): string {
  if (pct >= 90) return 'text-red-600 dark:text-red-400'
  if (pct >= 70) return 'text-amber-600 dark:text-amber-400'
  return 'text-violet-600 dark:text-violet-400'
}

// ── Plano atual ───────────────────────────────────────────────────────────────

function CurrentPlanSection({
  recommendation,
  projectName,
}: {
  recommendation: UpgradeRecommendation
  projectName: string
}) {
  const currentPlan = recommendation.current_plan
  const quota = recommendation.quota_usage
  const usagePct = Math.min(Math.round(quota?.events_pct ?? 0), 100)
  const eventsUsed = quota?.events_last_30d ?? 0
  const eventsLimit = quota?.events_limit ?? 10000
  const isUnlimited = ['pro', 'enterprise', 'internal'].includes(currentPlan)

  return (
    <section
      className="rounded-xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 p-6 space-y-5"
      aria-label="Plano atual"
    >
      <div className="flex items-center gap-3">
        <div className="w-9 h-9 rounded-lg bg-violet-100 dark:bg-violet-900/40 flex items-center justify-center flex-shrink-0">
          <CreditCard className="w-4 h-4 text-violet-600 dark:text-violet-400" aria-hidden="true" />
        </div>
        <div className="min-w-0">
          <h2 className="text-base font-semibold text-slate-900 dark:text-slate-100">Plano atual</h2>
          <p className="text-xs text-slate-500 dark:text-slate-400 truncate">{projectName}</p>
        </div>
        <span className={`ml-auto inline-flex items-center px-2.5 py-1 rounded-full text-xs font-semibold ${
          currentPlan === 'enterprise' || currentPlan === 'internal'
            ? 'bg-amber-100 text-amber-700 dark:bg-amber-900/40 dark:text-amber-300'
            : currentPlan === 'pro'
            ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300'
            : currentPlan === 'growth'
            ? 'bg-violet-100 text-violet-700 dark:bg-violet-900/40 dark:text-violet-300'
            : currentPlan === 'starter'
            ? 'bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-300'
            : 'bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-300'
        }`}>
          {PLAN_LABELS[currentPlan] ?? currentPlan}
        </span>
      </div>

      {/* Quota de eventos */}
      <div>
        <div className="flex items-center justify-between mb-1.5">
          <span className="text-xs font-medium text-slate-500 dark:text-slate-400">
            Eventos usados este mês
          </span>
          <span className={`text-xs font-semibold tabular-nums ${quotaTextColor(usagePct)}`}>
            {isUnlimited ? '∞' : `${usagePct}%`}
          </span>
        </div>
        <div className="h-2 w-full rounded-full bg-slate-100 dark:bg-slate-800 overflow-hidden">
          <div
            className={`h-full rounded-full transition-all ${isUnlimited ? 'bg-emerald-500' : quotaBarColor(usagePct)}`}
            style={{ width: isUnlimited ? '100%' : `${usagePct}%` }}
            role="progressbar"
            aria-valuenow={usagePct}
            aria-valuemin={0}
            aria-valuemax={100}
          />
        </div>
        <p className="mt-1 text-xs text-slate-400 dark:text-slate-500">
          {eventsUsed.toLocaleString('pt-BR')} /{' '}
          {isUnlimited ? '∞' : eventsLimit.toLocaleString('pt-BR')} eventos
        </p>
      </div>

      {/* KPIs */}
      <div className="grid grid-cols-2 gap-3">
        {[
          { icon: Flag, label: 'Flags ativas', value: quota?.flags_used ?? 0 },
          { icon: FolderOpen, label: 'Projetos', value: quota?.projects_used ?? 0 },
        ].map(({ icon: Icon, label, value }) => (
          <div key={label} className="rounded-lg bg-slate-50 dark:bg-slate-800 border border-slate-100 dark:border-slate-700 p-3 flex items-center gap-3">
            <Icon className="w-4 h-4 text-violet-500 flex-shrink-0" aria-hidden="true" />
            <div>
              <p className="text-xs text-slate-500 dark:text-slate-400">{label}</p>
              <p className="text-lg font-bold text-slate-900 dark:text-slate-100 tabular-nums">{value}</p>
            </div>
          </div>
        ))}
      </div>
    </section>
  )
}

// ── Toggle mensal / anual ─────────────────────────────────────────────────────

function BillingToggle({ billing, onChange }: { billing: BillingCycle; onChange: (v: BillingCycle) => void }) {
  return (
    <div
      className="inline-flex items-center rounded-lg border border-slate-200 dark:border-slate-700 bg-slate-100 dark:bg-slate-800 p-0.5"
      role="group"
      aria-label="Ciclo de cobrança"
    >
      {(['monthly', 'yearly'] as const).map((v) => (
        <button
          key={v}
          onClick={() => onChange(v)}
          aria-pressed={billing === v}
          className={[
            'px-3 py-1 rounded-md text-sm font-medium transition-colors flex items-center gap-1.5',
            billing === v
              ? 'bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 shadow-sm'
              : 'text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-200',
          ].join(' ')}
        >
          {v === 'monthly' ? 'Mensal' : 'Anual'}
          {v === 'yearly' && (
            <span className="text-[10px] font-bold px-1.5 py-0.5 rounded-full bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300">
              -20%
            </span>
          )}
        </button>
      ))}
    </div>
  )
}

// ── Self-Hosted card ──────────────────────────────────────────────────────────

function SelfHostedCard() {
  return (
    <div className="rounded-xl border border-amber-200 dark:border-amber-800/60 bg-amber-50 dark:bg-amber-900/10 p-6 flex flex-col gap-4">
      <div className="flex items-start gap-3">
        <div className="w-9 h-9 rounded-lg bg-amber-100 dark:bg-amber-900/40 flex items-center justify-center flex-shrink-0">
          <Server className="w-4 h-4 text-amber-600 dark:text-amber-400" aria-hidden="true" />
        </div>
        <div className="flex-1">
          <div className="flex items-center gap-2 flex-wrap">
            <p className="text-sm font-semibold text-slate-900 dark:text-slate-100">Self-Hosted</p>
            <span className="text-[10px] font-bold px-2 py-0.5 rounded-full bg-amber-200 text-amber-800 dark:bg-amber-800/60 dark:text-amber-300">
              Licença única
            </span>
          </div>
          <p className="text-2xl font-bold text-amber-700 dark:text-amber-400 mt-1">
            R$ 50.000
            <span className="text-xs font-normal text-slate-500 dark:text-slate-400"> /ano</span>
          </p>
        </div>
      </div>

      <p className="text-xs text-slate-600 dark:text-slate-400 leading-relaxed">
        Seus dados na sua infraestrutura. Deploy via Docker Compose ou K3s — sem dependência de cloud externa.
      </p>

      <ul className="space-y-1.5">
        {[
          'Projetos e eventos ilimitados',
          'Docker Compose ou K3s — sua escolha',
          'Suporte dedicado de onboarding',
          'SLA 99,9% garantido em contrato',
          'SSO / SAML incluído',
        ].map((f) => (
          <li key={f} className="flex items-center gap-2 text-xs text-slate-700 dark:text-slate-300">
            <CheckCircle className="w-3.5 h-3.5 text-amber-500 flex-shrink-0" aria-hidden="true" />
            {f}
          </li>
        ))}
      </ul>

      <a
        href="https://wa.me/5571999999999?text=Olá%2C%20quero%20saber%20mais%20sobre%20o%20Pulse%20Self-Hosted"
        target="_blank"
        rel="noopener noreferrer"
        className="w-full flex items-center justify-center gap-2 py-2.5 rounded-lg text-sm font-semibold bg-amber-600 hover:bg-amber-500 text-white transition-colors"
      >
        <MessageCircle className="w-4 h-4" aria-hidden="true" />
        Falar com o time
        <ExternalLink className="w-3 h-3" aria-hidden="true" />
      </a>
    </div>
  )
}

// ── Cards de planos ───────────────────────────────────────────────────────────

function PlanCards({
  currentPlan,
  projectId,
  billing,
  onBillingChange,
}: {
  currentPlan: string
  projectId: string
  billing: BillingCycle
  onBillingChange: (v: BillingCycle) => void
}) {
  const [plans, setPlans] = useState<PlanResponse[]>([])
  const [loadingPlans, setLoadingPlans] = useState(true)
  const currentRank = planRank(currentPlan)

  useEffect(() => {
    let cancelled = false
    fetchPulsePlans()
      .then((data) => { if (!cancelled) setPlans(data) })
      .catch(() => { /* planos não carregados — UI mostra skeleton */ })
      .finally(() => { if (!cancelled) setLoadingPlans(false) })
    return () => { cancelled = true }
  }, [])

  const upgradeUrl = (slug: string): string => {
    const base = buildUpgradeURL(slug as PlanSlug, projectId)
    return billing === 'yearly' ? `${base}&billing=yearly` : base
  }

  return (
    <section aria-label="Planos disponíveis">
      <div className="flex items-center justify-between gap-3 mb-5 flex-wrap">
        <div className="flex items-center gap-2">
          <TrendingUp className="w-4 h-4 text-violet-500" aria-hidden="true" />
          <h2 className="text-base font-semibold text-slate-900 dark:text-slate-100">Planos</h2>
        </div>
        <BillingToggle billing={billing} onChange={onBillingChange} />
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-5">
        {loadingPlans
          ? [1, 2, 3, 4].map((i) => (
              <div key={i} className="rounded-xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 p-5 h-64 animate-pulse" />
            ))
          : plans.map((plan) => {
              // Extrai slug sem prefixo "pulse-" para comparar com currentPlan
              const bareSlug = plan.slug.replace(/^pulse-/, '')
              const isCurrent = bareSlug === currentPlan || plan.slug === currentPlan
              const rank = planRank(bareSlug)
              const isDowngrade = rank < currentRank
              const isUpgrade = rank > currentRank

              return (
                <div
                  key={plan.slug}
                  className={[
                    'relative rounded-xl border p-5 flex flex-col gap-4 bg-white dark:bg-slate-900 transition-all',
                    isCurrent
                      ? 'border-violet-400 dark:border-violet-600 ring-1 ring-violet-400/30 dark:ring-violet-600/30'
                      : plan.highlighted
                      ? 'border-violet-300 dark:border-violet-700 shadow-md shadow-violet-100 dark:shadow-violet-900/20'
                      : 'border-slate-200 dark:border-slate-700',
                  ].join(' ')}
                >
                  {(plan.highlighted || isCurrent) && (
                    <div className={`absolute -top-2.5 left-1/2 -translate-x-1/2 px-3 py-0.5 rounded-full text-xs font-semibold whitespace-nowrap ${
                      isCurrent
                        ? 'bg-slate-100 dark:bg-slate-700 text-slate-600 dark:text-slate-300 border border-slate-200 dark:border-slate-600'
                        : 'bg-violet-600 text-white'
                    }`}>
                      {isCurrent ? 'Plano atual' : 'Mais popular'}
                    </div>
                  )}

                  <div>
                    <p className="text-sm font-semibold text-slate-900 dark:text-slate-100">{plan.name}</p>
                    <div className="flex items-baseline gap-1 mt-1">
                      <span className="text-2xl font-bold text-slate-900 dark:text-slate-100">
                        {resolveDisplayPrice(plan, billing)}
                      </span>
                      {plan.price_monthly_cents > 0 && (
                        <span className="text-xs text-slate-400">/mês</span>
                      )}
                    </div>
                    {plan.price_monthly_cents > 0 && billing === 'yearly' && (
                      <p className="text-[11px] text-emerald-600 dark:text-emerald-400 mt-0.5">
                        Economize 20% no anual
                      </p>
                    )}
                  </div>

                  <ul className="space-y-1.5 flex-1">
                    {plan.features.map((f) => (
                      <li key={f} className="flex items-start gap-1.5 text-xs text-slate-500 dark:text-slate-400">
                        <CheckCircle className="w-3.5 h-3.5 text-emerald-500 flex-shrink-0 mt-px" aria-hidden="true" />
                        {f}
                      </li>
                    ))}
                  </ul>

                  {isCurrent ? (
                    <div className="w-full text-center py-2 rounded-lg text-xs font-medium text-slate-400 dark:text-slate-500 bg-slate-50 dark:bg-slate-800">
                      Plano atual
                    </div>
                  ) : isUpgrade ? (
                    <a
                      href={upgradeUrl(plan.slug)}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="w-full flex items-center justify-center gap-1.5 py-2 rounded-lg text-xs font-semibold bg-violet-600 hover:bg-violet-500 text-white transition-colors"
                      aria-label={`Fazer upgrade para ${plan.name}`}
                    >
                      Fazer upgrade
                      <ExternalLink className="w-3 h-3" aria-hidden="true" />
                    </a>
                  ) : isDowngrade ? (
                    <div className="w-full text-center py-2 rounded-lg text-xs font-medium text-slate-300 dark:text-slate-600 bg-slate-50 dark:bg-slate-800 cursor-not-allowed">
                      Plano inferior
                    </div>
                  ) : null}
                </div>
              )
            })}
      </div>

      <SelfHostedCard />
    </section>
  )
}

// ── Histórico de cobranças ────────────────────────────────────────────────────

interface HistoryItem {
  id: string
  project_id: string
  project_name: string | null
  project_slug: string | null
  from_plan: string
  to_plan: string
  checkout_order_id: string
  created_at: string
}

function BillingHistory() {
  const [items, setItems] = useState<HistoryItem[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    api
      .get<{ history: HistoryItem[] }>('/billing/history')
      .then((r) => setItems(r.data.history ?? []))
      .catch((err) => {
        const e = err as { response?: { data?: { error?: string } }; message?: string }
        setError(e?.response?.data?.error ?? e?.message ?? 'Falha ao carregar histórico')
      })
      .finally(() => setLoading(false))
  }, [])

  return (
    <section aria-label="Histórico de cobranças">
      <div className="flex items-center gap-2 mb-4">
        <Receipt className="w-4 h-4 text-violet-500" aria-hidden="true" />
        <h2 className="text-base font-semibold text-slate-900 dark:text-slate-100">Histórico de cobranças</h2>
      </div>

      <div className="rounded-xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 overflow-hidden">
        {loading ? (
          <div className="flex items-center justify-center gap-2 py-10 text-xs text-slate-400">
            <RefreshCw className="w-4 h-4 animate-spin" />
            Carregando histórico...
          </div>
        ) : error ? (
          <div className="px-4 py-8 text-center text-sm text-red-500 dark:text-red-400">{error}</div>
        ) : items.length === 0 ? (
          <div className="px-4 py-10 text-center flex flex-col items-center gap-2">
            <Receipt className="w-8 h-8 text-slate-300 dark:text-slate-600" />
            <p className="text-sm text-slate-500">Nenhuma cobrança encontrada</p>
            <p className="text-xs text-slate-400 dark:text-slate-600">
              As cobranças aparecerão aqui após o primeiro ciclo de faturamento.
            </p>
          </div>
        ) : (
          <table className="w-full text-sm">
            <thead className="border-b border-slate-100 dark:border-slate-800 bg-slate-50 dark:bg-slate-800/50">
              <tr>
                {['Data', 'Projeto', 'Mudança de plano', 'Order ID'].map((h) => (
                  <th key={h} scope="col" className="px-4 py-3 text-left text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider">
                    {h}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-50 dark:divide-slate-800">
              {items.map((item) => (
                <tr key={item.id} className="hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-colors">
                  <td className="px-4 py-3 text-xs text-slate-600 dark:text-slate-300 whitespace-nowrap">
                    {new Date(item.created_at).toLocaleDateString('pt-BR')}
                  </td>
                  <td className="px-4 py-3">
                    <p className="text-xs font-medium text-slate-800 dark:text-slate-200">{item.project_name ?? '—'}</p>
                    <p className="text-[11px] text-slate-400 font-mono">{item.project_slug ?? ''}</p>
                  </td>
                  <td className="px-4 py-3">
                    <span className="inline-flex items-center gap-1.5 text-xs">
                      <span className="text-slate-500 dark:text-slate-400">{PLAN_LABELS[item.from_plan] ?? item.from_plan}</span>
                      <ArrowRight className="w-3 h-3 text-violet-500 flex-shrink-0" />
                      <span className="font-semibold text-violet-600 dark:text-violet-400">{PLAN_LABELS[item.to_plan] ?? item.to_plan}</span>
                    </span>
                  </td>
                  <td className="px-4 py-3">
                    <code className="text-[11px] font-mono text-slate-400 dark:text-slate-500">
                      {item.checkout_order_id || '—'}
                    </code>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </section>
  )
}

// ── Skeleton ──────────────────────────────────────────────────────────────────

function BillingSkeleton() {
  return (
    <div className="space-y-8 animate-pulse">
      <div className="rounded-xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 p-6 h-48" />
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {[1, 2, 3, 4].map((i) => (
          <div key={i} className="rounded-xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 p-5 h-64" />
        ))}
      </div>
      <div className="rounded-xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 p-6 h-32" />
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
      const e = err as { response?: { data?: { error?: string }; status?: number }; message?: string }
      if (e.response?.status === 404 || !e.response) {
        setRecommendation({
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
        })
      } else {
        setError(e?.response?.data?.error ?? e?.message ?? 'Falha ao carregar dados de faturamento')
      }
    } finally {
      setLoading(false)
    }
  }, [selectedProject, projects.length])

  useEffect(() => {
    if (projectId) void fetchRecommendation(projectId)
    else setRecommendation(null)
  }, [projectId, fetchRecommendation])

  return (
    <div className="space-y-8">
      {/* Cabeçalho */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="w-9 h-9 rounded-lg bg-violet-100 dark:bg-violet-900/40 flex items-center justify-center">
            <CreditCard className="w-4 h-4 text-violet-600 dark:text-violet-400" />
          </div>
          <div>
            <h1 className="text-xl font-bold text-slate-900 dark:text-slate-100">Faturamento</h1>
            <p className="text-sm text-slate-500 dark:text-slate-400 mt-0.5">Gerencie seu plano e acompanhe o consumo</p>
          </div>
        </div>
        {projectId && (
          <button
            onClick={() => void fetchRecommendation(projectId)}
            disabled={loading}
            className="p-2 rounded-lg text-slate-500 hover:bg-slate-100 dark:text-slate-400 dark:hover:bg-slate-800 transition-colors disabled:opacity-50"
            aria-label="Recarregar dados de faturamento"
          >
            <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
          </button>
        )}
      </div>

      {/* Conteúdo */}
      {!projectId ? (
        <div className="rounded-xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 p-12 text-center">
          <FolderOpen className="w-10 h-10 text-slate-300 dark:text-slate-600 mx-auto mb-3" />
          <p className="text-sm font-medium text-slate-600 dark:text-slate-400">Nenhum projeto selecionado</p>
          <p className="text-xs text-slate-400 dark:text-slate-600 mt-1">
            Selecione um projeto no seletor do topo para ver o faturamento.
          </p>
        </div>
      ) : loading ? (
        <BillingSkeleton />
      ) : error ? (
        <div className="rounded-xl border border-red-200 dark:border-red-900/40 bg-red-50 dark:bg-red-900/10 p-6 text-center" role="alert">
          <AlertTriangle className="w-8 h-8 text-red-500 mx-auto mb-2" />
          <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
          <button onClick={() => void fetchRecommendation(projectId)} className="mt-3 text-xs text-red-500 underline">
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
