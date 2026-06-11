#!/bin/sh
set -e

# ── Runtime env injection ────────────────────────────────────────────────
# Gera /tmp/env-config.js com variáveis de ambiente do pod k8s,
# evitando build-args que persistem secrets no docker history.
# BKL-1063 / #60 — VITE_SERVICE_TOKEN vazado no docker history
# ──────────────────────────────────────────────────────────────────────────
cat > /tmp/env-config.js << EOF
window.__ENV__ = {
  VITE_SERVICE_TOKEN: "${VITE_SERVICE_TOKEN:-}",
  VITE_API_URL: "${VITE_API_URL:-}",
  VITE_AUTH_SERVICE_URL: "${VITE_AUTH_SERVICE_URL:-}",
};
EOF

exec "$@"
