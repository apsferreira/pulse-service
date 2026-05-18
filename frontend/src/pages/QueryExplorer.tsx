// BKL-993: SQL Explorer — query builder para eventos brutos do Pulse.
// Permite explorar events por event_type, filtros JSONB, group_by e aggregates.

import { useState } from 'react'
import { useProject } from '../components/Layout'
import { Database, Play, Download, Trash2, Plus } from 'lucide-react'
import { api } from '../services/api'

// ── Tipos ─────────────────────────────────────────────────────────────────────

interface QueryFilter {
  property: string
  operator: string
  value: string
}

interface QueryResultRow {
  group: string
  value: number
}

interface QueryResponse {
  results: QueryResultRow[]
  count: number
  query: {
    event: string
    aggregate: string
    group_by: string[]
  }
}

// ── Constantes ────────────────────────────────────────────────────────────────

const AGGREGATES = ['count', 'sum', 'avg', 'distinct_count']

const GROUP_BY_OPTIONS = [
  { value: 'browser', label: 'Browser' },
  { value: 'country', label: 'País' },
  { value: 'os', label: 'Sistema Operacional' },
  { value: 'device', label: 'Dispositivo' },
  { value: 'event_type', label: 'Tipo de Evento' },
  { value: 'environment', label: 'Ambiente' },
  { value: 'user_id', label: 'Usuário' },
]

const OPERATORS = [
  { value: 'eq', label: '=' },
  { value: 'neq', label: '!=' },
  { value: 'gt', label: '>' },
  { value: 'gte', label: '>=' },
  { value: 'lt', label: '<' },
  { value: 'lte', label: '<=' },
  { value: 'contains', label: 'contém' },
  { value: 'starts_with', label: 'começa com' },
]

// ── Utilitário de exportação CSV ──────────────────────────────────────────────

function exportCSV(rows: QueryResultRow[], filename = 'pulse-query-result.csv') {
  const header = 'group,value'
  const lines = rows.map((r) => `"${r.group.replace(/"/g, '""')}",${r.value}`)
  const csv = [header, ...lines].join('\n')
  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  a.click()
  URL.revokeObjectURL(url)
}

// ── Componente principal ───────────────────────────────────────────────────────

