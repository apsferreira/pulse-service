import React, { createContext, useContext, useEffect, useState, useCallback } from 'react'
import axios from 'axios'
import api from '../services/api'

const AUTH_SERVICE_URL = (import.meta.env?.VITE_AUTH_SERVICE_URL as string) || 'http://localhost:3010'
const AUTH_MODE = ((import.meta.env?.VITE_AUTH_MODE as string) ?? 'external') === 'local' ? 'local' : 'external'

interface User {
  id: string
  name: string
  email: string
}

interface AuthContextType {
  user: User | null
  isAuthenticated: boolean
  isLoading: boolean
  authMode: 'external' | 'local'
  loginWithTokens: (accessToken: string, refreshToken: string) => Promise<void>
  loginLocal: (email: string, password: string) => Promise<void>
  logout: () => void
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  const restoreSession = useCallback(async () => {
    const token = sessionStorage.getItem('access_token')

    if (!token) {
      setIsLoading(false)
      return
    }

    if (AUTH_MODE === 'local') {
      // Valida o token local via /api/v1/auth/me
      try {
        const { data } = await api.get<User>('/auth/me')
        sessionStorage.setItem('pulse_user', JSON.stringify(data))
        setUser(data)
      } catch {
        sessionStorage.removeItem('access_token')
        sessionStorage.removeItem('pulse_user')
      }
    } else {
      // Modo external: restaura do sessionStorage sem chamar auth-service
      const storedUser = sessionStorage.getItem('pulse_user')
      if (storedUser) {
        try {
          setUser(JSON.parse(storedUser) as User)
        } catch {
          sessionStorage.removeItem('access_token')
          sessionStorage.removeItem('refresh_token')
          sessionStorage.removeItem('pulse_user')
        }
      }
    }

    setIsLoading(false)
  }, [])

  useEffect(() => {
    void restoreSession()
  }, [restoreSession])

  const loginWithTokens = async (accessToken: string, refreshToken: string) => {
    sessionStorage.setItem('access_token', accessToken)
    sessionStorage.setItem('refresh_token', refreshToken)
    // Buscar perfil do usuário no auth-service com o token recebido
    const { data } = await axios.post<{ access_token: string; refresh_token: string; user: User }>(
      `${AUTH_SERVICE_URL}/api/v1/auth/refresh`,
      { refresh_token: refreshToken }
    )
    sessionStorage.setItem('access_token', data.access_token)
    sessionStorage.setItem('refresh_token', data.refresh_token)
    sessionStorage.setItem('pulse_user', JSON.stringify(data.user))
    setUser(data.user)
  }

  const loginLocal = async (email: string, password: string) => {
    const { data } = await api.post<{ access_token: string; user: User }>(
      '/auth/login',
      { email, password }
    )
    sessionStorage.setItem('access_token', data.access_token)
    sessionStorage.setItem('pulse_user', JSON.stringify(data.user))
    setUser(data.user)
  }

  const logout = () => {
    if (AUTH_MODE === 'local') {
      const token = sessionStorage.getItem('access_token')
      if (token) {
        void api.post('/auth/logout').catch(() => { /* logout silencioso */ })
      }
      sessionStorage.removeItem('access_token')
      sessionStorage.removeItem('pulse_user')
      setUser(null)
      window.location.replace('/login')
    } else {
      const refreshToken = sessionStorage.getItem('refresh_token')
      if (refreshToken) {
        void axios.post(`${AUTH_SERVICE_URL}/api/v1/auth/logout`, { refresh_token: refreshToken })
          .catch(() => { /* logout silencioso */ })
      }
      sessionStorage.removeItem('access_token')
      sessionStorage.removeItem('refresh_token')
      sessionStorage.removeItem('pulse_user')
      setUser(null)
      const redirectUri = encodeURIComponent(`${window.location.origin}/auth/callback`)
      window.location.replace(`${AUTH_SERVICE_URL}/auth/login?redirect_uri=${redirectUri}&product=pulse`)
    }
  }

  return (
    <AuthContext.Provider value={{ user, isAuthenticated: !!user, isLoading, authMode: AUTH_MODE, loginWithTokens, loginLocal, logout }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth(): AuthContextType {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth deve ser usado dentro de AuthProvider')
  return ctx
}
