# Pulse Pricing Strategy

**Versão:** 1.0  
**Data:** 10 de abril de 2026  
**Produto:** Pulse (Feature Flags + A/B Testing + Product Analytics)  
**Mercado:** Brasil (BRL)  
**Status:** Análise competitiva completa + proposta de pricing v1

---

## Sumário Executivo

O **Pulse** é posicionado como alternativa brasileira com custo 40-60% menor que LaunchDarkly e VWO em escala de 10k-100k MAUs, com diferencial de self-hosting privado (reduz COGS em 60-80% vs cloud puro).

**Proposta de valor:**
- Feature flags com rollout gradual e A/B testing integrado
- Self-hosted ou SaaS cloud
- Compliance LGPD nativo (dados cliente não saem do Brasil)
- Pricing transparente e previsível (sem surpresas de MAU)

**Mercado-alvo:** Startups brasileiras, agências digitais, e-commerce, fintechs (10k-100k MAUs / ~R$50M-R$500M ARR)

---

## 1. Análise de Pricing — Concorrentes

### 1.1 LaunchDarkly

**Modelo:** Híbrido (Seats + MAUs + Features)

| Métrica | Valor |
|---------|-------|
| **Free Plan** | 2 seats, 1 project, 2 environments, 1k MAU (feature flags) |
| **Foundation (Paid)** | $12/mês por service connection + $10 por 1k MAU |
| **Pro Plan** | $12/seat/mês + MAU charges |
| **Experimentation Add-on** | +$3 por 1k MAU (feature flags: $0.01/MAU/mês) |
| **Enterprise** | Custom pricing (contato vendedor) |

**Custo para 10k MAUs (Foundation tier):**
- Base: $12/mês (1 service connection)
- MAUs: $10 × 10 = $100/mês
- **Total: ~$112/mês ≈ $1.344/ano**

**Custo para 100k MAUs (Foundation tier):**
- Base: $12/mês (1 service connection)
- MAUs: $10 × 100 = $1.000/mês
- Experimentation add-on: $3 × 100 = $300/mês
- **Total: ~$1.312/mês ≈ $15.744/ano**

**Característica:** Sem volume discount; custo linear com crescimento de usuários.

---

### 1.2 VWO (Visual Website Optimizer)

**Modelo:** MTU-based (Monthly Tracked Users) por módulo

| Métrica | Valor |
|---------|-------|
| **Free Plan** | Até 50k MAU (limitado a features básicas) |
| **Starter Plan** | ~$198/mês (30k MTUs anuais) |
| **Growth Plan** | ~$422/mês (100k MTUs anuais) |
| **Pro Plan** | ~$997/mês (500k+ MTUs anuais) |
| **Enterprise** | Custom pricing (contato vendedor) |

**Módulos separados:**
- VWO Testing (A/B testing)
- VWO Insights (Analytics, heatmaps, recordings)
- VWO Personalize
- VWO Web Rollouts (Feature flags)
- VWO FullStack (Mobile)

**Exemplo Growth Plan (30k MTUs):**
- VWO Testing: $179/mês
- VWO Insights: $199/mês
- VWO Web Rollouts: $149/mês
- **Total: ~$527/mês ≈ $6.324/ano**

**Para 100k MTUs:** Requer Pro/Enterprise = $997+ e múltiplos módulos. Estimado **$2.500-3.500/mês = $30k-42k/ano**

**Característica:** Modular (paga separado por feature), descontos anuais de 15-20%, requer contato sales para Enterprise.

---

### 1.3 PostHog (Referência — Freemium)

**Modelo:** Usage-based (eventos)

| Métrica | Valor |
|---------|-------|
| **Free Tier** | 1M analytics events/mês, 5k session replays, 1M feature flag requests |
| **Paid** | $0.00005-0.00001 por evento adicional (volume-based) |
| **Example (10M events)** | ~$45-50/mês |
| **Example (100M events)** | ~$300-500/mês |

**Característica:** Muito barato em larga escala, máxima transparência, open-source, self-hosting disponível.

---

### 1.4 Resumo Competitivo (10k MAU)

| Plataforma | Custo mensal USD | Custo anual BRL* | Modelo |
|---|---|---|---|
| **LaunchDarkly** | $112 | R$6.949 | Seats + MAU |
| **VWO** | $198+ | R$12.269 | MTU modular |
| **PostHog** | $45-50 | R$2.793-3.120 | Usage events |
| **Pulse (Proposta)** | R$299/mês | R$3.588/ano | Flat tier + self-hosted |

