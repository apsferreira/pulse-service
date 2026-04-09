// BKL-124: Página de A/B Testing — visualização de flags com variantes e distribuição de tráfego

import { useState, useEffect } from 'react'
import { FlaskConical, ToggleLeft, TrendingUp, Users } from 'lucide-react'
import { flagsApi } from '../services/api'
import type { FeatureFlag, FlagVariant } from '../services/api'

const VARIANT_COLORS = [
  'bg-violet-500',
  'bg-blue-500',
  'bg-emerald-500',
  'bg-amber-500',
  'bg-rose-500',
  'bg-cyan-500',
]

const VARIANT_TEXT_COLORS = [
  'text-violet-600 dark:text-violet-400',
  'text-blue-600 dark:text-blue-400',
  'text-emerald-600 dark:text-emerald-400',
  'text-amber-600 dark:text-amber-400',
  'text-rose-600 dark:text-rose-400',
  'text-cyan-600 dark:text-cyan-400',
]

function VariantBar({ variants }: { variants: FlagVariant[] }) {
  const total = variants.reduce((acc, v) => acc + v.weight, 0)
  return (
    <div className="w-full flex rounded-lg overflow-hidden h-6" title={`Total: ${total}%`}>
      {variants.map((v, i) => (
        <div
          key={v.key}
          className={`${VARIANT_COLORS[i % VARIANT_COLORS.length]} flex items-center justify-center text-white text-[10px] font-medium transition-all`}
          style={{ width: `${total > 0 ? (v.weight / total) * 100 : 0}%` }}
          title={`${v.key}: ${v.weight}%`}
        >
          {v.weight >= 15 ? `${v.key} ${v.weight}%` : v.weight >= 8 ? `${v.weight}%` : ''}
        </div>
      ))}
    </div>
  )
}

