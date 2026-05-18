// BKL-095: Página de Cohort Analysis e DAU/WAU/MAU
// BKL-1065: Card de Churn Risk com ML (regressão logística simplificada)

import { useState, useEffect } from 'react'
import { Users, TrendingUp, Calendar, AlertTriangle, CheckCircle, ShieldAlert } from 'lucide-react'
import { metricsApi, api } from '../services/api'
import type { ActiveUsersResponse, CohortsResponse } from '../services/api'
import { useProject } from '../components/Layout'
import UpgradeModal from '../components/UpgradeModal'
import { useUpgradeGate } from '../hooks/useUpgradeGate'
import { usePlanGate } from '../hooks/usePlanGate'

// --- BKL-1065: Churn Risk Card ---

interface ChurnFactor {
  name: string
  weight: number
  value: number
  score: number
}

interface ChurnRiskResponse {
  project_id: string
  score: number
  risk: 'low' | 'medium' | 'high'
  factors: ChurnFactor[]
  period: string
  computed_at: string
}

function ChurnRiskCard({ projectID }: { projectID: string }) {
  const [data, setData] = useState<ChurnRiskResponse | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!projectID) return
    setLoading(true)
    api
      .get<ChurnRiskResponse>(`/projects/${projectID}/churn-risk`)
      .then((r) => setData(r.data))
      .catch((err: unknown) => {
        const e = err as { response?: { data?: { error?: string } }; message?: string }
        setError(e?.response?.data?.error ?? e?.message ?? 'Erro ao carregar churn risk')
      })
      .finally(() => setLoading(false))
  }, [projectID])

  const riskConfig = {
    low: {
      label: 'Risco Baixo',
      icon: CheckCircle,
      color: 'text-emerald-600 dark:text-emerald-400',
      bg: 'bg-emerald-50 dark:bg-emerald-950/30 border-emerald-200 dark:border-emerald-800',
      bar: 'bg-emerald-500',
    },
    medium: {
      label: 'Risco Médio',
      icon: AlertTriangle,
      color: 'text-amber-600 dark:text-amber-400',
      bg: 'bg-amber-50 dark:bg-amber-950/30 border-amber-200 dark:border-amber-800',
      bar: 'bg-amber-500',
    },
    high: {
      label: 'Risco Alto',
      icon: ShieldAlert,
      color: 'text-red-600 dark:text-red-400',
      bg: 'bg-red-50 dark:bg-red-950/30 border-red-200 dark:border-red-800',
      bar: 'bg-red-500',
    },
  }

  if (!projectID) return null

  return (
    <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-5">
      <h2 className="text-sm font-semibold text-slate-900 dark:text-white mb-3 flex items-center gap-2">
        <ShieldAlert className="w-4 h-4 text-violet-500" />
        Churn Risk
        <span className="text-xs font-normal text-slate-400 ml-1">BKL-1065</span>
      </h2>

      {loading && (
        <div className="flex items-center gap-2 text-sm text-slate-400 py-4">
          <div className="w-4 h-4 border-2 border-violet-500 border-t-transparent rounded-full animate-spin" />
          Calculando score...
        </div>
      )}

      {error && (
        <p className="text-sm text-red-500 dark:text-red-400">{error}</p>
      )}

      {data && (() => {
        const cfg = riskConfig[data.risk]
        const Icon = cfg.icon
        const scorePct = Math.round(data.score * 100)
        return (
          <div className="space-y-4">
            {/* Score principal */}
            <div className={`rounded-lg border px-4 py-3 flex items-center gap-3 ${cfg.bg}`}>
              <Icon className={`w-5 h-5 shrink-0 ${cfg.color}`} />
              <div className="flex-1 min-w-0">
                <p className={`text-sm font-semibold ${cfg.color}`}>{cfg.label}</p>
                <p className="text-xs text-slate-500 dark:text-slate-400">{data.period}</p>
              </div>
              <span className={`text-2xl font-bold ${cfg.color}`}>{scorePct}%</span>
            </div>

            {/* Barra de score */}
            <div className="h-2 rounded-full bg-slate-100 dark:bg-slate-800 overflow-hidden">
              <div
                className={`h-full rounded-full transition-all ${cfg.bar}`}
                style={{ width: `${scorePct}%` }}
              />
            </div>

            {/* Fatores */}
            <div className="space-y-2">
              <p className="text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide">
                Fatores contribuintes
              </p>
              {data.factors.map((f) => (
                <div key={f.name} className="flex items-center gap-3 text-sm">
                  <span className="w-20 text-slate-600 dark:text-slate-300 capitalize text-xs">{f.name}</span>
                  <div className="flex-1 h-1.5 rounded-full bg-slate-100 dark:bg-slate-800 overflow-hidden">
                    <div
                      className="h-full rounded-full bg-violet-500"
                      style={{ width: `${Math.round(f.value * 100)}%` }}
                    />
                  </div>
                  <span className="text-xs text-slate-400 w-10 text-right">
                    {Math.round(f.value * 100)}%
                  </span>
                  <span className="text-xs text-slate-400 w-12 text-right">
                    ×{f.weight}
                  </span>
                </div>
              ))}
            </div>
          </div>
        )
      })()}
    </div>
  )
}

