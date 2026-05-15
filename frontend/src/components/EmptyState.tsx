import { type LucideIcon } from 'lucide-react'

interface EmptyStateProps {
  icon: LucideIcon
  title: string
  description?: string
  action?: {
    label: string
    onClick: () => void
  }
}

export function EmptyState({ icon: Icon, title, description, action }: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center py-16 text-center">
      <Icon className="w-10 h-10 text-slate-300 dark:text-slate-600 mb-3" strokeWidth={1.5} />
      <p className="text-base font-medium text-slate-500 dark:text-slate-400">{title}</p>
      {description && (
        <p className="text-sm text-slate-400 dark:text-slate-500 mt-1 max-w-xs">{description}</p>
      )}
      {action && (
        <button
          onClick={action.onClick}
          className="mt-4 px-4 py-2 rounded-lg bg-violet-600 hover:bg-violet-500 text-white text-sm font-medium transition-colors"
        >
          {action.label}
        </button>
      )}
    </div>
  )
}
