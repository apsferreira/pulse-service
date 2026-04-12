/**
 * @pulse-iit/node — SDK Node.js server-side para pulse-service
 * BKL-1003: event tracking com buffer e batch flush
 *
 * Uso minimo:
 *   const pulse = new PulseClient({ projectId: 'proj_xxx', apiUrl: 'https://pulse-api...' })
 *   pulse.track('page_view', { path: '/home' }, 'user_123')
 *   await pulse.shutdown()
 */

// --- Types ---

export interface PulseClientOptions {
  /** ID do projeto no pulse-service */
  projectId: string;
  /** Base URL do pulse-service (ex: "https://pulse-api.institutoitinerante.com.br") */
  apiUrl: string;
  /**
   * Intervalo de flush automatico em ms (default: 5000).
   * O buffer e enviado a cada `flushInterval` ms OU quando atinge `maxBatchSize`.
   */
  flushInterval?: number;
  /**
   * Numero maximo de eventos por batch antes de um flush forcado (default: 50).
   */
  maxBatchSize?: number;
  /**
   * Timeout por requisicao HTTP em ms (default: 10000).
   */
  requestTimeout?: number;
  /**
   * Custom fetch — util para testes ou ambientes sem fetch nativo (Node < 18).
   * Default: globalThis.fetch
   */
  fetch?: typeof globalThis.fetch;
}

export interface TrackEvent {
  type: "track";
  event: string;
  userId?: string;
  properties?: Record<string, unknown>;
  projectId: string;
  timestamp: string;
}

export interface IdentifyEvent {
  type: "identify";
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

export interface BatchResponse {
  received: number;
  errors?: string[];
}

// --- Helpers ---

function nowISO(): string {
  return new Date().toISOString();
}

// --- Client ---

export class PulseClient {
  private readonly projectId: string;
  private readonly apiUrl: string;
  private readonly flushInterval: number;
  private readonly maxBatchSize: number;
  private readonly requestTimeout: number;
  private readonly fetchFn: typeof globalThis.fetch;

  private buffer: PulseEvent[] = [];
  private timer: ReturnType<typeof setInterval> | null = null;
  private flushing = false;
  private shutdown_ = false;

  constructor(options: PulseClientOptions) {
    if (!options.projectId) throw new Error("PulseClient: projectId e obrigatorio");
    if (!options.apiUrl) throw new Error("PulseClient: apiUrl e obrigatorio");

    this.projectId = options.projectId;
    this.apiUrl = options.apiUrl.replace(/\/$/, "");
    this.flushInterval = options.flushInterval ?? 5_000;
    this.maxBatchSize = options.maxBatchSize ?? 50;
    this.requestTimeout = options.requestTimeout ?? 10_000;
    this.fetchFn = options.fetch ?? globalThis.fetch;

    this.startTimer();
  }

  /**
   * Registra um evento de tracking no buffer.
   * O buffer e enviado automaticamente a cada `flushInterval` ms
   * ou quando atinge `maxBatchSize` eventos.
   */
  track(
    event: string,
    properties?: Record<string, unknown>,
    userId?: string,
  ): void {
    if (this.shutdown_) {
      console.warn("[pulse] track() chamado apos shutdown() — evento descartado");
      return;
    }

    this.buffer.push({
      type: "track",
      event,
      userId,
      properties,
      projectId: this.projectId,
      timestamp: nowISO(),
    });

    if (this.buffer.length >= this.maxBatchSize) {
      void this.flush();
    }
  }

  /**
   * Registra um evento de identificacao de usuario no buffer.
   */
  identify(userId: string, traits?: Record<string, unknown>): void {
    if (this.shutdown_) {
      console.warn("[pulse] identify() chamado apos shutdown() — evento descartado");
      return;
    }

    this.buffer.push({
      type: "identify",
      userId,
      traits,
      projectId: this.projectId,
      timestamp: nowISO(),
    });

    if (this.buffer.length >= this.maxBatchSize) {
      void this.flush();
    }
  }

  /**
   * Envia todos os eventos em buffer para o pulse-service imediatamente.
   * Chamadas concorrentes sao serializadas — apenas uma flush roda por vez.
   */
  async flush(): Promise<void> {
    if (this.flushing || this.buffer.length === 0) return;

    this.flushing = true;
    const batch = this.buffer.splice(0, this.buffer.length);

    try {
      await this.sendBatch(batch);
    } catch (err) {
      // Re-enqueue para nao perder eventos — insere no inicio do buffer
      this.buffer.unshift(...batch);
      console.error("[pulse] flush falhou — eventos re-enfileirados:", err);
    } finally {
      this.flushing = false;
    }
  }

  /**
   * Faz flush final do buffer e para o timer automatico.
   * Deve ser chamado no encerramento do processo (SIGTERM/SIGINT).
   */
  async shutdown(): Promise<void> {
    this.shutdown_ = true;
    this.stopTimer();
    await this.flush();
  }

  // --- Private ---

  private startTimer(): void {
    if (this.flushInterval <= 0) return;

    this.timer = setInterval(() => {
      void this.flush();
    }, this.flushInterval);

    // Nao impede o processo de encerrar enquanto o timer esta ativo
    if (this.timer.unref) this.timer.unref();
  }

  private stopTimer(): void {
    if (this.timer !== null) {
      clearInterval(this.timer);
      this.timer = null;
    }
  }

  private async sendBatch(events: PulseEvent[]): Promise<void> {
    const payload: BatchPayload = {
      projectId: this.projectId,
      events,
    };

    const controller = new AbortController();
    const timeout = setTimeout(
      () => controller.abort(),
      this.requestTimeout,
    );

    try {
      const response = await this.fetchFn(
        `${this.apiUrl}/api/v1/events/batch`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(payload),
          signal: controller.signal,
        },
      );

      if (!response.ok) {
        const body = await response.text().catch(() => "(sem corpo)");
        throw new Error(
          `pulse-service retornou ${response.status}: ${body}`,
        );
      }
    } finally {
      clearTimeout(timeout);
    }
  }
}

export default PulseClient;
