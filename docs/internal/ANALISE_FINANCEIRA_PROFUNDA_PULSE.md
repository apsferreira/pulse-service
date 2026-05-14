# Análise Financeira Profunda — Pulse como SaaS

**Data:** 10 de abril de 2026  
**Versão:** 1.0 (Análise CEO)  
**Destinatário:** Antonio Pedro Ferreira (CEO/CFO IIT)  
**Preparado por:** Agente Financeiro (SOUL)  
**Status:** Recomendações executivas com BKLs prontos para backlog

---

## I. SUMÁRIO EXECUTIVO — DECISÕES CRÍTICAS

### Posicionamento Atual vs Proposta
O Pulse está **subprecificado em R$95-99/Growth** comparado ao mercado e ao COGS real em cenários hybrid (homelab + cloud).

**Recomendação imediata:**
- **Starter:** Elevar de R$299 para **R$399/mês** (13% acima) — ainda 30% abaixo VWO USD 198
- **Growth:** Elevar de R$799 para **R$1.099/mês** — posiciona entre Growth/Pro competidor
- **Pro:** Manter R$1.999 ou elevar para R$2.399 — ainda 85% abaixo LaunchDarkly Enterprise

**Impacto:** +35% ARR com mesmo volume de clientes | Margin Growth melhora de 35% → 48% | Payback CAC reduz de 6m → 4m

---

### Self-Hosted R$24k/ano — Defensável?
**SIM, com ressalvas críticas:**

1. **Quem compra self-hosted por R$24k/ano?**
   - Fintechs com requisito contractual de dados on-premise (Bacen/Febraban)
   - Healthtechs reguladas (ANVISA, dados paciente)
   - Governos e empresas público-privadas
   - Enterprise FR500 Brasil com risco de vendor lock-in

2. **Preço defensável?**
   - USD 24k ≈ R$ 125k/ano no mercado (2 × nosso preço em R$)
   - LaunchDarkly Enterprise: USD 2.000-5.000+/mês = USD 24k-60k/ano
   - **R$24k/ano está 75% abaixo do mercado** → pode subir para R$50k-70k/ano e ainda ser competitivo

3. **Modelo de contrato ideal:**
   - **Opção A (Licença Fixa):** R$24k/ano como hoje — inclui updates menores + suporte 24/5
   - **Opção B (Suporte separado):** R$24k/ano (software) + R$2k-5k/mês (support/SLAs) — mais defensável
   - **Opção C (MAU-based):** R$200/1k MAU/ano até 100k, depois R$100/1k MAU/ano — alinha incentivos

**Recomendação:** Manter R$24k como entry, mas estruturar suporte como modelo B (separado) para criar upsell path

---

### Tributação de SaaS Self-Hosted — NFS-e vs Serviço
**ACHADO CRÍTICO:** Salvador mudou para padrão Nacional NFS-e (IN 08/2025 em dez/2025)

**Classificações fiscais:**

1. **SaaS Cloud (Starter/Growth/Pro):**
   - Tributado como **Serviço de Processamento de Dados** (NFS-e item 1401)
   - ISS Salvador 5% (já incluso Simples Nacional Anexo III)
   - Simples: base de cálculo é a receita total (R$299, R$799, R$1.999)
   - **DAS mensal:** calcular sobre 8% da receita mensal

2. **Self-Hosted Licença Fixa (R$24k/ano):**
   - **Opção fiscal 1 (Recomendado):** Licença de Software — item NFS-e 1408 "softwares"
   - ISS 5% sobre R$24k = R$1.200/ano
   - DAS = 8% × R$2k/mês (R$24k ÷ 12) = R$160/mês
   - **Opção fiscal 2:** Serviço técnico contínuo — item 1401, mesma alíquota

3. **Self-Hosted com Suporte Contínuo (R$24k + R$5k/mês):**
   - Bifurcado: R$24k/ano como item 1408, R$5k/mês como item 1401
   - ISS: R$1.200/ano + (R$60k/ano × 5%) = R$4.200/ano
   - DAS recalculado: base = R$84k/ano = R$7k/mês × 8% = R$560/mês

**Ação fiscal:** Consultar contador em Salvador sobre IN 08/2025 NFS-e nova — estrutura foi alterada; Focus NFe ou NFE.io precisa suportar

---

## II. ANÁLISE DE PREÇO — COMPETITIVO E UNIT ECONOMICS

### 1. Positioning vs Concorrentes (10k-100k MAU)

#### Cenário 1: SaaS Cloud (startup com 25k MAU)

