/**
 * BKL-420: Pulse — Session Replay (gravação de sessão do usuário)
 * Grava sessões anonimizadas usando uma implementação leve de captura de DOM mutations.
 * LGPD compliant: não captura dados pessoais (inputs de senha/email são mascarados).
 * Opt-out obrigatório: respeitado via localStorage 'pulse_replay_optout'.
 *
 * Uso:
 *   import { SessionReplay } from "@iit/pulse-sdk/session-replay";
 *   SessionReplay.start({ apiKey: "...", projectId: "...", userId: "..." });
 *   SessionReplay.stop();
 */

export interface SessionReplayOptions {
  apiKey: string;
  projectId: string;
  userId?: string;
  baseURL?: string;
  /** Seletores CSS de elementos que devem ser mascarados (ex: inputs, .confidential) */
  maskSelectors?: string[];
  /** Tamanho máximo do batch de eventos antes de flush (default: 50) */
  batchSize?: number;
  /** Intervalo de flush em ms (default: 5000) */
  flushInterval?: number;
  /** Habilita captura de cliques com posição (default: true) */
  captureClicks?: boolean;
  /** Habilita captura de scroll (default: false — reduz volume) */
  captureScroll?: boolean;
}

interface ReplayEvent {
  type: 'dom_mutation' | 'click' | 'scroll' | 'viewport' | 'url_change';
  timestamp: number;
  data: Record<string, unknown>;
}

const DEFAULT_MASK_SELECTORS = [
  'input[type="password"]',
  'input[type="email"]',
  'input[name*="card"]',
  'input[name*="cpf"]',
  'input[name*="senha"]',
  '[data-pulse-mask]',
];

let sessionId: string | null = null;
let eventBuffer: ReplayEvent[] = [];
let flushTimer: ReturnType<typeof setInterval> | null = null;
let observer: MutationObserver | null = null;
let options: SessionReplayOptions | null = null;
let isStarted = false;

/**
 * Verifica se o usuário optou por não ser rastreado.
 */
export function hasOptedOut(): boolean {
  if (typeof localStorage === 'undefined') return false;
  return localStorage.getItem('pulse_replay_optout') === 'true';
}

/**
 * Registra opt-out do usuário para Session Replay (LGPD Art. 18).
 */
export function optOut(): void {
  if (typeof localStorage !== 'undefined') {
    localStorage.setItem('pulse_replay_optout', 'true');
  }
  stop();
}

/**
 * Remove opt-out do usuário.
 */
export function optIn(): void {
  if (typeof localStorage !== 'undefined') {
    localStorage.removeItem('pulse_replay_optout');
  }
}

/**
 * Inicia a gravação de sessão.
 */
export function start(opts: SessionReplayOptions): void {
  if (typeof document === 'undefined') return; // SSR guard
  if (hasOptedOut()) return;
  if (isStarted) stop();

  options = opts;
  sessionId = `replay_${opts.projectId}_${Date.now()}_${Math.random().toString(36).slice(2, 9)}`;
  isStarted = true;

  const batchSize = opts.batchSize ?? 50;
  const flushIntervalMs = opts.flushInterval ?? 5000;

  // Captura mudanças no DOM
  observer = new MutationObserver((mutations) => {
    for (const mutation of mutations) {
      const maskedSelectors = [...DEFAULT_MASK_SELECTORS, ...(opts.maskSelectors ?? [])];
      // Ignora nós de elementos mascarados
      if (mutation.target instanceof Element) {
        const isMasked = maskedSelectors.some(sel => {
          try { return (mutation.target as Element).closest(sel) !== null; } catch { return false; }
        });
        if (isMasked) continue;
      }

      addEvent({
        type: 'dom_mutation',
        timestamp: Date.now(),
        data: {
          type: mutation.type,
          target: mutation.target instanceof Element ? mutation.target.tagName.toLowerCase() : 'text',
          added: mutation.addedNodes.length,
          removed: mutation.removedNodes.length,
        },
      });
    }

    if (eventBuffer.length >= batchSize) flush();
  });

  observer.observe(document.body, {
    childList: true,
    subtree: true,
    attributes: true,
    characterData: false,
  });

  // Captura cliques
  if (opts.captureClicks !== false) {
    document.addEventListener('click', handleClick, { passive: true });
  }

  // Captura scroll
  if (opts.captureScroll === true) {
    document.addEventListener('scroll', handleScroll, { passive: true });
  }

  // Captura mudanças de URL (SPA navigation)
  window.addEventListener('popstate', handleUrlChange);

  // Captura viewport inicial
  addEvent({
    type: 'viewport',
    timestamp: Date.now(),
    data: { width: window.innerWidth, height: window.innerHeight, url: window.location.href },
  });

  // Flush periódico
  flushTimer = setInterval(flush, flushIntervalMs);
}

/**
 * Para a gravação de sessão e faz flush final.
 */
export function stop(): void {
  if (!isStarted) return;
  isStarted = false;

  observer?.disconnect();
  observer = null;

  document.removeEventListener('click', handleClick);
  document.removeEventListener('scroll', handleScroll);
  window.removeEventListener('popstate', handleUrlChange);

  if (flushTimer) {
    clearInterval(flushTimer);
    flushTimer = null;
  }

  flush(); // flush final
  sessionId = null;
  eventBuffer = [];
}

function handleClick(event: MouseEvent): void {
  const target = event.target as Element | null;
  if (!target) return;

  const maskedSelectors = [...DEFAULT_MASK_SELECTORS, ...(options?.maskSelectors ?? [])];
  const isMasked = maskedSelectors.some(sel => {
    try { return target.closest(sel) !== null; } catch { return false; }
  });

  addEvent({
    type: 'click',
    timestamp: Date.now(),
    data: {
      x: event.clientX,
      y: event.clientY,
      tag: isMasked ? '[masked]' : target.tagName.toLowerCase(),
      id: isMasked ? '[masked]' : (target.id || undefined),
      class: isMasked ? '[masked]' : (target.className?.toString().slice(0, 50) || undefined),
    },
  });
}

function handleScroll(): void {
  addEvent({
    type: 'scroll',
    timestamp: Date.now(),
    data: { x: window.scrollX, y: window.scrollY },
  });
}

function handleUrlChange(): void {
  addEvent({
    type: 'url_change',
    timestamp: Date.now(),
    data: { url: window.location.href },
  });
}

function addEvent(event: ReplayEvent): void {
  if (!isStarted || hasOptedOut()) return;
  eventBuffer.push(event);
}

async function flush(): Promise<void> {
  if (!options || !sessionId || eventBuffer.length === 0) return;

  const events = [...eventBuffer];
  eventBuffer = [];

  const baseURL = options.baseURL ?? '';
  const payload = {
    session_id: sessionId,
    project_id: options.projectId,
    user_id: options.userId ?? 'anonymous',
    events,
    page_url: window.location.href,
    user_agent: navigator.userAgent.slice(0, 200),
  };

  try {
    await fetch(`${baseURL}/api/v1/replay/events`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${options.apiKey}`,
      },
      body: JSON.stringify(payload),
      keepalive: true,
    });
  } catch {
    // Silent fail — session replay is non-critical
  }
}

export const SessionReplay = { start, stop, optOut, optIn, hasOptedOut };
export default SessionReplay;
