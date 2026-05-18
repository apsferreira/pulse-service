/**
 * @pulse-iit/node — SDK Node.js server-side para pulse-service
 * Feature Flags + Event Tracking com fallback e cache local.
 *
 * Uso mínimo — feature flags:
 *   const pulse = new PulseClient({ apiUrl: 'https://pulse-api...', serviceToken: 'svc_xxx' })
 *   const enabled = await pulse.isEnabled('minha-flag', false)
 *   pulse.track('page_view', { path: '/home' }, 'user_123')
 *   await pulse.shutdown()
 */

// --- Types ---

export interface PulseClientOptions {
  /** Base URL do pulse-service (ex: "https://pulse-api.institutoitinerante.com.br") */
  apiUrl: string;
  /**
   * Service token — necessário para avaliação de flags (X-Service-Token).
   * Opcional se usar apenas tracking de eventos públicos.
   */
  serviceToken?: string;
  /** ID do projeto para tracking de eventos */
  projectId?: string;
  /**
   * TTL do cache local de flags em ms (default: 30000 — 30s).
   * Flags são recarregadas em background quando o TTL expira.
   * 0 = desabilita cache (não recomendado para produção).
   */
  flagCacheTTL?: number;
  /**
   * Valores default por flag — usados quando o Pulse está indisponível
   * e não há cache local. Garante comportamento determinístico offline.
   * Ex: { 'nova-feature': false, 'modo-manutencao': false }
   */
  defaults?: Record<string, boolean>;
  /** Timeout por requisição HTTP em ms (default: 2000 para flags, 10000 para eventos) */
  flagTimeout?: number;
  /** Intervalo de flush do buffer de eventos em ms (default: 5000) */
  flushInterval?: number;
  /** Tamanho máximo do buffer de eventos antes de flush forçado (default: 50) */
  maxBatchSize?: number;
  /** Custom fetch — útil para testes ou Node < 18 */
  fetch?: typeof globalThis.fetch;
}

export interface FlagsMap {
  [key: string]: boolean;
}

export interface TrackEvent {
  type: 'track';
  event: string;
  userId?: string;
  properties?: Record<string, unknown>;
  projectId: string;
  timestamp: string;
}

export interface IdentifyEvent {
  type: 'identify';
  userId: string;
  traits?: Record<string, unknown>;
  projectId: string;
  timestamp: string;
}

export type PulseEvent = TrackEvent | IdentifyEvent;

export interface BatchPayload {
  projectId: string;
  events: PulseEvent[];
}

// --- Helpers ---

function nowISO(): string {
  return new Date().toISOString();
}

// --- Client ---

export class PulseClient {
  private readonly apiUrl: string;
  private readonly serviceToken: string;
  private readonly projectId: string;
  private readonly defaults: Record<string, boolean>;
  private readonly flagCacheTTL: number;
  private readonly flagTimeout: number;
  private readonly flushInterval: number;
  private readonly maxBatchSize: number;
  private readonly fetchFn: typeof globalThis.fetch;

  // Cache local de flags
  private flagCache: FlagsMap | null = null;
  private flagCacheTime = 0;
  private flagFetchPromise: Promise<FlagsMap> | null = null;

  // Buffer de eventos
  private buffer: PulseEvent[] = [];
  private timer: ReturnType<typeof setInterval> | null = null;
  private flushing = false;
  private shutdown_ = false;

  constructor(options: PulseClientOptions) {
    if (!options.apiUrl) throw new Error('PulseClient: apiUrl é obrigatório');

    this.apiUrl = options.apiUrl.replace(/\/$/, '');
    this.serviceToken = options.serviceToken ?? '';
    this.projectId = options.projectId ?? '';
    this.defaults = options.defaults ?? {};
    this.flagCacheTTL = options.flagCacheTTL ?? 30_000;
    this.flagTimeout = options.flagTimeout ?? 2_000;
    this.flushInterval = options.flushInterval ?? 5_000;
    this.maxBatchSize = options.maxBatchSize ?? 50;
    this.fetchFn = options.fetch ?? globalThis.fetch;

    this.startTimer();
  }

  // ─── Feature Flags ──────────────────────────────────────────────────────────

  /**
   * Retorna se uma flag está ativa.
   * Fail-safe: retorna `defaultValue` (ou o default configurado no init) se:
   *   - Pulse estiver indisponível
   *   - Timeout expirar
   *   - Flag não existir
   * Nunca lança erro — seguro para usar em hot paths.
   */
  async isEnabled(flagKey: string, defaultValue?: boolean): Promise<boolean> {
    const fallback = defaultValue ?? this.defaults[flagKey] ?? false;
    try {
      const flags = await this.getFlags();
      if (!(flagKey in flags)) return fallback;
      return flags[flagKey] === true;
    } catch {
      return fallback;
    }
  }

