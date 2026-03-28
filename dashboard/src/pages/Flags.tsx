import { useCallback, useEffect, useState } from 'react'
import { Plus, RefreshCw, Search, Pencil, Trash2, X, AlertTriangle } from 'lucide-react'
import toast from 'react-hot-toast'
import { flagsApi } from '@/lib/api'
import type { FeatureFlag, CreateFlagPayload, UpdateFlagPayload } from '@/lib/api'

// ── Validation ──────────────────────────────────────────────────────

const KEY_REGEX = /^[a-zA-Z0-9_\-.]{2,100}$/

function sanitize(value: string): string {
  return value.replace(/[<>"'&]/g, '').trim()
}

// ── Toggle Switch ───────────────────────────────────────────────────

function ToggleSwitch({
  flagKey,
  enabled,
  onToggle,
}: {
  flagKey: string
  enabled: boolean
  onToggle: (key: string, enabled: boolean) => Promise<void>
}) {
  const [busy, setBusy] = useState(false)

  const handleClick = async () => {
    setBusy(true)
    try {
      await onToggle(flagKey, !enabled)
    } finally {
      setBusy(false)
    }
  }

  return (
    <button
      onClick={handleClick}
      disabled={busy}
      className={[
        'relative inline-flex w-10 h-6 rounded-full transition-colors focus:outline-none focus:ring-2 focus:ring-violet-500 focus:ring-offset-2 dark:focus:ring-offset-slate-900',
        enabled ? 'bg-violet-600' : 'bg-slate-300 dark:bg-slate-600',
        busy ? 'opacity-60 cursor-wait' : 'cursor-pointer',
      ].join(' ')}
      aria-label={enabled ? 'Desativar flag' : 'Ativar flag'}
    >
      <span
        className={[
          'absolute top-0.5 left-0.5 w-5 h-5 bg-white rounded-full shadow transition-transform',
          enabled ? 'translate-x-4' : 'translate-x-0',
        ].join(' ')}
      />
    </button>
  )
}

// ── Delete Confirmation ─────────────────────────────────────────────

function DeleteConfirm({
  flag,
  onConfirm,
  onCancel,
  loading,
}: {
  flag: FeatureFlag
  onConfirm: () => void
  onCancel: () => void
  loading: boolean
}) {
  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-fade-in"
      onClick={(e) => e.target === e.currentTarget && onCancel()}
    >
      <div className="bg-white dark:bg-slate-900 rounded-xl shadow-soft-lg w-full max-w-sm border border-slate-200 dark:border-slate-700 animate-scale-in p-5 space-y-4">
        <div className="flex items-start gap-3">
          <div className="p-2 bg-red-100 dark:bg-red-900/40 rounded-lg shrink-0">
            <AlertTriangle className="w-5 h-5 text-red-600 dark:text-red-400" />
          </div>
          <div>
            <h3 className="font-semibold text-slate-900 dark:text-slate-100 text-sm">
              Deletar flag?
            </h3>
            <p className="text-sm text-slate-500 dark:text-slate-400 mt-1">
              A flag <strong className="text-slate-700 dark:text-slate-300">{flag.key}</strong> sera
              removida permanentemente.
            </p>
          </div>
        </div>
        <div className="flex justify-end gap-2">
          <button
            onClick={onCancel}
            className="px-4 py-2 rounded-lg text-sm font-medium text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
          >
            Cancelar
          </button>
          <button
            onClick={onConfirm}
            disabled={loading}
            className="px-4 py-2 rounded-lg text-sm font-medium bg-red-600 hover:bg-red-700 disabled:opacity-60 text-white transition-colors"
          >
            Deletar
          </button>
        </div>
      </div>
    </div>
  )
}

// ── Flag Modal ──────────────────────────────────────────────────────

function FlagModal({
  flag,
  onClose,
  onSubmit,
  loading,
}: {
  flag: FeatureFlag | null
  onClose: () => void
  onSubmit: (payload: CreateFlagPayload | UpdateFlagPayload) => Promise<void>
  loading: boolean
}) {
  const isEdit = !!flag
  const [key, setKey] = useState(flag?.key ?? '')
  const [description, setDescription] = useState(flag?.description ?? '')
  const [enabled, setEnabled] = useState(flag?.enabled ?? false)
  const [keyError, setKeyError] = useState('')

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    const sanitizedDesc = sanitize(description)

    if (!isEdit) {
      const sanitizedKey = sanitize(key)
      if (!KEY_REGEX.test(sanitizedKey)) {
        setKeyError('Key: 2-100 caracteres, apenas letras, numeros, _, - e .')
        return
      }
      await onSubmit({ key: sanitizedKey, description: sanitizedDesc, enabled })
    } else {
      await onSubmit({ description: sanitizedDesc, enabled })
    }
  }

  const inputClass =
    'w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 text-sm focus:outline-none focus:ring-2 focus:ring-violet-500 focus:border-transparent'

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-fade-in"
      onClick={(e) => e.target === e.currentTarget && onClose()}
    >
      <div className="bg-white dark:bg-slate-900 rounded-xl shadow-soft-lg w-full max-w-md border border-slate-200 dark:border-slate-700 animate-scale-in">
        <div className="flex items-center justify-between px-5 py-4 border-b border-slate-200 dark:border-slate-700">
          <h2 className="text-base font-semibold text-slate-900 dark:text-slate-100">
            {isEdit ? `Editar: ${flag.key}` : 'Nova feature flag'}
          </h2>
          <button
            onClick={onClose}
            className="p-1.5 rounded-lg text-slate-400 hover:text-slate-600 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="px-5 py-4 space-y-4">
          {!isEdit && (
            <div>
              <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">
                Key *
              </label>
              <input
                type="text"
                value={key}
                onChange={(e) => { setKey(e.target.value); setKeyError('') }}
                placeholder="ex: new-checkout-flow"
                className={inputClass}
                maxLength={100}
              />
              {keyError && <p className="mt-1 text-xs text-red-500">{keyError}</p>}
            </div>
          )}

          <div>
            <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">
              Descricao
            </label>
            <textarea
              rows={2}
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Descreva o proposito desta flag..."
              className={`${inputClass} resize-none`}
              maxLength={500}
            />
          </div>

          <div className="flex items-center gap-3">
            <button
              type="button"
              onClick={() => setEnabled(!enabled)}
              className={[
                'relative inline-flex w-10 h-6 rounded-full transition-colors',
                enabled ? 'bg-violet-600' : 'bg-slate-300 dark:bg-slate-600',
              ].join(' ')}
            >
              <span
                className={[
                  'absolute top-0.5 left-0.5 w-5 h-5 bg-white rounded-full shadow transition-transform',
                  enabled ? 'translate-x-4' : 'translate-x-0',
                ].join(' ')}
              />
            </button>
            <span className="text-sm text-slate-700 dark:text-slate-300">
              {enabled ? 'Ativa' : 'Inativa'}
            </span>
          </div>

          <div className="flex justify-end gap-2 pt-2 border-t border-slate-100 dark:border-slate-800">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 rounded-lg text-sm font-medium text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
            >
              Cancelar
            </button>
            <button
              type="submit"
              disabled={loading}
              className="px-4 py-2 rounded-lg text-sm font-medium bg-violet-600 hover:bg-violet-700 disabled:opacity-60 text-white transition-colors"
            >
              {loading ? 'Salvando...' : isEdit ? 'Salvar' : 'Criar flag'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

// ── Main Flags Page ─────────────────────────────────────────────────

export default function Flags() {
  const [flags, setFlags] = useState<FeatureFlag[]>([])
  const [loading, setLoading] = useState(true)
  const [submitting, setSubmitting] = useState(false)
  const [search, setSearch] = useState('')
  const [modalOpen, setModalOpen] = useState(false)
  const [editTarget, setEditTarget] = useState<FeatureFlag | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<FeatureFlag | null>(null)
  const [deleteLoading, setDeleteLoading] = useState(false)

  const fetchFlags = useCallback(async () => {
    setLoading(true)
    try {
      const data = await flagsApi.list()
      setFlags(data)
    } catch (err) {
      toast.error('Falha ao carregar feature flags')
      console.error(err)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void fetchFlags()
  }, [fetchFlags])

  const handleCreate = () => {
    setEditTarget(null)
    setModalOpen(true)
  }

  const handleEdit = (flag: FeatureFlag) => {
    setEditTarget(flag)
    setModalOpen(true)
  }

  const handleModalClose = () => {
    setModalOpen(false)
    setEditTarget(null)
  }

  const handleSubmit = async (payload: CreateFlagPayload | UpdateFlagPayload) => {
    setSubmitting(true)
    try {
      if (editTarget) {
        const data = await flagsApi.update(editTarget.key, payload as UpdateFlagPayload)
        setFlags((prev) => prev.map((f) => (f.id === data.id ? data : f)))
        toast.success('Flag atualizada')
      } else {
        const data = await flagsApi.create(payload as CreateFlagPayload)
        setFlags((prev) => [data, ...prev])
        toast.success('Flag criada')
      }
      handleModalClose()
    } catch (err) {
      toast.error(editTarget ? 'Falha ao atualizar flag' : 'Falha ao criar flag')
      console.error(err)
    } finally {
      setSubmitting(false)
    }
  }

  const handleDelete = async () => {
    if (!deleteTarget) return
    setDeleteLoading(true)
    try {
      await flagsApi.remove(deleteTarget.key)
      setFlags((prev) => prev.filter((f) => f.key !== deleteTarget.key))
      toast.success('Flag deletada')
      setDeleteTarget(null)
    } catch (err) {
      toast.error('Falha ao deletar flag')
      console.error(err)
    } finally {
      setDeleteLoading(false)
    }
  }

  const handleToggle = async (key: string, enabled: boolean) => {
    try {
      await flagsApi.update(key, { enabled })
      setFlags((prev) => prev.map((f) => (f.key === key ? { ...f, enabled } : f)))
      toast.success(enabled ? 'Flag ativada' : 'Flag desativada')
    } catch (err) {
      toast.error('Falha ao alternar flag')
      console.error(err)
    }
  }

  const filtered = flags.filter((f) => {
    const q = search.trim().toLowerCase()
    if (!q) return true
    return f.key.toLowerCase().includes(q) || (f.description ?? '').toLowerCase().includes(q)
  })

  const enabledCount = flags.filter((f) => f.enabled).length

  const formatDate = (iso: string) => {
    try {
      return new Intl.DateTimeFormat('pt-BR', {
        day: '2-digit',
        month: '2-digit',
        year: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
      }).format(new Date(iso))
    } catch {
      return iso
    }
  }

  return (
    <div className="space-y-5">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-xl font-bold text-slate-900 dark:text-slate-100">Feature Flags</h1>
          <p className="text-sm text-slate-500 dark:text-slate-400 mt-0.5">
            {flags.length} flags cadastradas -- {enabledCount} ativas
          </p>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={fetchFlags}
            disabled={loading}
            className="p-2 rounded-lg text-slate-500 hover:bg-slate-100 dark:text-slate-400 dark:hover:bg-slate-800 transition-colors disabled:opacity-50"
            aria-label="Recarregar"
          >
            <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
          </button>
          <button
            onClick={handleCreate}
            className="flex items-center gap-2 px-4 py-2 rounded-lg bg-violet-600 hover:bg-violet-700 text-white text-sm font-medium transition-colors"
          >
            <Plus className="w-4 h-4" />
            Nova flag
          </button>
        </div>
      </div>

      <div className="relative">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
        <input
          type="text"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Buscar por key ou descricao..."
          className="w-full pl-9 pr-3 py-2 rounded-lg border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 text-sm focus:outline-none focus:ring-2 focus:ring-violet-500 focus:border-transparent placeholder:text-slate-400"
        />
      </div>

      {/* Stats */}
      <div className="grid grid-cols-2 gap-3">
        <div className="rounded-xl border bg-gradient-to-br from-violet-50 to-violet-100 dark:from-violet-950/40 dark:to-violet-900/20 border-violet-200 dark:border-violet-800 p-3">
          <p className="text-xs font-medium text-slate-500 dark:text-slate-400">Total</p>
          <p className="text-2xl font-bold text-slate-800 dark:text-slate-200 mt-0.5">{flags.length}</p>
        </div>
        <div className="rounded-xl border bg-gradient-to-br from-emerald-50 to-emerald-100 dark:from-emerald-950/40 dark:to-emerald-900/20 border-emerald-200 dark:border-emerald-800 p-3">
          <p className="text-xs font-medium text-slate-500 dark:text-slate-400">Ativas</p>
          <p className="text-2xl font-bold text-slate-800 dark:text-slate-200 mt-0.5">{enabledCount}</p>
        </div>
      </div>

      {/* Table */}
      {loading ? (
        <div className="flex items-center justify-center py-16 text-slate-400">
          <div className="w-6 h-6 border-2 border-violet-500 border-t-transparent rounded-full animate-spin mr-3" />
          <span className="text-sm">Carregando flags...</span>
        </div>
      ) : filtered.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-16 text-slate-400 gap-2">
          <p className="text-sm">Nenhuma feature flag encontrada</p>
        </div>
      ) : (
        <>
          {/* Desktop */}
          <div className="hidden md:block overflow-x-auto rounded-xl border border-slate-200 dark:border-slate-800">
            <table className="w-full text-sm">
              <thead>
                <tr className="bg-slate-50 dark:bg-slate-800/50 text-left text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide">
                  <th className="px-4 py-3">Key</th>
                  <th className="px-4 py-3">Descricao</th>
                  <th className="px-4 py-3 text-center">Ativo</th>
                  <th className="px-4 py-3">Criado em</th>
                  <th className="px-4 py-3 text-right">Acoes</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100 dark:divide-slate-800">
                {filtered.map((flag) => (
                  <tr
                    key={flag.id}
                    className="bg-white dark:bg-slate-900 hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-colors"
                  >
                    <td className="px-4 py-3 font-mono text-xs font-medium text-slate-800 dark:text-slate-200">
                      {flag.key}
                    </td>
                    <td className="px-4 py-3 text-slate-500 dark:text-slate-400 max-w-xs truncate">
                      {flag.description || <span className="text-slate-300 dark:text-slate-600">--</span>}
                    </td>
                    <td className="px-4 py-3 text-center">
                      <ToggleSwitch flagKey={flag.key} enabled={flag.enabled} onToggle={handleToggle} />
                    </td>
                    <td className="px-4 py-3 text-xs text-slate-400 font-mono whitespace-nowrap">
                      {formatDate(flag.created_at)}
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex items-center justify-end gap-1">
                        <button
                          onClick={() => handleEdit(flag)}
                          className="p-1.5 rounded-lg text-slate-400 hover:text-violet-600 hover:bg-violet-50 dark:hover:bg-violet-900/30 transition-colors"
                          aria-label="Editar"
                        >
                          <Pencil className="w-3.5 h-3.5" />
                        </button>
                        <button
                          onClick={() => setDeleteTarget(flag)}
                          className="p-1.5 rounded-lg text-slate-400 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-900/30 transition-colors"
                          aria-label="Deletar"
                        >
                          <Trash2 className="w-3.5 h-3.5" />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {/* Mobile cards */}
          <div className="md:hidden space-y-3">
            {filtered.map((flag) => (
              <div
                key={flag.id}
                className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-4 space-y-3"
              >
                <div className="flex items-start justify-between gap-3">
                  <div className="min-w-0">
                    <p className="font-mono text-sm font-medium text-slate-800 dark:text-slate-200 truncate">
                      {flag.key}
                    </p>
                    {flag.description && (
                      <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5 line-clamp-2">
                        {flag.description}
                      </p>
                    )}
                  </div>
                  <ToggleSwitch flagKey={flag.key} enabled={flag.enabled} onToggle={handleToggle} />
                </div>
                <div className="flex items-center justify-between pt-2 border-t border-slate-100 dark:border-slate-800">
                  <span className="text-xs text-slate-400 font-mono">{formatDate(flag.created_at)}</span>
                  <div className="flex items-center gap-1">
                    <button
                      onClick={() => handleEdit(flag)}
                      className="p-1.5 rounded-lg text-slate-400 hover:text-violet-600 hover:bg-violet-50 dark:hover:bg-violet-900/30 transition-colors"
                    >
                      <Pencil className="w-3.5 h-3.5" />
                    </button>
                    <button
                      onClick={() => setDeleteTarget(flag)}
                      className="p-1.5 rounded-lg text-slate-400 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-900/30 transition-colors"
                    >
                      <Trash2 className="w-3.5 h-3.5" />
                    </button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </>
      )}

      {/* Modals */}
      {modalOpen && (
        <FlagModal
          flag={editTarget}
          onClose={handleModalClose}
          onSubmit={handleSubmit}
          loading={submitting}
        />
      )}

      {deleteTarget && (
        <DeleteConfirm
          flag={deleteTarget}
          onConfirm={handleDelete}
          onCancel={() => setDeleteTarget(null)}
          loading={deleteLoading}
        />
      )}
    </div>
  )
}
