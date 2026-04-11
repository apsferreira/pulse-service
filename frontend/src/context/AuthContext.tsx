import React, { createContext, useContext, useEffect, useState, useCallback } from 'react'
import axios from 'axios'

const AUTH_SERVICE_URL = (import.meta.env?.VITE_AUTH_SERVICE_URL as string) || 'http://localhost:3010'

interface User {
  id: string
  name: string
  email: string
}

interface AuthContextType {
  user: User | null
  isAuthenticated: boolean
  isLoading: boolean
  loginWithTokens: (accessToken: string, refreshToken: string) => Promise<void>
  logout: () => void
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  const restoreSession = useCallback(async () => {
    const token = localStorage.getItem('access_token')
    const storedUser = localStorage.getItem('pulse_user')
    if (token && storedUser) {
      try {
        setUser(JSON.parse(storedUser) as User)
      } catch {
        localStorage.removeItem('access_token')
        localStorage.removeItem('refresh_token')
        localStorage.removeItem('pulse_user')
      }
    }
    setIsLoading(false)
  }, [])

  useEffect(() => {
    void restoreSession()
  }, [restoreSession])

  const loginWithTokens = async (accessToken: string, refreshToken: string) => {
    localStorage.setItem('access_token', accessToken)
    localStorage.setItem('refresh_token', refreshToken)
    // Buscar perfil do usuário no auth-service com o token recebido
    const { data } = await axios.post<{ access_token: string; refresh_token: string; user: User }>(
      `${AUTH_SERVICE_URL}/api/v1/auth/refresh`,
      { refresh_token: refreshToken }
    )
    localStorage.setItem('access_token', data.access_token)
    localStorage.setItem('refresh_token', data.refresh_token)
    localStorage.setItem('pulse_user', JSON.stringify(data.user))
    setUser(data.user)
  }

  const logout = () => {
    const refreshToken = localStorage.getItem('refresh_token')
    if (refreshToken) {
      void axios.post(`${AUTH_SERVICE_URL}/api/v1/auth/logout`, { refresh_token: refreshToken })
        .catch(() => { /* logout silencioso */ })
    }
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
    localStorage.removeItem('pulse_user')
    setUser(null)
    const redirectUri = encodeURIComponent(`${window.location.origin}/auth/callback`)
    window.location.replace(`${AUTH_SERVICE_URL}/auth/login?redirect_uri=${redirectUri}&product=pulse`)
  }

  return (
    <AuthContext.Provider value={{ user, isAuthenticated: !!user, isLoading, loginWithTokens, logout }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth(): AuthContextType {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth deve ser usado dentro de AuthProvider')
  return ctx
}
