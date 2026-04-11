import { useCallback, useEffect, useState } from 'react'
import { Plus, RefreshCw, Search, Lightbulb, AlertTriangle, CheckCircle, ArrowRight } from 'lucide-react'
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
  const disabledCount = flags.length - enabledCount
  const activePercent = Math.round((enabledCount / Math.max(flags.length, 1)) * 100)

  type RecommendedAction = {
    id: string
    icon: 'lightbulb' | 'alert' | 'check'
    title: string
    description: string
    href: string
    linkLabel: string
    priority: number
  }

  const recommendedActions: RecommendedAction[] = []

  if (flags.length === 0) {
    recommendedActions.push({
      id: 'no-flags',
      icon: 'lightbulb',
      title: 'Crie sua primeira feature flag',
      description: 'Comece adicionando uma flag para controlar funcionalidades do seu produto com segurança.',
      href: '#',
      linkLabel: 'Criar flag agora',
      priority: 1,
    })
  } else if (disabledCount > flags.length / 2) {
    recommendedActions.push({
      id: 'many-disabled',
      icon: 'alert',
      title: `${disabledCount} flags desabilitadas acumuladas`,
      description: 'Você tem mais flags desabilitadas do que ativas. Ative ou arquive as que não usa para manter o projeto limpo.',
      href: '#',
      linkLabel: 'Ver flags desabilitadas',
      priority: 1,
    })
  }

  const hasAnyRollout = flags.some((f) => f.rollout_percentage > 0)

  if (flags.length > 0 && !hasAnyRollout) {
    recommendedActions.push({
      id: 'no-rollout',
      icon: 'lightbulb',
      title: 'Configure rollout gradual',
      description: 'Nenhuma flag usa rollout percentual. Releases graduais reduzem risco e permitem rollback instantâneo.',
      href: '#',
      linkLabel: 'Editar uma flag',
      priority: 2,
    })
  }

  recommendedActions.push({
    id: 'health-scores',
    icon: 'check',
    title: 'Verifique o Health Score dos tenants',
    description: 'Acompanhe a saúde de cada tenant e identifique problemas de adoção antes que virem suporte.',
    href: '/health-scores',
    linkLabel: 'Ver Health Scores',
    priority: 3,
  })

  recommendedActions.push({
    id: 'ab-testing',
    icon: 'check',
    title: 'Consulte seus experimentos A/B ativos',
    description: 'Analise os resultados dos seus testes A/B em andamento e tome decisões baseadas em dados.',
    href: '/ab-testing',
    linkLabel: 'Ver experimentos',
    priority: 4,
  })

  const topActions = recommendedActions
    .sort((a, b) => a.priority - b.priority)
    .slice(0, 3)

  const ActionIcon = ({ type }: { type: RecommendedAction['icon'] }) => {
    if (type === 'alert') return <AlertTriangle className="w-4 h-4 text-amber-500" />
    if (type === 'check') return <CheckCircle className="w-4 h-4 text-emerald-500" />
    return <Lightbulb className="w-4 h-4 text-violet-500" />
  }

  return (
    <div className="space-y-5">
      {/* BKL-1066: KPI overview — 4 cards */}
      {!loading && (
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
          <div className="rounded-xl border bg-gradient-to-br from-violet-50 to-violet-100 dark:from-violet-950/40 dark:to-violet-900/20 border-violet-200 dark:border-violet-800 p-4">
            <p className="text-xs font-medium text-slate-500 dark:text-slate-400">Flags ativas</p>
            <p className="text-2xl font-bold text-violet-700 dark:text-violet-300 mt-0.5">{enabledCount}</p>
          </div>
          <div className="rounded-xl border bg-gradient-to-br from-emerald-50 to-emerald-100 dark:from-emerald-950/40 dark:to-emerald-900/20 border-emerald-200 dark:border-emerald-800 p-4">
            <p className="text-xs font-medium text-slate-500 dark:text-slate-400">Total de flags</p>
            <p className="text-2xl font-bold text-emerald-700 dark:text-emerald-300 mt-0.5">{flags.length}</p>
          </div>
          <div className="rounded-xl border bg-gradient-to-br from-amber-50 to-amber-100 dark:from-amber-950/40 dark:to-amber-900/20 border-amber-200 dark:border-amber-800 p-4">
            <p className="text-xs font-medium text-slate-500 dark:text-slate-400">Desabilitadas</p>
            <p className="text-2xl font-bold text-amber-700 dark:text-amber-300 mt-0.5">{disabledCount}</p>
          </div>
          <div className="rounded-xl border bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-800/60 dark:to-slate-800/30 border-slate-200 dark:border-slate-700 p-4">
            <p className="text-xs font-medium text-slate-500 dark:text-slate-400">% Ativas</p>
            <p className="text-2xl font-bold text-slate-700 dark:text-slate-200 mt-0.5">{activePercent}%</p>
          </div>
        </div>
      )}

      {/* BKL-1066: Ações recomendadas */}
      {!loading && topActions.length > 0 && (
        <div className="rounded-xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 p-4">
          <h2 className="text-sm font-semibold text-slate-700 dark:text-slate-200 mb-3">Ações recomendadas</h2>
          <ul className="space-y-2">
            {topActions.map((action) => (
              <li
                key={action.id}
                className="flex items-start gap-3 rounded-lg border border-slate-100 dark:border-slate-800 bg-slate-50 dark:bg-slate-800/50 p-3"
              >
                <span className="mt-0.5 shrink-0">
                  <ActionIcon type={action.icon} />
                </span>
                <div className="min-w-0 flex-1">
                  <p className="text-sm font-medium text-slate-800 dark:text-slate-100">{action.title}</p>
                  <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5">{action.description}</p>
                </div>
                <a
                  href={action.href}
                  className="shrink-0 flex items-center gap-1 text-xs font-medium text-violet-600 dark:text-violet-400 hover:text-violet-800 dark:hover:text-violet-200 transition-colors"
                  aria-label={action.linkLabel}
                >
                  {action.linkLabel}
                  <ArrowRight className="w-3 h-3" aria-hidden="true" />
                </a>
              </li>
            ))}
          </ul>
        </div>
      )}

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
