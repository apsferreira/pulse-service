import { useNavigate } from 'react-router-dom'
import { X, Lock, CheckCircle, ExternalLink } from 'lucide-react'

// ── Types ─────────────────────────────────────────────────────────────────────

export interface UpgradeModalProps {
  isOpen: boolean
  onClose: () => void
  currentPlan: string
  requiredPlan: string
  feature: string
  upgradeUrl: string
  message: string
}

// ── Plan benefits (hardcoded) ─────────────────────────────────────────────────

const PLAN_BENEFITS: Record<string, string[]> = {
  starter: ['Flags ilimitadas', '3 projetos', '500k eventos/mês'],
  growth: ['10 projetos', '5M eventos/mês', 'A/B testing com p-value', 'Health Score'],
  pro: ['Projetos ilimitados', 'Eventos ilimitados', 'SLA 99,9%', 'Onboarding dedicado'],
}

const PLAN_PRICES: Record<string, string> = {
  starter: 'R$ 399/mês',
  growth: 'R$ 1.099/mês',
  pro: 'R$ 2.399/mês',
}

const PLAN_LABELS: Record<string, string> = {
  starter: 'Starter',
  growth: 'Growth',
  pro: 'Pro',
}

// ── Component ─────────────────────────────────────────────────────────────────

export default function UpgradeModal({
  isOpen,
  onClose,
  currentPlan,
  requiredPlan,
  feature,
  upgradeUrl,
  message,
}: UpgradeModalProps) {
  const navigate = useNavigate()

  if (!isOpen) return null

  const benefits = PLAN_BENEFITS[requiredPlan] ?? []
  const price = PLAN_PRICES[requiredPlan] ?? ''
  const planLabel = PLAN_LABELS[requiredPlan] ?? requiredPlan

  const handleViewPlans = () => {
    onClose()
    navigate('/billing')
  }

  const handleKeyDown = (e: React.KeyboardEvent<HTMLDivElement>) => {
    if (e.key === 'Escape') onClose()
  }

  return (
    <div
      role="dialog"
      aria-modal="true"
      aria-labelledby="upgrade-modal-title"
      className="fixed inset-0 z-50 flex items-center justify-center p-4"
      onKeyDown={handleKeyDown}
    >
      {/* Overlay */}
      <div
        className="absolute inset-0 bg-black/70 backdrop-blur-sm"
        onClick={onClose}
        aria-hidden="true"
      />

      {/* Card */}
      <div
        className="relative w-full max-w-md rounded-2xl border border-violet-800/40 shadow-2xl"
        style={{ background: '#110f1a' }}
      >
        {/* Close */}
        <button
          onClick={onClose}
          className="absolute right-4 top-4 p-1.5 rounded-lg text-slate-400 hover:bg-white/10 transition-colors"
          aria-label="Fechar modal"
        >
          <X className="w-4 h-4" />
        </button>

        <div className="px-6 pt-8 pb-6 space-y-5">
          {/* Lock icon + title */}
          <div className="flex flex-col items-center text-center gap-3">
            <div
              className="w-14 h-14 rounded-2xl flex items-center justify-center"
              style={{ background: '#4c1d9520' }}
            >
              <Lock
                className="w-7 h-7"
                style={{ color: '#8B5CF6' }}
                aria-hidden="true"
              />
            </div>
            <div>
              <h2
                id="upgrade-modal-title"
                className="text-lg font-bold text-slate-100"
              >
                Recurso bloqueado
              </h2>
              <p className="text-sm text-slate-400 mt-1">
                Para{' '}
                <span className="text-slate-200 font-medium">{feature}</span>, você
                precisa do plano{' '}
                <span style={{ color: '#a78bfa' }} className="font-semibold">
                  {planLabel}
                </span>
                .
              </p>
            </div>
          </div>

          {/* Message */}
          {message && (
            <p className="text-sm text-slate-400 text-center">{message}</p>
          )}

          {/* Plan highlight */}
          <div
            className="rounded-xl border border-violet-700/40 p-4 space-y-3"
            style={{ background: '#1a1530' }}
          >
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-semibold text-slate-100">{planLabel}</p>
                <p className="text-xs text-slate-400">Plano necessário</p>
              </div>
              <span
                className="text-xl font-bold"
                style={{ color: '#a78bfa' }}
              >
                {price}
              </span>
            </div>

            <ul className="space-y-2">
              {benefits.map((b) => (
                <li key={b} className="flex items-center gap-2 text-xs text-slate-300">
                  <CheckCircle
                    className="w-3.5 h-3.5 flex-shrink-0"
                    style={{ color: '#10b981' }}
                    aria-hidden="true"
                  />
                  {b}
                </li>
              ))}
            </ul>
          </div>

          {/* Plano atual */}
          <p className="text-center text-xs text-slate-500">
            Plano atual:{' '}
            <span className="text-slate-400 font-medium capitalize">{currentPlan}</span>
          </p>

          {/* CTAs */}
          <div className="flex flex-col gap-2">
            <a
              href={upgradeUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="w-full flex items-center justify-center gap-2 py-2.5 rounded-xl text-sm font-semibold text-white transition-colors"
              style={{ background: '#7c3aed' }}
              onMouseEnter={(e) => {
                ;(e.currentTarget as HTMLAnchorElement).style.background = '#6d28d9'
              }}
              onMouseLeave={(e) => {
                ;(e.currentTarget as HTMLAnchorElement).style.background = '#7c3aed'
              }}
            >
              Fazer upgrade agora
              <ExternalLink className="w-3.5 h-3.5" aria-hidden="true" />
            </a>

            <button
              onClick={handleViewPlans}
              className="w-full py-2.5 rounded-xl text-sm font-medium text-violet-300 border border-violet-800/50 hover:bg-violet-900/20 transition-colors"
            >
              Ver planos
            </button>
          </div>

          {/* Dismiss */}
          <div className="text-center">
            <button
              onClick={onClose}
              className="text-xs text-slate-500 hover:text-slate-400 transition-colors underline underline-offset-2"
            >
              Continuar no plano gratuito
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
