/**
 * @iit/pulse-sdk — SDK JS/TS para pulse-service
 * Feature Flags + Event Tracking
 *
 * BKL-258: fail-open — retorna defaults quando backend indisponível.
 *   - Timeout de 2s para fetch de flags (não bloqueia a aplicação)
 *   - failOpen: true como padrão nas opções de inicialização
 *   - defaults por flag: pulse.init({ apiKey: '...', defaults: { 'nova-feature': false } })
 */

// --- Types ---

/** BKL-258: defaults por flag — usados quando fail-open ativa */
export type DefaultsMap = Record<string, boolean>;

export interface PulseClientOptions {
  /** Base URL do pulse-service (ex: "https://seu-pulse.exemplo.com") */
  baseURL: string;
  /**
   * Timeout em ms para fetch de flags (default: 2000).
   * BKL-258: 2s para não bloquear a aplicação — eventos usam timeout maior.
   */
  timeout?: number;
  /** Custom fetch implementation (default: globalThis.fetch) */
  fetch?: typeof globalThis.fetch;
  /**
   * BKL-258: comportamento fail-open.
   * Quando true (default), retorna defaults em vez de lançar erro
   * se o backend estiver indisponível ou timeout expirar.
   */
  failOpen?: boolean;
  /**
   * BKL-258: valores default por flag.
   * Usados quando fail-open ativa ou SDK não foi inicializado.
   * Ex: { 'nova-feature': false, 'dark-mode': true }
   */
  defaults?: DefaultsMap;
}

/**
 * BKL-258: opções de inicialização via pulse.init().
 * Alias conveniente para PulseClientOptions usando `apiKey` em vez de `baseURL`.
 */
export interface PulseInitOptions {
  /** API key do projeto (equivale a baseURL — usado como identificador) */
  apiKey: string;
  /** Base URL do pulse-service (ex: "https://seu-pulse.exemplo.com") */
  baseURL?: string;
  /** BKL-258: valores default por flag */
  defaults?: DefaultsMap;
  /** BKL-258: fail-open ativo (default: true) */
  failOpen?: boolean;
  /** Timeout em ms para fetch de flags (default: 2000) */
  timeout?: number;
  /** Custom fetch implementation */
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

// BKL-1054: chave do localStorage para flags de fallback
const FALLBACK_STORAGE_KEY = "pulse_flags_fallback";
const FALLBACK_TTL_MS = 5 * 60 * 1000; // 5 minutos

export class PulseClient {
  private readonly baseURL: string;
  private readonly timeout: number;
  private readonly fetchFn: typeof globalThis.fetch;
  private cachedFlags: FlagsMap | null = null;
  // BKL-258: fail-open config
  private readonly failOpen: boolean;
  private readonly defaults: DefaultsMap;

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
    // BKL-258: timeout padrão reduzido para 2s em flags para não bloquear a app
    this.timeout = options.timeout ?? 2000;
    // BKL-258: fail-open e defaults
    this.failOpen = options.failOpen ?? true;
    this.defaults = options.defaults ?? {};
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
   * BKL-1054: em caso de erro, usa fallback de localStorage (TTL 5min).
   * BKL-258: se failOpen=true e nenhum fallback disponível, retorna defaults
   * em vez de lançar erro.
   */
  async getFlags(): Promise<FlagsMap> {
    try {
      const response = await this.request<FlagsMapResponse>("GET", "/flags");
      const flags = response.flags ?? {};
      this.cachedFlags = flags;
      this.saveFallback(flags); // BKL-1054: persiste para fallback
      return flags;
    } catch (err) {
      // BKL-1054: tenta fallback de localStorage primeiro
      const fallback = this.loadFallback();
      if (fallback !== null) {
        this.cachedFlags = fallback;
        return fallback;
      }
      // BKL-258: fail-open — retorna defaults em vez de propagar o erro
      if (this.failOpen) {
        this.cachedFlags = { ...this.defaults };
        return this.cachedFlags;
      }
      throw err;
    }
  }

