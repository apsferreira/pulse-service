import { useState, useEffect, useCallback } from 'react'
import {
  Users as UsersIcon, Search, Flag, Activity, Copy, Check,
  ChevronLeft, ChevronRight, Link2, Clock, UserCheck, UserX,
  Video, MousePointer, RefreshCw, MessageSquare,
} from 'lucide-react'
import { api } from '../services/api'
import { useProject } from '../components/Layout'
import { replayApi, eventsApi, type ReplaySession, type RawEvent } from '../services/pulseApi'

// ─── Types ────────────────────────────────────────────────────────────────────

interface UserSummary {
  user_id: string
  event_count: number
  flag_count: number
  last_seen_at: string
  products: string[]
  is_anonymous: boolean
  name?: string
  email?: string
  avatar_url?: string
}

interface TimelineItem {
  kind: 'flag' | 'event'
  timestamp: string
  flag_key?: string
  flag_result?: boolean
  flag_reason?: string
  event_type?: string
  event_product?: string
  properties?: Record<string, unknown>
}

interface Identity {
  canonical_user_id: string
  product: string
  product_user_id: string
  email?: string
}

interface UserProfile {
  user_id: string
  is_anonymous: boolean
  name?: string
  email?: string
  avatar_url?: string
  identities: Identity[]
}

interface CSATResponse {
  id: string
  survey_id: string
  score: number
  comment?: string
  respondent?: string
  created_at: string
}

type DetailTab = 'perfil' | 'atividade' | 'eventos' | 'replay' | 'csat'

// ─── Constants ───────────────────────────────────────────────────────────────

const PRODUCT_COLORS: Record<string, { dot: string; badge: string; label: string }> = {
  nitro:      { dot: 'bg-violet-500',  badge: 'bg-violet-50 text-violet-700 border-violet-200 dark:bg-violet-950/40 dark:text-violet-300 dark:border-violet-800', label: 'Nitro' },
  libri:      { dot: 'bg-emerald-500', badge: 'bg-emerald-50 text-emerald-700 border-emerald-200 dark:bg-emerald-950/40 dark:text-emerald-300 dark:border-emerald-800', label: 'Librinho' },
  socialmake: { dot: 'bg-blue-500',    badge: 'bg-blue-50 text-blue-700 border-blue-200 dark:bg-blue-950/40 dark:text-blue-300 dark:border-blue-800', label: 'SocialMake' },
  brio:       { dot: 'bg-amber-500',   badge: 'bg-amber-50 text-amber-700 border-amber-200 dark:bg-amber-950/40 dark:text-amber-300 dark:border-amber-800', label: 'Brio' },
}

const REASON_LABELS: Record<string, string> = {
  kill_switch:      'Kill Switch',
  targeting_user:   'Targeting (usuário)',
  targeting_tenant: 'Targeting (tenant)',
  rollout:          'Rollout %',
  default:          'Padrão',
  disabled:         'Desabilitada',
}

const SINCE_OPTIONS = [
  { value: '24h', label: 'Últimas 24h' },
  { value: '7d',  label: 'Últimos 7 dias' },
  { value: '30d', label: 'Últimos 30 dias' },
]

const DETAIL_TABS: { id: DetailTab; label: string }[] = [
  { id: 'perfil',    label: 'Perfil' },
  { id: 'atividade', label: 'Atividade' },
  { id: 'eventos',   label: 'Eventos' },
  { id: 'replay',    label: 'Replay' },
  { id: 'csat',      label: 'CSAT' },
]

// ─── Helpers ──────────────────────────────────────────────────────────────────

function relativeTime(iso: string): string {
  const diff = Date.now() - new Date(iso).getTime()
  const m = Math.floor(diff / 60000)
  if (m < 1) return 'agora'
  if (m < 60) return `há ${m} min`
  const h = Math.floor(m / 60)
  if (h < 24) return `há ${h}h`
  const d = Math.floor(h / 24)
  return `há ${d}d`
}

function formatDate(iso: string): string {
  return new Intl.DateTimeFormat('pt-BR', {
    day: '2-digit', month: '2-digit', year: 'numeric',
    hour: '2-digit', minute: '2-digit',
  }).format(new Date(iso))
}