function StatCard({ label, value, sublabel }: { label: string; value: string | number; sublabel?: string }) {
  return (
    <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-4">
      <p className="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wide">{label}</p>
      <p className="text-2xl font-bold text-slate-900 dark:text-slate-100 mt-1">{value}</p>
      {sublabel && <p className="text-xs text-slate-400 mt-0.5">{sublabel}</p>}
    </div>
  )
}

function ActiveUsersChart({ data }: { data: ActiveUsersResponse }) {
  const maxCount = Math.max(
    ...data.dau.map(d => d.count),
    ...data.wau.map(d => d.count),
    ...data.mau.map(d => d.count),
    1,
  )

  const lastDAU = data.dau[data.dau.length - 1]?.count ?? 0
  const lastWAU = data.wau[data.wau.length - 1]?.count ?? 0
  const lastMAU = data.mau[data.mau.length - 1]?.count ?? 0

  const bars = data.dau.slice(-30)

  return (
    <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-5 space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-sm font-semibold text-slate-800 dark:text-slate-100">Usuarios Ativos</h2>
        <div className="flex items-center gap-3 text-xs text-slate-500 dark:text-slate-400">
          <span className="flex items-center gap-1"><span className="w-2 h-2 rounded-full bg-violet-500 inline-block" /> DAU</span>
          <span className="flex items-center gap-1"><span className="w-2 h-2 rounded-full bg-blue-400 inline-block" /> WAU</span>
          <span className="flex items-center gap-1"><span className="w-2 h-2 rounded-full bg-emerald-400 inline-block" /> MAU</span>
        </div>
      </div>

      <div className="grid grid-cols-3 gap-3">
        <StatCard label="DAU (hoje)" value={lastDAU} sublabel="usuarios unicos" />
        <StatCard label="WAU (semana)" value={lastWAU} sublabel="usuarios unicos" />
        <StatCard label="MAU (mes)" value={lastMAU} sublabel="usuarios unicos" />
      </div>

      {/* Bar chart for DAU */}
      <div>
        <p className="text-xs text-slate-500 dark:text-slate-400 mb-2">DAU — ultimos 30 dias</p>
        <div className="flex items-end gap-0.5 h-24">
          {bars.map((item, i) => {
            const height = maxCount > 0 ? (item.count / maxCount) * 100 : 0
            return (
              <div key={i} className="flex-1 flex flex-col items-center gap-0.5 group relative">
                <div
                  className="w-full bg-violet-500 dark:bg-violet-600 rounded-t-sm transition-all hover:bg-violet-400"
                  style={{ height: `${Math.max(height, 2)}%` }}
                />
                <div className="absolute -top-6 left-1/2 -translate-x-1/2 bg-slate-800 text-white text-[10px] px-1.5 py-0.5 rounded opacity-0 group-hover:opacity-100 whitespace-nowrap pointer-events-none z-10">
                  {item.date?.slice(5, 10) ?? ''}: {item.count}
                </div>
              </div>
            )
          })}
        </div>
        <div className="flex justify-between text-[10px] text-slate-400 mt-1">
          <span>{bars[0]?.date?.slice(5, 10) ?? ''}</span>
          <span>{bars[bars.length - 1]?.date?.slice(5, 10) ?? ''}</span>
        </div>
      </div>
    </div>
  )
}

