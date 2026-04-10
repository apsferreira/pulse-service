# Comparativo de Pricing Detalhado — Pulse vs Concorrentes

**Data:** 10 de abril de 2026  
**Análise:** Positioning competitivo por segmento de cliente

---

## I. SEGMENTO 1: Startup Seed (5k-25k MAU)

### Cenário: Equipe 5 pessoas, primeiro ano pós-seed, precisa flags + analytics

**Requisitos:**
- Feature flags com rollout gradual
- A/B testing básico
- Analytics de produto simples
- Max 25k MAU, 500k eventos/mês
- 1-3 projetos, 2 ambientes
- Email support

### Pricing Mensal

```
┌─────────────────┬──────────────┬──────────────┬──────────────┬──────────────┐
│ Produto         │ USD/mês      │ R$/mês (5.2) │ Inclui       │ Vantagem     │
├─────────────────┼──────────────┼──────────────┼──────────────┼──────────────┤
│ LaunchDarkly    │ ~$112        │ R$583        │ Flags only   │ Cheap entry  │
│ VWO Starter     │ ~$198        │ R$1.030      │ Flags + A/B  │ Full stack   │
│ PostHog Free    │ R$0          │ R$0          │ Analytics    │ Free tier    │
│ Pulse Starter   │ ~R$399       │ R$399        │ Flags + A/B  │ Melhor valor │
│ (PROPOSTO)      │ (+33% vs $)  │ (38% ≤ VWO)  │ + Analytics  │             │
└─────────────────┴──────────────┴──────────────┴──────────────┴──────────────┘
```

**Análise:**
- **Vs LaunchDarkly:** +R$184/mês, mas ganha A/B testing + analytics → **+35% valor**
- **Vs VWO:** -R$631/mês (-38%), mesmas features → **Compelling offer**
- **Vs PostHog:** +R$399/mês vs Free, mas com SLA + suporte + LGPD → **Trade-off aceitável**

**Argumento de venda:**
> "LaunchDarkly custa R$583/mês e não tem A/B testing. VWO custa R$1.030 e cobra módulos separados. Pulse é R$399 com tudo incluído + suporte em português + compliance LGPD."

**Público-alvo esperado:** 100-150 startups Seed/A BR (Loft, Plug, Sieve, Wemed, etc)

---

## II. SEGMENTO 2: Startup Série A (25k-100k MAU)

### Cenário: Equipe 20 pessoas, scaling fast, precisa flags + A/B + analytics + integrações

**Requisitos:**
- Flags com rollout gradual + targeting avançado
- A/B testing com significância estatística
- Analytics com funis, retenção, cohorts
- 100k MAU, 5M eventos/mês
- 10 projetos, 5 ambientes
- Slack support (24h)
- Webhooks, custom SSO

### Pricing Mensal

```
┌─────────────────┬──────────────┬──────────────┬──────────────┬──────────────┐
│ Produto         │ USD/mês      │ R$/mês (5.2) │ Valor        │ Vantagem     │
├─────────────────┼──────────────┼──────────────┼──────────────┼──────────────┤
│ LaunchDarkly    │ ~$400-800    │ R$2.080-4160 │ Flags only   │ Escalável    │
│ VWO Growth      │ ~$422-600    │ R$2.194-3120 │ Flags + A/B  │ Modular      │
│ PostHog Growth  │ ~$100-200    │ R$520-1040   │ Analytics    │ Mais barato  │
│ Pulse Growth    │ ~R$1.099     │ R$1.099      │ Full stack   │ Melhor deal  │
│ (PROPOSTO)      │              │              │ + LGPD       │             │
└─────────────────┴──────────────┴──────────────┴──────────────┴──────────────┘
```

**Análise:**
- **Vs LaunchDarkly Pro:** -R$1.081-3.061/mês (-50%), paridade de features → **10x ROI**
- **Vs VWO Growth:** -R$1.095-2.021/mês (-50%), inclui analytics → **Huge win**
- **Vs PostHog Growth:** +R$59-579/mês, mas com SLA + support → **Premium tier**

**Argumento de venda:**
> "VWO Growth custa R$2.194+ com módulos separados. LaunchDarkly Pro está em R$2.080+. Pulse é R$1.099 com TUDO: flags + A/B + analytics + webhooks + Slack support. E em português."

**Público-alvo esperado:** 30-50 startups Série A BR

**Impacto por cliente:** R$1.099 × 36 clientes = **R$39.564/mês = R$475k ARR** (cenário base)

---

## III. SEGMENTO 3: Mid-Market (100k-500k MAU)

### Cenário: Equipe 100+ pessoas, scale-up establecida, precisa Enterprise features