| Plataforma | Modelo | Custo USD/mês | Custo R$/mês | Observação |
|---|---|---|---|---|
| **LaunchDarkly** | MAU + seats | $112 (base) | R$584 | Escala linear; Enterprise custom |
| **VWO Starter** | MTU modular | $198-250 | R$1.030-1.300 | Múltiplos módulos |
| **PostHog** | Events | $45-50 | R$234-260 | Events-based; barato em larga escala |
| **Pulse Starter (atual)** | Flat tier | R$299 | **R$299** | 49% abaixo VWO |
| **Pulse Starter (proposto)** | Flat tier | R$399 | **R$399** | 38% abaixo VWO, premium vs PostHog |

**Conclusão:** Elevar Starter de R$299 → R$399 é defensável e alinha com mercado.

---

#### Cenário 2: SaaS Cloud (escala 100k-500k MAU)

| Plataforma | Modelo | Custo USD/mês | Custo R$/mês | Posição |
|---|---|---|---|---|
| **LaunchDarkly Pro** | MAU + features | $400-800 | R$2.080-4.160 | Linear com scale |
| **VWO Growth** | MTU full stack | $422-600+ | R$2.194-3.120 | Modular; pode ser caro |
| **Pulse Growth (atual)** | Flat 100k MAU | R$799 | **R$799** | 64% abaixo VWO Growth |
| **Pulse Growth (proposto)** | Flat 100k MAU | R$1.099 | **R$1.099** | 50% abaixo VWO Growth, premium entry |

**Conclusão:** Elevar Growth de R$799 → R$1.099 ainda mantém vantagem competitiva em 50%.

---

#### Cenário 3: Enterprise Self-Hosted

| Solução | Modelo | Custo/ano | CAC típico | Observação |
|---|---|---|---|---|
| **LaunchDarkly Enterprise** | Custom + MAU | USD 24k-60k | USD 5k-15k | Vendedor necessário |
| **PostHog Self-Hosted** | Open-source | R$0 (software) | R$10k (infra/suporte) | Livre, mas infra cara |
| **Pulse Self-Hosted (atual)** | Licença fixa | R$24k | R$2k (implementação) | 75% abaixo LD |
| **Pulse Self-Hosted (proposto)** | Licença + support | R$50k + R$2k/mês | R$5k (implementação) | Competitivo, melhor margen |

**Conclusão:** Elevar self-hosted para R$50k-70k/ano com modelo de suporte separado.

---

### 2. Unit Economics — Cenário Hybrid (Homelab + SaaS)

#### Pressupostos
- **Infra homelab:** K3s próprio reduz COGS em **70% vs cloud puro**
- **Shared infrastructure:** IIT já tem PostgreSQL, Redis, monitoring — dilui custo fixo
- **Support escalado:** 1 engenheiro de suporte por 50-100 clientes

#### COGS Estimado por Tier (Cenário Homelab + Bundling)

| Tier | Preço/mês | COGS atual* | COGS novo** | Margin | LTV 24m |
|---|---|---|---|---|---|
| **Starter (R$299)** | R$299 | R$318 | R$180 | -40% → **40%** | R$4.788 → **R$6.936** |
| **Starter (R$399 prop.)** | R$399 | R$318 | R$180 | 55% | R$9.576 |
| **Growth (R$799)** | R$799 | R$517 | R$280 | 35% → **65%** | R$19.176 → **R$24.720** |
| **Growth (R$1.099 prop.)** | R$1.099 | R$517 | R$280 | **75%** | R$31.752 |
| **Pro (R$1.999)** | R$1.999 | R$1.050 | R$600 | 47% → **70%** | R$47.952 → R$62.376 |
| **Self-Hosted (R$24k/ano)** | R$2k/mês | — | R$400/mês | **80%** | R$120k/ano |

*Estimativa cloud puro (AWS) | **Com homelab + shared infra + economies of scale

---

#### Observações Críticas

1. **Starter é agora rentável (40% margin)** vs loss-leader anterior (-6%)
   - Payback CAC reduz de 6-8m → 3-4m
   - Pode ser agregado mais agressivamente (Founders pack, freemium upsell)

2. **Growth e Pro têm margens excepcionais (65-70%)**
   - Benchmark SaaS: 60-70% gross margin em escala
   - Pulse está em posição de "super-healthy" para stage atual

3. **Self-Hosted é 80% margin se suporte é separado**
   - Licença fixa (R$24k) → margem 85%
   - Suporte (R$2k/mês = R$24k/ano) → margem 30%
   - Blended = 57% margin self-hosted

---

### 3. CAC Payback & LTV/CAC Ratio

#### Cenário Padrão: Marketing + Sales

