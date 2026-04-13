import { useCallback, useEffect, useState } from 'react'
import {
  Video,
  RefreshCw,
  X,
  MousePointer,
  Scroll,
  Monitor,
  Code,
  Layers,
  ChevronRight,
  Search,
} from 'lucide-react'
import toast from 'react-hot-toast'
import { replayApi, type ReplaySession, type ReplayEvent, type ReplayEventType } from '../services/pulseApi'

// ── Helpers ──────────────────────────────────────────────────────────────────

function formatDuration(ms: number): string {
  const totalSeconds = Math.floor(ms / 1000)
  const minutes = Math.floor(totalSeconds / 60)
  const seconds = totalSeconds % 60
  if (minutes === 0) return `${seconds}s`
  return `${minutes}m ${seconds}s`
}

function formatDate(iso: string): string {
  return new Intl.DateTimeFormat('pt-BR', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(iso))
}

function truncateUrl(url: string, max = 48): string {
  try {
    const u = new URL(url)
    const path = u.pathname + u.search
    if (path.length <= max) return path
    return path.slice(0, max) + '…'
  } catch {
    return url.length <= max ? url : url.slice(0, max) + '…'
  }
}

// ── Event type config ────────────────────────────────────────────────────────

interface EventTypeConfig {
  label: string
  color: string
  dotColor: string
  Icon: React.ElementType
}

const EVENT_TYPE_CONFIG: Record<ReplayEventType, EventTypeConfig> = {
  click: {
    label: 'Click',
    color: 'text-emerald-600 dark:text-emerald-400 bg-emerald-50 dark:bg-emerald-950/40 border-emerald-200 dark:border-emerald-800',
    dotColor: 'bg-emerald-500',
    Icon: MousePointer,
  },
  scroll: {
    label: 'Scroll',
    color: 'text-blue-600 dark:text-blue-400 bg-blue-50 dark:bg-blue-950/40 border-blue-200 dark:border-blue-800',
    dotColor: 'bg-blue-500',
    Icon: Scroll,
  },
  viewport: {
    label: 'Viewport',
    color: 'text-violet-600 dark:text-violet-400 bg-violet-50 dark:bg-violet-950/40 border-violet-200 dark:border-violet-800',
    dotColor: 'bg-violet-500',
    Icon: Monitor,
  },
  dom_mutation: {
    label: 'DOM',
    color: 'text-amber-600 dark:text-amber-400 bg-amber-50 dark:bg-amber-950/40 border-amber-200 dark:border-amber-800',
    dotColor: 'bg-amber-500',
    Icon: Layers,
  },
  custom: {
    label: 'Custom',
    color: 'text-slate-600 dark:text-slate-400 bg-slate-50 dark:bg-slate-800 border-slate-200 dark:border-slate-700',
    dotColor: 'bg-slate-500',
    Icon: Code,
  },
}

function getEventConfig(type: ReplayEventType): EventTypeConfig {
  return EVENT_TYPE_CONFIG[type] ?? EVENT_TYPE_CONFIG.custom
}

// ── Timeline Player ──────────────────────────────────────────────────────────

interface TimelinePlayerProps {
  events: ReplayEvent[]
  loading: boolean
  duration: number
}

