import { useCallback, useEffect, useRef, useState } from 'react'
import {
  RefreshCw,
  CheckCircle2,
  XCircle,
  Clock,
  AlertCircle,
  Plus,
  Pencil,
  Trash2,
  X,
} from 'lucide-react'
import toast from 'react-hot-toast'
import {
  monitorsApi,
  type Monitor,
  type MonitorStatus,
  type MonitorCheck,
  type CreateMonitorPayload,
  type UpdateMonitorPayload,
} from '../services/pulseApi'

// ── Sub-components ────────────────────────────────────────────────────────────

function StatusBadge({ status }: { status?: string }) {
  if (!status) {
    return (
      <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium bg-slate-100 text-slate-500 dark:bg-slate-800 dark:text-slate-400">
        <AlertCircle className="w-3 h-3" />
        sem dados
      </span>
    )
  }
  if (status === 'up') {
    return (
      <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-400">
        <CheckCircle2 className="w-3 h-3" />
        online
      </span>
    )
  }
  if (status === 'timeout') {
    return (
      <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium bg-amber-100 text-amber-700 dark:bg-amber-900/40 dark:text-amber-400">
        <Clock className="w-3 h-3" />
        timeout
      </span>
    )
  }
  return (
    <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium bg-red-100 text-red-700 dark:bg-red-900/40 dark:text-red-400">
      <XCircle className="w-3 h-3" />
      offline
    </span>
  )
}

function UptimeBar({ uptime }: { uptime: number }) {
  const pct = Math.min(100, Math.max(0, uptime))
  const color =
    pct >= 99 ? 'bg-emerald-500' : pct >= 95 ? 'bg-amber-400' : 'bg-red-500'

  return (
    <div className="flex items-center gap-2">
      <div className="flex-1 h-1.5 rounded-full bg-slate-200 dark:bg-slate-700 overflow-hidden">
        <div className={`h-full rounded-full ${color}`} style={{ width: `${pct}%` }} />
      </div>
      <span className="text-xs text-slate-500 dark:text-slate-400 w-12 text-right tabular-nums">
        {pct.toFixed(1)}%
      </span>
    </div>
  )
}

function formatMs(ms?: number): string {
  if (ms === undefined || ms === null) return '—'
  if (ms < 1000) return `${Math.round(ms)}ms`
  return `${(ms / 1000).toFixed(1)}s`
}

function HistoryDots({ checks }: { checks: MonitorCheck[] }) {
  const recent = checks.slice(0, 30).reverse()
  return (
    <div className="flex gap-0.5 flex-wrap">
      {recent.map((c) => {
        const color =
          c.status === 'up'
            ? 'bg-emerald-500'
            : c.status === 'timeout'
            ? 'bg-amber-400'
            : 'bg-red-500'
        return (
          <div
            key={c.id}
            title={`${c.status} — ${new Date(c.checked_at).toLocaleString('pt-BR')}${c.response_time_ms !== undefined ? ` — ${c.response_time_ms}ms` : ''}`}
            className={`w-2.5 h-2.5 rounded-sm ${color} opacity-80 hover:opacity-100 transition-opacity cursor-default`}
          />
        )
      })}
      {recent.length === 0 && (
        <span className="text-xs text-slate-400">nenhum check ainda</span>
      )}
    </div>
  )
}

// ── Monitor Form Modal ────────────────────────────────────────────────────────

interface MonitorFormData {
  name: string
  url: string
  interval_seconds: number
  timeout_seconds: number
}

const DEFAULT_FORM: MonitorFormData = {
  name: '',
  url: '',
  interval_seconds: 60,
  timeout_seconds: 10,
}

interface MonitorModalProps {
  editing: Monitor | null
  onClose: () => void
  onSaved: () => void
}

