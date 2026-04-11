import { useState } from 'react'
import { Pencil, Trash2, AlertTriangle } from 'lucide-react'
import type { FeatureFlag } from '../services/api'

function ToggleSwitch({
  flagKey,
  enabled,
  onToggle,
}: {
  flagKey: string
  enabled: boolean
  onToggle: (key: string, enabled: boolean) => Promise<void>
}) {
  const [loading, setLoading] = useState(false)

  const handleClick = async () => {
    setLoading(true)
    try {
      await onToggle(flagKey, !enabled)
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
              A flag <strong className="text-slate-700 dark:text-slate-300">{flag.key}</strong> sera
              removida permanentemente. Esta acao nao pode ser desfeita.
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
  onDelete: (key: string) => Promise<void>
  onToggle: (key: string, enabled: boolean) => Promise<void>
  loading: boolean
}

export default function FlagTable({ flags, onEdit, onDelete, onToggle, loading }: FlagTableProps) {
  const [deleteTarget, setDeleteTarget] = useState<FeatureFlag | null>(null)
  const [deleteLoading, setDeleteLoading] = useState(false)

  const handleConfirmDelete = async () => {
    if (!deleteTarget) return
    setDeleteLoading(true)
    try {
      await onDelete(deleteTarget.key)
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
    // BKL-1051: empty state com CTA e instruções de instalação do SDK
    return (
      <div className="rounded-xl border border-dashed border-violet-300 dark:border-violet-700 bg-violet-50/50 dark:bg-violet-950/20 p-10 flex flex-col items-center text-center gap-6">
        <div className="w-16 h-16 rounded-2xl bg-violet-100 dark:bg-violet-900/40 flex items-center justify-center">
          <svg className="w-8 h-8 text-violet-500 dark:text-violet-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M3 21v-4m0 0V5a2 2 0 012-2h6.5l1 1H21l-3 6 3 6h-8.5l-1-1H5a2 2 0 00-2 2zm9-13.5V9" />
          </svg>
        </div>
        <div className="space-y-1">
          <h2 className="text-base font-semibold text-slate-800 dark:text-slate-200">Nenhuma feature flag criada</h2>
          <p className="text-sm text-slate-500 dark:text-slate-400 max-w-md">
            Crie sua primeira flag acima e use o SDK para consultá-la nas suas aplicações.
          </p>
        </div>
        <div className="w-full max-w-xl text-left space-y-3">
          <p className="text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide">Instalação rápida</p>
          <div className="rounded-lg bg-slate-900 text-slate-100 p-4 text-xs font-mono space-y-1 overflow-x-auto">
            <p className="text-slate-400"># npm</p>
            <p>npm install @iit/pulse-sdk</p>
          </div>
          <div className="rounded-lg bg-slate-900 text-slate-100 p-4 text-xs font-mono space-y-1 overflow-x-auto">
            <p className="text-slate-400">// Uso</p>
            <p>{'import PulseClient from "@iit/pulse-sdk";'}</p>
            <p>{''}</p>
            <p>{'const pulse = new PulseClient({'}</p>
            <p>{'  baseURL: "https://pulse-api.institutoitinerante.com.br",'}</p>
            <p>{'});'}</p>
            <p>{''}</p>
            <p>{'const flags = await pulse.getFlags();'}</p>
            <p>{'if (flags["minha-nova-flag"]) {'}</p>
            <p>{'  // mostrar nova experiência'}</p>
            <p>{'}'}</p>
          </div>
        </div>
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
              <th className="px-4 py-3">Key</th>
              <th className="px-4 py-3">Descricao</th>
              <th className="px-4 py-3 text-center">Ativo</th>
              <th className="px-4 py-3 text-center">Rollout</th>
              <th className="px-4 py-3">Targeting</th>
              <th className="px-4 py-3">Criado em</th>
              <th className="px-4 py-3 text-right">Acoes</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-100 dark:divide-slate-800">
            {flags.map((flag) => {
              const rollout = flag.rollout_percentage ?? 100
              const hasTargeting = (flag.targeting?.user_ids?.length ?? 0) + (flag.targeting?.tenant_ids?.length ?? 0) > 0
              const targetingCount = (flag.targeting?.user_ids?.length ?? 0) + (flag.targeting?.tenant_ids?.length ?? 0)
              return (
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
                    <ToggleSwitch flagKey={flag.key} enabled={flag.enabled} onToggle={onToggle} />
                  </td>
                  <td className="px-4 py-3 text-center">
                    <span className={[
                      'inline-flex items-center px-2 py-0.5 rounded-full text-xs font-semibold',
                      rollout === 100
                        ? 'bg-green-100 dark:bg-green-900/40 text-green-700 dark:text-green-300'
                        : rollout === 0
                        ? 'bg-slate-100 dark:bg-slate-800 text-slate-500'
                        : 'bg-amber-100 dark:bg-amber-900/40 text-amber-700 dark:text-amber-300',
                    ].join(' ')}>
                      {rollout}%
                    </span>
                  </td>
                  <td className="px-4 py-3">
                    {hasTargeting ? (
                      <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full bg-violet-100 dark:bg-violet-900/40 text-violet-700 dark:text-violet-300 text-xs font-semibold">
                        {targetingCount} ID{targetingCount !== 1 ? 's' : ''}
                      </span>
                    ) : (
                      <span className="text-slate-300 dark:text-slate-600 text-xs">--</span>
                    )}
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
              )
            })}
          </tbody>
        </table>
      </div>

      {/* Mobile cards */}
      <div className="md:hidden space-y-3">
        {flags.map((flag) => {
          const rollout = flag.rollout_percentage ?? 100
          const targetingCount = (flag.targeting?.user_ids?.length ?? 0) + (flag.targeting?.tenant_ids?.length ?? 0)
          return (
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
                <ToggleSwitch flagKey={flag.key} enabled={flag.enabled} onToggle={onToggle} />
              </div>
              <div className="flex items-center gap-2">
                <span className={[
                  'inline-flex items-center px-2 py-0.5 rounded-full text-xs font-semibold',
                  rollout === 100
                    ? 'bg-green-100 dark:bg-green-900/40 text-green-700 dark:text-green-300'
                    : rollout === 0
                    ? 'bg-slate-100 dark:bg-slate-800 text-slate-500'
                    : 'bg-amber-100 dark:bg-amber-900/40 text-amber-700 dark:text-amber-300',
                ].join(' ')}>
                  {rollout}%
                </span>
                {targetingCount > 0 && (
                  <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full bg-violet-100 dark:bg-violet-900/40 text-violet-700 dark:text-violet-300 text-xs font-semibold">
                    {targetingCount} ID{targetingCount !== 1 ? 's' : ''}
                  </span>
                )}
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
          )
        })}
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
