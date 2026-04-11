// BKL-1073: Análise estatística de experimentos A/B — p-value + CI na UI
// Consome POST /api/v1/experiments/analyze (BKL-1043 — chi-squared backend)

import { useState } from 'react'
import { FlaskConical, TrendingUp, TrendingDown, Minus, AlertTriangle } from 'lucide-react'

// ── Types ──────────────────────────────────────────────────────────────

interface ExperimentResult {
  control_conversions: number
  control_total: number
  control_rate: number
  treatment_conversions: number
  treatment_total: number
  treatment_rate: number
  chi_squared: number
  p_value: number
  statistically_significant: boolean
  confidence_level: number
  relative_lift: number
  absolute_lift: number
  recommendation: 'keep' | 'reject' | 'collect_more_data'
  sample_size: number
}

// ── API ────────────────────────────────────────────────────────────────

const BASE_URL = (import.meta.env?.VITE_PULSE_API_URL as string) || ''

async function analyzeExperiment(
  controlConversions: number,
  controlTotal: number,
  treatmentConversions: number,
  treatmentTotal: number,
): Promise<ExperimentResult> {
  const res = await fetch(`${BASE_URL}/api/v1/experiments/analyze`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      control_conversions: controlConversions,
      control_total: controlTotal,
      treatment_conversions: treatmentConversions,
      treatment_total: treatmentTotal,
    }),
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: 'Erro desconhecido' }))
    throw new Error(err.error || `HTTP ${res.status}`)
  }
  return res.json()
}

// ── Sub-components ─────────────────────────────────────────────────────

function SignificanceBadge({ significant, pValue }: { significant: boolean; pValue: number }) {
  if (significant) {
    return (
      <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-semibold bg-emerald-100 text-emerald-800 dark:bg-emerald-900/30 dark:text-emerald-400">
        ✓ Significativo (p={pValue.toFixed(4)})
      </span>
    )
  }
  return (
    <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-semibold bg-amber-100 text-amber-800 dark:bg-amber-900/30 dark:text-amber-400">
      ⚠ Não significativo (p={pValue.toFixed(4)})
    </span>
  )
}

function RecommendationCard({ recommendation, relativeLift }: { recommendation: string; relativeLift: number }) {
  const config = {
    keep: {
      icon: TrendingUp,
      color: 'emerald',
      title: 'Adotar tratamento',
      text: 'O tratamento é estatisticamente superior ao controle. Recomenda-se fazer rollout completo.',
    },
    reject: {
      icon: TrendingDown,
      color: 'red',
      title: 'Reverter tratamento',
      text: 'O controle é estatisticamente superior ao tratamento. Reverta a mudança.',
    },
    collect_more_data: {
      icon: Minus,
      color: 'slate',
      title: 'Coletar mais dados',
      text: 'Ainda não há significância estatística. Continue o experimento para obter resultados confiáveis.',
    },
  }

  const { icon: Icon, color, title, text } = config[recommendation as keyof typeof config] ?? config.collect_more_data

  return (
    <div className={`p-4 rounded-lg border border-${color}-200 dark:border-${color}-800 bg-${color}-50 dark:bg-${color}-900/20`}>
      <div className="flex items-center gap-2 mb-2">
        <Icon className={`w-4 h-4 text-${color}-600`} />
        <span className={`font-semibold text-${color}-800 dark:text-${color}-200`}>{title}</span>
        {relativeLift !== 0 && (
          <span className={`text-sm text-${color}-600`}>
            ({relativeLift > 0 ? '+' : ''}{relativeLift.toFixed(1)}% lift relativo)
          </span>
        )}
      </div>
      <p className={`text-sm text-${color}-700 dark:text-${color}-300`}>{text}</p>
    </div>
  )
}

function ConfidenceBar({ confidence }: { confidence: number }) {
  const pct = Math.min(Math.round(confidence * 100), 100)
  const color = pct >= 95 ? 'bg-emerald-500' : pct >= 80 ? 'bg-amber-500' : 'bg-slate-400'

  return (
    <div className="space-y-1">
      <div className="flex justify-between text-xs text-slate-500">
        <span>Confiança estatística</span>
        <span className="font-semibold">{pct}%</span>
      </div>
      <div className="h-2 bg-slate-200 dark:bg-slate-700 rounded-full overflow-hidden">
        <div className={`h-full ${color} rounded-full transition-all duration-700`} style={{ width: `${pct}%` }} />
      </div>
      <div className="flex justify-between text-xs text-slate-400">
        <span>0%</span>
        <span className={pct >= 95 ? 'text-emerald-600 font-semibold' : ''}>95% (threshold)</span>
        <span>100%</span>
      </div>
    </div>
  )
}

function StatCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="bg-slate-50 dark:bg-slate-800 rounded-lg p-3">
      <div className="text-xs text-slate-500 dark:text-slate-400 mb-1">{label}</div>
      <div className="font-mono font-semibold text-slate-900 dark:text-slate-100">{value}</div>
    </div>
  )
}

// ── Main Page ──────────────────────────────────────────────────────────

interface InputRow {
  label: string
  conversions: string
  total: string
  setConversions: (v: string) => void
  setTotal: (v: string) => void
}