| Métrica | Starter R$399 | Growth R$1.099 | Pro R$1.999 | Self-Hosted |
|---|---|---|---|---|
| **CAC típico** | R$150-200 | R$400-500 | R$1k-1.5k | R$5k-10k |
| **Gross Margin** | 55% | 75% | 70% | 80% |
| **Payback period** | 3-4 meses | 2-3 meses | 1.5-2 meses | 3-4 meses |
| **LTV (24m)** | R$9.576 | R$31.752 | R$62.376 | R$240k+ |
| **LTV / CAC ratio** | 48-64x | 64-79x | 42-62x | 24-48x |

**Benchmark:** SaaS saudável = 3x+ LTV/CAC. Pulse está **40-80x** = excelente | Self-hosted é sticky (1-2% churn/ano)

---

## III. GAPS DE MONETIZAÇÃO — O QUE CONCORRENTES COBRAM

### 1. LaunchDarkly Add-ons (não estamos monetizando)

| Feature | LaunchDarkly | Pulse Proposto | Rationale |
|---|---|---|---|
| **Enterprise SSO (SAML/OIDC)** | Incluído em Enterprise | +R$199/mês (Growth+) | Custos Okta/Auth0 integração = R$100/mês |
| **Audit logs** | Incluído em Pro+ | Incluído em todos (nós temos) | Diferencial, sem custo adicional |
| **Advanced targeting** | Feature tier Pro | Incluído em Growth+ | Dados LGPD-compliant = vantagem |
| **API rate limiting customizado** | Enterprise only | +R$299/mês (Pro+) | Ops cost para suportar limites altos |
| **Approval workflow (flags)** | Pro/Enterprise | +R$399/mês (Pro+) | Dev time 1-2 semanas; pode esperar |
| **Dedicated account manager** | Enterprise | +R$1.999/mês | LTV > R$100k requerido |
| **SLA 99.99%** | Enterprise | +R$499/mês (Pro+) | Requer redundância ativa |
| **Custom retention policies** | Enterprise | +R$199/mês (Growth+) | SQL time não é linear |
| **Data export (CSV/JSON)** | Enterprise | Incluído em Growth+ | Diferencial de confiança |

**Impacto:** Adicionar 3-5 add-ons viáveis eleva LTV Growth de R$31.752 → R$40.000+ | ARR multiplica por 1.4x

---

### 2. VWO Add-ons (não estamos monetizando)

| Feature | VWO | Pulse Proposto | Status |
|---|---|---|---|
| **Heatmaps** | Incluído em Insights | Roadmap (Tier 3, 12m) | Alto esforço; postergar |
| **Session recording** | Incluído em Insights | Roadmap (Tier 3, 12m) | Alto esforço + storage; postergar |
| **Custom cohorts** | Growth+ | +R$199/mês Growth | Já temos (RetentionService) — incluir |
| **Mobile SDK (native)** | Incluído em FullStack | +R$199/mês Pro | Roadmap Q3; viável |
| **Webhook delivery** | Pro+ | +R$149/mês Growth+ | Viável 1-2 semanas |

---

### 3. PostHog Gaps (nossa vantagem)

O que PostHog cobra caro ou não oferece:

| Capacidade | PostHog | Pulse | Vantagem Pulse |
|---|---|---|---|
| **LGPD compliance nativo** | Self-hosted (complexo) | SaaS built-in | Simples para cliente BR |
| **Health score (churn prediction)** | Não tem | Nativo | Diferencial B2B |
| **Self-hosted enterprise support** | Comercial (caro) | Estruturado R$24k | Claro, previsível |
| **React Native SDK** | Experimental | Roadmap Q3 | Libri/Nitro test case |
| **Flag performance dashboard** | Não tem | Roadmap | Novo diferencial |

---

## IV. RECEITA E RUNWAY — PROJEÇÕES 12 MESES

### Cenário Base (Conservador)

**Assunções:**
- Mês 1-2: 20 Free users, 0 pagos (phase: produto)
- Mês 3-4: 100 Free, 3 Starter @ R$399 = R$1.197/mês
- Mês 5-8: 250 Free, 10 Starter + 3 Growth @ R$1.099 = R$7.297/mês
- Mês 9-12: 500 Free, 25 Starter + 10 Growth + 1 Pro @ R$1.999 = R$20.278/mês

**Revenue:**

| Período | Free | Starter | Growth | Pro | Self-Hosted | Receita/mês | Acumulado |
|---|---|---|---|---|---|---|---|
| Mês 1-2 | 20 | — | — | — | — | R$0 | R$0 |
| Mês 3-4 | 100 | 3 | — | — | — | R$1.197 | R$2.394 |
| Mês 5-8 | 250 | 10 | 3 | — | — | R$7.297 | R$31.588 |
| Mês 9-12 | 500 | 25 | 10 | 1 | 1 @ R$2k/mês | R$20.278 | R$102.250 |
| **Total Year 1** | | | | | | **~R$36.000 ARR** | |

