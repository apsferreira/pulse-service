# Pulse — SLA (Service Level Agreement)

> Documento público. Vigência: 2026-04-10 em diante. BKL-1069.

---

## Compromissos de Disponibilidade por Plano

| Plano | Uptime Garantido | Tempo de Resposta | Tempo de Resolução |
|-------|-----------------|------------------|-------------------|
| **Free** | Best effort (sem SLA) | Sem garantia | Sem garantia |
| **Starter** | 99,5% ao mês | 24h (dias úteis) | 48h |
| **Growth** | 99,5% ao mês | 24h (dias úteis) | 48h |
| **Pro** | 99,9% ao mês | 4h (7×24) | 8h |

### Cálculo de Downtime Permitido

| Plano | Downtime/mês permitido |
|-------|----------------------|
| Starter/Growth (99,5%) | ~3h 39min/mês |
| Pro (99,9%) | ~43min/mês |

---

## Janelas de Manutenção

- **Manutenções programadas:** comunicadas com 48h de antecedência via email + status page
- **Janela preferencial:** domingos 02:00–06:00 BRT
- **Janelas programadas não contam como downtime no SLA**

---

## Exclusões do SLA

Os seguintes eventos NÃO contam para cálculo de downtime:

1. Força maior: eventos climáticos, ataques DDoS em escala nacional, falhas de ISP primário
2. Manutenções programadas comunicadas com >48h de antecedência
3. Downtime causado por uso indevido do cliente (ex: eventos acima do limite do plano causando throttling intencional)
4. Erros de configuração do cliente (API key inválida, código do SDK incorreto)
5. Problemas na infraestrutura de terceiros fora do controle do Pulse (ex: Cloudflare, AWS, Asaas)

---

## Penalidades por Violação de SLA

### Starter e Growth
- Downtime acumulado > 3h39min no mês:
  - **1–4h extras de downtime:** crédito de 10% da mensalidade
  - **4–8h extras:** crédito de 25%
  - **>8h extras:** crédito de 50%

### Pro
- Downtime acumulado > 43min no mês:
  - **43min–2h de downtime:** crédito de 10% da mensalidade
  - **2–4h:** crédito de 25%
  - **>4h:** crédito de 50%

**Máximo de crédito:** 50% da mensalidade do mês afetado.
Créditos são aplicados na próxima fatura — não há reembolso em dinheiro.

---

## Como Solicitar Crédito

1. Abra ticket via email em suporte@institutoitinerante.com.br dentro de 7 dias após o incidente
2. Informe: datas/horários do downtime, projeto afetado, evidências (logs, screenshots)
3. O time avalia e aplica crédito em até 5 dias úteis após confirmação

---

## Monitoramento e Transparência

- **Status page:** pulse.institutoitinerante.com.br/status *(em construção — BKL-1068)*
- **Métricas públicas:** uptime dos últimos 90 dias disponível na status page
- **Incidentes:** comunicados via email para clientes afetados em até 15 minutos após detecção

---

## Contato para SLA

- **Email:** suporte@institutoitinerante.com.br
- **Pro (4h SLA):** Slack channel dedicado fornecido após contratação
- **Horário de suporte:** Starter/Growth: seg–sex 9h–18h BRT | Pro: 7×24

---

## Vigência e Alterações

Este SLA pode ser atualizado com aviso de 30 dias. Clientes ativos serão notificados por email.
Versão atual: **1.0 (2026-04-10)**.