function VariantInput({ label, conversions, total, setConversions, setTotal }: InputRow) {
  return (
    <div className="space-y-2">
      <div className="text-sm font-medium text-slate-700 dark:text-slate-300">{label}</div>
      <div className="grid grid-cols-2 gap-3">
        <div>
          <label className="block text-xs text-slate-500 mb-1">Conversões</label>
          <input
            type="number"
            min={0}
            value={conversions}
            onChange={(e) => setConversions(e.target.value)}
            className="w-full px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 text-sm focus:outline-none focus:ring-2 focus:ring-violet-500"
            placeholder="ex: 120"
          />
        </div>
        <div>
          <label className="block text-xs text-slate-500 mb-1">Total exposto</label>
          <input
            type="number"
            min={1}
            value={total}
            onChange={(e) => setTotal(e.target.value)}
            className="w-full px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 text-sm focus:outline-none focus:ring-2 focus:ring-violet-500"
            placeholder="ex: 1000"
          />
        </div>
      </div>
    </div>
  )
}

export default function ExperimentsPage() {
  const [cConv, setCConv] = useState('')
  const [cTotal, setCTotal] = useState('')
  const [tConv, setTConv] = useState('')
  const [tTotal, setTTotal] = useState('')
  const [result, setResult] = useState<ExperimentResult | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const handleAnalyze = async () => {
    setError('')
    setResult(null)
    setLoading(true)
    try {
      const res = await analyzeExperiment(
        parseInt(cConv || '0', 10),
        parseInt(cTotal || '0', 10),
        parseInt(tConv || '0', 10),
        parseInt(tTotal || '0', 10),
      )
      setResult(res)
    } catch (e) {
      setError((e as Error).message)
    } finally {
      setLoading(false)
    }
  }

  const canAnalyze = !!cTotal && parseInt(cTotal) > 0 && !!tTotal && parseInt(tTotal) > 0

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-3">
        <FlaskConical className="w-6 h-6 text-violet-500" />
        <div>
          <h1 className="text-xl font-bold text-slate-900 dark:text-slate-100">Experimentos A/B</h1>
          <p className="text-sm text-slate-500">Análise estatística com chi-squared test (α = 0,05)</p>
        </div>
      </div>

      {/* Input */}
      <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-6 space-y-4">
        <h2 className="font-semibold text-slate-800 dark:text-slate-200">Dados do experimento</h2>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <VariantInput
            label="Controle (A)"
            conversions={cConv}
            total={cTotal}
            setConversions={setCConv}
            setTotal={setCTotal}
          />
          <VariantInput
            label="Tratamento (B)"
            conversions={tConv}
            total={tTotal}
            setConversions={setTConv}
            setTotal={setTTotal}
          />
        </div>

        {error && (
          <div className="flex items-center gap-2 p-3 rounded-lg bg-red-50 dark:bg-red-900/20 text-red-700 dark:text-red-400 text-sm">
            <AlertTriangle className="w-4 h-4 shrink-0" />
            {error}
          </div>
        )}

        <button
          onClick={handleAnalyze}
          disabled={!canAnalyze || loading}
          className="w-full md:w-auto px-6 py-2.5 bg-violet-600 hover:bg-violet-700 disabled:opacity-50 disabled:cursor-not-allowed text-white font-medium rounded-lg transition-colors"
        >
          {loading ? 'Calculando...' : 'Analisar experimento'}
        </button>
      </div>

      {/* Results */}
      {result && (
        <div className="space-y-4">
          {/* Significance + Recommendation */}
          <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-6 space-y-4">
            <div className="flex items-center justify-between flex-wrap gap-3">
              <h2 className="font-semibold text-slate-800 dark:text-slate-200">Resultado</h2>
              <SignificanceBadge significant={result.statistically_significant} pValue={result.p_value} />
            </div>

            <ConfidenceBar confidence={result.confidence_level} />

            <RecommendationCard recommendation={result.recommendation} relativeLift={result.relative_lift} />
          </div>

          {/* Detailed stats */}
          <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-6 space-y-4">
            <h2 className="font-semibold text-slate-800 dark:text-slate-200">Estatísticas detalhadas</h2>

            {/* Conversion rates comparison */}
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <div className="text-xs font-medium text-slate-500 uppercase tracking-wide">Controle (A)</div>
                <div className="text-2xl font-bold text-slate-900 dark:text-slate-100">
                  {(result.control_rate * 100).toFixed(2)}%
                </div>
                <div className="text-xs text-slate-500">
                  {result.control_conversions} / {result.control_total}
                </div>
              </div>
              <div className="space-y-2">
                <div className="text-xs font-medium text-slate-500 uppercase tracking-wide">Tratamento (B)</div>
                <div className="text-2xl font-bold text-slate-900 dark:text-slate-100">
                  {(result.treatment_rate * 100).toFixed(2)}%
                </div>
                <div className="text-xs text-slate-500">
                  {result.treatment_conversions} / {result.treatment_total}
                </div>
              </div>
            </div>

            {/* Technical stats grid */}
            <div className="grid grid-cols-2 sm:grid-cols-3 gap-3">
              <StatCard label="Chi-squared (χ²)" value={result.chi_squared.toFixed(4)} />
              <StatCard label="p-value" value={result.p_value.toFixed(4)} />
              <StatCard label="Confiança" value={`${(result.confidence_level * 100).toFixed(1)}%`} />
              <StatCard label="Lift absoluto" value={`${result.absolute_lift > 0 ? '+' : ''}${(result.absolute_lift * 100).toFixed(2)}pp`} />
              <StatCard label="Lift relativo" value={`${result.relative_lift > 0 ? '+' : ''}${result.relative_lift.toFixed(1)}%`} />
              <StatCard label="Amostra total" value={result.sample_size.toLocaleString('pt-BR')} />
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