function formatDuration(ms: number): string {
  const s = Math.floor(ms / 1000)
  const m = Math.floor(s / 60)
  if (m === 0) return `${s}s`
  return `${m}m ${s % 60}s`
}

function avatarColor(userId: string): string {
  const colors = [
    'bg-violet-500', 'bg-emerald-500', 'bg-blue-500', 'bg-amber-500',
    'bg-rose-500', 'bg-cyan-500', 'bg-indigo-500', 'bg-teal-500',
  ]
  let hash = 0
  for (let i = 0; i < userId.length; i++) hash = (hash * 31 + userId.charCodeAt(i)) & 0xffffffff
  return colors[Math.abs(hash) % colors.length]
}

function avatarInitial(user: UserSummary): string {
  if (user.name) return user.name[0].toUpperCase()
  return (user.user_id[0] ?? 'U').toUpperCase()
}

function displayName(user: UserSummary): string {
  if (user.name) return user.name
  if (user.is_anonymous) return 'Visitante'
  return user.user_id.slice(0, 8) + '…'
}

function truncateUrl(url: string, max = 40): string {
  try {
    const u = new URL(url)
    const path = u.pathname + u.search
    if (path.length <= max) return path
    return path.slice(0, max) + '…'
  } catch {
    return url.length <= max ? url : url.slice(0, max) + '…'
  }
}

// ─── Sub-components ───────────────────────────────────────────────────────────

function CopyButton({ text }: { text: string }) {
  const [copied, setCopied] = useState(false)
  const copy = async () => {
    await navigator.clipboard.writeText(text)
    setCopied(true)
    setTimeout(() => setCopied(false), 1500)
  }
  return (
    <button onClick={copy} className="p-0.5 rounded text-slate-300 hover:text-slate-500 dark:text-slate-600 dark:hover:text-slate-400 transition-colors" title="Copiar">
      {copied ? <Check className="w-3 h-3 text-emerald-500" /> : <Copy className="w-3 h-3" />}
    </button>
  )
}

