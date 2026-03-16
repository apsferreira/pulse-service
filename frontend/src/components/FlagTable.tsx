import { useState } from 'react'
import { Pencil, Trash2, AlertTriangle } from 'lucide-react'
import type { FeatureFlag, Product } from '../services/api'

const PRODUCT_BADGES: Record<string, string> = {
  libri: 'bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-300',
  nitro: 'bg-orange-100 text-orange-700 dark:bg-orange-900/40 dark:text-orange-300',
  brio: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300',
  global: 'bg-slate-100 text-slate-600 dark:bg-slate-700 dark:text-slate-400',
}

function ProductBadge({ product }: { product: Product | null }) {
  if (!product) return <span className="text-slate-400 text-xs">—</span>
  const classes = PRODUCT_BADGES[product] ?? PRODUCT_BADGES.global
  return (
    <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${classes}`}>
      {product}
    </span>
  )
}

function RolloutBar({ pct }: { pct: number }) {
  return (
    <div className="flex items-center gap-2">
      <div className="w-16 h-1.5 bg-slate-200 dark:bg-slate-700 rounded-full overflow-hidden">
        <div
          className="h-full bg-violet-500 rounded-full transition-all"
          style={{ width: `${pct}%` }}
        />
      </div>
      <span className="text-xs text-slate-500 dark:text-slate-400 font-mono">{pct}%</span>
    </div>
  )
}

function ToggleSwitch({
  flagId,
  enabled,
  onToggle,
}: {
  flagId: string
  enabled: boolean
  onToggle: (id: string, enabled: boolean) => Promise<void>
}) {
  const [loading, setLoading] = useState(false)

  const handleClick = async () => {
    setLoading(true)
    try {
      await onToggle(flagId, !enabled)
    } finally {
      setLoading(false)
    }
  }

  return (
    <button
      onClick={handleClick}
      disabled={loading}
      className={[
        'relative inline-flex w-10 h-6 rounded-full transition-colors focus:outline-none focus:ring-2 focus:ring-violet-500 focus:ring-offset-2 dark:focus:ring-offset-slate-900',
        enabled ? 'bg-violet-600' : 'bg-slate-300 dark:bg-slate-600',
        loading ? 'opacity-60 cursor-wait' : 'cursor-pointer',
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

interface DeleteConfirmProps {
  flag: FeatureFlag
  onConfirm: () => Promise<void>
  onCancel: () => void
  loading: boolean
}

function DeleteConfirm({ flag, onConfirm, onCancel, loading }: DeleteConfirmProps) {
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
              A flag <strong className="text-slate-700 dark:text-slate-300">{flag.name}</strong> será
              removida permanentemente. Esta ação não pode ser desfeita.
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
            className="px-4 py-2 rounded-lg text-sm font-medium bg-red-600 hover:bg-red-700 disabled:opacity-60 disabled:cursor-not-allowed text-white transition-colors flex items-center gap-2"
          >
            {loading && (
              <div className="w-3.5 h-3.5 border-2 border-white/40 border-t-white rounded-full animate-spin" />
            )}
            Deletar
          </button>
        </div>
      </div>
    </div>
  )
}

interface FlagTableProps {
  flags: FeatureFlag[]
  onEdit: (flag: FeatureFlag) => void
  onDelete: (id: string) => Promise<void>
  onToggle: (id: string, enabled: boolean) => Promise<void>
  loading: boolean
}

export default function FlagTable({ flags, onEdit, onDelete, onToggle, loading }: FlagTableProps) {
  const [deleteTarget, setDeleteTarget] = useState<FeatureFlag | null>(null)
  const [deleteLoading, setDeleteLoading] = useState(false)

  const handleConfirmDelete = async () => {
    if (!deleteTarget) return
    setDeleteLoading(true)
    try {
      await onDelete(deleteTarget.id)
      setDeleteTarget(null)
    } finally {
      setDeleteLoading(false)
    }
  }

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

  if (loading) {
    return (
      <div className="flex items-center justify-center py-16 text-slate-400">
        <div className="w-6 h-6 border-2 border-violet-500 border-t-transparent rounded-full animate-spin mr-3" />
        <span className="text-sm">Carregando flags...</span>
      </div>
    )
  }

  if (flags.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-slate-400 gap-2">
        <svg className="w-10 h-10 opacity-30" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={1.5}
            d="M3 3h2l.4 2M7 13h10l4-8H5.4M7 13L5.4 5M7 13l-2 9m13-9l2 9M9 21h6"
          />
        </svg>
        <p className="text-sm">Nenhuma feature flag encontrada</p>
        <p className="text-xs">Crie a primeira flag usando o botão acima.</p>
      </div>
    )
  }

  return (
    <>
      {/* Desktop table */}
      <div className="hidden md:block overflow-x-auto rounded-xl border border-slate-200 dark:border-slate-800">
        <table className="w-full text-sm">
          <thead>
            <tr className="bg-slate-50 dark:bg-slate-800/50 text-left text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide">
              <th className="px-4 py-3">Nome</th>
              <th className="px-4 py-3">Descrição</th>
              <th className="px-4 py-3">Produto</th>
              <th className="px-4 py-3">Rollout</th>
              <th className="px-4 py-3 text-center">Ativo</th>
              <th className="px-4 py-3">Criado em</th>
              <th className="px-4 py-3 text-right">Ações</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-100 dark:divide-slate-800">
            {flags.map((flag) => (
              <tr
                key={flag.id}
                className="bg-white dark:bg-slate-900 hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-colors"
              >
                <td className="px-4 py-3 font-mono text-xs font-medium text-slate-800 dark:text-slate-200">
                  {flag.name}
                </td>
                <td className="px-4 py-3 text-slate-500 dark:text-slate-400 max-w-xs truncate">
                  {flag.description || <span className="text-slate-300 dark:text-slate-600">—</span>}
                </td>
                <td className="px-4 py-3">
                  <ProductBadge product={flag.product} />
                </td>
                <td className="px-4 py-3">
                  <RolloutBar pct={flag.rollout_pct} />
                </td>
                <td className="px-4 py-3 text-center">
                  <ToggleSwitch flagId={flag.id} enabled={flag.is_enabled} onToggle={onToggle} />
                </td>
                <td className="px-4 py-3 text-xs text-slate-400 font-mono whitespace-nowrap">
                  {formatDate(flag.created_at)}
                </td>
                <td className="px-4 py-3">
                  <div className="flex items-center justify-end gap-1">
                    <button
                      onClick={() => onEdit(flag)}
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
        {flags.map((flag) => (
          <div
            key={flag.id}
            className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-4 space-y-3"
          >
            <div className="flex items-start justify-between gap-3">
              <div className="min-w-0">
                <p className="font-mono text-sm font-medium text-slate-800 dark:text-slate-200 truncate">
                  {flag.name}
                </p>
                {flag.description && (
                  <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5 line-clamp-2">
                    {flag.description}
                  </p>
                )}
              </div>
              <ToggleSwitch flagId={flag.id} enabled={flag.is_enabled} onToggle={onToggle} />
            </div>
            <div className="flex items-center gap-3 flex-wrap">
              <ProductBadge product={flag.product} />
              <RolloutBar pct={flag.rollout_pct} />
            </div>
            <div className="flex items-center justify-between pt-2 border-t border-slate-100 dark:border-slate-800">
              <span className="text-xs text-slate-400 font-mono">{formatDate(flag.created_at)}</span>
              <div className="flex items-center gap-1">
                <button
                  onClick={() => onEdit(flag)}
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

      {/* Delete confirm dialog */}
      {deleteTarget && (
        <DeleteConfirm
          flag={deleteTarget}
          onConfirm={handleConfirmDelete}
          onCancel={() => setDeleteTarget(null)}
          loading={deleteLoading}
        />
      )}
    </>
  )
}
