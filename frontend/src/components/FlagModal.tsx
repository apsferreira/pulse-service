import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { X } from 'lucide-react'
import type { FeatureFlag, CreateFlagPayload, UpdateFlagPayload, Product } from '../services/api'

const PRODUCTS: { value: Product | ''; label: string }[] = [
  { value: '', label: 'Global (sem produto)' },
  { value: 'libri', label: 'Libri' },
  { value: 'nitro', label: 'Nitro' },
  { value: 'brio', label: 'Brio' },
  { value: 'global', label: 'Global (explícito)' },
]

const createSchema = z.object({
  name: z.string().min(2, 'Nome deve ter pelo menos 2 caracteres').max(100, 'Nome muito longo'),
  description: z.string().max(500, 'Descrição muito longa'),
  is_enabled: z.boolean(),
  product: z.string(),
  rollout_pct: z.number().min(0).max(100),
})

const editSchema = z.object({
  description: z.string().max(500, 'Descrição muito longa'),
  is_enabled: z.boolean(),
  rollout_pct: z.number().min(0).max(100),
})

type CreateFormValues = z.infer<typeof createSchema>
type EditFormValues = z.infer<typeof editSchema>

interface FlagModalProps {
  flag?: FeatureFlag | null
  onClose: () => void
  onSubmit: (payload: CreateFlagPayload | UpdateFlagPayload) => Promise<void>
  loading: boolean
}

