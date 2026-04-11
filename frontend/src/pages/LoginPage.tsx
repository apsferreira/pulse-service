import { useEffect } from 'react'
import { Navigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'

const AUTH_SERVICE_URL = (import.meta.env?.VITE_AUTH_SERVICE_URL as string) || 'http://localhost:3010'

export default function LoginPage() {
  const { isAuthenticated, isLoading } = useAuth()

  useEffect(() => {
    if (isLoading || isAuthenticated) return
    const redirectUri = encodeURIComponent(`${window.location.origin}/auth/callback`)
    window.location.replace(
      `${AUTH_SERVICE_URL}/auth/login?redirect_uri=${redirectUri}&product=pulse`
    )
  }, [isLoading, isAuthenticated])

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-slate-950">
        <div className="w-8 h-8 border-4 border-violet-500 border-t-transparent rounded-full animate-spin" />
      </div>
    )
  }

  if (isAuthenticated) return <Navigate to="/" replace />

  return (
    <div className="min-h-screen flex items-center justify-center bg-slate-950">
      <div className="flex flex-col items-center gap-3">
        <div className="w-8 h-8 border-4 border-violet-500 border-t-transparent rounded-full animate-spin" />
        <p className="text-sm text-slate-400">Redirecionando para login…</p>
      </div>
    </div>
  )
}