function CohortTable({ data }: { data: CohortsResponse }) {
  if (data.cohorts.length === 0) {
    return (
      <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-8 text-center">
        <p className="text-sm text-slate-500 dark:text-slate-400">Sem dados de cohort disponíveis.</p>
        <p className="text-xs text-slate-400 mt-1">Aguarde usuarios gerarem eventos de sessao.</p>
      </div>
    )
  }

  const maxBuckets = Math.max(...data.cohorts.map(c => c.buckets.length), 0)
  const headers = Array.from({ length: maxBuckets }, (_, i) => `+${i}${data.granularity === 'week' ? 'sem' : 'mes'}`)

  const retentionColor = (pct: number) => {
    if (pct >= 80) return 'bg-emerald-500'
    if (pct >= 60) return 'bg-emerald-400'
    if (pct >= 40) return 'bg-yellow-400'
    if (pct >= 20) return 'bg-orange-400'
    if (pct > 0) return 'bg-red-400'
    return 'bg-slate-200 dark:bg-slate-700'
  }

  return (
    <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-5 space-y-4">
      <h2 className="text-sm font-semibold text-slate-800 dark:text-slate-100">
        Analise de Cohort
        <span className="ml-2 text-xs text-slate-400 font-normal">
          Retencao por {data.granularity === 'week' ? 'semana' : 'mes'} apos cadastro
        </span>
      </h2>

      <div className="overflow-x-auto">
        <table className="w-full text-xs">
          <thead>
            <tr className="text-left text-slate-500 dark:text-slate-400">
              <th className="pb-2 pr-3 whitespace-nowrap font-medium">Cohort</th>
              <th className="pb-2 pr-3 whitespace-nowrap font-medium text-center">Tamanho</th>
              {headers.map((h, i) => (
                <th key={i} className="pb-2 px-1 whitespace-nowrap font-medium text-center">{h}</th>
              ))}
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-100 dark:divide-slate-800">
            {data.cohorts.map((cohort) => (
              <tr key={cohort.cohort_label}>
                <td className="py-1.5 pr-3 whitespace-nowrap font-mono text-slate-700 dark:text-slate-300">
                  {cohort.cohort_label}
                </td>
                <td className="py-1.5 pr-3 text-center font-semibold text-slate-800 dark:text-slate-200">
                  {cohort.cohort_size}
                </td>
                {Array.from({ length: maxBuckets }, (_, i) => {
                  const bucket = cohort.buckets[i]
                  const pct = bucket?.retention_pct ?? 0
                  return (
                    <td key={i} className="py-1.5 px-1 text-center">
                      {bucket ? (
                        <div
                          className={`inline-flex items-center justify-center w-12 h-7 rounded text-white text-[10px] font-semibold ${retentionColor(pct)}`}
                          title={`${bucket.active_users} usuarios (${pct.toFixed(1)}%)`}
                        >
                          {pct.toFixed(0)}%
                        </div>
                      ) : (
                        <span className="text-slate-300 dark:text-slate-700">-</span>
                      )}
                    </td>
                  )
                })}
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Legend */}
      <div className="flex items-center gap-3 text-[10px] text-slate-400 flex-wrap">
        <span className="font-medium">Retencao:</span>
        {[
          { label: '≥80%', cls: 'bg-emerald-500' },
          { label: '60-79%', cls: 'bg-emerald-400' },
          { label: '40-59%', cls: 'bg-yellow-400' },
          { label: '20-39%', cls: 'bg-orange-400' },
          { label: '<20%', cls: 'bg-red-400' },
        ].map(({ label, cls }) => (
          <span key={label} className="flex items-center gap-1">
            <span className={`w-3 h-3 rounded-sm inline-block ${cls}`} />
            {label}
          </span>
        ))}
      </div>
    </div>
  )
}

export default function Retention() {
  const { projectSlug, projectPlan } = useProject()
  const { gateState, triggerGate, closeGate } = useUpgradeGate()
  const canUseRetention = usePlanGate('starter')
  const [product, setProduct] = useState('')
  const [granularity, setGranularity] = useState<'week' | 'month'>('week')
  const [activeUsers, setActiveUsers] = useState<ActiveUsersResponse | null>(null)
  const [cohorts, setCohorts] = useState<CohortsResponse | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // Usa o slug do projeto selecionado como produto padrão
  const effectiveProduct = product || projectSlug || ''

  const fetchData = async () => {
    setLoading(true)
    setError(null)
    try {
      const [auRes, cohRes] = await Promise.all([
        metricsApi.activeUsers(effectiveProduct || undefined),
        metricsApi.cohorts(effectiveProduct || undefined, granularity),
      ])
      setActiveUsers(auRes.data)
      setCohorts(cohRes.data)
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Erro ao carregar metricas'
      setError(msg)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void fetchData()
  }, [effectiveProduct, granularity])

  if (!canUseRetention) {
    return (
      <div className="flex flex-col items-center justify-center py-24 gap-4 text-center">
        <Users className="w-12 h-12 text-violet-300" />
        <h2 className="text-xl font-bold text-slate-900 dark:text-slate-100">Retencao e Cohorts</h2>
        <p className="text-sm text-slate-500 dark:text-slate-400 max-w-sm">
          Analise de cohort, DAU/WAU/MAU e risco de churn estao disponiveis no plano Starter.
        </p>
        <button
          onClick={() => triggerGate({ plan: projectPlan, required_plan: 'starter', upgrade_url: 'https://checkout-api.institutoitinerante.com.br/checkout?plan=pulse-starter', message: 'Disponivel no plano Starter (R$399/mes). Comece com 14 dias gratis.', feature: 'Retencao e Cohorts' })}
          className="px-4 py-2 bg-violet-600 hover:bg-violet-700 text-white text-sm font-semibold rounded-lg transition-colors"
        >
          Ver planos
        </button>
        <UpgradeModal
          isOpen={gateState.isOpen}
          onClose={closeGate}
          currentPlan={gateState.currentPlan}
          requiredPlan={gateState.requiredPlan}
          feature={gateState.feature}
          upgradeUrl={gateState.upgradeUrl}
          message={gateState.message}
        />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-xl font-bold text-slate-900 dark:text-slate-100 flex items-center gap-2">
            <Users className="w-5 h-5 text-violet-600" />
            Retencao e Cohorts
          </h1>
          <p className="text-sm text-slate-500 dark:text-slate-400 mt-0.5">
            DAU/WAU/MAU e analise de cohort por periodo de cadastro
          </p>
        </div>
        <div className="flex items-center gap-2">
          {/* Product filter */}
          <input
            type="text"
            value={product}
            onChange={e => setProduct(e.target.value)}
            placeholder="Produto (ex: libri)"
            className="px-3 py-1.5 text-sm rounded-lg border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-violet-500 w-36"
          />
          {/* Granularity toggle */}
          <div className="flex rounded-lg border border-slate-300 dark:border-slate-600 overflow-hidden">
            {(['week', 'month'] as const).map(g => (
              <button
                key={g}
                onClick={() => setGranularity(g)}
                className={[
                  'px-3 py-1.5 text-xs font-medium transition-colors',
                  granularity === g
                    ? 'bg-violet-600 text-white'
                    : 'bg-white dark:bg-slate-800 text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-700',
                ].join(' ')}
              >
                {g === 'week' ? 'Semana' : 'Mes'}
              </button>
            ))}
          </div>
          <button
            onClick={fetchData}
            className="p-2 rounded-lg border border-slate-300 dark:border-slate-600 text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
            title="Atualizar"
          >
            <TrendingUp className="w-4 h-4" />
          </button>
        </div>
      </div>

      {error && (
        <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-xl p-4 text-sm text-red-700 dark:text-red-300">
          {error}
        </div>
      )}

      {/* BKL-1065: Churn Risk card — usa project_id do localStorage se disponível */}
      {(() => {
        const projectID = localStorage.getItem('pulse_project_id') ?? ''
        return projectID ? <ChurnRiskCard projectID={projectID} /> : null
      })()}

      {loading ? (
        <div className="flex items-center justify-center py-20 text-slate-400">
          <div className="w-6 h-6 border-2 border-violet-500 border-t-transparent rounded-full animate-spin mr-3" />
          <span className="text-sm">Carregando metricas...</span>
        </div>
      ) : (
        <div className="space-y-5">
          {activeUsers && <ActiveUsersChart data={activeUsers} />}
          {cohorts && <CohortTable data={cohorts} />}
          {!activeUsers && !cohorts && (
            <div className="text-center py-16 text-slate-400">
              <Calendar className="w-10 h-10 mx-auto mb-3 opacity-30" />
              <p className="text-sm">Nenhum dado disponivel.</p>
            </div>
          )}
        </div>
      )}

      <UpgradeModal
        isOpen={gateState.isOpen}
        onClose={closeGate}
        currentPlan={gateState.currentPlan}
        requiredPlan={gateState.requiredPlan}
        feature={gateState.feature}
        upgradeUrl={gateState.upgradeUrl}
        message={gateState.message}
      />
    </div>
  )
}