function MonitorModal({ editing, onClose, onSaved }: MonitorModalProps) {
  const [form, setForm] = useState<MonitorFormData>(() =>
    editing
      ? {
          name: editing.name,
          url: editing.url,
          interval_seconds: editing.interval_seconds,
          timeout_seconds: editing.timeout_seconds,
        }
      : DEFAULT_FORM,
  )
  const [saving, setSaving] = useState(false)
  const firstInputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    firstInputRef.current?.focus()
  }, [])

  // Close on Escape
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose()
    }
    window.addEventListener('keydown', handler)
    return () => window.removeEventListener('keydown', handler)
  }, [onClose])

  function handleChange(field: keyof MonitorFormData, value: string | number) {
    setForm((prev) => ({ ...prev, [field]: value }))
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()

    if (!form.name.trim()) {
      toast.error('Nome é obrigatório')
      return
    }
    if (!form.url.trim()) {
      toast.error('URL é obrigatória')
      return
    }
    if (form.interval_seconds < 30) {
      toast.error('Intervalo mínimo: 30s')
      return
    }
    if (form.timeout_seconds < 5) {
      toast.error('Timeout mínimo: 5s')
      return
    }

    setSaving(true)
    try {
      if (editing) {
        const payload: UpdateMonitorPayload = {
          name: form.name.trim(),
          url: form.url.trim(),
          interval_seconds: form.interval_seconds,
          timeout_seconds: form.timeout_seconds,
        }
        await monitorsApi.update(editing.id, payload)
        toast.success('Monitor atualizado')
      } else {
        const payload: CreateMonitorPayload = {
          name: form.name.trim(),
          url: form.url.trim(),
          interval_seconds: form.interval_seconds,
          timeout_seconds: form.timeout_seconds,
        }
        await monitorsApi.create(payload)
        toast.success('Monitor criado')
      }
      onSaved()
      onClose()
    } catch (err) {
      console.error(err)
      toast.error(editing ? 'Falha ao atualizar monitor' : 'Falha ao criar monitor')
    } finally {
      setSaving(false)
    }
  }

  return (
    /* Backdrop */
    <div
      className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm"
      role="dialog"
      aria-modal="true"
      aria-label={editing ? 'Editar monitor' : 'Novo monitor'}
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose()
      }}
    >
      {/* Panel */}
      <div className="w-full max-w-md rounded-2xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 shadow-2xl">
        {/* Modal header */}
        <div className="flex items-center justify-between px-5 py-4 border-b border-slate-100 dark:border-slate-800">
          <h2 className="text-base font-semibold text-slate-900 dark:text-slate-100">
            {editing ? 'Editar Monitor' : 'Novo Monitor'}
          </h2>
          <button
            onClick={onClose}
            className="p-1.5 rounded-lg text-slate-400 hover:text-slate-600 hover:bg-slate-100 dark:hover:text-slate-300 dark:hover:bg-slate-800 transition-colors"
            aria-label="Fechar modal"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit} noValidate>
          <div className="px-5 py-4 space-y-4">
            {/* Name */}
            <div>
              <label
                htmlFor="monitor-name"
                className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1.5"
              >
                Nome <span className="text-red-500" aria-hidden="true">*</span>
              </label>
              <input
                ref={firstInputRef}
                id="monitor-name"
                type="text"
                required
                value={form.name}
                onChange={(e) => handleChange('name', e.target.value)}
                placeholder="API de produção"
                className="w-full rounded-lg border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 px-3 py-2 text-sm text-slate-900 dark:text-slate-100 placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-emerald-500 dark:focus:ring-emerald-400 transition"
              />
            </div>

            {/* URL */}
            <div>
              <label
                htmlFor="monitor-url"
                className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1.5"
              >
                URL <span className="text-red-500" aria-hidden="true">*</span>
              </label>
              <input
                id="monitor-url"
                type="url"
                required
                value={form.url}
                onChange={(e) => handleChange('url', e.target.value)}
                placeholder="https://exemplo.com/health"
                className="w-full rounded-lg border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 px-3 py-2 text-sm text-slate-900 dark:text-slate-100 placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-emerald-500 dark:focus:ring-emerald-400 transition"
              />
            </div>

            {/* Interval + Timeout */}
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label
                  htmlFor="monitor-interval"
                  className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1.5"
                >
                  Intervalo (seg)
                </label>
                <input
                  id="monitor-interval"
                  type="number"
                  min={30}
                  value={form.interval_seconds}
                  onChange={(e) =>
                    handleChange('interval_seconds', Number(e.target.value))
                  }
                  className="w-full rounded-lg border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 px-3 py-2 text-sm text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-emerald-500 dark:focus:ring-emerald-400 transition"
                />
                <p className="text-xs text-slate-400 mt-1">mín. 30s</p>
              </div>

              <div>
                <label
                  htmlFor="monitor-timeout"
                  className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1.5"
                >
                  Timeout (seg)
                </label>
                <input
                  id="monitor-timeout"
                  type="number"
                  min={5}
                  value={form.timeout_seconds}
                  onChange={(e) =>
                    handleChange('timeout_seconds', Number(e.target.value))
                  }
                  className="w-full rounded-lg border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 px-3 py-2 text-sm text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-emerald-500 dark:focus:ring-emerald-400 transition"
                />
                <p className="text-xs text-slate-400 mt-1">mín. 5s</p>
              </div>
            </div>
          </div>

          {/* Footer */}
          <div className="flex items-center justify-end gap-2 px-5 py-4 border-t border-slate-100 dark:border-slate-800">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 rounded-lg text-sm font-medium text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
            >
              Cancelar
            </button>
            <button
              type="submit"
              disabled={saving}
              className="px-4 py-2 rounded-lg text-sm font-medium bg-emerald-600 hover:bg-emerald-500 disabled:opacity-50 disabled:cursor-not-allowed text-white transition-colors"
            >
              {saving ? 'Salvando…' : editing ? 'Salvar alterações' : 'Criar monitor'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

// ── Main page ─────────────────────────────────────────────────────────────────

export default function Uptime() {
  const [statuses, setStatuses] = useState<MonitorStatus[]>([])
  const [loading, setLoading] = useState(true)
  const [expandedId, setExpandedId] = useState<string | null>(null)
  const [historyMap, setHistoryMap] = useState<Record<string, MonitorCheck[]>>({})

  // Modal state
  const [modalOpen, setModalOpen] = useState(false)
  const [editingMonitor, setEditingMonitor] = useState<Monitor | null>(null)

  const fetchStatuses = useCallback(async () => {
    setLoading(true)
    try {
      const { data } = await monitorsApi.list()
      setStatuses(data)
    } catch (err) {
      toast.error('Falha ao carregar monitors')
      console.error(err)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void fetchStatuses()
    const interval = setInterval(() => void fetchStatuses(), 30_000)
    return () => clearInterval(interval)
  }, [fetchStatuses])

  const handleExpand = async (id: string) => {
    if (expandedId === id) {
      setExpandedId(null)
      return
    }
    setExpandedId(id)
    if (!historyMap[id]) {
      try {
        const { data } = await monitorsApi.history(id)
        setHistoryMap((prev) => ({ ...prev, [id]: data }))
      } catch {
        toast.error('Falha ao carregar histórico')
      }
    }
  }

  function openCreate() {
    setEditingMonitor(null)
    setModalOpen(true)
  }

  function openEdit(monitor: Monitor, e: React.MouseEvent) {
    e.stopPropagation()
    setEditingMonitor(monitor)
    setModalOpen(true)
  }

  function closeModal() {
    setModalOpen(false)
    setEditingMonitor(null)
  }

  async function handleDelete(monitor: Monitor, e: React.MouseEvent) {
    e.stopPropagation()
    const confirmed = window.confirm(
      `Remover o monitor "${monitor.name}"?\n\nEssa ação não pode ser desfeita.`,
    )
    if (!confirmed) return

    try {
      await monitorsApi.remove(monitor.id)
      toast.success(`Monitor "${monitor.name}" removido`)
      // Remove from local state immediately for snappy UX, then refresh
      setStatuses((prev) => prev.filter((s) => s.monitor.id !== monitor.id))
      if (expandedId === monitor.id) setExpandedId(null)
      void fetchStatuses()
    } catch (err) {
      console.error(err)
      toast.error('Falha ao remover monitor')
    }
  }

  const upCount = statuses.filter((s) => s.last_check?.status === 'up').length
  const downCount = statuses.filter(
    (s) => s.last_check?.status === 'down' || s.last_check?.status === 'timeout',
  ).length
  const noDataCount = statuses.filter((s) => !s.last_check).length

  return (
    <>
      <div className="space-y-5">
        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div>
            <h1 className="text-xl font-bold text-slate-900 dark:text-slate-100">
              Uptime Monitoring
            </h1>
            <p className="text-sm text-slate-500 dark:text-slate-400 mt-0.5">
              {statuses.length} serviços monitorados · verifica a cada 60s
            </p>
          </div>

          <div className="flex items-center gap-2 self-start sm:self-auto">
            {/* New monitor button */}
            <button
              onClick={openCreate}
              className="inline-flex items-center gap-1.5 px-3 py-2 rounded-lg text-sm font-medium bg-emerald-600 hover:bg-emerald-500 text-white transition-colors"
              aria-label="Criar novo monitor"
            >
              <Plus className="w-4 h-4" aria-hidden="true" />
              Novo Monitor
            </button>

            {/* Refresh */}
            <button
              onClick={fetchStatuses}
              disabled={loading}
              className="p-2 rounded-lg text-slate-500 hover:bg-slate-100 dark:text-slate-400 dark:hover:bg-slate-800 transition-colors disabled:opacity-50"
              aria-label="Recarregar"
            >
              <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
            </button>
          </div>
        </div>

        {/* Stats strip */}
        <div className="grid grid-cols-3 gap-3">
          <div className="rounded-xl border border-emerald-200 dark:border-emerald-800 bg-emerald-50 dark:bg-emerald-950/30 p-3">
            <p className="text-xs font-medium text-emerald-600 dark:text-emerald-400">Online</p>
            <p className="text-2xl font-bold text-emerald-700 dark:text-emerald-300">
              {loading ? '—' : upCount}
            </p>
          </div>
          <div className="rounded-xl border border-red-200 dark:border-red-800 bg-red-50 dark:bg-red-950/30 p-3">
            <p className="text-xs font-medium text-red-600 dark:text-red-400">Offline/Timeout</p>
            <p className="text-2xl font-bold text-red-700 dark:text-red-300">
              {loading ? '—' : downCount}
            </p>
          </div>
          <div className="rounded-xl border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800/30 p-3">
            <p className="text-xs font-medium text-slate-500 dark:text-slate-400">Sem dados</p>
            <p className="text-2xl font-bold text-slate-600 dark:text-slate-300">
              {loading ? '—' : noDataCount}
            </p>
          </div>
        </div>

        {/* Empty state */}
        {!loading && statuses.length === 0 && (
          <div className="text-center py-12 text-slate-400 dark:text-slate-500">
            Nenhum monitor configurado.{' '}
            <button
              onClick={openCreate}
              className="text-emerald-600 dark:text-emerald-400 hover:underline font-medium"
            >
              Criar o primeiro
            </button>
          </div>
        )}

        {/* Monitor list */}
        <div className="space-y-2">
          {statuses.map((ms) => {
            const lastStatus = ms.last_check?.status
            const isExpanded = expandedId === ms.monitor.id
            const history = historyMap[ms.monitor.id] ?? []

            return (
              <div
                key={ms.monitor.id}
                className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 overflow-hidden"
              >
                {/* Row */}
                <div className="px-4 py-3 flex items-center gap-3 hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-colors">
                  {/* Expand button (takes up most of the row) */}
                  <button
                    onClick={() => handleExpand(ms.monitor.id)}
                    className="flex-1 min-w-0 flex items-center gap-3 text-left"
                    aria-expanded={isExpanded}
                    aria-label={`${isExpanded ? 'Recolher' : 'Expandir'} ${ms.monitor.name}`}
                  >
                    {/* Status dot */}
                    <div
                      className={`w-2.5 h-2.5 rounded-full flex-shrink-0 ${
                        lastStatus === 'up'
                          ? 'bg-emerald-500'
                          : lastStatus === 'timeout'
                          ? 'bg-amber-400'
                          : lastStatus === 'down'
                          ? 'bg-red-500'
                          : 'bg-slate-300 dark:bg-slate-600'
                      }`}
                    />

                    {/* Name + URL */}
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium text-slate-800 dark:text-slate-200 truncate">
                        {ms.monitor.name}
                      </p>
                      <p className="text-xs text-slate-400 dark:text-slate-500 truncate">
                        {ms.monitor.url}
                      </p>
                    </div>

                    {/* Status badge */}
                    <div className="hidden sm:block flex-shrink-0">
                      <StatusBadge status={lastStatus} />
                    </div>

                    {/* Response time */}
                    <div className="hidden md:flex flex-col items-end w-20 flex-shrink-0">
                      <p className="text-xs text-slate-500 dark:text-slate-400 font-mono">
                        {formatMs(ms.last_check?.response_time_ms)}
                      </p>
                      <p className="text-xs text-slate-400 dark:text-slate-500">resp. time</p>
                    </div>

                    {/* Uptime bar */}
                    <div className="hidden lg:block w-36 flex-shrink-0">
                      <UptimeBar uptime={ms.uptime_24h} />
                    </div>

                    {/* Chevron */}
                    <svg
                      className={`w-4 h-4 text-slate-400 transition-transform flex-shrink-0 ${isExpanded ? 'rotate-180' : ''}`}
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                      aria-hidden="true"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M19 9l-7 7-7-7"
                      />
                    </svg>
                  </button>

                  {/* Action buttons */}
                  <div className="flex items-center gap-1 flex-shrink-0 ml-1">
                    <button
                      onClick={(e) => openEdit(ms.monitor, e)}
                      className="p-1.5 rounded-lg text-slate-400 hover:text-slate-600 hover:bg-slate-100 dark:hover:text-slate-300 dark:hover:bg-slate-800 transition-colors"
                      aria-label={`Editar ${ms.monitor.name}`}
                    >
                      <Pencil className="w-3.5 h-3.5" aria-hidden="true" />
                    </button>
                    <button
                      onClick={(e) => handleDelete(ms.monitor, e)}
                      className="p-1.5 rounded-lg text-slate-400 hover:text-red-600 hover:bg-red-50 dark:hover:text-red-400 dark:hover:bg-red-950/30 transition-colors"
                      aria-label={`Remover ${ms.monitor.name}`}
                    >
                      <Trash2 className="w-3.5 h-3.5" aria-hidden="true" />
                    </button>
                  </div>
                </div>

                {/* Expanded detail */}
                {isExpanded && (
                  <div className="border-t border-slate-100 dark:border-slate-800 px-4 py-3 space-y-3 bg-slate-50 dark:bg-slate-800/30">
                    {/* Stats row */}
                    <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
                      <div>
                        <p className="text-xs text-slate-500 dark:text-slate-400">Uptime 24h</p>
                        <p className="text-lg font-bold text-slate-800 dark:text-slate-200">
                          {ms.uptime_24h.toFixed(1)}%
                        </p>
                      </div>
                      <div>
                        <p className="text-xs text-slate-500 dark:text-slate-400">Resp. média 24h</p>
                        <p className="text-lg font-bold text-slate-800 dark:text-slate-200">
                          {formatMs(ms.avg_response_ms)}
                        </p>
                      </div>
                      <div>
                        <p className="text-xs text-slate-500 dark:text-slate-400">Último check</p>
                        <p className="text-sm font-medium text-slate-700 dark:text-slate-300">
                          {ms.last_check
                            ? new Date(ms.last_check.checked_at).toLocaleTimeString('pt-BR')
                            : '—'}
                        </p>
                      </div>
                      <div>
                        <p className="text-xs text-slate-500 dark:text-slate-400">Intervalo</p>
                        <p className="text-sm font-medium text-slate-700 dark:text-slate-300">
                          {ms.monitor.interval_seconds}s
                        </p>
                      </div>
                    </div>

                    {/* Error message if down */}
                    {ms.last_check?.error_message && (
                      <div className="rounded-lg bg-red-50 dark:bg-red-950/30 border border-red-200 dark:border-red-800 px-3 py-2">
                        <p className="text-xs text-red-600 dark:text-red-400 font-mono break-all">
                          {ms.last_check.error_message}
                        </p>
                      </div>
                    )}

                    {/* History dots */}
                    <div>
                      <p className="text-xs text-slate-500 dark:text-slate-400 mb-1.5">
                        Últimos 30 checks (mais recente à direita)
                      </p>
                      <HistoryDots checks={history} />
                    </div>
                  </div>
                )}
              </div>
            )
          })}
        </div>
      </div>

      {/* Modal */}
      {modalOpen && (
        <MonitorModal
          editing={editingMonitor}
          onClose={closeModal}
          onSaved={fetchStatuses}
        />
      )}
    </>
  )
}