function FlagCard({ flag }: { flag: FeatureFlag }) {
  const hasVariants = flag.variants && flag.variants.length > 0
  const weightSum = hasVariants ? flag.variants.reduce((acc, v) => acc + v.weight, 0) : 0
  const isBalanced = weightSum === 100

  return (
    <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-5 space-y-4">
      {/* Header */}
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <div className="flex items-center gap-2 flex-wrap">
            <code className="text-sm font-mono text-slate-900 dark:text-slate-100 font-semibold">
              {flag.key}
            </code>
            {hasVariants && (
              <span className="flex items-center gap-1 px-2 py-0.5 rounded-full bg-violet-100 dark:bg-violet-900/40 text-violet-600 dark:text-violet-300 text-xs font-medium">
                <FlaskConical className="w-3 h-3" />
                A/B — {flag.variants.length} variantes
              </span>
            )}
            {!hasVariants && (
              <span className="flex items-center gap-1 px-2 py-0.5 rounded-full bg-slate-100 dark:bg-slate-800 text-slate-500 text-xs">
                <ToggleLeft className="w-3 h-3" />
                On/Off
              </span>
            )}
          </div>
          {flag.description && (
            <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5 truncate">{flag.description}</p>
          )}
        </div>

        <span className={`shrink-0 text-xs px-2 py-1 rounded-full font-medium ${
          flag.enabled
            ? 'bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400'
            : 'bg-slate-100 dark:bg-slate-800 text-slate-500'
        }`}>
          {flag.enabled ? 'Ativa' : 'Inativa'}
        </span>
      </div>

      {/* Rollout info */}
      <div className="flex items-center gap-3 text-xs text-slate-500 dark:text-slate-400">
        <span className="flex items-center gap-1">
          <Users className="w-3.5 h-3.5" />
          Rollout: <strong className="text-slate-700 dark:text-slate-300">{flag.rollout_percentage ?? 100}%</strong>
        </span>
        {flag.targeting?.user_ids && flag.targeting.user_ids.length > 0 && (
          <span>{flag.targeting.user_ids.length} user(s) targeted</span>
        )}
        {flag.targeting?.tenant_ids && flag.targeting.tenant_ids.length > 0 && (
          <span>{flag.targeting.tenant_ids.length} tenant(s) targeted</span>
        )}
      </div>

      {/* Variants */}
      {hasVariants && (
        <div className="space-y-3">
          <VariantBar variants={flag.variants} />

          {!isBalanced && (
            <p className="text-xs text-amber-600 dark:text-amber-400 flex items-center gap-1">
              <TrendingUp className="w-3.5 h-3.5" />
              Soma dos pesos: {weightSum}% (recomendado: 100%)
            </p>
          )}

          <div className="grid grid-cols-2 sm:grid-cols-3 gap-2">
            {flag.variants.map((v, i) => (
              <div
                key={v.key}
                className="rounded-lg bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 p-3"
              >
                <div className="flex items-center gap-1.5 mb-1">
                  <div className={`w-2.5 h-2.5 rounded-full shrink-0 ${VARIANT_COLORS[i % VARIANT_COLORS.length]}`} />
                  <code className="text-xs font-mono text-slate-700 dark:text-slate-300 truncate">{v.key}</code>
                </div>
                <p className={`text-xl font-bold ${VARIANT_TEXT_COLORS[i % VARIANT_TEXT_COLORS.length]}`}>
                  {v.weight}%
                </p>
                <p className="text-[10px] text-slate-400 mt-0.5">
                  do tráfego
                </p>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Timestamps */}
      <div className="text-[10px] text-slate-400 flex gap-3">
        <span>Criada: {new Date(flag.created_at).toLocaleDateString('pt-BR')}</span>
        <span>Atualizada: {new Date(flag.updated_at).toLocaleDateString('pt-BR')}</span>
      </div>
    </div>
  )
}

export default function ABTesting() {
  const [flags, setFlags] = useState<FeatureFlag[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [filter, setFilter] = useState<'all' | 'ab' | 'onoff'>('all')

  useEffect(() => {
    flagsApi
      .list()
      .then(res => setFlags(res.data))
      .catch(() => setError('Falha ao carregar flags'))
      .finally(() => setLoading(false))
  }, [])

  const abFlags = flags.filter(f => f.variants && f.variants.length > 0)
  const onoffFlags = flags.filter(f => !f.variants || f.variants.length === 0)

  const displayed =
    filter === 'ab' ? abFlags :
    filter === 'onoff' ? onoffFlags :
    flags

  const activeAB = abFlags.filter(f => f.enabled).length
  const totalAB = abFlags.length

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-xl font-bold text-slate-900 dark:text-slate-100 flex items-center gap-2">
          <FlaskConical className="w-5 h-5 text-violet-500" />
          A/B Testing
        </h1>
        <p className="text-sm text-slate-500 dark:text-slate-400 mt-0.5">
          Flags com variantes de distribuição — experimentos ativos no ecossistema IIT
        </p>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
        {[
          { label: 'Total flags', value: flags.length },
          { label: 'Experimentos A/B', value: totalAB },
          { label: 'A/B ativos', value: activeAB },
          { label: 'On/Off flags', value: onoffFlags.length },
        ].map(s => (
          <div key={s.label} className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-4">
            <p className="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wide">{s.label}</p>
            <p className="text-2xl font-bold text-slate-900 dark:text-slate-100 mt-1">{s.value}</p>
          </div>
        ))}
      </div>

      {/* Filter tabs */}
      <div className="flex gap-2">
        {(['all', 'ab', 'onoff'] as const).map(f => (
          <button
            key={f}
            onClick={() => setFilter(f)}
            className={`px-3 py-1.5 rounded-lg text-sm font-medium transition-colors ${
              filter === f
                ? 'bg-violet-600 text-white'
                : 'bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-700 text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-800'
            }`}
          >
            {f === 'all' ? 'Todas' : f === 'ab' ? `A/B (${totalAB})` : `On/Off (${onoffFlags.length})`}
          </button>
        ))}
      </div>

      {/* Content */}
      {loading && (
        <div className="text-center py-16 text-slate-400">
          <div className="w-8 h-8 border-2 border-violet-500 border-t-transparent rounded-full animate-spin mx-auto mb-3" />
          Carregando flags...
        </div>
      )}

      {error && (
        <div className="rounded-xl border border-red-200 dark:border-red-800 bg-red-50 dark:bg-red-900/20 px-4 py-3 text-sm text-red-600 dark:text-red-400">
          {error}
        </div>
      )}

      {!loading && !error && displayed.length === 0 && (
        <div className="text-center py-16 text-slate-400">
          <FlaskConical className="w-12 h-12 mx-auto mb-3 opacity-30" />
          <p className="font-medium text-slate-500">Nenhuma flag encontrada</p>
          {filter === 'ab' && (
            <p className="text-sm mt-1">
              Para criar um experimento A/B, edite uma flag e adicione variantes com pesos percentuais.
            </p>
          )}
        </div>
      )}

      {!loading && !error && displayed.length > 0 && (
        <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
          {displayed.map(flag => (
            <FlagCard key={flag.id} flag={flag} />
          ))}
        </div>
      )}
    </div>
  )
}