export default function FlagModal({ flag, onClose, onSubmit, loading }: FlagModalProps) {
  const isEdit = !!flag

  const createForm = useForm<CreateFormValues>({
    resolver: zodResolver(createSchema),
    defaultValues: {
      name: '',
      description: '',
      is_enabled: false,
      product: '',
      rollout_pct: 100,
    },
  })

  const editForm = useForm<EditFormValues>({
    resolver: zodResolver(editSchema),
    defaultValues: {
      description: flag?.description ?? '',
      is_enabled: flag?.is_enabled ?? false,
      rollout_pct: flag?.rollout_pct ?? 100,
    },
  })

  useEffect(() => {
    if (flag) {
      editForm.reset({
        description: flag.description,
        is_enabled: flag.is_enabled,
        rollout_pct: flag.rollout_pct,
      })
    }
  }, [flag, editForm])

  const handleCreate = createForm.handleSubmit(async (values) => {
    const payload: CreateFlagPayload = {
      name: values.name,
      description: values.description,
      is_enabled: values.is_enabled,
      product: (values.product as Product) || null,
      rollout_pct: values.rollout_pct,
    }
    await onSubmit(payload)
  })

  const handleEdit = editForm.handleSubmit(async (values) => {
    const payload: UpdateFlagPayload = {
      description: values.description,
      is_enabled: values.is_enabled,
      rollout_pct: values.rollout_pct,
    }
    await onSubmit(payload)
  })

  const inputClass =
    'w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 text-sm focus:outline-none focus:ring-2 focus:ring-violet-500 focus:border-transparent placeholder:text-slate-400 dark:placeholder:text-slate-500'

  const labelClass = 'block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1'

  const errorClass = 'mt-1 text-xs text-red-500'

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-fade-in"
      onClick={(e) => e.target === e.currentTarget && onClose()}
    >
      <div className="bg-white dark:bg-slate-900 rounded-xl shadow-soft-lg w-full max-w-md border border-slate-200 dark:border-slate-700 animate-scale-in">
        {/* Header */}
        <div className="flex items-center justify-between px-5 py-4 border-b border-slate-200 dark:border-slate-700">
          <h2 className="text-base font-semibold text-slate-900 dark:text-slate-100">
            {isEdit ? `Editar flag: ${flag.name}` : 'Nova feature flag'}
          </h2>
          <button
            onClick={onClose}
            className="p-1.5 rounded-lg text-slate-400 hover:text-slate-600 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* Form */}
        <form onSubmit={isEdit ? handleEdit : handleCreate} className="px-5 py-4 space-y-4">
          {!isEdit && (
            <div>
              <label className={labelClass} htmlFor="flag-name">
                Nome <span className="text-red-500">*</span>
              </label>
              <input
                id="flag-name"
                type="text"
                placeholder="ex: new-checkout-flow"
                className={inputClass}
                {...createForm.register('name')}
              />
              {createForm.formState.errors.name && (
                <p className={errorClass}>{createForm.formState.errors.name.message}</p>
              )}
            </div>
          )}

          <div>
            <label className={labelClass} htmlFor="flag-desc">
              Descrição
            </label>
            <textarea
              id="flag-desc"
              rows={2}
              placeholder="Descreva o propósito desta flag..."
              className={`${inputClass} resize-none`}
              {...(isEdit ? editForm.register('description') : createForm.register('description'))}
            />
            {isEdit
              ? editForm.formState.errors.description && (
                  <p className={errorClass}>{editForm.formState.errors.description.message}</p>
                )
              : createForm.formState.errors.description && (
                  <p className={errorClass}>{createForm.formState.errors.description.message}</p>
                )}
          </div>

          {!isEdit && (
            <div>
              <label className={labelClass} htmlFor="flag-product">
                Produto
              </label>
              <select id="flag-product" className={inputClass} {...createForm.register('product')}>
                {PRODUCTS.map(({ value, label }) => (
                  <option key={value} value={value}>
                    {label}
                  </option>
                ))}
              </select>
            </div>
          )}

          <div>
            <label className={labelClass}>
              Rollout (
              {isEdit
                ? editForm.watch('rollout_pct')
                : createForm.watch('rollout_pct')}
              %)
            </label>
            <input
              type="range"
              min={0}
              max={100}
              step={5}
              className="w-full h-2 bg-slate-200 dark:bg-slate-700 rounded-full appearance-none accent-violet-600 cursor-pointer"
              {...(isEdit
                ? editForm.register('rollout_pct', { valueAsNumber: true })
                : createForm.register('rollout_pct', { valueAsNumber: true }))}
            />
            <div className="flex justify-between text-xs text-slate-400 mt-1">
              <span>0%</span>
              <span>100%</span>
            </div>
          </div>

          <div className="flex items-center gap-3">
            <div className="relative inline-flex items-center">
              <input
                id="flag-enabled"
                type="checkbox"
                className="sr-only peer"
                {...(isEdit ? editForm.register('is_enabled') : createForm.register('is_enabled'))}
              />
              <label
                htmlFor="flag-enabled"
                className="w-10 h-6 bg-slate-300 peer-checked:bg-violet-600 rounded-full cursor-pointer transition-colors peer-focus:ring-2 peer-focus:ring-violet-500 peer-focus:ring-offset-2 dark:peer-focus:ring-offset-slate-900 relative after:content-[''] after:absolute after:top-0.5 after:left-0.5 after:bg-white after:rounded-full after:w-5 after:h-5 after:transition-transform peer-checked:after:translate-x-4 dark:bg-slate-600"
              />
            </div>
            <label htmlFor="flag-enabled" className="text-sm text-slate-700 dark:text-slate-300 cursor-pointer select-none">
              Ativa ao criar
            </label>
          </div>

          {/* Actions */}
          <div className="flex justify-end gap-2 pt-2 border-t border-slate-100 dark:border-slate-800">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 rounded-lg text-sm font-medium text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
            >
              Cancelar
            </button>
            <button
              type="submit"
              disabled={loading}
              className="px-4 py-2 rounded-lg text-sm font-medium bg-violet-600 hover:bg-violet-700 disabled:opacity-60 disabled:cursor-not-allowed text-white transition-colors flex items-center gap-2"
            >
              {loading && (
                <div className="w-3.5 h-3.5 border-2 border-white/40 border-t-white rounded-full animate-spin" />
              )}
              {isEdit ? 'Salvar alterações' : 'Criar flag'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