**Requisitos:**
- Tudo do Growth, mais:
  - Multi-variate A/B testing (até 16 variantes)
  - Análise Bayesiana (não só frequentista)
  - Custom segments + ML recommendations
  - Approval workflow para flags
  - Dedicated account manager
  - Phone + Slack support (4h SLA)
  - 500k MAU, ilimitados eventos
  - Audit logs + compliance reports

### Pricing Mensal

```
┌─────────────────┬──────────────┬──────────────┬──────────────┬──────────────┐
│ Produto         │ USD/mês      │ R$/mês       │ SLA          │ Vantagem     │
├─────────────────┼──────────────┼──────────────┼──────────────┼──────────────┤
│ LaunchDarkly    │ USD 2k+      │ R$10.400+    │ 99.99%       │ Enterprise   │
│ VWO Enterprise  │ USD 1k+      │ R$5.200+     │ Customizado  │ Customizado  │
│ PostHog Enterprise│ USD 500-1k  │ R$2.600-5200 │ Negociável   │ Mais barato  │
│ Pulse Pro       │ ~R$2.399     │ R$2.399      │ 99.9%        │ 75% discount │
│ (PROPOSTO)      │              │              │ Roadmap 99.99│             │
└─────────────────┴──────────────┴──────────────┴──────────────┴──────────────┘
```

**Análise:**
- **Vs LaunchDarkly Enterprise:** -R$8.001+/mês (-75%), paridade features → **Steal**
- **Vs VWO Enterprise:** -R$2.801+/mês (-50%), mais features (webhooks, LGPD) → **Strong**
- **Vs PostHog Enterprise:** Comparável em preço, melhor UX → **Different positioning**

**Argumento de venda:**
> "LaunchDarkly Enterprise em R$10.400/mês é feito para empresas que pagam em USD. Pulse Pro é R$2.399 com compliance LGPD built-in, dados no Brasil, e integração com auth/billing brasileiros."

**Público-alvo esperado:** 2-5 empresas mid-market BR (fintechs, healthtechs, e-commerce)

**Impacto por cliente:** R$2.399 × 3 clientes = **R$7.197/mês = R$86k ARR**

---

## IV. SEGMENTO 4: Enterprise Self-Hosted (Unlimited MAU)

### Cenário: Banco, fintech, governo — precisa dados on-premise (LGPD contractual)

**Requisitos:**
- Deploy on-prem/VPC/Kubernetes
- Unlimited MAU, flags, projects, seats
- Compliance: LGPD audit trail, dados zero cloud
- Support: 24/5 (Slack + phone)
- SLA: 99.9% → 99.99% em roadmap

### Pricing Anual

```
┌─────────────────────┬──────────────┬──────────────┬──────────────┬──────────────┐
│ Solução             │ USD/ano      │ R$/ano       │ Modelo       │ Vantagem     │
├─────────────────────┼──────────────┼──────────────┼──────────────┼──────────────┤
│ LaunchDarkly Ent.   │ USD 24k-60k  │ R$125k-312k  │ Seats + MAU  │ Enterprise   │
│ PostHog Self-Hosted │ USD 0        │ R$0          │ Open-source  │ Livre (infra |
│                     │              │              │              │  caro)      │
│ Pulse Self-Hosted   │ ~R$50k       │ R$50k        │ Licença fixa │ 75% discount │
│ (PROPOSTO)          │              │ + R$24k/ano  │ + suporte    │ vs LD       │
│                     │              │ suporte      │              │             │
└─────────────────────┴──────────────┴──────────────┴──────────────┴──────────────┘
```

**Análise:**
- **Vs LaunchDarkly Enterprise:** -R$75k-262k/ano (-75%), mesmo scope → **Defensável**
- **Vs PostHog:** +R$74k/ano, mas gerenciado + suporte → **Trade-off**

**Argumento de venda:**
> "Você pode comprar LaunchDarkly Enterprise por R$125k-312k/ano. Ou PostHog free + contratar DevOps (R$100k+/ano). Pulse é R$50k licença + R$24-60k/ano suporte = R$74-110k total com tudo gerenciado + compliance report automático."

**Público-alvo esperado:** 1-2 deals/ano (fintechs, bancos, gov)

**Impacto por cliente:** R$74k ARR × 2 = **R$148k ARR** (vs R$312k/ano se LD)

---

## V. ANÁLISE DE SENSIBILIDADE (CAC + LTV)

### Assumindo CAC padrão por segmento

```
Tier        | CAC típico  | LTV (24m) | Payback | LTV/CAC | Viável?
────────────┼─────────────┼───────────┼─────────┼─────────┼─────────
Starter     | R$150-200   | R$9.576   | 3-4m    | 48-64x  | ✅ Sim
Growth      | R$400-500   | R$31.752  | 2-3m    | 64-79x  | ✅ Sim
Pro         | R$1k-1.5k   | R$62.376  | 1.5-2m  | 42-62x  | ✅ Sim
Self-Hosted | R$5k-10k    | R$240k    | 3-4m    | 24-48x  | ✅ Sim
```