  // BKL-1054: salva flags no localStorage com timestamp para TTL
  private saveFallback(flags: FlagsMap): void {
    try {
      if (typeof localStorage !== "undefined") {
        localStorage.setItem(
          FALLBACK_STORAGE_KEY,
          JSON.stringify({ flags, savedAt: Date.now() })
        );
      }
    } catch {
      // localStorage indisponível (SSR, cookies bloqueados) — ignora
    }
  }

  // BKL-1054: carrega flags do localStorage se dentro do TTL
  private loadFallback(): FlagsMap | null {
    try {
      if (typeof localStorage === "undefined") return null;
      const raw = localStorage.getItem(FALLBACK_STORAGE_KEY);
      if (!raw) return null;
      const parsed = JSON.parse(raw) as { flags: FlagsMap; savedAt: number };
      if (Date.now() - parsed.savedAt > FALLBACK_TTL_MS) {
        localStorage.removeItem(FALLBACK_STORAGE_KEY);
        return null;
      }
      return parsed.flags;
    } catch {
      return null;
    }
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
   * BKL-258: se failOpen=true, retorna o default definido no init (ou false)
   * quando o backend estiver indisponível.
   */
  async isEnabled(key: string, defaultValue?: boolean): Promise<boolean> {
    validateKey(key);
    if (this.cachedFlags === null) {
      await this.getFlags();
    }
    if (this.cachedFlags === null) {
      // Nenhum cache disponível mesmo após getFlags — retorna default
      return defaultValue ?? this.defaults[key] ?? false;
    }
    // Se a flag não existe no mapa, usa default do caller ou do init
    if (!(key in this.cachedFlags)) {
      return defaultValue ?? this.defaults[key] ?? false;
    }
    return this.cachedFlags[key] === true;
  }

  /**
   * BKL-258: retorna o default configurado para uma flag específica.
   * Útil para leitura síncrona antes do carregamento das flags.
   */
  getDefault(key: string): boolean {
    return this.defaults[key] ?? false;
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

// ─── BKL-258: Singleton com API pulse.init() ──────────────────────────────

/**
 * Singleton de conveniência para uso sem instanciação explícita.
 *
 * @example
 * import pulse from '@iit/pulse-sdk'
 *
 * pulse.init({
 *   apiKey: 'proj_xxx',
 *   baseURL: 'https://seu-pulse.exemplo.com',
 *   failOpen: true,
 *   defaults: { 'nova-feature': false, 'dark-mode': true },
 * })
 *
 * const enabled = await pulse.isEnabled('nova-feature')
 */
class PulseSingleton {
  private client: PulseClient | null = null;

  /**
   * Inicializa o SDK com as opções fornecidas.
   * BKL-258: failOpen=true e timeout=2000ms são os defaults.
   */
  init(options: PulseInitOptions): this {
    const baseURL = options.baseURL;
    this.client = new PulseClient({
      baseURL,
      timeout: options.timeout ?? 2000,
      failOpen: options.failOpen ?? true,
      defaults: options.defaults ?? {},
      fetch: options.fetch,
    });
    return this;
  }

  private requireClient(): PulseClient {
    if (!this.client) {
      throw new PulseError("pulse.init() must be called before using the SDK");
    }
    return this.client;
  }

  async getFlags(): Promise<FlagsMap> {
    return this.requireClient().getFlags();
  }

  async isEnabled(key: string, defaultValue?: boolean): Promise<boolean> {
    if (!this.client) {
      // BKL-258: fail-open antes mesmo do init
      return defaultValue ?? false;
    }
    return this.client.isEnabled(key, defaultValue);
  }

  getDefault(key: string): boolean {
    return this.client?.getDefault(key) ?? false;
  }

  async getFlag(key: string): Promise<FeatureFlag | null> {
    return this.requireClient().getFlag(key);
  }

  async trackEvent(event: TrackEventRequest): Promise<TrackEventResponse> {
    return this.requireClient().trackEvent(event);
  }

  queue(event: TrackEventRequest): void {
    this.requireClient().queue(event);
  }

  flush(): void {
    this.requireClient().flush();
  }

  invalidateCache(): void {
    this.requireClient().invalidateCache();
  }
}

export const pulse = new PulseSingleton();
