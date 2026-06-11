// BKL-124: Página de A/B Testing — visualização de flags com variantes e distribuição de tráfego
// BKL-1049: Painel de análise estatística (p-value, CI, lift, badge significativo)
// BKL-1001: Tab de análise Bayesiana (Beta distribution + Monte Carlo P(B > A))

import { useState, useEffect, useCallback } from 'react'
import { FlaskConical, ToggleLeft, TrendingUp, Users, BarChart2, CheckCircle, Clock, Atom, Lock, Info, ArrowRight } from 'lucide-react'
import { flagsApi, api, UpgradeRequiredError } from '../services/api'
import type { FeatureFlag, FlagVariant } from '../services/api'
import { useProject } from '../components/Layout'
import UpgradeModal from '../components/UpgradeModal'
import { readEnv } from '../env'
import { useUpgradeGate } from '../hooks/useUpgradeGate'
import { usePlanGate } from '../hooks/usePlanGate'

// --- BKL-1049: Stat Analyzer Component ---

interface StatResult {
  chi_squared: number
  p_value: number
  statistically_significant: boolean
  control_rate: number
  treatment_rate: number
  lift_pct: number
  confidence_interval: { lower: number; upper: number }
  recommendation: string
}

interface StatAnalyzerProps {
  onUpgradeRequired?: (err: UpgradeRequiredError) => void
}