  /**
   * Retorna o mapa completo de flags { key: enabled }.
   * Usa cache local (TTL configurável). Em caso de falha, retorna cache
   * expirado se disponível, senão retorna o mapa de defaults.
   * Múltiplas chamadas simultâneas compartilham um único fetch em andamento (dedup).
   */
  async getFlags(): Promise<FlagsMap> {
    const now = Date.now();
    const cacheValid = this.flagCache !== null &&
      (this.flagCacheTTL === 0 || now - this.flagCacheTime < this.flagCacheTTL);

    if (cacheValid) return this.flagCache!;

    // Dedup: se já há um fetch em andamento, aguarda o mesmo
    if (this.flagFetchPromise !== null) return this.flagFetchPromise;

    this.flagFetchPromise = this.fetchFlags().finally(() => {
      this.flagFetchPromise = null;
    });

    return this.flagFetchPromise;
  }

  private async fetchFlags(): Promise<FlagsMap> {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), this.flagTimeout);

    try {
      const headers: Record<string, string> = { Accept: 'application/json' };
      if (this.serviceToken) headers['X-Service-Token'] = this.serviceToken;

      const resp = await this.fetchFn(`${this.apiUrl}/flags`, {
        method: 'GET',
        headers,
        signal: controller.signal,
      });

      if (!resp.ok) throw new Error(`pulse: status ${resp.status}`);

      const body = await resp.json() as { flags?: FlagsMap };
      const flags = body.flags ?? {};

      this.flagCache = flags;
      this.flagCacheTime = Date.now();
      return flags;
    } catch {
      // Fail-safe: retorna cache expirado se houver, senão defaults
      if (this.flagCache !== null) return this.flagCache;
      return { ...this.defaults };
    } finally {
      clearTimeout(timeoutId);
    }
  }

  /**
   * Invalida o cache de flags, forçando reload no próximo isEnabled/getFlags.
   */
  invalidateFlagCache(): void {
    this.flagCache = null;
    this.flagCacheTime = 0;
  }

  // ─── Event Tracking ─────────────────────────────────────────────────────────

  /**
   * Enfileira um evento de tracking.
   * Fire-and-forget — não bloqueia o caller.
   */
  track(event: string, properties?: Record<string, unknown>, userId?: string): void {
    if (this.shutdown_) return;
    if (!this.projectId) return;

    this.buffer.push({
      type: 'track',
      event,
      userId,
      properties,
      projectId: this.projectId,
      timestamp: nowISO(),
    });

    if (this.buffer.length >= this.maxBatchSize) void this.flush();
  }

  /**
   * Enfileira um evento de identificação de usuário.
   */
  identify(userId: string, traits?: Record<string, unknown>): void {
    if (this.shutdown_) return;
    if (!this.projectId) return;

    this.buffer.push({
      type: 'identify',
      userId,
      traits,
      projectId: this.projectId,
      timestamp: nowISO(),
    });

    if (this.buffer.length >= this.maxBatchSize) void this.flush();
  }

  /**
   * Envia todos os eventos em buffer imediatamente.
   * Chamadas concorrentes são serializadas.
   */
  async flush(): Promise<void> {
    if (this.flushing || this.buffer.length === 0) return;

    this.flushing = true;
    const batch = this.buffer.splice(0, this.buffer.length);

    try {
      await this.sendBatch(batch);
    } catch {
      this.buffer.unshift(...batch);
    } finally {
      this.flushing = false;
    }
  }

  /**
   * Flush final + para o timer. Chamar no SIGTERM/SIGINT do processo.
   */
  async shutdown(): Promise<void> {
    this.shutdown_ = true;
    this.stopTimer();
    await this.flush();
  }

  // ─── Internal ───────────────────────────────────────────────────────────────

  private startTimer(): void {
    if (this.flushInterval <= 0 || !this.projectId) return;
    this.timer = setInterval(() => { void this.flush() }, this.flushInterval);
    if (this.timer.unref) this.timer.unref();
  }

  private stopTimer(): void {
    if (this.timer !== null) {
      clearInterval(this.timer);
      this.timer = null;
    }
  }

  private async sendBatch(events: PulseEvent[]): Promise<void> {
    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), 10_000);
    try {
      const resp = await this.fetchFn(`${this.apiUrl}/api/v1/events/batch`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ projectId: this.projectId, events }),
        signal: controller.signal,
      });
      if (!resp.ok) throw new Error(`pulse: batch status ${resp.status}`);
    } finally {
      clearTimeout(timeout);
    }
  }
}

export default PulseClient;
