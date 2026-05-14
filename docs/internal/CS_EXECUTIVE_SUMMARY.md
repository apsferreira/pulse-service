# Pulse — Customer Success Strategy (Executive Summary)

**Prepared for:** Antonio Pedro Ferreira (Founder)  
**Date:** 10 de abril de 2026  
**Status:** Ready for decision

---

## The Problem

Pulse é uma ferramenta técnica complexa (feature flags, A/B testing, event tracking). Seus clientes são **Product Managers, Tech Leads, PMs** — personas altamente técnicas, baixa tolerância para onboarding ruim.

**Desafio:** IIT é uma startup com time pequeno. Não há orçamento para CSM full-time por cliente (LaunchDarkly modelo). **Precisamos estruturar CS que seja:**

1. ✅ Escalável (1 pessoa suporte resolve 90% dos problemas)
2. ✅ Proativo (detecta churn antes de cancelamento)
3. ✅ Automático (email, health scores, alerts rodam sozinhos)
4. ✅ Defensável vs concorrentes (diferenciador)

---

## O Que Faz o Pulse Diferente

LaunchDarkly, VWO e PostHog têm **suporte reativo e caro**:
- LaunchDarkly: email 48h, Slack só em Enterprise (USD 2k+/mês)
- VWO: implementação consultant incluída (caro para eles), suporte 24/7 chat só em Growth+
- PostHog: documentação boa, mas suporte é GitHub issues (lento)

**Pulse pode ser diferente:**
- Documentação super clara (70% das dúvidas resolvidas antes de ticket)
- Onboarding automático (email sequence + checklist)
- Health score do cliente (nossa ferramenta detecta churn deles!)
- Brio IA (L1 24/7 via WhatsApp, 60% resolve sem escalação)
- Feedback loop automático (tickets agrupados → backlog produto)

**Resultado:** Churn 3% vs 5-8% indústria = defensável, rentável.

---

## Estrutura Recomendada (3 Tiers)

### Tier 1 — MVP Support (0-30 dias) — BLOQUEADOR
**Objetivo:** Ir para market com suporte funcional.

| BKL | O Quê | Por Quê | Investimento |
|-----|-------|---------|--------------|
| **CS-001/006** | 10 Knowledge Base articles (KBAs) | 70% tickets são perguntas repetidas. KBAs resolvem antes do ticket chegar. | 6 dias |
| **CS-002** | Health score (0-100) por cliente | Detectar churn passivo (0 logins, eventos zerados) automático. | 2-3 dias |
| **CS-003** | Email sequence (6 emails) | Ativar 70% dos clientes em 30 dias (criar flags, enviar eventos). | 2 dias |
| **CS-004** | Brio integração (WhatsApp L1) | Suporte 24/7. 60% dos tickets resolvem em L1 sem escalação. | 3 dias |
| **CS-005** | Churn alerts | Quando health score < 30, email + Brio automático detecta risco. | 1 dia |

**Total Tier 1:** ~16-18 dias eng + 5 dias PM. **Viável em 4 semanas.**

**Resultado:** Pulse pronto para venda. Suporte robusto. Churn reduzido.

---

### Tier 2 — Diferenciadores (30-120 dias)
**Objetivo:** Proativo, feedback loop, community.

| BKL | O Quê | ROI |
|-----|-------|-----|
| CS-010 | Dashboard isolado por projeto | Pronto para SaaS real (clientes veem só dados próprios). |
| CS-012 | Support patterns detector | Tickets → backlog automático. Roadmap data-driven. |
| CS-013 | Churn prediction (ML) | 2-4 semanas antes de cancelamento. Oferta resgate proativa. |
| CS-011 | Community Discord | 30% menos suporte. Customer engagement. |
| CS-015/016 | ROI calc + upgrade recommender | Reduz fricção Free → Growth (aumenta conversion 2-3x). |
| CS-017 | Win-back campaign | Reactiva Free lapsed customers. |

