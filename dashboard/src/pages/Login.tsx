import { useEffect } from 'react'
import { redirectToLogin } from '../lib/auth'

export default function Login() {
  useEffect(() => {
    redirectToLogin()
  }, [])

  return (
    <div className="min-h-screen flex items-center justify-center bg-slate-50 dark:bg-slate-950">
      <div className="text-center">
        <div className="w-8 h-8 border-2 border-slate-200 border-t-violet-600 rounded-full animate-spin mx-auto mb-4" />
        <p className="text-sm text-slate-500 dark:text-slate-400">Redirecionando para login...</p>
      </div>
    </div>
  )
}
