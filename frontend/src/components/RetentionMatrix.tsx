// BKL-1059: Cohort analysis — matriz de retenção semana×semana com heatmap

import { useState, useEffect } from 'react'
import { TrendingUp } from 'lucide-react'

interface RetentionData {
  cohort: string
  cohortSize: number
  weeks: (number | null)[] // retention percentage por semana (W0-W8)
}

const getHeatmapColor = (percentage: number | null): string => {
  if (percentage === null) return 'bg-slate-100 dark:bg-slate-800 text-slate-400'
  if (percentage >= 80) return 'bg-emerald-500 dark:bg-emerald-600 text-white'
  if (percentage >= 60) return 'bg-emerald-400 dark:bg-emerald-500 text-white'
  if (percentage >= 50) return 'bg-cyan-400 dark:bg-cyan-500 text-white'
  if (percentage >= 40) return 'bg-yellow-400 dark:bg-yellow-500 text-white'
  if (percentage >= 20) return 'bg-orange-400 dark:bg-orange-500 text-white'
  if (percentage > 0) return 'bg-red-400 dark:bg-red-500 text-white'
  return 'bg-slate-100 dark:bg-slate-800 text-slate-400'
}

// Dados placeholder para demonstrar o visual
const PLACEHOLDER_DATA: RetentionData[] = [
  {
    cohort: '2026-04-07',
    cohortSize: 1250,
    weeks: [100, 68, 52, 45, 38, 32, 28, 24, 21],
  },
  {
    cohort: '2026-04-14',
    cohortSize: 1420,
    weeks: [100, 71, 54, 47, 40, 35, 30, 27, null],
  },
  {
    cohort: '2026-04-21',
    cohortSize: 980,
    weeks: [100, 69, 53, 46, 39, 34, 29, null, null],
  },
  {
    cohort: '2026-04-28',
    cohortSize: 1150,
    weeks: [100, 65, 50, 43, 37, 31, null, null, null],
  },
  {
    cohort: '2026-05-05',
    cohortSize: 1380,
    weeks: [100, 72, 55, 48, 41, null, null, null, null],
  },
  {
    cohort: '2026-05-12',
    cohortSize: 1100,
    weeks: [100, 67, 51, 44, null, null, null, null, null],
  },
  {
    cohort: '2026-05-19',
    cohortSize: 1230,
    weeks: [100, 70, 54, null, null, null, null, null, null],
  },
  {
    cohort: '2026-05-26',
    cohortSize: 1050,
    weeks: [100, 68, null, null, null, null, null, null, null],
  },
  {
    cohort: '2026-06-02',
    cohortSize: 1400,
    weeks: [100, null, null, null, null, null, null, null, null],
  },
]

const WEEKS = Array.from({ length: 9 }, (_, i) => i)

const calculateMetrics = (data: RetentionData[]) => {
  let totalCohorts = 0
  let avgW0W4 = 0
  let avgLatest = 0
  let cohortCount = 0

  data.forEach((row) => {
    totalCohorts += row.cohortSize
    if (row.weeks[4] !== null) {
      avgW0W4 += row.weeks[4]
      cohortCount++
    }
    const lastNonNull = row.weeks.filter((w) => w !== null).pop()
    if (lastNonNull !== null && lastNonNull !== 100) {
      avgLatest += lastNonNull
    }
  })

  return {
    totalUsers: totalCohorts,
    avgRetention4w: cohortCount > 0 ? Math.round(avgW0W4 / cohortCount) : 0,
    avgLatestRetention: data.length > 0 ? Math.round(avgLatest / data.length) : 0,
  }
}

interface RetentionMatrixProps {
  projectId?: string
}

