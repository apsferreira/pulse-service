// Cores semânticas do Pulse — usar em todas as páginas
export const STATUS_COLORS = {
  active:     'bg-emerald-100 text-emerald-700 border-emerald-200 dark:bg-emerald-950/30 dark:text-emerald-400 dark:border-emerald-800',
  partial:    'bg-amber-100 text-amber-700 border-amber-200 dark:bg-amber-950/30 dark:text-amber-400 dark:border-amber-800',
  inactive:   'bg-slate-100 text-slate-500 border-slate-200 dark:bg-slate-800 dark:text-slate-400 dark:border-slate-700',
  error:      'bg-red-100 text-red-700 border-red-200 dark:bg-red-950/30 dark:text-red-400 dark:border-red-800',
  primary:    'bg-violet-100 text-violet-700 border-violet-200 dark:bg-violet-950/30 dark:text-violet-400 dark:border-violet-800',
  identified: 'bg-emerald-50 text-emerald-700 border-emerald-200 dark:bg-emerald-950/20 dark:text-emerald-400 dark:border-emerald-800',
  anonymous:  'bg-slate-100 text-slate-500 border-slate-300 dark:bg-slate-800 dark:text-slate-400 dark:border-slate-600',
} as const

export type StatusKey = keyof typeof STATUS_COLORS

export function statusBadge(key: StatusKey): string {
  return `inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium border ${STATUS_COLORS[key]}`
}

// Dot indicator (w-2 h-2 rounded-full)
export const STATUS_DOTS = {
  active:   'bg-emerald-500',
  partial:  'bg-amber-500',
  inactive: 'bg-slate-400',
  error:    'bg-red-500',
  primary:  'bg-violet-600',
} as const

export function statusDot(key: keyof typeof STATUS_DOTS): string {
  return `w-2 h-2 rounded-full ${STATUS_DOTS[key]}`
}
