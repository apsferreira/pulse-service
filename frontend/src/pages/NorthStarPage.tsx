// BKL-767: Dashboard de North Star Metrics — 6 produtos IIT
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
      {/* Ponto final destacado */}
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

type BadgeStatus = 'green' | 'yellow' | 'red'

interface NorthStarProduct {
  name: string
  icon: React.ElementType
  nsm: string
  unit: string
  colorClass: string
  sparkColor: string
  badgeBg: string
  currentValue: number
  prevValue: number
  sparkData: number[]
  status: BadgeStatus
}

function generateSparkline(base: number): number[] {
  return Array.from({ length: 8 }, () =>
    Math.round(base * (0.85 + Math.random() * 0.3))
  )
}

const PRODUCTS: NorthStarProduct[] = [
  {
    name: 'SocialMake',
    icon: Share2,
    nsm: 'Posts publicados',
    unit: 'por semana',
    colorClass: 'text-amber-600 dark:text-amber-400',
    sparkColor: '#d97706',
    badgeBg: 'bg-amber-100 text-amber-700 dark:bg-amber-900/40 dark:text-amber-300',
    currentValue: 47,
    prevValue: 38,
    sparkData: generateSparkline(45),
    status: 'green',
  },
  {
    name: 'Nitro',
    icon: Timer,
    nsm: 'Sessões de foco completadas',
    unit: 'por semana',
    colorClass: 'text-indigo-600 dark:text-indigo-400',
    sparkColor: '#4f46e5',
    badgeBg: 'bg-indigo-100 text-indigo-700 dark:bg-indigo-900/40 dark:text-indigo-300',
    currentValue: 312,
    prevValue: 298,
    sparkData: generateSparkline(305),
    status: 'green',
  },
  {
    name: 'Libri',
    icon: BookOpen,
    nsm: 'Livros concluídos',
    unit: 'no mês',
    colorClass: 'text-emerald-600 dark:text-emerald-400',
    sparkColor: '#059669',
    badgeBg: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300',
    currentValue: 84,
    prevValue: 91,
    sparkData: generateSparkline(87),
    status: 'yellow',
  },
  {
    name: 'MeuIngresso',
    icon: Ticket,
    nsm: 'Ingressos vendidos',
    unit: 'por evento',
    colorClass: 'text-sky-600 dark:text-sky-400',
    sparkColor: '#0284c7',
    badgeBg: 'bg-sky-100 text-sky-700 dark:bg-sky-900/40 dark:text-sky-300',
    currentValue: 215,
    prevValue: 180,
    sparkData: generateSparkline(200),
    status: 'green',
  },
  {
    name: 'Brio',
    icon: MessageSquare,
    nsm: 'Conversas resolvidas pelo agente',
    unit: '% do total',
    colorClass: 'text-violet-600 dark:text-violet-400',
    sparkColor: '#7c3aed',
    badgeBg: 'bg-violet-100 text-violet-700 dark:bg-violet-900/40 dark:text-violet-300',
    currentValue: 68,
    prevValue: 72,
    sparkData: generateSparkline(70),
    status: 'yellow',
  },
  {
    name: 'Pulse',
    icon: Zap,
    nsm: 'Projetos com flag tracked',
    unit: 'ativos',
    colorClass: 'text-orange-600 dark:text-orange-400',
    sparkColor: '#ea580c',
    badgeBg: 'bg-orange-100 text-orange-700 dark:bg-orange-900/40 dark:text-orange-300',
    currentValue: 9,
    prevValue: 6,
    sparkData: generateSparkline(8),
    status: 'green',
  },
]

const STATUS_CONFIG: Record<
  BadgeStatus,
  { label: string; className: string }
> = {
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

interface NorthStarCardProps {
  product: NorthStarProduct
}

function NorthStarCard({ product }: NorthStarCardProps) {
  const Icon = product.icon
  const status = STATUS_CONFIG[product.status]

  return (
    <article
      className="rounded-2xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 p-5 flex flex-col gap-4 shadow-sm hover:shadow-md transition-shadow"
      aria-label={`North Star — ${product.name}`}
    >
      {/* Header */}
      <div className="flex items-start justify-between gap-2">
        <div className="flex items-center gap-2">
          <div className={`p-2 rounded-lg bg-slate-100 dark:bg-slate-800 ${product.colorClass}`}>
            <Icon className="w-4 h-4" aria-hidden="true" />
          </div>
          <span className="font-semibold text-slate-900 dark:text-slate-100 text-sm">
            {product.name}
          </span>
        </div>
        <span
          className={`text-xs font-medium px-2 py-0.5 rounded-full ${status.className}`}
          aria-label={`Status: ${status.label}`}
        >
          {status.label}
        </span>
      </div>

      {/* NSM label */}
      <div>
        <p className="text-xs text-slate-500 dark:text-slate-400 leading-tight">{product.nsm}</p>
        <p className="text-xs text-slate-400 dark:text-slate-500">{product.unit}</p>
      </div>

      {/* Value + sparkline */}
      <div className="flex items-end justify-between gap-3">
        <div>
          <p className={`text-3xl font-black tabular-nums ${product.colorClass}`}>
            {product.currentValue.toLocaleString('pt-BR')}
            {product.unit === '% do total' && (
              <span className="text-lg font-semibold">%</span>
            )}
          </p>
          <DeltaBadge current={product.currentValue} prev={product.prevValue} />
        </div>
        <Sparkline data={product.sparkData} color={product.sparkColor} />
      </div>
    </article>
  )
}

export default function NorthStarPage() {
  return (
    <div className="space-y-6">
      {/* Page header */}
      <div className="flex items-center justify-between flex-wrap gap-2">
        <div>
          <h1 className="text-2xl font-bold text-slate-900 dark:text-slate-100">
            North Star Metrics
          </h1>
          <p className="text-sm text-slate-500 dark:text-slate-400 mt-0.5">
            Indicadores principais dos 6 produtos IIT — dados placeholder, backend em breve
          </p>
        </div>
        <span className="text-xs text-slate-400 dark:text-slate-500 bg-slate-100 dark:bg-slate-800 px-2.5 py-1 rounded-lg">
          Semana atual
        </span>
      </div>

      {/* Grid: 3 colunas desktop / 2 tablet / 1 mobile */}
      <div
        className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4"
        role="list"
        aria-label="North Star Metrics por produto"
      >
        {PRODUCTS.map((product) => (
          <div key={product.name} role="listitem">
            <NorthStarCard product={product} />
          </div>
        ))}
      </div>

      {/* Disclaimer */}
      <p className="text-xs text-slate-400 dark:text-slate-500 text-center pt-2">
        Os sparklines exibem dados de placeholder (8 semanas simuladas). A integração com a API de eventos está prevista no backlog.
      </p>
    </div>
  )
}
