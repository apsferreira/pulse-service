#!/usr/bin/env bash
# Falha o build se houver PII (email) em chamadas de log.
# LGPD Art. 46 — logs do K3s retidos sem mascaramento.
set -euo pipefail

# Procura por OwnerEmail / ContactEmail / CustomerEmail / adminEmail
# dentro de chamadas de log.Printf|Println|Print|Fatal*|Panic* no diretório internal/.
matches=$(grep -rnE 'log\.(Printf|Println|Print|Fatal|Fatalf|Panic|Panicf)\(.*\b(OwnerEmail|ContactEmail|CustomerEmail|adminEmail)\b' \
  --include='*.go' internal/ 2>/dev/null || true)

if [ -n "$matches" ]; then
  echo "PII detectada em logs (LGPD violation):"
  echo "$matches"
  echo
  echo "Remova OwnerEmail/ContactEmail/CustomerEmail/adminEmail das chamadas de log."
  echo "Use IDs (project.ID, req.Slug) ou hash SHA-256."
  exit 1
fi

echo "OK: nenhum email de cliente em chamadas de log."
