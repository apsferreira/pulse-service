// CSAT — Customer Satisfaction & NPS Dashboard

import { useState, useEffect, useCallback } from 'react'
import { Star, Plus, Copy, Check, X, ChevronDown } from 'lucide-react'
import { api } from '../services/api'
import { useProject } from '../components/Layout'
import UpgradeModal from '../components/UpgradeModal'
import { useUpgradeGate } from '../hooks/useUpgradeGate'
import { usePlanGate } from '../hooks/usePlanGate'

function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const method = (init?.method ?? 'GET') as 'get' | 'post' | 'put' | 'delete' | 'patch'
  const body = init?.body ? JSON.parse(init.body as string) : undefined
  return api.request<T>({ method, url: path, data: body }).then((r) => r.data)
}

// ── Types ─────────────────────────────────────────────────────────────────────

interface Survey {
  id: string
  product: string
  question: string
  type: 'NPS' | 'CSAT'
  scale: number
  active: boolean
  created_at: string
}

interface NpsResult {
  survey_id: string
  nps_score: number
  promoters: number
  neutrals: number
  detractors: number
  total: number
  average_score: number
}

type TabKey = 'nps' | 'surveys' | 'embed'

const PERIODS = [
  { label: '7d', days: 7 },
  { label: '30d', days: 30 },
  { label: '90d', days: 90 },
]

// ── Sub-components ─────────────────────────────────────────────────────────────

function SkeletonRect({ className }: { className: string }) {
  return <div className={`bg-slate-100 dark:bg-slate-800 rounded animate-pulse ${className}`} />
}

function ErrorBanner({ message }: { message: string }) {
  return (
    <div className="rounded-lg bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 px-4 py-3 text-sm text-red-700 dark:text-red-400">
      {message}
    </div>
  )
}

function NpsBar({ promoters, neutrals, detractors }: { promoters: number; neutrals: number; detractors: number }) {
  const total = promoters + neutrals + detractors
  if (total === 0) return null

  const pPct = (promoters / total) * 100
  const nPct = (neutrals / total) * 100
  const dPct = (detractors / total) * 100

  return (
    <div className="w-full flex rounded-lg overflow-hidden h-7" aria-label="Distribuição NPS">
      {pPct > 0 && (
        <div
          className="bg-emerald-500 flex items-center justify-center text-white text-xs font-medium transition-all"
          style={{ width: `${pPct}%` }}
          title={`Promotores: ${promoters} (${pPct.toFixed(1)}%)`}
        >
          {pPct >= 12 ? `${pPct.toFixed(0)}%` : ''}
        </div>
      )}
      {nPct > 0 && (
        <div
          className="bg-amber-400 flex items-center justify-center text-white text-xs font-medium transition-all"
          style={{ width: `${nPct}%` }}
          title={`Neutros: ${neutrals} (${nPct.toFixed(1)}%)`}
        >
          {nPct >= 12 ? `${nPct.toFixed(0)}%` : ''}
        </div>
      )}
      {dPct > 0 && (
        <div
          className="bg-red-500 flex items-center justify-center text-white text-xs font-medium transition-all"
          style={{ width: `${dPct}%` }}
          title={`Detratores: ${detractors} (${dPct.toFixed(1)}%)`}
        >
          {dPct >= 12 ? `${dPct.toFixed(0)}%` : ''}
        </div>
      )}
    </div>
  )
}

// ── NPS Tab ───────────────────────────────────────────────────────────────────