**Total Tier 2:** ~35-40 dias eng. **Viável em 3 meses.**

**Resultado:** CS proativo. Churn 3%. Conversion Free → Growth 5%.

---

### Tier 3 — Premium Services (6+ meses)
**Objetivo:** CSM, SLA enterprise, community.

| BKL | O Quê |
|-----|-------|
| CS-020 | Contratar CSM (1 FTE) quando tiver 50+ Pro customers |
| CS-023 | SLA 99.99% (Enterprise) |
| CS-024 | Private Slack channel (Pro) |

**Viável quando:** 50+ clientes pagantes, NPS >= 50, ARR > R$ 150k/ano.

---

## Impacto Esperado

### Churn Rate
| Cenário | Hoje | 12 meses (Tier 1+2) | Delta |
|---------|------|---------------------|-------|
| **Free** | 2-3%/mês | 1-2%/mês | -50% |
| **Growth/Pro** | 5-8%/mês | 2-3%/mês | -50-60% |
| **Média ponderada** | 4-5%/mês | 2-2.5%/mês | **-45%** |

**Impact MRR:** Se 50 clientes Growth (R$ 99 = R$ 4,950/mês):
- Hoje: 50 × 6% = 3 clientes churn/mês = -R$ 297/mês
- Future: 50 × 2% = 1 cliente churn/mês = -R$ 99/mês
- **Economia:** +R$ 198/mês, +R$ 2,376/ano retention gain

### Conversion Free → Paid
| Métrica | Hoje | 12 meses |
|---------|------|----------|
| Free activation rate (day 7) | ~40% | **70%** |
| Free → Growth (30d) | 1% | **5%** |
| LTV aumento | +4x | Ticket médio Growth cresce |

**Impact:** Se 1,000 Free users, 5% upgrade para Growth:
- 50 clientes × R$ 99 = R$ 4,950/mês = R$ 59,400/ano **novo ARR**

---

## Investimento Necessário

### Tier 1 (0-30 dias)
```
Recursos:
  - 1 pessoa tech (eng/PM): 16-18 dias
  - 1 pessoa content/PM: 5 dias
  - Brio (attend-service) já existe
  
Ferramentas:
  - Resend (email) — R$ 0-100/mês
  - Telegram API (churn alerts) — grátis
  - Discord bot — grátis
  
Total $ cost: ~R$ 200-300/mês
```

### Tier 2 (30-120 dias)
```
Recursos:
  - Tech: 35-40 dias (distribuído em 3 meses)
  - PM: 10 dias
  - Data: 3 dias (ML)
  
Total: +80-90 dias eng
```

### Tier 3 (6+ meses)
```
Recursos:
  - 1 FTE CSM: R$ 6-8k/mês (quando tiver 50+ Pro customers)
  - Ops: 15 dias PM
  
Trigger: ARR > R$ 150k/ano em tier premium
```

---

## Go/No-Go Checklist (Antes de Launch)

### Documentação
- [ ] KBA-001 a KBA-010 publicadas
- [ ] Getting Started: SDK até primeiro evento em 5 min
- [ ] API docs 100% coverage

### Infra
- [ ] Health score cron: rodando 24/7, sem crashes
- [ ] Email service: Resend testado
- [ ] Brio WhatsApp: integrado, responde <15s

### Compliance
- [ ] LGPD: dados sensíveis não em logs
- [ ] Privacy policy publicada
- [ ] ToS criado

### Commercial
- [ ] Asaas integração: cobrança automática testada
- [ ] Upgrade flow: sem fricção

### Monitoring
- [ ] Uptime monitoring: status page pública
- [ ] Error tracking: Sentry ativo
- [ ] Support queue: tickets chegam

**Go-to-Market Go:** Se 18/20 checkboxes estão checked.

---

## Recomendação

### Começar imediatamente com Tier 1 (Semana que vem)

