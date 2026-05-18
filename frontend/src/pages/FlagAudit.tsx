// BKL-1074: Audit log de avaliações de feature flags.
import { useState, useEffect, useCallback } from 'react'
import { History, ChevronLeft, ChevronRight, CheckCircle, XCircle } from 'lucide-react'
import { api } from '../services/api'
import { useProject } from '../components/Layout'
import UpgradeModal from '../components/UpgradeModal'
import { useUpgradeGate } from '../hooks/useUpgradeGate'
import { usePlanGate } from '../hooks/usePlanGate'

interface FlagEvaluation {
  id: string
  flag_key: string
  project_id: string | null
  user_id: string | null
  tenant_id: string | null
  result: boolean
  variant: string | null
  reason: string
  evaluated_at: string
}

interface AuditResponse {
  items: FlagEvaluation[]
  total: number
  limit: number
  offset: number
}

const REASON_LABELS: Record<string, string> = {
  kill_switch: 'Kill Switch',
  targeting_user: 'Targeting (usuário)',
  targeting_tenant: 'Targeting (tenant)',
  rollout: 'Rollout %',
  default: 'Padrão',
  disabled: 'Desabilitada',
}

const REASON_COLORS: Record<string, string> = {
  kill_switch: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400',
  targeting_user: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400',
  targeting_tenant: 'bg-indigo-100 text-indigo-700 dark:bg-indigo-900/30 dark:text-indigo-400',
  rollout: 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400',
  default: 'bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-400',
  disabled: 'bg-gray-100 text-gray-500 dark:bg-gray-800 dark:text-gray-400',
}

const PAGE_SIZE = 50

