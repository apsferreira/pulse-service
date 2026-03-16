import { useCallback, useEffect, useState } from 'react'
import { Plus, RefreshCw, Search } from 'lucide-react'
import toast from 'react-hot-toast'
import FlagTable from '../components/FlagTable'
import FlagModal from '../components/FlagModal'
import { flagsApi } from '../services/api'
import type { FeatureFlag, CreateFlagPayload, UpdateFlagPayload, Product } from '../services/api'

type FilterProduct = Product | 'all'

const PRODUCT_OPTIONS: { value: FilterProduct; label: string }[] = [
  { value: 'all', label: 'Todos os produtos' },
  { value: 'libri', label: 'Libri' },
  { value: 'nitro', label: 'Nitro' },
  { value: 'brio', label: 'Brio' },
  { value: 'global', label: 'Global' },
]

export default function Dashboard() {
  const [flags, setFlags] = useState<FeatureFlag[]>([])
  const [loading, setLoading] = useState(true)
  const [submitting, setSubmitting] = useState(false)

  // Filters
  const [search, setSearch] = useState('')
  const [filterProduct, setFilterProduct] = useState<FilterProduct>('all')

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
        const { data } = await flagsApi.update(editTarget.id, payload as UpdateFlagPayload)
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

  const handleDelete = async (id: string) => {
    await flagsApi.remove(id)
    setFlags((prev) => prev.filter((f) => f.id !== id))
    toast.success('Flag deletada')
  }

  const handleToggle = async (id: string, enabled: boolean) => {
    await flagsApi.update(id, { is_enabled: enabled })
    setFlags((prev) => prev.map((f) => (f.id === id ? { ...f, is_enabled: enabled } : f)))
    toast.success(enabled ? 'Flag ativada' : 'Flag desativada')
  }

  const filtered = flags.filter((f) => {
    const matchSearch =
      search.trim() === '' ||
      f.name.toLowerCase().includes(search.toLowerCase()) ||
      (f.description ?? '').toLowerCase().includes(search.toLowerCase())

    const matchProduct =
      filterProduct === 'all' ||
      (filterProduct === 'global' ? !f.product || f.product === 'global' : f.product === filterProduct)

    return matchSearch && matchProduct
  })

  const enabledCount = flags.filter((f) => f.is_enabled).length

  return (
    <div className="space-y-5">
      {/* Page header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-xl font-bold text-slate-900 dark:text-slate-100">Feature Flags</h1>
          <p className="text-sm text-slate-500 dark:text-slate-400 mt-0.5">
            {flags.length} flags cadastradas · {enabledCount} ativas
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
            placeholder="Buscar por nome ou descrição..."
            className="w-full pl-9 pr-3 py-2 rounded-lg border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 text-sm focus:outline-none focus:ring-2 focus:ring-violet-500 focus:border-transparent placeholder:text-slate-400 dark:placeholder:text-slate-500"
          />
        </div>
        <select
          value={filterProduct}
          onChange={(e) => setFilterProduct(e.target.value as FilterProduct)}
          className="px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 text-sm focus:outline-none focus:ring-2 focus:ring-violet-500 focus:border-transparent"
        >
          {PRODUCT_OPTIONS.map(({ value, label }) => (
            <option key={value} value={value}>
              {label}
            </option>
          ))}
        </select>
      </div>

      {/* Stats strip */}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
        {(['libri', 'nitro', 'brio', 'global'] as const).map((product) => {
          const count = flags.filter(
            (f) => f.product === product || (!f.product && product === 'global')
          ).length
          const active = flags.filter(
            (f) =>
              f.is_enabled &&
              (f.product === product || (!f.product && product === 'global'))
          ).length
          const COLORS: Record<string, string> = {
            libri: 'from-blue-50 to-blue-100 dark:from-blue-950/40 dark:to-blue-900/20 border-blue-200 dark:border-blue-800',
            nitro: 'from-orange-50 to-orange-100 dark:from-orange-950/40 dark:to-orange-900/20 border-orange-200 dark:border-orange-800',
            brio: 'from-emerald-50 to-emerald-100 dark:from-emerald-950/40 dark:to-emerald-900/20 border-emerald-200 dark:border-emerald-800',
            global: 'from-slate-50 to-slate-100 dark:from-slate-800/40 dark:to-slate-700/20 border-slate-200 dark:border-slate-700',
          }
          return (
            <div
              key={product}
              className={`rounded-xl border bg-gradient-to-br ${COLORS[product]} p-3 cursor-pointer transition-all hover:shadow-soft`}
              onClick={() => setFilterProduct(filterProduct === product ? 'all' : product)}
            >
              <p className="text-xs font-medium text-slate-500 dark:text-slate-400 capitalize">{product}</p>
              <p className="text-2xl font-bold text-slate-800 dark:text-slate-200 mt-0.5">{count}</p>
              <p className="text-xs text-slate-400 mt-0.5">{active} ativas</p>
            </div>
          )
        })}
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
