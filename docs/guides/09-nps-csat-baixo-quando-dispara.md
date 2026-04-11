---
title: "NPS/CSAT baixo — quando dispara"
description: "Como o widget de NPS/CSAT funciona, quais thresholds disparam alertas, como o health score é calculado e quais ações tomar quando o score cai."
---

# NPS/CSAT baixo — quando dispara

O Pulse inclui um módulo de Voice of Customer (VoC) que coleta NPS e CSAT diretamente no app, calcula um health score por tenant/usuário e dispara alertas quando os scores caem abaixo dos thresholds configurados.

---

## Como o widget funciona

O widget é exibido em overlay após eventos-gatilho configurados no dashboard. Você define:

- **Tipo de pesquisa:** NPS (0–10) ou CSAT (1–5)
- **Evento-gatilho:** ex: `checkout_completed`, `onboarding_finished`, após N dias de uso
- **Frequência:** mínimo de dias entre exibições para o mesmo usuário (padrão: 90 dias para NPS, 30 dias para CSAT)
- **Delay:** segundos após o evento antes de exibir (padrão: 3s)

### Instalação do widget

```typescript
// Inicializar com suporte a NPS/CSAT
const pulse = new PulseSDK({
  apiUrl: 'https://pulse-api.institutoitinerante.com.br',
  product: 'minha-app',
  serviceToken: process.env.PULSE_SERVICE_TOKEN!,
  userId: currentUser.id,
  voc: {
    enabled: true,
    containerId: 'pulse-voc-widget',  // div onde o widget será montado
  },
})

await pulse.getFlags()

// Disparar pesquisa manualmente após evento de conversão
await pulse.triggerSurvey('checkout_completed')
```

```html
<!-- Adicionar no index.html ou layout raiz -->
<div id="pulse-voc-widget"></div>
```

O widget é renderizado automaticamente quando as condições de frequência e gatilho são atendidas. Se o usuário já respondeu nos últimos N dias, nada é exibido.

---

## Thresholds de alerta

### NPS (Net Promoter Score — escala 0 a 10)

| Score | Classificação | Alerta |
|-------|---------------|--------|
| 9 – 10 | Promotor | Nenhum |
| 7 – 8 | Neutro | Nenhum |
| 0 – 6 | Detrator | Alerta disparado |

**Threshold padrão:** score < 7 = detrator = alerta.

Você pode ajustar em **Settings → Alertas → NPS Threshold**.

### CSAT (Customer Satisfaction Score — escala 1 a 5)

| Score | Classificação | Alerta |
|-------|---------------|--------|
| 4 – 5 | Satisfeito | Nenhum |
| 3 | Neutro | Nenhum |
| 1 – 2 | Insatisfeito | Alerta disparado |

**Threshold padrão:** score ≤ 2 = insatisfeito = alerta.

---

## Como o health score é calculado

O health score é um índice de 0 a 100 calculado por tenant/usuário, combinando múltiplos sinais:

```
health_score = (
  nps_score         * 0.35 +   // NPS normalizado (0-10 → 0-100)
  csat_score        * 0.25 +   // CSAT normalizado (1-5 → 0-100)
  engagement_score  * 0.25 +   // frequência de uso nos últimos 30 dias
  conversion_score  * 0.15     // eventos de conversão nos últimos 30 dias
)
```

**Engagement score:** baseado em `page_viewed`, `feature_used` e outros eventos de atividade. Um tenant que não faz login há 14 dias tem engagement_score próximo de zero.

**Conversion score:** baseado em eventos marcados como conversão no dashboard (ex: `checkout_completed`, `report_generated`).

### Consultar health score via API

```bash
# Health score de um tenant específico
curl "https://pulse-api.institutoitinerante.com.br/api/v1/health/tenant-456?product=minha-app" \
  -H "X-Service-Token: $SERVICE_TOKEN"
```

