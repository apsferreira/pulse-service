/**
 * @iit/pulse-sdk — SDK JS/TS para pulse-service
 * Feature Flags + Event Tracking
 */

// --- Types ---

export interface PulseClientOptions {
  /** Base URL do pulse-service (ex: "https://pulse.institutoitinerante.com.br") */
  baseURL: string;
  /** Timeout em ms (default: 10000) */
  timeout?: number;
  /** Custom fetch implementation (default: globalThis.fetch) */
  fetch?: typeof globalThis.fetch;
}

export interface FlagsMap {
  [key: string]: boolean;
}

export interface FlagsMapResponse {
  flags: FlagsMap;
}

export interface FeatureFlag {
  id: string;
  key: string;
  description: string;
  enabled: boolean;
  created_at: string;
  updated_at: string;
}

export interface TrackEventRequest {
  event_name: string;
  user_id?: string;
  environment?: string;
  properties?: Record<string, unknown>;
}

export interface TrackEventResponse {
  id: string;
  product: string;
  event_type: string;
  user_id?: string;
  session_id?: string;
  properties?: Record<string, unknown>;
  created_at: string;
}

export interface ErrorResponse {
  error: string;
}

// --- Validation ---

const VALID_KEY_PATTERN = /^[a-zA-Z0-9_\-\.]{2,100}$/;

function validateKey(key: string): void {
  if (!VALID_KEY_PATTERN.test(key)) {
    throw new PulseError(`Invalid flag key: "${key}". Must be 2-100 chars, alphanumeric/underscore/hyphen/dot only.`);
  }
}

// --- Error ---

export class PulseError extends Error {
  public readonly statusCode?: number;

  constructor(message: string, statusCode?: number) {
    super(message);
    this.name = "PulseError";
    this.statusCode = statusCode;
  }
}

// --- Client ---

export class PulseClient {
  private readonly baseURL: string;
  private readonly timeout: number;
  private readonly fetchFn: typeof globalThis.fetch;
  private cachedFlags: FlagsMap | null = null;

  constructor(options: PulseClientOptions) {
    if (!options.baseURL) {
      throw new PulseError("baseURL is required");
    }
    this.baseURL = options.baseURL.replace(/\/+$/, "");
    this.timeout = options.timeout ?? 10000;
    this.fetchFn = options.fetch ?? globalThis.fetch.bind(globalThis);
  }

  /**
   * GET /flags — retorna mapa de flags públicas {key: enabled}.
   */
  async getFlags(): Promise<FlagsMap> {
    const response = await this.request<FlagsMapResponse>("GET", "/flags");
    this.cachedFlags = response.flags ?? {};
    return this.cachedFlags;
  }

  /**
   * GET /api/v1/flags/:key — retorna flag individual com detalhes completos.
   * Requer autenticação (service token no servidor).
   */
  async getFlag(key: string): Promise<FeatureFlag | null> {
    validateKey(key);
    try {
      return await this.request<FeatureFlag>("GET", `/api/v1/flags/${encodeURIComponent(key)}`);
    } catch (err) {
      if (err instanceof PulseError && err.statusCode === 404) {
        return null;
      }
      throw err;
    }
  }

  /**
   * Verifica se uma flag está ativa usando o endpoint público GET /flags.
   * Usa cache local se getFlags() já foi chamado.
   * Retorna false se a flag não existe.
   */
  async isEnabled(key: string): Promise<boolean> {
    validateKey(key);
    if (this.cachedFlags === null) {
      await this.getFlags();
    }
    return this.cachedFlags?.[key] === true;
  }

  /**
   * Invalida o cache de flags, forçando nova requisição no próximo isEnabled().
   */
  invalidateCache(): void {
    this.cachedFlags = null;
  }

  /**
   * POST /api/v1/events/track — envia evento de tracking.
   */
  async trackEvent(event: TrackEventRequest): Promise<TrackEventResponse> {
    if (!event.event_name || event.event_name.trim() === "") {
      throw new PulseError("event_name is required");
    }
    return this.request<TrackEventResponse>("POST", "/api/v1/events/track", event);
  }

  // --- Internal ---

  private async request<T>(method: string, path: string, body?: unknown): Promise<T> {
    const url = this.baseURL + path;

    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), this.timeout);

    try {
      const init: RequestInit = {
        method,
        headers: {
          "Content-Type": "application/json",
          Accept: "application/json",
        },
        signal: controller.signal,
      };

      if (body !== undefined) {
        init.body = JSON.stringify(body);
      }

      const response = await this.fetchFn(url, init);

      if (!response.ok) {
        let errorMsg = `HTTP ${response.status}`;
        try {
          const errBody = (await response.json()) as ErrorResponse;
          if (errBody.error) {
            errorMsg = errBody.error;
          }
        } catch {
          // ignore parse error
        }
        throw new PulseError(errorMsg, response.status);
      }

      return (await response.json()) as T;
    } catch (err) {
      if (err instanceof PulseError) throw err;
      if (err instanceof DOMException && err.name === "AbortError") {
        throw new PulseError(`Request timeout after ${this.timeout}ms`);
      }
      throw new PulseError(`Request failed: ${(err as Error).message}`);
    } finally {
      clearTimeout(timeoutId);
    }
  }
}

export default PulseClient;