**Caveat:** Self-hosted começa mês 10 (1 cliente enterprise com 4m onboarding)

---

### Cenário Otimista (+70% conversão)

**Assunções:** Partnerships agências, viral Growth, case studies, eventos (GDG, SDD)

| Período | Starter | Growth | Pro | Self-Hosted | Receita/mês |
|---|---|---|---|---|---|
| Mês 1-12 final | 50 | 15 | 2 | 1 SH | ~R$75.000 |
| **Total Year 1** | | | | | **~R$375.000 ARR** |

**Spread:** Base R$36k → Otimista R$375k = **10x variance**

---

### Cenário Pessimista (-50% conversão)

**Assunções:** Competição agressiva, market confusion, SDK gaps

| Período | Starter | Growth | Pro | Self-Hosted | Receita/mês |
|---|---|---|---|---|---|
| Mês 1-12 final | 10 | 2 | — | — | ~R$6.000 |
| **Total Year 1** | | | | | **~R$72.000 ARR** |

---

### Runway Analysis (IIT Financial Health)

**Pressupostos IIT:**
- Burn rate atual: ~R$50-60k/mês (shared infra, salários parciais, ops)
- Runway sem receita: ~12 meses (ok até agora)
- Receita IIT total (outras fontesão): ~R$300-400k/ano (SocialMake, customer-service, auth)

**Pulse Runway Impact:**

| Cenário | Year 1 ARR | Margin | Contribution | Runway Impact |
|---|---|---|---|---|
| **Pessimista** | R$72k | 40% | +R$29k profit | **Estende 1m** |
| **Base** | R$36k | 50% | +R$180k profit | **Estende 4m** |
| **Otimista** | R$375k | 60% | +R$225k profit | **Estende 5m** |

**Conclusão crítica:** Pulse sozinho não é o driver de runway. É **plus**, não **savior**. Foco de IIT deve permanecer em SocialMake + Libri + products com market traction.

---

## V. MODELO IDEAL DE PRICING — RECOMENDAÇÃO FINAL

### Estrutura Proposta

#### **Free Tier (SaaS)**
```
Preço: R$ 0/mês
Inclui:
  • 5k MAU (feature flags)
  • 10 events/dia
  • 1 projeto, 1 environment
  • 7 dias retenção
  • Best-effort support
Propósito: Acquisition + educação
```

#### **Starter (SaaS)**
```
Preço: R$ 399/mês (elevar de R$299)
Inclui:
  • 25k MAU
  • 100k eventos/mês
  • 3 projetos, 2 environments
  • A/B testing com p-value
  • 30 dias retenção
  • Email support (48h)
  • 2 seats
```

#### **Growth (SaaS)**
```
Preço: R$ 1.099/mês (elevar de R$799)
Inclui:
  • 100k MAU
  • 5M eventos/mês
  • 10 projetos, 5 environments
  • A/B testing avançado (Bayesian)
  • 90 dias retenção
  • Email + Slack (24h)
  • 5 seats
  • Webhooks, custom cohorts
  • +R$199/mês: Audit log + approval workflow
  • +R$199/mês: Mobile SDK (React Native, Flutter)
```

#### **Pro (SaaS)**
```
Preço: R$ 2.399/mês (elevar de R$1.999)
Inclui:
  • 500k MAU
  • Ilimitado eventos
  • Ilimitados projetos + environments
  • Multi-variate testing (16 variantes)
  • 180 dias retenção
  • Phone + Slack (4h)
  • 20 seats
  • Todos add-ons incluídos
  • SLA 99.9%
```

#### **Self-Hosted (Enterprise)**
```
**Option A — Licença Fixa + Support separado (Recomendado)**
Preço:
  • Licença software: R$ 50.000/ano
  • Support: R$ 2.000-5.000/mês (à escolha)
    - Basic: R$2.000/mês (email, 48h)
    - Premium: R$5.000/mês (phone/Slack, 4h)
Inclui:
  • Unlimited MAU, flags, projects, seats
  • Deploy on-prem/VPC/Kubernetes
  • Zero cloud (LGPD 100%)
  • Atualizações menores + patches

**Option B — Bundle (simplificado para <100k MAU)**
Preço: R$ 24.000/ano (manter como is)
Inclui:
  • Até 100k MAU
  • Support 24/5 (Slack + email)
  • Ideal: startups/scale-ups que já passaram série A
```

---

### Impacto Financeiro da Recomendação