export default function FlagAudit() {
  const { projectId, projectPlan } = useProject()
  const { gateState, triggerGate, closeGate } = useUpgradeGate()
  const canUseFlagAudit = usePlanGate('starter')
  const [items, setItems] = useState<FlagEvaluation[]>([])
  const [total, setTotal] = useState(0)
  const [offset, setOffset] = useState(0)
  const [loading, setLoading] = useState(true)
  const [filterKey, setFilterKey] = useState('')

  const fetchAudit = useCallback(async () => {
    setLoading(true)
    try {
      const params: Record<string, string | number> = { limit: PAGE_SIZE, offset }
      if (filterKey) params.flag_key = filterKey
      if (projectId) params.project_id = projectId

      const { data } = await api.get<AuditResponse>('/flag-evaluations/audit', { params })
      setItems(data.items ?? [])
      setTotal(data.total ?? 0)
    } catch (err) {
      console.error('audit load error', err)
    } finally {
      setLoading(false)
    }
  }, [projectId, offset, filterKey])

  useEffect(() => {
    void fetchAudit()
  }, [fetchAudit])

  useEffect(() => {
    setOffset(0)
  }, [projectId, filterKey])

  const totalPages = Math.ceil(total / PAGE_SIZE)
  const currentPage = Math.floor(offset / PAGE_SIZE) + 1

  if (!canUseFlagAudit) {
    return (
      <div className="flex flex-col items-center justify-center py-24 gap-4 text-center">
        <History className="w-12 h-12 text-violet-300" />
        <h2 className="text-xl font-bold text-slate-900 dark:text-slate-100">Audit Log — Feature Flags</h2>
        <p className="text-sm text-slate-500 dark:text-slate-400 max-w-sm">
          Historico de avaliacoes de flags com motivo da decisao esta disponivel no plano Starter.
        </p>
        <button
          onClick={() => triggerGate({ plan: projectPlan, required_plan: 'starter', upgrade_url: 'https://checkout-api.institutoitinerante.com.br/checkout?plan=pulse-starter', message: 'Disponivel no plano Starter (R$399/mes). Comece com 14 dias gratis.', feature: 'Audit Log de Feature Flags' })}
          className="px-4 py-2 bg-violet-600 hover:bg-violet-700 text-white text-sm font-semibold rounded-lg transition-colors"
        >
          Ver planos
        </button>
        <UpgradeModal
          isOpen={gateState.isOpen}
          onClose={closeGate}
          currentPlan={gateState.currentPlan}
          requiredPlan={gateState.requiredPlan}
          feature={gateState.feature}
          upgradeUrl={gateState.upgradeUrl}
          message={gateState.message}
        />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight flex items-center gap-2">
            <History className="w-6 h-6 text-violet-500" />
            Audit Log — Feature Flags
          </h1>
          <p className="text-slate-500 dark:text-slate-400 text-sm mt-1">
            Histórico de avaliações de flags com motivo da decisão.
          </p>
        </div>
        <span className="text-sm text-slate-400 dark:text-slate-500">
          {total.toLocaleString('pt-BR')} avaliações
        </span>
      </div>

      {/* Filters */}
      <div className="flex items-center gap-3">
        <input
          type="text"
          placeholder="Filtrar por flag key..."
          value={filterKey}
          onChange={(e) => setFilterKey(e.target.value)}
          className="px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-sm w-64 focus:outline-none focus:ring-2 focus:ring-violet-500"
        />
      </div>

      {/* Table */}
      <div className="rounded-xl border border-slate-200 dark:border-slate-800 overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-slate-50 dark:bg-slate-900 border-b border-slate-200 dark:border-slate-800">
            <tr>
              <th className="px-4 py-3 text-left font-semibold text-slate-600 dark:text-slate-400">Flag</th>
              <th className="px-4 py-3 text-left font-semibold text-slate-600 dark:text-slate-400">Resultado</th>
              <th className="px-4 py-3 text-left font-semibold text-slate-600 dark:text-slate-400">Motivo</th>
              <th className="px-4 py-3 text-left font-semibold text-slate-600 dark:text-slate-400">Variante</th>
              <th className="px-4 py-3 text-left font-semibold text-slate-600 dark:text-slate-400">Usuário</th>
              <th className="px-4 py-3 text-left font-semibold text-slate-600 dark:text-slate-400">Data</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-100 dark:divide-slate-800 bg-white dark:bg-slate-950">
            {loading ? (
              <tr>
                <td colSpan={6} className="px-4 py-8 text-center text-slate-400">
                  Carregando...
                </td>
              </tr>
            ) : items.length === 0 ? (
              <tr>
                <td colSpan={6} className="px-4 py-8 text-center text-slate-400">
                  Nenhuma avaliação encontrada.
                </td>
              </tr>
            ) : (
              items.map((item) => (
                <tr key={item.id} className="hover:bg-slate-50 dark:hover:bg-slate-900/50 transition-colors">
                  <td className="px-4 py-3 font-mono text-sm text-slate-800 dark:text-slate-200">
                    {item.flag_key}
                  </td>
                  <td className="px-4 py-3">
                    {item.result ? (
                      <span className="flex items-center gap-1 text-emerald-600 dark:text-emerald-400 font-medium">
                        <CheckCircle className="w-3.5 h-3.5" /> true
                      </span>
                    ) : (
                      <span className="flex items-center gap-1 text-slate-400 dark:text-slate-500">
                        <XCircle className="w-3.5 h-3.5" /> false
                      </span>
                    )}
                  </td>
                  <td className="px-4 py-3">
                    <span className={`inline-flex px-2 py-0.5 rounded-full text-xs font-medium ${REASON_COLORS[item.reason] ?? REASON_COLORS.default}`}>
                      {REASON_LABELS[item.reason] ?? item.reason}
                    </span>
                  </td>
                  <td className="px-4 py-3 font-mono text-xs text-slate-500 dark:text-slate-400">
                    {item.variant ?? '—'}
                  </td>
                  <td className="px-4 py-3 text-xs text-slate-400 dark:text-slate-500 max-w-[160px] truncate">
                    {item.user_id ?? item.tenant_id ?? '—'}
                  </td>
                  <td className="px-4 py-3 text-xs text-slate-400 dark:text-slate-500 whitespace-nowrap">
                    {new Date(item.evaluated_at).toLocaleString('pt-BR', {
                      day: '2-digit', month: '2-digit', year: '2-digit',
                      hour: '2-digit', minute: '2-digit', second: '2-digit',
                    })}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between text-sm text-slate-500 dark:text-slate-400">
          <span>Página {currentPage} de {totalPages} ({total.toLocaleString('pt-BR')} registros)</span>
          <div className="flex items-center gap-2">
            <button
              onClick={() => setOffset(Math.max(0, offset - PAGE_SIZE))}
              disabled={offset === 0}
              className="p-1.5 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
            >
              <ChevronLeft className="w-4 h-4" />
            </button>
            <button
              onClick={() => setOffset(offset + PAGE_SIZE)}
              disabled={offset + PAGE_SIZE >= total}
              className="p-1.5 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
            >
              <ChevronRight className="w-4 h-4" />
            </button>
          </div>
        </div>
      )}

      <UpgradeModal
        isOpen={gateState.isOpen}
        onClose={closeGate}
        currentPlan={gateState.currentPlan}
        requiredPlan={gateState.requiredPlan}
        feature={gateState.feature}
        upgradeUrl={gateState.upgradeUrl}
        message={gateState.message}
      />
    </div>
  )
}