*Câmbio: R$5,20 = USD 1,00; preço anual = mensal × 12

---

## 2. Proposta de Pricing — Pulse

### 2.1 Tiers Propostos

#### **Tier 1: Free** — Comunidade e MVPs

```
Preço: R$ 0/mês
Inclui:
  • Até 5k MAU
  • 10 feature flags (total)
  • 1 environment (dev/staging/prod)
  • 1 projeto
  • Feature flags com rollout percentual
  • Sem A/B testing
  • Comunidade Slack (suporte best-effort)
  • Dados em cloud SaaS (Brasil)
  • 7 dias retenção de eventos

Limite: 1 mês inatividade → purga de dados
```

**Posicionamento:** Aquisição, educação, community building.

---

#### **Tier 2: Starter** — Pequenas equipes e startups MVP

```
Preço: R$ 299/mês (R$ 299 × 12 = R$ 3.588/ano)
Inclui tudo do Free, mais:
  • Até 25k MAU
  • 50 feature flags (total)
  • 3 environments
  • 3 projetos
  • A/B testing com cálculo de significância estatística
  • Análise básica de impacto (conversão, CTR)
  • Email support (72h SLA)
  • Dados em cloud SaaS (Brasil) com criptografia
  • 30 dias retenção de eventos
  • API básica de flags
  • 2 seats (usuários admin)

Público: Startups Seed/Série A, agências pequenas, equipes de 5-20 pessoas
```

**Argumento de venda:** "Metade do preço do LaunchDarkly, mesmo valor que VWO Starter, mas com A/B testing incluído."

---

#### **Tier 3: Growth** — Startups em crescimento

```
Preço: R$ 799/mês (R$ 799 × 12 = R$ 9.588/ano)
Inclui tudo do Starter, mais:
  • Até 100k MAU
  • 500 feature flags (total)
  • 5 environments
  • 10 projetos
  • Análise avançada (múltiplas variantes, segmentação por coorte)
  • User cohort tagging e targeting avançado
  • Webhooks para integrações (Slack, Zapier, etc)
  • Priority email support (24h SLA)
  • Dados em cloud SaaS (Brasil)
  • 90 dias retenção de eventos
  • 5 seats
  • Custom SSO (SAML)
  • API avançada com rate limiting generoso

Público: Startups Série A/B, médias agências digitais, 20-100 pessoas
```

**Argumento de venda:** "A/B testing + feature flags num único plano. VWO cobrava módulos separados; nós incluímos tudo."

---

#### **Tier 4: Pro** — Empresas de scale

```
Preço: R$ 1.999/mês (R$ 1.999 × 12 = R$ 23.988/ano)
Inclui tudo do Growth, mais:
  • Até 500k MAU
  • 5.000 feature flags (total)
  • Unlimited environments
  • Unlimited projetos
  • Multi-variate testing (até 16 variantes)
  • Análise Bayesiana + frequentista
  • Custom segments e ML-powered recommendations
  • Phone + priority email support (4h SLA)
  • Dados em cloud SaaS (Brasil) com backup automático
  • 180 dias retenção de eventos
  • 20 seats
  • Advanced RBAC (roles customizados)
  • API com limite customizável
  • Premium SDKs (React Native, Flutter, mobile nativo)

Público: Empresas médias de tech, fintechs, e-commerce 20-500M ARR, 100-500+ pessoas
```

**Argumento de venda:** "LaunchDarkly + experimentation, mas 30% mais barato. Compliance LGPD built-in."

---

#### **Tier 5: Self-Hosted (Enterprise)** — Compliance + privacidade

```
Preço: Licença anual (modelos abaixo)

**Modelos:**
a) Licença Fixa: R$ 24.000/ano
   • Deploy ilimitado (on-prem, VPC, Kubernetes)
   • Unlimited MAU, flags, projects, seats
   • Suporte técnico 24/5 (Slack + email)
   • Atualizações menores garantidas
   • Dados 100% no cliente (zero cloud)
   • Compliance LGPD/LGPD-compliant

b) Licença por Tier (MAU) + Suporte:
   • Até 100k MAU: R$ 12.000/ano + R$ 2.000/mês support
   • 100k-1M MAU: R$ 24.000/ano + R$ 3.000/mês support
   • 1M+ MAU: R$ 48.000/ano + R$ 5.000/mês support

Público: Bancos, fintechs reguladas, healthcare, governo, empresas Fortune 500 Brasil
```

