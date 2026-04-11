---
title: "Como rodar A/B test e interpretar p-value"
description: "Passo a passo para criar um experimento no Pulse, definir variantes, rastrear conversão e interpretar os resultados com rigor estatístico."
---

# Como rodar A/B test e interpretar p-value

O Pulse suporta experimentos A/B com análise estatística embutida. Este guia cobre desde a criação do experimento até a decisão de qual variante subir para produção.

Para referência completa da API de A/B testing, veja também [ab-testing.md](./ab-testing.md).

---

## Conceitos básicos

| Termo | Definição |
|-------|-----------|
| **Variante de controle** | A versão atual (baseline) |
| **Variante de tratamento** | A nova versão sendo testada |
| **Métrica de conversão** | O evento que você quer maximizar (ex: `checkout_completed`) |
| **Sample size** | Número mínimo de usuários por variante para resultado confiável |
| **p-value** | Probabilidade de o resultado ser aleatório. Menor = mais confiante |
| **Significância estatística** | Convencionalmente p < 0.05 (95% de confiança) |

---

## Passo 1 — Criar o experimento

```bash
curl -X POST https://pulse-api.institutoitinerante.com.br/api/v1/experiments \
  -H "X-Service-Token: $SERVICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "botao_checkout_cor",
    "product": "minha-app",
    "description": "Teste se botão verde converte mais que azul",
    "metric": "checkout_completed",
    "variants": [
      { "name": "controle", "weight": 50 },
      { "name": "verde",    "weight": 50 }
    ],
    "min_sample_size": 500
  }'
```

**Resposta:**
```json
{
  "id": "exp_abc123",
  "name": "botao_checkout_cor",
  "status": "running",
  "variants": [
    { "id": "var_ctrl", "name": "controle", "weight": 50 },
    { "id": "var_verde","name": "verde",    "weight": 50 }
  ]
}
```

Os pesos (`weight`) são percentuais e devem somar 100. Para testar gradualmente uma variante nova, use `{ "controle": 80, "nova": 20 }`.

---

## Passo 2 — Expor a variante no código

```typescript
await pulse.getFlags()

// getVariant retorna o nome da variante para este usuário
const variante = pulse.getVariant('botao_checkout_cor')

// O mesmo usuário sempre recebe a mesma variante (hash consistente)
if (variante === 'verde') {
  return <BotaoCheckout cor="green" />
}
return <BotaoCheckout cor="blue" />  // controle
```

---

## Passo 3 — Rastrear o evento de conversão

Rastreie o evento exato que você definiu como `metric` no experimento. O Pulse correlaciona automaticamente o evento com a variante do usuário.

```typescript
async function finalizarCompra(orderId: string, valor: number) {
  await processarPagamento()

  // Rastrear conversão — Pulse associa ao experiment automaticamente
  await pulse.track('checkout_completed', {
    order_id: orderId,
    value: valor,
    currency: 'BRL',
  })
}
```

**Importante:** rastreie apenas conversões reais. Não dispare o evento de conversão sem a ação ter ocorrido — isso invalida o teste.

---

## Passo 4 — Aguardar o sample size mínimo

Não analise os resultados antes de atingir o `min_sample_size` por variante. Olhar cedo e parar o teste quando parece que uma variante está ganhando é o erro mais comum em A/B testing (problema de "peeking").

Verifique o progresso:

```bash
curl "https://pulse-api.institutoitinerante.com.br/api/v1/experiments/exp_abc123/results" \
  -H "X-Service-Token: $SERVICE_TOKEN"
```

```json
{
  "experiment": "botao_checkout_cor",
  "status": "running",
  "sample_size_reached": false,
  "variants": [
    {
      "name": "controle",
      "participants": 312,
      "conversions": 47,
      "conversion_rate": 0.1506
    },
    {
      "name": "verde",
      "participants": 298,
      "conversions": 61,
      "conversion_rate": 0.2047
    }
  ],
  "p_value": 0.031,
  "confidence": 0.969,
  "winner": null
}
```

O campo `winner` só é preenchido quando `sample_size_reached: true` e `p_value < 0.05`.

---

## Passo 5 — Interpretar o p-value

O p-value responde à pergunta: **"se não houvesse diferença real entre as variantes, qual a chance de observar esse resultado por acaso?"**

| p-value | Interpretação | Ação |
|---------|---------------|------|
| `> 0.10` | Sem evidência de diferença | Continue coletando dados |
| `0.05 – 0.10` | Tendência, mas inconclusivo | Continue coletando dados |
| `< 0.05` | Significativo (95% confiança) | Pode tomar decisão |
| `< 0.01` | Altamente significativo (99%) | Resultado robusto |

**No exemplo acima:** p-value = 0.031 < 0.05. A variante `verde` converteu 20.47% vs 15.06% do controle. Com 96.9% de confiança, a diferença não é ruído aleatório.

### O que o p-value NÃO diz

- Não mede o tamanho do efeito (uma diferença de 0.01% pode ser significativa com amostra enorme mas não vale o deploy).
- Não garante que o resultado vai se repetir em produção para todos os usuários.
- Não substitui julgamento de negócio — avalie também o impacto em receita e UX.

---

## Passo 6 — Tomar a decisão

### Se a variante de tratamento ganhou (p < 0.05):

```bash
# 1. Marcar o experimento como concluído
curl -X PATCH https://pulse-api.institutoitinerante.com.br/api/v1/experiments/exp_abc123 \
  -H "X-Service-Token: $SERVICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{ "status": "completed", "winner": "verde" }'

# 2. Fazer rollout da variante vencedora para 100%
curl -X PATCH https://pulse-api.institutoitinerante.com.br/api/v1/flags/FLAG_ID \
  -H "X-Service-Token: $SERVICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{ "rollout_percentage": 100 }'

# 3. Remover o código da variante perdedora no próximo sprint
```

### Se o controle ganhou ou resultado inconclusivo:

```bash
# Encerrar experimento sem vencedor
curl -X PATCH https://pulse-api.institutoitinerante.com.br/api/v1/experiments/exp_abc123 \
  -H "X-Service-Token: $SERVICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{ "status": "completed", "winner": "controle" }'

# Desativar a flag da nova variante
curl -X PATCH https://pulse-api.institutoitinerante.com.br/api/v1/flags/FLAG_ID \
  -H "X-Service-Token: $SERVICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{ "enabled": false }'
```

---

## Erros comuns

**Parar o teste cedo demais (peeking):** Resistir à tentação de olhar os resultados a cada hora. Defina uma data de análise com antecedência.

**Testar múltiplas métricas:** Se você mede 10 métricas, estatisticamente uma vai parecer significativa por acaso. Defina a métrica primária antes de iniciar o teste.

**Poluir o experimento:** Mudar o código da variante ou o peso do rollout após o início invalida os resultados.

**Amostra contaminada:** Usuários que viram as duas variantes (ex: limparam cookies) devem ser excluídos. O Pulse faz isso automaticamente via `userId` persistente.

---

## Próximos passos

- [07-exportar-dados-testes.md](./07-exportar-dados-testes.md) — exportar dados brutos para análise em BI/planilha
- [10-integrar-pulse-ci-cd.md](./10-integrar-pulse-ci-cd.md) — automatizar rollout da variante vencedora no CI/CD
