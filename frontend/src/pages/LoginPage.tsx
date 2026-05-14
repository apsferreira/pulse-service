import { useEffect, useState } from 'react'
import { Navigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'

const AUTH_SERVICE_URL = (import.meta.env?.VITE_AUTH_SERVICE_URL as string) || 'http://localhost:3010'

export default function LoginPage() {
  const { isAuthenticated, isLoading, authMode, loginLocal } = useAuth()

  // ── Modo external ────────────────────────────────────────────────────────────
  useEffect(() => {
    if (authMode !== 'external') return
    if (isLoading || isAuthenticated) return
    const redirectUri = encodeURIComponent(`${window.location.origin}/auth/callback`)
    window.location.replace(
      `${AUTH_SERVICE_URL}/auth/login?redirect_uri=${redirectUri}&product=pulse`
    )
  }, [authMode, isLoading, isAuthenticated])

  // ── Modo local — estado do formulário ───────────────────────────────────────
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [formError, setFormError] = useState('')
  const [submitting, setSubmitting] = useState(false)

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-slate-950">
        <div className="w-8 h-8 border-4 border-violet-500 border-t-transparent rounded-full animate-spin" />
      </div>
    )
  }

  if (isAuthenticated) return <Navigate to="/" replace />

  // ── Modo external: aguardando redirect ──────────────────────────────────────
  if (authMode === 'external') {
    return (
      <div className="min-h-screen flex items-center justify-center bg-slate-950">
        <div className="flex flex-col items-center gap-3">
          <div className="w-8 h-8 border-4 border-violet-500 border-t-transparent rounded-full animate-spin" />
          <p className="text-sm text-slate-400">Redirecionando para login…</p>
        </div>
      </div>
    )
  }

  // ── Modo local: formulário email+senha ───────────────────────────────────────
  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    setFormError('')
    setSubmitting(true)
    try {
      await loginLocal(email, password)
      // AuthContext atualiza o estado; Navigate abaixo cuida do redirect
    } catch (err: unknown) {
      const message =
        err instanceof Error
          ? err.message
          : 'Erro ao autenticar. Verifique suas credenciais.'
      setFormError(message)
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-slate-950 px-4">
      <div className="w-full max-w-sm">
        <div className="rounded-xl border border-slate-800 bg-slate-900 p-8 shadow-xl">
          {/* Logo */}
          <div className="mb-8 text-center">
            <span className="text-2xl font-bold tracking-tight text-violet-400">Pulse</span>
          </div>

          <form onSubmit={(e) => { void handleSubmit(e) }} noValidate>
            <div className="flex flex-col gap-4">
              {/* Email */}
              <div className="flex flex-col gap-1.5">
                <label htmlFor="email" className="text-sm font-medium text-slate-300">
                  E-mail
                </label>
                <input
                  id="email"
                  type="email"
                  autoComplete="email"
                  placeholder="admin@suaempresa.com"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  required
                  disabled={submitting}
                  className="rounded-md border border-slate-700 bg-slate-800 px-3 py-2 text-sm text-slate-100 placeholder:text-slate-500 focus:border-violet-500 focus:outline-none focus:ring-1 focus:ring-violet-500 disabled:opacity-50"
                />
              </div>

              {/* Senha */}
              <div className="flex flex-col gap-1.5">
                <label htmlFor="password" className="text-sm font-medium text-slate-300">
                  Senha
                </label>
                <input
                  id="password"
                  type="password"
                  autoComplete="current-password"
                  placeholder="••••••••"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                  disabled={submitting}
                  className="rounded-md border border-slate-700 bg-slate-800 px-3 py-2 text-sm text-slate-100 placeholder:text-slate-500 focus:border-violet-500 focus:outline-none focus:ring-1 focus:ring-violet-500 disabled:opacity-50"
                />
              </div>

              {/* Botão submit */}
              <button
                type="submit"
                disabled={submitting}
                className="mt-2 flex items-center justify-center gap-2 rounded-md bg-violet-600 px-4 py-2 text-sm font-semibold text-white transition-colors hover:bg-violet-500 focus:outline-none focus:ring-2 focus:ring-violet-500 focus:ring-offset-2 focus:ring-offset-slate-900 disabled:cursor-not-allowed disabled:opacity-60"
              >
                {submitting && (
                  <span
                    className="h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent"
                    aria-hidden="true"
                  />
                )}
                {submitting ? 'Entrando…' : 'Entrar'}
              </button>

              {/* Mensagem de erro */}
              {formError && (
                <p role="alert" className="text-center text-sm text-red-400">
                  {formError}
                </p>
              )}
            </div>
          </form>
        </div>
      </div>
    </div>
  )
}
