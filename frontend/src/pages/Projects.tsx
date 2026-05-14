import { useState, useEffect, useCallback } from 'react'
import {
  Plus,
  RefreshCw,
  Key,
  Eye,
  EyeOff,
  RotateCcw,
  Pencil,
  CheckCircle,
  XCircle,
  ChevronRight,
  FolderOpen,
  Server,
} from 'lucide-react'
import toast from 'react-hot-toast'
import { projectsApi } from '../services/projectsApi'
import type { Project, Environment } from '../services/projectsApi'

// ── Helpers ──────────────────────────────────────────────────────────────────

const PLAN_LABELS: Record<string, string> = {
  free: 'Free',
  starter: 'Starter',
  growth: 'Growth',
  pro: 'Pro',
  enterprise: 'Enterprise',
  internal: 'Internal',
}

const PLAN_COLORS: Record<string, string> = {
  free: 'bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-300',
  starter: 'bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-300',
  growth: 'bg-violet-100 text-violet-700 dark:bg-violet-900/40 dark:text-violet-300',
  pro: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300',
  enterprise: 'bg-amber-100 text-amber-700 dark:bg-amber-900/40 dark:text-amber-300',
  internal: 'bg-rose-100 text-rose-700 dark:bg-rose-900/40 dark:text-rose-300',
}

