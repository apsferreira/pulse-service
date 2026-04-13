// BKL-1076 + BKL-123 + BKL-124: Gestão completa de Feature Flags
import { useState, useEffect, useCallback } from 'react'
import {
  Flag, Plus, Search, ToggleLeft, ToggleRight, Pencil, Trash2,
  ChevronDown, ChevronUp, X, Save, AlertCircle, Copy, Check,
  Users, Percent, FlaskConical, Shield,
} from 'lucide-react'
import { flagsApi, type FeatureFlag, type CreateFlagPayload, type UpdateFlagPayload } from '../services/api'
import { useProject } from '../components/Layout'

// ── Helpers ───────────────────────────────────────────────────────────────────

function Badge({ children, color }: { children: React.ReactNode; color: string }) {
  return (
    <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium ${color}`}>
      {children}
    </span>
  )
}

function Spinner() {
  return (
    <div className="flex items-center justify-center py-16">
      <div className="w-7 h-7 border-2 border-violet-500 border-t-transparent rounded-full animate-spin" />
    </div>
  )
}

function EmptyState({ onNew }: { onNew: () => void }) {
  return (
    <div className="flex flex-col items-center justify-center py-20 text-center">
      <div className="w-14 h-14 rounded-2xl bg-violet-100 dark:bg-violet-900/30 flex items-center justify-center mb-4">
        <Flag className="w-7 h-7 text-violet-600 dark:text-violet-400" />
      </div>
      <h3 className="text-lg font-semibold text-slate-800 dark:text-slate-200 mb-1">
        Nenhuma flag ainda
      </h3>
      <p className="text-sm text-slate-500 dark:text-slate-400 max-w-xs mb-6">
        Crie sua primeira feature flag para controlar o lançamento de funcionalidades em tempo real.
      </p>
      <button onClick={onNew} className="btn-primary flex items-center gap-2">
        <Plus className="w-4 h-4" /> Nova Flag
      </button>
    </div>
  )
}

// ── FlagModal ─────────────────────────────────────────────────────────────────

interface FlagModalProps {
  flag?: FeatureFlag | null
  onClose: () => void
  onSaved: () => void
}

function FlagModal({ flag, onClose, onSaved }: FlagModalProps) {
  const isEdit = !!flag
  const [key, setKey] = useState(flag?.key ?? '')
  const [description, setDescription] = useState(flag?.description ?? '')
  const [enabled, setEnabled] = useState(flag?.enabled ?? false)
  const [rollout, setRollout] = useState(String(flag?.rollout_percentage ?? 100))
  const [userIds, setUserIds] = useState((flag?.targeting?.user_ids ?? []).join(', '))
  const [tenantIds, setTenantIds] = useState((flag?.targeting?.tenant_ids ?? []).join(', '))
  const [variants, setVariants] = useState<{ key: string; weight: number }[]>(
    flag?.variants?.length ? flag.variants : []
  )
  const [abMode, setAbMode] = useState((flag?.variants?.length ?? 0) > 0)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  const parseIds = (raw: string) =>
    raw
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean)

  const addVariant = () =>
    setVariants((v) => [...v, { key: `variant-${v.length + 1}`, weight: 0 }])

  const removeVariant = (i: number) =>
    setVariants((v) => v.filter((_, idx) => idx !== i))

  const updateVariant = (i: number, field: 'key' | 'weight', val: string) =>
    setVariants((v) =>
      v.map((item, idx) =>
        idx === i ? { ...item, [field]: field === 'weight' ? Number(val) : val } : item
      )
    )

  const totalWeight = variants.reduce((s, v) => s + v.weight, 0)

  const validate = () => {
    if (!key.trim()) return 'Chave é obrigatória'
    if (!/^[a-z0-9_-]+$/.test(key.trim()))
      return 'Chave deve conter apenas letras minúsculas, números, _ ou -'
    const pct = Number(rollout)
    if (isNaN(pct) || pct < 0 || pct > 100) return 'Rollout deve estar entre 0 e 100'
    if (abMode && variants.length < 2) return 'A/B Testing precisa de pelo menos 2 variantes'
    if (abMode && totalWeight !== 100) return `Pesos das variantes devem somar 100 (atual: ${totalWeight})`
    return ''
  }

  const handleSave = async () => {
    const err = validate()
    if (err) { setError(err); return }
    setSaving(true)
    setError('')
    try {
      const payload: CreateFlagPayload = {
        key: key.trim(),
        description: description.trim(),
        enabled,
        rollout_percentage: Number(rollout),
        targeting: {
          user_ids: parseIds(userIds),
          tenant_ids: parseIds(tenantIds),
        },
        variants: abMode ? variants : [],
      }
      if (isEdit) {
        const update: UpdateFlagPayload = { ...payload }
        delete (update as Record<string, unknown>).key
        await flagsApi.update(flag!.key, update)
      } else {
        await flagsApi.create(payload)
      }
      onSaved()
      onClose()
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : 'Erro ao salvar'
      setError(msg)
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm">
      <div className="w-full max-w-lg bg-white dark:bg-slate-900 rounded-2xl shadow-2xl border border-slate-200 dark:border-slate-700 max-h-[90vh] overflow-y-auto">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-slate-200 dark:border-slate-800">
          <h2 className="font-semibold text-slate-900 dark:text-slate-100">
            {isEdit ? 'Editar Flag' : 'Nova Feature Flag'}
          </h2>
          <button onClick={onClose} className="p-1.5 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors">
            <X className="w-4 h-4 text-slate-500" />
          </button>
        </div>

        {/* Body */}
        <div className="px-6 py-5 space-y-5">
          {error && (
            <div className="flex items-start gap-2 p-3 rounded-lg bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800">
              <AlertCircle className="w-4 h-4 text-red-600 dark:text-red-400 mt-0.5 flex-shrink-0" />
              <p className="text-sm text-red-700 dark:text-red-400">{error}</p>
            </div>
          )}

          {/* Key */}
          <div>
            <label className="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1.5">
              Chave <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              value={key}
              onChange={(e) => setKey(e.target.value.toLowerCase().replace(/\s/g, '-'))}
              disabled={isEdit}
              placeholder="ex: new-checkout-ui"
              className="input-base font-mono disabled:opacity-60 disabled:cursor-not-allowed"
            />
            <p className="mt-1 text-xs text-slate-400">Apenas letras minúsculas, números, _ ou -</p>
          </div>

          {/* Description */}
          <div>
            <label className="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1.5">
              Descrição
            </label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows={2}
              placeholder="Para que serve essa flag?"
              className="input-base resize-none"
            />
          </div>

          {/* Enabled + Rollout row */}
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1.5">
                Status
              </label>
              <button
                onClick={() => setEnabled((v) => !v)}
                className={`flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-medium w-full transition-colors ${
                  enabled
                    ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400'
                    : 'bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-400'
                }`}
              >
                {enabled ? <ToggleRight className="w-4 h-4" /> : <ToggleLeft className="w-4 h-4" />}
                {enabled ? 'Ativa' : 'Inativa'}
              </button>
            </div>
            <div>
              <label className="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1.5">
                <span className="flex items-center gap-1"><Percent className="w-3 h-3" /> Rollout %</span>
              </label>
              <input
                type="number"
                min={0}
                max={100}
                value={rollout}
                onChange={(e) => setRollout(e.target.value)}
                className="input-base"
              />
            </div>
          </div>

          {/* Targeting */}
          <div>
            <label className="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1.5">
              <span className="flex items-center gap-1"><Users className="w-3 h-3" /> Targeting — User IDs</span>
            </label>
            <input
              type="text"
              value={userIds}
              onChange={(e) => setUserIds(e.target.value)}
              placeholder="uuid1, uuid2, ... (separados por vírgula)"
              className="input-base font-mono text-xs"
            />
          </div>
          <div>
            <label className="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1.5">
              <span className="flex items-center gap-1"><Shield className="w-3 h-3" /> Targeting — Tenant IDs</span>
            </label>
            <input
              type="text"
              value={tenantIds}
              onChange={(e) => setTenantIds(e.target.value)}
              placeholder="tenant-1, tenant-2, ... (separados por vírgula)"
              className="input-base font-mono text-xs"
            />
          </div>

          {/* A/B Variants */}
          <div>
            <div className="flex items-center justify-between mb-2">
              <label className="text-xs font-medium text-slate-600 dark:text-slate-400 flex items-center gap-1">
                <FlaskConical className="w-3 h-3" /> A/B Testing
              </label>
              <button
                onClick={() => { setAbMode((v) => !v); if (!abMode && variants.length === 0) setVariants([{ key: 'control', weight: 50 }, { key: 'challenger', weight: 50 }]) }}
                className={`px-2.5 py-1 rounded-lg text-xs font-medium transition-colors ${
                  abMode
                    ? 'bg-violet-100 text-violet-700 dark:bg-violet-900/40 dark:text-violet-300'
                    : 'bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-400 hover:bg-slate-200'
                }`}
              >
                {abMode ? 'Ativado' : 'Ativar'}
              </button>
            </div>

            {abMode && (
              <div className="space-y-2">
                {variants.map((v, i) => (
                  <div key={i} className="flex items-center gap-2">
                    <input
                      type="text"
                      value={v.key}
                      onChange={(e) => updateVariant(i, 'key', e.target.value)}
                      placeholder="nome da variante"
                      className="input-base flex-1 font-mono text-xs"
                    />
                    <input
                      type="number"
                      min={0}
                      max={100}
                      value={v.weight}
                      onChange={(e) => updateVariant(i, 'weight', e.target.value)}
                      className="input-base w-20 text-right"
                    />
                    <span className="text-xs text-slate-400">%</span>
                    <button onClick={() => removeVariant(i)} className="p-1 rounded hover:bg-red-50 dark:hover:bg-red-900/20 text-slate-400 hover:text-red-500 transition-colors">
                      <X className="w-3.5 h-3.5" />
                    </button>
                  </div>
                ))}
                <div className="flex items-center justify-between">
                  <button onClick={addVariant} className="text-xs text-violet-600 dark:text-violet-400 hover:underline flex items-center gap-1">
                    <Plus className="w-3 h-3" /> Adicionar variante
                  </button>
                  <span className={`text-xs font-mono ${totalWeight === 100 ? 'text-emerald-600' : 'text-amber-600'}`}>
                    Total: {totalWeight}%
                  </span>
                </div>
              </div>
            )}
          </div>
        </div>

        {/* Footer */}
        <div className="flex items-center justify-end gap-3 px-6 py-4 border-t border-slate-200 dark:border-slate-800">
          <button onClick={onClose} className="btn-ghost">Cancelar</button>
          <button onClick={handleSave} disabled={saving} className="btn-primary flex items-center gap-2">
            {saving ? <div className="w-3.5 h-3.5 border-2 border-white/40 border-t-white rounded-full animate-spin" /> : <Save className="w-3.5 h-3.5" />}
            {isEdit ? 'Salvar alterações' : 'Criar flag'}
          </button>
        </div>
      </div>
    </div>
  )
}

// ── FlagRow ───────────────────────────────────────────────────────────────────

interface FlagRowProps {
  flag: FeatureFlag
  onToggle: (flag: FeatureFlag) => void
  onEdit: (flag: FeatureFlag) => void
  onDelete: (flag: FeatureFlag) => void
}

function FlagRow({ flag, onToggle, onEdit, onDelete }: FlagRowProps) {
  const [expanded, setExpanded] = useState(false)
  const [copied, setCopied] = useState(false)

  const copyKey = async () => {
    await navigator.clipboard.writeText(flag.key)
    setCopied(true)
    setTimeout(() => setCopied(false), 1500)
  }

  const hasTargeting =
    (flag.targeting?.user_ids?.length ?? 0) > 0 ||
    (flag.targeting?.tenant_ids?.length ?? 0) > 0

  return (
    <div className={`rounded-xl border transition-all ${
      flag.enabled
        ? 'border-emerald-200 dark:border-emerald-800/50 bg-white dark:bg-slate-900'
        : 'border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-900/50'
    }`}>
      <div className="flex items-center gap-3 px-4 py-3">
        {/* Toggle */}
        <button
          onClick={() => onToggle(flag)}
          className={`flex-shrink-0 transition-colors ${
            flag.enabled
              ? 'text-emerald-500 hover:text-emerald-600'
              : 'text-slate-300 dark:text-slate-600 hover:text-slate-400'
          }`}
          title={flag.enabled ? 'Desativar' : 'Ativar'}
        >
          {flag.enabled
            ? <ToggleRight className="w-7 h-7" />
            : <ToggleLeft className="w-7 h-7" />}
        </button>

        {/* Key + description */}
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 flex-wrap">
            <button
              onClick={copyKey}
              className="flex items-center gap-1.5 group"
              title="Copiar chave"
            >
              <code className="text-sm font-mono font-medium text-slate-800 dark:text-slate-200">
                {flag.key}
              </code>
              {copied
                ? <Check className="w-3.5 h-3.5 text-emerald-500" />
                : <Copy className="w-3.5 h-3.5 text-slate-300 dark:text-slate-600 group-hover:text-slate-500 transition-colors" />}
            </button>

            {/* Badges */}
            {flag.rollout_percentage < 100 && (
              <Badge color="bg-amber-50 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400">
                <Percent className="w-3 h-3" />
                {flag.rollout_percentage}%
              </Badge>
            )}
            {hasTargeting && (
              <Badge color="bg-blue-50 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400">
                <Users className="w-3 h-3" />
                Targeting
              </Badge>
            )}
            {(flag.variants?.length ?? 0) > 0 && (
              <Badge color="bg-violet-50 text-violet-700 dark:bg-violet-900/30 dark:text-violet-400">
                <FlaskConical className="w-3 h-3" />
                A/B ({flag.variants.length} variantes)
              </Badge>
            )}
            {!flag.enabled && (
              <Badge color="bg-slate-100 text-slate-500 dark:bg-slate-800 dark:text-slate-400">
                Inativa
              </Badge>
            )}
          </div>
          {flag.description && (
            <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5 truncate">{flag.description}</p>
          )}
        </div>

        {/* Actions */}
        <div className="flex items-center gap-1 flex-shrink-0">
          <button
            onClick={() => setExpanded((v) => !v)}
            className="p-1.5 rounded-lg text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
            title="Detalhes"
          >
            {expanded ? <ChevronUp className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />}
          </button>
          <button
            onClick={() => onEdit(flag)}
            className="p-1.5 rounded-lg text-slate-400 hover:bg-blue-50 hover:text-blue-600 dark:hover:bg-blue-900/20 transition-colors"
            title="Editar"
          >
            <Pencil className="w-4 h-4" />
          </button>
          <button
            onClick={() => onDelete(flag)}
            className="p-1.5 rounded-lg text-slate-400 hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20 transition-colors"
            title="Excluir"
          >
            <Trash2 className="w-4 h-4" />
          </button>
        </div>
      </div>

      {/* Expanded detail */}
      {expanded && (
        <div className="px-4 pb-4 border-t border-slate-100 dark:border-slate-800 pt-3 space-y-3">
          <div className="grid grid-cols-2 gap-3 text-xs">
            <div>
              <span className="text-slate-400">Rollout</span>
              <div className="mt-1">
                <div className="h-1.5 rounded-full bg-slate-100 dark:bg-slate-800 overflow-hidden">
                  <div
                    className="h-full rounded-full bg-violet-500 transition-all"
                    style={{ width: `${flag.rollout_percentage}%` }}
                  />
                </div>
                <span className="text-slate-600 dark:text-slate-400 font-mono">{flag.rollout_percentage}%</span>
              </div>
            </div>
            <div>
              <span className="text-slate-400">Criado em</span>
              <p className="text-slate-700 dark:text-slate-300 font-mono mt-1">
                {new Date(flag.created_at).toLocaleDateString('pt-BR')}
              </p>
            </div>
          </div>

          {hasTargeting && (
            <div className="text-xs space-y-1">
              {(flag.targeting?.user_ids?.length ?? 0) > 0 && (
                <div>
                  <span className="text-slate-400">User IDs:</span>
                  <p className="font-mono text-slate-600 dark:text-slate-400 break-all">
                    {flag.targeting.user_ids!.join(', ')}
                  </p>
                </div>
              )}
              {(flag.targeting?.tenant_ids?.length ?? 0) > 0 && (
                <div>
                  <span className="text-slate-400">Tenant IDs:</span>
                  <p className="font-mono text-slate-600 dark:text-slate-400 break-all">
                    {flag.targeting.tenant_ids!.join(', ')}
                  </p>
                </div>
              )}
            </div>
          )}

          {(flag.variants?.length ?? 0) > 0 && (
            <div className="text-xs">
              <span className="text-slate-400 block mb-1">Variantes A/B:</span>
              <div className="space-y-1.5">
                {flag.variants.map((v) => (
                  <div key={v.key} className="flex items-center gap-2">
                    <code className="text-slate-700 dark:text-slate-300 w-32 truncate">{v.key}</code>
                    <div className="flex-1 h-1.5 rounded-full bg-slate-100 dark:bg-slate-800 overflow-hidden">
                      <div
                        className="h-full rounded-full bg-violet-500"
                        style={{ width: `${v.weight}%` }}
                      />
                    </div>
                    <span className="text-slate-500 font-mono w-8 text-right">{v.weight}%</span>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* SDK snippet */}
          <div className="bg-slate-900 dark:bg-slate-950 rounded-lg p-3">
            <p className="text-xs text-slate-500 mb-1.5">SDK — JavaScript</p>
            <pre className="text-xs text-emerald-400 font-mono overflow-x-auto">{`const enabled = await pulse.isEnabled('${flag.key}', context)${(flag.variants?.length ?? 0) > 0 ? `\nconst variant = await pulse.getVariant('${flag.key}', context)` : ''}`}</pre>
          </div>
        </div>
      )}
    </div>
  )
}

// ── DeleteConfirm ─────────────────────────────────────────────────────────────

function DeleteConfirm({ flag, onConfirm, onCancel }: { flag: FeatureFlag; onConfirm: () => void; onCancel: () => void }) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm">
      <div className="w-full max-w-sm bg-white dark:bg-slate-900 rounded-2xl shadow-2xl border border-slate-200 dark:border-slate-700 p-6">
        <div className="flex items-start gap-3 mb-4">
          <div className="w-9 h-9 rounded-xl bg-red-100 dark:bg-red-900/30 flex items-center justify-center flex-shrink-0">
            <Trash2 className="w-4 h-4 text-red-600 dark:text-red-400" />
          </div>
          <div>
            <h3 className="font-semibold text-slate-900 dark:text-slate-100">Excluir flag</h3>
            <p className="text-sm text-slate-500 dark:text-slate-400 mt-1">
              Tem certeza que deseja excluir <code className="font-mono text-red-600 dark:text-red-400">{flag.key}</code>? Esta ação não pode ser desfeita.
            </p>
          </div>
        </div>
        <div className="flex gap-3 justify-end">
          <button onClick={onCancel} className="btn-ghost">Cancelar</button>
          <button onClick={onConfirm} className="px-4 py-2 rounded-lg text-sm font-medium bg-red-600 text-white hover:bg-red-700 transition-colors">
            Excluir
          </button>
        </div>
      </div>
    </div>
  )
}

// ── Main Page ─────────────────────────────────────────────────────────────────

export default function FeatureFlags() {
  const { projectId } = useProject()
  const [flags, setFlags] = useState<FeatureFlag[]>([])
  const [loading, setLoading] = useState(true)
  const [search, setSearch] = useState('')
  const [filterStatus, setFilterStatus] = useState<'all' | 'active' | 'inactive'>('all')
  const [modalFlag, setModalFlag] = useState<FeatureFlag | null | undefined>(undefined) // undefined = fechado, null = novo
  const [deleteTarget, setDeleteTarget] = useState<FeatureFlag | null>(null)
  const [error, setError] = useState('')

  const fetchFlags = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const { data } = await flagsApi.list()
      setFlags(Array.isArray(data) ? data : [])
    } catch {
      setError('Não foi possível carregar as flags. Verifique a conexão com a API.')
    } finally {
      setLoading(false)
    }
  }, [projectId])

  useEffect(() => { void fetchFlags() }, [fetchFlags])

  const handleToggle = async (flag: FeatureFlag) => {
    try {
      await flagsApi.update(flag.key, { enabled: !flag.enabled })
      setFlags((prev) =>
        prev.map((f) => (f.key === flag.key ? { ...f, enabled: !f.enabled } : f))
      )
    } catch {
      setError('Erro ao atualizar flag.')
    }
  }

  const handleDelete = async () => {
    if (!deleteTarget) return
    try {
      await flagsApi.remove(deleteTarget.key)
      setFlags((prev) => prev.filter((f) => f.key !== deleteTarget.key))
    } catch {
      setError('Erro ao excluir flag.')
    } finally {
      setDeleteTarget(null)
    }
  }

  const filtered = flags.filter((f) => {
    const matchSearch =
      f.key.toLowerCase().includes(search.toLowerCase()) ||
      f.description?.toLowerCase().includes(search.toLowerCase())
    const matchStatus =
      filterStatus === 'all' ||
      (filterStatus === 'active' && f.enabled) ||
      (filterStatus === 'inactive' && !f.enabled)
    return matchSearch && matchStatus
  })

  const activeCount = flags.filter((f) => f.enabled).length
  const abCount = flags.filter((f) => (f.variants?.length ?? 0) > 0).length

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-900 dark:text-slate-100 flex items-center gap-2">
            <Flag className="w-6 h-6 text-violet-500" />
            Feature Flags
          </h1>
          <p className="text-sm text-slate-500 dark:text-slate-400 mt-0.5">
            Controle o lançamento de funcionalidades em tempo real, sem redeploy.
          </p>
        </div>
        <button
          onClick={() => setModalFlag(null)}
          className="btn-primary flex items-center gap-2"
        >
          <Plus className="w-4 h-4" /> Nova Flag
        </button>
      </div>

      {/* Stats */}
      {flags.length > 0 && (
        <div className="grid grid-cols-3 gap-4">
          <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 px-4 py-3">
            <p className="text-xs text-slate-500 dark:text-slate-400">Total</p>
            <p className="text-2xl font-bold text-slate-900 dark:text-slate-100 mt-0.5">{flags.length}</p>
          </div>
          <div className="bg-white dark:bg-slate-900 rounded-xl border border-emerald-200 dark:border-emerald-800/40 px-4 py-3">
            <p className="text-xs text-slate-500 dark:text-slate-400">Ativas</p>
            <p className="text-2xl font-bold text-emerald-600 dark:text-emerald-400 mt-0.5">{activeCount}</p>
          </div>
          <div className="bg-white dark:bg-slate-900 rounded-xl border border-violet-200 dark:border-violet-800/40 px-4 py-3">
            <p className="text-xs text-slate-500 dark:text-slate-400">A/B Tests</p>
            <p className="text-2xl font-bold text-violet-600 dark:text-violet-400 mt-0.5">{abCount}</p>
          </div>
        </div>
      )}

      {/* Filters */}
      {flags.length > 0 && (
        <div className="flex items-center gap-3 flex-wrap">
          <div className="relative flex-1 min-w-48">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
            <input
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Buscar por chave ou descrição..."
              className="input-base pl-9"
            />
          </div>
          <div className="flex rounded-lg border border-slate-200 dark:border-slate-700 overflow-hidden text-sm">
            {(['all', 'active', 'inactive'] as const).map((v) => (
              <button
                key={v}
                onClick={() => setFilterStatus(v)}
                className={`px-3 py-1.5 font-medium transition-colors ${
                  filterStatus === v
                    ? 'bg-violet-600 text-white'
                    : 'bg-white dark:bg-slate-900 text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-800'
                }`}
              >
                {{ all: 'Todas', active: 'Ativas', inactive: 'Inativas' }[v]}
              </button>
            ))}
          </div>
        </div>
      )}

      {/* Error */}
      {error && (
        <div className="flex items-center gap-2 p-3 rounded-lg bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-sm text-red-700 dark:text-red-400">
          <AlertCircle className="w-4 h-4 flex-shrink-0" />
          {error}
          <button onClick={() => setError('')} className="ml-auto"><X className="w-4 h-4" /></button>
        </div>
      )}

      {/* Content */}
      {loading ? (
        <Spinner />
      ) : flags.length === 0 ? (
        <EmptyState onNew={() => setModalFlag(null)} />
      ) : filtered.length === 0 ? (
        <div className="text-center py-12 text-slate-500 dark:text-slate-400">
          <p>Nenhuma flag encontrada para "{search}"</p>
        </div>
      ) : (
        <div className="space-y-2">
          {filtered.map((flag) => (
            <FlagRow
              key={flag.key}
              flag={flag}
              onToggle={handleToggle}
              onEdit={(f) => setModalFlag(f)}
              onDelete={(f) => setDeleteTarget(f)}
            />
          ))}
          <p className="text-xs text-slate-400 dark:text-slate-500 text-right pt-1">
            {filtered.length} de {flags.length} flags
          </p>
        </div>
      )}

      {/* Modals */}
      {modalFlag !== undefined && (
        <FlagModal
          flag={modalFlag}
          onClose={() => setModalFlag(undefined)}
          onSaved={fetchFlags}
        />
      )}
      {deleteTarget && (
        <DeleteConfirm
          flag={deleteTarget}
          onConfirm={handleDelete}
          onCancel={() => setDeleteTarget(null)}
        />
      )}
    </div>
  )
}
