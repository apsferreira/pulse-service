/**
 * @pulse-iit/react-native — Tipos públicos
 */

export interface PulseConfig {
  /** URL base da API Pulse. Ex: "https://pulse-api.institutoitinerante.com.br" */
  apiUrl: string
  /** Chave do projeto (X-Project-Key). */
  projectKey: string
  /** Token de serviço para autenticação (X-Service-Token). */
  serviceToken?: string
  /** Intervalo de flush do batch queue em ms. Default: 5000 */
  flushInterval?: number
  /** Timeout de requisições em ms. Default: 10000 */
  timeout?: number
  /** Ativa logs de debug. Default: false */
  debug?: boolean
}

export interface FeatureFlag {
  key: string
  enabled: boolean
}

export interface TrackEvent {
  event_name: string
  user_id: string
  project_key: string
  platform: 'react-native'
  timestamp: string
  properties?: Record<string, unknown>
}

export interface FlagsMap {
  [key: string]: boolean
}

export interface VariantsMap {
  [key: string]: string
}

export interface ExperimentResult {
  variant: string
  trackExposure: () => void
}

export interface FeatureFlagResult {
  value: boolean
  loading: boolean
}

export interface FeatureFlagsResult {
  flags: FlagsMap
  loading: boolean
}