**Diferencial:** Dados nunca saem do Brasil. LGPD compliance. Deploy control total.

---

### 2.2 Resumo Dos Tiers

| Tier | Preço/mês | MAU | Flags | Envs | A/B testing | Suporte | Público |
|---|---|---|---|---|---|---|---|
| **Free** | R$ 0 | 5k | 10 | 1 | ❌ | Community | MVPs, educação |
| **Starter** | R$ 299 | 25k | 50 | 3 | ✅ | Email (72h) | Startups Seed/A |
| **Growth** | R$ 799 | 100k | 500 | 5 | ✅ | Email (24h) | Startups A/B |
| **Pro** | R$ 1.999 | 500k | 5k | ∞ | ✅ Multi-variate | Phone (4h) | Mid-market |
| **Self-Hosted** | R$ 12k-48k/ano | ∞ | ∞ | ∞ | ✅ | 24/5 | Enterprise |

---

### 2.3 Add-ons (Todos os tiers pagos)

| Add-on | Preço | Descrição |
|---|---|---|
| **Advanced Analytics Addon** | +R$ 299/mês | Funnel analysis, retention cohorts, custom metrics |
| **Enterprise SSO (SAML/OIDC)** | +R$ 199/mês | Custom SSO (fora do Pro) |
| **Dedicated Support** | +R$ 999/mês | Engenheiro dedicado, SLA 1h, planejamento estratégico |
| **Data Residency Audit** | +R$ 499/mês | Conformidade auditável (LGPD, ISO 27001) |
| **Premium SDK Suite** | +R$ 199/mês | Todos SDKs premium: React Native, Flutter, native iOS/Android, Go, Python |

---

## 3. Unit Economics

### 3.1 COGS — Cloud SaaS (Tier Starter/Growth/Pro)

**Infraestrutura mensal por tier (estimado):**

| Componente | Starter (25k MAU) | Growth (100k MAU) | Pro (500k MAU) |
|---|---|---|---|
| **PostgreSQL (RDS)** | R$ 180 | R$ 350 | R$ 850 |
| **Redis (ElastiCache)** | R$ 100 | R$ 200 | R$ 500 |
| **Go API Server (K3s/ECS)** | R$ 150 | R$ 300 | R$ 600 |
| **Data egress** | R$ 50 | R$ 100 | R$ 250 |
| **S3 storage (events)** | R$ 80 | R$ 200 | R$ 500 |
| **Monitoring (Datadog)** | R$ 150 | R$ 150 | R$ 150 |
| **CDN (CloudFlare)** | R$ 50 | R$ 75 | R$ 100 |
| **Backup e replicação** | R$ 100 | R$ 150 | R$ 250 |
| **Support overhead** | R$ 200 | R$ 200 | R$ 300 |
| **Total COGS/mês** | **R$ 1.060** | **R$ 1.725** | **R$ 3.500** |

*Estimativa baseada em providers AWS/GCP/Azure; em homelab IIT (infra própria), reduz 60-80%.*

---

### 3.2 Gross Margin

| Tier | Preço/mês | COGS/mês | Margin | % |
|---|---|---|---|---|
| **Starter** | R$ 299 | R$ 1.060* | (R$ 761) | -254% ⚠️ |
| **Growth** | R$ 799 | R$ 1.725* | (R$ 926) | -116% ⚠️ |
| **Pro** | R$ 1.999 | R$ 3.500* | (R$ 1.501) | -75% ⚠️ |

*Nota:* Estes números indicam que em cloud puro (AWS/GCP/Azure), os tiers baixos são **não-rentáveis**. Soluções:

1. **Usar infra IIT homelab** (reduz COGS 60-80%) → Starter: 70% margin, Growth: 60%, Pro: 50%
2. **Modelo freemium agressivo** (Free tier com ads ou watered-down features) → ganhar scale, depois monetizar
3. **Bundling com outros serviços IIT** (auth-service, customer-service) → dilui custo fixo

---

### 3.3 Gross Margin (Cenário Homelab + Bundling)

**Com infra IIT própria (redução 70% COGS) + economia de escala:**

