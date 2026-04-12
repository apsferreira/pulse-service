// BKL-769: North Star Metrics — dados reais via GET /api/v1/north-star
import { useEffect, useState } from 'react'
import {
  Share2,
  Timer,
  BookOpen,
  Ticket,
  MessageSquare,
  Zap,
  TrendingUp,
  TrendingDown,
  Minus,
} from 'lucide-react'

// ─── Sparkline ───────────────────────────────────────────────────────────────

interface SparklineProps {
  data: number[]
  color: string
}

function Sparkline({ data, color }: SparklineProps) {
  const min = Math.min(...data)
  const max = Math.max(...data)
  const range = max - min || 1
  const width = 96
  const height = 32
  const step = width / (data.length - 1)

  const points = data
    .map((v, i) => {
      const x = i * step
      const y = height - ((v - min) / range) * height
      return `${x},${y}`
    })
    .join(' ')

  return (
    <svg
      width={width}
      height={height}
      viewBox={`0 0 ${width} ${height}`}
      aria-hidden="true"
      className="overflow-visible"
    >
      <polyline
        fill="none"
        stroke={color}
        strokeWidth="2"
        strokeLinejoin="round"
        strokeLinecap="round"
        points={points}
      />
      {data.length > 0 && (
        <circle
          cx={(data.length - 1) * step}
          cy={height - ((data[data.length - 1] - min) / range) * height}
          r="3"
          fill={color}
        />
      )}
    </svg>
  )
}

// ─── Tipos ───────────────────────────────────────────────────────────────────

type BadgeStatus = 'green' | 'yellow' | 'red'

interface ApiMetric {
  product: string
  metric_name: string
  unit: string
  current_value: number
  previous_value: number
  sparkline: number[]
  updated_at: string
}

interface ProductConfig {
  product: string
  icon: React.ElementType
  colorClass: string
  sparkColor: string
}

const PRODUCT_CONFIG: Record<string, ProductConfig> = {
  socialmake: {
    product: 'SocialMake',
    icon: Share2,
    colorClass: 'text-amber-600 dark:text-amber-400',
    sparkColor: '#d97706',
  },
  nitro: {
    product: 'Nitro',
    icon: Timer,
    colorClass: 'text-indigo-600 dark:text-indigo-400',
    sparkColor: '#4f46e5',
  },
  libri: {
    product: 'Libri',
    icon: BookOpen,
    colorClass: 'text-emerald-600 dark:text-emerald-400',
    sparkColor: '#059669',
  },
  meuingresso: {
    product: 'MeuIngresso',
    icon: Ticket,
    colorClass: 'text-sky-600 dark:text-sky-400',
    sparkColor: '#0284c7',
  },
  brio: {
    product: 'Brio',
    icon: MessageSquare,
    colorClass: 'text-violet-600 dark:text-violet-400',
    sparkColor: '#7c3aed',
  },
  pulse: {
    product: 'Pulse',
    icon: Zap,
    colorClass: 'text-orange-600 dark:text-orange-400',
    sparkColor: '#ea580c',
  },
}

const STATUS_CONFIG: Record<BadgeStatus, { label: string; className: string }> = {
  green: {
    label: 'No alvo',
    className: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300',
  },
  yellow: {
    label: 'Atenção',
    className: 'bg-amber-100 text-amber-700 dark:bg-amber-900/40 dark:text-amber-300',
  },
  red: {
    label: 'Crítico',
    className: 'bg-red-100 text-red-700 dark:bg-red-900/40 dark:text-red-300',
  },
}

function deriveStatus(current: number, previous: number): BadgeStatus {
  if (previous === 0) return 'green'
  const pct = ((current - previous) / previous) * 100
  if (pct >= -5) return 'green'
  if (pct >= -20) return 'yellow'
  return 'red'
}

// ─── DeltaBadge ──────────────────────────────────────────────────────────────

function DeltaBadge({ current, prev }: { current: number; prev: number }) {
  if (prev === 0) return null
  const pct = ((current - prev) / prev) * 100
  const sign = pct > 0 ? '+' : ''
  const isUp = pct > 0
  const isFlat = Math.abs(pct) < 0.5

  const cls = isFlat
    ? 'text-slate-500 dark:text-slate-400'
    : isUp
      ? 'text-emerald-600 dark:text-emerald-400'
      : 'text-red-500 dark:text-red-400'

  const Icon = isFlat ? Minus : isUp ? TrendingUp : TrendingDown

  return (
    <span className={`inline-flex items-center gap-0.5 text-xs font-medium ${cls}`}>
      <Icon className="w-3.5 h-3.5" aria-hidden="true" />
      {sign}{pct.toFixed(1)}% vs semana anterior
    </span>
  )
}

// ─── Skeleton card ───────────────────────────────────────────────────────────

function SkeletonCard() {
  return (
    <article className="rounded-2xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 p-5 flex flex-col gap-4 shadow-sm animate-pulse">
      <div className="flex items-start justify-between gap-2">
        <div className="flex items-center gap-2">
          <div className="w-8 h-8 rounded-lg bg-slate-200 dark:bg-slate-700" />
          <div className="w-20 h-4 rounded bg-slate-200 dark:bg-slate-700" />
        </div>
        <div className="w-14 h-5 rounded-full bg-slate-200 dark:bg-slate-700" />
      </div>
      <div className="space-y-1">
        <div className="w-36 h-3 rounded bg-slate-200 dark:bg-slate-700" />
        <div className="w-16 h-3 rounded bg-slate-200 dark:bg-slate-700" />
      </div>
      <div className="flex items-end justify-between gap-3">
        <div className="space-y-1">
          <div className="w-20 h-9 rounded bg-slate-200 dark:bg-slate-700" />
          <div className="w-28 h-3 rounded bg-slate-200 dark:bg-slate-700" />
        </div>
        <div className="w-24 h-8 rounded bg-slate-200 dark:bg-slate-700" />
      </div>
    </article>
  )
}

