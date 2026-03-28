import { useCallback, useEffect, useState } from 'react'
import { RefreshCw, Users, TrendingUp } from 'lucide-react'
import toast from 'react-hot-toast'
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts'
import { retentionApi } from '@/lib/api'
import type { ActiveUsersItem, CohortItem } from '@/lib/api'

const PRODUCTS = ['', 'libri', 'nitro', 'brio']

function formatDate30DaysAgo(): string {
  const d = new Date()
  d.setDate(d.getDate() - 30)
  return d.toISOString().split('T')[0]
}

function formatDateToday(): string {
  return new Date().toISOString().split('T')[0]
}

const inputClass =
  'px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 text-sm focus:outline-none focus:ring-2 focus:ring-violet-500 focus:border-transparent'

export default function Retention() {
  const [product, setProduct] = useState('')
  const [startDate, setStartDate] = useState(formatDate30DaysAgo)
  const [endDate, setEndDate] = useState(formatDateToday)
  const [granularity, setGranularity] = useState<'week' | 'month'>('week')

  const [retentionData, setRetentionData] = useState<ActiveUsersItem[]>([])
  const [cohortData, setCohortData] = useState<CohortItem[]>([])
  const [loading, setLoading] = useState(true)

  const fetchData = useCallback(async () => {
    setLoading(true)
    try {
      const [retention, cohorts] = await Promise.all([
        retentionApi.retention({
          product: product || undefined,
          start_date: startDate,
          end_date: endDate,
        }),
        retentionApi.cohorts({
          product: product || undefined,
          granularity,
        }),
      ])
      setRetentionData(retention.items ?? [])
      setCohortData(cohorts.cohorts ?? [])
    } catch (err) {
      toast.error('Falha ao carregar dados de retencao')
      console.error(err)
    } finally {
      setLoading(false)
    }
  }, [product, startDate, endDate, granularity])

  useEffect(() => {
    void fetchData()
  }, [fetchData])

  // Latest DAU/WAU/MAU
  const latest = retentionData.length > 0 ? retentionData[retentionData.length - 1] : null

  return (
    <div className="space-y-5">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-xl font-bold text-slate-900 dark:text-slate-100">
            Retencao
          </h1>
          <p className="text-sm text-slate-500 dark:text-slate-400 mt-0.5">
            DAU / WAU / MAU e analise de coortes
          </p>
        </div>
        <button
          onClick={fetchData}
          disabled={loading}
          className="p-2 rounded-lg text-slate-500 hover:bg-slate-100 dark:text-slate-400 dark:hover:bg-slate-800 transition-colors disabled:opacity-50 self-start sm:self-auto"
          aria-label="Recarregar"
        >
          <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
        </button>
      </div>

      {/* Filters */}
      <div className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-4">
        <div className="grid grid-cols-1 sm:grid-cols-5 gap-3">
          <div>
            <label className="block text-xs font-medium text-slate-500 dark:text-slate-400 mb-1">
              Produto
            </label>
            <select
              value={product}
              onChange={(e) => setProduct(e.target.value)}
              className={`w-full ${inputClass}`}
            >
              <option value="">Todos</option>
              {PRODUCTS.filter(Boolean).map((p) => (
                <option key={p} value={p}>
                  {p.charAt(0).toUpperCase() + p.slice(1)}
                </option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-xs font-medium text-slate-500 dark:text-slate-400 mb-1">
              Data inicio
            </label>
            <input
              type="date"
              value={startDate}
              onChange={(e) => setStartDate(e.target.value)}
              className={`w-full ${inputClass}`}
            />
          </div>
          <div>
            <label className="block text-xs font-medium text-slate-500 dark:text-slate-400 mb-1">
              Data fim
            </label>
            <input
              type="date"
              value={endDate}
              onChange={(e) => setEndDate(e.target.value)}
              className={`w-full ${inputClass}`}
            />
          </div>
          <div>
            <label className="block text-xs font-medium text-slate-500 dark:text-slate-400 mb-1">
              Granularidade coortes
            </label>
            <select
              value={granularity}
              onChange={(e) => setGranularity(e.target.value as 'week' | 'month')}
              className={`w-full ${inputClass}`}
            >
              <option value="week">Semana</option>
              <option value="month">Mes</option>
            </select>
          </div>
          <div className="flex items-end">
            <button
              onClick={fetchData}
              className="px-4 py-2 rounded-lg bg-violet-600 hover:bg-violet-700 text-white text-sm font-medium transition-colors"
            >
              Atualizar
            </button>
          </div>
        </div>
      </div>

      {loading ? (
        <div className="flex items-center justify-center py-16 text-slate-400">
          <div className="w-6 h-6 border-2 border-violet-500 border-t-transparent rounded-full animate-spin mr-3" />
          <span className="text-sm">Carregando metricas...</span>
        </div>
      ) : (
        <>
          {/* KPI cards */}
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <KpiCard label="DAU (hoje)" value={latest?.dau ?? 0} icon={<Users className="w-5 h-5" />} />
            <KpiCard label="WAU (7d)" value={latest?.wau ?? 0} icon={<Users className="w-5 h-5" />} />
            <KpiCard label="MAU (30d)" value={latest?.mau ?? 0} icon={<TrendingUp className="w-5 h-5" />} />
          </div>

          {/* Retention chart */}
          {retentionData.length > 0 && (
            <div className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-4">
              <h2 className="text-sm font-semibold text-slate-700 dark:text-slate-300 mb-4">
                DAU / WAU / MAU ao longo do tempo
              </h2>
              <ResponsiveContainer width="100%" height={320}>
                <LineChart data={retentionData}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#374151" opacity={0.2} />
                  <XAxis
                    dataKey="date"
                    tick={{ fontSize: 11 }}
                    stroke="#94a3b8"
                    tickFormatter={(v: string) => v.slice(5)}
                  />
                  <YAxis tick={{ fontSize: 11 }} stroke="#94a3b8" />
                  <Tooltip
                    contentStyle={{
                      backgroundColor: '#1e293b',
                      border: 'none',
                      borderRadius: 8,
                      color: '#e2e8f0',
                      fontSize: 12,
                    }}
                  />
                  <Legend />
                  <Line type="monotone" dataKey="dau" name="DAU" stroke="#8b5cf6" strokeWidth={2} dot={false} />
                  <Line type="monotone" dataKey="wau" name="WAU" stroke="#06b6d4" strokeWidth={2} dot={false} />
                  <Line type="monotone" dataKey="mau" name="MAU" stroke="#f59e0b" strokeWidth={2} dot={false} />
                </LineChart>
              </ResponsiveContainer>
            </div>
          )}

          {/* Cohort table */}
          {cohortData.length > 0 && (
            <div className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-4">
              <h2 className="text-sm font-semibold text-slate-700 dark:text-slate-300 mb-4">
                Analise de coortes ({granularity === 'week' ? 'semanal' : 'mensal'})
              </h2>
              <div className="overflow-x-auto">
                <table className="w-full text-xs">
                  <thead>
                    <tr className="bg-slate-50 dark:bg-slate-800/50 text-left font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide">
                      <th className="px-3 py-2">Coorte</th>
                      <th className="px-3 py-2">Tamanho</th>
                      {maxPeriods(cohortData).map((p) => (
                        <th key={p} className="px-3 py-2 text-center">
                          {granularity === 'week' ? `S${p}` : `M${p}`}
                        </th>
                      ))}
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-100 dark:divide-slate-800">
                    {cohortData.map((cohort) => (
                      <tr key={cohort.cohort_label} className="bg-white dark:bg-slate-900 hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-colors">
                        <td className="px-3 py-2 font-mono text-slate-700 dark:text-slate-300 whitespace-nowrap">
                          {cohort.cohort_label}
                        </td>
                        <td className="px-3 py-2 text-slate-600 dark:text-slate-400 text-center">
                          {cohort.cohort_size}
                        </td>
                        {maxPeriods(cohortData).map((p) => {
                          const bucket = cohort.buckets.find((b) => b.period === p)
                          const pct = bucket?.retention_pct ?? 0
                          return (
                            <td key={p} className="px-3 py-2 text-center">
                              <span
                                className="inline-block px-2 py-0.5 rounded text-xs font-medium"
                                style={{
                                  backgroundColor: `rgba(139, 92, 246, ${Math.min(pct / 100, 1) * 0.5 + 0.05})`,
                                  color: pct > 50 ? '#fff' : '#c4b5fd',
                                }}
                              >
                                {pct > 0 ? `${pct.toFixed(0)}%` : '--'}
                              </span>
                            </td>
                          )
                        })}
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {retentionData.length === 0 && cohortData.length === 0 && (
            <div className="flex flex-col items-center justify-center py-16 text-slate-400 gap-2">
              <p className="text-sm">Nenhum dado de retencao encontrado</p>
              <p className="text-xs">Ajuste os filtros ou aguarde novos eventos com user_id.</p>
            </div>
          )}
        </>
      )}
    </div>
  )
}

function KpiCard({ label, value, icon }: { label: string; value: number; icon: React.ReactNode }) {
  return (
    <div className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-4 flex items-center gap-4">
      <div className="w-10 h-10 rounded-lg bg-violet-100 dark:bg-violet-900/30 flex items-center justify-center text-violet-600 dark:text-violet-400">
        {icon}
      </div>
      <div>
        <p className="text-2xl font-bold text-slate-900 dark:text-slate-100">
          {value.toLocaleString('pt-BR')}
        </p>
        <p className="text-xs text-slate-500 dark:text-slate-400">{label}</p>
      </div>
    </div>
  )
}

function maxPeriods(cohorts: CohortItem[]): number[] {
  let max = 0
  for (const c of cohorts) {
    for (const b of c.buckets) {
      if (b.period > max) max = b.period
    }
  }
  return Array.from({ length: max + 1 }, (_, i) => i)
}
