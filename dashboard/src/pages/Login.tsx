import { useEffect } from 'react'
import { redirectToLogin } from '../lib/auth'

export default function Login() {
  useEffect(() => {
    redirectToLogin()
  }, [])

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="text-center">
        <div className="w-8 h-8 border-2 border-gray-200 border-t-indigo-600 rounded-full animate-spin mx-auto mb-4" />
        <p className="text-sm text-gray-500">Redirecionando para login...</p>
      </div>
    </div>
  )
}
