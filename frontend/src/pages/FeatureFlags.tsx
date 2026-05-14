import { useState, useEffect, useCallback, useRef } from 'react'
import {
  Flag, Plus, Search, Pencil, Trash2,
  ChevronDown, ChevronUp, X, Save, AlertCircle, Copy, Check,
  Users, Percent, FlaskConical, Shield, RefreshCw,
} from 'lucide-react'
import { flagsApi, type FeatureFlag, type CreateFlagPayload, type UpdateFlagPayload } from '../services/api'
import { useProject } from '../components/Layout'

// ── Helpers ───────────────────────────────────────────────────────────────────

function Badge({ children, color }: { children: React.ReactNode; color: string }) {
  return (
    <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[11px] font-medium ${color}`}>
      {children}
    </span>
  )
}

function StatCardSkeleton() {
  return (
    <div className="stat-card border-slate-200 dark:border-slate-800">
      <div className="skeleton w-10 h-10 rounded-xl" />
      <div>
        <div className="skeleton h-7 w-10 mb-1" />
        <div className="skeleton h-3 w-16" />
      </div>
    </div>
  )
}

function FlagRowSkeleton() {
  return (
    <div className="relative border-b border-slate-100 dark:border-slate-800 px-4 h-[3.75rem] flex items-center gap-3">
      <div className="skeleton w-9 h-5 rounded-full" />
      <div className="flex-1 space-y-1.5">
        <div className="skeleton h-3.5 w-40 rounded" />
        <div className="skeleton h-2.5 w-64 rounded" />
      </div>
      <div className="flex gap-1">
        <div className="skeleton w-7 h-7 rounded-lg" />
        <div className="skeleton w-7 h-7 rounded-lg" />
        <div className="skeleton w-7 h-7 rounded-lg" />
      </div>
    </div>
  )
}

function EmptyState({ onNew }: { onNew: () => void }) {
  return (
    <div className="flex flex-col items-center justify-center py-20 text-center">
      <div className="w-16 h-16 rounded-2xl bg-violet-50 dark:bg-violet-900/20 flex items-center justify-center mb-5 shadow-flag-focus">
        <Flag className="w-8 h-8 text-violet-500 dark:text-violet-400" />
      </div>
      <h3 className="text-base font-semibold text-slate-800 dark:text-slate-200 mb-1">
        Nenhuma flag ainda
      </h3>
      <p className="text-sm text-slate-500 dark:text-slate-400 max-w-xs mb-6">
        Crie sua primeira feature flag para controlar o lançamento de funcionalidades em tempo real, sem redeploy.
      </p>
      <button onClick={onNew} className="btn-primary flex items-center gap-2">
        <Plus className="w-4 h-4" /> Nova Flag
      </button>
    </div>
  )
}

// ── Toggle Component ───────────────────────────────────────────────────────────

function FlagToggle({
  enabled,
  loading,
  onChange,
}: {
  enabled: boolean
  loading: boolean
  onChange: () => void
}) {
  return (
    <button
      role="switch"
      aria-checked={enabled}
      onClick={onChange}
      disabled={loading}
      title={enabled ? 'Desativar flag' : 'Ativar flag'}
      className={`flag-toggle ${enabled ? 'flag-toggle--on' : 'flag-toggle--off'} ${loading ? 'opacity-60 cursor-wait' : ''}`}
    >
      <span
        className={`flag-toggle__thumb ${
          loading
            ? 'translate-x-2 bg-white/70'
            : enabled
            ? 'flag-toggle__thumb--on'
            : 'flag-toggle__thumb--off'
        }`}
      />
    </button>
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
  const [enabled, setEnabled] = useState(flag?.enabled ?? true)
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
    raw.split(',').map((s) => s.trim()).filter(Boolean)

  const userIdCount = parseIds(userIds).length
  const tenantIdCount = parseIds(tenantIds).length

  const rolloutNum = Number(rollout)
  const showRolloutWarning = enabled && !isNaN(rolloutNum) && rolloutNum === 0

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
      const msg = e instanceof Error ? e.message : 'Erro ao salvar flag'
      setError(msg)
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm">
      <div className="w-full max-w-lg bg-white dark:bg-slate-900 rounded-2xl shadow-soft-lg border border-slate-200 dark:border-slate-700 max-h-[90vh] overflow-y-auto animate-scale-in">

        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-slate-100 dark:border-slate-800">
          <div className="flex items-center gap-2.5">
            <div className="w-7 h-7 rounded-lg bg-violet-100 dark:bg-violet-900/40 flex items-center justify-center">
              <Flag className="w-3.5 h-3.5 text-violet-600 dark:text-violet-400" />
            </div>
            <h2 className="font-semibold text-slate-900 dark:text-slate-100 text-sm">
              {isEdit ? `Editar: ${flag!.key}` : 'Nova Feature Flag'}
            </h2>
          </div>
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

          {/* ── Configuração básica ── */}
          <div className="modal-section">Configuração básica</div>

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
            <p className="input-hint">Apenas letras minúsculas, números, _ ou -. Imutável após criar.</p>
          </div>

          <div>
            <label className="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1.5">Descrição</label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows={2}
              placeholder="Para que serve essa flag? Quem é responsável?"
              className="input-base resize-none"
            />
          </div>

          {/* ── Distribuição ── */}
          <div className="modal-section">Distribuição</div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1.5">Status</label>
              <button
                onClick={() => setEnabled((v) => !v)}
                className={`flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-medium w-full transition-colors border ${
                  enabled
                    ? 'bg-emerald-50 text-emerald-700 border-emerald-200 dark:bg-emerald-900/20 dark:text-emerald-400 dark:border-emerald-800/40'
                    : 'bg-slate-100 text-slate-600 border-slate-200 dark:bg-slate-800 dark:text-slate-400 dark:border-slate-700'
                }`}
              >
                <FlagToggle enabled={enabled} loading={false} onChange={() => setEnabled((v) => !v)} />
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
              {showRolloutWarning && (
                <p className="input-hint--warning">
                  <AlertCircle className="w-3 h-3 flex-shrink-0" />
                  Flag ativa com 0% — nenhum usuário verá esta flag
                </p>
              )}
            </div>
          </div>

          <div>
            <div className="flex items-center justify-between mb-2">
              <label className="text-xs font-medium text-slate-600 dark:text-slate-400 flex items-center gap-1">
                <FlaskConical className="w-3 h-3" /> A/B Testing
              </label>
              <button
                onClick={() => {
                  setAbMode((v) => !v)
                  if (!abMode && variants.length === 0)
                    setVariants([{ key: 'control', weight: 50 }, { key: 'challenger', weight: 50 }])
                }}
                className={`px-2.5 py-1 rounded-lg text-xs font-medium transition-colors ${
                  abMode
                    ? 'bg-violet-100 text-violet-700 dark:bg-violet-900/40 dark:text-violet-300'
                    : 'bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-400 hover:bg-slate-200 dark:hover:bg-slate-700'
                }`}
              >
                {abMode ? 'Ativado' : 'Ativar'}
              </button>
            </div>
            {abMode && (
              <div className="space-y-2">
                <p className="text-xs text-slate-400 dark:text-slate-500 italic">
                  Modo A/B desativa o rollout percentual — as variantes controlam a distribuição.
                </p>
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
                    <span className="text-xs text-slate-400 w-4">%</span>
                    <button onClick={() => removeVariant(i)} className="p-1 rounded hover:bg-red-50 dark:hover:bg-red-900/20 text-slate-400 hover:text-red-500 transition-colors">
                      <X className="w-3.5 h-3.5" />
                    </button>
                  </div>
                ))}
                <div className="flex items-center justify-between pt-1">
                  <button onClick={addVariant} className="text-xs text-violet-600 dark:text-violet-400 hover:underline flex items-center gap-1">
                    <Plus className="w-3 h-3" /> Adicionar variante
                  </button>
                  <span className={`text-xs font-mono font-medium ${totalWeight === 100 ? 'text-emerald-600 dark:text-emerald-400' : 'text-amber-600 dark:text-amber-400'}`}>
                    Total: {totalWeight}%
                  </span>
                </div>
              </div>
            )}
          </div>

          {/* ── Targeting ── */}
          <div className="modal-section">Targeting</div>

          <div>
            <label className="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1.5">
              <span className="flex items-center gap-1"><Users className="w-3 h-3" /> User IDs</span>
            </label>
            <input
              type="text"
              value={userIds}
              onChange={(e) => setUserIds(e.target.value)}
              placeholder="uuid1, uuid2, ... (separados por vírgula)"
              className="input-base font-mono text-xs"
            />
            <p className="input-hint">
              {userIdCount > 0 ? `${userIdCount} ID${userIdCount > 1 ? 's' : ''} configurado${userIdCount > 1 ? 's' : ''}` : 'Sem targeting específico — aplica rollout a todos'}
            </p>
          </div>

          <div>
            <label className="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1.5">
              <span className="flex items-center gap-1"><Shield className="w-3 h-3" /> Tenant IDs</span>
            </label>
            <input
              type="text"
              value={tenantIds}
              onChange={(e) => setTenantIds(e.target.value)}
              placeholder="tenant-1, tenant-2, ... (separados por vírgula)"
              className="input-base font-mono text-xs"
            />
            <p className="input-hint">
              {tenantIdCount > 0 ? `${tenantIdCount} tenant${tenantIdCount > 1 ? 's' : ''} configurado${tenantIdCount > 1 ? 's' : ''}` : 'Sem targeting por tenant'}
            </p>
          </div>
        </div>

        {/* Footer */}
        <div className="flex items-center justify-end gap-3 px-6 py-4 border-t border-slate-100 dark:border-slate-800">
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
  toggling: boolean
  onToggle: (flag: FeatureFlag) => void
  onEdit: (flag: FeatureFlag) => void
  onDelete: (flag: FeatureFlag) => void
  flashClass: string
}

function FlagRow({ flag, toggling, onToggle, onEdit, onDelete, flashClass }: FlagRowProps) {
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

  const accentColor = toggling
    ? 'bg-violet-400'
    : flag.enabled
    ? 'bg-emerald-500'
    : 'bg-transparent'

  return (
    <div className={`${flashClass}`}>
      {/* Main row */}
      <div
        className={`relative flex items-center gap-3 px-4 h-[3.75rem] transition-colors border-b border-slate-100 dark:border-slate-800 ${
          flag.enabled
            ? 'bg-white dark:bg-slate-900 hover:bg-slate-50/50 dark:hover:bg-slate-800/30'
            : 'bg-slate-50/60 dark:bg-slate-900/40 hover:bg-slate-50 dark:hover:bg-slate-800/20'
        }`}
      >
        {/* Left accent */}
        <div className={`absolute left-0 top-2 bottom-2 w-[3px] rounded-r-sm transition-colors ${accentColor}`} />

        {/* Toggle */}
        <div className="pl-2">
          <FlagToggle enabled={flag.enabled} loading={toggling} onChange={() => onToggle(flag)} />
        </div>

        {/* Key + description */}
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 flex-wrap">
            <button
              onClick={copyKey}
              className="flex items-center gap-1.5 group"
              title="Copiar chave"
            >
              <code className={`text-sm font-mono font-semibold ${flag.enabled ? 'text-slate-800 dark:text-slate-100' : 'text-slate-500 dark:text-slate-400'}`}>
                {flag.key}
              </code>
              {copied
                ? <Check className="w-3.5 h-3.5 text-emerald-500" />
                : <Copy className="w-3.5 h-3.5 text-slate-300 dark:text-slate-600 opacity-0 group-hover:opacity-100 transition-opacity" />}
            </button>

            {/* Badges */}
            {flag.rollout_percentage < 100 && flag.enabled && (
              <Badge color="bg-amber-50 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400 border border-amber-200 dark:border-amber-800/40">
                <Percent className="w-2.5 h-2.5" />
                {flag.rollout_percentage}%
              </Badge>
            )}
            {hasTargeting && (
              <Badge color="bg-blue-50 text-blue-700 dark:bg-blue-900/20 dark:text-blue-400 border border-blue-200 dark:border-blue-800/40">
                <Users className="w-2.5 h-2.5" />
                Targeting
              </Badge>
            )}
            {(flag.variants?.length ?? 0) > 0 && (
              <Badge color="bg-violet-50 text-violet-700 dark:bg-violet-900/20 dark:text-violet-400 border border-violet-200 dark:border-violet-800/40">
                <FlaskConical className="w-2.5 h-2.5" />
                A/B
              </Badge>
            )}
          </div>
          {flag.description && (
            <p className="text-xs text-slate-400 dark:text-slate-500 mt-0.5 truncate">{flag.description}</p>
          )}
        </div>

        {/* Status pill */}
        <div className="hidden sm:flex items-center gap-1.5 flex-shrink-0 w-20">
          <span className={`w-1.5 h-1.5 rounded-full flex-shrink-0 ${flag.enabled ? 'bg-emerald-500' : 'bg-slate-300 dark:bg-slate-600'}`} />
          <span className={`text-xs font-medium ${flag.enabled ? 'text-emerald-600 dark:text-emerald-400' : 'text-slate-400 dark:text-slate-500'}`}>
            {flag.enabled ? 'Ativa' : 'Inativa'}
          </span>
        </div>

        {/* Actions */}
        <div className="flex items-center gap-0.5 flex-shrink-0">
          <button
            onClick={() => setExpanded((v) => !v)}
            className="p-1.5 rounded-lg text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
            title="Detalhes"
          >
            {expanded ? <ChevronUp className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />}
          </button>
          <button
            onClick={() => onEdit(flag)}
            className="p-1.5 rounded-lg text-slate-400 hover:bg-blue-50 hover:text-blue-600 dark:hover:bg-blue-900/20 dark:hover:text-blue-400 transition-colors"
            title="Editar"
          >
            <Pencil className="w-4 h-4" />
          </button>
          <button
            onClick={() => onDelete(flag)}
            className="p-1.5 rounded-lg text-slate-400 hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20 dark:hover:text-red-400 transition-colors"
            title="Excluir"
          >
            <Trash2 className="w-4 h-4" />
          </button>
        </div>
      </div>

      {/* Expanded detail */}
      {expanded && (
        <div className="px-8 pb-4 pt-3 bg-slate-50/80 dark:bg-slate-900/60 border-b border-slate-100 dark:border-slate-800 animate-slide-down space-y-3">
          <div className="grid grid-cols-2 gap-4 text-xs">
            <div>
              <span className="text-slate-400 dark:text-slate-500 block mb-1.5">Rollout</span>
              <div className="h-1.5 rounded-full bg-slate-200 dark:bg-slate-700 overflow-hidden mb-1">
                <div
                  className="h-full rounded-full bg-violet-500 transition-all"
                  style={{ width: `${flag.rollout_percentage}%` }}
                />
              </div>
              <span className="text-slate-600 dark:text-slate-400 font-mono">{flag.rollout_percentage}%</span>
            </div>
            <div>
              <span className="text-slate-400 dark:text-slate-500 block mb-1.5">Criado em</span>
              <p className="text-slate-700 dark:text-slate-300 font-mono">
                {new Date(flag.created_at).toLocaleDateString('pt-BR')}
              </p>
            </div>
          </div>

          {hasTargeting && (
            <div className="text-xs space-y-1.5">
              {(flag.targeting?.user_ids?.length ?? 0) > 0 && (
                <div>
                  <span className="text-slate-400 dark:text-slate-500">
                    User IDs ({flag.targeting.user_ids!.length}):
                  </span>
                  <p className="font-mono text-slate-600 dark:text-slate-400 break-all mt-0.5">
                    {flag.targeting.user_ids!.join(', ')}
                  </p>
                </div>
              )}
              {(flag.targeting?.tenant_ids?.length ?? 0) > 0 && (
                <div>
                  <span className="text-slate-400 dark:text-slate-500">
                    Tenant IDs ({flag.targeting.tenant_ids!.length}):
                  </span>
                  <p className="font-mono text-slate-600 dark:text-slate-400 break-all mt-0.5">
                    {flag.targeting.tenant_ids!.join(', ')}
                  </p>
                </div>
              )}
            </div>
          )}

          {(flag.variants?.length ?? 0) > 0 && (
            <div className="text-xs">
              <span className="text-slate-400 dark:text-slate-500 block mb-2">Variantes A/B</span>
              <div className="space-y-1.5">
                {flag.variants.map((v) => (
                  <div key={v.key} className="flex items-center gap-3">
                    <code className="text-slate-700 dark:text-slate-300 w-28 truncate">{v.key}</code>
                    <div className="flex-1 h-1.5 rounded-full bg-slate-200 dark:bg-slate-700 overflow-hidden">
                      <div className="h-full rounded-full bg-violet-500" style={{ width: `${v.weight}%` }} />
                    </div>
                    <span className="text-slate-500 dark:text-slate-400 font-mono w-8 text-right">{v.weight}%</span>
                  </div>
                ))}
              </div>
            </div>
          )}

          <div className="bg-slate-900 dark:bg-slate-950 rounded-lg p-3 border border-slate-800">
            <p className="text-[10px] text-slate-500 mb-1.5 uppercase tracking-wider">SDK — JavaScript</p>
            <pre className="text-xs text-emerald-400 font-mono overflow-x-auto whitespace-pre-wrap">{`const enabled = await pulse.isEnabled('${flag.key}', context)${(flag.variants?.length ?? 0) > 0 ? `\nconst variant = await pulse.getVariant('${flag.key}', context)` : ''}`}</pre>
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
      <div className="w-full max-w-sm bg-white dark:bg-slate-900 rounded-2xl shadow-soft-lg border border-slate-200 dark:border-slate-700 p-6 animate-scale-in">
        <div className="flex items-start gap-3 mb-5">
          <div className="w-9 h-9 rounded-xl bg-red-100 dark:bg-red-900/30 flex items-center justify-center flex-shrink-0">
            <Trash2 className="w-4 h-4 text-red-600 dark:text-red-400" />
          </div>
          <div>
            <h3 className="font-semibold text-slate-900 dark:text-slate-100 text-sm">Excluir flag</h3>
            <p className="text-sm text-slate-500 dark:text-slate-400 mt-1">
              Tem certeza que deseja excluir{' '}
              <code className="font-mono text-red-600 dark:text-red-400 text-xs">{flag.key}</code>?
              Esta ação não pode ser desfeita.
            </p>
          </div>
        </div>
        <div className="flex gap-3 justify-end">
          <button onClick={onCancel} className="btn-ghost">Cancelar</button>
          <button
            onClick={onConfirm}
            className="px-4 py-2 rounded-lg text-sm font-medium bg-red-600 text-white hover:bg-red-700 transition-colors"
          >
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
  const [refreshing, setRefreshing] = useState(false)
  const [search, setSearch] = useState('')
  const [filterStatus, setFilterStatus] = useState<'all' | 'active' | 'inactive'>('all')
  const [modalFlag, setModalFlag] = useState<FeatureFlag | null | undefined>(undefined)
  const [deleteTarget, setDeleteTarget] = useState<FeatureFlag | null>(null)
  const [error, setError] = useState('')
  // key → flash class para animação de confirmação de toggle
  const [flashMap, setFlashMap] = useState<Record<string, string>>({})
  // key → toggling state por linha (sem bloquear a lista inteira)
  const [togglingMap, setTogglingMap] = useState<Record<string, boolean>>({})
  const flashTimeouts = useRef<Record<string, ReturnType<typeof setTimeout>>>({})

  const fetchFlags = useCallback(async (isRefresh = false) => {
    if (isRefresh) setRefreshing(true)
    else setLoading(true)
    setError('')
    try {
      const { data } = await flagsApi.list()
      setFlags(Array.isArray(data) ? data : [])
    } catch {
      setError('Não foi possível carregar as flags. Verifique a conexão com a API.')
    } finally {
      setLoading(false)
      setRefreshing(false)
    }
  }, [projectId])

  useEffect(() => { void fetchFlags() }, [fetchFlags])

  const triggerFlash = (key: string, enabled: boolean) => {
    const cls = enabled ? 'animate-flag-flash-on' : 'animate-flag-flash-off'
    setFlashMap((prev) => ({ ...prev, [key]: cls }))
    if (flashTimeouts.current[key]) clearTimeout(flashTimeouts.current[key])
    flashTimeouts.current[key] = setTimeout(() => {
      setFlashMap((prev) => {
        const next = { ...prev }
        delete next[key]
        return next
      })
    }, 750)
  }

  const handleToggle = async (flag: FeatureFlag) => {
    const newEnabled = !flag.enabled
    setTogglingMap((prev) => ({ ...prev, [flag.key]: true }))
    // Otimistic update
    setFlags((prev) => prev.map((f) => (f.key === flag.key ? { ...f, enabled: newEnabled } : f)))
    try {
      await flagsApi.update(flag.key, { enabled: newEnabled })
      triggerFlash(flag.key, newEnabled)
    } catch {
      // Reverter se falhou
      setFlags((prev) => prev.map((f) => (f.key === flag.key ? { ...f, enabled: !newEnabled } : f)))
      setError(`Erro ao ${newEnabled ? 'ativar' : 'desativar'} a flag "${flag.key}". Tente novamente.`)
    } finally {
      setTogglingMap((prev) => {
        const next = { ...prev }
        delete next[flag.key]
        return next
      })
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
      (f.description ?? '').toLowerCase().includes(search.toLowerCase())
    const matchStatus =
      filterStatus === 'all' ||
      (filterStatus === 'active' && f.enabled) ||
      (filterStatus === 'inactive' && !f.enabled)
    return matchSearch && matchStatus
  })

  const activeCount = flags.filter((f) => f.enabled).length
  const abCount = flags.filter((f) => (f.variants?.length ?? 0) > 0).length
  const rolloutCount = flags.filter((f) => f.enabled && f.rollout_percentage < 100).length

  return (
    <div className="space-y-6">
      {/* Page header */}
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-xl font-bold text-slate-900 dark:text-slate-100 flex items-center gap-2.5">
            <div className="w-7 h-7 rounded-lg bg-violet-100 dark:bg-violet-900/40 flex items-center justify-center">
              <Flag className="w-4 h-4 text-violet-600 dark:text-violet-400" />
            </div>
            Feature Flags
          </h1>
          <p className="text-sm text-slate-500 dark:text-slate-400 mt-0.5 ml-9">
            Controle lançamentos em tempo real, sem redeploy.
          </p>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => void fetchFlags(true)}
            disabled={refreshing || loading}
            title="Atualizar"
            className="p-2 rounded-lg text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors disabled:opacity-40"
          >
            <RefreshCw className={`w-4 h-4 ${refreshing ? 'animate-spin' : ''}`} />
          </button>
          <button onClick={() => setModalFlag(null)} className="btn-primary flex items-center gap-2">
            <Plus className="w-4 h-4" /> Nova Flag
          </button>
        </div>
      </div>

      {/* Stat cards */}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
        {loading ? (
          <>
            <StatCardSkeleton />
            <StatCardSkeleton />
            <StatCardSkeleton />
            <StatCardSkeleton />
          </>
        ) : (
          <>
            <div className="stat-card border-slate-200 dark:border-slate-800">
              <div className="stat-card__icon bg-slate-100 dark:bg-slate-800">
                <Flag className="w-5 h-5 text-slate-500 dark:text-slate-400" />
              </div>
              <div>
                <p className="stat-card__value text-slate-900 dark:text-slate-100">{flags.length}</p>
                <p className="stat-card__label">Total</p>
              </div>
            </div>
            <div className="stat-card border-emerald-200 dark:border-emerald-800/40">
              <div className="stat-card__icon bg-emerald-50 dark:bg-emerald-900/20">
                <span className="w-2.5 h-2.5 rounded-full bg-emerald-500" />
              </div>
              <div>
                <p className="stat-card__value text-emerald-600 dark:text-emerald-400">{activeCount}</p>
                <p className="stat-card__label">Ativas</p>
              </div>
            </div>
            <div className="stat-card border-violet-200 dark:border-violet-800/40">
              <div className="stat-card__icon bg-violet-50 dark:bg-violet-900/20">
                <FlaskConical className="w-5 h-5 text-violet-500 dark:text-violet-400" />
              </div>
              <div>
                <p className="stat-card__value text-violet-600 dark:text-violet-400">{abCount}</p>
                <p className="stat-card__label">A/B Tests</p>
              </div>
            </div>
            <div className="stat-card border-amber-200 dark:border-amber-800/40">
              <div className="stat-card__icon bg-amber-50 dark:bg-amber-900/20">
                <Percent className="w-5 h-5 text-amber-500 dark:text-amber-400" />
              </div>
              <div>
                <p className="stat-card__value text-amber-600 dark:text-amber-400">{rolloutCount}</p>
                <p className="stat-card__label">Em Rollout</p>
              </div>
            </div>
          </>
        )}
      </div>

      {/* Toolbar */}
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
          {search && (
            <button
              onClick={() => setSearch('')}
              className="absolute right-2.5 top-1/2 -translate-y-1/2 p-0.5 rounded text-slate-400 hover:text-slate-600 transition-colors"
            >
              <X className="w-3.5 h-3.5" />
            </button>
          )}
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

      {/* Error */}
      {error && (
        <div className="flex items-center gap-2 p-3 rounded-lg bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-sm text-red-700 dark:text-red-400">
          <AlertCircle className="w-4 h-4 flex-shrink-0" />
          <span className="flex-1">{error}</span>
          <button onClick={() => setError('')} className="p-0.5 rounded hover:bg-red-100 dark:hover:bg-red-900/40 transition-colors">
            <X className="w-4 h-4" />
          </button>
        </div>
      )}

      {/* Content */}
      {loading ? (
        <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 overflow-hidden">
          {Array.from({ length: 5 }).map((_, i) => <FlagRowSkeleton key={i} />)}
        </div>
      ) : flags.length === 0 ? (
        <EmptyState onNew={() => setModalFlag(null)} />
      ) : filtered.length === 0 ? (
        <div className="text-center py-14">
          <Search className="w-8 h-8 text-slate-300 dark:text-slate-600 mx-auto mb-3" />
          <p className="text-sm text-slate-500 dark:text-slate-400">
            Nenhuma flag encontrada {search ? `para "${search}"` : 'com esse filtro'}
          </p>
          <button
            onClick={() => { setSearch(''); setFilterStatus('all') }}
            className="mt-3 text-xs text-violet-600 dark:text-violet-400 hover:underline"
          >
            Limpar filtros
          </button>
        </div>
      ) : (
        <div>
          <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 overflow-hidden">
            {/* Table header */}
            <div className="flex items-center gap-3 px-8 h-9 bg-slate-50 dark:bg-slate-900/80 border-b border-slate-200 dark:border-slate-800">
              <span className="flex-1 text-[11px] font-medium text-slate-400 dark:text-slate-500 uppercase tracking-wider">
                Chave / Descrição
              </span>
              <span className="hidden sm:block w-20 text-[11px] font-medium text-slate-400 dark:text-slate-500 uppercase tracking-wider">
                Status
              </span>
              <span className="w-24 text-[11px] font-medium text-slate-400 dark:text-slate-500 uppercase tracking-wider text-right">
                Ações
              </span>
            </div>

            {filtered.map((flag) => (
              <FlagRow
                key={flag.key}
                flag={flag}
                toggling={!!togglingMap[flag.key]}
                onToggle={handleToggle}
                onEdit={(f) => setModalFlag(f)}
                onDelete={(f) => setDeleteTarget(f)}
                flashClass={flashMap[flag.key] ?? ''}
              />
            ))}
          </div>

          <p className="text-xs text-slate-400 dark:text-slate-500 text-right mt-2">
            {filtered.length === flags.length
              ? `${flags.length} flag${flags.length !== 1 ? 's' : ''}`
              : `${filtered.length} de ${flags.length} flags`}
          </p>
        </div>
      )}

      {/* Modals */}
      {modalFlag !== undefined && (
        <FlagModal
          flag={modalFlag}
          onClose={() => setModalFlag(undefined)}
          onSaved={() => void fetchFlags()}
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