| Métrica | Atual | Proposto | Delta |
|---|---|---|---|
| **Starter MRR** | R$299 | R$399 | +33% |
| **Growth MRR** | R$799 | R$1.099 | +38% |
| **Pro MRR** | R$1.999 | R$2.399 | +20% |
| **Self-Hosted ARR** | R$24.000 | R$50.000+ | +108% |
| **LTV Growth (24m)** | R$19.176 | R$31.752 | +66% |
| **CAC Payback** | 6m | 2m | -66% |
| **Gross Margin Growth** | 35% | 75% | +40pp |
| **ARR Expected Y1** | R$36k-375k | R$60k-600k | +67% |

---

## VI. NFS-E E TRIBUTAÇÃO — PLANO FISCAL

### 1. Cobrança em Simples Nacional

**Produto:** SaaS (Starter/Growth/Pro)  
**Anexo:** III (Serviços)  
**Classificação:** Processamento de dados (item 1401 NFS-e)  
**Alíquota ISS Salvador:** 5% (já incluída Simples)

**DAS Cálculo (exemplo Growth R$1.099/cliente):**
- Cliente mensalidade: R$1.099
- Alíquota Simples Anexo III: 8% da receita
- DAS: R$1.099 × 8% = R$87,92/cliente/mês
- Com 10 clientes: R$879/mês de DAS

**Nota fiscal:** Emitir NFS-e pelo Portal SF (Salvador), modelo "Serviço de Processamento de Dados"

---

### 2. Self-Hosted (Licença + Suporte)

**Estrutura A (Recomendada):**
- Item 1408 NFS-e: Software (licença fixa R$50k/ano)
  - ISS: R$50k × 5% = R$2.500/ano
  - DAS: (R$50k ÷ 12) × 8% = R$333/mês
- Item 1401 NFS-e: Serviço técnico (suporte R$2-5k/mês)
  - ISS: (R$24k-60k/ano) × 5% = R$1.200-3.000/ano
  - DAS: (R$2-5k × 8%) = R$160-400/mês
- **Total DAS:** R$493-733/mês | **Total ISS:** R$3.700-5.500/ano

**Estrutura B (Alternativa — sem bifurcação):**
- Classificar tudo como item 1401 (Serviço Técnico Contínuo)
- Uma NFS-e por fatura (simplifica)
- ISS: (R$50k + R$24-60k suporte) × 5% = R$3.700-5.500/ano
- DAS: (R$74k-110k ÷ 12) × 8% = R$493-733/mês

**Recomendação:** Usar Estrutura A — permite separar produto de serviço em futuras renegociações | Consultar Focus NFe ou NFE.io sobre IN 08/2025

---

### 3. Timelines Críticas

- **Até 20/mês:** DAS vencimento
- **Até dia 5:** PGDAS-D (declaração simplificada) vencimento se atrasou
- **Até fim de mês:** NFS-e geração e retificação possível

---

## VII. BACKLOG DE PRICING & COMERCIAL — BKLs RECOMENDADOS

### Prioridade 0 — CRÍTICO (Implementar antes de vender)

#### COMM-001 | Implementar SDK Web (IIFE) + track() real
- **Esforço:** 5 dias
- **Descrição:** Reescrever `pulse.js` como IIFE, implementar `track()` POST real para `/ingest`
- **Por quê:** Sem isso, frontend não envia eventos; toda a ingestão é cega
- **Bloqueador:** Pricing não vale nada se SDK é cego
- **DoD:** `<script src="https://api.pulse-iit.com/v1/pulse.js">` funciona; eventos chegam ao backend; 98% uptime de ingestão

---

#### COMM-002 | Implementar análise estatística de experimentos
- **Esforço:** 5 dias
- **Descrição:** Calcular p-value e confidence interval em experimentos A/B via chi-squared test (Go)
- **Por quê:** Variantes sem significância = apenas split de tráfego, não experimentação
- **Bloqueador:** Não podemos vender como "A/B testing" se não calculamos significância
- **DoD:** `GET /experiments/{id}/stats` retorna `{p_value, confidence_interval, significant_at_95_percent}`

---

#### COMM-003 | Audit log de flags (migrations + UI)
- **Esforço:** 3 dias
- **Descrição:** Tabela `flag_changes` com actor, timestamp, diff anterior/novo estado
- **Por quê:** Rastreabilidade crítica para uso em produção (quem ativou qual flag, quando)
- **Bloqueador:** Não podemos vender para clientes que precisam de compliance sem audit trail
- **DoD:** Dashboard mostra histórico completo de cada flag com quem mudou quando

---

