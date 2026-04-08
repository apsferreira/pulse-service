const TOKEN_KEY = 'pulse_token'
const AUTH_SERVICE_URL = (import.meta.env?.VITE_AUTH_SERVICE_URL as string) || 'https://auth.institutoitinerante.com.br'

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

export function setToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token)
}

export function clearToken(): void {
  localStorage.removeItem(TOKEN_KEY)
}

export function isAuthenticated(): boolean {
  const token = getToken()
  if (!token) return false
  try {
    const payload = JSON.parse(atob(token.split('.')[1]))
    return payload.exp * 1000 > Date.now()
  } catch {
    return false
  }
}

export function redirectToLogin(): void {
  const redirectUri = encodeURIComponent(`${window.location.origin}/auth/callback`)
  window.location.href = `${AUTH_SERVICE_URL}/auth/login?redirect_uri=${redirectUri}&product=pulse`
}

export function logout(): void {
  clearToken()
  redirectToLogin()
}
