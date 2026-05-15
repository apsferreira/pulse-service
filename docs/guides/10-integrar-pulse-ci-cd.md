---
title: "Integrar Pulse em CI/CD"
description: "Como usar feature flags como gates de deployment, fazer gradual rollout e rollback instantâneo, com exemplos práticos para GitHub Actions."
---

# Integrar Pulse em CI/CD

Feature flags não servem apenas para controlar UI — são uma ferramenta poderosa de deployment seguro. Com o Pulse integrado ao CI/CD, você pode deployar código novo com a flag desligada, ativar gradualmente e reverter em segundos sem precisar de rollback de código.

---

## O padrão: flag como gate de deployment

A ideia central é separar **deploy** de **release**:

```
Deploy  → código vai para produção, flag OFF → zero impacto para usuários
Release → flag ON (rollout 5%) → monitorar → aumentar gradualmente → 100%
Rollback → flag OFF imediatamente → zero downtime, zero redeploy
```

Isso elimina a pressão de um deploy "tudo ou nada" e permite detectar problemas com apenas uma fração dos usuários afetados.

---

## Criar ou atualizar flag via API

Você pode gerenciar flags diretamente da pipeline com `curl`:

```bash
# Criar flag (idempotente — retorna 200 se já existe)
curl -X POST https://pulse.institutoitinerante.com.br/api/v1/flags \
  -H "X-Service-Token: $PULSE_SERVICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "nova_feature_v2",
    "product": "minha-app",
    "description": "Nova feature deployada em 2026-04-11",
    "enabled": false,
    "rollout_percentage": 0
  }'

# Atualizar percentual de rollout
curl -X PATCH https://pulse.institutoitinerante.com.br/api/v1/flags/FLAG_ID \
  -H "X-Service-Token: $PULSE_SERVICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{ "rollout_percentage": 5 }'

# Rollback instantâneo (desativar flag)
curl -X PATCH https://pulse.institutoitinerante.com.br/api/v1/flags/FLAG_ID \
  -H "X-Service-Token: $PULSE_SERVICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{ "enabled": false, "rollout_percentage": 0 }'
```

---

## Workflow completo no GitHub Actions

### 1. Deploy seguro (flag OFF)

```yaml
# .github/workflows/deploy.yml
name: Deploy

on:
  push:
    branches: [main]

env:
  PULSE_API: https://pulse.institutoitinerante.com.br
  FLAG_NAME: nova_feature_v2
  PRODUCT: minha-app

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Build e push imagem Docker
        run: |
          docker build -t ghcr.io/${{ github.repository }}:${{ github.sha }} .
          docker push ghcr.io/${{ github.repository }}:${{ github.sha }}

      - name: Deploy no K3s (ArgoCD)
        run: |
          # Atualizar tag da imagem no manifesto e push para Git
          # ArgoCD detecta e sincroniza automaticamente
          sed -i "s|image: .*|image: ghcr.io/${{ github.repository }}:${{ github.sha }}|" k8s/deployment.yaml
          git config user.email "ci@institutoitinerante.com.br"
          git config user.name "CI Bot"
          git add k8s/deployment.yaml
          git commit -m "ci: atualiza imagem para ${{ github.sha }}"
          git push

      - name: Garantir flag desligada pós-deploy
        run: |
          # Buscar ID da flag
          FLAG_ID=$(curl -s "${{ env.PULSE_API }}/api/v1/flags?product=${{ env.PRODUCT }}" \
            -H "X-Service-Token: ${{ secrets.PULSE_SERVICE_TOKEN }}" | \
            python3 -c "import sys,json; flags=json.load(sys.stdin); \
            f=[x for x in flags if x['name']=='${{ env.FLAG_NAME }}']; \
            print(f[0]['id'] if f else '')")

          if [ -n "$FLAG_ID" ]; then
            curl -s -X PATCH "${{ env.PULSE_API }}/api/v1/flags/$FLAG_ID" \
              -H "X-Service-Token: ${{ secrets.PULSE_SERVICE_TOKEN }}" \
              -H "Content-Type: application/json" \
              -d '{ "enabled": false, "rollout_percentage": 0 }'
            echo "Flag $FLAG_ID garantida como desligada pós-deploy"
          fi
```

### 2. Gradual rollout (workflow manual)