function StatAnalyzer({ onUpgradeRequired }: StatAnalyzerProps) {
  const [form, setForm] = useState({
    control_conversions: '',
    control_total: '',
    treatment_conversions: '',
    treatment_total: '',
  })
  const [result, setResult] = useState<StatResult | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleAnalyze = async () => {
    setLoading(true)
    setError(null)
    try {
      const { data } = await api.post<StatResult>('/experiments/analyze', {
        control_conversions: Number(form.control_conversions),
        control_total: Number(form.control_total),
        treatment_conversions: Number(form.treatment_conversions),
        treatment_total: Number(form.treatment_total),
      })
      setResult(data)
    } catch (err: unknown) {
      if (err instanceof UpgradeRequiredError) {
        onUpgradeRequired?.(err)
        return
      }
      const e = err as { response?: { data?: { error?: string } }; message?: string }
      setError(e?.response?.data?.error ?? e?.message ?? 'Erro ao analisar')
    } finally {
      setLoading(false)
    }
  }

  const isValid = ['control_conversions', 'control_total', 'treatment_conversions', 'treatment_total']
    .every((k) => Number(form[k as keyof typeof form]) > 0)

  // Medidor de confiança: 1 - p_value (0 a 1), representado como 0-100%
  const confidencePct = result ? Math.round((1 - result.p_value) * 100) : null

  return (
    <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-5">
      {/* Header com badge frequentista */}
      <div className="flex items-center justify-between gap-3 mb-4">
        <h2 className="text-base font-semibold flex items-center gap-2">
          <BarChart2 className="w-4 h-4 text-blue-500" />
          Analisador Estatístico
        </h2>
        <span className="px-2.5 py-1 rounded-full text-xs font-semibold bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-300">
          Frequentista — Chi-squared
        </span>
      </div>

      <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 mb-4">
        {[
          { key: 'control_conversions', label: 'Controle: conversões' },
          { key: 'control_total', label: 'Controle: total' },
          { key: 'treatment_conversions', label: 'Variante: conversões' },
          { key: 'treatment_total', label: 'Variante: total' },
        ].map(({ key, label }) => (
          <div key={key}>
            <label className="block text-xs text-slate-500 dark:text-slate-400 mb-1">{label}</label>
            <input
              type="number"
              min={0}
              value={form[key as keyof typeof form]}
              onChange={(e) => setForm((f) => ({ ...f, [key]: e.target.value }))}
              className="w-full px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="0"
            />
          </div>
        ))}
      </div>

      <button
        onClick={handleAnalyze}
        disabled={!isValid || loading}
        className="px-4 py-2 rounded-lg bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white text-sm font-medium transition-colors"
      >
        {loading ? 'Analisando...' : 'Analisar experimento'}
      </button>

      {error && (
        <p className="mt-3 text-sm text-red-500 dark:text-red-400">{error}</p>
      )}

      {result && confidencePct !== null && (
        <div className="mt-4 space-y-3">
          {/* Badge de significância */}
          <div className={`inline-flex items-center gap-2 px-3 py-1.5 rounded-full text-sm font-semibold ${
            result.statistically_significant
              ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400'
              : 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'
          }`}>
            {result.statistically_significant
              ? <><CheckCircle className="w-4 h-4" /> Resultado significativo (p &lt; 0,05)</>
              : <><Clock className="w-4 h-4" /> Em coleta (p ≥ 0,05)</>
            }
          </div>

          {/* Medidor visual de confiança */}
          <div className="space-y-1.5">
            <div className="flex items-center justify-between text-xs">
              <span className="text-slate-500 dark:text-slate-400">Nível de confiança (1 − p-value)</span>
              <span className={`font-semibold ${
                confidencePct >= 95 ? 'text-emerald-600 dark:text-emerald-400' :
                confidencePct >= 85 ? 'text-amber-600 dark:text-amber-400' :
                'text-slate-500 dark:text-slate-400'
              }`}>
                {confidencePct}%
              </span>
            </div>
            <div className="w-full h-2 rounded-full bg-slate-100 dark:bg-slate-800 overflow-hidden">
              <div
                className={`h-full rounded-full transition-all duration-500 ${
                  confidencePct >= 95 ? 'bg-emerald-500' :
                  confidencePct >= 85 ? 'bg-amber-500' :
                  'bg-blue-400'
                }`}
                style={{ width: `${confidencePct}%` }}
              />
            </div>
            <p className="text-[10px] text-slate-400">
              {confidencePct >= 95
                ? 'Confiança suficiente para tomar decisão'
                : confidencePct >= 85
                ? 'Quase lá — colete mais dados para confirmar'
                : 'Insuficiente — continue o experimento'}
            </p>
          </div>

          <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
            {[
              { label: 'p-value', value: result.p_value.toFixed(4) },
              { label: 'Chi-squared', value: result.chi_squared.toFixed(4) },
              { label: 'Lift', value: `${result.lift_pct > 0 ? '+' : ''}${result.lift_pct.toFixed(1)}%` },
              { label: 'IC 95%', value: `[${result.confidence_interval.lower.toFixed(3)}, ${result.confidence_interval.upper.toFixed(3)}]` },
            ].map(({ label, value }) => (
              <div key={label} className="bg-slate-50 dark:bg-slate-800 rounded-lg p-3">
                <p className="text-xs text-slate-500 dark:text-slate-400">{label}</p>
                <p className="text-sm font-bold text-slate-900 dark:text-slate-100 mt-0.5">{value}</p>
              </div>
            ))}
          </div>

          <div className="grid grid-cols-2 gap-3">
            {[
              { label: 'Taxa controle', value: `${(result.control_rate * 100).toFixed(2)}%` },
              { label: 'Taxa variante', value: `${(result.treatment_rate * 100).toFixed(2)}%` },
            ].map(({ label, value }) => (
              <div key={label} className="bg-slate-50 dark:bg-slate-800 rounded-lg p-3">
                <p className="text-xs text-slate-500 dark:text-slate-400">{label}</p>
                <p className="text-sm font-bold text-slate-900 dark:text-slate-100 mt-0.5">{value}</p>
              </div>
            ))}
          </div>

          {result.recommendation && (
            <p className="text-sm text-slate-600 dark:text-slate-400 bg-slate-50 dark:bg-slate-800 rounded-lg p-3">
              {result.recommendation}
            </p>
          )}
        </div>
      )}
    </div>
  )
}

const VARIANT_COLORS = [
  'bg-violet-500',
  'bg-blue-500',
  'bg-emerald-500',
  'bg-amber-500',
  'bg-rose-500',
  'bg-cyan-500',
]

const VARIANT_TEXT_COLORS = [
  'text-violet-600 dark:text-violet-400',
  'text-blue-600 dark:text-blue-400',
  'text-emerald-600 dark:text-emerald-400',
  'text-amber-600 dark:text-amber-400',
  'text-rose-600 dark:text-rose-400',
  'text-cyan-600 dark:text-cyan-400',
]

