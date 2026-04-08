import { useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { setToken } from '../lib/auth'

export default function AuthCallback() {
  const navigate = useNavigate()

  useEffect(() => {
    const params = new URLSearchParams(window.location.search)
    const token = params.get('token')

    if (token) {
      setToken(token)
      navigate('/', { replace: true })
    } else {
      navigate('/login', { replace: true })
    }
  }, [navigate])

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="text-center">
        <div className="w-8 h-8 border-2 border-gray-200 border-t-indigo-600 rounded-full animate-spin mx-auto mb-4" />
        <p className="text-sm text-gray-500">Autenticando...</p>
      </div>
    </div>
  )
}
