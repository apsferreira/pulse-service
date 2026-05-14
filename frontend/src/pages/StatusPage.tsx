// BKL-1068: Status page pública — uptime dos serviços IIT sem autenticação.
// Acesso: /status (rota pública registrada no App.tsx)
import { useEffect, useState } from 'react'
import { CheckCircle2, XCircle, Clock, AlertCircle, RefreshCw } from 'lucide-react'
import type { MonitorStatus } from '../services/pulseApi'

const API_BASE = import.meta.env.VITE_API_URL ?? '/api/v1'

function fetchPublic<T>(path: string): Promise<T> {
  return fetch(`${API_BASE}${path}`).then(r => {
    if (!r.ok) throw new Error(`HTTP ${r.status}`)
    return r.json() as Promise<T>
  })
}

function StatusDot({ status }: { status?: string }) {
  if (status === 'up') {
    return <span className="inline-block w-3 h-3 rounded-full bg-emerald-500 ring-4 ring-emerald-100 animate-pulse" />
  }
  if (status === 'down') {
    return <span className="inline-block w-3 h-3 rounded-full bg-red-500 ring-4 ring-red-100" />
  }
  if (status === 'timeout') {
    return <span className="inline-block w-3 h-3 rounded-full bg-amber-400 ring-4 ring-amber-100" />
  }
  return <span className="inline-block w-3 h-3 rounded-full bg-slate-300 ring-4 ring-slate-100" />
}

function StatusLabel({ status }: { status?: string }) {
  if (status === 'up') return <span className="text-emerald-600 font-semibold text-sm">Operacional</span>
  if (status === 'down') return <span className="text-red-600 font-semibold text-sm">Indisponível</span>
  if (status === 'timeout') return <span className="text-amber-600 font-semibold text-sm">Degradado</span>
  return <span className="text-slate-400 text-sm">Sem dados</span>
}

function UptimeBar({ uptime }: { uptime: number }) {
  const pct = Math.min(100, Math.max(0, uptime))
  const color =
    pct >= 99 ? 'bg-emerald-500' :
    pct >= 95 ? 'bg-amber-400' :
    'bg-red-500'
  return (
    <div className="flex items-center gap-2 text-xs text-slate-500">
      <div className="flex-1 bg-slate-100 rounded-full h-1.5 overflow-hidden">
        <div className={`h-full rounded-full ${color}`} style={{ width: `${pct}%` }} />
      </div>
      <span className="w-12 text-right tabular-nums">{pct.toFixed(1)}%</span>
    </div>
  )
}

function OverallBanner({ statuses }: { statuses: MonitorStatus[] }) {
  if (statuses.length === 0) return null
  const allUp = statuses.every(s => s.last_check?.status === 'up')
  const anyDown = statuses.some(s => s.last_check?.status === 'down')

  if (allUp) {
    return (
      <div className="flex items-center gap-3 bg-emerald-50 border border-emerald-200 rounded-xl px-5 py-4 mb-8">
        <CheckCircle2 className="w-6 h-6 text-emerald-600 shrink-0" />
        <div>
          <p className="font-semibold text-emerald-800">Todos os sistemas operacionais</p>
          <p className="text-sm text-emerald-600">Nenhum incidente ativo.</p>
        </div>
      </div>
    )
  }
  if (anyDown) {
    return (
      <div className="flex items-center gap-3 bg-red-50 border border-red-200 rounded-xl px-5 py-4 mb-8">
        <XCircle className="w-6 h-6 text-red-600 shrink-0" />
        <div>
          <p className="font-semibold text-red-800">Incidente em andamento</p>
          <p className="text-sm text-red-600">Um ou mais serviços estão indisponíveis. Trabalhando para resolver.</p>
        </div>
      </div>
    )
  }
  return (
    <div className="flex items-center gap-3 bg-amber-50 border border-amber-200 rounded-xl px-5 py-4 mb-8">
      <AlertCircle className="w-6 h-6 text-amber-600 shrink-0" />
      <div>
        <p className="font-semibold text-amber-800">Degradação parcial</p>
        <p className="text-sm text-amber-600">Alguns serviços podem estar lentos. Monitorando.</p>
      </div>
    </div>
  )
}

