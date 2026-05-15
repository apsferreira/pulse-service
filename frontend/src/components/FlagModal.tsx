// BKL-123: Formulário de feature flag com rollout % e targeting por user_id/tenant_id
// BKL-124: Suporte a variantes A/B testing
// BKL-989: Visual rule builder para targeting avançado (substitui campos user_ids/tenant_ids)

import { useEffect, useRef, useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Check, ClipboardCopy, X, Plus, Trash2, FlaskConical, SlidersHorizontal, TestTube2 } from 'lucide-react'
import type { FeatureFlag, CreateFlagPayload, UpdateFlagPayload, FlagVariant, FlagPreviewResponse } from '../services/api'
import { flagsApi } from '../services/api'
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

  // Preview de avaliação — "em qual grupo esse user_id cai?"
  const [previewUserId, setPreviewUserId] = useState('')
  const [previewResult, setPreviewResult] = useState<FlagPreviewResponse | null>(null)
  const [previewLoading, setPreviewLoading] = useState(false)
  const [previewCopied, setPreviewCopied] = useState(false)
  const previewDebounceRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const runPreview = (userId: string) => {
    if (!flag?.key || !userId.trim()) { setPreviewResult(null); return }
    setPreviewLoading(true)
    flagsApi.preview(flag.key, userId.trim())
      .then(r => setPreviewResult(r.data))
      .catch(() => setPreviewResult(null))
      .finally(() => setPreviewLoading(false))
  }

  const handlePreviewInput = (userId: string) => {
    setPreviewUserId(userId)
    if (previewDebounceRef.current) clearTimeout(previewDebounceRef.current)
    previewDebounceRef.current = setTimeout(() => runPreview(userId), 400)
  }

  const copyPreviewUrl = () => {
    if (!previewResult?.preview_url) return
    navigator.clipboard.writeText(previewResult.preview_url).then(() => {
      setPreviewCopied(true)
      setTimeout(() => setPreviewCopied(false), 2000)
    })
  }

  const reasonLabel: Record<string, string> = {
    disabled:         'Flag desabilitada — sempre OFF',
    always_on:        'Sem rollout/targeting — sempre ON para todos',
    targeting_user:   'Forçado ON por targeting (user_id na lista)',
    targeting_tenant: 'Forçado ON por targeting (tenant_id na lista)',
    rollout:          'Dentro do rollout',
    rollout_miss:     'Fora do rollout',
    targeting_miss:   'Não está na lista de targeting',
  }

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
  const buildTargeting = (): import('../services/api').FlagTargeting | undefined => {
    if (!showRuleBuilder) return undefined
    const backend = ruleGroupsToBackend(ruleGroups) as { groups?: unknown[] }
    if (!backend.groups || backend.groups.length === 0) return undefined
    return backend as import('../services/api').FlagTargeting
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

          {/* Testar avaliação — só no modo edição */}
          {isEdit && (
            <div className="rounded-xl border border-slate-200 dark:border-slate-700 overflow-hidden">
              <div className="flex items-center gap-2 px-3 py-2 bg-slate-50 dark:bg-slate-800/60 border-b border-slate-200 dark:border-slate-700">
                <TestTube2 className="w-3.5 h-3.5 text-violet-500" />
                <span className="text-xs font-semibold text-slate-700 dark:text-slate-300">Testar avaliação</span>
                <span className="text-xs text-slate-400 font-normal">— em qual grupo esse usuário cai?</span>
              </div>
              <div className="px-3 py-3 space-y-2">
                <input
                  type="text"
                  value={previewUserId}
                  onChange={e => handlePreviewInput(e.target.value)}
                  placeholder="user_id para simular (ex: user-123, email@...)"
                  className="w-full px-3 py-2 text-sm rounded-lg border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-violet-500/40"
                />
                {previewLoading && (
                  <div className="flex items-center gap-2 text-xs text-slate-400">
                    <div className="w-3 h-3 border-2 border-slate-300 border-t-violet-500 rounded-full animate-spin" />
                    Calculando...
                  </div>
                )}
                {!previewLoading && previewResult && (
                  <div className="space-y-2">
                    {/* Resultado principal */}
                    <div className={`flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-semibold ${
                      previewResult.enabled
                        ? 'bg-green-50 dark:bg-green-900/20 text-green-700 dark:text-green-400 border border-green-200 dark:border-green-800'
                        : 'bg-red-50 dark:bg-red-900/20 text-red-700 dark:text-red-400 border border-red-200 dark:border-red-800'
                    }`}>
                      <span className={`w-2 h-2 rounded-full ${previewResult.enabled ? 'bg-green-500' : 'bg-red-500'}`} />
                      {previewResult.enabled ? 'ON — flag ativa para esse usuário' : 'OFF — flag inativa para esse usuário'}
                      {previewResult.variant && (
                        <span className="ml-auto font-mono text-xs px-2 py-0.5 rounded-md bg-violet-100 dark:bg-violet-900/40 text-violet-700 dark:text-violet-300">
                          variante: {previewResult.variant}
                        </span>
                      )}
                    </div>

                    {/* Detalhes */}
                    <div className="text-xs text-slate-500 dark:text-slate-400 space-y-1 px-1">
                      <div>{reasonLabel[previewResult.reason] ?? previewResult.reason}</div>
                      {previewResult.bucket !== null && (
                        <div>
                          Bucket: <span className="font-mono font-semibold text-slate-700 dark:text-slate-300">{previewResult.bucket}</span>
                          <span className="text-slate-400"> / 99</span>
                          {' '}— rollout: <span className="font-semibold">{previewResult.rollout_percentage}%</span>
                          {' '}(limite: {previewResult.rollout_percentage - 1})
                        </div>
                      )}
                    </div>

                    {/* Link de preview */}
                    <div className="flex items-center gap-2">
                      <code className="flex-1 text-xs font-mono bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400 px-2 py-1.5 rounded-lg overflow-x-auto whitespace-nowrap">
                        {previewResult.preview_url}
                      </code>
                      <button
                        type="button"
                        onClick={copyPreviewUrl}
                        title="Copiar URL de override"
                        className="flex-shrink-0 p-1.5 rounded-lg border border-slate-200 dark:border-slate-700 text-slate-500 hover:text-violet-600 hover:border-violet-400 transition-colors"
                      >
                        {previewCopied ? <Check className="w-3.5 h-3.5 text-green-500" /> : <ClipboardCopy className="w-3.5 h-3.5" />}
                      </button>
                    </div>
                    <p className="text-xs text-slate-400 leading-relaxed">
                      Cole esse parâmetro na URL do seu app para forçar esse grupo.
                      O SDK JS lê <code className="font-mono">pulse_override[flag_key]</code> automaticamente.
                    </p>
                  </div>
                )}
              </div>
            </div>
          )}

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