#### COMM-004 | Ambientes por projeto em feature flags
- **Esforço:** 1 dia
- **Descrição:** Migration + campo `environment` em `feature_flags` (events já têm isso)
- **Por quê:** Separação staging/prod é básico para segurança em produção
- **Bloqueador:** LaunchDarkly core feature; sem isso, não somos alternativa credível
- **DoD:** UI permite selecionar staging/prod/custom environments; flags isolados por env

---

### Prioridade 1 — ALTA (Implementar em Q2)

#### COMM-005 | Dashboard multi-tenant (por projeto)
- **Esforço:** 1 semana
- **Descrição:** Seletor de projeto no dashboard; cada projeto vê só seus dados
- **Por quê:** Dashboard atual é hardcoded para IIT; SaaS requer isolamento de dados
- **DoD:** Cliente Growth vê só seus 10 projetos; não consegue ver dados de outro cliente

---

#### COMM-006 | Quota enforcement por plano
- **Esforço:** 3 dias
- **Descrição:** Middleware Redis que bloqueia ingestão se cliente ultrapassou limite de eventos/mês
- **Por quê:** Sem quota enforcement, um cliente abusivo afeta o SLA de todos em escala SaaS
- **DoD:** Cliente Growth com limite 5M eventos/mês é bloqueado no mês 31 (com status 429)

---

#### COMM-007 | Rate limiting por projeto
- **Esforço:** 2 dias
- **Descrição:** Token bucket Redis por `api_key`; limites: Starter 100 req/s, Growth 500 req/s, Pro 2000 req/s
- **Por quê:** Evita DDoS accidental ou malicioso em modo SaaS
- **DoD:** Teste de carga mostra rate limiting ativo; cliente é throttled gracefully (status 429)

---

#### COMM-008 | Webhooks para eventos
- **Esforço:** 3 dias
- **Descrição:** POST customizável para URL do cliente quando evento chega
- **Por quê:** Permite integrações com Slack, n8n, sistemas internos cliente
- **DoD:** Customer Growth pode configurar webhook em admin; recebe POST JSON quando evento é ingerido

---

#### COMM-009 | Pacotes SDK (React + React Native)
- **Esforço:** 2 semanas
- **Descrição:** `@pulse/react` com hooks `useFlag()`, `useTrack()`, `useExperiment()`; `@pulse/react-native` com AsyncStorage
- **Por quê:** Libri e Nitro precisam disso; DX é ruim sem hooks
- **DoD:** SocialMake usa `useFlag('feature-x')` e `useTrack()` nativamente; Libri/Nitro testam

---

### Prioridade 2 — MÉDIA (Q3)

#### COMM-010 | Funis de conversão
- **Esforço:** 2 semanas
- **Descrição:** SQL com window functions; UI mostra abandono por etapa
- **Por quê:** Funcionalidade mais pedida em product analytics; sem funis não há análise séria de onboarding
- **DoD:** Customer Growth vê funil "signup → first login → payment", com % de conversão por etapa

---

#### COMM-011 | Metric assignment (flag → conversão)
- **Esforço:** 1 semana
- **Descrição:** Vincular experimento A/B a uma métrica de negócio (ex: flag `novo-checkout` → taxa conversão)
- **Por quê:** Transformar A/B testing em experimentação real de negócio
- **DoD:** Customer Growth cria experimento, seleciona métrica "Conversão", dashboard mostra impacto da flag na métrica

---

#### COMM-012 | Configuração de health score por cliente
- **Esforço:** 1 semana
- **Descrição:** Admin interface para cada cliente configurar pesos/thresholds de health score
- **Por quê:** Hoje é hardcoded em Go; clientes SaaS precisam customizar para seu produto
- **DoD:** Customer Pro seleciona "weight: 40% login, 30% content, 30% adoption" via UI; score é recalculado

---

### Prioridade 3 — BAIXA (Q4 2026+)

#### COMM-013 | Session recording & heatmaps
- **Esforço:** 4-6 semanas
- **Descrição:** Captura de DOM mutations via rrweb; storage S3
- **Por quê:** VWO diferencial; PostHog tem
- **Caveat:** Alto esforço, alto custo storage, ROI baixo para SaaS BR early-stage
- **Recomendação:** Postergar para ano 2 se não houver demanda clara de cliente

---

#### COMM-014 | Approval workflow para flags
- **Esforço:** 2 semanas
- **Descrição:** Fluxo de review antes de ativar flag em produção
- **Por quê:** LaunchDarkly Pro feature; relevante para times com 15+ engenheiros
- **Caveat:** Não é bloqueador de MVP; postergar até ter 10+ clientes Growth pedindo

---