export default function StatusPage() {
  const [statuses, setStatuses] = useState<MonitorStatus[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null)

  const load = () => {
    setLoading(true)
    setError(null)
    fetchPublic<MonitorStatus[]>('/monitors')
      .then(data => {
        setStatuses(data ?? [])
        setLastUpdated(new Date())
      })
      .catch(() => setError('Não foi possível carregar os dados de status.'))
      .finally(() => setLoading(false))
  }

  useEffect(() => {
    load()
    // Auto-refresh a cada 60s
    const id = setInterval(load, 60_000)
    return () => clearInterval(id)
  }, [])

  return (
    <div className="min-h-screen bg-[#F5F8FA] font-sans">
      {/* Header */}
      <header className="bg-white border-b border-slate-200 py-5 px-4">
        <div className="max-w-3xl mx-auto flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-8 h-8 rounded-lg bg-[#0097D6] flex items-center justify-center text-white font-bold text-sm">IIT</div>
            <div>
              <h1 className="text-base font-bold text-[#0D1B26]">Instituto Itinerante de Tecnologia</h1>
              <p className="text-xs text-slate-500">Status dos Serviços</p>
            </div>
          </div>
          <button
            onClick={load}
            disabled={loading}
            className="flex items-center gap-1.5 text-xs text-slate-500 hover:text-[#0097D6] transition-colors disabled:opacity-50"
            title="Atualizar"
          >
            <RefreshCw className={`w-3.5 h-3.5 ${loading ? 'animate-spin' : ''}`} />
            {lastUpdated && (
              <span>atualizado às {lastUpdated.toLocaleTimeString('pt-BR', { hour: '2-digit', minute: '2-digit' })}</span>
            )}
          </button>
        </div>
      </header>

      <main className="max-w-3xl mx-auto px-4 py-10">
        {/* Overall banner */}
        {!loading && !error && <OverallBanner statuses={statuses} />}

        {/* Error */}
        {error && (
          <div className="bg-red-50 border border-red-200 rounded-xl px-5 py-4 mb-8 text-red-700 text-sm">{error}</div>
        )}

        {/* Loading skeleton */}
        {loading && statuses.length === 0 && (
          <div className="space-y-3">
            {[1, 2, 3, 4, 5].map(i => (
              <div key={i} className="bg-white rounded-xl p-5 animate-pulse">
                <div className="flex items-center gap-3 mb-3">
                  <div className="w-3 h-3 rounded-full bg-slate-200" />
                  <div className="h-4 bg-slate-200 rounded w-40" />
                  <div className="ml-auto h-3 bg-slate-100 rounded w-20" />
                </div>
                <div className="h-1.5 bg-slate-100 rounded-full w-full" />
              </div>
            ))}
          </div>
        )}

        {/* Services list */}
        {!loading && statuses.length > 0 && (
          <div className="space-y-3">
            <h2 className="text-xs font-semibold uppercase tracking-widest text-slate-400 mb-4">Serviços</h2>
            {statuses.map(s => (
              <div key={s.monitor.id} className="bg-white rounded-xl p-5 shadow-sm border border-slate-100">
                <div className="flex items-center gap-3 mb-3">
                  <StatusDot status={s.last_check?.status} />
                  <span className="font-medium text-[#0D1B26] text-sm">{s.monitor.name}</span>
                  <div className="ml-auto flex items-center gap-3">
                    {s.avg_response_ms > 0 && (
                      <span className="text-xs text-slate-400 flex items-center gap-1">
                        <Clock className="w-3 h-3" />
                        {s.avg_response_ms.toFixed(0)} ms
                      </span>
                    )}
                    <StatusLabel status={s.last_check?.status} />
                  </div>
                </div>
                <UptimeBar uptime={s.uptime_24h} />
                <p className="text-right text-xs text-slate-300 mt-1">últimas 24h</p>
              </div>
            ))}
          </div>
        )}

        {!loading && statuses.length === 0 && !error && (
          <div className="text-center py-16 text-slate-400 text-sm">
            Nenhum monitor configurado.
          </div>
        )}
      </main>

      <footer className="text-center text-xs text-slate-400 py-8">
        <p>Instituto Itinerante de Tecnologia · Salvador, BA</p>
        <p className="mt-1">Dados atualizados a cada 60 segundos</p>
      </footer>
    </div>
  )
}
