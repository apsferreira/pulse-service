import { useEffect, useRef, useCallback } from 'react'
import { readEnv } from '../env'

const API_BASE_URL = readEnv('VITE_API_URL')
// SERVICE_TOKEN é o token mestre de serviço — o mesmo que protege endpoints
// administrativos sensíveis (billing, license, user export, etc). É mantido
// aqui apenas como fallback para ambientes self-hosted sem auth-service
// (mesmo padrão já usado em services/api.ts). Em produção, a sessão do
// dashboard autentica via JWT do auth-service (sessionStorage.access_token) —
// nunca o token mestre.
const SERVICE_TOKEN = readEnv('VITE_SERVICE_TOKEN')

export interface RealtimeEvent {
  type: string
  event: {
    id?: string
    product: string
    event_type: string
    user_id?: string
    session_id?: string
    properties?: Record<string, unknown>
    created_at?: string
  }
}

interface UseRealtimeEventsOptions {
  /** ID do projeto/produto para filtrar eventos. Vazio = todos. */
  projectId?: string
  /** Callback invocado a cada evento recebido via WebSocket */
  onEvent: (event: RealtimeEvent) => void
  /** Habilitar/desabilitar a conexão (default: true) */
  enabled?: boolean
}

/**
 * useRealtimeEvents — conecta ao endpoint WebSocket do pulse-service
 * e invoca onEvent para cada mensagem recebida.
 *
 * Reconecta automaticamente com backoff exponencial em caso de erro.
 */
export function useRealtimeEvents({ projectId = '', onEvent, enabled = true }: UseRealtimeEventsOptions) {
  const wsRef = useRef<WebSocket | null>(null)
  const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const reconnectAttemptsRef = useRef(0)
  const onEventRef = useRef(onEvent)
  onEventRef.current = onEvent

  const connect = useCallback(() => {
    // SEC: usar o JWT de sessão do dashboard (emitido pelo auth-service, curta
    // duração, iss/aud validados pelo backend) em vez do token mestre de
    // serviço sempre que uma sessão autenticada existir — mesmo padrão de
    // services/api.ts. O token mestre só é usado em ambientes self-hosted
    // sem auth-service (AuthMode=local), nunca em produção com login ativo.
    const sessionToken = sessionStorage.getItem('access_token')
    const authToken = sessionToken || SERVICE_TOKEN

    // WebSocket endpoint não implementado — desabilita silenciosamente
    if (!enabled || !authToken || !API_BASE_URL) return

    // Converte URL HTTP para WS
    const base = API_BASE_URL.replace(/^http/, 'ws')
    const params = new URLSearchParams({ token: authToken })
    if (projectId) params.set('project_id', projectId)
    const url = `${base}/api/v1/ws?${params.toString()}`

    const ws = new WebSocket(url)
    wsRef.current = ws

    ws.onopen = () => {
      reconnectAttemptsRef.current = 0
    }

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data) as RealtimeEvent
        onEventRef.current(data)
      } catch {
        // mensagem não-JSON ignorada
      }
    }

    ws.onerror = () => {
      // onerror sempre é seguido de onclose — reconexão ocorre lá
    }

    ws.onclose = () => {
      wsRef.current = null
      if (!enabled) return

      // Backoff exponencial: 1s, 2s, 4s, 8s, máximo 30s
      const delay = Math.min(1000 * Math.pow(2, reconnectAttemptsRef.current), 30_000)
      reconnectAttemptsRef.current++

      reconnectTimeoutRef.current = setTimeout(() => {
        connect()
      }, delay)
    }
  }, [enabled, projectId])

  useEffect(() => {
    if (!enabled) return

    connect()

    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current)
      }
      if (wsRef.current) {
        // Fecha sem reconectar
        wsRef.current.onclose = null
        wsRef.current.close()
        wsRef.current = null
      }
    }
  }, [enabled, connect])
}
