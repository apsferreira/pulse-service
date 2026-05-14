import { useCallback, useEffect, useState } from 'react'
import {
  Activity,
  RefreshCw,
  Wifi,
  Users as UsersIcon,
  ChevronRight,
  ChevronDown,
  Zap,
  BarChart2,
  TrendingUp,
  Clock,
  Monitor,
} from 'lucide-react'
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

// Cor da barra de progresso por produto
const PRODUCT_BAR_COLORS: Record<string, string> = {
  libri:  'bg-blue-400 dark:bg-blue-500',
  nitro:  'bg-orange-400 dark:bg-orange-500',
  brio:   'bg-emerald-400 dark:bg-emerald-500',
}

const defaultColor =
  'from-slate-50 to-slate-100 dark:from-slate-800/40 dark:to-slate-700/20 border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-300'

const defaultBarColor = 'bg-violet-400 dark:bg-violet-500'

// ── Helpers ───────────────────────────────────────────────────────────────────

function getProductColor(product: string): string {
  return PRODUCT_COLORS[product] ?? defaultColor
}

function getProductBarColor(product: string): string {
  return PRODUCT_BAR_COLORS[product] ?? defaultBarColor
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

// ── Sub-componentes compartilhados ─────────────────────────────────────────────

/** Skeleton genérico de card para loading states */
function SkeletonMetricCard() {
  return (
    <div className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-4 animate-pulse">
      <div className="flex items-start justify-between">
        <div className="space-y-2">
          <div className="h-3 w-28 bg-slate-200 dark:bg-slate-700 rounded" />
          <div className="h-8 w-16 bg-slate-200 dark:bg-slate-700 rounded" />
          <div className="h-3 w-20 bg-slate-100 dark:bg-slate-800 rounded" />
        </div>
        <div className="h-9 w-9 bg-slate-100 dark:bg-slate-800 rounded-lg" />
      </div>
    </div>
  )
}

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

// ── Sub-componentes (aba Usuários) ────────────────────────────────────────────

interface UserCardProps {
  profile: UserProfile
  rank: number
}

function UserCard({ profile, rank }: UserCardProps) {
  const [expanded, setExpanded] = useState(false)

  const uniqueEventTypes = new Set(profile.events.map((e) => e.event_type)).size

  return (
    <div className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 overflow-hidden transition-shadow hover:shadow-sm">
      <button
        type="button"
        onClick={() => setExpanded((v) => !v)}
        className="w-full flex items-center gap-3 px-4 py-3 text-left hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-colors"
        aria-expanded={expanded}
        aria-label={`Usuário ${abbreviateId(profile.userId)} — ${profile.eventCount} eventos`}
      >
        {/* Rank badge */}
        <span
          className="hidden sm:flex items-center justify-center w-6 h-6 rounded-full bg-slate-100 dark:bg-slate-800 text-xs font-bold text-slate-500 dark:text-slate-400 shrink-0"
          aria-hidden="true"
        >
          {rank}
        </span>

        {/* User ID */}
        <span className="font-mono text-sm font-medium text-slate-800 dark:text-slate-200 shrink-0">
          {abbreviateId(profile.userId)}
        </span>

        {/* Produto badges */}
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

        {/* Separador flexível */}
        <span className="flex-1" />

        {/* Meta info (escondida em mobile pequeno) */}
        <div className="hidden md:flex items-center gap-3 text-xs text-slate-400 dark:text-slate-500 shrink-0">
          <span className="flex items-center gap-1">
            <Monitor className="w-3 h-3" aria-hidden="true" />
            {parseUserAgent(profile.userAgent)}
          </span>
          <span className="flex items-center gap-1">
            <Clock className="w-3 h-3" aria-hidden="true" />
            {timeAgo(profile.lastSeen)}
          </span>
        </div>

        {/* Contadores */}
        <div className="flex items-center gap-2 shrink-0">
          <span className="text-sm font-semibold text-slate-700 dark:text-slate-300">
            {profile.eventCount}
            <span className="hidden sm:inline font-normal text-slate-400 dark:text-slate-500 ml-1">
              eventos
            </span>
          </span>
          {uniqueEventTypes > 1 && (
            <span className="hidden sm:inline px-1.5 py-0.5 rounded text-xs bg-slate-100 dark:bg-slate-800 text-slate-500 dark:text-slate-400">
              {uniqueEventTypes} tipos
            </span>
          )}
        </div>

        {expanded ? (
          <ChevronDown className="w-4 h-4 text-slate-400 shrink-0" aria-hidden="true" />
        ) : (
          <ChevronRight className="w-4 h-4 text-slate-400 shrink-0" aria-hidden="true" />
        )}
      </button>

      {/* Info secundária visível em mobile */}
      <div className="px-4 pb-2 -mt-1 flex items-center gap-3 md:hidden">
        <span className="text-xs text-slate-400 dark:text-slate-500 flex items-center gap-1">
          <Clock className="w-3 h-3" aria-hidden="true" />
          {timeAgo(profile.lastSeen)}
        </span>
      </div>

      {/* Histórico expandido */}
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

  // Usuário mais ativo
  const topUser = profiles[0] ?? null

  return (
    <div className="space-y-4">
      {/* Header com stats + botão */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
        <div className="flex items-center gap-3">
          {loading ? (
            <>
              <div className="h-8 w-28 bg-slate-100 dark:bg-slate-800 rounded-lg animate-pulse" />
              <div className="h-8 w-28 bg-slate-100 dark:bg-slate-800 rounded-lg animate-pulse" />
            </>
          ) : (
            <>
              <div className="flex items-center gap-2 px-3 py-1.5 rounded-lg bg-violet-50 dark:bg-violet-950/30 border border-violet-200 dark:border-violet-800">
                <UsersIcon className="w-3.5 h-3.5 text-violet-500" aria-hidden="true" />
                <span className="text-sm font-semibold text-violet-700 dark:text-violet-300">
                  {profiles.length}
                </span>
                <span className="text-xs text-violet-500 dark:text-violet-400">
                  usuário{profiles.length !== 1 ? 's' : ''}
                </span>
              </div>
              <div className="flex items-center gap-2 px-3 py-1.5 rounded-lg bg-slate-50 dark:bg-slate-800/50 border border-slate-200 dark:border-slate-700">
                <Activity className="w-3.5 h-3.5 text-slate-400" aria-hidden="true" />
                <span className="text-sm font-semibold text-slate-700 dark:text-slate-300">
                  {totalEvents.toLocaleString('pt-BR')}
                </span>
                <span className="text-xs text-slate-400 dark:text-slate-500">eventos</span>
              </div>
            </>
          )}
        </div>
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

      {/* Top user highlight (quando há dados) */}
      {!loading && topUser && (
        <div className="rounded-xl border border-violet-200 dark:border-violet-800 bg-gradient-to-r from-violet-50 to-purple-50 dark:from-violet-950/20 dark:to-purple-950/20 px-4 py-3 flex items-center gap-3">
          <div className="w-8 h-8 rounded-full bg-violet-100 dark:bg-violet-900/50 flex items-center justify-center shrink-0">
            <TrendingUp className="w-4 h-4 text-violet-500" aria-hidden="true" />
          </div>
          <div className="min-w-0">
            <p className="text-xs text-violet-500 dark:text-violet-400 font-medium">Usuário mais ativo</p>
            <p className="text-sm font-semibold text-slate-800 dark:text-slate-200 font-mono truncate">
              {abbreviateId(topUser.userId)}
            </p>
          </div>
          <div className="ml-auto flex items-center gap-3 shrink-0 text-right">
            <div>
              <p className="text-lg font-bold text-violet-700 dark:text-violet-300">
                {topUser.eventCount}
              </p>
              <p className="text-xs text-slate-400 dark:text-slate-500">eventos</p>
            </div>
          </div>
        </div>
      )}

      {/* Loading skeletons */}
      {loading && (
        <div className="space-y-2">
          <SkeletonCard />
          <SkeletonCard />
          <SkeletonCard />
        </div>
      )}

      {/* Empty state */}
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

      {/* Lista de usuários */}
      {!loading && profiles.length > 0 && (
        <div className="space-y-2">
          {profiles.map((profile, index) => (
            <UserCard key={profile.userId} profile={profile} rank={index + 1} />
          ))}
        </div>
      )}
    </div>
  )
}

// ── Componente de barra horizontal de distribuição ─────────────────────────────

interface EventBarProps {
  item: EventSummaryItem
  maxCount: number
  barColorClass: string
}

function EventBar({ item, maxCount, barColorClass }: EventBarProps) {
  const pct = maxCount > 0 ? Math.max(2, Math.round((item.count / maxCount) * 100)) : 0

  return (
    <div className="flex items-center gap-3 group">
      {/* Nome do evento */}
      <span
        className="font-mono text-xs text-slate-600 dark:text-slate-300 truncate shrink-0"
        style={{ minWidth: '8rem', maxWidth: '12rem' }}
        title={item.event_type}
      >
        {item.event_type}
      </span>

      {/* Barra */}
      <div className="flex-1 flex items-center gap-2 min-w-0">
        <div className="flex-1 h-2 rounded-full bg-slate-100 dark:bg-slate-800 overflow-hidden">
          <div
            className={`h-full rounded-full transition-all duration-500 ${barColorClass}`}
            style={{ width: `${pct}%` }}
            role="progressbar"
            aria-valuenow={item.count}
            aria-valuemin={0}
            aria-valuemax={maxCount}
            aria-label={`${item.event_type}: ${item.count} eventos`}
          />
        </div>
        {/* Contagem */}
        <span className="text-xs font-semibold text-slate-700 dark:text-slate-200 shrink-0 text-right tabular-nums w-10">
          {item.count.toLocaleString('pt-BR')}
        </span>
      </div>
    </div>
  )
}

// ── Page principal ────────────────────────────────────────────────────────────

type Tab = 'resumo' | 'usuarios'

export default function Events() {
  const { projectId } = useProject()
  const [activeTab, setActiveTab] = useState<Tab>('resumo')
  const [summary, setSummary] = useState<EventsSummaryResponse | null>(null)
  const [loading, setLoading] = useState(true)
  const [realtimeCount, setRealtimeCount] = useState(0)

  const fetchSummary = useCallback(async () => {
    setLoading(true)
    try {
      const { data } = await eventsApi.summary(projectId)
      setSummary(data)
    } catch (err) {
      toast.error('Falha ao carregar resumo de eventos')
      console.error(err)
    } finally {
      setLoading(false)
    }
  }, [projectId])

  // BKL-150: receber eventos em tempo real via WebSocket
  // Filtra por project_id quando um projeto está selecionado
  const handleRealtimeEvent = useCallback((msg: RealtimeEvent) => {
    if (msg.type !== 'event') return
    if (projectId && msg.event.properties?.['project_id'] && msg.event.properties['project_id'] !== projectId) return
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
  }, [projectId])

  useRealtimeEvents({ onEvent: handleRealtimeEvent })

  // Reset ao mudar projeto
  useEffect(() => {
    setRealtimeCount(0)
    setSummary(null)
  }, [projectId])

  useEffect(() => {
    void fetchSummary()
    const interval = setInterval(() => void fetchSummary(), 30_000)
    return () => clearInterval(interval)
  }, [fetchSummary])

  const grouped = summary ? groupByProduct(summary.items) : {}
  const products = Object.keys(grouped).sort()

  // Métricas derivadas para os cards de topo
  const totalEvents = summary?.total_last_24h ?? 0
  const uniqueEventTypes = summary ? new Set(summary.items.map((i) => i.event_type)).size : 0
  const topEvent = summary
    ? summary.items.reduce<EventSummaryItem | null>((best, item) => {
        return !best || item.count > best.count ? item : best
      }, null)
    : null
  const topProductName =
    products.length > 0
      ? products.reduce((best, p) => {
          const sumBest = grouped[best]?.reduce((s, i) => s + i.count, 0) ?? 0
          const sumP = grouped[p]?.reduce((s, i) => s + i.count, 0) ?? 0
          return sumP > sumBest ? p : best
        }, products[0])
      : null

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
            <Wifi className="w-3 h-3" aria-hidden="true" />
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
            <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} aria-hidden="true" />
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

      {/* ── Aba Resumo ─────────────────────────────────────────────────────── */}
      {activeTab === 'resumo' && (
        <div className="space-y-5">

          {/* Cards de métricas */}
          {loading ? (
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
              <SkeletonMetricCard />
              <SkeletonMetricCard />
              <SkeletonMetricCard />
              <SkeletonMetricCard />
            </div>
          ) : (
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">

              {/* Total de eventos */}
              <div className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-4 flex items-start justify-between gap-3">
                <div className="min-w-0">
                  <p className="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wide">
                    Total de eventos
                  </p>
                  <p className="text-3xl font-bold text-slate-800 dark:text-slate-200 mt-1 tabular-nums">
                    {totalEvents.toLocaleString('pt-BR')}
                  </p>
                  <p className="text-xs text-slate-400 dark:text-slate-500 mt-0.5">últimas 24h</p>
                </div>
                <div className="w-9 h-9 rounded-lg bg-violet-50 dark:bg-violet-950/40 flex items-center justify-center shrink-0">
                  <Activity className="w-4.5 h-4.5 text-violet-500" aria-hidden="true" />
                </div>
              </div>

              {/* Tipos únicos */}
              <div className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-4 flex items-start justify-between gap-3">
                <div className="min-w-0">
                  <p className="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wide">
                    Tipos de evento
                  </p>
                  <p className="text-3xl font-bold text-slate-800 dark:text-slate-200 mt-1 tabular-nums">
                    {uniqueEventTypes}
                  </p>
                  <p className="text-xs text-slate-400 dark:text-slate-500 mt-0.5">distintos nas 24h</p>
                </div>
                <div className="w-9 h-9 rounded-lg bg-blue-50 dark:bg-blue-950/40 flex items-center justify-center shrink-0">
                  <BarChart2 className="w-4.5 h-4.5 text-blue-500" aria-hidden="true" />
                </div>
              </div>

              {/* Produto mais ativo */}
              <div className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-4 flex items-start justify-between gap-3">
                <div className="min-w-0">
                  <p className="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wide">
                    Produto mais ativo
                  </p>
                  <p className="text-2xl font-bold text-slate-800 dark:text-slate-200 mt-1 capitalize truncate">
                    {topProductName ?? '—'}
                  </p>
                  <p className="text-xs text-slate-400 dark:text-slate-500 mt-0.5">
                    {topProductName && grouped[topProductName]
                      ? grouped[topProductName].reduce((s, i) => s + i.count, 0).toLocaleString('pt-BR') + ' eventos'
                      : 'sem dados'}
                  </p>
                </div>
                <div className="w-9 h-9 rounded-lg bg-emerald-50 dark:bg-emerald-950/40 flex items-center justify-center shrink-0">
                  <TrendingUp className="w-4.5 h-4.5 text-emerald-500" aria-hidden="true" />
                </div>
              </div>

              {/* Evento mais disparado */}
              <div className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-4 flex items-start justify-between gap-3">
                <div className="min-w-0">
                  <p className="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wide">
                    Evento top
                  </p>
                  <p
                    className="text-lg font-bold text-slate-800 dark:text-slate-200 mt-1 font-mono truncate"
                    title={topEvent?.event_type}
                  >
                    {topEvent?.event_type ?? '—'}
                  </p>
                  <p className="text-xs text-slate-400 dark:text-slate-500 mt-0.5">
                    {topEvent
                      ? `${topEvent.count.toLocaleString('pt-BR')} ocorrências · ${topEvent.product}`
                      : 'sem dados'}
                  </p>
                </div>
                <div className="w-9 h-9 rounded-lg bg-orange-50 dark:bg-orange-950/40 flex items-center justify-center shrink-0">
                  <Zap className="w-4.5 h-4.5 text-orange-500" aria-hidden="true" />
                </div>
              </div>
            </div>
          )}

          {/* Empty state */}
          {!loading && (!summary || summary.items.length === 0) && (
            <div className="rounded-xl border border-dashed border-violet-300 dark:border-violet-700 bg-violet-50/50 dark:bg-violet-950/20 p-12 flex flex-col items-center text-center gap-4">
              <div className="w-16 h-16 rounded-2xl bg-violet-100 dark:bg-violet-900/40 flex items-center justify-center">
                <Activity className="w-8 h-8 text-violet-500 dark:text-violet-400" aria-hidden="true" />
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

          {/* Distribuição por produto — barras horizontais */}
          {products.length > 0 && (
            <div className="space-y-4">
              {products.map((product) => {
                const items = [...grouped[product]].sort((a, b) => b.count - a.count)
                const total = items.reduce((sum, i) => sum + i.count, 0)
                const maxCount = items[0]?.count ?? 0
                const color = getProductColor(product)
                const barColor = getProductBarColor(product)

                return (
                  <div
                    key={product}
                    className={`rounded-xl border bg-gradient-to-br ${color} overflow-hidden`}
                  >
                    {/* Header do produto */}
                    <div className="flex items-center justify-between px-4 pt-4 pb-3">
                      <div className="flex items-center gap-2">
                        <h3 className="font-semibold capitalize text-sm">{product}</h3>
                        <span className="px-1.5 py-0.5 rounded text-xs font-medium bg-white/50 dark:bg-slate-900/30">
                          {items.length} tipo{items.length !== 1 ? 's' : ''}
                        </span>
                      </div>
                      <span className="text-sm font-semibold tabular-nums">
                        {total.toLocaleString('pt-BR')} eventos
                      </span>
                    </div>

                    {/* Barras de distribuição */}
                    <div className="px-4 pb-4 space-y-2.5">
                      {items.map((item) => (
                        <EventBar
                          key={`${item.product}-${item.event_type}`}
                          item={item}
                          maxCount={maxCount}
                          barColorClass={barColor}
                        />
                      ))}
                    </div>
                  </div>
                )
              })}
            </div>
          )}
        </div>
      )}

      {/* ── Aba Usuários ───────────────────────────────────────────────────── */}
      {activeTab === 'usuarios' && <UsersTab />}
    </div>
  )
}
