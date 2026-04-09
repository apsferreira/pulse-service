import { useState, useEffect } from 'react'
import { costsApi } from '../services/api'

interface CostSummary {
  product: string
  operation: string
  count: number
  total_usd: number
  total_brl: number
}

interface CostOverview {
  from: string
  to: string
  total_usd: number
  total_brl: number
  breakdown: CostSummary[]
}

const PERIODS = [
  { label: '7d', days: 7 },
  { label: '30d', days: 30 },
  { label: '90d', days: 90 },
]

function fmtUSD(v: number) {
  return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD', minimumFractionDigits: 4 }).format(v)
}

function fmtBRL(v: number) {
  return new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL', minimumFractionDigits: 2 }).format(v)
}

export default function Costs() {
  const [days, setDays] = useState(30)
  const [overview, setOverview] = useState<CostOverview | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    setLoading(true)
    setError(null)
    costsApi
      .overview(undefined, days)
      .then((data: CostOverview) => setOverview(data))
      .catch(() => setError('Erro ao carregar dados de custos.'))
      .finally(() => setLoading(false))
  }, [days])

  const totalCalls = overview?.breakdown.reduce((acc, r) => acc + r.count, 0) ?? 0

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
        <div>
          <h1 className="text-2xl font-bold text-slate-900 dark:text-slate-100">Cost Tracking</h1>
          <p className="text-sm text-slate-500 dark:text-slate-400 mt-0.5">
            Custos de API por produto e operacao
          </p>
        </div>

        {/* Period selector */}
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

      {/* Error */}
      {error && (
        <div className="rounded-lg bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 px-4 py-3 text-sm text-red-700 dark:text-red-400">
          {error}
        </div>
      )}

      {/* Summary cards */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl p-5">
          <p className="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
            Total USD
          </p>
          {loading ? (
            <div className="mt-2 h-7 w-28 bg-slate-100 dark:bg-slate-800 rounded animate-pulse" />
          ) : (
            <p className="mt-2 text-2xl font-bold text-slate-900 dark:text-slate-100">
              {fmtUSD(overview?.total_usd ?? 0)}
            </p>
          )}
        </div>

        <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl p-5">
          <p className="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
            Total BRL
          </p>
          {loading ? (
            <div className="mt-2 h-7 w-28 bg-slate-100 dark:bg-slate-800 rounded animate-pulse" />
          ) : (
            <p className="mt-2 text-2xl font-bold text-slate-900 dark:text-slate-100">
              {fmtBRL(overview?.total_brl ?? 0)}
            </p>
          )}
        </div>

        <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl p-5">
          <p className="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
            Total de Chamadas
          </p>
          {loading ? (
            <div className="mt-2 h-7 w-16 bg-slate-100 dark:bg-slate-800 rounded animate-pulse" />
          ) : (
            <p className="mt-2 text-2xl font-bold text-slate-900 dark:text-slate-100">
              {totalCalls.toLocaleString('pt-BR')}
            </p>
          )}
        </div>
      </div>

      {/* Breakdown table */}
      <div className="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl overflow-hidden">
        <div className="px-5 py-4 border-b border-slate-200 dark:border-slate-800">
          <h2 className="text-sm font-semibold text-slate-900 dark:text-slate-100">
            Breakdown por Produto / Operacao
          </h2>
        </div>

        {loading ? (
          <div className="p-6 space-y-3">
            {[...Array(4)].map((_, i) => (
              <div key={i} className="h-4 bg-slate-100 dark:bg-slate-800 rounded animate-pulse" />
            ))}
          </div>
        ) : (overview?.breakdown.length ?? 0) === 0 ? (
          <div className="p-10 text-center text-sm text-slate-400 dark:text-slate-500">
            Nenhum custo registrado no periodo selecionado.
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-slate-200 dark:divide-slate-800">
              <thead>
                <tr className="bg-slate-50 dark:bg-slate-800/50">
                  {['Produto', 'Operacao', 'Chamadas', 'Custo USD', 'Custo BRL'].map((col) => (
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
                {overview!.breakdown.map((row, idx) => (
                  <tr
                    key={`${row.product}-${row.operation}-${idx}`}
                    className="hover:bg-slate-50 dark:hover:bg-slate-800/40 transition-colors"
                  >
                    <td className="px-5 py-3 text-sm font-medium text-slate-900 dark:text-slate-100">
                      {row.product}
                    </td>
                    <td className="px-5 py-3 text-sm text-slate-600 dark:text-slate-400">
                      {row.operation}
                    </td>
                    <td className="px-5 py-3 text-sm text-slate-600 dark:text-slate-400 tabular-nums">
                      {row.count.toLocaleString('pt-BR')}
                    </td>
                    <td className="px-5 py-3 text-sm text-slate-600 dark:text-slate-400 tabular-nums">
                      {fmtUSD(row.total_usd)}
                    </td>
                    <td className="px-5 py-3 text-sm text-slate-600 dark:text-slate-400 tabular-nums">
                      {fmtBRL(row.total_brl)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  )
}
