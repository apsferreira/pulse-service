import { useCallback, useEffect, useState } from 'react'
import { CheckSquare, RefreshCw, Clock, CheckCircle2, XCircle, MessageSquare, X } from 'lucide-react'
import toast from 'react-hot-toast'
import { approvalsApi, type Approval, type ApprovalStatus } from '../services/pulseApi'

// ── Helpers ──────────────────────────────────────────────────────────────────

function formatDate(iso: string): string {
  return new Intl.DateTimeFormat('pt-BR', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(iso))
}

function formatRelative(iso: string): string {
  const diff = Date.now() - new Date(iso).getTime()
  const mins = Math.floor(diff / 60_000)
  if (mins < 1) return 'agora'
  if (mins < 60) return `${mins}m atrás`
  const hours = Math.floor(mins / 60)
  if (hours < 24) return `${hours}h atrás`
  return formatDate(iso)
}

// ── Status badge ─────────────────────────────────────────────────────────────

interface StatusBadgeProps {
  status: ApprovalStatus
}

function StatusBadge({ status }: StatusBadgeProps) {
  if (status === 'pending') {
    return (
      <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium bg-amber-50 dark:bg-amber-950/40 text-amber-700 dark:text-amber-400 border border-amber-200 dark:border-amber-800">
        <Clock className="w-3 h-3" aria-hidden="true" />
        Pendente
      </span>
    )
  }
  if (status === 'approved') {
    return (
      <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium bg-emerald-50 dark:bg-emerald-950/40 text-emerald-700 dark:text-emerald-400 border border-emerald-200 dark:border-emerald-800">
        <CheckCircle2 className="w-3 h-3" aria-hidden="true" />
        Aprovada
      </span>
    )
  }
  return (
    <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium bg-red-50 dark:bg-red-950/40 text-red-700 dark:text-red-400 border border-red-200 dark:border-red-800">
      <XCircle className="w-3 h-3" aria-hidden="true" />
      Rejeitada
    </span>
  )
}

// ── Changes diff ─────────────────────────────────────────────────────────────

interface ChangesDiffProps {
  changes: Record<string, unknown>
}

function ChangesDiff({ changes }: ChangesDiffProps) {
  const entries = Object.entries(changes)
  if (entries.length === 0) {
    return <p className="text-xs text-slate-400 italic">Sem mudanças especificadas</p>
  }
  return (
    <div className="space-y-1">
      {entries.map(([key, value]) => {
        const v = value as { before?: unknown; after?: unknown } | unknown
        const isBefore = typeof v === 'object' && v !== null && 'before' in v
        if (isBefore) {
          const { before, after } = v as { before: unknown; after: unknown }
          return (
            <div key={key} className="flex items-center gap-2 text-xs font-mono">
              <span className="text-slate-500 dark:text-slate-400 min-w-[80px] font-sans font-medium">
                {key}
              </span>
              <span className="px-1.5 py-0.5 rounded bg-red-50 dark:bg-red-950/30 text-red-600 dark:text-red-400 line-through">
                {JSON.stringify(before)}
              </span>
              <span className="text-slate-400" aria-hidden="true">→</span>
              <span className="px-1.5 py-0.5 rounded bg-emerald-50 dark:bg-emerald-950/30 text-emerald-600 dark:text-emerald-400">
                {JSON.stringify(after)}
              </span>
            </div>
          )
        }
        return (
          <div key={key} className="flex items-center gap-2 text-xs font-mono">
            <span className="text-slate-500 dark:text-slate-400 min-w-[80px] font-sans font-medium">
              {key}
            </span>
            <span className="px-1.5 py-0.5 rounded bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300">
              {JSON.stringify(v)}
            </span>
          </div>
        )
      })}
    </div>
  )
}

// ── Review modal ─────────────────────────────────────────────────────────────

interface ReviewModalProps {
  approval: Approval
  action: 'approve' | 'reject'
  onClose: () => void
  onSuccess: () => void
}