function VariantBar({ variants }: { variants: FlagVariant[] }) {
  const total = variants.reduce((acc, v) => acc + v.weight, 0)
  return (
    <div className="w-full flex rounded-lg overflow-hidden h-6" title={`Total: ${total}%`}>
      {variants.map((v, i) => (
        <div
          key={v.key}
          className={`${VARIANT_COLORS[i % VARIANT_COLORS.length]} flex items-center justify-center text-white text-[10px] font-medium transition-all`}
          style={{ width: `${total > 0 ? (v.weight / total) * 100 : 0}%` }}
          title={`${v.key}: ${v.weight}%`}
        >
          {v.weight >= 15 ? `${v.key} ${v.weight}%` : v.weight >= 8 ? `${v.weight}%` : ''}
        </div>
      ))}
    </div>
  )
}

function getDaysActive(createdAt: string): number {
  const created = new Date(createdAt)
  const now = new Date()
  const diffMs = now.getTime() - created.getTime()
  return Math.max(0, Math.floor(diffMs / (1000 * 60 * 60 * 24)))
}

function FlagCard({ flag }: { flag: FeatureFlag }) {
  const hasVariants = flag.variants && flag.variants.length > 0
  const weightSum = hasVariants ? flag.variants.reduce((acc, v) => acc + v.weight, 0) : 0
  const isBalanced = weightSum === 100
  const daysActive = getDaysActive(flag.created_at)
  const rollout = flag.rollout_percentage ?? 100

  return (
    <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-5 space-y-4 hover:shadow-md hover:border-violet-300 dark:hover:border-violet-700 transition-all duration-200">
      {/* Header */}
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <div className="flex items-center gap-2 flex-wrap">
            <code className="text-sm font-mono text-slate-900 dark:text-slate-100 font-semibold">
              {flag.key}
            </code>
            {hasVariants && (
              <span className="flex items-center gap-1 px-2 py-0.5 rounded-full bg-violet-100 dark:bg-violet-900/40 text-violet-600 dark:text-violet-300 text-xs font-medium">
                <FlaskConical className="w-3 h-3" />
                A/B — {flag.variants.length} variantes
              </span>
            )}
            {!hasVariants && (
              <span className="flex items-center gap-1 px-2 py-0.5 rounded-full bg-slate-100 dark:bg-slate-800 text-slate-500 text-xs">
                <ToggleLeft className="w-3 h-3" />
                On/Off
              </span>
            )}
          </div>
          {flag.description && (
            <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5 truncate">{flag.description}</p>
          )}
        </div>

        <div className="flex flex-col items-end gap-1.5 shrink-0">
          <span className={`text-xs px-2 py-1 rounded-full font-medium ${
            flag.enabled
              ? 'bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400'
              : 'bg-slate-100 dark:bg-slate-800 text-slate-500'
          }`}>
            {flag.enabled ? 'Ativa' : 'Inativa'}
          </span>
          {/* Timeline de status: badge "Ativo há X dias" */}
          <span className="flex items-center gap-1 text-[10px] text-slate-400 dark:text-slate-500">
            <Clock className="w-3 h-3" />
            {daysActive === 0 ? 'Criada hoje' : `Ativa há ${daysActive}d`}
          </span>
        </div>
      </div>

      {/* Rollout info */}
      <div className="flex items-center gap-3 text-xs text-slate-500 dark:text-slate-400">
        <span className="flex items-center gap-1">
          <Users className="w-3.5 h-3.5" />
          Rollout: <strong className="text-slate-700 dark:text-slate-300">{rollout}%</strong>
        </span>
        {flag.targeting?.user_ids && flag.targeting.user_ids.length > 0 && (
          <span>{flag.targeting.user_ids.length} user(s) targeted</span>
        )}
        {flag.targeting?.tenant_ids && flag.targeting.tenant_ids.length > 0 && (
          <span>{flag.targeting.tenant_ids.length} tenant(s) targeted</span>
        )}
      </div>

      {/* Mini funnel de rollout — barra fina */}
      <div className="space-y-1">
        <div className="flex items-center justify-between text-[10px] text-slate-400">
          <span>Cobertura do rollout</span>
          <span className="font-medium text-slate-600 dark:text-slate-300">{rollout}%</span>
        </div>
        <div className="w-full h-1.5 rounded-full bg-slate-100 dark:bg-slate-800 overflow-hidden">
          <div
            className={`h-full rounded-full transition-all duration-500 ${
              rollout === 100 ? 'bg-emerald-500' :
              rollout >= 50 ? 'bg-violet-500' :
              'bg-amber-500'
            }`}
            style={{ width: `${rollout}%` }}
          />
        </div>
      </div>

      {/* Variants */}
      {hasVariants && (
        <div className="space-y-3">
          <VariantBar variants={flag.variants} />

          {!isBalanced && (
            <p className="text-xs text-amber-600 dark:text-amber-400 flex items-center gap-1">
              <TrendingUp className="w-3.5 h-3.5" />
              Soma dos pesos: {weightSum}% (recomendado: 100%)
            </p>
          )}

          <div className="grid grid-cols-2 sm:grid-cols-3 gap-2">
            {flag.variants.map((v, i) => (
              <div
                key={v.key}
                className="rounded-lg bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 p-3"
              >
                <div className="flex items-center gap-1.5 mb-1">
                  <div className={`w-2.5 h-2.5 rounded-full shrink-0 ${VARIANT_COLORS[i % VARIANT_COLORS.length]}`} />
                  <code className="text-xs font-mono text-slate-700 dark:text-slate-300 truncate">{v.key}</code>
                </div>
                <p className={`text-xl font-bold ${VARIANT_TEXT_COLORS[i % VARIANT_TEXT_COLORS.length]}`}>
                  {v.weight}%
                </p>
                <p className="text-[10px] text-slate-400 mt-0.5">
                  do tráfego
                </p>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Timestamps */}
      <div className="text-[10px] text-slate-400 flex gap-3">
        <span>Criada: {new Date(flag.created_at).toLocaleDateString('pt-BR')}</span>
        <span>Atualizada: {new Date(flag.updated_at).toLocaleDateString('pt-BR')}</span>
      </div>
    </div>
  )
}

// --- BKL-1001: Bayesian Analysis Component ---

interface BayesianVariant {
  alpha: number
  beta: number
  mean: number
}

interface BayesianResult {
  variant_a: BayesianVariant
  variant_b: BayesianVariant
  prob_b_beats_a: number
  recommendation: string
  sample_size: number
}

function BayesianAnalysis() {
  const [form, setForm] = useState({
    a_conversions: '',
    a_total: '',
    b_conversions: '',
    b_total: '',
  })
  const [result, setResult] = useState<BayesianResult | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleAnalyze = async () => {
    setLoading(true)
    setError(null)
    try {
      const params = new URLSearchParams({
        a_conversions: form.a_conversions,
        a_total: form.a_total,
        b_conversions: form.b_conversions,
        b_total: form.b_total,
      })
      // Usa URL base + path absoluto pois o endpoint não tem prefixo /api/v1/projects
      const baseURL = readEnv('VITE_API_URL')
      const resp = await fetch(
        `${baseURL}/api/v1/experiments/mock/bayesian?${params}`,
        { headers: { 'X-Service-Token': readEnv('VITE_SERVICE_TOKEN') } },
      )
      // Na prática usa o id do experimento — usamos "mock" como placeholder
      // O endpoint ignora o :id pois é stateless (só matemática)
      if (!resp.ok) {
        const body = await resp.json().catch(() => ({})) as { error?: string }
        throw new Error(body?.error ?? `HTTP ${resp.status}`)
      }
      setResult(await resp.json() as BayesianResult)
    } catch (err: unknown) {
      const e = err as { message?: string }
      setError(e?.message ?? 'Erro ao analisar')
    } finally {
      setLoading(false)
    }
  }

  const isValid = ['a_conversions', 'a_total', 'b_conversions', 'b_total'].every(
    (k) => Number(form[k as keyof typeof form]) > 0,
  )

  const probPct = result ? Math.round(result.prob_b_beats_a * 100) : null

  const recommendationLabel: Record<string, string> = {
    deploy_b: 'Fazer deploy de B',
    keep_a: 'Manter variante A',
    collect_more_data: 'Coletar mais dados',
  }

  return (
    <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-5">
      {/* Header com badge bayesiana */}
      <div className="flex items-center justify-between gap-3 mb-4">
        <h2 className="text-base font-semibold flex items-center gap-2">
          <Atom className="w-4 h-4 text-violet-500" />
          Análise Bayesiana
        </h2>
        <span className="px-2.5 py-1 rounded-full text-xs font-semibold bg-violet-100 text-violet-700 dark:bg-violet-900/40 dark:text-violet-300">
          Bayesiana — Beta Distribution
        </span>
      </div>

      <div className="grid sm:grid-cols-2 gap-4 mb-4">
        {(['a', 'b'] as const).map((v) => (
          <div key={v} className="space-y-2">
            <p className="text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide">
              Variante {v.toUpperCase()}
            </p>
            <div className="flex gap-2">
              <div className="flex-1">
                <label className="text-xs text-slate-500 dark:text-slate-400">Conversões</label>
                <input
                  type="number"
                  min="0"
                  value={form[`${v}_conversions` as keyof typeof form]}
                  onChange={(e) => setForm((f) => ({ ...f, [`${v}_conversions`]: e.target.value }))}
                  className="w-full mt-1 rounded-lg border border-slate-300 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 px-3 py-2 text-sm text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-violet-500"
                />
              </div>
              <div className="flex-1">
                <label className="text-xs text-slate-500 dark:text-slate-400">Total</label>
                <input
                  type="number"
                  min="1"
                  value={form[`${v}_total` as keyof typeof form]}
                  onChange={(e) => setForm((f) => ({ ...f, [`${v}_total`]: e.target.value }))}
                  className="w-full mt-1 rounded-lg border border-slate-300 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 px-3 py-2 text-sm text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-violet-500"
                />
              </div>
            </div>
          </div>
        ))}
      </div>

      <button
        onClick={handleAnalyze}
        disabled={!isValid || loading}
        className="bg-violet-600 hover:bg-violet-700 disabled:opacity-50 text-white px-4 py-2 rounded-lg text-sm font-medium transition-colors"
      >
        {loading ? 'Calculando...' : 'Calcular P(B > A)'}
      </button>

      {error && (
        <p className="mt-3 text-sm text-red-500 dark:text-red-400">{error}</p>
      )}

      {result && probPct !== null && (
        <div className="mt-5 space-y-4">
          {/* Gauge visual */}
          <div className="flex flex-col items-center gap-2 py-4">
            <div className="relative w-36 h-36">
              <svg viewBox="0 0 100 100" className="w-full h-full -rotate-90">
                <circle cx="50" cy="50" r="40" fill="none" stroke="#e2e8f0" strokeWidth="12" />
                <circle
                  cx="50"
                  cy="50"
                  r="40"
                  fill="none"
                  stroke={probPct >= 95 ? '#16a34a' : probPct <= 5 ? '#dc2626' : '#7c3aed'}
                  strokeWidth="12"
                  strokeDasharray={`${probPct * 2.513} 251.3`}
                  strokeLinecap="round"
                />
              </svg>
              <div className="absolute inset-0 flex flex-col items-center justify-center">
                <span className="text-2xl font-bold text-slate-900 dark:text-white">{probPct}%</span>
                <span className="text-xs text-slate-500 dark:text-slate-400">P(B &gt; A)</span>
              </div>
            </div>
            <p className="text-sm text-slate-600 dark:text-slate-300 text-center max-w-xs">
              Há <strong>{probPct}%</strong> de probabilidade que a variante B supere a variante controle A.
            </p>
          </div>

          {/* Beta params por variante */}
          <div className="grid grid-cols-2 gap-3 text-sm">
            {(['a', 'b'] as const).map((v) => {
              const vData = v === 'a' ? result.variant_a : result.variant_b
              return (
                <div key={v} className="rounded-lg bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 p-3">
                  <p className="text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase mb-2">
                    Variante {v.toUpperCase()}
                  </p>
                  <div className="space-y-1 text-xs text-slate-600 dark:text-slate-300">
                    <p>α (alpha): <strong>{vData.alpha}</strong></p>
                    <p>β (beta): <strong>{vData.beta}</strong></p>
                    <p>Média: <strong>{(vData.mean * 100).toFixed(2)}%</strong></p>
                  </div>
                </div>
              )
            })}
          </div>

          {/* Recomendação */}
          <div className={`rounded-lg px-4 py-3 text-sm font-medium flex items-center gap-2 ${
            result.recommendation === 'deploy_b'
              ? 'bg-green-50 dark:bg-green-950/30 text-green-700 dark:text-green-400'
              : result.recommendation === 'keep_a'
              ? 'bg-red-50 dark:bg-red-950/30 text-red-700 dark:text-red-400'
              : 'bg-amber-50 dark:bg-amber-950/30 text-amber-700 dark:text-amber-400'
          }`}>
            <CheckCircle className="w-4 h-4 shrink-0" />
            Recomendação: {recommendationLabel[result.recommendation] ?? result.recommendation}
            <span className="ml-auto text-xs font-normal opacity-70">{result.sample_size} amostras · 10k simulações MC</span>
          </div>
        </div>
      )}
    </div>
  )
}

// --- Hero metric tooltip ---

function MetricTooltip({ text }: { text: string }) {
  const [visible, setVisible] = useState(false)
  return (
    <span className="relative inline-flex items-center">
      <button
        onMouseEnter={() => setVisible(true)}
        onMouseLeave={() => setVisible(false)}
        onFocus={() => setVisible(true)}
        onBlur={() => setVisible(false)}
        className="text-slate-400 hover:text-slate-600 dark:hover:text-slate-300 transition-colors"
        aria-label="Mais informações"
        type="button"
      >
        <Info className="w-3.5 h-3.5" />
      </button>
      {visible && (
        <span className="absolute z-10 bottom-full left-1/2 -translate-x-1/2 mb-2 w-52 rounded-lg bg-slate-900 dark:bg-slate-700 text-white text-xs px-3 py-2 shadow-lg text-center leading-relaxed pointer-events-none">
          {text}
        </span>
      )}
    </span>
  )
}

export default function ABTesting() {
  const { projectId, projectSlug, projectPlan } = useProject()
  const { gateState, triggerGate, closeGate } = useUpgradeGate()
  const canUseAB = usePlanGate('starter')
  const [flags, setFlags] = useState<FeatureFlag[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [filter, setFilter] = useState<'all' | 'ab' | 'onoff'>('all')

  const fetchFlags = useCallback(() => {
    if (!projectId) {
      setFlags([])
      setLoading(false)
      return
    }
    setLoading(true)
    flagsApi
      .list(projectId)
      .then(res => setFlags(res.data))
      .catch(() => setError('Falha ao carregar flags'))
      .finally(() => setLoading(false))
  }, [projectId])

  useEffect(() => {
    fetchFlags()
  }, [fetchFlags])

  const abFlags = flags.filter(f => f.variants && f.variants.length > 0)
  const onoffFlags = flags.filter(f => !f.variants || f.variants.length === 0)

  const displayed =
    filter === 'ab' ? abFlags :
    filter === 'onoff' ? onoffFlags :
    flags

  const activeAB = abFlags.filter(f => f.enabled).length
  const totalAB = abFlags.length

  if (!projectId) {
    return (
      <div className="flex flex-col items-center justify-center py-24 text-center gap-4">
        <div className="w-14 h-14 rounded-2xl bg-violet-100 dark:bg-violet-900/40 flex items-center justify-center">
          <FlaskConical className="w-7 h-7 text-violet-500" />
        </div>
        <h2 className="text-lg font-semibold text-slate-800 dark:text-slate-200">Selecione um projeto</h2>
        <p className="text-sm text-slate-500 dark:text-slate-400 max-w-xs">
          Escolha um projeto no seletor do topo para ver os experimentos A/B.
        </p>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Hero section aprimorada */}
      <div className="rounded-xl border border-slate-200 dark:border-slate-800 bg-gradient-to-br from-violet-50 to-slate-50 dark:from-violet-950/20 dark:to-slate-900 p-5">
        <div className="flex items-start justify-between gap-4 flex-wrap">
          <div>
            <h1 className="text-xl font-bold text-slate-900 dark:text-slate-100 flex items-center gap-2">
              <FlaskConical className="w-5 h-5 text-violet-500" />
              A/B Testing
            </h1>
            <p className="text-sm text-slate-500 dark:text-slate-400 mt-0.5">
              Experimentos do projeto <span className="font-medium text-violet-500">{projectSlug}</span>
            </p>
          </div>

          {/* Resumo rápido — 3 métricas hero */}
          <div className="flex items-center gap-4 flex-wrap">
            <div className="text-center">
              <p className="text-2xl font-bold text-slate-900 dark:text-slate-100">{activeAB}</p>
              <div className="flex items-center justify-center gap-1 text-xs text-slate-500 dark:text-slate-400 mt-0.5">
                <span>experimentos ativos</span>
                <MetricTooltip text="Flags A/B com variantes e status habilitado no momento." />
              </div>
            </div>
            <div className="w-px h-10 bg-slate-200 dark:bg-slate-700" />
            <div className="text-center">
              <p className="text-2xl font-bold text-slate-900 dark:text-slate-100">0</p>
              <div className="flex items-center justify-center gap-1 text-xs text-slate-500 dark:text-slate-400 mt-0.5">
                <span>conversões testadas (30d)</span>
                <MetricTooltip text="Total de eventos de conversão registrados nos últimos 30 dias em todos os experimentos ativos." />
              </div>
            </div>
            <div className="w-px h-10 bg-slate-200 dark:bg-slate-700" />
            <div className="text-center">
              <p className="text-2xl font-bold text-slate-900 dark:text-slate-100">0%</p>
              <div className="flex items-center justify-center gap-1 text-xs text-slate-500 dark:text-slate-400 mt-0.5">
                <span>usuários em teste</span>
                <MetricTooltip text="Percentual de usuários únicos que receberam pelo menos uma variante de experimento nos últimos 30 dias." />
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Gate de plano — A/B testing requer starter+ */}
      {!canUseAB && (
        <div className="rounded-xl border border-violet-500/30 bg-violet-950/30 p-5 flex items-start gap-4">
          <div className="w-10 h-10 rounded-xl bg-violet-900/50 flex items-center justify-center flex-shrink-0">
            <Lock className="w-5 h-5 text-violet-400" />
          </div>
          <div className="flex-1">
            <p className="text-sm font-semibold text-violet-200">A/B Testing disponível no plano Starter</p>
            <p className="text-xs text-violet-300/70 mt-1">
              Crie experimentos com múltiplas variantes, análise estatística com p-value e Monte Carlo.
              Plano atual: <span className="font-medium capitalize">{projectPlan}</span>.
            </p>
          </div>
          <button
            onClick={() => triggerGate({
              plan: projectPlan,
              required_plan: 'starter',
              upgrade_url: 'https://checkout-api.institutoitinerante.com.br/checkout?plan=pulse-starter',
              message: 'Faca upgrade para o plano Starter para usar A/B Testing.',
              feature: 'usar A/B Testing',
            })}
            className="flex-shrink-0 px-3 py-1.5 rounded-lg text-xs font-semibold text-white bg-violet-600 hover:bg-violet-500 transition-colors"
          >
            Ver planos
          </button>
        </div>
      )}

      {/* Stats */}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
        {[
          { label: 'Total flags', value: flags.length },
          { label: 'Experimentos A/B', value: totalAB },
          { label: 'A/B ativos', value: activeAB },
          { label: 'On/Off flags', value: onoffFlags.length },
        ].map(s => (
          <div key={s.label} className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-4">
            <p className="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wide">{s.label}</p>
            <p className="text-2xl font-bold text-slate-900 dark:text-slate-100 mt-1">{s.value}</p>
          </div>
        ))}
      </div>

      {/* Filter tabs */}
      <div className="flex gap-2">
        {(['all', 'ab', 'onoff'] as const).map(f => (
          <button
            key={f}
            onClick={() => setFilter(f)}
            className={`px-3 py-1.5 rounded-lg text-sm font-medium transition-colors ${
              filter === f
                ? 'bg-violet-600 text-white'
                : 'bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-700 text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-800'
            }`}
          >
            {f === 'all' ? 'Todas' : f === 'ab' ? `A/B (${totalAB})` : `On/Off (${onoffFlags.length})`}
          </button>
        ))}
      </div>

      {/* Seção de Análise — layout lado a lado */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        {/* BKL-1049: Painel de análise estatística (chi-squared) */}
        <StatAnalyzer
          onUpgradeRequired={(err) =>
            triggerGate({
              plan: err.plan,
              upgrade_url: err.upgradeUrl,
              message: err.message,
              required_plan: err.required_plan,
              feature: 'analisar um experimento A/B',
            })
          }
        />

        {/* BKL-1001: Análise Bayesiana (Beta distribution + Monte Carlo) */}
        <BayesianAnalysis />
      </div>

      {/* Content */}
      {loading && (
        <div className="text-center py-16 text-slate-400">
          <div className="w-8 h-8 border-2 border-violet-500 border-t-transparent rounded-full animate-spin mx-auto mb-3" />
          Carregando flags...
        </div>
      )}

      {error && (
        <div className="rounded-xl border border-red-200 dark:border-red-800 bg-red-50 dark:bg-red-900/20 px-4 py-3 text-sm text-red-600 dark:text-red-400">
          {error}
        </div>
      )}

      {/* BKL-1051: empty state com "como começar" em 3 passos */}
      {!loading && !error && displayed.length === 0 && (
        <div className="rounded-xl border border-dashed border-violet-300 dark:border-violet-700 bg-violet-50/50 dark:bg-violet-950/20 p-10 flex flex-col items-center text-center gap-6">
          <div className="w-16 h-16 rounded-2xl bg-violet-100 dark:bg-violet-900/40 flex items-center justify-center">
            <FlaskConical className="w-8 h-8 text-violet-500 dark:text-violet-400" />
          </div>
          <div className="space-y-1">
            <h2 className="text-base font-semibold text-slate-800 dark:text-slate-200">
              {filter === 'ab' ? 'Nenhum experimento A/B configurado' : 'Nenhuma flag encontrada'}
            </h2>
            <p className="text-sm text-slate-500 dark:text-slate-400 max-w-md">
              {filter === 'ab'
                ? 'Para criar um experimento A/B, edite uma flag e adicione variantes com pesos percentuais.'
                : 'Crie sua primeira flag de feature para controlar o lançamento de funcionalidades.'}
            </p>
          </div>

          {/* Como começar — 3 passos numerados */}
          <div className="w-full max-w-lg text-left space-y-3">
            <p className="text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide text-center mb-4">
              Como começar
            </p>
            {[
              {
                step: '1',
                title: 'Crie uma flag de feature',
                description: 'Na aba Feature Flags, crie uma nova flag com uma chave identificadora única.',
              },
              {
                step: '2',
                title: 'Adicione variantes com pesos',
                description: 'Edite a flag e adicione variantes (ex: control 50%, challenger 50%). A soma deve ser 100%.',
              },
              {
                step: '3',
                title: 'Use o SDK no seu código',
                description: (
                  <code className="text-violet-500 dark:text-violet-400 font-mono text-xs">
                    pulse.getVariant(&apos;minha-flag&apos;)
                  </code>
                ),
              },
            ].map(({ step, title, description }) => (
              <div key={step} className="flex items-start gap-3">
                <div className="w-6 h-6 rounded-full bg-violet-600 text-white text-xs font-bold flex items-center justify-center shrink-0 mt-0.5">
                  {step}
                </div>
                <div>
                  <p className="text-sm font-semibold text-slate-700 dark:text-slate-300 flex items-center gap-1.5">
                    {title}
                    {step !== '3' && <ArrowRight className="w-3.5 h-3.5 text-slate-400" />}
                  </p>
                  <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
                    {description}
                  </p>
                </div>
              </div>
            ))}
          </div>

          {filter === 'ab' && (
            <div className="w-full max-w-xl text-left space-y-3">
              <p className="text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide">Consultar variante no SDK</p>
              <div className="rounded-lg bg-slate-900 text-slate-100 p-4 text-xs font-mono space-y-1 overflow-x-auto">
                <p>{'const variants = await pulse.getVariants({ user_id: userId });'}</p>
                <p>{'const variant = variants["pricing-test"]; // "control" | "challenger"'}</p>
                <p>{'if (variant === "challenger") {'}</p>
                <p>{'  // mostrar novo preço'}</p>
                <p>{'}'}</p>
              </div>
            </div>
          )}
        </div>
      )}

      {!loading && !error && displayed.length > 0 && (
        <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
          {displayed.map(flag => (
            <FlagCard key={flag.id} flag={flag} />
          ))}
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
