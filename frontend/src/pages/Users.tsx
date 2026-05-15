import { useState, useEffect, useCallback } from 'react'
import {
  Users as UsersIcon, Search, Flag, Activity, Copy, Check,
  ChevronLeft, ChevronRight, Link2, Clock
} from 'lucide-react'
import { api } from '../services/api'
import { useProject } from '../components/Layout'

// ─── Types ───────────────────────────────────────────────────────────────────

interface UserSummary {
  user_id: string
  event_count: number
  flag_count: number
  last_seen_at: string
  products: string[]
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

// ─── Helpers ─────────────────────────────────────────────────────────────────

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

function avatarColor(userId: string): string {
  const colors = [
    'bg-violet-500', 'bg-emerald-500', 'bg-blue-500', 'bg-amber-500',
    'bg-rose-500', 'bg-cyan-500', 'bg-indigo-500', 'bg-teal-500',
  ]
  let hash = 0
  for (let i = 0; i < userId.length; i++) hash = (hash * 31 + userId.charCodeAt(i)) & 0xffffffff
  return colors[Math.abs(hash) % colors.length]
}

function avatarLetter(userId: string): string {
  return (userId[0] ?? 'U').toUpperCase()
}

// ─── Sub-components ──────────────────────────────────────────────────────────

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

function ProductDot({ product }: { product: string }) {
  const cfg = PRODUCT_COLORS[product]
  if (!cfg) return <span className={`w-2 h-2 rounded-full bg-slate-400`} title={product} />
  return <span className={`w-2 h-2 rounded-full ${cfg.dot}`} title={cfg.label} />
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

function UserCard({
  user, selected, onClick,
}: { user: UserSummary; selected: boolean; onClick: () => void }) {
  const letter = avatarLetter(user.user_id)
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
        <span className={`w-7 h-7 rounded-full flex items-center justify-center text-white text-xs font-bold shrink-0 ${avatarCls}`}>
          {letter}
        </span>
        <span className="font-mono text-xs text-slate-700 dark:text-slate-300 truncate">
          {user.user_id.slice(0, 16)}{user.user_id.length > 16 ? '…' : ''}
        </span>
        <span onClick={e => e.stopPropagation()} className="ml-auto">
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
    ? isTrue
      ? 'bg-emerald-50 dark:bg-emerald-950/20'
      : 'bg-slate-50 dark:bg-slate-800/40'
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
                Feature{' '}
                <span className="font-medium font-mono">{item.flag_key}</span>
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
                Evento{' '}
                <span className="font-medium">{item.event_type}</span>
                {item.event_product && (
                  <span className="text-slate-400 dark:text-slate-500 text-xs ml-1">
                    em {item.event_product}
                  </span>
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

function EmptyPanel() {
  return (
    <div className="flex flex-col items-center justify-center h-full py-24 text-center">
      <UsersIcon className="w-12 h-12 text-slate-200 dark:text-slate-700 mb-3" />
      <p className="text-base font-medium text-slate-400 dark:text-slate-500">Selecione um usuário</p>
      <p className="text-sm text-slate-300 dark:text-slate-600 mt-1">O histórico de atividade aparece aqui</p>
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
  const [timeline, setTimeline] = useState<TimelineItem[]>([])
  const [identities, setIdentities] = useState<Identity[]>([])
  const [detailLoading, setDetailLoading] = useState(false)

  const fetchUsers = useCallback(async () => {
    setLoading(true)
    try {
      const params: Record<string, string | number> = {
        limit: PAGE_SIZE, offset,
        since: sinceFilter,
      }
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
    setDetailLoading(true)
    setTimeline([])
    setIdentities([])
    try {
      const [tl, ids] = await Promise.all([
        api.get<{ items: TimelineItem[] }>(`/users/${encodeURIComponent(user.user_id)}/timeline`, {
          params: { since: sinceFilter, limit: 100 },
        }),
        api.get<{ identities: Identity[] }>(`/users/${encodeURIComponent(user.user_id)}/identities`),
      ])
      setTimeline(tl.data.items ?? [])
      setIdentities(ids.data.identities ?? [])
    } catch {
      // non-fatal
    } finally {
      setDetailLoading(false)
    }
  }

  const totalPages = Math.ceil(total / PAGE_SIZE)
  const currentPage = Math.floor(offset / PAGE_SIZE) + 1

  // ─── Render ────────────────────────────────────────────────────────────────

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
            Comportamento individual por flag e evento, com identidade cross-produto.
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
            placeholder="Buscar por user ID..."
            value={search}
            onChange={e => setSearch(e.target.value)}
            className="pl-8 pr-3 py-2 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-sm w-60 focus:outline-none focus:ring-2 focus:ring-violet-500"
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
        <div className={`w-full md:w-[340px] md:shrink-0 flex flex-col gap-2 ${mobileTab === 'detail' ? 'hidden md:flex' : 'flex'}`}>
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
              <div className="border-b border-slate-100 dark:border-slate-800 px-5 py-4">
                <div className="flex items-center gap-2 mb-2">
                  <span className={`w-8 h-8 rounded-full flex items-center justify-center text-white text-sm font-bold ${avatarColor(selectedUser.user_id)}`}>
                    {avatarLetter(selectedUser.user_id)}
                  </span>
                  <span className="font-mono text-sm text-slate-800 dark:text-slate-200 break-all">
                    {selectedUser.user_id}
                  </span>
                  <CopyButton text={selectedUser.user_id} />
                </div>

                <div className="flex flex-wrap items-center gap-2 ml-10">
                  {identities.length > 0 && (
                    <>
                      <span className="flex items-center gap-1 text-xs text-slate-400 dark:text-slate-500">
                        <Link2 className="w-3 h-3" />
                        Identidade unificada
                      </span>
                      {identities.map(id => (
                        <ProductBadge key={`${id.product}-${id.product_user_id}`} product={id.product} />
                      ))}
                    </>
                  )}
                  {selectedUser.products.filter(p => !identities.find(id => id.product === p)).map(p => (
                    <ProductBadge key={p} product={p} />
                  ))}
                </div>
              </div>

              {/* Timeline */}
              <div className="flex-1 overflow-y-auto px-5 py-4">
                {detailLoading ? (
                  <div className="space-y-3">
                    {Array.from({ length: 4 }).map((_, i) => (
                      <div key={i} className="border-l-2 border-slate-200 dark:border-slate-700 pl-4 pb-4 animate-pulse">
                        <div className="h-10 bg-slate-100 dark:bg-slate-800 rounded-lg" />
                      </div>
                    ))}
                  </div>
                ) : timeline.length === 0 ? (
                  <div className="text-center py-12 text-slate-400 text-sm">
                    <Activity className="w-8 h-8 mx-auto mb-2 text-slate-200 dark:text-slate-700" />
                    Nenhuma atividade encontrada para este usuário.
                  </div>
                ) : (
                  <div>
                    <p className="text-xs font-semibold uppercase tracking-widest text-slate-400 dark:text-slate-500 mb-4">
                      Timeline de atividade
                    </p>
                    {timeline.map((item, i) => (
                      <TimelineEntry key={i} item={item} />
                    ))}
                  </div>
                )}
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  )
}