**Por quê?**
1. Bloqueador para venda (não dá para ir ao market sem suporte)
2. Viável em 4 semanas (16-18 dias eng)
3. ROI claro (reduz churn, aumenta conversion)
4. De baixo risco (documentação + email, fácil reverter)

**Próximas ações:**
- [ ] Semana 1-2: CS-001 + CS-002 + CS-003 paralelo
- [ ] Semana 3: CS-004 + CS-005 + review
- [ ] Semana 4: CS-006 + polish, go-to-market
- [ ] Dia 28: Go/No-Go decision

**Owner:** @pm (Antonio) coordena. @tl para health score + Brio. Support para KBAs.

---

## Alternativa: Não fazer (Risks)

❌ **Sem Tier 1 CS:**
- Lançar Pulse com suporte reativo = 8-10% churn (indústria)
- KBAs faltando = 50+ tickets/mês sobre mesmas coisas
- Saúde cliente invisível = surpresa quando cancela
- Brio não integrado = manual support, não escalável
- Resultado: time queimado em suporte, zero energia para product

---

## Conclusão

Pulse é um produto bom, mas **clientes precisam saber usá-lo**. Estrutura de CS implementada em 4 semanas (Tier 1) é o **mínimo viável para não falhar**.

Tier 2 (3 meses) transforma isso em vantagem competitiva vs LaunchDarkly/VWO — churn 3% vs 5-8%.

**Recomendação:** ✅ **Go com Tier 1 — begin Sprint 1 Monday**

---

## Apêndice: Estimativas Detalhadas

### CS-001: KBA 1-5 (3-4 dias)
```
- KBA-001: 3h (5 min read + 3 workflows)
- KBA-002: 4h (SDK install, copy-paste, React)
- KBA-003: 5h (debug tree, 4 causas, curl examples)
- KBA-004: 4h (A/B testing, p-value, interpretation)
- KBA-005: 2h (pricing matrix, recommender)
- Review + video: 3h

Total: 21h ≈ 3 dias
Owner: @pm + support
```

### CS-002: Health Score (2-3 dias)
```
- Design: 2h (4 dimensions, formula)
- Migration: 3h (create table, indices)
- Cron job: 4h (Go code, error handling)
- Queries: 3h (aggregate logins, flags, events)
- Tests: 4h (10 tenant scenarios)
- Dashboard widget: 3h (React component)

Total: 19h ≈ 2.5 dias
Owner: @tl
```

### CS-003: Email Sequence (2 dias)
```
- Design: 2h (6 email scenarios)
- Resend templates: 6h (HTML, variables)
- Cron triggers: 3h (day 0, 1, 7, 30, 60)
- Tests: 3h (sandbox, delivery)
- Analytics: 1h (tracking setup)

Total: 15h ≈ 2 dias
Owner: @pm
```

### CS-004: Brio Integration (3 dias)
```
- Brio SOUL.md update: 2h
- KBA index embed: 2h
- Webhook endpoint: 4h
- Escalation flow: 3h
- Training + tests: 4h

Total: 15h ≈ 2 dias eng + 1 dia training
Owner: @ai (Brio) + support
```

### CS-005: Churn Alerts (1 dia)
```
- Alert logic: 2h
- Email template: 1h
- Brio message: 1h
- Integrations (Resend + Telegram): 2h
- Tests: 1h

Total: 7h ≈ 1 dia
Owner: @tl
```

### CS-006: KBA 6-10 (3 dias)
```
- Each KBA: 4-5h (similar to CS-001)
- Total: 5 KBAs × 4.5h = 22.5h ≈ 3 dias
Owner: @pm + support
```

**Tier 1 Total:** 16-18 days (assuming 8h/day)

---

**Document prepared by:** IIT Customer Success Framework  
**Reviewed by:** @pm (strategic approval pending)  
**Next:** Weekly standup to track Tier 1 progress