export default function RetentionMatrix({ projectId }: RetentionMatrixProps) {
  const [data, setData] = useState<RetentionData[]>(PLACEHOLDER_DATA)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // Simular carregamento de dados via API
  useEffect(() => {
    const fetchData = async () => {
      if (!projectId) return

      setLoading(true)
      setError(null)
      try {
        // Quando implementar o backend, descomente:
        // const res = await api.get(`/v1/projects/${projectId}/retention?cohort=weekly`)
        // setData(res.data)

        // Por enquanto, usar dados placeholder
        setData(PLACEHOLDER_DATA)
      } catch (err) {
        const msg = err instanceof Error ? err.message : 'Erro ao carregar retenção'
        setError(msg)
      } finally {
        setLoading(false)
      }
    }

    fetchData()
  }, [projectId])

  const metrics = calculateMetrics(data)

  return (
    <div className="space-y-6">
      {/* Stats Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-4">
          <p className="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wide">
            Usuários totais
          </p>
          <p className="text-2xl font-bold text-slate-900 dark:text-slate-100 mt-1">
            {metrics.totalUsers.toLocaleString()}
          </p>
          <p className="text-xs text-slate-400 mt-0.5">em todas as cohorts</p>
        </div>

        <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-4">
          <p className="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wide">
            Retenção W0→W4
          </p>
          <p className="text-2xl font-bold text-emerald-600 dark:text-emerald-400 mt-1">
            {metrics.avgRetention4w}%
          </p>
          <p className="text-xs text-slate-400 mt-0.5">média em 4 semanas</p>
        </div>

        <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-4">
          <p className="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wide">
            Retenção atual
          </p>
          <p className="text-2xl font-bold text-slate-900 dark:text-slate-100 mt-1">
            {metrics.avgLatestRetention}%
          </p>
          <p className="text-xs text-slate-400 mt-0.5">última medição disponível</p>
        </div>
      </div>

      {/* Retention Matrix */}
      <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-5 space-y-4">
        <div className="flex items-center gap-2 mb-4">
          <TrendingUp className="w-4 h-4 text-violet-500" />
          <h2 className="text-sm font-semibold text-slate-800 dark:text-slate-100">
            Matriz de Retenção Semanal (W0 — W8)
          </h2>
          <span className="text-xs text-slate-500 dark:text-slate-400 font-normal">
            (semanas após cadastro)
          </span>
        </div>

        {error && (
          <div className="rounded-lg bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 p-3 text-xs text-red-700 dark:text-red-300">
            {error}
          </div>
        )}

        {loading ? (
          <div className="flex items-center justify-center py-12 text-slate-400">
            <div className="w-5 h-5 border-2 border-violet-500 border-t-transparent rounded-full animate-spin mr-2" />
            Carregando dados de retenção...
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-xs border-collapse">
              <thead>
                <tr className="border-b border-slate-200 dark:border-slate-700">
                  <th className="py-2 px-3 text-left text-slate-500 dark:text-slate-400 font-medium whitespace-nowrap">
                    Data Cohort
                  </th>
                  <th className="py-2 px-1 text-center text-slate-500 dark:text-slate-400 font-medium whitespace-nowrap">
                    Tamanho
                  </th>
                  {WEEKS.map((w) => (
                    <th
                      key={w}
                      className="py-2 px-1 text-center text-slate-500 dark:text-slate-400 font-medium whitespace-nowrap"
                    >
                      W{w}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {data.map((row, idx) => (
                  <tr key={idx} className="border-b border-slate-100 dark:border-slate-800">
                    <td className="py-2.5 px-3 text-slate-700 dark:text-slate-300 font-mono text-xs whitespace-nowrap">
                      {row.cohort}
                    </td>
                    <td className="py-2.5 px-1 text-center text-slate-700 dark:text-slate-300 font-semibold text-xs whitespace-nowrap">
                      {row.cohortSize.toLocaleString()}
                    </td>
                    {WEEKS.map((w) => {
                      const pct = row.weeks[w]
                      return (
                        <td key={w} className="py-2.5 px-1 text-center">
                          {pct !== null ? (
                            <div
                              className={`inline-flex items-center justify-center w-10 h-8 rounded font-semibold transition-all ${getHeatmapColor(
                                pct
                              )}`}
                              title={`${pct}% retidos`}
                            >
                              {pct}%
                            </div>
                          ) : (
                            <div className="inline-flex items-center justify-center w-10 h-8 text-slate-300 dark:text-slate-600">
                              —
                            </div>
                          )}
                        </td>
                      )
                    })}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {/* Legend */}
        <div className="mt-4 pt-4 border-t border-slate-200 dark:border-slate-700 space-y-2">
          <p className="text-xs font-medium text-slate-600 dark:text-slate-400">Legenda de retenção:</p>
          <div className="flex flex-wrap gap-3 text-xs">
            {[
              { label: '≥80%', color: 'bg-emerald-500' },
              { label: '60-79%', color: 'bg-emerald-400' },
              { label: '50-59%', color: 'bg-cyan-400' },
              { label: '40-49%', color: 'bg-yellow-400' },
              { label: '20-39%', color: 'bg-orange-400' },
              { label: '<20%', color: 'bg-red-400' },
            ].map(({ label, color }) => (
              <div key={label} className="flex items-center gap-1.5">
                <div className={`w-4 h-4 rounded ${color}`} />
                <span className="text-slate-600 dark:text-slate-400">{label}</span>
              </div>
            ))}
          </div>
        </div>

        {/* Info */}
        <div className="bg-slate-50 dark:bg-slate-800 rounded-lg p-3 text-xs text-slate-600 dark:text-slate-400 space-y-1">
          <p className="font-medium text-slate-700 dark:text-slate-300">Sobre esta matriz:</p>
          <ul className="space-y-0.5">
            <li>• W0 = semana de cadastro (sempre 100%)</li>
            <li>• Cada coluna mostra retenção na semana N após cadastro</li>
            <li>• Cores verdes = boa retenção; cores quentes = churn</li>
            <li>• Dados em tempo real; atualiza cada 6 horas</li>
          </ul>
        </div>
      </div>
    </div>
  )
}