**Benchmark SaaS:** 3x+ LTV/CAC = saudável. **Pulse está 48-79x = excelente**

---

## VI. MATRIZ DE COMPETITIVIDADE

### Por feature + preço

```
                  LaunchDarkly  VWO          PostHog      Pulse
Preço             ⭐⭐⭐        ⭐⭐⭐        ⭐⭐⭐⭐⭐    ⭐⭐⭐⭐⭐
(vs preço market)

Feature flags     ⭐⭐⭐⭐⭐    ⭐⭐⭐⭐⭐    ⭐⭐⭐⭐⭐    ⭐⭐⭐⭐⭐

A/B testing       ⭐⭐⭐⭐⭐    ⭐⭐⭐⭐⭐    ⭐⭐⭐⭐⭐    ⭐⭐⭐⭐ (stats 🚧)

Analytics         ❌            ⭐⭐⭐⭐⭐    ⭐⭐⭐⭐⭐    ⭐⭐⭐⭐ (básico)

Self-hosted       ❌            ❌            ⭐⭐⭐⭐⭐    ⭐⭐⭐⭐⭐

LGPD compliance   ❌            ⚠️ (LatAm)    ⭐⭐⭐⭐⭐    ⭐⭐⭐⭐⭐

Suporte PT-BR     ❌            ❌            ❌            ⭐⭐⭐⭐⭐

Integração IIT    ❌            ❌            ❌            ⭐⭐⭐⭐⭐
```

**Conclusão:** Pulse vence em preço + compliance + localização. Cede em analytics maduro e UX (PostHog/VWO mais polidas visualmente).

---

## VII. DECISÃO: QUAL PRICING IMPLEMENTAR?

### Opção A: Agressivo (Recomendado)
```
Starter:   R$399/mês  (+33% vs R$299)
Growth:    R$1.099/mês (+38% vs R$799)
Pro:       R$2.399/mês (+20% vs R$1.999)
Self-Hosted: R$50k/ano + R$2-5k/mês suporte
```

**Rationale:** Ainda 30-50% abaixo concorrentes, margins se tornam saudáveis (45-75%), não attracts bargain hunters, CAC payback reduz 50%.

### Opção B: Conservador
```
Starter:   R$349/mês  (+16% vs R$299)
Growth:    R$999/mês  (+25% vs R$799)
Pro:       R$2.199/mês (+10% vs R$1.999)
```

**Rationale:** Menos agressivo, pero margins still weak para Starter. Viável se volume é 2x.

### Opção C: Alinhar com PostHog
```
Starter:   R$299/mês  (sem mudança)
Growth:    R$599/mês  (-25%)
```

**Rationale:** Race to bottom. Nunca recomendado.

---

## VIII. IMPLEMENTAÇÃO RECOMENDADA

### Phase 1 (Imediato): Publicar Preço Opção A
- Página públicamente com comparativo lado-a-lado
- CTA: "Free tier, depois Starter"
- FAQ: "Por que R$399? Porque não cobramos por MAU — você paga flat, nós assumimos risco"

### Phase 2 (Mês 2): A/B teste discreto
- 20% do traffic: vende Starter R$349 (controle)
- 80% do traffic: vende Starter R$399 (teste)
- Métrica: conversion rate (Free → Starter)
- Se R$399 ≥ conversion rate de R$349, mantém R$399

### Phase 3 (Mês 3): Feedback de primeiros clientes
- Recolher NPS + feedback de pricing de 5-10 clientes Growth
- Ajustar se necessário (improvável; mercado está esperando opção brasileira cara)

---

## IX. CONCLUSÃO COMPETITIVA

**Pulse com pricing proposto (Opção A) é:**

| Vs | Posição | Estratégia |
|---|---|---|
| LaunchDarkly | Cheaper + LGPD | Atraira startups BR risk-conscious |
| VWO | Cheaper + all-in | Vence em valor (não modular) |
| PostHog | Pricier + managed | Vence em confiabilidade + LGPD + suporte |
| Avo/Statsig | Diferente (tracking plan) | Não compete; complementar |

**Mensagem de venda unificada:**
> "Pulse é LaunchDarkly + VWO + PostHog em um produto, 30-50% mais barato, com compliance LGPD built-in e suporte em português. Escolha: cloud SaaS ou self-hosted."

**Público:** Startups brasileiras, agências de growth, fintech/healthtech regulated

---

**Documento finalizado:** 10 de abril de 2026  
**Próximo passo:** Aprovação CEO + publicação pricing page
