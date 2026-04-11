# Pulse — Planos de Preços

## Visão Geral

Pulse é um serviço de feature flags e event tracking para aplicações na escala do Instituto Itinerante. Oferecemos 4 planos: **Free**, **Starter**, **Growth** e **Pro**, calibrados para startups brasileiras e empresas em escala.

> **Atualização BKL-1070 (2026-04-10):** Repricing baseado em análise de custo operacional.
> Custo estimado por cliente: R$180/mês no homelab. Margem mínima de 40% para sustentabilidade.
> Novo pricing: Starter R$399 (era R$299), Growth R$999, Pro R$2.499.

---

## Planos

### Free
**Sem custo | Sem cartão de crédito**

Ideal para protótipos, MVPs e pequenas aplicações.

- **Eventos/mês:** até 10.000
- **Feature flags:** até 5
- **Projetos:** 1
- **Usuários:** 1
- **A/B testing:** não
- **Histórico de eventos:** 7 dias
- **SLA:** Best effort (sem garantia)
- **Suporte:** Community (Discord/GitHub Issues)

**Limite:** 1 projeto, 5 flags, 10k eventos/mês.

---

### Starter
**R$ 399/mês | Faturamento mensal**

Para produtos em fase de validação com primeiros usuários reais.

- **Eventos/mês:** até 500.000
- **Feature flags:** ilimitadas
- **Projetos:** até 3
- **Usuários:** até 3
- **A/B testing:** básico (2 variantes)
- **Histórico de eventos:** 30 dias
- **Rollout percentual:** sim
- **SLA:** 99.5% uptime | Resposta em 24h
- **Suporte:** Email (resposta < 24h)

**Upgrade trigger:** Quando usar >10k eventos/mês OU precisar de mais de 1 projeto.

---

### Growth
**R$ 999/mês | Faturamento mensal**

Para startups e equipes em crescimento acelerado.

- **Eventos/mês:** até 2.000.000
- **Feature flags:** ilimitadas
- **Projetos:** até 10
- **Usuários:** até 10
- **A/B testing:** avançado (múltiplas variantes + chi-squared)
- **Histórico de eventos:** 90 dias
- **Rollout percentual:** sim
- **Overrides:** por usuário/tenant
- **CSAT/NPS tracking:** sim
- **Audit log:** sim
- **Ambientes por projeto:** prod/staging/dev
- **Eventos por segundo:** ~0.77 (2M/mês ÷ 30 dias ÷ 86.4k segundos)
- **SLA:** 99.5% uptime | Resposta em 24h
- **Suporte:** Email (resposta < 24h)

**Upgrade trigger:** Quando usar >500k eventos/mês OU precisar de A/B testing avançado OU CSAT/NPS.

---

### Pro
**R$ 2.499/mês | Faturamento mensal**

Para empresas que precisam de confiabilidade enterprise e escala.

- **Eventos/mês:** ilimitados
- **Feature flags:** ilimitadas
- **Projetos:** ilimitados
- **Usuários:** ilimitados
- **A/B testing:** avançado + experimentos simultâneos ilimitados
- **Histórico de eventos:** 365 dias
- **CSAT/NPS tracking:** sim
- **Retention analysis:** sim
- **Webhooks customizados:** sim
- **Audit log:** sim (LGPD/compliance)
- **Ambientes por projeto:** ilimitados
- **API priority:** SLA 4h
- **SLA:** 99.9% uptime | Resposta em 4h
- **Suporte:** Email + Slack channel (resposta < 4h)

**Upgrade trigger:** Quando usar >2M eventos/mês OU precisar de >10 projetos OU dados históricos 1 ano OU SLA 99.9%.

---

## Pricing Justificativa

### Custo Operacional (BKL-1070)

Custo estimado por cliente ativo no homelab (homelab-infrastructure):
- PostgreSQL: ~R$ 30/cliente (armazenamento + CPU)
- Redis: ~R$ 20/cliente (memória)
- K3s (compute + networking): ~R$ 80/cliente
- Cloudflare: ~R$ 30/cliente
- Support & overhead: ~R$ 20/cliente
- **Total: ~R$ 180/cliente/mês**

Margem operacional por plano:
- Starter R$399: margem ~55% (R$219 de lucro bruto)
- Growth R$999: margem ~82% (R$819)
- Pro R$2.499: margem ~93% (R$2.319)

### Comparativo com Concorrentes

| Recurso | Pulse Free | Pulse Starter | Pulse Growth | Pulse Pro | PostHog | LaunchDarkly |
|---------|-----------|--------------|-------------|----------|---------|--------------|
| Feature flags | ✅ 5 | ✅ Ilimitado | ✅ Ilimitado | ✅ Ilimitado | ✅ Ilimitado | ✅ Ilimitado |
| Event tracking | ✅ 10k/mês | ✅ 500k/mês | ✅ 2M/mês | ✅ Ilimitado | ✅ Ilimitado | ❌ |
| A/B testing | ❌ | Básico | Avançado | Avançado | ✅ | ✅ |
| Preço (BRL) | Grátis | R$399 | R$999 | R$2.499 | ~R$500 | ~R$2.500 |
| **Posicionamento** | MVP/dev | Validação | Startup BR | Empresa BR | SaaS maduro | Enterprise |

---

## Modelo de Faturamento

- **Cobrança:** mensal, automática via Asaas
- **Ciclo:** 1º ao último dia do mês
- **Cancelamento:** a qualquer momento, sem penalidade
- **Upgrade:** imediato, diferença pro-rata
- **Downgrade:** no mês seguinte

### Exemplo: Upgrade Mid-month
Usuário em **Free** (R$ 0) e faz upgrade para **Starter** (R$ 399) no dia 15 de um mês com 30 dias.
- Custo pro-rata: 399 × (16 dias ÷ 30) = R$ 212,80
- Próxima cobrança: R$ 399 (ciclo cheio)

---

## Overage Pricing

Se cliente em **Starter** ultrapassar 500k eventos:
- **Overage:** R$ 0,20 por 1k eventos acima do limite
- **Exemplo:** 600k eventos = 399 + (100k ÷ 1k × 0,20) = 399 + 20 = R$ 419

Se cliente em **Growth** ultrapassar 2M eventos:
- **Overage:** R$ 0,10 por 1k eventos acima do limite

Clientes em **Pro** não têm overage (ilimitado).

---

## FAQ

**P: Posso usar Free indefinidamente?**
R: Sim. Sem limite de tempo. Quando ultrapassar 10k eventos/mês, sugerimos upgrade para Starter.

**P: E se eu exceder o limite de eventos?**
R: Você paga overage até o final do mês. O serviço não para — apenas cobra a diferença.

**P: Vocês oferecem desconto anual?**
R: Sim, 2 meses grátis no plano anual (Growth e Pro).

**P: Qual SLA tem o Free?**
R: Best effort (sem garantia formal). Starter e Growth têm 99.5%, Pro tem 99.9%.

**P: Existe plano Enterprise?**
R: Estamos mapeando demanda. Entre em contato para orçamento customizado.