| Tier | Preço/mês | COGS/mês | Margin | % | LTV (24m) |
|---|---|---|---|---|---|
| **Starter** | R$ 299 | R$ 318 | (R$ 19) | -6% | R$ 4.788 |
| **Growth** | R$ 799 | R$ 517 | R$ 282 | 35% | R$ 19.176 |
| **Pro** | R$ 1.999 | R$ 1.050 | R$ 949 | 47% | R$ 47.952 |
| **Self-Hosted** | R$ 12k/ano | R$ 2k/ano | R$ 10k/ano | 83% | R$ 120k |

**Observações:**
- Starter é loss-leader (adquisição) — payback em ~4 meses se upgrade para Growth
- Growth e Pro viáveis com 35-47% margin
- Self-Hosted é high-margin (83%) — foco B2B Enterprise

---

### 3.4 LTV / CAC Ratio

**Cenário padrão de aquisição:**

| Métrica | Starter | Growth | Pro |
|---|---|---|---|
| **CAC** (marketing spend) | R$ 150 | R$ 400 | R$ 800 |
| **LTV 24 meses** | R$ 4.788 | R$ 19.176 | R$ 47.952 |
| **LTV / CAC ratio** | 32x | 48x | 60x | ✅
| **Payback period** | 6-8 meses | 4-6 meses | 2-3 meses |

**Benchmark:** SaaS saudável = 3x+ LTV/CAC. Pulse está em **32-60x** = excelente.

---

### 3.5 Churn & Retention Esperada

**Assumindo:**
- Starter: 5% churn/mês (novo, sensível preço)
- Growth: 2% churn/mês (produto-market fit)
- Pro: 1% churn/mês (sticky, enterprise)
- Self-Hosted: 0.5% churn/ano (contractual)

**Lifetime (até churn zero):**
| Tier | Lifetime (meses) | LTV |
|---|---|---|
| **Starter** | 20 | R$ 5.980 |
| **Growth** | 50 | R$ 39.950 |
| **Pro** | 100+ | R$ 199.900+ |
| **Self-Hosted** | 240+ | R$ 288k+ |

---

## 4. Análise de Mercado — Brasil

### 4.1 TAM / SAM / SOM

**TAM (Total Addressable Market):** Feature flags + A/B testing globalmente  
- Mercado global: ~$3-5B (2024-2025)
- Brasil share: ~3% = **R$ 1-1.5B** (~$200-300M)

**SAM (Serviceable Available Market):** Brasil, startups tech + agências + e-commerce + fintech  
- ~5.000 empresas brasileiras com 10k+ MAU (VC-backed, public, scale-up)
- Renda média: R$ 200/ano per company = **R$ 1B** potencial

**SOM (Serviceable Obtainable Market):** Year 1 alvo IIT  
- Captura 2% de SAM = **20-30 clientes pagos** (~R$ 2-3M ARR)
- Captura 5% de SAM = **100-150 clientes** (~R$ 10-15M ARR) — Year 3

---

### 4.2 Competição no Brasil

| Concorrente | Presença | Status |
|---|---|---|
| **LaunchDarkly** | Global (não tem data center BR) | Presença via cloud global; não focado em compliance LGPD |
| **VWO** | Global + LatAm (dados LatAm) | Cloud LatAm; presença forte em CRO/conversion |
| **PostHog** | Global + open-source | Trending entre startups; self-hosted easy |
| **GrowthBook** | Open-source | Crescimento lento; falta de commercialization |
| **Statsig** | Crescente | Series A; modelo PLG |
| **Flagsmith** | Open-source | Niche; developers |
| **Pulse (IIT)** | **Local, brasil-first** | **Oportunidade: compliance LGPD nativo + preço local** |

**Vantagem competitiva Pulse:**
1. Dados BR (LGPD compliance sem workaround)
2. Pricing 30-50% menor que LaunchDarkly/VWO
3. Self-hosted option (atrai enterprise risk-averse)
4. Integrado com ecossistema IIT (auth, customer, billing)

---

### 4.3 Estratégia GTM — Brasil

#### **Fase 1 (Meses 1-3): Freemium Acquisition**
- Free tier agressivo (5k MAU grátis) → educação + community
- Founders pack (desconto 50% primeiro ano, Starter) → viralidade
- Content: "LaunchDarkly custa $X/mês; Pulse custa R$299. Aqui está o lado a lado."

