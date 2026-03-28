import { useCallback, useEffect, useState } from 'react'
import { Flag, Activity, TrendingUp, RefreshCw } from 'lucide-react'
import toast from 'react-hot-toast'
import {
  ResponsiveContainer,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
} from 'recharts'
import { flagsApi, eventsApi } from '@/lib/api'
import type { FeatureFlag } from '@/lib/api'

interface DayCount {
  date: string
  label: string
  count: number
}

function getLast7Days(): { date: string; label: string }[] {
  const days: { date: string; label: string }[] = []
  const now = new Date()
  for (let i = 6; i >= 0; i--) {
    const d = new Date(now)
    d.setDate(d.getDate() - i)
    const iso = d.toISOString().split('T')[0]
    const label = d.toLocaleDateString('pt-BR', { day: '2-digit', month: '2-digit' })
    days.push({ date: iso, label })
  }
  return days
}

export default function Overview() {
  const [flags, setFlags] = useState<FeatureFlag[]>([])
  const [totalEvents, setTotalEvents] = useState(0)
  const [chartData, setChartData] = useState<DayCount[]>([])
  const [loading, setLoading] = useState(true)

  const fetchData = useCallback(async () => {
    setLoading(true)
    try {
      const days = getLast7Days()
      const startDate = days[0].date
      const endDate = days[days.length - 1].date

      const [flagList, countResp] = await Promise.all([
        flagsApi.list(),
        eventsApi.count({ start_date: startDate, end_date: endDate }),
      ])

      setFlags(flagList)
      setTotalEvents(countResp.total)

      // Build per-day counts by fetching individual day counts
      // The count endpoint groups by event_name, so we sum totals per day
      const dayPromises = days.map(async (day) => {
        try {
          const resp = await eventsApi.count({
            start_date: day.date,
            end_date: day.date,
          })
          return { ...day, count: resp.total }
        } catch {
          return { ...day, count: 0 }
        }
      })

      const dayResults = await Promise.all(dayPromises)
      setChartData(dayResults)
    } catch (err) {
      toast.error('Falha ao carregar dados do overview')
      console.error(err)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void fetchData()
  }, [fetchData])

  const activeFlags = flags.filter((f) => f.enabled).length

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-slate-900 dark:text-slate-100">Overview</h1>
          <p className="text-sm text-slate-500 dark:text-slate-400 mt-0.5">
            Visao geral do Pulse Service
          </p>
        </div>
        <button
          onClick={fetchData}
          disabled={loading}
          className="p-2 rounded-lg text-slate-500 hover:bg-slate-100 dark:text-slate-400 dark:hover:bg-slate-800 transition-colors disabled:opacity-50"
          aria-label="Recarregar"
        >
          <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
        </button>
      </div>

      {/* Metric cards */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-5">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-violet-100 dark:bg-violet-900/40">
              <Flag className="w-5 h-5 text-violet-600 dark:text-violet-400" />
            </div>
            <div>
              <p className="text-xs font-medium text-slate-500 dark:text-slate-400">Flags ativas</p>
              <p className="text-2xl font-bold text-slate-800 dark:text-slate-200">
                {loading ? '--' : `${activeFlags} / ${flags.length}`}
              </p>
            </div>
          </div>
        </div>

        <div className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-5">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-emerald-100 dark:bg-emerald-900/40">
              <Activity className="w-5 h-5 text-emerald-600 dark:text-emerald-400" />
            </div>
            <div>
              <p className="text-xs font-medium text-slate-500 dark:text-slate-400">Total de eventos (7d)</p>
              <p className="text-2xl font-bold text-slate-800 dark:text-slate-200">
                {loading ? '--' : totalEvents.toLocaleString('pt-BR')}
              </p>
            </div>
          </div>
        </div>

        <div className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-5">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-blue-100 dark:bg-blue-900/40">
              <TrendingUp className="w-5 h-5 text-blue-600 dark:text-blue-400" />
            </div>
            <div>
              <p className="text-xs font-medium text-slate-500 dark:text-slate-400">Media/dia (7d)</p>
              <p className="text-2xl font-bold text-slate-800 dark:text-slate-200">
                {loading ? '--' : Math.round(totalEvents / 7).toLocaleString('pt-BR')}
              </p>
            </div>
          </div>
        </div>
      </div>

      {/* Chart */}
      <div className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-5">
        <h2 className="text-sm font-semibold text-slate-700 dark:text-slate-300 mb-4">
          Eventos por dia (ultimos 7 dias)
        </h2>
        {loading ? (
          <div className="flex items-center justify-center h-64 text-slate-400">
            <div className="w-6 h-6 border-2 border-violet-500 border-t-transparent rounded-full animate-spin mr-3" />
            <span className="text-sm">Carregando...</span>
          </div>
        ) : (
          <ResponsiveContainer width="100%" height={280}>
            <BarChart data={chartData} margin={{ top: 5, right: 20, left: 0, bottom: 5 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
              <XAxis
                dataKey="label"
                tick={{ fontSize: 12, fill: '#94a3b8' }}
                axisLine={{ stroke: '#e2e8f0' }}
                tickLine={false}
              />
              <YAxis
                tick={{ fontSize: 12, fill: '#94a3b8' }}
                axisLine={{ stroke: '#e2e8f0' }}
                tickLine={false}
                allowDecimals={false}
              />
              <Tooltip
                contentStyle={{
                  background: '#1e293b',
                  border: 'none',
                  borderRadius: '8px',
                  color: '#f1f5f9',
                  fontSize: '13px',
                }}
                labelStyle={{ color: '#94a3b8' }}
                formatter={(value: number) => [value.toLocaleString('pt-BR'), 'Eventos']}
              />
              <Bar dataKey="count" fill="#7C3AED" radius={[4, 4, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        )}
      </div>
    </div>
  )
}
