/**
 * BKL-1062: Badge "Powered by Pulse" — componente opt-in para PLG.
 * Injeta um badge discreto na aplicação do cliente linkando para a landing page do Pulse.
 *
 * Uso (opt-in — cliente deve chamar explicitamente):
 *   import { injectPoweredByBadge } from "@iit/pulse-sdk/badge";
 *   injectPoweredByBadge(); // injeta badge no canto inferior direito
 *   injectPoweredByBadge({ position: "bottom-left", theme: "dark" });
 */

export interface BadgeOptions {
  /** Posição do badge na tela (default: "bottom-right") */
  position?: "bottom-right" | "bottom-left" | "top-right" | "top-left";
  /** Tema do badge (default: "auto" — segue prefers-color-scheme) */
  theme?: "light" | "dark" | "auto";
  /** URL da landing page (default: https://pulse.institutoitinerante.com.br) */
  landingURL?: string;
  /** z-index do badge (default: 9999) */
  zIndex?: number;
}

const BADGE_ID = "__pulse_powered_by_badge__";
const DEFAULT_LANDING_URL = "https://pulse.institutoitinerante.com.br";

/**
 * Injeta o badge "Powered by Pulse" no DOM da aplicação.
 * Se já estiver injetado, não duplica.
 * Safe para SSR: verifica typeof document antes de executar.
 */
export function injectPoweredByBadge(options: BadgeOptions = {}): void {
  if (typeof document === "undefined") return;
  if (document.getElementById(BADGE_ID)) return; // já injetado

  const {
    position = "bottom-right",
    theme = "auto",
    landingURL = DEFAULT_LANDING_URL,
    zIndex = 9999,
  } = options;

  const positionStyles = {
    "bottom-right": "bottom: 16px; right: 16px;",
    "bottom-left": "bottom: 16px; left: 16px;",
    "top-right": "top: 16px; right: 16px;",
    "top-left": "top: 16px; left: 16px;",
  };

  // Estilos base (funciona em qualquer framework)
  const badge = document.createElement("a");
  badge.id = BADGE_ID;
  badge.href = landingURL;
  badge.target = "_blank";
  badge.rel = "noopener noreferrer";
  badge.setAttribute("aria-label", "Powered by Pulse — Feature Flags & Analytics");
  badge.setAttribute(
    "title",
    "Powered by Pulse — Feature Flags & Analytics by Instituto Itinerante"
  );

  const isDark =
    theme === "dark" ||
    (theme === "auto" &&
      typeof window !== "undefined" &&
      window.matchMedia?.("(prefers-color-scheme: dark)").matches);

  const bg = isDark ? "#1e1b4b" : "#f5f3ff";
  const textColor = isDark ? "#c4b5fd" : "#6d28d9";
  const borderColor = isDark ? "#4c1d95" : "#ddd6fe";
  const logoColor = isDark ? "#8b5cf6" : "#7c3aed";

  badge.style.cssText = `
    position: fixed;
    ${positionStyles[position]}
    z-index: ${zIndex};
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 6px 10px;
    background: ${bg};
    border: 1px solid ${borderColor};
    border-radius: 999px;
    text-decoration: none;
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
    font-size: 11px;
    font-weight: 500;
    color: ${textColor};
    letter-spacing: 0.01em;
    box-shadow: 0 1px 3px rgba(0,0,0,0.12);
    transition: opacity 0.2s ease;
    opacity: 0.85;
  `;

  // SVG logo Pulse (simplificado — ícone de atividade)
  const svg = document.createElementNS("http://www.w3.org/2000/svg", "svg");
  svg.setAttribute("width", "14");
  svg.setAttribute("height", "14");
  svg.setAttribute("viewBox", "0 0 24 24");
  svg.setAttribute("fill", "none");
  svg.setAttribute("stroke", logoColor);
  svg.setAttribute("stroke-width", "2.5");
  svg.setAttribute("stroke-linecap", "round");
  svg.setAttribute("stroke-linejoin", "round");
  const path = document.createElementNS("http://www.w3.org/2000/svg", "path");
  path.setAttribute("d", "M22 12h-4l-3 9L9 3l-3 9H2");
  svg.appendChild(path);

  const text = document.createTextNode("Powered by Pulse");

  badge.appendChild(svg);
  badge.appendChild(text);

  // Hover effect
  badge.addEventListener("mouseenter", () => {
    badge.style.opacity = "1";
  });
  badge.addEventListener("mouseleave", () => {
    badge.style.opacity = "0.85";
  });

  document.body.appendChild(badge);
}

/**
 * Remove o badge do DOM (caso necessário).
 */
export function removePoweredByBadge(): void {
  if (typeof document === "undefined") return;
  document.getElementById(BADGE_ID)?.remove();
}
