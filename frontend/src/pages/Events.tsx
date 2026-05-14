import { useCallback, useEffect, useState } from 'react'
import { Activity, RefreshCw, Wifi, Users as UsersIcon, ChevronRight, ChevronDown } from 'lucide-react'
import toast from 'react-hot-toast'
import {
  eventsApi,
  type EventSummaryItem,
  type EventsSummaryResponse,
  type RawEvent,
} from '../services/pulseApi'
import { useRealtimeEvents, type RealtimeEvent } from '../hooks/useRealtimeEvents'
import { useProject } from '../components/Layout'

// ── Types (aba Usuários) ───────────────────────────────────────────────────────

interface UserProfile {
  userId: string
  eventCount: number
  lastSeen: string
  products: string[]
  userAgent: string
  events: RawEvent[]
}

// ── Constantes de cor por produto ─────────────────────────────────────────────

const PRODUCT_COLORS: Record<string, string> = {
  libri:  'from-blue-50 to-blue-100 dark:from-blue-950/40 dark:to-blue-900/20 border-blue-200 dark:border-blue-800 text-blue-700 dark:text-blue-300',
  nitro:  'from-orange-50 to-orange-100 dark:from-orange-950/40 dark:to-orange-900/20 border-orange-200 dark:border-orange-800 text-orange-700 dark:text-orange-300',
  brio:   'from-emerald-50 to-emerald-100 dark:from-emerald-950/40 dark:to-emerald-900/20 border-emerald-200 dark:border-emerald-800 text-emerald-700 dark:text-emerald-300',
}

const defaultColor =
  'from-slate-50 to-slate-100 dark:from-slate-800/40 dark:to-slate-700/20 border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-300'

// ── Helpers ───────────────────────────────────────────────────────────────────

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

function timeAgo(isoDate: string): string {
  const diffMs = Date.now() - new Date(isoDate).getTime()
  const diffSec = Math.floor(diffMs / 1000)

  if (diffSec < 60) return `há ${diffSec}s`
  if (diffSec < 3600) return `há ${Math.floor(diffSec / 60)}min`
  if (diffSec < 86400) return `há ${Math.floor(diffSec / 3600)}h`
  if (diffSec < 86400 * 30) return `há ${Math.floor(diffSec / 86400)}d`
  return `há ${Math.floor(diffSec / (86400 * 30))}m`
}

function abbreviateId(id: string): string {
  if (id.length <= 12) return id
  return `${id.slice(0, 8)}…${id.slice(-4)}`
}

function parseUserAgent(ua: string): string {
  if (!ua) return 'Desconhecido'
  const browser = ua.includes('Chrome')
    ? 'Chrome'
    : ua.includes('Firefox')
    ? 'Firefox'
    : ua.includes('Safari')
    ? 'Safari'
    : ua.includes('Edge')
    ? 'Edge'
    : 'Browser'
  const os = ua.includes('Mac')
    ? 'macOS'
    : ua.includes('Windows')
    ? 'Windows'
    : ua.includes('Linux')
    ? 'Linux'
    : ua.includes('Android')
    ? 'Android'
    : ua.includes('iPhone') || ua.includes('iPad')
    ? 'iOS'
    : 'OS'
  return `${browser}/${os}`
}

function summarizeProperties(props: Record<string, unknown>): string {
  const parts: string[] = []

  const path = props['path'] ?? props['page'] ?? props['url']
  if (path != null) parts.push(String(path))

  const variant = props['variant'] ?? props['ab_variant']
  if (variant != null) parts.push(`variant=${String(variant)}`)

  const referrer = props['referrer'] ?? props['ref']
  if (referrer != null) parts.push(`referrer=${String(referrer)}`)

  return parts.join('  ')
}