#### **Fase 2 (Meses 4-8): Product-Market Fit**
- Eventos (GDG Salvador, SDD, Loft) para pitch Pulse
- Case study de cliente Série A → prova de valor
- Benchmark público: Pulse vs LaunchDarkly vs VWO (dashboard aberto)

#### **Fase 3 (Meses 9-12): B2B Enterprise**
- Self-hosted sales → fintech, healthcare, bancos
- Partnerships: agências de growth, consultoria digital
- Conformidade formal: ISO 27001, SOC 2, LGPD audit trail

---

## 5. Projeção de Receita — 24 Meses

### 5.1 Cenário Base (Conservador)

**Assunções:**
- Mês 1-2: 10 clientes Free (acquisition cost $50 em events), 0 pagos
- Mês 3-6: 50 Free, 5 Starter (R$ 299/mês) = R$ 1.495/mês
- Mês 7-12: 200 Free, 20 Starter + 5 Growth = R$ 10.475/mês
- Mês 13-24: 500 Free, 50 Starter + 20 Growth + 2 Pro = R$ 28.000/mês
- Self-Hosted: 1 customer Year 1, 2 Year 2 (R$ 12k each/ano)

**Receita:**

| Período | Free Users | Starter | Growth | Pro | Self-Hosted | Receita mensal | Receita acumulada |
|---|---|---|---|---|---|---|---|
| Mês 1-2 | 10 | — | — | — | — | R$ 0 | R$ 0 |
| Mês 3-6 | 50 | 5 | — | — | — | R$ 1.495 | R$ 5.980 |
| Mês 7-12 | 200 | 20 | 5 | — | 1 SH (R$1k/mês) | R$ 10.475 | R$ 58.700 |
| Mês 13-24 | 500 | 50 | 20 | 2 | 2 SH | R$ 28.000 | R$ 393.200 |
| **Total Year 1** | | | | | | **~R$ 38.000** | (acum.) |
| **Total Year 2** | | | | | | **~R$ 280.000** | (acum.) |

---

### 5.2 Cenário Otimista (+50% conversão)

**Assunções:** Partnerships com agências, viral growth, case studies publicados.

| Período | Starter | Growth | Pro | Self-Hosted | Receita mensal |
|---|---|---|---|---|---|
| Mês 1-12 | 30 | 8 | 1 | 2 SH | ~R$ 95.000/mês ao final |
| Mês 13-24 | 80 | 35 | 5 | 4 SH | ~R$ 180.000/mês ao final |
| **Total Year 2** | | | | | **~R$ 1.650.000** |

---

### 5.3 Breakeven Analysis

**Com COGS em homelab (70% reduction):**

- **Fixed costs/mês:** R$ 15.000 (marketing, salários suporte, infrastructure)
- **Starter contribution:** (R$ 299 - R$ 318) ≈ -R$ 20/cliente (loss-leader)
- **Growth contribution:** (R$ 799 - R$ 517) = R$ 282/cliente
- **Pro contribution:** (R$ 1.999 - R$ 1.050) = R$ 949/cliente
- **Breakeven:** ~10-15 Growth customers + marketing para Free → Starter pipeline

**Estimado:** Breakeven em **Mês 9-10** (Year 1) com scenario base.

---

## 6. Comparação Competitiva — Pulse vs Concorrentes

### Tabela Side-by-Side (10k MAU, Brazil-focused)

| Atributo | Pulse (Starter) | LaunchDarkly | VWO | PostHog Free | Flagsmith |
|---|---|---|---|---|---|
| **Preço/mês USD** | ~R$ 299 ($57) | $112 | $198+ | Free | Free |
| **Preço/mês BRL** | R$ 299 | R$ 584 | R$ 1.030+ | R$ 0 | R$ 0 |
| **A/B Testing** | ✅ | ✅ | ✅ | ❌ | ❌ |
| **Feature Flags** | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Rollout Gradual** | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Analytics Integrado** | ✅ (básico) | ❌ (add-on) | ✅ (modular) | ✅ | ❌ |
| **Self-Hosted** | ✅ | ❌ | ❌ | ✅ | ✅ |
| **LGPD Data Residency** | ✅ Brasil | ❌ Global | ⚠️ LatAm | ✅ Self-hosted | ✅ Self-hosted |
| **Suporte PT-BR** | ✅ | ❌ | ❌ | ❌ | ❌ |
| **Integração IIT** | ✅ (auth, customer) | ❌ | ❌ | ❌ | ❌ |