```yaml
# .github/workflows/rollout.yml
name: Gradual Rollout

on:
  workflow_dispatch:
    inputs:
      flag_name:
        description: "Nome da flag"
        required: true
        default: "nova_feature_v2"
      rollout_percentage:
        description: "Percentual de rollout (0-100)"
        required: true
        default: "5"
      enabled:
        description: "Habilitar flag? (true/false)"
        required: true
        default: "true"

env:
  PULSE_API: https://pulse.institutoitinerante.com.br
  PRODUCT: minha-app

jobs:
  rollout:
    runs-on: ubuntu-latest
    steps:
      - name: Buscar ID da flag
        id: get-flag
        run: |
          FLAG_ID=$(curl -s "${{ env.PULSE_API }}/api/v1/flags?product=${{ env.PRODUCT }}" \
            -H "X-Service-Token: ${{ secrets.PULSE_SERVICE_TOKEN }}" | \
            python3 -c "
          import sys, json
          flags = json.load(sys.stdin)
          match = [f for f in flags if f['name'] == '${{ github.event.inputs.flag_name }}']
          print(match[0]['id'] if match else 'NOT_FOUND')
          ")
          echo "flag_id=$FLAG_ID" >> $GITHUB_OUTPUT

      - name: Falhar se flag não encontrada
        if: steps.get-flag.outputs.flag_id == 'NOT_FOUND'
        run: |
          echo "Flag '${{ github.event.inputs.flag_name }}' não encontrada no produto ${{ env.PRODUCT }}"
          exit 1

      - name: Atualizar rollout
        run: |
          curl -s -X PATCH "${{ env.PULSE_API }}/api/v1/flags/${{ steps.get-flag.outputs.flag_id }}" \
            -H "X-Service-Token: ${{ secrets.PULSE_SERVICE_TOKEN }}" \
            -H "Content-Type: application/json" \
            -d '{
              "enabled": ${{ github.event.inputs.enabled }},
              "rollout_percentage": ${{ github.event.inputs.rollout_percentage }}
            }'

      - name: Resumo do rollout
        run: |
          echo "### Rollout atualizado" >> $GITHUB_STEP_SUMMARY
          echo "- **Flag:** ${{ github.event.inputs.flag_name }}" >> $GITHUB_STEP_SUMMARY
          echo "- **Enabled:** ${{ github.event.inputs.enabled }}" >> $GITHUB_STEP_SUMMARY
          echo "- **Rollout:** ${{ github.event.inputs.rollout_percentage }}%" >> $GITHUB_STEP_SUMMARY
```

Execute via **Actions → Gradual Rollout → Run workflow**.

### 3. Rollback de emergência

```yaml
# .github/workflows/rollback-flag.yml
name: Rollback de Flag (Emergência)

on:
  workflow_dispatch:
    inputs:
      flag_name:
        description: "Nome da flag para desativar"
        required: true

env:
  PULSE_API: https://pulse.institutoitinerante.com.br
  PRODUCT: minha-app

jobs:
  rollback:
    runs-on: ubuntu-latest
    steps:
      - name: Desativar flag imediatamente
        run: |
          FLAG_ID=$(curl -s "${{ env.PULSE_API }}/api/v1/flags?product=${{ env.PRODUCT }}" \
            -H "X-Service-Token: ${{ secrets.PULSE_SERVICE_TOKEN }}" | \
            python3 -c "
          import sys, json
          flags = json.load(sys.stdin)
          match = [f for f in flags if f['name'] == '${{ github.event.inputs.flag_name }}']
          print(match[0]['id'] if match else 'NOT_FOUND')
          ")

          if [ "$FLAG_ID" = "NOT_FOUND" ]; then
            echo "Flag não encontrada!"
            exit 1
          fi

          HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
            -X PATCH "${{ env.PULSE_API }}/api/v1/flags/$FLAG_ID" \
            -H "X-Service-Token: ${{ secrets.PULSE_SERVICE_TOKEN }}" \
            -H "Content-Type: application/json" \
            -d '{ "enabled": false, "rollout_percentage": 0 }')

          if [ "$HTTP_STATUS" = "200" ]; then
            echo "Rollback concluido — flag '${{ github.event.inputs.flag_name }}' desligada."
            echo "Impacto para usuarios: 0% (imediato, sem redeploy necessario)"
          else
            echo "Falha no rollback! HTTP $HTTP_STATUS"
            exit 1
          fi
```

---

## Protocolo de rollout gradual recomendado

```
Dia 1  → rollout_percentage: 5   (canary — time interno + beta users)
Dia 2  → verificar métricas de erro e NPS
Dia 3  → rollout_percentage: 25  (se métricas OK)
Dia 5  → rollout_percentage: 50
Dia 7  → rollout_percentage: 100 (se nenhum incidente)
Dia 14 → remover flag do código no próximo sprint (limpeza de flag debt)
```

Automatize a progressão com o workflow manual acima ou com um script de verificação de métricas:

```bash
#!/bin/bash
# check-and-rollout.sh
# Aumenta rollout automaticamente se error rate < 1%

PRODUCT=minha-app
FLAG_NAME=nova_feature_v2
CURRENT_ROLLOUT=5
NEXT_ROLLOUT=25

ERROR_RATE=$(curl -s "$PULSE_API/api/v1/metrics?product=$PRODUCT&metric=error_rate&minutes=60" \
  -H "X-Service-Token: $SERVICE_TOKEN" | python3 -c "import sys,json; print(json.load(sys.stdin)['value'])")

if (( $(echo "$ERROR_RATE < 1.0" | bc -l) )); then
  echo "Error rate $ERROR_RATE% — aumentando rollout para $NEXT_ROLLOUT%"
  # Atualizar flag via API...
else
  echo "Error rate $ERROR_RATE% — ACIMA DO LIMITE. Mantendo $CURRENT_ROLLOUT%."
fi
```

---

## Configurar o secret no GitHub

No repositório GitHub: **Settings → Secrets and variables → Actions → New repository secret**

| Nome | Valor |
|------|-------|
| `PULSE_SERVICE_TOKEN` | Token de API gerado no dashboard Pulse |

---

## Próximos passos

- [03-debug-flag-nao-ativa.md](./03-debug-flag-nao-ativa.md) — debug quando a flag não reflete no app após rollout
- [08-self-hosted-como-funciona.md](./08-self-hosted-como-funciona.md) — deploy da instância self-hosted também pode usar este workflow
