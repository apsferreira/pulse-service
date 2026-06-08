/**
 * Runtime env vars helper — lê de window.__ENV__ (injetado via entrypoint.sh
 * em produção/k8s) com fallback para import.meta.env (build-time, dev local).
 *
 * BKL-1063 / #60 — Evita secrets no docker history usando injeção runtime.
 */
export function readEnv(key: string, fallback = ''): string {
  if (typeof window !== 'undefined' && (window as Record<string, unknown>).__ENV__?.[key as keyof typeof window.__ENV__]) {
    return (window as Record<string, unknown>).__ENV__[key as keyof typeof window.__ENV__] as string
  }
  try {
    return (import.meta.env as Record<string, unknown>)?.[key] as string ?? fallback
  } catch {
    return fallback
  }
}