function TimelinePlayer({ events, loading, duration }: TimelinePlayerProps) {
  const [selectedEvent, setSelectedEvent] = useState<ReplayEvent | null>(null)

  if (loading) {
    return (
      <div className="flex items-center justify-center h-40">
        <RefreshCw className="w-5 h-5 animate-spin text-slate-400" />
      </div>
    )
  }

  if (events.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-40 text-slate-400 gap-2">
        <Video className="w-8 h-8" aria-hidden="true" />
        <p className="text-sm">Nenhum evento registrado nesta sessão</p>
      </div>
    )
  }

  const minTs = events[0]?.timestamp_ms ?? 0
  const maxTs = duration > 0 ? duration : (events[events.length - 1]?.timestamp_ms ?? 1) - minTs

  return (
    <div className="space-y-4">
      {/* Timeline track */}
      <div className="relative">
        <div className="h-2 bg-slate-200 dark:bg-slate-700 rounded-full relative overflow-visible">
          {events.map((ev) => {
            const pct = maxTs > 0 ? ((ev.timestamp_ms - minTs) / maxTs) * 100 : 0
            const cfg = getEventConfig(ev.type)
            return (
              <button
                key={ev.id}
                title={`${cfg.label} — ${ev.timestamp_ms - minTs}ms`}
                onClick={() => setSelectedEvent(selectedEvent?.id === ev.id ? null : ev)}
                style={{ left: `${Math.min(pct, 98)}%` }}
                className={`absolute -top-1 w-4 h-4 rounded-full border-2 border-white dark:border-slate-900 ${cfg.dotColor} cursor-pointer hover:scale-125 transition-transform`}
                aria-label={`Evento ${cfg.label} em ${ev.timestamp_ms - minTs}ms`}
              />
            )
          })}
        </div>
        <div className="flex justify-between text-xs text-slate-400 mt-2">
          <span>0s</span>
          <span>{formatDuration(maxTs)}</span>
        </div>
      </div>

      {/* Legend */}
      <div className="flex flex-wrap gap-2">
        {(Object.keys(EVENT_TYPE_CONFIG) as ReplayEventType[]).map((type) => {
          const cfg = getEventConfig(type)
          const count = events.filter((e) => e.type === type).length
          if (count === 0) return null
          return (
            <span
              key={type}
              className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium border ${cfg.color}`}
            >
              <span className={`w-1.5 h-1.5 rounded-full ${cfg.dotColor}`} aria-hidden="true" />
              {cfg.label} ({count})
            </span>
          )
        })}
      </div>

      {/* Event list */}
      <div className="space-y-1 max-h-64 overflow-y-auto pr-1">
        {events.map((ev) => {
          const cfg = getEventConfig(ev.type)
          const Icon = cfg.Icon
          const isSelected = selectedEvent?.id === ev.id
          return (
            <button
              key={ev.id}
              onClick={() => setSelectedEvent(isSelected ? null : ev)}
              className={`w-full flex items-start gap-3 px-3 py-2 rounded-lg text-left text-sm transition-colors ${
                isSelected
                  ? 'bg-slate-100 dark:bg-slate-800'
                  : 'hover:bg-slate-50 dark:hover:bg-slate-800/50'
              }`}
            >
              <span className={`mt-0.5 p-1 rounded border ${cfg.color} flex-shrink-0`}>
                <Icon className="w-3 h-3" aria-hidden="true" />
              </span>
              <span className="flex-1 min-w-0">
                <span className="font-medium text-slate-700 dark:text-slate-300">{cfg.label}</span>
                <span className="ml-2 text-xs text-slate-400 font-mono">
                  +{formatDuration(ev.timestamp_ms - minTs)}
                </span>
              </span>
              <ChevronRight
                className={`w-3.5 h-3.5 text-slate-400 flex-shrink-0 mt-0.5 transition-transform ${isSelected ? 'rotate-90' : ''}`}
                aria-hidden="true"
              />
            </button>
          )
        })}
      </div>

      {/* Selected event data */}
      {selectedEvent !== null && (
        <div className="rounded-xl border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-900 p-3">
          <p className="text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-2">
            Dados do evento
          </p>
          <pre className="text-xs font-mono text-slate-700 dark:text-slate-300 overflow-x-auto whitespace-pre-wrap break-all">
            {JSON.stringify(selectedEvent.data, null, 2)}
          </pre>
        </div>
      )}
    </div>
  )
}

// ── Session slide-over panel ─────────────────────────────────────────────────

interface SessionPanelProps {
  session: ReplaySession
  onClose: () => void
}

function SessionPanel({ session, onClose }: SessionPanelProps) {
  const [events, setEvents] = useState<ReplayEvent[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    replayApi
      .events(session.session_id)
      .then((res) => {
        if (!cancelled) setEvents(res.data ?? [])
      })
      .catch(() => {
        if (!cancelled) toast.error('Falha ao carregar eventos da sessão')
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [session.session_id])

  return (
    <>
      {/* Backdrop */}
      <div
        className="fixed inset-0 bg-black/40 z-40 backdrop-blur-sm"
        onClick={onClose}
        aria-hidden="true"
      />

      {/* Panel */}
      <aside
        role="dialog"
        aria-modal="true"
        aria-label={`Sessão ${session.session_id}`}
        className="fixed right-0 top-0 h-full w-full max-w-lg bg-white dark:bg-slate-900 shadow-2xl z-50 flex flex-col"
      >
        {/* Header */}
        <div className="flex items-center justify-between px-5 py-4 border-b border-slate-200 dark:border-slate-800 flex-shrink-0">
          <div className="flex items-center gap-2">
            <Video className="w-4 h-4 text-emerald-500" aria-hidden="true" />
            <h2 className="font-semibold text-slate-800 dark:text-slate-100 text-sm">
              Session Replay
            </h2>
          </div>
          <button
            onClick={onClose}
            className="p-1.5 rounded-lg text-slate-500 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
            aria-label="Fechar painel"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* Session meta */}
        <div className="px-5 py-4 border-b border-slate-200 dark:border-slate-800 flex-shrink-0 space-y-2">
          <div className="grid grid-cols-2 gap-3">
            <div>
              <p className="text-xs text-slate-400 uppercase tracking-wide font-medium">Usuário</p>
              <p className="text-sm font-mono text-slate-700 dark:text-slate-300 truncate" title={session.user_id}>
                {session.user_id}
              </p>
            </div>
            <div>
              <p className="text-xs text-slate-400 uppercase tracking-wide font-medium">Duração</p>
              <p className="text-sm font-semibold text-emerald-600 dark:text-emerald-400">
                {formatDuration(session.duration_ms)}
              </p>
            </div>
            <div>
              <p className="text-xs text-slate-400 uppercase tracking-wide font-medium">Tela</p>
              <p className="text-sm text-slate-700 dark:text-slate-300">
                {session.screen_width}×{session.screen_height}
              </p>
            </div>
            <div>
              <p className="text-xs text-slate-400 uppercase tracking-wide font-medium">Início</p>
              <p className="text-sm text-slate-700 dark:text-slate-300">{formatDate(session.started_at)}</p>
            </div>
          </div>
          <div>
            <p className="text-xs text-slate-400 uppercase tracking-wide font-medium">URL</p>
            <p className="text-sm font-mono text-slate-700 dark:text-slate-300 break-all">
              {session.page_url}
            </p>
          </div>
        </div>

        {/* Player */}
        <div className="flex-1 overflow-y-auto px-5 py-4">
          <TimelinePlayer events={events} loading={loading} duration={session.duration_ms} />
        </div>
      </aside>
    </>
  )
}

// ── Main page ────────────────────────────────────────────────────────────────

export default function Replay() {
  const [sessions, setSessions] = useState<ReplaySession[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [selectedSession, setSelectedSession] = useState<ReplaySession | null>(null)
  const [filterUrl, setFilterUrl] = useState('')
  const [filterUserId, setFilterUserId] = useState('')

  const fetchSessions = useCallback(async () => {
    setLoading(true)
    try {
      const params: { url?: string; user_id?: string } = {}
      if (filterUrl.trim()) params.url = filterUrl.trim()
      if (filterUserId.trim()) params.user_id = filterUserId.trim()
      const res = await replayApi.sessions(params)
      const body = res.data
      // API pode retornar array direto ou objeto com sessions[]
      if (Array.isArray(body)) {
        setSessions(body)
        setTotal(body.length)
      } else {
        setSessions(body.sessions ?? [])
        setTotal(body.total ?? 0)
      }
    } catch {
      toast.error('Falha ao carregar sessões de replay')
    } finally {
      setLoading(false)
    }
  }, [filterUrl, filterUserId])

  useEffect(() => {
    void fetchSessions()
  }, [fetchSessions])

  return (
    <div className="space-y-5">
      {/* Page header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-xl font-bold text-slate-900 dark:text-slate-100">Session Replay</h1>
          <p className="text-sm text-slate-500 dark:text-slate-400 mt-0.5">
            {loading ? 'Carregando…' : `${total.toLocaleString('pt-BR')} sessões registradas`}
          </p>
        </div>
        <button
          onClick={fetchSessions}
          disabled={loading}
          className="self-start sm:self-auto p-2 rounded-lg text-slate-500 hover:bg-slate-100 dark:text-slate-400 dark:hover:bg-slate-800 transition-colors disabled:opacity-50"
          aria-label="Recarregar sessões"
        >
          <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
        </button>
      </div>

      {/* Filters */}
      <div className="flex flex-col sm:flex-row gap-3">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-slate-400" aria-hidden="true" />
          <input
            type="text"
            placeholder="Filtrar por URL…"
            value={filterUrl}
            onChange={(e) => setFilterUrl(e.target.value)}
            className="w-full pl-9 pr-3 py-2 text-sm rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-800 dark:text-slate-200 placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-emerald-500/40 focus:border-emerald-500"
          />
        </div>
        <div className="relative flex-1 sm:max-w-xs">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-slate-400" aria-hidden="true" />
          <input
            type="text"
            placeholder="Filtrar por user ID…"
            value={filterUserId}
            onChange={(e) => setFilterUserId(e.target.value)}
            className="w-full pl-9 pr-3 py-2 text-sm rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-800 dark:text-slate-200 placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-emerald-500/40 focus:border-emerald-500"
          />
        </div>
      </div>

      {/* Table */}
      <div className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 overflow-hidden">
        {loading ? (
          <div className="flex items-center justify-center h-40">
            <RefreshCw className="w-5 h-5 animate-spin text-slate-400" />
          </div>
        ) : sessions.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-40 text-slate-400 gap-2">
            <Video className="w-8 h-8" aria-hidden="true" />
            <p className="text-sm">Nenhuma sessão encontrada</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-800/50">
                  <th className="px-4 py-3 text-left text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide">
                    Usuário
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide">
                    URL
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide">
                    Duração
                  </th>
                  <th className="px-4 py-3 text-right text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide">
                    Eventos
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide">
                    Data
                  </th>
                  <th className="px-4 py-3" />
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100 dark:divide-slate-800">
                {sessions.map((session) => (
                  <tr
                    key={session.id}
                    className="hover:bg-slate-50 dark:hover:bg-slate-800/40 transition-colors cursor-pointer"
                    onClick={() => setSelectedSession(session)}
                  >
                    <td className="px-4 py-3">
                      <span
                        className="font-mono text-xs text-slate-600 dark:text-slate-400 truncate block max-w-[160px]"
                        title={session.user_id}
                      >
                        {session.user_id}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <span
                        className="text-slate-700 dark:text-slate-300 truncate block max-w-[220px]"
                        title={session.page_url}
                      >
                        {truncateUrl(session.page_url)}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-emerald-50 dark:bg-emerald-950/40 text-emerald-700 dark:text-emerald-400 border border-emerald-200 dark:border-emerald-800">
                        {formatDuration(session.duration_ms)}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-right">
                      <span className="font-mono text-slate-700 dark:text-slate-300">
                        {session.event_count.toLocaleString('pt-BR')}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-slate-500 dark:text-slate-400 text-xs whitespace-nowrap">
                      {formatDate(session.started_at)}
                    </td>
                    <td className="px-4 py-3">
                      <ChevronRight className="w-4 h-4 text-slate-400" aria-hidden="true" />
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Slide-over panel */}
      {selectedSession !== null && (
        <SessionPanel session={selectedSession} onClose={() => setSelectedSession(null)} />
      )}
    </div>
  )
}
