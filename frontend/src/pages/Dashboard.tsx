import { useCallback, useEffect, useState } from 'react'
import { Plus, RefreshCw, Search } from 'lucide-react'
import toast from 'react-hot-toast'
import FlagTable from '../components/FlagTable'
import FlagModal from '../components/FlagModal'
import OnboardingWizard from '../components/OnboardingWizard'
import { flagsApi } from '../services/api'
import type { FeatureFlag, CreateFlagPayload, UpdateFlagPayload } from '../services/api'

export default function Dashboard() {
  const [flags, setFlags] = useState<FeatureFlag[]>([])
  const [loading, setLoading] = useState(true)
  const [submitting, setSubmitting] = useState(false)

  // Filters
  const [search, setSearch] = useState('')

  // Modal state
  const [modalOpen, setModalOpen] = useState(false)
  const [editTarget, setEditTarget] = useState<FeatureFlag | null>(null)

  const fetchFlags = useCallback(async () => {
    setLoading(true)
    try {
      const { data } = await flagsApi.list()
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
        const { data } = await flagsApi.update(editTarget.key, payload as UpdateFlagPayload)
        setFlags((prev) => prev.map((f) => (f.id === data.id ? data : f)))
        toast.success('Flag atualizada com sucesso')
      } else {
        const { data } = await flagsApi.create(payload as CreateFlagPayload)
        setFlags((prev) => [data, ...prev])
        toast.success('Flag criada com sucesso')
      }
      handleModalClose()
    } catch (err) {
      toast.error(editTarget ? 'Falha ao atualizar flag' : 'Falha ao criar flag')
      console.error(err)
    } finally {
      setSubmitting(false)
    }
  }

  const handleDelete = async (key: string) => {
    await flagsApi.remove(key)
    setFlags((prev) => prev.filter((f) => f.key !== key))
    toast.success('Flag deletada')
  }

  const handleToggle = async (key: string, enabled: boolean) => {
    await flagsApi.update(key, { enabled })
    setFlags((prev) => prev.map((f) => (f.key === key ? { ...f, enabled } : f)))
    toast.success(enabled ? 'Flag ativada' : 'Flag desativada')
  }

  const filtered = flags.filter((f) => {
    return (
      search.trim() === '' ||
      f.key.toLowerCase().includes(search.toLowerCase()) ||
      (f.description ?? '').toLowerCase().includes(search.toLowerCase())
    )
  })

  const enabledCount = flags.filter((f) => f.enabled).length

  return (
    <div className="space-y-5">
      {/* BKL-1050: Onboarding wizard — visível apenas quando não há flags */}
      <OnboardingWizard flagsCount={flags.length} />

      {/* Page header */}
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
            className="flex items-center gap-2 px-4 py-2 rounded-lg bg-violet-600 hover:bg-violet-700 text-white text-sm font-medium transition-colors shadow-pulse"
          >
            <Plus className="w-4 h-4" />
            Nova flag
          </button>
        </div>
      </div>

      {/* Filters */}
      <div className="flex flex-col sm:flex-row gap-3">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Buscar por key ou descricao..."
            className="w-full pl-9 pr-3 py-2 rounded-lg border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 text-sm focus:outline-none focus:ring-2 focus:ring-violet-500 focus:border-transparent placeholder:text-slate-400 dark:placeholder:text-slate-500"
          />
        </div>
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
      <FlagTable
        flags={filtered}
        onEdit={handleEdit}
        onDelete={handleDelete}
        onToggle={handleToggle}
        loading={loading}
      />

      {/* Modal */}
      {modalOpen && (
        <FlagModal
          flag={editTarget}
          onClose={handleModalClose}
          onSubmit={handleSubmit}
          loading={submitting}
        />
      )}
    </div>
  )
}
