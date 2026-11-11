/**
 * Runtime env vars helper — lê de window.__ENV__ (injetado via entrypoint.sh
 * em produção/k8s) com fallback para import.meta.env (build-time, dev local).
 *
 * BKL-1063 / #60 — Evita secrets no docker history usando injeção runtime.
 */
export function readEnv(key: string, fallback = ''): string {
  if (typeof window !== 'undefined') {
    const runtimeEnv = (window as unknown as { __ENV__?: Record<string, string> }).__ENV__
    if (runtimeEnv?.[key]) {
      return runtimeEnv[key]
    }
  }
  try {
    return (import.meta.env as Record<string, unknown>)?.[key] as string ?? fallback
  } catch {
    return fallback
  }
}