**Mensagem de venda:**
> "Pulse combina feature flags + A/B testing com compliance LGPD nativo, self-hosted opcional, e preço 50% menor que LaunchDarkly. Dados nunca deixam o Brasil."

---

## 7. Recomendações Executivas

### 7.1 Recomendação de Pricing

**Adotar estrutura proposta (Free + Starter + Growth + Pro + Self-Hosted):**

1. **Free tier:** Aquisição agressiva (objetivo: 1.000 free users em Year 1)
2. **Starter (R$ 299):** Loss-leader; objetivo = conversão 10% free → starter
3. **Growth (R$ 799):** Core do modelo (50% da receita esperada)
4. **Pro (R$ 1.999):** High-margin; foco em mid-market
5. **Self-Hosted (R$ 12k-48k/ano):** Enterprise, high-margin, sticky

### 7.2 Metricas de Sucesso — Trimestral

| Métrica | T1 | T2 | T3 | T4 |
|---|---|---|---|---|
| **Free users** | 100 | 300 | 500 | 1k |
| **Paying customers** | 5 | 25 | 50 | 100+ |
| **ARR** | ~R$ 18k | ~R$ 150k | ~R$ 350k | ~R$ 600k+ |
| **Gross Margin (%)** | 30% | 35% | 40% | 45% |
| **CAC Payback (meses)** | 8 | 6 | 5 | 4 |

### 7.3 Roadmap Recomendado

**Q2 2026 (abr-jun):**
- [ ] Publicar pricing public (este documento)
- [ ] Implementar self-hosted distribution (Docker, Helm, Terraform)
- [ ] Criar caso de uso: "Startups Série A — como usar feature flags para go/no-go"
- [ ] Integrar analytics SDK (simples events logging)

**Q3 2026 (jul-set):**
- [ ] Beta de clientes pagos (Starter + Growth)
- [ ] Case study publicado (1 cliente série A)
- [ ] Suporte PT-BR ao vivo
- [ ] SLA SaaS vs self-hosted documentado

**Q4 2026 (out-dez):**
- [ ] 50+ clientes pagos
- [ ] Self-hosted enterprise closing (1-2 deals)
- [ ] Integração com sistemas de billing IIT (checkout-service)
- [ ] LGPD audit trail + compliance report automático

---

## 8. Appendix: Detalhes Técnicos

### 8.1 SDKs & Integração

Pulse suporta:
- **JavaScript/TypeScript** (npm + cdn)
- **React/React Native** (expo ready)
- **Go** (stdlib + chi/fiber)
- **Python** (requests)
- Roadmap: Java, Node.js, Ruby

### 8.2 Compliance & Segurança

- **Data encryption:** TLS 1.3 em transit, AES-256 at rest
- **Backup:** Daily snapshots, 30-day retention
- **Audit log:** Todos os flag changes loggados (LGPD compliance)
- **DPO ready:** LGPD data processing agreement padrão
- **Roadmap:** SOC 2 Type II (Q4 2026), ISO 27001 (2027)

### 8.3 Escalabilidade

- PostgreSQL 15 + pgxpool (connection pooling)
- Redis 7 (flag cache, TTL 60s)
- K3s + horizontal pod autoscaler (mín 2, máx 10 replicas)
- CDN CloudFlare (cacheable GET endpoints)
- Estimado: até 10M events/dia com current setup

---

## 9. Conclusão

O **Pulse** está posicionado para capturar 2-5% do mercado brasileiro de feature flags + A/B testing (SAM ~R$ 1B) através de:

1. **Preço local:** 30-50% menor que LaunchDarkly/VWO
2. **Compliance nativo:** LGPD data residency no Brasil
3. **Self-hosting:** Atrai enterprise risk-averse
4. **Integração:** Ecossistema IIT (auth, customer, billing)

**Projeção Year 1:** R$ 38k-95k ARR | **Projeção Year 2:** R$ 280k-1.65M ARR  
**Breakeven:** ~Mês 9-10 (scenario base, homelab infra)

**Próximos passos:** Aprovação de pricing, publicação pública, onboarding primeiros 5-10 beta customers (Q2 2026).

---

**Documento:** `/Users/antonio.ferreira/projects/pulse-service/docs/PRICING-STRATEGY.md`  
**Autor:** @finance agent  
**Status:** Draft v1.0 — Aguardando aprovação CEO (Antonio Pedro Ferreira)