function ReviewModal({ approval, action, onClose, onSuccess }: ReviewModalProps) {
  const [comment, setComment] = useState('')
  const [submitting, setSubmitting] = useState(false)

  const isApprove = action === 'approve'

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!comment.trim()) {
      toast.error('O comentário é obrigatório')
      return
    }
    setSubmitting(true)
    try {
      await approvalsApi.review(approval.id, { action, comment: comment.trim() })
      toast.success(isApprove ? 'Aprovação confirmada' : 'Aprovação rejeitada')
      onSuccess()
      onClose()
    } catch {
      toast.error('Falha ao processar revisão')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <>
      {/* Backdrop */}
      <div
        className="fixed inset-0 bg-black/50 z-40 backdrop-blur-sm"
        onClick={onClose}
        aria-hidden="true"
      />

      {/* Modal */}
      <div
        role="dialog"
        aria-modal="true"
        aria-label={isApprove ? 'Confirmar aprovação' : 'Confirmar rejeição'}
        className="fixed inset-0 z-50 flex items-center justify-center p-4"
      >
        <div className="w-full max-w-md bg-white dark:bg-slate-900 rounded-2xl shadow-2xl border border-slate-200 dark:border-slate-800 overflow-hidden">
          {/* Header */}
          <div className="flex items-center justify-between px-5 py-4 border-b border-slate-200 dark:border-slate-800">
            <div className="flex items-center gap-2">
              {isApprove ? (
                <CheckCircle2 className="w-4 h-4 text-emerald-500" aria-hidden="true" />
              ) : (
                <XCircle className="w-4 h-4 text-red-500" aria-hidden="true" />
              )}
              <h2 className="font-semibold text-slate-800 dark:text-slate-100 text-sm">
                {isApprove ? 'Aprovar mudança' : 'Rejeitar mudança'}
              </h2>
            </div>
            <button
              onClick={onClose}
              className="p-1.5 rounded-lg text-slate-500 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
              aria-label="Fechar modal"
            >
              <X className="w-4 h-4" />
            </button>
          </div>

          {/* Context */}
          <div className="px-5 py-4 border-b border-slate-100 dark:border-slate-800 bg-slate-50 dark:bg-slate-800/40">
            <p className="text-xs text-slate-500 dark:text-slate-400 font-medium mb-1">Flag</p>
            <code className="text-sm font-mono font-semibold text-violet-600 dark:text-violet-400">
              {approval.flag_key}
            </code>
          </div>

          {/* Form */}
          <form onSubmit={handleSubmit} className="px-5 py-4 space-y-4">
            <div>
              <label
                htmlFor="review-comment"
                className="block text-xs font-semibold text-slate-600 dark:text-slate-400 mb-1.5"
              >
                Comentário <span className="text-red-500" aria-hidden="true">*</span>
              </label>
              <textarea
                id="review-comment"
                value={comment}
                onChange={(e) => setComment(e.target.value)}
                rows={3}
                placeholder={
                  isApprove
                    ? 'Descreva por que está aprovando esta mudança…'
                    : 'Explique o motivo da rejeição…'
                }
                className="w-full px-3 py-2 text-sm rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-800 dark:text-slate-200 placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-emerald-500/40 focus:border-emerald-500 resize-none"
                required
              />
            </div>

            <div className="flex gap-3 justify-end">
              <button
                type="button"
                onClick={onClose}
                disabled={submitting}
                className="px-4 py-2 text-sm font-medium rounded-lg text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors disabled:opacity-50"
              >
                Cancelar
              </button>
              <button
                type="submit"
                disabled={submitting || !comment.trim()}
                className={`px-4 py-2 text-sm font-medium rounded-lg text-white transition-colors disabled:opacity-50 ${
                  isApprove
                    ? 'bg-emerald-600 hover:bg-emerald-700'
                    : 'bg-red-600 hover:bg-red-700'
                }`}
              >
                {submitting ? (
                  <RefreshCw className="w-4 h-4 animate-spin" aria-hidden="true" />
                ) : isApprove ? (
                  'Aprovar'
                ) : (
                  'Rejeitar'
                )}
              </button>
            </div>
          </form>
        </div>
      </div>
    </>
  )
}

