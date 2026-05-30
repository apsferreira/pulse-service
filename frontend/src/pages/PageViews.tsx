// PULSE-PV-03: Dashboard de Page Views — Top Pages + User Journey

import { useState, useEffect, useCallback } from 'react'
import { Globe, Clock, Eye, ArrowRight, Search } from 'lucide-react'
import { pageViewsApi } from '../services/api'
import type { TopPageItem, UserJourneyItem } from '../services/api'
import { useProject } from '../components/Layout'

// ─── Helpers ─────────────────────────────────────────────────────────────────

function formatMs(ms: number | undefined): string {
  if (ms === undefined || ms === null) return '—'
  if (ms < 1000) return `${Math.round(ms)}ms`
  const s = (ms / 1000).toFixed(1)
  return `${s}s`
}

function formatDate(iso: string): string {
  try {
    return new Date(iso).toLocaleString('pt-BR', {
      day: '2-digit',
      month: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    })
  } catch {
    return iso
  }
}

// ─── Top Pages ───────────────────────────────────────────────────────────────

interface TopPagesProps {
  product: string
  from: string
  to: string
}

function TopPagesTable({ product, from, to }: TopPagesProps) {
  const [items, setItems] = useState<TopPageItem[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!product) return
    setLoading(true)
    setError(null)
    pageViewsApi
      .topPages(product, from, to, 15)
      .then((r) => setItems(r.data.items ?? []))
      .catch((err: unknown) => {
        const e = err as { response?: { data?: { error?: string } }; message?: string }
        setError(e?.response?.data?.error ?? e?.message ?? 'Erro ao carregar top pages')
      })
      .finally(() => setLoading(false))
  }, [product, from, to])

  if (!product)
    return (
      <div className="text-sm text-slate-400 text-center py-8">
        Selecione um projeto para ver as páginas mais acessadas.
      </div>
    )

  return (
    <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 overflow-hidden">
      <div className="px-5 py-4 border-b border-slate-200 dark:border-slate-800 flex items-center gap-2">
        <Globe className="w-4 h-4 text-violet-500" />
        <h2 className="text-sm font-semibold text-slate-900 dark:text-white">Top Pages</h2>
        <span className="text-xs text-slate-400 ml-auto">produto: {product}</span>
      </div>

      {loading && (
        <div className="py-10 text-center text-sm text-slate-400">Carregando...</div>
      )}
      {!loading && error && (
        <div className="py-8 text-center text-sm text-red-500">{error}</div>
      )}
      {!loading && !error && items.length === 0 && (
        <div className="py-10 text-center text-sm text-slate-400">
          Nenhuma page view registrada no período.
        </div>
      )}
      {!loading && !error && items.length > 0 && (
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-slate-100 dark:border-slate-800 bg-slate-50 dark:bg-slate-800/50">
              <th className="px-5 py-2.5 text-left text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wide">
                Página
              </th>
              <th className="px-5 py-2.5 text-right text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wide">
                Views
              </th>
              <th className="px-5 py-2.5 text-right text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wide">
                Tempo médio
              </th>
            </tr>
          </thead>
          <tbody>
            {items.map((item, idx) => (
              <tr
                key={item.page_path + idx}
                className="border-b border-slate-100 dark:border-slate-800 hover:bg-slate-50 dark:hover:bg-slate-800/40 transition-colors"
              >
                <td className="px-5 py-3 font-mono text-xs text-slate-700 dark:text-slate-300 max-w-xs truncate">
                  {item.page_path}
                </td>
                <td className="px-5 py-3 text-right font-semibold text-slate-900 dark:text-white">
                  <span className="inline-flex items-center gap-1.5">
                    <Eye className="w-3.5 h-3.5 text-violet-400" />
                    {item.total_views.toLocaleString('pt-BR')}
                  </span>
                </td>
                <td className="px-5 py-3 text-right text-slate-500 dark:text-slate-400">
                  <span className="inline-flex items-center gap-1.5">
                    <Clock className="w-3.5 h-3.5 text-slate-400" />
                    {formatMs(item.avg_time_on_page_ms)}
                  </span>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  )
}

// ─── User Journey ─────────────────────────────────────────────────────────────

interface UserJourneyProps {
  product: string
}

function UserJourneyTimeline({ product }: UserJourneyProps) {
  const [userId, setUserId] = useState('')
  const [search, setSearch] = useState('')
  const [items, setItems] = useState<UserJourneyItem[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [searched, setSearched] = useState(false)

  const fetchJourney = useCallback(
    (uid: string) => {
      if (!uid.trim()) return
      setLoading(true)
      setError(null)
      setSearched(true)
      pageViewsApi
        .userJourney(uid.trim(), product || undefined, 30)
        .then((r) => setItems(r.data.items ?? []))
        .catch((err: unknown) => {
          const e = err as { response?: { data?: { error?: string } }; message?: string }
          setError(e?.response?.data?.error ?? e?.message ?? 'Erro ao carregar jornada')
        })
        .finally(() => setLoading(false))
    },
    [product],
  )

  function handleSearch(e: React.FormEvent) {
    e.preventDefault()
    setUserId(search)
    fetchJourney(search)
  }

  return (
    <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 overflow-hidden">
      <div className="px-5 py-4 border-b border-slate-200 dark:border-slate-800 flex items-center gap-2">
        <ArrowRight className="w-4 h-4 text-violet-500" />
        <h2 className="text-sm font-semibold text-slate-900 dark:text-white">User Journey</h2>
        <span className="text-xs text-slate-400 ml-auto">últimas 30 páginas por usuário</span>
      </div>

      <div className="px-5 py-4 border-b border-slate-100 dark:border-slate-800">
        <form onSubmit={handleSearch} className="flex gap-2">
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
            <input
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="UUID do usuário..."
              className="w-full pl-9 pr-3 py-2 text-sm rounded-lg border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 text-slate-900 dark:text-white placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-violet-500"
            />
          </div>
          <button
            type="submit"
            disabled={!search.trim()}
            className="px-4 py-2 text-sm font-medium rounded-lg bg-violet-600 text-white hover:bg-violet-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            Buscar
          </button>
        </form>
      </div>

      <div className="px-5 py-4">
        {loading && (
          <div className="text-center text-sm text-slate-400 py-6">Carregando...</div>
        )}
        {!loading && error && (
          <div className="text-center text-sm text-red-500 py-6">{error}</div>
        )}
        {!loading && !error && !searched && (
          <div className="text-center text-sm text-slate-400 py-6">
            Digite o UUID de um usuário para ver sua jornada.
          </div>
        )}
        {!loading && !error && searched && items.length === 0 && (
          <div className="text-center text-sm text-slate-400 py-6">
            Nenhuma page view encontrada para este usuário.
          </div>
        )}
        {!loading && !error && items.length > 0 && (
          <div className="space-y-0">
            <p className="text-xs text-slate-400 mb-4">
              Usuário: <span className="font-mono text-violet-500">{userId}</span>
              {product && (
                <span className="ml-2">
                  — produto: <span className="font-mono">{product}</span>
                </span>
              )}
            </p>
            <ol className="relative border-l border-slate-200 dark:border-slate-700 ml-3">
              {items.map((item, idx) => (
                <li key={item.id + idx} className="mb-4 ml-5">
                  <span className="absolute -left-[9px] flex items-center justify-center w-4 h-4 rounded-full bg-violet-100 dark:bg-violet-900 ring-2 ring-white dark:ring-slate-900">
                    <span className="w-1.5 h-1.5 rounded-full bg-violet-500" />
                  </span>
                  <div className="flex items-start justify-between gap-2">
                    <div>
                      <span className="font-mono text-xs text-slate-700 dark:text-slate-300">
                        {item.page_path ?? '—'}
                      </span>
                      {item.page_title && (
                        <span className="ml-2 text-xs text-slate-400">({item.page_title})</span>
                      )}
                      {item.referrer && (
                        <div className="text-xs text-slate-400 mt-0.5">
                          de: <span className="font-mono">{item.referrer}</span>
                        </div>
                      )}
                    </div>
                    <div className="shrink-0 text-right">
                      <div className="text-xs text-slate-400">{formatDate(item.created_at)}</div>
                      {item.time_on_page_ms !== undefined && (
                        <div className="text-xs text-slate-400 flex items-center gap-1 justify-end">
                          <Clock className="w-3 h-3" />
                          {formatMs(item.time_on_page_ms)}
                        </div>
                      )}
                    </div>
                  </div>
                </li>
              ))}
            </ol>
          </div>
        )}
      </div>
    </div>
  )
}

// ─── Página principal ─────────────────────────────────────────────────────────

export default function PageViews() {
  const { projectSlug } = useProject()

  const now = new Date()
  const thirtyDaysAgo = new Date(now)
  thirtyDaysAgo.setDate(thirtyDaysAgo.getDate() - 30)

  const [from, setFrom] = useState(thirtyDaysAgo.toISOString().slice(0, 10))
  const [to, setTo] = useState(now.toISOString().slice(0, 10))

  const product = projectSlug ?? ''

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6 space-y-6">
      {/* Cabeçalho */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-xl font-bold text-slate-900 dark:text-white flex items-center gap-2">
            <Globe className="w-5 h-5 text-violet-500" />
            Page Views
          </h1>
          <p className="text-sm text-slate-500 dark:text-slate-400 mt-0.5">
            Rastreamento automático de páginas via SDK — PULSE-PV-01/02/03
          </p>
        </div>

        {/* Seletor de período */}
        <div className="flex items-center gap-2 text-sm">
          <label className="text-slate-500 dark:text-slate-400 whitespace-nowrap">De</label>
          <input
            type="date"
            value={from}
            onChange={(e) => setFrom(e.target.value)}
            className="px-3 py-1.5 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-900 dark:text-white text-sm focus:outline-none focus:ring-2 focus:ring-violet-500"
          />
          <label className="text-slate-500 dark:text-slate-400">até</label>
          <input
            type="date"
            value={to}
            onChange={(e) => setTo(e.target.value)}
            className="px-3 py-1.5 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-900 dark:text-white text-sm focus:outline-none focus:ring-2 focus:ring-violet-500"
          />
        </div>
      </div>

      {/* Aviso se nenhum projeto selecionado */}
      {!product && (
        <div className="rounded-xl border border-amber-200 dark:border-amber-800 bg-amber-50 dark:bg-amber-950/30 px-5 py-4 text-sm text-amber-700 dark:text-amber-300">
          Selecione um projeto no menu superior para filtrar os dados de page views.
        </div>
      )}

      {/* Top Pages */}
      <TopPagesTable product={product} from={from} to={to} />

      {/* User Journey */}
      <UserJourneyTimeline product={product} />
    </div>
  )
}
