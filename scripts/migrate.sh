#!/usr/bin/env bash
# scripts/migrate.sh — roda todas as migrations do Pulse em ordem
#
# Uso:
#   ./scripts/migrate.sh                          # usa DATABASE_URL do ambiente
#   DATABASE_URL=postgres://... ./scripts/migrate.sh
#   ./scripts/migrate.sh --dry-run               # lista arquivos sem executar
#
# Para self-hosted com docker compose:
#   docker exec -i pulse-postgres psql -U pulse -d pulse_db < migrations/001_create_feature_flags.sql
#   Ou passe DATABASE_URL apontando para o container exposto.

set -euo pipefail

MIGRATIONS_DIR="$(cd "$(dirname "$0")/../migrations" && pwd)"
DRY_RUN=false

for arg in "$@"; do
  case "$arg" in
    --dry-run) DRY_RUN=true ;;
    *) echo "Uso: $0 [--dry-run]" >&2; exit 1 ;;
  esac
done

if [[ -z "${DATABASE_URL:-}" ]]; then
  echo "Erro: DATABASE_URL não definida." >&2
  echo "Exemplo: DATABASE_URL=postgres://pulse:senha@localhost:5432/pulse_db ./scripts/migrate.sh" >&2
  exit 1
fi

files=$(find "$MIGRATIONS_DIR" -name "*.sql" | sort)

if [[ -z "$files" ]]; then
  echo "Nenhum arquivo .sql encontrado em $MIGRATIONS_DIR" >&2
  exit 1
fi

count=0
for f in $files; do
  name=$(basename "$f")
  if $DRY_RUN; then
    echo "[dry-run] $name"
  else
    echo "→ $name"
    psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f "$f"
    count=$((count + 1))
  fi
done

if ! $DRY_RUN; then
  echo ""
  echo "✓ $count migrations aplicadas com sucesso."
fi