function ProductBadge({ product }: { product: string }) {
  const cfg = PRODUCT_COLORS[product]
  const label = cfg?.label ?? product
  const cls = cfg?.badge ?? 'bg-slate-100 text-slate-600 border-slate-200 dark:bg-slate-800 dark:text-slate-400 dark:border-slate-700'
  return (
    <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium border ${cls}`}>
      {label}
    </span>
  )
}

function ProductDot({ product }: { product: string }) {
  const cfg = PRODUCT_COLORS[product]
  if (!cfg) return <span className="w-2 h-2 rounded-full bg-slate-400" title={product} />
  return <span className={`w-2 h-2 rounded-full ${cfg.dot}`} title={cfg.label} />
}

function UserCard({
  user, selected, onClick,
}: { user: UserSummary; selected: boolean; onClick: () => void }) {
  const avatarCls = avatarColor(user.user_id)
  return (
    <button
      onClick={onClick}
      className={[
        'w-full text-left rounded-xl border px-3 py-3 transition-all',
        selected
          ? 'bg-violet-50 border-violet-300 dark:bg-violet-950/30 dark:border-violet-700'
          : 'bg-white border-slate-200 hover:bg-slate-50 dark:bg-slate-900 dark:border-slate-800 dark:hover:bg-slate-800/60',
      ].join(' ')}
    >
      <div className="flex items-center gap-2 mb-1.5">
        {user.avatar_url ? (
          <img src={user.avatar_url} alt="" className="w-7 h-7 rounded-full object-cover shrink-0" />
        ) : (
          <span className={`w-7 h-7 rounded-full flex items-center justify-center text-white text-xs font-bold shrink-0 ${avatarCls}`}>
            {avatarInitial(user)}
          </span>
        )}
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-1">
            {user.is_anonymous ? (
              <UserX className="w-3 h-3 text-slate-400 shrink-0" />
            ) : (
              <UserCheck className="w-3 h-3 text-emerald-500 shrink-0" />
            )}
            <span className="text-xs font-medium text-slate-700 dark:text-slate-300 truncate">
              {displayName(user)}
            </span>
          </div>
          {user.email && (
            <span className="text-xs text-slate-400 dark:text-slate-500 truncate block">{user.email}</span>
          )}
          {user.is_anonymous && (
            <span className="font-mono text-xs text-slate-400 dark:text-slate-600 truncate block">
              {user.user_id.slice(0, 14)}…
            </span>
          )}
        </div>
        <span onClick={e => e.stopPropagation()} className="ml-auto shrink-0">
          <CopyButton text={user.user_id} />
        </span>
      </div>

      {user.products.length > 0 && (
        <div className="flex items-center gap-1 mb-1.5 ml-9">
          {user.products.map(p => <ProductDot key={p} product={p} />)}
        </div>
      )}

      <div className="flex items-center gap-2 ml-9 text-xs text-slate-400 dark:text-slate-500">
        <span>{user.event_count} eventos</span>
        <span>·</span>
        <span>{user.flag_count} flags</span>
        <span className="ml-auto">{relativeTime(user.last_seen_at)}</span>
      </div>
    </button>
  )
}

function TimelineEntry({ item }: { item: TimelineItem }) {
  const isFlag = item.kind === 'flag'
  const isTrue = item.flag_result === true

  const borderColor = isFlag
    ? isTrue ? 'border-l-emerald-500' : 'border-l-slate-400'
    : 'border-l-violet-400'
  const bg = isFlag
    ? isTrue ? 'bg-emerald-50 dark:bg-emerald-950/20' : 'bg-slate-50 dark:bg-slate-800/40'
    : 'bg-violet-50 dark:bg-violet-950/20'
  const dot = isFlag
    ? isTrue ? 'bg-emerald-500' : 'bg-slate-400'
    : 'bg-violet-400'

  return (
    <div className={`relative border-l-2 pl-4 pb-4 last:pb-0 ${borderColor}`}>
      <span className={`absolute -left-[5px] top-3 w-2 h-2 rounded-full ${dot}`} />
      <div className={`rounded-lg px-3 py-2.5 ${bg}`}>
        <div className="flex items-start justify-between gap-2">
          <div className="text-sm text-slate-700 dark:text-slate-300">
            {isFlag ? (
              <>
                Flag <span className="font-medium font-mono">{item.flag_key}</span>
                {' '}avaliada como{' '}
                <span className={`font-semibold ${isTrue ? 'text-emerald-600 dark:text-emerald-400' : 'text-slate-500'}`}>
                  {isTrue ? 'true' : 'false'}
                </span>
                {item.flag_reason && (
                  <span className="text-slate-400 dark:text-slate-500 text-xs ml-1">
                    via {REASON_LABELS[item.flag_reason] ?? item.flag_reason}
                  </span>
                )}
              </>
            ) : (
              <>
                <Flag className="inline w-3.5 h-3.5 text-violet-400 mr-1 -mt-0.5" />
                Evento <span className="font-medium">{item.event_type}</span>
                {item.event_product && (
                  <span className="text-slate-400 dark:text-slate-500 text-xs ml-1">em {item.event_product}</span>
                )}
              </>
            )}
          </div>
          <span className="text-xs text-slate-400 dark:text-slate-500 whitespace-nowrap flex items-center gap-1 shrink-0">
            <Clock className="w-3 h-3" />
            {relativeTime(item.timestamp)}
          </span>
        </div>
      </div>
    </div>
  )
}

// ─── Detail Tabs ──────────────────────────────────────────────────────────────

function TabPerfil({ profile, identities }: { profile: UserProfile | null; identities: Identity[] }) {
  if (!profile) {
    return <div className="text-center py-12 text-slate-400 text-sm">Carregando perfil…</div>
  }
  return (
    <div className="space-y-5">
      {/* Identity card */}
      <div className="rounded-xl border border-slate-200 dark:border-slate-800 p-4 space-y-3">
        <p className="text-xs font-semibold uppercase tracking-widest text-slate-400 dark:text-slate-500">Identidade</p>
        <div className="flex items-center gap-3">
          {profile.avatar_url ? (
            <img src={profile.avatar_url} alt="" className="w-10 h-10 rounded-full object-cover" />
          ) : (
            <span className={`w-10 h-10 rounded-full flex items-center justify-center text-white font-bold ${avatarColor(profile.user_id)}`}>
              {profile.name ? profile.name[0].toUpperCase() : (profile.user_id[0] ?? 'U').toUpperCase()}
            </span>
          )}
          <div>
            {profile.name && <p className="font-semibold text-slate-800 dark:text-slate-100">{profile.name}</p>}
            {profile.email && <p className="text-sm text-slate-500 dark:text-slate-400">{profile.email}</p>}
            {!profile.name && !profile.email && (
              <p className="text-sm text-slate-400 dark:text-slate-500 italic">
                {profile.is_anonymous ? 'Visitante anônimo' : 'Sem dados de perfil no auth-service'}
              </p>
            )}
          </div>
        </div>
        <div className="pt-1 border-t border-slate-100 dark:border-slate-800">
          <p className="text-xs text-slate-400 dark:text-slate-500 mb-1">User ID</p>
          <div className="flex items-center gap-1">
            <span className="font-mono text-xs text-slate-600 dark:text-slate-400 break-all">{profile.user_id}</span>
            <CopyButton text={profile.user_id} />
          </div>
        </div>
        <div className="flex items-center gap-1.5">
          {profile.is_anonymous ? (
            <><UserX className="w-3.5 h-3.5 text-slate-400" /><span className="text-xs text-slate-400">Anônimo</span></>
          ) : (
            <><UserCheck className="w-3.5 h-3.5 text-emerald-500" /><span className="text-xs text-emerald-600 dark:text-emerald-400">Identificado</span></>
          )}
        </div>
      </div>

      {/* Cross-product identities */}
      {identities.length > 0 && (
        <div className="rounded-xl border border-slate-200 dark:border-slate-800 p-4 space-y-3">
          <p className="text-xs font-semibold uppercase tracking-widest text-slate-400 dark:text-slate-500 flex items-center gap-1.5">
            <Link2 className="w-3.5 h-3.5" />
            Identidades cross-produto
          </p>
          <div className="space-y-2">
            {identities.map(id => (
              <div key={`${id.product}-${id.product_user_id}`} className="flex items-center justify-between gap-2 text-sm">
                <ProductBadge product={id.product} />
                <div className="flex items-center gap-1 min-w-0">
                  <span className="font-mono text-xs text-slate-500 dark:text-slate-400 truncate">{id.product_user_id}</span>
                  <CopyButton text={id.product_user_id} />
                </div>
                {id.email && <span className="text-xs text-slate-400 truncate">{id.email}</span>}
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}

function TabAtividade({ timeline, loading }: { timeline: TimelineItem[]; loading: boolean }) {
  if (loading) {
    return (
      <div className="space-y-3">
        {Array.from({ length: 4 }).map((_, i) => (
          <div key={i} className="border-l-2 border-slate-200 dark:border-slate-700 pl-4 pb-4 animate-pulse">
            <div className="h-10 bg-slate-100 dark:bg-slate-800 rounded-lg" />
          </div>
        ))}
      </div>
    )
  }
  if (timeline.length === 0) {
    return (
      <div className="text-center py-12 text-slate-400 text-sm">
        <Activity className="w-8 h-8 mx-auto mb-2 text-slate-200 dark:text-slate-700" />
        Nenhuma atividade encontrada.
      </div>
    )
  }
  return (
    <div>
      <p className="text-xs font-semibold uppercase tracking-widest text-slate-400 dark:text-slate-500 mb-4">
        Flags + Eventos combinados
      </p>
      {timeline.map((item, i) => <TimelineEntry key={i} item={item} />)}
    </div>
  )
}

function TabEventos({ userID, sinceFilter }: { userID: string; sinceFilter: string }) {
  const [events, setEvents] = useState<RawEvent[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    setLoading(true)
    eventsApi.list({ user_id: userID, since: sinceFilter, limit: 100 })
      .then(r => setEvents(r.data.items ?? []))
      .catch(() => setEvents([]))
      .finally(() => setLoading(false))
  }, [userID, sinceFilter])

  if (loading) {
    return (
      <div className="space-y-2">
        {Array.from({ length: 5 }).map((_, i) => (
          <div key={i} className="h-14 bg-slate-100 dark:bg-slate-800 rounded-lg animate-pulse" />
        ))}
      </div>
    )
  }
  if (events.length === 0) {
    return (
      <div className="text-center py-12 text-slate-400 text-sm">
        <Flag className="w-8 h-8 mx-auto mb-2 text-slate-200 dark:text-slate-700" />
        Nenhum evento registrado.
      </div>
    )
  }
  return (
    <div className="space-y-2">
      {events.map(ev => (
        <div key={ev.id} className="rounded-lg border border-slate-100 dark:border-slate-800 bg-white dark:bg-slate-900 px-3 py-2.5">
          <div className="flex items-center justify-between gap-2 mb-1">
            <span className="font-medium text-sm text-slate-700 dark:text-slate-300">{ev.event_type}</span>
            <span className="text-xs text-slate-400 shrink-0">{relativeTime(ev.created_at)}</span>
          </div>
          <div className="flex items-center gap-2">
            <ProductBadge product={ev.product} />
            <span className="text-xs text-slate-400">{ev.environment}</span>
          </div>
          {Object.keys(ev.properties ?? {}).length > 0 && (
            <pre className="mt-2 text-xs text-slate-500 dark:text-slate-400 bg-slate-50 dark:bg-slate-800 rounded p-2 overflow-x-auto">
              {JSON.stringify(ev.properties, null, 2)}
            </pre>
          )}
        </div>
      ))}
    </div>
  )
}

function TabReplay({ userID, projectId }: { userID: string; projectId: string | null }) {
  const [sessions, setSessions] = useState<ReplaySession[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    setLoading(true)
    replayApi.sessions({ user_id: userID, ...(projectId ? { project_id: projectId } : {}) })
      .then(r => {
        const body = r.data as { sessions?: ReplaySession[] } | ReplaySession[]
        if (Array.isArray(body)) setSessions(body)
        else setSessions((body as { sessions?: ReplaySession[] }).sessions ?? [])
      })
      .catch(() => setSessions([]))
      .finally(() => setLoading(false))
  }, [userID, projectId])

  if (loading) {
    return (
      <div className="space-y-2">
        {Array.from({ length: 3 }).map((_, i) => (
          <div key={i} className="h-16 bg-slate-100 dark:bg-slate-800 rounded-lg animate-pulse" />
        ))}
      </div>
    )
  }
  if (sessions.length === 0) {
    return (
      <div className="text-center py-12 text-slate-400 text-sm">
        <Video className="w-8 h-8 mx-auto mb-2 text-slate-200 dark:text-slate-700" />
        Nenhuma sessão gravada.
      </div>
    )
  }
  return (
    <div className="space-y-2">
      {sessions.map(s => (
        <div key={s.session_id} className="rounded-lg border border-slate-100 dark:border-slate-800 bg-white dark:bg-slate-900 px-3 py-2.5">
          <div className="flex items-center justify-between gap-2 mb-1">
            <div className="flex items-center gap-2">
              <Video className="w-3.5 h-3.5 text-violet-400 shrink-0" />
              <span className="font-mono text-xs text-slate-600 dark:text-slate-400">{s.session_id.slice(0, 12)}…</span>
            </div>
            <span className="text-xs text-slate-400 shrink-0">{relativeTime(s.started_at)}</span>
          </div>
          <div className="flex items-center gap-3 text-xs text-slate-400 dark:text-slate-500 ml-5">
            <span className="flex items-center gap-1">
              <MousePointer className="w-3 h-3" />
              {s.event_count} eventos
            </span>
            <span>{formatDuration(s.duration_ms)}</span>
            {s.page_url && <span className="truncate max-w-[140px]">{truncateUrl(s.page_url)}</span>}
          </div>
        </div>
      ))}
    </div>
  )
}

function TabCSAT({ userID }: { userID: string }) {
  const [responses, setResponses] = useState<CSATResponse[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    setLoading(true)
    api.get<{ responses: CSATResponse[] }>('/csat/responses', { params: { respondent: userID, limit: 50 } })
      .then(r => setResponses(r.data.responses ?? []))
      .catch(() => setResponses([]))
      .finally(() => setLoading(false))
  }, [userID])

  if (loading) {
    return (
      <div className="space-y-2">
        {Array.from({ length: 3 }).map((_, i) => (
          <div key={i} className="h-14 bg-slate-100 dark:bg-slate-800 rounded-lg animate-pulse" />
        ))}
      </div>
    )
  }
  if (responses.length === 0) {
    return (
      <div className="text-center py-12 text-slate-400 text-sm">
        <MessageSquare className="w-8 h-8 mx-auto mb-2 text-slate-200 dark:text-slate-700" />
        Nenhuma resposta de survey registrada.
      </div>
    )
  }

  const scoreColor = (score: number, scale = 10) => {
    const pct = score / scale
    if (pct >= 0.9) return 'text-emerald-600 dark:text-emerald-400'
    if (pct >= 0.7) return 'text-amber-600 dark:text-amber-400'
    return 'text-rose-600 dark:text-rose-400'
  }

  return (
    <div className="space-y-2">
      {responses.map(r => (
        <div key={r.id} className="rounded-lg border border-slate-100 dark:border-slate-800 bg-white dark:bg-slate-900 px-3 py-2.5">
          <div className="flex items-center justify-between gap-2 mb-1">
            <span className={`text-2xl font-bold ${scoreColor(r.score)}`}>{r.score}</span>
            <span className="text-xs text-slate-400 shrink-0">{formatDate(r.created_at)}</span>
          </div>
          {r.comment && (
            <p className="text-sm text-slate-600 dark:text-slate-400 mt-1">{r.comment}</p>
          )}
        </div>
      ))}
    </div>
  )
}

// ─── Empty Panel ─────────────────────────────────────────────────────────────

function EmptyPanel() {
  return (
    <div className="flex flex-col items-center justify-center h-full py-24 text-center">
      <UsersIcon className="w-12 h-12 text-slate-200 dark:text-slate-700 mb-3" />
      <p className="text-base font-medium text-slate-400 dark:text-slate-500">Selecione um usuário</p>
      <p className="text-sm text-slate-300 dark:text-slate-600 mt-1">O perfil completo aparece aqui</p>
    </div>
  )
}

// ─── Main Page ────────────────────────────────────────────────────────────────

const PAGE_SIZE = 50

export default function UsersPage() {
  const { projectId } = useProject()

  // List state
  const [users, setUsers] = useState<UserSummary[]>([])
  const [total, setTotal] = useState(0)
  const [offset, setOffset] = useState(0)
  const [loading, setLoading] = useState(true)
  const [search, setSearch] = useState('')
  const [sinceFilter, setSinceFilter] = useState('7d')
  const [productFilter, setProductFilter] = useState('')
  const [mobileTab, setMobileTab] = useState<'list' | 'detail'>('list')

  // Detail state
  const [selectedUser, setSelectedUser] = useState<UserSummary | null>(null)
  const [profile, setProfile] = useState<UserProfile | null>(null)
  const [timeline, setTimeline] = useState<TimelineItem[]>([])
  const [identities, setIdentities] = useState<Identity[]>([])
  const [detailLoading, setDetailLoading] = useState(false)
  const [detailTab, setDetailTab] = useState<DetailTab>('perfil')

  const fetchUsers = useCallback(async () => {
    setLoading(true)
    try {
      const params: Record<string, string | number> = { limit: PAGE_SIZE, offset, since: sinceFilter }
      if (search) params.search = search
      if (projectId) params.project_id = projectId
      if (productFilter) params.product = productFilter

      const { data } = await api.get<{ users: UserSummary[]; total: number }>('/users', { params })
      setUsers(data.users ?? [])
      setTotal(data.total ?? 0)
    } catch {
      setUsers([])
    } finally {
      setLoading(false)
    }
  }, [projectId, offset, search, sinceFilter, productFilter])

  useEffect(() => { void fetchUsers() }, [fetchUsers])
  useEffect(() => { setOffset(0) }, [projectId, search, sinceFilter, productFilter])

  const selectUser = async (user: UserSummary) => {
    setSelectedUser(user)
    setMobileTab('detail')
    setDetailTab('perfil')
    setDetailLoading(true)
    setProfile(null)
    setTimeline([])
    setIdentities([])
    try {
      const [profileRes, tlRes, idsRes] = await Promise.all([
        api.get<UserProfile>(`/users/${encodeURIComponent(user.user_id)}/profile`),
        api.get<{ items: TimelineItem[] }>(`/users/${encodeURIComponent(user.user_id)}/timeline`, {
          params: { since: sinceFilter, limit: 100 },
        }),
        api.get<{ identities: Identity[] }>(`/users/${encodeURIComponent(user.user_id)}/identities`),
      ])
      setProfile(profileRes.data)
      setTimeline(tlRes.data.items ?? [])
      setIdentities(idsRes.data.identities ?? [])
    } catch {
      // non-fatal
    } finally {
      setDetailLoading(false)
    }
  }

  const totalPages = Math.ceil(total / PAGE_SIZE)
  const currentPage = Math.floor(offset / PAGE_SIZE) + 1

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight flex items-center gap-2">
            <UsersIcon className="w-6 h-6 text-violet-500" />
            Usuários
          </h1>
          <p className="text-slate-500 dark:text-slate-400 text-sm mt-1">
            Perfil, flags, eventos, replays e NPS por usuário.
          </p>
        </div>
        <span className="text-sm text-slate-400 dark:text-slate-500">
          {total.toLocaleString('pt-BR')} usuários
        </span>
      </div>

      {/* Filters */}
      <div className="flex flex-wrap items-center gap-2">
        <div className="relative">
          <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-slate-400" />
          <input
            type="text"
            placeholder="Buscar por nome, email ou ID…"
            value={search}
            onChange={e => setSearch(e.target.value)}
            className="pl-8 pr-3 py-2 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-sm w-64 focus:outline-none focus:ring-2 focus:ring-violet-500"
          />
        </div>

        <select
          value={productFilter}
          onChange={e => setProductFilter(e.target.value)}
          className="px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-sm focus:outline-none focus:ring-2 focus:ring-violet-500"
        >
          <option value="">Todos os produtos</option>
          {Object.entries(PRODUCT_COLORS).map(([k, v]) => (
            <option key={k} value={k}>{v.label}</option>
          ))}
        </select>

        <select
          value={sinceFilter}
          onChange={e => setSinceFilter(e.target.value)}
          className="px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-sm focus:outline-none focus:ring-2 focus:ring-violet-500"
        >
          {SINCE_OPTIONS.map(o => <option key={o.value} value={o.value}>{o.label}</option>)}
        </select>
      </div>

      {/* Mobile tabs */}
      <div className="md:hidden flex rounded-lg border border-slate-200 dark:border-slate-700 overflow-hidden text-sm font-medium">
        <button
          onClick={() => setMobileTab('list')}
          className={`flex-1 py-2 transition-colors ${mobileTab === 'list' ? 'bg-violet-600 text-white' : 'bg-white dark:bg-slate-900 text-slate-600 dark:text-slate-400'}`}
        >
          Usuários ({total})
        </button>
        <button
          onClick={() => setMobileTab('detail')}
          disabled={!selectedUser}
          className={`flex-1 py-2 transition-colors disabled:opacity-40 ${mobileTab === 'detail' ? 'bg-violet-600 text-white' : 'bg-white dark:bg-slate-900 text-slate-600 dark:text-slate-400'}`}
        >
          Detalhe
        </button>
      </div>

      {/* Split panel */}
      <div className="flex gap-4 min-h-[600px]">

        {/* Left — user list */}
        <div className={`w-full md:w-[320px] md:shrink-0 flex flex-col gap-2 ${mobileTab === 'detail' ? 'hidden md:flex' : 'flex'}`}>
          {loading && users.length === 0 ? (
            Array.from({ length: 5 }).map((_, i) => (
              <div key={i} className="rounded-xl border border-slate-100 dark:border-slate-800 bg-white dark:bg-slate-900 p-3 animate-pulse">
                <div className="flex gap-2 mb-2">
                  <div className="w-7 h-7 rounded-full bg-slate-200 dark:bg-slate-700" />
                  <div className="h-3 bg-slate-200 dark:bg-slate-700 rounded w-32 mt-2" />
                </div>
                <div className="h-2 bg-slate-100 dark:bg-slate-800 rounded w-24 ml-9" />
              </div>
            ))
          ) : users.length === 0 ? (
            <div className="text-center py-12 text-slate-400 text-sm">Nenhum usuário encontrado.</div>
          ) : (
            users.map(u => (
              <UserCard
                key={u.user_id}
                user={u}
                selected={selectedUser?.user_id === u.user_id}
                onClick={() => selectUser(u)}
              />
            ))
          )}

          {totalPages > 1 && (
            <div className="flex items-center justify-between pt-2 text-sm text-slate-500 dark:text-slate-400">
              <span>Pág. {currentPage}/{totalPages}</span>
              <div className="flex gap-1">
                <button
                  onClick={() => setOffset(Math.max(0, offset - PAGE_SIZE))}
                  disabled={offset === 0}
                  className="p-1.5 rounded hover:bg-slate-100 dark:hover:bg-slate-800 disabled:opacity-40 transition-colors"
                >
                  <ChevronLeft className="w-4 h-4" />
                </button>
                <button
                  onClick={() => setOffset(offset + PAGE_SIZE)}
                  disabled={offset + PAGE_SIZE >= total}
                  className="p-1.5 rounded hover:bg-slate-100 dark:hover:bg-slate-800 disabled:opacity-40 transition-colors"
                >
                  <ChevronRight className="w-4 h-4" />
                </button>
              </div>
            </div>
          )}
        </div>

        {/* Right — detail panel */}
        <div className={`flex-1 rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 overflow-hidden flex flex-col ${mobileTab === 'list' ? 'hidden md:flex' : 'flex'}`}>
          {!selectedUser ? (
            <EmptyPanel />
          ) : (
            <>
              {/* Panel header */}
              <div className="border-b border-slate-100 dark:border-slate-800 px-5 pt-4 pb-0">
                <div className="flex items-center gap-2 mb-3">
                  {selectedUser.avatar_url ? (
                    <img src={selectedUser.avatar_url} alt="" className="w-8 h-8 rounded-full object-cover shrink-0" />
                  ) : (
                    <span className={`w-8 h-8 rounded-full flex items-center justify-center text-white text-sm font-bold ${avatarColor(selectedUser.user_id)}`}>
                      {avatarInitial(selectedUser)}
                    </span>
                  )}
                  <div className="min-w-0">
                    <div className="flex items-center gap-1.5">
                      {selectedUser.is_anonymous
                        ? <UserX className="w-3.5 h-3.5 text-slate-400 shrink-0" />
                        : <UserCheck className="w-3.5 h-3.5 text-emerald-500 shrink-0" />
                      }
                      <span className="font-semibold text-sm text-slate-800 dark:text-slate-100 truncate">
                        {selectedUser.name ?? (selectedUser.is_anonymous ? 'Visitante anônimo' : displayName(selectedUser))}
                      </span>
                    </div>
                    {selectedUser.email && (
                      <p className="text-xs text-slate-500 dark:text-slate-400">{selectedUser.email}</p>
                    )}
                    {!selectedUser.email && (
                      <p className="font-mono text-xs text-slate-400 dark:text-slate-600 truncate">{selectedUser.user_id}</p>
                    )}
                  </div>
                  <div className="ml-auto flex items-center gap-1 shrink-0">
                    {detailLoading && <RefreshCw className="w-3.5 h-3.5 animate-spin text-slate-400" />}
                    <CopyButton text={selectedUser.user_id} />
                  </div>
                </div>

                {/* Tab bar */}
                <div className="flex border-b-0 gap-0 -mb-px overflow-x-auto">
                  {DETAIL_TABS.map(tab => (
                    <button
                      key={tab.id}
                      onClick={() => setDetailTab(tab.id)}
                      className={[
                        'px-4 py-2 text-sm font-medium border-b-2 whitespace-nowrap transition-colors',
                        detailTab === tab.id
                          ? 'border-violet-500 text-violet-600 dark:text-violet-400'
                          : 'border-transparent text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-300',
                      ].join(' ')}
                    >
                      {tab.label}
                    </button>
                  ))}
                </div>
              </div>

              {/* Tab content */}
              <div className="flex-1 overflow-y-auto px-5 py-4">
                {detailTab === 'perfil' && (
                  <TabPerfil profile={profile} identities={identities} />
                )}
                {detailTab === 'atividade' && (
                  <TabAtividade timeline={timeline} loading={detailLoading} />
                )}
                {detailTab === 'eventos' && (
                  <TabEventos userID={selectedUser.user_id} sinceFilter={sinceFilter} />
                )}
                {detailTab === 'replay' && (
                  <TabReplay userID={selectedUser.user_id} projectId={projectId} />
                )}
                {detailTab === 'csat' && (
                  <TabCSAT userID={selectedUser.user_id} />
                )}
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  )
}
