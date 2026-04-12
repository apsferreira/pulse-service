// BKL-123: Formulário de feature flag com rollout % e targeting por user_id/tenant_id
// BKL-124: Suporte a variantes A/B testing
// BKL-989: Visual rule builder para targeting avançado (substitui campos user_ids/tenant_ids)

import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { X, Plus, Trash2, FlaskConical, SlidersHorizontal } from 'lucide-react'
import type { FeatureFlag, CreateFlagPayload, UpdateFlagPayload, FlagVariant } from '../services/api'
import RuleBuilder, { type RuleGroup, ruleGroupsToBackend, backendToRuleGroups } from './RuleBuilder'

const createSchema = z.object({
  key: z
    .string()
    .min(2, 'Key deve ter pelo menos 2 caracteres')
    .max(100, 'Key muito longa')
    .regex(/^[a-zA-Z0-9_\-\.]+$/, 'Apenas letras, numeros, underscore, hifen e ponto'),
  description: z.string().max(500, 'Descricao muito longa'),
  enabled: z.boolean(),
  rollout_percentage: z.number().min(0).max(100),
})

const editSchema = z.object({
  description: z.string().max(500, 'Descricao muito longa'),
  enabled: z.boolean(),
  rollout_percentage: z.number().min(0).max(100),
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

  // BKL-989: Rule builder state — substitui campos user_ids/tenant_ids por regras visuais
  const [ruleGroups, setRuleGroups] = useState<RuleGroup[]>(
    backendToRuleGroups(flag?.targeting as Record<string, unknown> | null | undefined)
  )
  const [showRuleBuilder, setShowRuleBuilder] = useState(
    // Abrir automaticamente se já houver regras salvas no formato v2
    !!(flag?.targeting && (flag.targeting as Record<string, unknown>).version === 2)
  )

  // BKL-124: Variantes A/B
  const [variants, setVariants] = useState<FlagVariant[]>(flag?.variants ?? [])
  const [variantKeyInput, setVariantKeyInput] = useState('')
  const [variantWeightInput, setVariantWeightInput] = useState(50)
  const variantWeightSum = variants.reduce((acc, v) => acc + v.weight, 0)

  const createForm = useForm<CreateFormValues>({
    resolver: zodResolver(createSchema),
    defaultValues: {
      key: '',
      description: '',
      enabled: false,
      rollout_percentage: 100,
    },
  })

  const editForm = useForm<EditFormValues>({
    resolver: zodResolver(editSchema),
    defaultValues: {
      description: flag?.description ?? '',
      enabled: flag?.enabled ?? false,
      rollout_percentage: flag?.rollout_percentage ?? 100,
    },
  })

  useEffect(() => {
    if (flag) {
      editForm.reset({
        description: flag.description,
        enabled: flag.enabled,
        rollout_percentage: flag.rollout_percentage ?? 100,
      })
      setRuleGroups(backendToRuleGroups(flag.targeting as Record<string, unknown> | null | undefined))
      setShowRuleBuilder(!!(flag.targeting && (flag.targeting as Record<string, unknown>).version === 2))
      setVariants(flag.variants ?? [])
    }
  }, [flag, editForm])

  // BKL-989: buildTargeting serializa as RuleGroups no formato v2 para o backend
  const buildTargeting = () => {
    if (!showRuleBuilder) return undefined
    const backend = ruleGroupsToBackend(ruleGroups) as { groups?: unknown[] }
    if (!backend.groups || backend.groups.length === 0) return undefined
    return backend
  }

  const addVariant = () => {
    const key = variantKeyInput.trim()
    if (!key || variants.some(v => v.key === key)) return
    setVariants([...variants, { key, weight: variantWeightInput }])
    setVariantKeyInput('')
    setVariantWeightInput(50)
  }

  const removeVariant = (key: string) => setVariants(variants.filter(v => v.key !== key))

  const updateVariantWeight = (key: string, weight: number) => {
    setVariants(variants.map(v => v.key === key ? { ...v, weight } : v))
  }

  const handleCreate = createForm.handleSubmit(async (values) => {
    const payload: CreateFlagPayload = {
      key: values.key,
      description: values.description,
      enabled: values.enabled,
      rollout_percentage: values.rollout_percentage,
      targeting: buildTargeting(),
      variants: variants.length > 0 ? variants : undefined,
    }
    await onSubmit(payload)
  })

  const handleEdit = editForm.handleSubmit(async (values) => {
    const payload: UpdateFlagPayload = {
      description: values.description,
      enabled: values.enabled,
      rollout_percentage: values.rollout_percentage,
      targeting: buildTargeting(),
      variants: variants.length > 0 ? variants : [],
    }
    await onSubmit(payload)
  })

  const inputClass =
    'w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 text-sm focus:outline-none focus:ring-2 focus:ring-violet-500 focus:border-transparent placeholder:text-slate-400 dark:placeholder:text-slate-500'

  const labelClass = 'block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1'
  const errorClass = 'mt-1 text-xs text-red-500'

  const rolloutValue = isEdit
    ? editForm.watch('rollout_percentage')
    : createForm.watch('rollout_percentage')

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-fade-in"
      onClick={(e) => e.target === e.currentTarget && onClose()}
    >
      <div className="bg-white dark:bg-slate-900 rounded-xl shadow-soft-lg w-full max-w-lg border border-slate-200 dark:border-slate-700 animate-scale-in max-h-[90vh] overflow-y-auto">
        {/* Header */}
        <div className="flex items-center justify-between px-5 py-4 border-b border-slate-200 dark:border-slate-700 sticky top-0 bg-white dark:bg-slate-900 z-10">
          <h2 className="text-base font-semibold text-slate-900 dark:text-slate-100">
            {isEdit ? `Editar flag: ${flag.key}` : 'Nova feature flag'}
          </h2>
          <button
            onClick={onClose}
            className="p-1.5 rounded-lg text-slate-400 hover:text-slate-600 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* Form */}
        <form onSubmit={isEdit ? handleEdit : handleCreate} className="px-5 py-4 space-y-5">
          {!isEdit && (
            <div>
              <label className={labelClass} htmlFor="flag-key">
                Key <span className="text-red-500">*</span>
              </label>
              <input
                id="flag-key"
                type="text"
                placeholder="ex: new-checkout-flow"
                className={inputClass}
                {...createForm.register('key')}
              />
              {createForm.formState.errors.key && (
                <p className={errorClass}>{createForm.formState.errors.key.message}</p>
              )}
            </div>
          )}

          <div>
            <label className={labelClass} htmlFor="flag-desc">
              Descricao
            </label>
            <textarea
              id="flag-desc"
              rows={2}
              placeholder="Descreva o proposito desta flag..."
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

          {/* Toggle ativo */}
          <div className="flex items-center gap-3">
            <div className="relative inline-flex items-center">
              <input
                id="flag-enabled"
                type="checkbox"
                className="sr-only peer"
                {...(isEdit ? editForm.register('enabled') : createForm.register('enabled'))}
              />
              <label
                htmlFor="flag-enabled"
                className="w-10 h-6 bg-slate-300 peer-checked:bg-violet-600 rounded-full cursor-pointer transition-colors peer-focus:ring-2 peer-focus:ring-violet-500 peer-focus:ring-offset-2 dark:peer-focus:ring-offset-slate-900 relative after:content-[''] after:absolute after:top-0.5 after:left-0.5 after:bg-white after:rounded-full after:w-5 after:h-5 after:transition-transform peer-checked:after:translate-x-4 dark:bg-slate-600"
              />
            </div>
            <label htmlFor="flag-enabled" className="text-sm text-slate-700 dark:text-slate-300 cursor-pointer select-none">
              Flag ativa
            </label>
          </div>

          {/* BKL-123: Rollout % */}
          <div>
            <label className={labelClass} htmlFor="flag-rollout">
              Rollout gradual
              <span className="ml-2 font-semibold text-violet-600 dark:text-violet-400">{rolloutValue ?? 100}%</span>
            </label>
            <input
              id="flag-rollout"
              type="range"
              min={0}
              max={100}
              step={5}
              className="w-full accent-violet-600 cursor-pointer"
              {...(isEdit
                ? editForm.register('rollout_percentage', { valueAsNumber: true })
                : createForm.register('rollout_percentage', { valueAsNumber: true }))}
            />
            <div className="flex justify-between text-xs text-slate-400 mt-0.5">
              <span>0%</span>
              <span className="text-slate-500 dark:text-slate-400 text-center text-[11px]">
                {rolloutValue === 100 ? 'Todos os usuarios' : rolloutValue === 0 ? 'Nenhum usuario' : `${rolloutValue}% do trafego`}
              </span>
              <span>100%</span>
            </div>
          </div>

          {/* BKL-989: Visual Rule Builder para targeting avançado */}
          <div>
            <div className="flex items-center justify-between mb-1">
              <label className={labelClass}>
                <span className="flex items-center gap-1.5">
                  <SlidersHorizontal className="w-3.5 h-3.5 text-indigo-500" />
                  Targeting por regras
                </span>
                <span className="ml-1 text-xs text-slate-400 font-normal">(opcional — acesso direto, ignora rollout%)</span>
              </label>
              <button
                type="button"
                onClick={() => setShowRuleBuilder(v => !v)}
                className="text-xs text-indigo-600 hover:text-indigo-800 font-medium"
              >
                {showRuleBuilder ? 'remover targeting' : '+ adicionar regras'}
              </button>
            </div>

            {showRuleBuilder && (
              <RuleBuilder
                groups={ruleGroups}
                onChange={setRuleGroups}
                disabled={loading}
              />
            )}
          </div>

          {/* BKL-124: Variantes A/B Testing */}
          <div>
            <label className={labelClass}>
              <span className="flex items-center gap-1.5">
                <FlaskConical className="w-3.5 h-3.5 text-violet-500" />
                Variantes A/B Testing
              </span>
              <span className="ml-1 text-xs text-slate-400 font-normal">(opcional — vazio = on/off normal)</span>
            </label>

            {variants.length > 0 && (
              <div className="mb-3 space-y-2">
                {variants.map(v => (
                  <div key={v.key} className="flex items-center gap-2">
                    <span className="flex-1 text-sm font-mono text-slate-700 dark:text-slate-300 bg-slate-100 dark:bg-slate-800 px-2 py-1 rounded-lg">
                      {v.key}
                    </span>
                    <input
                      type="number"
                      min={0}
                      max={100}
                      value={v.weight}
                      onChange={e => updateVariantWeight(v.key, Number(e.target.value))}
                      className="w-20 px-2 py-1 text-sm rounded-lg border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 text-center"
                    />
                    <span className="text-xs text-slate-400">%</span>
                    <button type="button" onClick={() => removeVariant(v.key)} className="text-slate-400 hover:text-red-500 transition-colors">
                      <Trash2 className="w-4 h-4" />
                    </button>
                  </div>
                ))}
                <div className={`text-xs font-medium mt-1 ${variantWeightSum === 100 ? 'text-green-600 dark:text-green-400' : 'text-red-500'}`}>
                  Total: {variantWeightSum}% {variantWeightSum !== 100 && '(deve somar 100%)'}
                </div>
              </div>
            )}

            <div className="flex gap-2">
              <input
                type="text"
                value={variantKeyInput}
                onChange={e => setVariantKeyInput(e.target.value)}
                onKeyDown={e => e.key === 'Enter' && (e.preventDefault(), addVariant())}
                placeholder="ex: control, challenger..."
                className={inputClass}
              />
              <input
                type="number"
                min={0}
                max={100}
                value={variantWeightInput}
                onChange={e => setVariantWeightInput(Number(e.target.value))}
                className="w-24 px-2 py-2 text-sm rounded-lg border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 text-center"
                placeholder="%"
              />
              <button
                type="button"
                onClick={addVariant}
                className="p-2 rounded-lg border border-slate-300 dark:border-slate-600 text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
              >
                <Plus className="w-4 h-4" />
              </button>
            </div>
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
              {isEdit ? 'Salvar alteracoes' : 'Criar flag'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