// ── Approval card ─────────────────────────────────────────────────────────────

interface ApprovalCardProps {
  approval: Approval
  onAction: (approval: Approval, action: 'approve' | 'reject') => void
  readonly?: boolean
}

function ApprovalCard({ approval, onAction, readonly = false }: ApprovalCardProps) {
  return (
    <article className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-4 space-y-3">
      {/* Top row */}
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2 flex-wrap">
            <code className="text-sm font-mono font-semibold text-violet-600 dark:text-violet-400 truncate">
              {approval.flag_key}
            </code>
            <StatusBadge status={approval.status} />
          </div>
          <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
            Solicitado por{' '}
            <span className="font-medium text-slate-700 dark:text-slate-300">
              {approval.requested_by}
            </span>
            {' · '}
            <time dateTime={approval.created_at} title={formatDate(approval.created_at)}>
              {formatRelative(approval.created_at)}
            </time>
          </p>
        </div>
      </div>

      {/* Message */}
      {approval.message && (
        <div className="flex items-start gap-2 text-sm text-slate-600 dark:text-slate-400">
          <MessageSquare className="w-3.5 h-3.5 mt-0.5 flex-shrink-0 text-slate-400" aria-hidden="true" />
          <p className="italic">&ldquo;{approval.message}&rdquo;</p>
        </div>
      )}

      {/* Changes diff */}
      <div className="rounded-lg bg-slate-50 dark:bg-slate-800/50 border border-slate-100 dark:border-slate-800 p-3">
        <p className="text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-2">
          Mudanças
        </p>
        <ChangesDiff changes={approval.changes} />
      </div>

      {/* Comment (history) */}
      {approval.comment && (
        <div className="rounded-lg bg-slate-50 dark:bg-slate-800/50 border border-slate-100 dark:border-slate-800 p-3">
          <p className="text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-1">
            Comentário do revisor
          </p>
          <p className="text-sm text-slate-700 dark:text-slate-300 italic">
            &ldquo;{approval.comment}&rdquo;
          </p>
        </div>
      )}

      {/* Actions */}
      {!readonly && approval.status === 'pending' && (
        <div className="flex gap-2 pt-1">
          <button
            onClick={() => onAction(approval, 'approve')}
            className="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 text-sm font-medium rounded-lg bg-emerald-600 hover:bg-emerald-700 text-white transition-colors"
          >
            <CheckCircle2 className="w-3.5 h-3.5" aria-hidden="true" />
            Aprovar
          </button>
          <button
            onClick={() => onAction(approval, 'reject')}
            className="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 text-sm font-medium rounded-lg bg-red-50 dark:bg-red-950/30 hover:bg-red-100 dark:hover:bg-red-950/50 text-red-700 dark:text-red-400 border border-red-200 dark:border-red-800 transition-colors"
          >
            <XCircle className="w-3.5 h-3.5" aria-hidden="true" />
            Rejeitar
          </button>
        </div>
      )}
    </article>
  )
}

// ── Empty state ───────────────────────────────────────────────────────────────

interface EmptyStateProps {
  tab: 'pending' | 'history'
}

function EmptyState({ tab }: EmptyStateProps) {
  return (
    <div className="rounded-xl border border-dashed border-slate-300 dark:border-slate-700 p-12 flex flex-col items-center text-center gap-3">
      <div className="w-12 h-12 rounded-2xl bg-slate-100 dark:bg-slate-800 flex items-center justify-center">
        <CheckSquare className="w-6 h-6 text-slate-400" aria-hidden="true" />
      </div>
      <div>
        <p className="text-sm font-semibold text-slate-700 dark:text-slate-300">
          {tab === 'pending' ? 'Nenhuma aprovação pendente' : 'Nenhum histórico encontrado'}
        </p>
        <p className="text-xs text-slate-400 mt-0.5">
          {tab === 'pending'
            ? 'Aprovações de mudanças em feature flags aparecerão aqui.'
            : 'Aprovações e rejeições anteriores aparecerão aqui.'}
        </p>
      </div>
    </div>
  )
}