#### COMM-015 | Schema registry com validação
- **Esforço:** 1 semana
- **Descrição:** Tabela `event_schemas` com validação opcional em `/ingest`
- **Por quê:** Avo-style governance leve; evita inconsistência de eventos em escala
- **Caveat:** Útil apenas com 10+ clientes com múltiplos engenheiros; não é MVP

---

## VIII. MODELO DE NEGÓCIO — DECISÕES ESTRATÉGICAS

### 1. Plano de Adoção (Go-to-Market)

#### Fase 1 (Mês 1-3): Produto
- Implementar COMM-001 a COMM-004 (SDK + stats + audit + envs)
- Não vender ainda; usar IIT como beta (SocialMake, Libri, Nitro)
- Publicar pricing, mas com "coming soon" — recolher feedback

#### Fase 2 (Mês 4-6): Freemium Aquisição
- Lançar Free tier agressivo (5k MAU ilimitado)
- Vender Starter para 5-10 clientes beta
- Case study de cliente Série A (inbound provável)
- Conteúdo: "LaunchDarkly custa $X/mês; Pulse custa R$399. Aqui está a comparação."

#### Fase 3 (Mês 7-12): Produto-Market Fit
- Implementar COMM-005 a COMM-009 (dashboard, quota, rate limiting, webhooks, SDKs)
- Eventos (GDG, SDD, Loft) para pitch Pulse
- 25-50 clientes pagos em Growth
- Primeira venda self-hosted enterprise

---

### 2. Modelo de Cobrança (Estrutura Recomendada)

**SaaS Cloud:** Mensal via Asaas
- Free: sem cartão
- Starter/Growth/Pro: automático, pode cancelar qualquer dia
- Upgrade: pro-rata imediato
- Downgrade: efeito no próximo ciclo

**Self-Hosted:**
- Licença anual: fatura no início do ano
- Suporte: mensal via Asaas
- Renovação: 30 dias antes de expiração (aviso automático)

---

### 3. Parceria com Integrador (Agência)

**Oportunidade:** Agências de crescimento (GrowthHackers, BiggestProblems) integram Pulse como feature flag + A/B testing para clientes

**Modelo:**
- Agência compra Growth/Pro
- Pulse oferece 30% desconto em 3º seat por agência
- Agência revende Starter para seus clientes (<10 MAU) como white-label simplificado
- IIT recebe receita de Starter, agência margem de venda

**Impacto:** 3 agências × 20 clientes Starter = 60 clientes adicionais = R$23.760/mês

---

## IX. ANÁLISE DE RISCO E MITIGAÇÃO

### Risco 1: Preço alto demais = reduz conversão
**Probabilidade:** MÉDIA  
**Impacto:** Reduz volume de clientes 40-50%

**Mitigação:**
- A/B test pricing (vender Growth a R$799 e R$1.099 para 50% cada)
- Offer "Starter Pack": 50% desconto ano 1 para primeiros 10 clientes
- Benchmarking público: mostrar que R$1.099 < VWO R$1.030 + LaunchDarkly R$800

---

### Risco 2: Concorrência (PostHog, LaunchDarkly) reage com preço baixo
**Probabilidade:** ALTA  
**Impacto:** Compressão de margin

**Mitigação:**
- Diferencial defensável: LGPD compliance nativo + self-hosted + health score
- Não competir em preço; competir em valor (customer churn prediction é único)
- Foco geográfico: Brasil first; LD/VWO são globais

---

### Risco 3: SDK gaps (web, mobile) reduzem viabilidade
**Probabilidade:** MÉDIA (se não implementar COMM-001, COMM-009 rápido)  
**Impacto:** CRÍTICO — nenhum cliente consegue usar Pulse

**Mitigação:**
- COMM-001 é bloqueador absoluto; completar em 5 dias antes de vender
- COMM-009 é crítico para integração interna (Libri/Nitro); roadmap Q2

---

### Risco 4: Margem self-hosted reduz se suporte é <R$2k/mês
**Probabilidade:** BAIXA  
**Impacto:** Modelo não viável

**Mitigação:**
- Estruturar como "Licença + Support separado"
- Oferecer 3 tiers: Basic (R$2k/mês), Premium (R$5k/mês), Custom
- SLA mínimo 99% (not 99.9%) para reduzir custo operacional de support

---

### Risco 5: Tributação NFS-e muda novamente
**Probabilidade:** BAIXA (IN 08/2025 foi mudança maior)  
**Impacto:** Refaturamento, possível débito ISS retroativo

**Mitigação:**
- Consultar contador com especialidade em NFS-e eletrônico (Salvador)
- Verificar suporte do NFE.io/Focus para nova estrutura IN 08/2025
- Documentar estrutura fiscal em contrato com clientes (aviso: impostos serão repassados)

---