// ─── NorthStarCard ───────────────────────────────────────────────────────────

interface NorthStarCardProps {
  metric: ApiMetric
}

function NorthStarCard({ metric }: NorthStarCardProps) {
  const cfg = PRODUCT_CONFIG[metric.product]
  const Icon = cfg?.icon ?? Zap
  const status = deriveStatus(metric.current_value, metric.previous_value)
  const statusCfg = STATUS_CONFIG[status]

  const displayName = cfg?.product ?? metric.product
  const colorClass = cfg?.colorClass ?? 'text-slate-600 dark:text-slate-400'
  const sparkColor = cfg?.sparkColor ?? '#64748b'

  return (
    <article
      className="rounded-2xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 p-5 flex flex-col gap-4 shadow-sm hover:shadow-md transition-shadow"
      aria-label={`North Star — ${displayName}`}
    >
      {/* Header */}
      <div className="flex items-start justify-between gap-2">
        <div className="flex items-center gap-2">
          <div className={`p-2 rounded-lg bg-slate-100 dark:bg-slate-800 ${colorClass}`}>
            <Icon className="w-4 h-4" aria-hidden="true" />
          </div>
          <span className="font-semibold text-slate-900 dark:text-slate-100 text-sm">
            {displayName}
          </span>
        </div>
        <span
          className={`text-xs font-medium px-2 py-0.5 rounded-full ${statusCfg.className}`}
          aria-label={`Status: ${statusCfg.label}`}
        >
          {statusCfg.label}
        </span>
      </div>

      {/* NSM label */}
      <div>
        <p className="text-xs text-slate-500 dark:text-slate-400 leading-tight">{metric.metric_name}</p>
        <p className="text-xs text-slate-400 dark:text-slate-500">{metric.unit}</p>
      </div>

      {/* Value + sparkline */}
      <div className="flex items-end justify-between gap-3">
        <div>
          <p className={`text-3xl font-black tabular-nums ${colorClass}`}>
            {metric.current_value.toLocaleString('pt-BR')}
            {metric.unit === '% do total' && (
              <span className="text-lg font-semibold">%</span>
            )}
          </p>
          <DeltaBadge current={metric.current_value} prev={metric.previous_value} />
        </div>
        <Sparkline data={metric.sparkline} color={sparkColor} />
      </div>
    </article>
  )
}

// ─── Page ─────────────────────────────────────────────────────────────────────

const PULSE_API_URL = import.meta.env.VITE_PULSE_API_URL ?? ''

export default function NorthStarPage() {
  const [metrics, setMetrics] = useState<ApiMetric[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [generatedAt, setGeneratedAt] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError(null)

    fetch(`${PULSE_API_URL}/api/v1/north-star`)
      .then((res) => {
        if (!res.ok) throw new Error(`HTTP ${res.status}`)
        return res.json()
      })
      .then((data: { metrics: ApiMetric[]; generated_at: string }) => {
        if (cancelled) return
        setMetrics(data.metrics ?? [])
        setGeneratedAt(data.generated_at ?? null)
        setLoading(false)
      })
      .catch((err: Error) => {
        if (cancelled) return
        setError(err.message)
        setLoading(false)
      })

    return () => {
      cancelled = true
    }
  }, [])

  const formattedAt = generatedAt
    ? new Date(generatedAt).toLocaleString('pt-BR', {
        day: '2-digit',
        month: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
      })
    : null

  return (
    <div className="space-y-6">
      {/* Page header */}
      <div className="flex items-center justify-between flex-wrap gap-2">
        <div>
          <h1 className="text-2xl font-bold text-slate-900 dark:text-slate-100">
            North Star Metrics
          </h1>
          <p className="text-sm text-slate-500 dark:text-slate-400 mt-0.5">
            Indicadores principais dos 6 produtos IIT — dados reais de eventos
          </p>
        </div>
        <span className="text-xs text-slate-400 dark:text-slate-500 bg-slate-100 dark:bg-slate-800 px-2.5 py-1 rounded-lg">
          Semana atual
        </span>
      </div>

      {/* Error banner */}
      {error && (
        <div
          role="alert"
          className="rounded-xl border border-red-200 bg-red-50 dark:bg-red-900/20 dark:border-red-800 px-4 py-3 text-sm text-red-700 dark:text-red-400"
        >
          Erro ao carregar métricas: {error}. Os valores exibidos podem estar desatualizados.
        </div>
      )}

      {/* Grid: 3 colunas desktop / 2 tablet / 1 mobile */}
      <div
        className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4"
        role="list"
        aria-label="North Star Metrics por produto"
      >
        {loading
          ? Array.from({ length: 6 }).map((_, i) => (
              <div key={i} role="listitem">
                <SkeletonCard />
              </div>
            ))
          : metrics.map((metric) => (
              <div key={metric.product} role="listitem">
                <NorthStarCard metric={metric} />
              </div>
            ))}
      </div>

      {/* Footer */}
      {!loading && !error && formattedAt && (
        <p className="text-xs text-slate-400 dark:text-slate-500 text-center pt-2">
          Dados reais de eventos. Atualizado em {formattedAt}.
        </p>
      )}
    </div>
  )
}