// ── Main page ─────────────────────────────────────────────────────────────────

type TabKey = 'pending' | 'history'

export default function Approvals() {
  const [approvals, setApprovals] = useState<Approval[]>([])
  const [loading, setLoading] = useState(true)
  const [tab, setTab] = useState<TabKey>('pending')
  const [reviewTarget, setReviewTarget] = useState<{
    approval: Approval
    action: 'approve' | 'reject'
  } | null>(null)

  const fetchApprovals = useCallback(async () => {
    setLoading(true)
    try {
      const res = await approvalsApi.listAll()
      const data = res.data
      setApprovals(Array.isArray(data) ? data : [])
    } catch {
      toast.error('Falha ao carregar aprovações')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void fetchApprovals()
  }, [fetchApprovals])

  const pending = approvals.filter((a) => a.status === 'pending')
  const history = approvals.filter((a) => a.status !== 'pending')

  const displayed = tab === 'pending' ? pending : history

  function handleAction(approval: Approval, action: 'approve' | 'reject') {
    setReviewTarget({ approval, action })
  }

  return (
    <div className="space-y-5">
      {/* Page header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-xl font-bold text-slate-900 dark:text-slate-100">Aprovações</h1>
          <p className="text-sm text-slate-500 dark:text-slate-400 mt-0.5">
            Revisão de mudanças em feature flags
          </p>
        </div>
        <button
          onClick={fetchApprovals}
          disabled={loading}
          className="self-start sm:self-auto p-2 rounded-lg text-slate-500 hover:bg-slate-100 dark:text-slate-400 dark:hover:bg-slate-800 transition-colors disabled:opacity-50"
          aria-label="Recarregar aprovações"
        >
          <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
        </button>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 p-1 rounded-xl bg-slate-100 dark:bg-slate-800 w-fit">
        {(['pending', 'history'] as const).map((t) => {
          const isActive = tab === t
          const label = t === 'pending' ? 'Pendentes' : 'Histórico'
          const count = t === 'pending' ? pending.length : history.length
          return (
            <button
              key={t}
              onClick={() => setTab(t)}
              className={`flex items-center gap-1.5 px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                isActive
                  ? 'bg-white dark:bg-slate-900 text-slate-800 dark:text-slate-100 shadow-sm'
                  : 'text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-300'
              }`}
            >
              {label}
              {count > 0 && (
                <span
                  className={`inline-flex items-center justify-center w-5 h-5 rounded-full text-xs font-bold ${
                    t === 'pending'
                      ? 'bg-amber-100 dark:bg-amber-950/50 text-amber-700 dark:text-amber-400'
                      : 'bg-slate-200 dark:bg-slate-700 text-slate-600 dark:text-slate-400'
                  }`}
                  aria-label={`${count} ${label.toLowerCase()}`}
                >
                  {count}
                </span>
              )}
            </button>
          )
        })}
      </div>

      {/* Content */}
      {loading ? (
        <div className="flex items-center justify-center h-40">
          <RefreshCw className="w-5 h-5 animate-spin text-slate-400" />
        </div>
      ) : displayed.length === 0 ? (
        <EmptyState tab={tab} />
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {displayed.map((approval) => (
            <ApprovalCard
              key={approval.id}
              approval={approval}
              onAction={handleAction}
              readonly={tab === 'history'}
            />
          ))}
        </div>
      )}

      {/* Review modal */}
      {reviewTarget !== null && (
        <ReviewModal
          approval={reviewTarget.approval}
          action={reviewTarget.action}
          onClose={() => setReviewTarget(null)}
          onSuccess={fetchApprovals}
        />
      )}
    </div>
  )
}
