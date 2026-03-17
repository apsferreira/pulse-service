import { useCallback, useEffect, useState } from 'react'
import { Activity, RefreshCw } from 'lucide-react'
import toast from 'react-hot-toast'
import { eventsApi, type EventSummaryItem, type EventsSummaryResponse } from '../services/pulseApi'

const PRODUCT_COLORS: Record<string, string> = {
  libri:  'from-blue-50 to-blue-100 dark:from-blue-950/40 dark:to-blue-900/20 border-blue-200 dark:border-blue-800 text-blue-700 dark:text-blue-300',
  nitro:  'from-orange-50 to-orange-100 dark:from-orange-950/40 dark:to-orange-900/20 border-orange-200 dark:border-orange-800 text-orange-700 dark:text-orange-300',
  brio:   'from-emerald-50 to-emerald-100 dark:from-emerald-950/40 dark:to-emerald-900/20 border-emerald-200 dark:border-emerald-800 text-emerald-700 dark:text-emerald-300',
}

const defaultColor =
  'from-slate-50 to-slate-100 dark:from-slate-800/40 dark:to-slate-700/20 border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-300'

function getProductColor(product: string): string {
  return PRODUCT_COLORS[product] ?? defaultColor
}

function groupByProduct(items: EventSummaryItem[]): Record<string, EventSummaryItem[]> {
  return items.reduce<Record<string, EventSummaryItem[]>>((acc, item) => {
    if (!acc[item.product]) acc[item.product] = []
    acc[item.product].push(item)
    return acc
  }, {})
}

export default function Events() {
  const [summary, setSummary] = useState<EventsSummaryResponse | null>(null)
  const [loading, setLoading] = useState(true)

  const fetchSummary = useCallback(async () => {
    setLoading(true)
    try {
      const { data } = await eventsApi.summary()
      setSummary(data)
    } catch (err) {
      toast.error('Falha ao carregar resumo de eventos')
      console.error(err)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void fetchSummary()
    const interval = setInterval(() => void fetchSummary(), 30_000)
    return () => clearInterval(interval)
  }, [fetchSummary])

  const grouped = summary ? groupByProduct(summary.items) : {}
  const products = Object.keys(grouped).sort()

  return (
    <div className="space-y-5">
      {/* Page header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-xl font-bold text-slate-900 dark:text-slate-100">Eventos</h1>
          <p className="text-sm text-slate-500 dark:text-slate-400 mt-0.5">
            Rastreamento de eventos — últimas 24h
          </p>
        </div>
        <button
          onClick={fetchSummary}
          disabled={loading}
          className="p-2 rounded-lg text-slate-500 hover:bg-slate-100 dark:text-slate-400 dark:hover:bg-slate-800 transition-colors disabled:opacity-50 self-start sm:self-auto"
          aria-label="Recarregar"
        >
          <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
        </button>
      </div>

      {/* Summary strip */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
        <div className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-4">
          <p className="text-xs text-slate-500 dark:text-slate-400 font-medium">Total de eventos</p>
          <p className="text-3xl font-bold text-slate-800 dark:text-slate-200 mt-1">
            {loading ? '—' : (summary?.total_last_24h ?? 0).toLocaleString('pt-BR')}
          </p>
          <p className="text-xs text-slate-400 mt-0.5">últimas 24h</p>
        </div>

        <div className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-4">
          <p className="text-xs text-slate-500 dark:text-slate-400 font-medium">Produtos ativos</p>
          <p className="text-3xl font-bold text-slate-800 dark:text-slate-200 mt-1">
            {loading ? '—' : products.length}
          </p>
          <p className="text-xs text-slate-400 mt-0.5">com eventos nas últimas 24h</p>
        </div>

        <div className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-4">
          <p className="text-xs text-slate-500 dark:text-slate-400 font-medium">Tipos de evento</p>
          <p className="text-3xl font-bold text-slate-800 dark:text-slate-200 mt-1">
            {loading ? '—' : summary ? new Set(summary.items.map((i) => i.event_type)).size : 0}
          </p>
          <p className="text-xs text-slate-400 mt-0.5">distintos nas últimas 24h</p>
        </div>
      </div>

      {/* Empty state */}
      {!loading && (!summary || summary.items.length === 0) && (
        <div className="rounded-xl border border-dashed border-violet-300 dark:border-violet-700 bg-violet-50/50 dark:bg-violet-950/20 p-12 flex flex-col items-center text-center gap-4">
          <div className="w-16 h-16 rounded-2xl bg-violet-100 dark:bg-violet-900/40 flex items-center justify-center">
            <Activity className="w-8 h-8 text-violet-500 dark:text-violet-400" />
          </div>
          <div className="space-y-1">
            <h2 className="text-base font-semibold text-slate-800 dark:text-slate-200">
              Nenhum evento registado nas últimas 24h
            </h2>
            <p className="text-sm text-slate-500 dark:text-slate-400 max-w-md">
              Os eventos aparecerão aqui quando as aplicações Libri, Nitro e Brio começarem a enviar dados para
              o endpoint <code className="font-mono text-xs bg-slate-100 dark:bg-slate-800 px-1 py-0.5 rounded">POST /api/v1/events</code>.
            </p>
          </div>
        </div>
      )}

      {/* Per-product breakdown */}
      {products.length > 0 && (
        <div className="space-y-4">
          {products.map((product) => {
            const items = grouped[product]
            const total = items.reduce((sum, i) => sum + i.count, 0)
            const color = getProductColor(product)

            return (
              <div
                key={product}
                className={`rounded-xl border bg-gradient-to-br ${color} p-4 space-y-3`}
              >
                <div className="flex items-center justify-between">
                  <h3 className="font-semibold capitalize">{product}</h3>
                  <span className="text-sm font-medium opacity-70">
                    {total.toLocaleString('pt-BR')} eventos
                  </span>
                </div>

                <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-2">
                  {items.map((item) => (
                    <div
                      key={`${item.product}-${item.event_type}`}
                      className="rounded-lg bg-white/60 dark:bg-slate-900/40 px-3 py-2"
                    >
                      <p className="text-xs text-slate-500 dark:text-slate-400 font-mono truncate">
                        {item.event_type}
                      </p>
                      <p className="text-lg font-bold text-slate-800 dark:text-slate-200">
                        {item.count.toLocaleString('pt-BR')}
                      </p>
                    </div>
                  ))}
                </div>
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}