function planBadge(plan: string) {
  return (
    <span
      className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${
        PLAN_COLORS[plan] ?? PLAN_COLORS.free
      }`}
    >
      {PLAN_LABELS[plan] ?? plan}
    </span>
  )
}

function maskKey(key: string) {
  if (key.length <= 8) return '••••••••'
  return key.slice(0, 6) + '••••••••••••' + key.slice(-4)
}

// ── Create Project Modal ──────────────────────────────────────────────────────

interface CreateProjectModalProps {
  onClose: () => void
  onSuccess: (p: Project) => void
}

function CreateProjectModal({ onClose, onSuccess }: CreateProjectModalProps) {
  const [name, setName] = useState('')
  const [slug, setSlug] = useState('')
  const [plan, setPlan] = useState('free')
  const [loading, setLoading] = useState(false)

  // Auto-generate slug from name
  const handleNameChange = (v: string) => {
    setName(v)
    setSlug(
      v
        .toLowerCase()
        .replace(/\s+/g, '-')
        .replace(/[^a-z0-9-]/g, '')
        .slice(0, 40),
    )
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!name.trim() || !slug.trim()) return
    setLoading(true)
    try {
      const project = await projectsApi.create({ name: name.trim(), slug: slug.trim(), plan })
      toast.success(`Projeto "${project.name}" criado`)
      onSuccess(project)
    } catch (err) {
      const e = err as { response?: { data?: { error?: string } }; message?: string }
      toast.error(e?.response?.data?.error ?? 'Falha ao criar projeto')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40 backdrop-blur-sm"
      onClick={(e) => e.target === e.currentTarget && onClose()}
    >
      <div className="w-full max-w-md rounded-2xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-700 shadow-xl">
        <div className="flex items-center justify-between px-6 py-4 border-b border-slate-200 dark:border-slate-700">
          <h2 className="text-base font-semibold text-slate-900 dark:text-slate-100">
            Novo projeto
          </h2>
          <button
            onClick={onClose}
            className="p-1.5 rounded-lg text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
            aria-label="Fechar"
          >
            <XCircle className="w-4 h-4" />
          </button>
        </div>

        <form onSubmit={(e) => void handleSubmit(e)} className="px-6 py-5 space-y-4">
          <div>
            <label className="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">
              Nome
            </label>
            <input
              type="text"
              value={name}
              onChange={(e) => handleNameChange(e.target.value)}
              placeholder="Meu Produto"
              required
              className="w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent placeholder:text-slate-400 dark:placeholder:text-slate-500"
            />
          </div>

          <div>
            <label className="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">
              Slug
            </label>
            <input
              type="text"
              value={slug}
              onChange={(e) =>
                setSlug(e.target.value.toLowerCase().replace(/[^a-z0-9-]/g, ''))
              }
              placeholder="meu-produto"
              required
              pattern="^[a-z0-9-]+$"
              className="w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent placeholder:text-slate-400 dark:placeholder:text-slate-500"
            />
          </div>

          <div>
            <label className="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">
              Plano
            </label>
            <select
              value={plan}
              onChange={(e) => setPlan(e.target.value)}
              className="w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent"
            >
              {Object.entries(PLAN_LABELS).map(([k, v]) => (
                <option key={k} value={k}>
                  {v}
                </option>
              ))}
            </select>
          </div>

          <div className="flex justify-end gap-2 pt-1">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 rounded-lg text-sm font-medium text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
            >
              Cancelar
            </button>
            <button
              type="submit"
              disabled={loading}
              className="flex items-center gap-2 px-4 py-2 rounded-lg bg-emerald-600 hover:bg-emerald-700 text-white text-sm font-medium transition-colors disabled:opacity-60"
            >
              {loading && <RefreshCw className="w-3.5 h-3.5 animate-spin" />}
              Criar projeto
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

// ── Edit Project Modal ────────────────────────────────────────────────────────

interface ProjectDetailModalProps {
  project: Project
  onClose: () => void
  onUpdated: (p: Project) => void
}

function ProjectDetailModal({ project, onClose, onUpdated }: ProjectDetailModalProps) {
  const [name, setName] = useState(project.name)
  const [plan, setPlan] = useState(project.plan)
  const [saving, setSaving] = useState(false)
  const [editMode, setEditMode] = useState(false)

  const [keyVisible, setKeyVisible] = useState(false)
  const [apiKey, setApiKey] = useState(project.api_key)
  const [regenerating, setRegenerating] = useState(false)

  const [environments, setEnvironments] = useState<Environment[]>([])
  const [envsLoading, setEnvsLoading] = useState(true)
  const [envsError, setEnvsError] = useState<string | null>(null)

  const [newEnvName, setNewEnvName] = useState('')
  const [newEnvSlug, setNewEnvSlug] = useState('')
  const [creatingEnv, setCreatingEnv] = useState(false)
  const [showEnvForm, setShowEnvForm] = useState(false)

  const [envKeyVisible, setEnvKeyVisible] = useState<Record<string, boolean>>({})
  const [envKeys, setEnvKeys] = useState<Record<string, string>>({})
  const [envRegenerating, setEnvRegenerating] = useState<Record<string, boolean>>({})

  const loadEnvironments = useCallback(async () => {
    setEnvsLoading(true)
    setEnvsError(null)
    try {
      const data = await projectsApi.listEnvironments(project.slug)
      setEnvironments(data.environments ?? [])
      const keys: Record<string, string> = {}
      ;(data.environments ?? []).forEach((env) => {
        keys[env.id] = env.api_key
      })
      setEnvKeys(keys)
    } catch (err) {
      const e = err as { response?: { data?: { error?: string } }; message?: string }
      setEnvsError(e?.response?.data?.error ?? 'Falha ao carregar ambientes')
    } finally {
      setEnvsLoading(false)
    }
  }, [project.slug])

  useEffect(() => {
    void loadEnvironments()
  }, [loadEnvironments])

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault()
    setSaving(true)
    try {
      const updated = await projectsApi.update(project.slug, {
        name: name.trim(),
        plan,
      })
      toast.success('Projeto atualizado')
      setEditMode(false)
      onUpdated(updated)
    } catch (err) {
      const e = err as { response?: { data?: { error?: string } }; message?: string }
      toast.error(e?.response?.data?.error ?? 'Falha ao atualizar')
    } finally {
      setSaving(false)
    }
  }

  const handleRegenerateKey = async () => {
    if (!confirm('Regenerar a API key vai invalidar a key atual. Continuar?')) return
    setRegenerating(true)
    try {
      const { api_key } = await projectsApi.regenerateKey(project.slug)
      setApiKey(api_key)
      setKeyVisible(true)
      toast.success('API key regenerada')
    } catch {
      toast.error('Falha ao regenerar key')
    } finally {
      setRegenerating(false)
    }
  }

  const handleCreateEnv = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!newEnvName.trim() || !newEnvSlug.trim()) return
    setCreatingEnv(true)
    try {
      const env = await projectsApi.createEnvironment(project.slug, {
        name: newEnvName.trim(),
        slug: newEnvSlug.trim(),
      })
      setEnvironments((prev) => [...prev, env])
      setEnvKeys((prev) => ({ ...prev, [env.id]: env.api_key }))
      setNewEnvName('')
      setNewEnvSlug('')
      setShowEnvForm(false)
      toast.success(`Ambiente "${env.name}" criado`)
    } catch (err) {
      const e = err as { response?: { data?: { error?: string } }; message?: string }
      toast.error(e?.response?.data?.error ?? 'Falha ao criar ambiente')
    } finally {
      setCreatingEnv(false)
    }
  }

  const handleRegenerateEnvKey = async (envId: string) => {
    if (!confirm('Regenerar a key deste ambiente vai invalidar a atual. Continuar?')) return
    setEnvRegenerating((prev) => ({ ...prev, [envId]: true }))
    try {
      const { api_key } = await projectsApi.regenerateEnvKey(project.slug, envId)
      setEnvKeys((prev) => ({ ...prev, [envId]: api_key }))
      setEnvKeyVisible((prev) => ({ ...prev, [envId]: true }))
      toast.success('Key do ambiente regenerada')
    } catch {
      toast.error('Falha ao regenerar key')
    } finally {
      setEnvRegenerating((prev) => ({ ...prev, [envId]: false }))
    }
  }

  const handleEnvNameChange = (v: string) => {
    setNewEnvName(v)
    setNewEnvSlug(
      v
        .toLowerCase()
        .replace(/\s+/g, '-')
        .replace(/[^a-z0-9-]/g, '')
        .slice(0, 40),
    )
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40 backdrop-blur-sm"
      onClick={(e) => e.target === e.currentTarget && onClose()}
    >
      <div className="w-full max-w-lg rounded-2xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-700 shadow-xl max-h-[90vh] flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-slate-200 dark:border-slate-700 flex-shrink-0">
          <div className="flex items-center gap-2">
            <FolderOpen className="w-4 h-4 text-emerald-500" aria-hidden="true" />
            <h2 className="text-base font-semibold text-slate-900 dark:text-slate-100 truncate max-w-[260px]">
              {project.name}
            </h2>
            {planBadge(project.plan)}
          </div>
          <button
            onClick={onClose}
            className="p-1.5 rounded-lg text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
            aria-label="Fechar"
          >
            <XCircle className="w-4 h-4" />
          </button>
        </div>

        <div className="overflow-y-auto flex-1 px-6 py-5 space-y-6">
          {/* Info section */}
          <section>
            <div className="flex items-center justify-between mb-3">
              <h3 className="text-xs font-semibold uppercase tracking-wider text-slate-400">
                Configurações
              </h3>
              {!editMode && (
                <button
                  onClick={() => setEditMode(true)}
                  className="flex items-center gap-1 text-xs text-emerald-600 dark:text-emerald-400 hover:underline"
                >
                  <Pencil className="w-3 h-3" />
                  Editar
                </button>
              )}
            </div>

            {editMode ? (
              <form onSubmit={(e) => void handleSave(e)} className="space-y-3">
                <div>
                  <label className="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">
                    Nome
                  </label>
                  <input
                    type="text"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    required
                    className="w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent"
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">
                    Slug
                  </label>
                  <input
                    type="text"
                    value={project.slug}
                    disabled
                    className="w-full px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800/50 text-slate-400 dark:text-slate-500 text-sm font-mono cursor-not-allowed"
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">
                    Plano
                  </label>
                  <select
                    value={plan}
                    onChange={(e) => setPlan(e.target.value)}
                    className="w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent"
                  >
                    {Object.entries(PLAN_LABELS).map(([k, v]) => (
                      <option key={k} value={k}>
                        {v}
                      </option>
                    ))}
                  </select>
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-xs text-slate-500 dark:text-slate-400">
                    Status:{' '}
                    {project.is_active ? (
                      <span className="text-emerald-600 font-medium">Ativo</span>
                    ) : (
                      <span className="text-amber-500 font-medium">Aguardando SDK</span>
                    )}
                  </span>
                  <span className="text-xs text-slate-400">(definido automaticamente pelo SDK)</span>
                </div>
                <div className="flex justify-end gap-2 pt-1">
                  <button
                    type="button"
                    onClick={() => {
                      setEditMode(false)
                      setName(project.name)
                      setPlan(project.plan)
                    }}
                    className="px-3 py-1.5 rounded-lg text-sm text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
                  >
                    Cancelar
                  </button>
                  <button
                    type="submit"
                    disabled={saving}
                    className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-emerald-600 hover:bg-emerald-700 text-white text-sm font-medium transition-colors disabled:opacity-60"
                  >
                    {saving && <RefreshCw className="w-3 h-3 animate-spin" />}
                    Salvar
                  </button>
                </div>
              </form>
            ) : (
              <dl className="grid grid-cols-2 gap-x-4 gap-y-3 text-sm">
                <div>
                  <dt className="text-xs text-slate-400 dark:text-slate-500">Nome</dt>
                  <dd className="font-medium text-slate-800 dark:text-slate-200 mt-0.5">
                    {project.name}
                  </dd>
                </div>
                <div>
                  <dt className="text-xs text-slate-400 dark:text-slate-500">Slug</dt>
                  <dd className="font-mono text-slate-800 dark:text-slate-200 mt-0.5">
                    {project.slug}
                  </dd>
                </div>
                <div>
                  <dt className="text-xs text-slate-400 dark:text-slate-500">Plano</dt>
                  <dd className="mt-0.5">{planBadge(project.plan)}</dd>
                </div>
                <div>
                  <dt className="text-xs text-slate-400 dark:text-slate-500">Status</dt>
                  <dd className="flex items-center gap-1 mt-0.5">
                    {project.is_active ? (
                      <>
                        <CheckCircle className="w-3.5 h-3.5 text-emerald-500" aria-hidden="true" />
                        <span className="text-emerald-600 dark:text-emerald-400 text-xs font-medium">
                          Ativo
                        </span>
                      </>
                    ) : (
                      <>
                        <Server className="w-3.5 h-3.5 text-amber-500" aria-hidden="true" />
                        <span className="text-amber-600 dark:text-amber-400 text-xs font-medium">
                          Aguardando SDK
                        </span>
                      </>
                    )}
                  </dd>
                </div>
                {typeof project.mau_current === 'number' && (
                  <div>
                    <dt className="text-xs text-slate-400 dark:text-slate-500">MAU atual</dt>
                    <dd className="font-medium text-slate-800 dark:text-slate-200 mt-0.5">
                      {project.mau_current.toLocaleString('pt-BR')}
                    </dd>
                  </div>
                )}
              </dl>
            )}
          </section>

          {/* API Key */}
          <section>
            <h3 className="text-xs font-semibold uppercase tracking-wider text-slate-400 mb-3">
              API Key
            </h3>
            <div className="flex items-center gap-2 p-3 rounded-lg bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700">
              <Key className="w-4 h-4 text-slate-400 flex-shrink-0" aria-hidden="true" />
              <code className="flex-1 text-xs font-mono text-slate-700 dark:text-slate-300 truncate">
                {keyVisible ? apiKey : maskKey(apiKey)}
              </code>
              <button
                onClick={() => setKeyVisible((v) => !v)}
                className="p-1.5 rounded text-slate-400 hover:bg-slate-200 dark:hover:bg-slate-700 transition-colors flex-shrink-0"
                aria-label={keyVisible ? 'Ocultar key' : 'Revelar key'}
              >
                {keyVisible ? (
                  <EyeOff className="w-3.5 h-3.5" />
                ) : (
                  <Eye className="w-3.5 h-3.5" />
                )}
              </button>
              <button
                onClick={() => void handleRegenerateKey()}
                disabled={regenerating}
                className="p-1.5 rounded text-slate-400 hover:bg-slate-200 dark:hover:bg-slate-700 transition-colors flex-shrink-0 disabled:opacity-50"
                aria-label="Regenerar API key"
                title="Regenerar API key"
              >
                <RotateCcw
                  className={`w-3.5 h-3.5 ${regenerating ? 'animate-spin' : ''}`}
                />
              </button>
            </div>
          </section>

          {/* Environments */}
          <section>
            <div className="flex items-center justify-between mb-3">
              <h3 className="text-xs font-semibold uppercase tracking-wider text-slate-400">
                Ambientes
              </h3>
              <button
                onClick={() => setShowEnvForm((v) => !v)}
                className="flex items-center gap-1 text-xs text-emerald-600 dark:text-emerald-400 hover:underline"
              >
                <Plus className="w-3 h-3" />
                Novo ambiente
              </button>
            </div>

            {showEnvForm && (
              <form
                onSubmit={(e) => void handleCreateEnv(e)}
                className="mb-3 p-3 rounded-lg border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800/50 space-y-2"
              >
                <div className="grid grid-cols-2 gap-2">
                  <div>
                    <label className="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">
                      Nome
                    </label>
                    <input
                      type="text"
                      value={newEnvName}
                      onChange={(e) => handleEnvNameChange(e.target.value)}
                      placeholder="Production"
                      required
                      className="w-full px-2.5 py-1.5 rounded-lg border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 text-xs focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent placeholder:text-slate-400"
                    />
                  </div>
                  <div>
                    <label className="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">
                      Slug
                    </label>
                    <input
                      type="text"
                      value={newEnvSlug}
                      onChange={(e) =>
                        setNewEnvSlug(
                          e.target.value.toLowerCase().replace(/[^a-z0-9-]/g, ''),
                        )
                      }
                      placeholder="production"
                      required
                      pattern="^[a-z0-9-]+$"
                      className="w-full px-2.5 py-1.5 rounded-lg border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 text-xs font-mono focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent placeholder:text-slate-400"
                    />
                  </div>
                </div>
                <div className="flex justify-end gap-2">
                  <button
                    type="button"
                    onClick={() => {
                      setShowEnvForm(false)
                      setNewEnvName('')
                      setNewEnvSlug('')
                    }}
                    className="px-3 py-1.5 rounded-lg text-xs text-slate-700 dark:text-slate-300 hover:bg-slate-200 dark:hover:bg-slate-700 transition-colors"
                  >
                    Cancelar
                  </button>
                  <button
                    type="submit"
                    disabled={creatingEnv}
                    className="flex items-center gap-1 px-3 py-1.5 rounded-lg bg-emerald-600 hover:bg-emerald-700 text-white text-xs font-medium transition-colors disabled:opacity-60"
                  >
                    {creatingEnv && <RefreshCw className="w-3 h-3 animate-spin" />}
                    Criar
                  </button>
                </div>
              </form>
            )}

            {envsLoading ? (
              <div className="flex items-center gap-2 py-4 text-xs text-slate-400">
                <RefreshCw className="w-3.5 h-3.5 animate-spin" />
                Carregando ambientes...
              </div>
            ) : envsError ? (
              <p className="text-xs text-rose-500 py-2">{envsError}</p>
            ) : environments.length === 0 ? (
              <p className="text-xs text-slate-400 py-2">Nenhum ambiente configurado.</p>
            ) : (
              <ul className="space-y-2">
                {environments.map((env) => (
                  <li
                    key={env.id}
                    className="p-3 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800"
                  >
                    <div className="flex items-center justify-between mb-2">
                      <div className="flex items-center gap-2">
                        <Server className="w-3.5 h-3.5 text-slate-400" aria-hidden="true" />
                        <span className="text-sm font-medium text-slate-800 dark:text-slate-200">
                          {env.name}
                        </span>
                        <code className="text-xs text-slate-400 font-mono">{env.slug}</code>
                      </div>
                    </div>
                    <div className="flex items-center gap-2 p-2 rounded bg-slate-50 dark:bg-slate-700/50">
                      <Key className="w-3.5 h-3.5 text-slate-400 flex-shrink-0" aria-hidden="true" />
                      <code className="flex-1 text-xs font-mono text-slate-600 dark:text-slate-300 truncate">
                        {envKeyVisible[env.id] ? envKeys[env.id] : maskKey(envKeys[env.id] ?? '')}
                      </code>
                      <button
                        onClick={() =>
                          setEnvKeyVisible((prev) => ({ ...prev, [env.id]: !prev[env.id] }))
                        }
                        className="p-1 rounded text-slate-400 hover:bg-slate-200 dark:hover:bg-slate-600 transition-colors flex-shrink-0"
                        aria-label={envKeyVisible[env.id] ? 'Ocultar key' : 'Revelar key'}
                      >
                        {envKeyVisible[env.id] ? (
                          <EyeOff className="w-3 h-3" />
                        ) : (
                          <Eye className="w-3 h-3" />
                        )}
                      </button>
                      <button
                        onClick={() => void handleRegenerateEnvKey(env.id)}
                        disabled={envRegenerating[env.id]}
                        className="p-1 rounded text-slate-400 hover:bg-slate-200 dark:hover:bg-slate-600 transition-colors flex-shrink-0 disabled:opacity-50"
                        aria-label="Regenerar key do ambiente"
                        title="Regenerar key"
                      >
                        <RotateCcw
                          className={`w-3 h-3 ${envRegenerating[env.id] ? 'animate-spin' : ''}`}
                        />
                      </button>
                    </div>
                  </li>
                ))}
              </ul>
            )}
          </section>
        </div>
      </div>
    </div>
  )
}

// ── Project Card ──────────────────────────────────────────────────────────────

interface ProjectCardProps {
  project: Project
  onClick: () => void
}

function ProjectCard({ project, onClick }: ProjectCardProps) {
  return (
    <button
      onClick={onClick}
      className="w-full text-left rounded-xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 hover:border-emerald-400 dark:hover:border-emerald-600 hover:shadow-md transition-all p-5 group"
    >
      <div className="flex items-start justify-between gap-3">
        <div className="flex items-center gap-3 min-w-0">
          <div className="w-9 h-9 rounded-lg bg-emerald-50 dark:bg-emerald-900/30 flex items-center justify-center flex-shrink-0">
            <FolderOpen
              className="w-4.5 h-4.5 text-emerald-600 dark:text-emerald-400"
              aria-hidden="true"
            />
          </div>
          <div className="min-w-0">
            <p className="font-semibold text-slate-900 dark:text-slate-100 truncate">
              {project.name}
            </p>
            <p className="text-xs text-slate-400 font-mono mt-0.5">{project.slug}</p>
          </div>
        </div>
        <ChevronRight
          className="w-4 h-4 text-slate-300 dark:text-slate-600 group-hover:text-emerald-500 transition-colors flex-shrink-0 mt-1"
          aria-hidden="true"
        />
      </div>

      <div className="flex items-center gap-3 mt-4">
        {planBadge(project.plan)}
        {project.is_active ? (
          <span className="flex items-center gap-1 text-xs text-emerald-600 dark:text-emerald-400">
            <CheckCircle className="w-3.5 h-3.5" aria-hidden="true" />
            Ativo
          </span>
        ) : (
          <span className="flex items-center gap-1 text-xs text-amber-500 dark:text-amber-400">
            <Server className="w-3.5 h-3.5" aria-hidden="true" />
            Aguardando SDK
          </span>
        )}
        {typeof project.mau_current === 'number' && (
          <span className="ml-auto text-xs text-slate-400">
            {project.mau_current.toLocaleString('pt-BR')} MAU
          </span>
        )}
      </div>
    </button>
  )
}

// ── Main Page ─────────────────────────────────────────────────────────────────

export default function Projects() {
  const [projects, setProjects] = useState<Project[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [createModalOpen, setCreateModalOpen] = useState(false)
  const [selectedProject, setSelectedProject] = useState<Project | null>(null)

  const fetchProjects = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const data = await projectsApi.list()
      setProjects(data.projects ?? [])
    } catch (err) {
      const e = err as { response?: { data?: { error?: string } }; message?: string }
      const msg = e?.response?.data?.error ?? e?.message ?? 'Falha ao carregar projetos'
      setError(msg)
      toast.error(msg)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void fetchProjects()
  }, [fetchProjects])

  const handleCreated = (project: Project) => {
    setProjects((prev) => [project, ...prev])
    setCreateModalOpen(false)
  }

  const handleUpdated = (updated: Project) => {
    setProjects((prev) => prev.map((p) => (p.id === updated.id ? updated : p)))
    setSelectedProject(updated)
  }

  const activeCount = projects.filter((p) => p.is_active).length

  return (
    <div className="space-y-5">
      {/* KPI strip */}
      {!loading && (
        <div className="grid grid-cols-2 sm:grid-cols-3 gap-3">
          <div className="rounded-xl border bg-gradient-to-br from-emerald-50 to-emerald-100 dark:from-emerald-950/40 dark:to-emerald-900/20 border-emerald-200 dark:border-emerald-800 p-4">
            <p className="text-xs font-medium text-slate-500 dark:text-slate-400">Total</p>
            <p className="text-2xl font-bold text-emerald-700 dark:text-emerald-300 mt-0.5">
              {projects.length}
            </p>
          </div>
          <div className="rounded-xl border bg-gradient-to-br from-violet-50 to-violet-100 dark:from-violet-950/40 dark:to-violet-900/20 border-violet-200 dark:border-violet-800 p-4">
            <p className="text-xs font-medium text-slate-500 dark:text-slate-400">Ativos</p>
            <p className="text-2xl font-bold text-violet-700 dark:text-violet-300 mt-0.5">
              {activeCount}
            </p>
          </div>
          <div className="rounded-xl border bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-800/60 dark:to-slate-800/30 border-slate-200 dark:border-slate-700 p-4 col-span-2 sm:col-span-1">
            <p className="text-xs font-medium text-slate-500 dark:text-slate-400">Inativos</p>
            <p className="text-2xl font-bold text-slate-700 dark:text-slate-200 mt-0.5">
              {projects.length - activeCount}
            </p>
          </div>
        </div>
      )}

      {/* Page header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-xl font-bold text-slate-900 dark:text-slate-100">Projetos</h1>
          <p className="text-sm text-slate-500 dark:text-slate-400 mt-0.5">
            Gerencie projetos, planos e ambientes multi-tenant
          </p>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => void fetchProjects()}
            disabled={loading}
            className="p-2 rounded-lg text-slate-500 hover:bg-slate-100 dark:text-slate-400 dark:hover:bg-slate-800 transition-colors disabled:opacity-50"
            aria-label="Recarregar projetos"
          >
            <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
          </button>
          <button
            onClick={() => setCreateModalOpen(true)}
            className="flex items-center gap-2 px-4 py-2 rounded-lg bg-emerald-600 hover:bg-emerald-700 text-white text-sm font-medium transition-colors"
          >
            <Plus className="w-4 h-4" aria-hidden="true" />
            Novo projeto
          </button>
        </div>
      </div>

      {/* Content */}
      {loading ? (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {[1, 2, 3].map((i) => (
            <div
              key={i}
              className="rounded-xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 p-5 animate-pulse"
            >
              <div className="flex items-center gap-3 mb-4">
                <div className="w-9 h-9 rounded-lg bg-slate-200 dark:bg-slate-700" />
                <div className="space-y-1.5 flex-1">
                  <div className="h-3.5 bg-slate-200 dark:bg-slate-700 rounded w-3/4" />
                  <div className="h-2.5 bg-slate-100 dark:bg-slate-800 rounded w-1/2" />
                </div>
              </div>
              <div className="h-5 bg-slate-100 dark:bg-slate-800 rounded w-1/3" />
            </div>
          ))}
        </div>
      ) : error ? (
        <div className="rounded-xl border border-rose-200 dark:border-rose-800 bg-rose-50 dark:bg-rose-900/20 p-6 text-center">
          <p className="text-sm text-rose-600 dark:text-rose-400">{error}</p>
          <button
            onClick={() => void fetchProjects()}
            className="mt-3 text-xs text-rose-700 dark:text-rose-300 underline"
          >
            Tentar novamente
          </button>
        </div>
      ) : projects.length === 0 ? (
        <div className="rounded-xl border border-dashed border-slate-300 dark:border-slate-600 bg-slate-50 dark:bg-slate-800/30 p-12 text-center">
          <FolderOpen
            className="w-10 h-10 text-slate-300 dark:text-slate-600 mx-auto mb-3"
            aria-hidden="true"
          />
          <p className="text-sm font-medium text-slate-600 dark:text-slate-400">
            Nenhum projeto encontrado
          </p>
          <p className="text-xs text-slate-400 dark:text-slate-500 mt-1">
            Crie seu primeiro projeto para começar a rastrear eventos e feature flags.
          </p>
          <button
            onClick={() => setCreateModalOpen(true)}
            className="mt-4 flex items-center gap-2 px-4 py-2 rounded-lg bg-emerald-600 hover:bg-emerald-700 text-white text-sm font-medium transition-colors mx-auto"
          >
            <Plus className="w-4 h-4" aria-hidden="true" />
            Criar primeiro projeto
          </button>
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {projects.map((p) => (
            <ProjectCard key={p.id} project={p} onClick={() => setSelectedProject(p)} />
          ))}
        </div>
      )}

      {/* Modals */}
      {createModalOpen && (
        <CreateProjectModal
          onClose={() => setCreateModalOpen(false)}
          onSuccess={handleCreated}
        />
      )}
      {selectedProject && (
        <ProjectDetailModal
          project={selectedProject}
          onClose={() => setSelectedProject(null)}
          onUpdated={handleUpdated}
        />
      )}
    </div>
  )
}
