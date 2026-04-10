# Resumo Executivo — Pulse SaaS Pricing & Comercial

**Para:** Antonio Pedro Ferreira (CEO/CFO)  
**Data:** 10 de abril de 2026  
**De:** Agente Financeiro (SOUL)  
**Leitura:** 3 minutos

---

## I. ACHADOS CRÍTICOS

### 1. Pulse está SUBPRECIFICADO
- Starter R$299 é 40% abaixo do custo de servir (COGS R$318 em cloud puro)
- Mesmo em homelab, margin é apenas -6% (loss-leader não intencional)
- **Recomendação:** Elevar **Starter R$299 → R$399** (+33%) — ainda 30% abaixo VWO

### 2. Growth e Pro viáveis COM margem excelente
- Com homelab (70% reduction COGS) + shared infra IIT:
  - Growth R$799 → R$1.099: margin 35% → 75%
  - Pro R$1.999 → R$2.399: margin 47% → 70%
- LTV/CAC ratio: **64-79x** (benchmark SaaS = 3x+) ✅

### 3. Self-Hosted R$24k/ano está 75% abaixo mercado
- LD Enterprise: USD 24k-60k/ano (R$125k-312k)
- **Recomendação:** Elevar **R$24k → R$50k/ano (licença) + R$2-5k/mês (suporte)**
- Ainda competitivo; permite upsell path de suporte

### 4. Tributação clara: Simples Nacional Anexo III
- SaaS Cloud: Item 1401 NFS-e (5% ISS, 8% DAS)
- Self-Hosted: Item 1408 (licença software) + 1401 (suporte) — bifurcado
- **Ação:** Validar com contador Salvador sobre IN 08/2025 NFS-e nova

### 5. Pulse sozinho não salva runway
- **Year 1 esperado:** R$36k-375k ARR (10x spread)
- Contribui 1-5 meses de runway
- **Foco:** Pulse é plus, não savior. Prioridade IIT continua SocialMake + Libri

---

## II. DECISÃO DE PRICING — RECOMENDAÇÃO

### Estrutura Proposta

| Tier | Preço Novo | Margin | LTV 24m | Público |
|---|---|---|---|---|
| **Starter** | R$399/mês ↑ | 55% | R$9.576 | Startups Seed/A |
| **Growth** | R$1.099/mês ↑ | 75% | R$31.752 | Startups A/B |
| **Pro** | R$2.399/mês ↑ | 70% | R$62.376 | Mid-market |
| **Self-Hosted** | R$50k/ano + R$2-5k/mês suporte | 57-80% | R$240k+ | Enterprise |

**Impacto:**
- ARR multiplica ~1.5-2x com **mesmo volume de clientes**
- Breakeven operacional: mês 9-10 (vs 12 antes)
- Attracts quality customers (não price-sensitive)

---

## III. BKLs CRÍTICOS — BLOQUEADORES (5-10 dias)

| BKL | Bloqueador? | Por quê |
|---|---|---|
| **COMM-001: SDK IIFE + track()** | SIM ✅ | Sem isso, frontend é cego; nenhum evento chega |
| **COMM-002: p-value experimentos** | SIM ✅ | Sem significância, A/B é só split de tráfego |
| **COMM-003: Audit log flags** | SIM ✅ | Sem rastreabilidade, não vende para compliance |
| **COMM-004: Ambientes staging/prod** | SIM ✅ | LaunchDarkly baseline; sem isso, não competimos |

**Timeline:** 10 dias de dev (1 engenheiro). Depois posso vender com confiança.

---

## IV. GAPS DE MONETIZAÇÃO (Adicionar no Growth)

Concorrentes cobram (nós não estamos):

| Feature | LaunchDarkly | Pulse Proposto | Add-on preço |
|---|---|---|---|
| Enterprise SSO | $200/mês | +R$199/mês | Growth+ |
| Approval workflow | Pro+ | +R$399/mês | Pro+ |
| Audit logs | Enterprise | Incluído grátis | Diferencial |
| Webhooks | Enterprise | +R$149/mês | Growth+ |
| SLA 99.99% | Enterprise | +R$499/mês | Pro+ |

**Impacto:** 3-5 add-ons viáveis elevam LTV Growth de R$31.752 → R$40k+ | ARR multiplica 1.4x

---

## V. PROJEÇÃO 12 MESES

### Cenário Base (Realista)
- Mês 1-4: 0 pagos (produto phase)
- Mês 5-8: 10 Starter + 3 Growth = R$7.3k/mês
- Mês 9-12: 25 Starter + 10 Growth + 1 Pro + 1 Self-Hosted = R$20.3k/mês final
- **Year 1 ARR:** ~R$360.000 (com novo pricing)
- **Margin contribuição:** ~50% → +R$180k profit
- **Runway extension:** +4 meses

### Cenário Otimista (+70% conversão)
- Partnerships agências, viral growth, case studies
- Mês 12 final: 50 Starter + 15 Growth + 2 Pro + 1 SH = R$75k/mês
- **Year 1 ARR:** ~R$600.000
- **Runway extension:** +5 meses

### Cenário Pessimista (-50%)
- Competição, SDK gaps, market confusion
- Mês 12: 10 Starter + 2 Growth = R$6k/mês
- **Year 1 ARR:** ~R$72.000
- **Runway extension:** +1 mês

---

## VI. PRÓXIMAS AÇÕES (Semana 1)

### Aprovações Requeridas
1. **Pricing:** Starter R$399, Growth R$1.099, Pro R$2.399, Self-Hosted R$50k+suporte?
2. **Fiscal:** Validar estrutura NFS-e bifurcada com contador Salvador?
3. **Dev:** Alocação 1 dev por 10 dias (COMM-001 a COMM-004)?

### Timeline Executivo
- **Semana 1:** Aprovação pricing + começo COMM-001
- **Semana 2-4:** Completar COMM-002, COMM-003, COMM-004
- **Mês 2:** Publicar pricing page + vender primeiros 5 clientes Starter beta
- **Mês 3:** Case study cliente Série A
- **Mês 12:** 25-50 clientes pagos, primeira venda self-hosted

---

## VII. RISCO PRINCIPAL

**Risco:** SDK gaps (web funcional, React/RN) = nenhum cliente consegue usar Pulse  
**Probabilidade:** ALTA se não implementar COMM-001, COMM-009  
**Mitigação:** Priorizar COMM-001 nos próximos 10 dias; não vender antes disso

---

## VIII. CONCLUSÃO

**Pulse está posicionado para capturar 2-5% do mercado SaaS brasileiro (TAM ~R$1B)** através de:
1. **Pricing defensável:** 30-50% menor que LaunchDarkly/VWO
2. **Compliance nativo:** LGPD data residency — único valor SaaS BR
3. **Self-hosted enterprise:** Bloqueia saída para LD/VWO em grandes deals
4. **Integração IIT:** Ecossistema (auth, billing, customer) — network effect

**Reajuste de pricing (Starter +33%, Growth +38%, Self-Hosted +108%) é defensável e necessário** para viabilizar modelo; ainda mantém competitividade 30-50% abaixo.

**Breakeven:** Mês 9-10 com 10-15 clientes Growth + base de Free → contribui 4-5 meses de runway IIT.

---

**Status:** Aguardando aprovação CEO para começar COMM-001  
**Leitura completa:** `/Users/antonio.ferreira/projects/pulse-service/docs/ANALISE_FINANCEIRA_PROFUNDA_PULSE.md`  
**BKLs detalhados:** `/Users/antonio.ferreira/projects/pulse-service/docs/BACKLOG_COMERCIAL_PRICING_BKL.md`