export default function QueryExplorer() {
  const { projectId } = useProject()
  const projectID = projectId ?? ''

  const [eventName, setEventName] = useState('')
  const [aggregate, setAggregate] = useState('count')
  const [groupBy, setGroupBy] = useState<string[]>([])
  const [filters, setFilters] = useState<QueryFilter[]>([])
  const [dateFrom, setDateFrom] = useState('')
  const [dateTo, setDateTo] = useState('')

  const [results, setResults] = useState<QueryResultRow[] | null>(null)
  const [resultMeta, setResultMeta] = useState<QueryResponse['query'] | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // ── Group By ────────────────────────────────────────────────────────────────

  const toggleGroupBy = (val: string) => {
    setGroupBy((prev) =>
      prev.includes(val) ? prev.filter((g) => g !== val) : prev.length < 3 ? [...prev, val] : prev,
    )
  }

  // ── Filters ─────────────────────────────────────────────────────────────────

  const addFilter = () =>
    setFilters((prev) => [...prev, { property: '', operator: 'eq', value: '' }])

  const updateFilter = (i: number, patch: Partial<QueryFilter>) =>
    setFilters((prev) => prev.map((f, idx) => (idx === i ? { ...f, ...patch } : f)))

  const removeFilter = (i: number) =>
    setFilters((prev) => prev.filter((_, idx) => idx !== i))

  // ── Execução ─────────────────────────────────────────────────────────────────

  const runQuery = async () => {
    if (!eventName.trim()) {
      setError('Informe o nome do evento.')
      return
    }
    if (!projectID) {
      setError('Selecione um projeto no menu superior para usar o Query Explorer.')
      return
    }

    setLoading(true)
    setError(null)
    setResults(null)

    try {
      const payload = {
        event: eventName.trim(),
        aggregate,
        group_by: groupBy,
        filters: filters.filter((f) => f.property && f.value),
        date_range: {
          from: dateFrom || undefined,
          to: dateTo || undefined,
        },
      }

      const { data } = await api.post<QueryResponse>(
        `/projects/${projectID}/query`,
        payload,
      )
      setResults(data.results ?? [])
      setResultMeta(data.query)
    } catch (err: unknown) {
      const e = err as { response?: { data?: { error?: string } }; message?: string }
      setError(e?.response?.data?.error ?? e?.message ?? 'Erro ao executar query')
    } finally {
      setLoading(false)
    }
  }

  // ── Render ───────────────────────────────────────────────────────────────────

  return (
    <div className="space-y-6 p-4 max-w-4xl mx-auto">
      <div className="flex items-center gap-2">
        <Database className="w-5 h-5 text-violet-500" />
        <h1 className="text-xl font-bold text-slate-900 dark:text-white">SQL Explorer</h1>
        <span className="ml-auto text-xs text-slate-500 dark:text-slate-400">BKL-993</span>
      </div>

      <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-5 space-y-5">

        {/* Evento */}
        <div>
          <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">
            Tipo de Evento
          </label>
          <input
            type="text"
            placeholder="ex: page_view, button_click, purchase"
            value={eventName}
            onChange={(e) => setEventName(e.target.value)}
            className="w-full rounded-lg border border-slate-300 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 px-3 py-2 text-sm text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-violet-500"
          />
        </div>

        {/* Aggregate */}
        <div>
          <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">
            Agregação
          </label>
          <select
            value={aggregate}
            onChange={(e) => setAggregate(e.target.value)}
            className="rounded-lg border border-slate-300 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 px-3 py-2 text-sm text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-violet-500"
          >
            {AGGREGATES.map((a) => (
              <option key={a} value={a}>{a}</option>
            ))}
          </select>
        </div>

        {/* Group By */}
        <div>
          <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">
            Agrupar por <span className="text-slate-400">(máx. 3)</span>
          </label>
          <div className="flex flex-wrap gap-2">
            {GROUP_BY_OPTIONS.map((opt) => (
              <button
                key={opt.value}
                onClick={() => toggleGroupBy(opt.value)}
                className={`px-3 py-1 rounded-full text-xs font-medium border transition-colors ${
                  groupBy.includes(opt.value)
                    ? 'bg-violet-600 text-white border-violet-600'
                    : 'bg-white dark:bg-slate-800 text-slate-600 dark:text-slate-300 border-slate-300 dark:border-slate-700 hover:border-violet-400'
                }`}
              >
                {opt.label}
              </button>
            ))}
          </div>
        </div>

        {/* Filtros */}
        <div>
          <div className="flex items-center justify-between mb-2">
            <label className="text-sm font-medium text-slate-700 dark:text-slate-300">
              Filtros
            </label>
            <button
              onClick={addFilter}
              className="flex items-center gap-1 text-xs text-violet-600 dark:text-violet-400 hover:underline"
            >
              <Plus className="w-3 h-3" /> Adicionar filtro
            </button>
          </div>
          <div className="space-y-2">
            {filters.map((f, i) => (
              <div key={i} className="flex items-center gap-2">
                <input
                  type="text"
                  placeholder="propriedade (ex: country)"
                  value={f.property}
                  onChange={(e) => updateFilter(i, { property: e.target.value })}
                  className="flex-1 rounded-lg border border-slate-300 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 px-3 py-2 text-sm text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-violet-500"
                />
                <select
                  value={f.operator}
                  onChange={(e) => updateFilter(i, { operator: e.target.value })}
                  className="rounded-lg border border-slate-300 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 px-2 py-2 text-sm text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-violet-500"
                >
                  {OPERATORS.map((op) => (
                    <option key={op.value} value={op.value}>{op.label}</option>
                  ))}
                </select>
                <input
                  type="text"
                  placeholder="valor"
                  value={f.value}
                  onChange={(e) => updateFilter(i, { value: e.target.value })}
                  className="flex-1 rounded-lg border border-slate-300 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 px-3 py-2 text-sm text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-violet-500"
                />
                <button
                  onClick={() => removeFilter(i)}
                  className="text-slate-400 hover:text-red-500 transition-colors"
                >
                  <Trash2 className="w-4 h-4" />
                </button>
              </div>
            ))}
          </div>
        </div>

        {/* Date Range */}
        <div className="flex gap-4">
          <div className="flex-1">
            <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">De</label>
            <input
              type="date"
              value={dateFrom}
              onChange={(e) => setDateFrom(e.target.value)}
              className="w-full rounded-lg border border-slate-300 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 px-3 py-2 text-sm text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-violet-500"
            />
          </div>
          <div className="flex-1">
            <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">Até</label>
            <input
              type="date"
              value={dateTo}
              onChange={(e) => setDateTo(e.target.value)}
              className="w-full rounded-lg border border-slate-300 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 px-3 py-2 text-sm text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-violet-500"
            />
          </div>
        </div>

        {/* Botão executar */}
        <button
          onClick={runQuery}
          disabled={loading}
          className="flex items-center gap-2 bg-violet-600 hover:bg-violet-700 disabled:opacity-50 text-white px-4 py-2 rounded-lg text-sm font-medium transition-colors"
        >
          <Play className="w-4 h-4" />
          {loading ? 'Executando...' : 'Executar Query'}
        </button>

        {error && (
          <p className="text-sm text-red-500 dark:text-red-400 bg-red-50 dark:bg-red-950/30 rounded-lg px-3 py-2">
            {error}
          </p>
        )}
      </div>

      {/* Resultados */}
      {results !== null && (
        <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-5">
          <div className="flex items-center justify-between mb-4">
            <div>
              <h2 className="text-sm font-semibold text-slate-900 dark:text-white">
                Resultados
                <span className="ml-2 text-slate-400 font-normal">{results.length} linhas</span>
              </h2>
              {resultMeta && (
                <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
                  {resultMeta.event} · {resultMeta.aggregate}
                  {resultMeta.group_by.length > 0 && ` · agrupado por ${resultMeta.group_by.join(', ')}`}
                </p>
              )}
            </div>
            {results.length > 0 && (
              <button
                onClick={() => exportCSV(results)}
                className="flex items-center gap-1 text-xs text-violet-600 dark:text-violet-400 hover:underline"
              >
                <Download className="w-3 h-3" /> Export CSV
              </button>
            )}
          </div>

          {results.length === 0 ? (
            <p className="text-sm text-slate-500 dark:text-slate-400 text-center py-8">
              Nenhum resultado encontrado para os critérios informados.
            </p>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-slate-200 dark:border-slate-800">
                    <th className="text-left py-2 px-3 text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide">
                      {groupBy.length > 0 ? groupBy.join(' / ') : 'Resultado'}
                    </th>
                    <th className="text-right py-2 px-3 text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide">
                      {aggregate}
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {results.map((row, i) => (
                    <tr
                      key={i}
                      className="border-b border-slate-100 dark:border-slate-800/60 hover:bg-slate-50 dark:hover:bg-slate-800/40 transition-colors"
                    >
                      <td className="py-2 px-3 text-slate-700 dark:text-slate-300">
                        {row.group || <span className="text-slate-400">(total)</span>}
                      </td>
                      <td className="py-2 px-3 text-right font-mono text-slate-900 dark:text-white">
                        {Number.isInteger(row.value) ? row.value : row.value.toFixed(4)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
