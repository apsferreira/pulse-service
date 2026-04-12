// BKL-124: Página de A/B Testing — visualização de flags com variantes e distribuição de tráfego
// BKL-1049: Painel de análise estatística (p-value, CI, lift, badge significativo)
// BKL-1001: Tab de análise Bayesiana (Beta distribution + Monte Carlo P(B > A))

import { useState, useEffect } from 'react'
import { FlaskConical, ToggleLeft, TrendingUp, Users, BarChart2, CheckCircle, Clock, Atom } from 'lucide-react'
import { flagsApi, api } from '../services/api'
import type { FeatureFlag, FlagVariant } from '../services/api'

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

function StatAnalyzer() {
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
      const e = err as { response?: { data?: { error?: string } }; message?: string }
      setError(e?.response?.data?.error ?? e?.message ?? 'Erro ao analisar')
    } finally {
      setLoading(false)
    }
  }

  const isValid = ['control_conversions', 'control_total', 'treatment_conversions', 'treatment_total']
    .every((k) => Number(form[k as keyof typeof form]) > 0)

  return (
    <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-5">
      <h2 className="text-base font-semibold flex items-center gap-2 mb-4">
        <BarChart2 className="w-4 h-4 text-violet-500" />
        Analisador Estatístico — Chi-squared Test
      </h2>

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
              className="w-full px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-sm focus:outline-none focus:ring-2 focus:ring-violet-500"
              placeholder="0"
            />
          </div>
        ))}
      </div>

      <button
        onClick={handleAnalyze}
        disabled={!isValid || loading}
        className="px-4 py-2 rounded-lg bg-violet-600 hover:bg-violet-500 disabled:opacity-50 text-white text-sm font-medium transition-colors"
      >
        {loading ? 'Analisando...' : 'Analisar experimento'}
      </button>

      {error && (
        <p className="mt-3 text-sm text-red-500 dark:text-red-400">{error}</p>
      )}

      {result && (
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
              💡 {result.recommendation}
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

function FlagCard({ flag }: { flag: FeatureFlag }) {
  const hasVariants = flag.variants && flag.variants.length > 0
  const weightSum = hasVariants ? flag.variants.reduce((acc, v) => acc + v.weight, 0) : 0
  const isBalanced = weightSum === 100

  return (
    <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-5 space-y-4">
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

        <span className={`shrink-0 text-xs px-2 py-1 rounded-full font-medium ${
          flag.enabled
            ? 'bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400'
            : 'bg-slate-100 dark:bg-slate-800 text-slate-500'
        }`}>
          {flag.enabled ? 'Ativa' : 'Inativa'}
        </span>
      </div>

      {/* Rollout info */}
      <div className="flex items-center gap-3 text-xs text-slate-500 dark:text-slate-400">
        <span className="flex items-center gap-1">
          <Users className="w-3.5 h-3.5" />
          Rollout: <strong className="text-slate-700 dark:text-slate-300">{flag.rollout_percentage ?? 100}%</strong>
        </span>
        {flag.targeting?.user_ids && flag.targeting.user_ids.length > 0 && (
          <span>{flag.targeting.user_ids.length} user(s) targeted</span>
        )}
        {flag.targeting?.tenant_ids && flag.targeting.tenant_ids.length > 0 && (
          <span>{flag.targeting.tenant_ids.length} tenant(s) targeted</span>
        )}
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
      const baseURL = (import.meta.env?.VITE_API_URL as string) ?? ''
      const resp = await fetch(
        `${baseURL}/api/v1/experiments/mock/bayesian?${params}`,
        { headers: { 'X-Service-Token': (import.meta.env?.VITE_SERVICE_TOKEN as string) ?? '' } },
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
      <h2 className="text-base font-semibold flex items-center gap-2 mb-4">
        <Atom className="w-4 h-4 text-violet-500" />
        Análise Bayesiana — Beta Distribution
      </h2>

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

export default function ABTesting() {
  const [flags, setFlags] = useState<FeatureFlag[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [filter, setFilter] = useState<'all' | 'ab' | 'onoff'>('all')

  useEffect(() => {
    flagsApi
      .list()
      .then(res => setFlags(res.data))
      .catch(() => setError('Falha ao carregar flags'))
      .finally(() => setLoading(false))
  }, [])

  const abFlags = flags.filter(f => f.variants && f.variants.length > 0)
  const onoffFlags = flags.filter(f => !f.variants || f.variants.length === 0)

  const displayed =
    filter === 'ab' ? abFlags :
    filter === 'onoff' ? onoffFlags :
    flags

  const activeAB = abFlags.filter(f => f.enabled).length
  const totalAB = abFlags.length

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-xl font-bold text-slate-900 dark:text-slate-100 flex items-center gap-2">
          <FlaskConical className="w-5 h-5 text-violet-500" />
          A/B Testing
        </h1>
        <p className="text-sm text-slate-500 dark:text-slate-400 mt-0.5">
          Flags com variantes de distribuição — experimentos ativos no ecossistema IIT
        </p>
      </div>

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

      {/* BKL-1049: Painel de análise estatística (chi-squared) */}
      <StatAnalyzer />

      {/* BKL-1001: Análise Bayesiana (Beta distribution + Monte Carlo) */}
      <BayesianAnalysis />

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

      {/* BKL-1051: empty state com instruções de SDK */}
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
                ? 'Para criar um experimento A/B, edite uma flag e adicione variantes com pesos percentuais (ex: control 50%, challenger 50%).'
                : 'Crie uma flag de feature para começar a controlar o lançamento de funcionalidades.'}
            </p>
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
    </div>
  )
}