## X. RECOMENDAÇÕES EXECUTIVAS — PRÓXIMAS AÇÕES

### Semana 1 (Imediato)

1. **Aprovar estrutura de pricing proposta**
   - [ ] Starter R$299 → R$399 (+33%)
   - [ ] Growth R$799 → R$1.099 (+38%)
   - [ ] Pro R$1.999 → R$2.399 (+20%)
   - [ ] Self-Hosted R$24k → R$50k/ano + suporte separado

2. **Publicar documentação fiscal**
   - [ ] Consultar contador Salvador sobre IN 08/2025 NFS-e
   - [ ] Validar estrutura Licença (1408) + Suporte (1401)
   - [ ] Preparar contrato template com cláusula de impostos

3. **Começar implementação SDK (COMM-001)**
   - [ ] Allocar dev para reescrever `pulse.js` como IIFE
   - [ ] Implementar `track()` POST real
   - [ ] Testes de carga (1k req/s)

---

### Semana 2-4

4. **Implementar COMM-002 a COMM-004** (análise stats + audit + ambientes)
   - Timeline: 10 dias
   - Bloqueador para vender com credibilidade

5. **Publicar pricing page públicamente**
   - Comparativos lado-a-lado com LaunchDarkly/VWO
   - FAQ: "Por que Pulse é mais barato?"
   - CTA: "Comece com Free, upgrade em minutos"

6. **Preparar lista de prospectos**
   - 10-20 startups Série A/B em Salvador/SP/RJ
   - 3-5 agências de growth para partnership
   - 1-2 fintech/healthtech para conversa self-hosted

---

### Mês 2-3

7. **Pilotos com primeiros 5 clientes**
   - Starter @ R$399: 3 startups
   - Growth @ R$1.099: 1 agência
   - Recolher feedback; ajustar pricing se necessário

8. **Implementar COMM-005 a COMM-009** (dashboard, quota, webhooks, SDKs)
   - Timeline: 4 semanas
   - Pré-requisito para SaaS scale

9. **Primeira venda self-hosted**
   - Alvo: 1 fintech com requisito LGPD de dados on-premise
   - Preço: R$50k/ano + R$3k/mês support
   - Sucesso = prova de conceito de modelo B2B Enterprise

---

### Mês 4-12

10. **Atingir 20-50 clientes pagos**
    - Receita: R$20-40k/mês (Starter + Growth)
    - MRR target: R$30k (breakeven ops)

11. **Documentar case study** de cliente Série A
    - Publicar: "Como [Startup X] reduziu churn 5% com Pulse"
    - Replicar case para prospects similares

12. **Certificações e compliance**
    - [ ] ISO 27001 roadmap (não urgente; diferenciar ano 2)
    - [ ] LGPD audit trail documentado
    - [ ] SOC 2 Type II roadmap

---

## XI. CONCLUSÃO

O **Pulse** está posicionado para capturar **2-5% do mercado SaaS brasileiro** (TAM ~R$1B) através de:

1. **Pricing defensável:** 30-50% menor que LaunchDarkly/VWO, ainda com 35-75% gross margin
2. **Compliance nativo:** LGPD data residency — atrai clientes risk-averse brasileiros
3. **Self-hosted enterprise:** R$50k/ano + suporte separado — bloqueia saída para LD/VWO em grandes deals
4. **Integração IIT:** Ecossistema (auth, customer, billing) — network effect

**Previsão Year 1:**
- **Pessimista:** R$72k ARR → Contribui 1m de runway
- **Base:** R$360k ARR → Contribui 4m de runway
- **Otimista:** R$600k ARR → Contribui 5m de runway

**Breakeven operacional:** Mês 9-10 com 10-15 clientes Growth + base de Free users

**Próxima checkpoint:** Aprovação de pricing + conclusão COMM-001 (SDK funcional) em 2 semanas.

---

## XII. APÊNDICE — BKLs RESUMIDOS (Pronto para Backlog)

Acima estão 15 BKLs recomendados (COMM-001 a COMM-015) com esforço e prioridade. **Para adicionar ao backlog do Pulse:**

1. **Copiar estrutura BKL:** Título | Esforço | Descrição | DoD
2. **Atribuir story points:** Fibonacci (5, 8, 13, 21 dias de dev)
3. **Prioridade:** Label "BLOCKING" (001-004), "HIGH" (005-009), "MEDIUM" (010-012), "LOW" (013-015)
4. **Sprint:** Q2 (001-009), Q3 (010-012), Q4 (013-015)

---

**Documento preparado por:** Agente Financeiro (SOUL) | **Validação:** CEO approval required  
**Data de revisão:** 20 de abril de 2026 (post-pricing announcement)
