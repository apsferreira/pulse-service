import { Activity, Clock } from 'lucide-react'

export default function Events() {
  return (
    <div className="space-y-5">
      {/* Page header */}
      <div>
        <h1 className="text-xl font-bold text-slate-900 dark:text-slate-100">Eventos</h1>
        <p className="text-sm text-slate-500 dark:text-slate-400 mt-0.5">
          Rastreamento de eventos em tempo real
        </p>
      </div>

      {/* Placeholder */}
      <div className="rounded-xl border border-dashed border-violet-300 dark:border-violet-700 bg-violet-50/50 dark:bg-violet-950/20 p-12 flex flex-col items-center text-center gap-4">
        <div className="w-16 h-16 rounded-2xl bg-violet-100 dark:bg-violet-900/40 flex items-center justify-center">
          <Activity className="w-8 h-8 text-violet-500 dark:text-violet-400" />
        </div>
        <div className="space-y-1">
          <h2 className="text-base font-semibold text-slate-800 dark:text-slate-200">
            Em breve: rastreamento de eventos em tempo real
          </h2>
          <p className="text-sm text-slate-500 dark:text-slate-400 max-w-md">
            Aqui você poderá visualizar os eventos disparados pelas aplicações Libri, Nitro e Brio —
            conversões, cliques, erros e métricas de uso em tempo real.
          </p>
        </div>
        <div className="flex items-center gap-2 px-4 py-2 rounded-full bg-violet-100 dark:bg-violet-900/40 text-violet-600 dark:text-violet-400 text-xs font-medium">
          <Clock className="w-3.5 h-3.5" />
          Planejado para uma próxima versão
        </div>
      </div>

      {/* Preview cards */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 opacity-40 pointer-events-none select-none">
        {[
          { label: 'Eventos hoje', value: '—', desc: 'últimas 24h' },
          { label: 'Usuários únicos', value: '—', desc: 'últimas 24h' },
          { label: 'Taxa de erros', value: '—%', desc: 'últimas 24h' },
        ].map(({ label, value, desc }) => (
          <div
            key={label}
            className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-4"
          >
            <p className="text-xs text-slate-500 dark:text-slate-400 font-medium">{label}</p>
            <p className="text-3xl font-bold text-slate-800 dark:text-slate-200 mt-1">{value}</p>
            <p className="text-xs text-slate-400 mt-0.5">{desc}</p>
          </div>
        ))}
      </div>
    </div>
  )
}