```json
{
  "tenant_id": "tenant-456",
  "health_score": 62,
  "trend": "declining",
  "components": {
    "nps_score": 55,
    "csat_score": 80,
    "engagement_score": 45,
    "conversion_score": 70
  },
  "last_nps": {
    "score": 5,
    "submitted_at": "2026-04-08T14:00:00Z",
    "comment": "O relatório demora muito para carregar"
  },
  "last_active": "2026-04-09T09:30:00Z",
  "alerts": [
    {
      "type": "nps_detractor",
      "triggered_at": "2026-04-08T14:00:00Z",
      "score": 5
    }
  ]
}
```

### Listar tenants em risco

```bash
# Tenants com health score abaixo de 60 (em risco de churn)
curl "https://pulse-api.institutoitinerante.com.br/api/v1/health?product=minha-app&max_score=60&limit=20" \
  -H "X-Service-Token: $SERVICE_TOKEN"
```

---

## Quando os alertas disparam

Os alertas são enviados por:

1. **Dashboard:** notificação em tempo real no painel de health
2. **Email:** para os membros do projeto configurados em **Settings → Alertas → Emails**
3. **Webhook** (plano Pro): POST para a URL configurada

### Payload do webhook de alerta

```json
{
  "event": "nps_detractor",
  "product": "minha-app",
  "tenant_id": "tenant-456",
  "user_id": "user-123",
  "score": 5,
  "comment": "O relatório demora muito para carregar",
  "health_score": 62,
  "health_trend": "declining",
  "triggered_at": "2026-04-08T14:00:00Z"
}
```

Configura o webhook em **Settings → Alertas → Webhook URL**.

---

## Ações recomendadas quando o alerta dispara

### NPS detrator (score < 7)

| Urgência | Condição | Ação |
|----------|----------|------|
| Alta | Score 0–4 + comment preenchido | CS contata o cliente em até 24h |
| Média | Score 5–6 | CS agenda check-in na semana |
| Baixa | Score 5–6 + cliente novo (<30 dias) | Email automático de onboarding avançado |

```typescript
// Exemplo: webhook handler no customer-service
app.post('/webhooks/pulse-alert', async (req, res) => {
  const { event, tenant_id, score, comment } = req.body

  if (event === 'nps_detractor' && score <= 4) {
    await criarTicketCS({
      tenant_id,
      prioridade: 'alta',
      descricao: `NPS detrator: ${score}/10. Comentário: "${comment || 'sem comentário'}"`,
      sla_horas: 24,
    })
  }

  res.status(200).json({ received: true })
})
```

### CSAT insatisfeito (score ≤ 2)

1. Verificar no audit log se o usuário teve erros recentes (HTTP 4xx/5xx correlacionados ao evento-gatilho).
2. Abrir ticket de bug se o problema for técnico.
3. Se for UX, criar tarefa no backlog com prioridade baseada na frequência do problema.

### Health score em queda (trend: declining)

```bash
# Ver histórico de health score dos últimos 30 dias
curl "https://pulse-api.institutoitinerante.com.br/api/v1/health/tenant-456/history?product=minha-app&days=30" \
  -H "X-Service-Token: $SERVICE_TOKEN"
```

Um tenant com `trend: declining` por mais de 7 dias e `health_score < 50` está em risco alto de churn. Acione o fluxo de resgate imediatamente.

---

## Configurar thresholds personalizados

```bash
curl -X PATCH https://pulse-api.institutoitinerante.com.br/api/v1/settings/alerts \
  -H "X-Service-Token: $SERVICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "nps_detractor_threshold": 7,
    "csat_unsatisfied_threshold": 3,
    "health_score_at_risk": 60,
    "health_score_critical": 40,
    "alert_channels": ["email", "webhook"]
  }'
```

---

## Próximos passos

- [07-exportar-dados-testes.md](./07-exportar-dados-testes.md) — exportar dados de NPS/CSAT para análise em BI
- [04-ab-test-interpretar-p-value.md](./04-ab-test-interpretar-p-value.md) — testar variações de UX que afetam o CSAT