function groupByUser(events: RawEvent[]): UserProfile[] {
  const map = new Map<string, RawEvent[]>()

  for (const ev of events) {
    const rawId = ev.properties['user_id']
    if (typeof rawId !== 'string' || rawId === '') continue
    const existing = map.get(rawId)
    if (existing) {
      existing.push(ev)
    } else {
      map.set(rawId, [ev])
    }
  }

  const profiles: UserProfile[] = []

  map.forEach((userEvents, userId) => {
    const sorted = [...userEvents].sort(
      (a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime(),
    )
    const latest = sorted[0]
    const products = [...new Set(userEvents.map((e) => e.product))]

    profiles.push({
      userId,
      eventCount: userEvents.length,
      lastSeen: latest.created_at,
      products,
      userAgent: latest.user_agent,
      events: sorted,
    })
  })

  return profiles.sort(
    (a, b) => new Date(b.lastSeen).getTime() - new Date(a.lastSeen).getTime(),
  )
}

// ── Sub-componentes (aba Usuários) ────────────────────────────────────────────

function SkeletonCard() {
  return (
    <div className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-4 animate-pulse">
      <div className="flex items-center gap-3">
        <div className="h-4 w-32 bg-slate-200 dark:bg-slate-700 rounded" />
        <div className="h-4 w-16 bg-slate-200 dark:bg-slate-700 rounded-full" />
        <div className="ml-auto h-4 w-20 bg-slate-200 dark:bg-slate-700 rounded" />
      </div>
      <div className="mt-2 h-3 w-24 bg-slate-100 dark:bg-slate-800 rounded" />
    </div>
  )
}

interface UserCardProps {
  profile: UserProfile
}

function UserCard({ profile }: UserCardProps) {
  const [expanded, setExpanded] = useState(false)

  return (
    <div className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 overflow-hidden">
      <button
        type="button"
        onClick={() => setExpanded((v) => !v)}
        className="w-full flex items-center gap-3 px-4 py-3 text-left hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-colors"
        aria-expanded={expanded}
        aria-label={`Usuário ${abbreviateId(profile.userId)} — ${profile.eventCount} eventos`}
      >
        <span className="font-mono text-sm font-medium text-slate-800 dark:text-slate-200 shrink-0">
          {abbreviateId(profile.userId)}
        </span>

        <div className="flex items-center gap-1 flex-wrap">
          {profile.products.map((p) => (
            <span
              key={p}
              className="px-2 py-0.5 rounded-full text-xs font-medium bg-violet-100 text-violet-700 dark:bg-violet-900/40 dark:text-violet-300"
            >
              {p}
            </span>
          ))}
        </div>

        <span className="hidden sm:block text-xs text-slate-400 dark:text-slate-500 shrink-0">
          {parseUserAgent(profile.userAgent)}
        </span>

        <span className="flex-1" />

        <span className="text-sm font-semibold text-slate-700 dark:text-slate-300 shrink-0">
          {profile.eventCount} eventos
        </span>

        {expanded ? (
          <ChevronDown className="w-4 h-4 text-slate-400 shrink-0" aria-hidden="true" />
        ) : (
          <ChevronRight className="w-4 h-4 text-slate-400 shrink-0" aria-hidden="true" />
        )}
      </button>

      <div className="px-4 pb-2 -mt-1">
        <span className="text-xs text-slate-400 dark:text-slate-500">
          Último evento: {timeAgo(profile.lastSeen)}
        </span>
      </div>

      {expanded && (
        <div className="border-t border-slate-100 dark:border-slate-800 overflow-x-auto">
          <table className="w-full text-xs">
            <thead>
              <tr className="bg-slate-50 dark:bg-slate-800/60">
                <th className="px-4 py-2 text-left font-medium text-slate-500 dark:text-slate-400 whitespace-nowrap">
                  Data/Hora
                </th>
                <th className="px-4 py-2 text-left font-medium text-slate-500 dark:text-slate-400 whitespace-nowrap">
                  Tipo
                </th>
                <th className="px-4 py-2 text-left font-medium text-slate-500 dark:text-slate-400">
                  Properties
                </th>
              </tr>
            </thead>
            <tbody>
              {profile.events.map((ev) => (
                <tr
                  key={ev.id}
                  className="border-t border-slate-100 dark:border-slate-800 hover:bg-slate-50 dark:hover:bg-slate-800/40 transition-colors"
                >
                  <td className="px-4 py-2 font-mono text-slate-500 dark:text-slate-400 whitespace-nowrap">
                    {new Date(ev.created_at).toLocaleString('pt-BR', {
                      year: 'numeric',
                      month: '2-digit',
                      day: '2-digit',
                      hour: '2-digit',
                      minute: '2-digit',
                    })}
                  </td>
                  <td className="px-4 py-2 font-mono text-slate-700 dark:text-slate-300 whitespace-nowrap">
                    {ev.event_type}
                  </td>
                  <td className="px-4 py-2 text-slate-500 dark:text-slate-400 font-mono">
                    {summarizeProperties(ev.properties) || '—'}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}

// ── Aba Usuários ──────────────────────────────────────────────────────────────

function UsersTab() {
  const { projectId } = useProject()
  const [profiles, setProfiles] = useState<UserProfile[]>([])
  const [totalEvents, setTotalEvents] = useState(0)
  const [loading, setLoading] = useState(true)

  const fetchUsers = useCallback(async () => {
    setLoading(true)
    try {
      const params: Record<string, string | number> = { limit: 500 }
      if (projectId) params['project_id'] = projectId

      const { data } = await eventsApi.list(params)
      setTotalEvents(data.total)
      setProfiles(groupByUser(data.items))
    } catch (err) {
      toast.error('Falha ao carregar usuários')
      console.error(err)
    } finally {
      setLoading(false)
    }
  }, [projectId])

  useEffect(() => {
    void fetchUsers()
  }, [fetchUsers])

  return (
    <div className="space-y-4">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
        {!loading && (
          <p className="text-sm text-slate-500 dark:text-slate-400">
            {totalEvents.toLocaleString('pt-BR')} eventos · {profiles.length} usuário
            {profiles.length !== 1 ? 's' : ''} único{profiles.length !== 1 ? 's' : ''}
          </p>
        )}
        <button
          onClick={() => void fetchUsers()}
          disabled={loading}
          className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-medium border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-600 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors disabled:opacity-50 self-start sm:self-auto"
          aria-label="Atualizar lista de usuários"
        >
          <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} aria-hidden="true" />
          Atualizar
        </button>
      </div>

      {loading && (
        <div className="space-y-2">
          <SkeletonCard />
          <SkeletonCard />
          <SkeletonCard />
        </div>
      )}

      {!loading && profiles.length === 0 && (
        <div className="rounded-xl border border-dashed border-violet-300 dark:border-violet-700 bg-violet-50/50 dark:bg-violet-950/20 p-12 flex flex-col items-center text-center gap-4">
          <div className="w-16 h-16 rounded-2xl bg-violet-100 dark:bg-violet-900/40 flex items-center justify-center">
            <UsersIcon className="w-8 h-8 text-violet-500 dark:text-violet-400" aria-hidden="true" />
          </div>
          <div className="space-y-1">
            <h2 className="text-base font-semibold text-slate-800 dark:text-slate-200">
              Nenhum usuário rastreado ainda
            </h2>
            <p className="text-sm text-slate-500 dark:text-slate-400 max-w-md">
              Instale o SDK e envie eventos com{' '}
              <code className="font-mono text-violet-600 dark:text-violet-400">user_id</code> nas
              properties para que os usuários apareçam aqui.
            </p>
          </div>
          <div className="w-full max-w-sm text-left">
            <div className="rounded-lg bg-slate-900 text-slate-100 p-4 text-xs font-mono space-y-1 overflow-x-auto">
              <p>{'await pulse.trackEvent({'}</p>
              <p>{'  event_name: "page_view",'}</p>
              <p>{'  product: "libri",'}</p>
              <p className="text-emerald-400">{'  properties: { user_id: userId },'}</p>
              <p>{'});'}</p>
            </div>
          </div>
        </div>
      )}

      {!loading && profiles.length > 0 && (
        <div className="space-y-2">
          {profiles.map((profile) => (
            <UserCard key={profile.userId} profile={profile} />
          ))}
        </div>
      )}
    </div>
  )
}

// ── Page principal ────────────────────────────────────────────────────────────

type Tab = 'resumo' | 'usuarios'

export default function Events() {
  const [activeTab, setActiveTab] = useState<Tab>('resumo')
  const [summary, setSummary] = useState<EventsSummaryResponse | null>(null)
  const [loading, setLoading] = useState(true)
  const [realtimeCount, setRealtimeCount] = useState(0)

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

  // BKL-150: receber eventos em tempo real via WebSocket
  const handleRealtimeEvent = useCallback((msg: RealtimeEvent) => {
    if (msg.type !== 'event') return
    setRealtimeCount((prev) => prev + 1)
    setSummary((prev) => {
      if (!prev) return prev
      const items = [...prev.items]
      const idx = items.findIndex(
        (i) => i.product === msg.event.product && i.event_type === msg.event.event_type,
      )
      if (idx >= 0) {
        items[idx] = { ...items[idx], count: items[idx].count + 1 }
      } else {
        items.push({ product: msg.event.product, event_type: msg.event.event_type, count: 1 })
      }
      return { ...prev, items, total_last_24h: prev.total_last_24h + 1 }
    })
  }, [])

  useRealtimeEvents({ onEvent: handleRealtimeEvent })

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
        <div className="flex items-center gap-2 self-start sm:self-auto">
          <span
            title={`${realtimeCount} eventos recebidos em tempo real`}
            className="flex items-center gap-1 text-xs text-emerald-600 dark:text-emerald-400 bg-emerald-50 dark:bg-emerald-950/30 border border-emerald-200 dark:border-emerald-800 px-2 py-1 rounded-full"
          >
            <Wifi className="w-3 h-3" />
            <span>real-time</span>
            {realtimeCount > 0 && (
              <span className="font-mono font-bold">+{realtimeCount}</span>
            )}
          </span>
          <button
            onClick={fetchSummary}
            disabled={loading}
            className="p-2 rounded-lg text-slate-500 hover:bg-slate-100 dark:text-slate-400 dark:hover:bg-slate-800 transition-colors disabled:opacity-50"
            aria-label="Recarregar"
          >
            <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
          </button>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex items-center border-b border-slate-200 dark:border-slate-800">
        <button
          type="button"
          onClick={() => setActiveTab('resumo')}
          className={[
            'px-4 py-2 text-sm font-medium transition-colors border-b-2 -mb-px',
            activeTab === 'resumo'
              ? 'border-violet-600 text-violet-700 dark:border-violet-400 dark:text-violet-300'
              : 'border-transparent text-slate-500 hover:text-slate-700 dark:text-slate-400 dark:hover:text-slate-200',
          ].join(' ')}
        >
          Resumo
        </button>
        <button
          type="button"
          onClick={() => setActiveTab('usuarios')}
          className={[
            'px-4 py-2 text-sm font-medium transition-colors border-b-2 -mb-px',
            activeTab === 'usuarios'
              ? 'border-violet-600 text-violet-700 dark:border-violet-400 dark:text-violet-300'
              : 'border-transparent text-slate-500 hover:text-slate-700 dark:text-slate-400 dark:hover:text-slate-200',
          ].join(' ')}
        >
          Usuários
        </button>
      </div>

      {/* Aba Resumo */}
      {activeTab === 'resumo' && (
        <div className="space-y-4">
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
                  Os eventos aparecerão aqui quando suas aplicações começarem a enviar dados via SDK.
                </p>
              </div>
              {/* BKL-1051: instrução de SDK no empty state */}
              <div className="w-full max-w-xl text-left space-y-3">
                <p className="text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide">Rastrear evento via SDK</p>
                <div className="rounded-lg bg-slate-900 text-slate-100 p-4 text-xs font-mono space-y-1 overflow-x-auto">
                  <p>{'import PulseClient from "@iit/pulse-sdk";'}</p>
                  <p>{''}</p>
                  <p>{'const pulse = new PulseClient({'}</p>
                  <p>{'  baseURL: "https://pulse-api.institutoitinerante.com.br",'}</p>
                  <p>{'});'}</p>
                  <p>{''}</p>
                  <p>{'// Evento imediato'}</p>
                  <p>{'await pulse.trackEvent({'}</p>
                  <p>{'  event_name: "checkout_completed",'}</p>
                  <p>{'  user_id: userId,'}</p>
                  <p>{'  product: "libri",'}</p>
                  <p>{'});'}</p>
                  <p>{''}</p>
                  <p>{'// Ou com batch queue (mais eficiente)'}</p>
                  <p>{'pulse.queue({ event_name: "page_view", product: "nitro" });'}</p>
                </div>
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
      )}

      {/* Aba Usuários */}
      {activeTab === 'usuarios' && <UsersTab />}
    </div>
  )
}