function NpsTab({ surveys }: { surveys: Survey[] }) {
  const [selectedId, setSelectedId] = useState<string>('')
  const [days, setDays] = useState(30)
  const [result, setResult] = useState<NpsResult | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const npsSurveys = surveys.filter((s) => s.type === 'NPS')

  useEffect(() => {
    if (npsSurveys.length > 0 && !selectedId) {
      setSelectedId(npsSurveys[0].id)
    }
  }, [npsSurveys, selectedId])

  const fetchNps = useCallback(() => {
    if (!selectedId) return
    setLoading(true)
    setError(null)
    apiFetch<NpsResult>(`/csat/surveys/${selectedId}/nps?days=${days}`)
      .then((data) => setResult(data))
      .catch((err: unknown) => {
        const msg = err instanceof Error ? err.message : 'Erro ao carregar NPS'
        setError(msg)
      })
      .finally(() => setLoading(false))
  }, [selectedId, days])

  useEffect(() => {
    fetchNps()
  }, [fetchNps])

  const npsColor =
    result === null
      ? 'text-slate-400'
      : result.nps_score >= 50
        ? 'text-emerald-600 dark:text-emerald-400'
        : result.nps_score >= 0
          ? 'text-amber-600 dark:text-amber-400'
          : 'text-red-600 dark:text-red-400'

  return (
    <div className="space-y-6">
      {/* Controles */}
      <div className="flex flex-col sm:flex-row sm:items-center gap-3">
        <div className="relative flex-1 max-w-xs">
          <select
            value={selectedId}
            onChange={(e) => setSelectedId(e.target.value)}
            className="w-full appearance-none bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-700 rounded-lg px-3 py-2 pr-8 text-sm text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-violet-500"
            aria-label="Selecionar survey"
          >
            {npsSurveys.length === 0 && (
              <option value="">Nenhuma survey NPS encontrada</option>
            )}
            {npsSurveys.map((s) => (
              <option key={s.id} value={s.id}>
                {s.product} — {s.question.slice(0, 40)}{s.question.length > 40 ? '…' : ''}
              </option>
            ))}
          </select>
          <ChevronDown className="pointer-events-none absolute right-2 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
        </div>

        <div className="flex items-center gap-1 bg-slate-100 dark:bg-slate-800 rounded-lg p-1">
          {PERIODS.map((p) => (
            <button
              key={p.days}
              onClick={() => setDays(p.days)}
              className={[
                'px-3 py-1.5 rounded-md text-sm font-medium transition-colors',
                days === p.days
                  ? 'bg-white dark:bg-slate-700 text-slate-900 dark:text-slate-100 shadow-sm'
                  : 'text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-300',
              ].join(' ')}
            >
              {p.label}
            </button>
          ))}
        </div>
      </div>

      {error && <ErrorBanner message={error} />}

      {/* Score NPS em destaque */}
      <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl p-6">
        <p className="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider mb-2">
          Score NPS
        </p>
        {loading ? (
          <SkeletonRect className="h-20 w-40" />
        ) : (
          <p className={`text-7xl font-extrabold leading-none ${npsColor}`}>
            {result ? (result.nps_score > 0 ? `+${result.nps_score}` : String(result.nps_score)) : '—'}
          </p>
        )}
        {!loading && result && (
          <p className="text-sm text-slate-500 dark:text-slate-400 mt-2">
            Baseado em <strong className="text-slate-700 dark:text-slate-300">{result.total}</strong> respostas
            nos ultimos <strong className="text-slate-700 dark:text-slate-300">{days} dias</strong>
            {' '}&middot; media{' '}
            <strong className="text-slate-700 dark:text-slate-300">
              {result.average_score.toFixed(1)}
            </strong>
          </p>
        )}
      </div>

      {/* Counters */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        {[
          {
            label: 'Promotores',
            value: result?.promoters,
            color: 'text-emerald-600 dark:text-emerald-400',
            border: 'border-emerald-200 dark:border-emerald-800',
            bg: 'bg-emerald-50 dark:bg-emerald-900/20',
            dot: 'bg-emerald-500',
            tip: 'Score 9–10',
          },
          {
            label: 'Neutros',
            value: result?.neutrals,
            color: 'text-amber-600 dark:text-amber-400',
            border: 'border-amber-200 dark:border-amber-800',
            bg: 'bg-amber-50 dark:bg-amber-900/20',
            dot: 'bg-amber-400',
            tip: 'Score 7–8',
          },
          {
            label: 'Detratores',
            value: result?.detractors,
            color: 'text-red-600 dark:text-red-400',
            border: 'border-red-200 dark:border-red-800',
            bg: 'bg-red-50 dark:bg-red-900/20',
            dot: 'bg-red-500',
            tip: 'Score 0–6',
          },
        ].map(({ label, value, color, border, bg, dot, tip }) => (
          <div key={label} className={`rounded-xl border ${border} ${bg} p-5`}>
            <div className="flex items-center gap-1.5 mb-2">
              <div className={`w-2.5 h-2.5 rounded-full ${dot}`} />
              <p className="text-xs font-medium text-slate-600 dark:text-slate-400 uppercase tracking-wider">
                {label}
              </p>
            </div>
            {loading ? (
              <SkeletonRect className="h-8 w-16" />
            ) : (
              <p className={`text-3xl font-bold ${color}`}>
                {value ?? '—'}
              </p>
            )}
            <p className="text-[10px] text-slate-400 mt-0.5">{tip}</p>
          </div>
        ))}
      </div>

      {/* Barra de distribuicao */}
      {!loading && result && result.total > 0 && (
        <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl p-5 space-y-3">
          <p className="text-sm font-semibold text-slate-700 dark:text-slate-300">Distribuicao</p>
          <NpsBar
            promoters={result.promoters}
            neutrals={result.neutrals}
            detractors={result.detractors}
          />
          <div className="flex items-center gap-4 text-xs text-slate-500 dark:text-slate-400">
            <span className="flex items-center gap-1">
              <span className="w-2 h-2 rounded-full bg-emerald-500 inline-block" />
              Promotores {result.total > 0 ? ((result.promoters / result.total) * 100).toFixed(1) : 0}%
            </span>
            <span className="flex items-center gap-1">
              <span className="w-2 h-2 rounded-full bg-amber-400 inline-block" />
              Neutros {result.total > 0 ? ((result.neutrals / result.total) * 100).toFixed(1) : 0}%
            </span>
            <span className="flex items-center gap-1">
              <span className="w-2 h-2 rounded-full bg-red-500 inline-block" />
              Detratores {result.total > 0 ? ((result.detractors / result.total) * 100).toFixed(1) : 0}%
            </span>
          </div>
        </div>
      )}

      {!loading && result && result.total === 0 && (
        <div className="text-center py-12 text-slate-400 dark:text-slate-500">
          <Star className="w-10 h-10 mx-auto mb-2 opacity-30" />
          <p>Nenhuma resposta no periodo selecionado.</p>
        </div>
      )}
    </div>
  )
}

// ── Modal Nova Survey ─────────────────────────────────────────────────────────

interface NewSurveyPayload {
  product: string
  question: string
  type: 'NPS' | 'CSAT'
  scale: number
}

function NewSurveyModal({
  onClose,
  onCreated,
}: {
  onClose: () => void
  onCreated: () => void
}) {
  const [form, setForm] = useState<NewSurveyPayload>({
    product: '',
    question: '',
    type: 'NPS',
    scale: 10,
  })
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault()
    if (!form.product.trim() || !form.question.trim()) {
      setError('Produto e pergunta sao obrigatorios.')
      return
    }
    setLoading(true)
    setError(null)
    apiFetch<Survey>('/csat/surveys', {
      method: 'POST',
      body: JSON.stringify(form),
    })
      .then(() => {
        onCreated()
        onClose()
      })
      .catch((err: unknown) => {
        const msg = err instanceof Error ? err.message : 'Erro ao criar survey'
        setError(msg)
      })
      .finally(() => setLoading(false))
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm px-4"
      role="dialog"
      aria-modal="true"
      aria-label="Nova survey"
    >
      <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-700 rounded-2xl w-full max-w-md shadow-xl">
        <div className="flex items-center justify-between px-5 py-4 border-b border-slate-200 dark:border-slate-700">
          <h2 className="font-semibold text-slate-900 dark:text-slate-100">Nova Survey</h2>
          <button
            onClick={onClose}
            className="p-1.5 rounded-lg text-slate-400 hover:text-slate-600 dark:hover:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
            aria-label="Fechar modal"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-5 space-y-4">
          {error && <ErrorBanner message={error} />}

          <div>
            <label
              htmlFor="csat-product"
              className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1"
            >
              Produto
            </label>
            <input
              id="csat-product"
              type="text"
              value={form.product}
              onChange={(e) => setForm((f) => ({ ...f, product: e.target.value }))}
              placeholder="ex: socialmake, libri, nitro"
              className="w-full bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg px-3 py-2 text-sm text-slate-900 dark:text-slate-100 placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-violet-500"
              required
            />
          </div>

          <div>
            <label
              htmlFor="csat-question"
              className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1"
            >
              Pergunta
            </label>
            <input
              id="csat-question"
              type="text"
              value={form.question}
              onChange={(e) => setForm((f) => ({ ...f, question: e.target.value }))}
              placeholder="ex: De 0 a 10, quanto voce recomendaria?"
              className="w-full bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg px-3 py-2 text-sm text-slate-900 dark:text-slate-100 placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-violet-500"
              required
            />
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label
                htmlFor="csat-type"
                className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1"
              >
                Tipo
              </label>
              <div className="relative">
                <select
                  id="csat-type"
                  value={form.type}
                  onChange={(e) => setForm((f) => ({ ...f, type: e.target.value as 'NPS' | 'CSAT' }))}
                  className="w-full appearance-none bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg px-3 py-2 pr-8 text-sm text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-violet-500"
                >
                  <option value="NPS">NPS</option>
                  <option value="CSAT">CSAT</option>
                </select>
                <ChevronDown className="pointer-events-none absolute right-2 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
              </div>
            </div>

            <div>
              <label
                htmlFor="csat-scale"
                className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1"
              >
                Escala
              </label>
              <input
                id="csat-scale"
                type="number"
                min={1}
                max={100}
                value={form.scale}
                onChange={(e) => setForm((f) => ({ ...f, scale: Number(e.target.value) }))}
                className="w-full bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg px-3 py-2 text-sm text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-violet-500"
              />
            </div>
          </div>

          <div className="flex justify-end gap-2 pt-1">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 rounded-lg text-sm font-medium text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
            >
              Cancelar
            </button>
            <button
              type="submit"
              disabled={loading}
              className="px-4 py-2 rounded-lg text-sm font-medium bg-violet-600 hover:bg-violet-700 disabled:opacity-60 text-white transition-colors"
            >
              {loading ? 'Criando...' : 'Criar'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

// ── Surveys Tab ───────────────────────────────────────────────────────────────

function SurveysTab({
  surveys,
  loading,
  error,
  onRefresh,
}: {
  surveys: Survey[]
  loading: boolean
  error: string | null
  onRefresh: () => void
}) {
  const [showModal, setShowModal] = useState(false)
  const [togglingId, setTogglingId] = useState<string | null>(null)

  function handleToggle(survey: Survey) {
    setTogglingId(survey.id)
    apiFetch<Survey>(`/csat/surveys/${survey.id}`, {
      method: 'PUT',
      body: JSON.stringify({ active: !survey.active }),
    })
      .then(() => onRefresh())
      .catch(() => {
        // silencioso — o refresh nao acontece, estado permanece
      })
      .finally(() => setTogglingId(null))
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <p className="text-sm text-slate-500 dark:text-slate-400">
          {surveys.length} survey{surveys.length !== 1 ? 's' : ''} cadastrada{surveys.length !== 1 ? 's' : ''}
        </p>
        <button
          onClick={() => setShowModal(true)}
          className="flex items-center gap-1.5 px-3 py-2 rounded-lg bg-violet-600 hover:bg-violet-700 text-white text-sm font-medium transition-colors"
        >
          <Plus className="w-4 h-4" aria-hidden="true" />
          Nova Survey
        </button>
      </div>

      {error && <ErrorBanner message={error} />}

      {loading ? (
        <div className="space-y-3">
          {[...Array(3)].map((_, i) => (
            <SkeletonRect key={i} className="h-14 w-full" />
          ))}
        </div>
      ) : surveys.length === 0 ? (
        <div className="text-center py-12 text-slate-400 dark:text-slate-500">
          <Star className="w-10 h-10 mx-auto mb-2 opacity-30" />
          <p>Nenhuma survey cadastrada.</p>
        </div>
      ) : (
        <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl overflow-hidden">
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-slate-100 dark:divide-slate-800">
              <thead>
                <tr className="bg-slate-50 dark:bg-slate-800/50">
                  {['Produto', 'Pergunta', 'Tipo', 'Ativa', 'Criada em', 'Acoes'].map((col) => (
                    <th
                      key={col}
                      className="px-5 py-3 text-left text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider"
                    >
                      {col}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100 dark:divide-slate-800">
                {surveys.map((s) => (
                  <tr
                    key={s.id}
                    className="hover:bg-slate-50 dark:hover:bg-slate-800/40 transition-colors"
                  >
                    <td className="px-5 py-3 text-sm font-medium text-slate-900 dark:text-slate-100">
                      {s.product}
                    </td>
                    <td className="px-5 py-3 text-sm text-slate-600 dark:text-slate-400 max-w-xs truncate">
                      {s.question}
                    </td>
                    <td className="px-5 py-3">
                      <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${
                        s.type === 'NPS'
                          ? 'bg-violet-100 dark:bg-violet-900/40 text-violet-700 dark:text-violet-300'
                          : 'bg-blue-100 dark:bg-blue-900/40 text-blue-700 dark:text-blue-300'
                      }`}>
                        {s.type}
                      </span>
                    </td>
                    <td className="px-5 py-3">
                      <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${
                        s.active
                          ? 'bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400'
                          : 'bg-slate-100 dark:bg-slate-800 text-slate-500'
                      }`}>
                        {s.active ? 'Sim' : 'Nao'}
                      </span>
                    </td>
                    <td className="px-5 py-3 text-sm text-slate-500 dark:text-slate-400 tabular-nums whitespace-nowrap">
                      {new Date(s.created_at).toLocaleDateString('pt-BR')}
                    </td>
                    <td className="px-5 py-3">
                      <button
                        onClick={() => handleToggle(s)}
                        disabled={togglingId === s.id}
                        className={`px-3 py-1 rounded-lg text-xs font-medium transition-colors disabled:opacity-50 ${
                          s.active
                            ? 'bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400 hover:bg-slate-200 dark:hover:bg-slate-700'
                            : 'bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400 hover:bg-green-200 dark:hover:bg-green-800/40'
                        }`}
                        aria-label={s.active ? `Desativar survey ${s.product}` : `Ativar survey ${s.product}`}
                      >
                        {togglingId === s.id ? '...' : s.active ? 'Desativar' : 'Ativar'}
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {showModal && (
        <NewSurveyModal
          onClose={() => setShowModal(false)}
          onCreated={onRefresh}
        />
      )}
    </div>
  )
}

// ── Embed Tab ─────────────────────────────────────────────────────────────────

function EmbedTab({ surveys }: { surveys: Survey[] }) {
  const [selectedId, setSelectedId] = useState<string>('')
  const [copied, setCopied] = useState(false)

  useEffect(() => {
    if (surveys.length > 0 && !selectedId) {
      setSelectedId(surveys[0].id)
    }
  }, [surveys, selectedId])

  const snippet = `<script>
  window.PulseCsat = { surveyId: "${selectedId || 'SURVEY_ID'}" };
</script>
<script src="https://pulse.institutoitinerante.com.br/public/csat/widget.js" async></script>`

  function handleCopy() {
    navigator.clipboard.writeText(snippet).then(() => {
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    })
  }

  return (
    <div className="space-y-5">
      <div>
        <p className="text-sm text-slate-600 dark:text-slate-400 mb-4">
          Incorpore o widget de feedback em qualquer site externo. Selecione a survey desejada e copie o snippet abaixo.
        </p>

        {surveys.length > 0 && (
          <div className="relative max-w-xs mb-4">
            <label htmlFor="embed-survey" className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">
              Survey
            </label>
            <select
              id="embed-survey"
              value={selectedId}
              onChange={(e) => setSelectedId(e.target.value)}
              className="w-full appearance-none bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-700 rounded-lg px-3 py-2 pr-8 text-sm text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-violet-500"
            >
              {surveys.map((s) => (
                <option key={s.id} value={s.id}>
                  {s.product} ({s.type})
                </option>
              ))}
            </select>
            <ChevronDown className="pointer-events-none absolute right-2 bottom-2.5 w-4 h-4 text-slate-400" />
          </div>
        )}
      </div>

      <div className="bg-slate-900 dark:bg-slate-950 border border-slate-700 rounded-xl overflow-hidden">
        <div className="flex items-center justify-between px-4 py-2.5 border-b border-slate-700">
          <span className="text-xs text-slate-400 font-mono">HTML</span>
          <button
            onClick={handleCopy}
            className="flex items-center gap-1.5 px-2.5 py-1 rounded-md text-xs font-medium text-slate-300 hover:text-white hover:bg-slate-700 transition-colors"
            aria-label="Copiar snippet"
          >
            {copied ? (
              <>
                <Check className="w-3.5 h-3.5 text-emerald-400" />
                <span className="text-emerald-400">Copiado!</span>
              </>
            ) : (
              <>
                <Copy className="w-3.5 h-3.5" />
                Copiar
              </>
            )}
          </button>
        </div>
        <pre className="p-4 text-sm text-slate-200 font-mono overflow-x-auto leading-relaxed">
          {snippet}
        </pre>
      </div>

      <div className="rounded-xl border border-amber-200 dark:border-amber-800 bg-amber-50 dark:bg-amber-900/20 px-4 py-3 text-sm text-amber-700 dark:text-amber-400">
        <strong>Nota:</strong> O arquivo <code className="font-mono text-xs">widget.js</code> ainda nao esta disponivel em producao. Este snippet representa a integracao prevista.
      </div>
    </div>
  )
}

// ── Main Page ─────────────────────────────────────────────────────────────────

const TABS: { key: TabKey; label: string }[] = [
  { key: 'nps', label: 'NPS Dashboard' },
  { key: 'surveys', label: 'Surveys' },
  { key: 'embed', label: 'Widget Embed' },
]

export default function CSAT() {
  const { gateState, triggerGate, closeGate } = useUpgradeGate()
  const canUseCSAT = usePlanGate('starter')
  const { projectPlan } = useProject()
  const [activeTab, setActiveTab] = useState<TabKey>('nps')
  const [surveys, setSurveys] = useState<Survey[]>([])
  const [surveysLoading, setSurveysLoading] = useState(true)
  const [surveysError, setSurveysError] = useState<string | null>(null)

  const fetchSurveys = useCallback(() => {
    setSurveysLoading(true)
    setSurveysError(null)
    apiFetch<Survey[] | { surveys: Survey[]; count: number }>('/csat/surveys')
      .then((data) => {
        const raw = Array.isArray(data)
          ? data
          : Array.isArray((data as { surveys: Survey[] }).surveys)
            ? (data as { surveys: Survey[] }).surveys
            : []
        // normalize type to uppercase (API returns 'nps'/'csat')
        setSurveys(raw.map((s) => ({ ...s, type: (s.type as string).toUpperCase() as 'NPS' | 'CSAT' })))
      })
      .catch((err: unknown) => {
        const msg = err instanceof Error ? err.message : 'Erro ao carregar surveys'
        setSurveysError(msg)
      })
      .finally(() => setSurveysLoading(false))
  }, [])

  useEffect(() => {
    fetchSurveys()
  }, [fetchSurveys])

  if (!canUseCSAT) {
    return (
      <div className="flex flex-col items-center justify-center py-24 gap-4 text-center">
        <Star className="w-12 h-12 text-violet-300" />
        <h2 className="text-xl font-bold text-slate-900 dark:text-slate-100">CSAT / NPS</h2>
        <p className="text-sm text-slate-500 dark:text-slate-400 max-w-sm">
          Pesquisas de satisfacao e Net Promoter Score estao disponiveis no plano Starter.
        </p>
        <button
          onClick={() => triggerGate({ plan: projectPlan, required_plan: 'starter', upgrade_url: 'https://checkout-api.institutoitinerante.com.br/checkout?plan=pulse-starter', message: 'Disponivel no plano Starter (R$399/mes). Comece com 14 dias gratis.', feature: 'CSAT / NPS' })}
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
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
        <div>
          <h1 className="text-2xl font-bold text-slate-900 dark:text-slate-100 flex items-center gap-2">
            <Star className="w-6 h-6 text-violet-500" aria-hidden="true" />
            CSAT / NPS
          </h1>
          <p className="text-sm text-slate-500 dark:text-slate-400 mt-0.5">
            Customer Satisfaction Score e Net Promoter Score
          </p>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 border-b border-slate-200 dark:border-slate-800">
        {TABS.map(({ key, label }) => (
          <button
            key={key}
            onClick={() => setActiveTab(key)}
            className={[
              'px-4 py-2.5 text-sm font-medium border-b-2 -mb-px transition-colors',
              activeTab === key
                ? 'border-violet-600 text-violet-700 dark:text-violet-400'
                : 'border-transparent text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-300',
            ].join(' ')}
            aria-selected={activeTab === key}
            role="tab"
          >
            {label}
          </button>
        ))}
      </div>

      {/* Content */}
      <div role="tabpanel">
        {activeTab === 'nps' && (
          <NpsTab surveys={surveys} />
        )}
        {activeTab === 'surveys' && (
          <SurveysTab
            surveys={surveys}
            loading={surveysLoading}
            error={surveysError}
            onRefresh={fetchSurveys}
          />
        )}
        {activeTab === 'embed' && (
          <EmbedTab surveys={surveys} />
        )}
      </div>

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
