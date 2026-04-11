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
  product?: string;
  user_id?: string;
  environment?: string;
  properties?: Record<string, unknown>;
}

// BKL-1047: batch queue de eventos — acumula eventos e faz flush automático.
export interface BatchQueueOptions {
  /** Intervalo de flush em ms (default: 5000) */
  flushInterval?: number;
  /** Tamanho máximo do buffer antes de flush imediato (default: 25) */
  maxBatchSize?: number;
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

  // BKL-1047: batch queue
  private eventQueue: TrackEventRequest[] = [];
  private flushTimer: ReturnType<typeof setTimeout> | null = null;
  private readonly flushInterval: number;
  private readonly maxBatchSize: number;

  constructor(options: PulseClientOptions & BatchQueueOptions) {
    if (!options.baseURL) {
      throw new PulseError("baseURL is required");
    }
    this.baseURL = options.baseURL.replace(/\/+$/, "");
    this.timeout = options.timeout ?? 10000;
    // BKL-1047: batch queue config
    this.flushInterval = options.flushInterval ?? 5000;
    this.maxBatchSize = options.maxBatchSize ?? 25;
    // BKL-1046: guard SSR — globalThis.fetch é undefined em Node < 18 e Deno sem --allow-net
    // Usar options.fetch injetado pelo caller (ex: node-fetch) para compatibilidade SSR.
    if (options.fetch) {
      this.fetchFn = options.fetch;
    } else if (typeof globalThis !== "undefined" && typeof globalThis.fetch === "function") {
      this.fetchFn = globalThis.fetch.bind(globalThis);
    } else {
      // SSR sem fetch disponível — lança erro descritivo ao invés de crash silencioso
      this.fetchFn = () => {
        throw new PulseError(
          "fetch não disponível neste ambiente. Passe uma implementação via options.fetch (ex: import fetch from 'node-fetch')"
        );
      };
    }
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
   * POST /api/v1/events/track — envia evento de tracking imediatamente (sem queue).
   */
  async trackEvent(event: TrackEventRequest): Promise<TrackEventResponse> {
    if (!event.event_name || event.event_name.trim() === "") {
      throw new PulseError("event_name is required");
    }
    return this.request<TrackEventResponse>("POST", "/api/v1/events/track", event);
  }

  /**
   * BKL-1047: Adiciona evento à fila batch.
   * Faz flush automático quando a fila atinge maxBatchSize ou após flushInterval ms.
   * Não lança erro — falhas de rede são silenciosas (fire-and-forget).
   */
  queue(event: TrackEventRequest): void {
    if (!event.event_name || event.event_name.trim() === "") return;

    this.eventQueue.push(event);

    // Flush imediato se atingiu o limite
    if (this.eventQueue.length >= this.maxBatchSize) {
      this.flush();
      return;
    }

    // Agenda flush periódico se não houver um ativo
    if (this.flushTimer === null) {
      this.flushTimer = setTimeout(() => this.flush(), this.flushInterval);
    }
  }

  /**
   * BKL-1047: Envia todos os eventos acumulados via POST /api/v1/events/batch.
   * Chamado automaticamente pelo timer ou quando o buffer está cheio.
   * Pode ser chamado manualmente (ex: beforeunload).
   */
  flush(): void {
    if (this.flushTimer !== null) {
      clearTimeout(this.flushTimer);
      this.flushTimer = null;
    }
    if (this.eventQueue.length === 0) return;

    const batch = this.eventQueue.splice(0, this.maxBatchSize);
    const payload = {
      events: batch.map((e) => ({
        product: e.product ?? "default",
        event_type: e.event_name,
        user_id: e.user_id ? [e.user_id] : undefined,
        properties: e.properties,
        environment: e.environment,
      })),
    };

    // Fire-and-forget — falhas de rede não afetam o caller
    this.request("POST", "/api/v1/events/batch", payload).catch(() => {
      // Re-enqueue eventos que falharam (no início da fila)
      this.eventQueue.unshift(...batch);
    });
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
