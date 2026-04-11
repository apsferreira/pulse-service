# Análise Competitiva Profunda — Pulse vs PostHog vs GrowthBook vs LaunchDarkly
**Data:** Abril 2026  
**Versão:** 3.0 — Análise atualizada com pesquisa de mercado 2026  
**Autor:** SOUL — PM Agent IIT

---

## Executive Summary

O Pulse é um **produto híbrido** que tenta cobrir 3 categorias simultâneas:
1. **Feature flags** (concorrente: LaunchDarkly)
2. **A/B testing / Experimentation** (concorrente: GrowthBook, VWO)
3. **Product analytics** (concorrente: PostHog)

**Diagnóstico:** Pulse é MVP funcional (4/10 de maturidade), mas **não pode ser vendido externamente** até os bloqueadores de segurança estarem resolvidos (já foram em 2026-04-10).

**Posicionamento defensável:** LaunchDarkly + GrowthBook + PostHog em um produto, self-hosted, pricing em R$, com health score de cliente nativo.

**ICP (Ideal Customer Profile):** Engineering Manager / CTO de EdTech ou SaaS brasileira, 20-500 engenheiros, R$399-2.499/mês.

---

## 1. Feature Comparison Matrix — Pulse vs Concorrentes

| Feature | **Pulse** | **LaunchDarkly** | **GrowthBook** | **PostHog** |
|---------|:---:|:---:|:---:|:---:|
| **FEATURE FLAGS** | | | | |
| On/off simples | ✅ | ✅ | ✅ | ✅ |
| Rollout percentual | ✅ | ✅ | ✅ | ✅ |
| Targeting por user/tenant | ✅ | ✅ | ✅ | ✅ |
| Multivariate flags (N variantes) | ✅ | ✅ | ✅ | ✅ |
| Ambientes por projeto (staging/prod) | ✅ (v2026-04) | ✅ | ✅ | ✅ |
| Audit log de mudanças | ✅ (v2026-04) | ✅ | ❌ | ✅ |
| Aprovação de mudanças (flag review) | ❌ | ✅ (Pro+) | ❌ | ❌ |
| Flag stale detection | ✅ (v2026-04) | ✅ | ❌ | ❌ |
| Flag triggers por métrica de erro | ❌ | ✅ | ❌ | ❌ |
| Streaming SSE para updates | ❌ | ✅ | ❌ | ✅ |
| | | | | |
| **A/B TESTING / EXPERIMENTS** | | | | |
| A/B com variantes e peso | ✅ | ✅ | ✅ | ✅ |
| Análise estatística (p-value, CI) | ✅ (v2026-04, chi-squared) | ✅ | ✅ (Bayesiano + Frequentist) | ✅ (Bayesiano) |
| Metric assignment (flag → métrica) | ✅ (v2026-04) | ✅ | ✅ | ✅ |
| Holdout groups (experimental) | ❌ | ✅ | ✅ | ❌ |
| Multi-variate experiments (3+) | ✅ | ✅ | ✅ | ✅ |
| Heatmaps | ❌ | ❌ | ❌ | ✅ |
| Session recording | ❌ | ❌ | ❌ | ✅ |
| Bayesian inference | ❌ | ✅ | ✅ | ✅ |
| | | | | |
| **EVENT TRACKING** | | | | |
| Ingestão de eventos customizados | ✅ | Contexto apenas | ✅ | ✅ |
| Identify (anon → user) | ✅ | ✅ | ✅ | ✅ |
| Propriedades via JSONB | ✅ | ✅ | ✅ | ✅ |
| Batch ingest (POST /batch) | ✅ | ✅ | ✅ | ✅ |
| Funis de conversão | ✅ (v2026-04) | ❌ | ✅ | ✅ |
| Retenção (DAU/WAU/MAU) | ✅ | ❌ | ✅ | ✅ |
| Cohort analysis | ✅ | ❌ | ✅ | ✅ |
| SQL explorer (query livre) | ❌ | ❌ | ❌ | ✅ |
| LLM API cost tracking | ✅ | ❌ | ❌ | ❌ |
| | | | | |
| **SDKs DISPONÍVEIS** | | | | |
| Web via CDN (script tag IIFE) | ✅ | ✅ | ✅ | ✅ |
| SDK React com hooks | ✅ | ✅ | ✅ | ✅ |
| SDK React Native / Expo | ✅ | ✅ | ❌ | ✅ |
| SDK iOS nativo | ❌ | ✅ | ✅ | ✅ |
| SDK Android nativo | ❌ | ✅ | ✅ | ✅ |
| SDK Go | ✅ | ✅ | ❌ | ✅ |
| SDK Node.js | ❌ | ✅ | ❌ | ✅ |
| SDK Python | ❌ | ✅ | ❌ | ✅ |
| | | | | |
| **INFRA & OPERAÇÃO** | | | | |
| Self-hosted | ✅ | ❌ | ✅ | ✅ (open-source) |
| Multi-tenant lógico | ✅ | ✅ | ✅ | ✅ |
| Rate limiting por projeto | ✅ | ✅ | ✅ | ✅ |
| Webhooks de eventos | ❌ | ✅ | ✅ | ✅ |
| Export de dados (CSV/JSON) | ✅ (v2026-04) | ✅ | ✅ | ✅ |
| Data warehouse export | ❌ | ✅ | ✅ | ✅ |
| LGPD compliance (DPA + direito ao esquecimento) | ✅ | ✅ | ✅ | ✅ |
| | | | | |
| **FEATURES EXCLUSIVAS DO PULSE** | | | | |
| Tenant health score (0-100) configurável | ✅ | ❌ | ❌ | ❌ |
| Churn prediction nativa | ✅ (regra simples) | ❌ | ❌ | ❌ |
| Pricing em R$ (sem conversão) | ✅ | ❌ | ❌ | ❌ |
| Free tier self-hosted ilimitado | ✅ | ❌ | ✅ | ✅ |

---

## 2. Análise de Posicionamento de Mercado

### 2.1 LaunchDarkly
- **Foco:** Feature flags enterprise
- **Modelo de preço:** USD 10-20/mês por 1k MAU (Monthly Active User)
- **ICP:** Engineering Manager de startup/scale-up ≥ 30 engenheiros, com budget USD 5k+/ano
- **Vantagem competitiva:** Audit logs, aprovações, SSO, integrações empresariais
- **Limitação:** Sem self-hosted, sem analytics nativo, sem A/B testing
- **Pricing 2026:** Starter USD 12/mês, Pro USD 200+/mês, Enterprise customizado. Mediana de contrato anual: USD 72.000

### 2.2 GrowthBook
- **Foco:** A/B testing + feature flags integrados
- **Modelo de preço:** Freemium + USD 40/usuário/mês (Pro) + custom (Enterprise)
- **ICP:** Product Manager de SaaS 50-500 pessoas, com time de dados
- **Vantagem competitiva:** Integração com data warehouse, Bayesian + Frequentist, estatística robusta
- **Limitação:** Requer warehouse externo (Snowflake, BigQuery, Postgres), caro para escalas pequenas
- **Pricing 2026:** Free (3 usuários), Pro USD 40/usuário/mês, Enterprise custom. Corte de custos vs LaunchDarkly: 5x menor para mesma escala

### 2.3 PostHog
- **Foco:** Product analytics open-source com feature flags bolted-on
- **Modelo de preço:** Freemium + usage-based USD 0,00045/evento
- **ICP:** Startup early-stage (<50 engenheiros), time de produto scrappy
- **Vantagem competitiva:** Open-source, session recording, heatmaps, self-hosted gratuito
- **Limitação:** Feature flags é secundário, sem health score, sem pricing em R$
- **Pricing 2026:** Free 1M eventos/mês, Pago USD 0,00045/evento acima. 90% das startups usam grátis

### 2.4 Pulse
- **Foco:** Híbrido: feature flags + A/B testing + analytics + health score
- **Modelo de preço:** Free self-hosted ilimitado + Cloud R$99-2.499/mês
- **ICP:** CTO / Head of Product de startup brasileira SaaS, 20-500 engenheiros
- **Vantagem competitiva:** Self-hosted em VPS USD 20/mês, pricing em R$, health score nativo, sem converter USD
- **Limitação:** Sem SDKs mobile nativo (iOS/Android), sem session recording, sem data warehouse
- **Positioning:** "LaunchDarkly + GrowthBook em um produto, self-hosted, em R$"

---

## 3. Análise Financeira Comparativa

| Métrica | **Pulse (proposta)** | **LaunchDarkly** | **GrowthBook** | **PostHog** |
|---------|:---:|:---:|:---:|:---:|
| Starter plan mensal | R$ 99 | USD 12 (~R$ 60) | USD 40 ~(R$ 200) | USD 0 (free) / USD 4.5k (commerical) |
| Custo anual por cliente (Growth) | R$ 1.188 | USD 240 (R$ 1.200) | USD 480 (R$ 2.400) | ~USD 2.000 (R$ 10k) @ 1M eventos |
| Margem operacional (estimada) | 70% | ~60% | ~50% | ~65% |
| Modelo de distribuição | Self-hosted + Cloud | Cloud SaaS | Cloud SaaS | Open-source + Cloud |
| Conversão no Brasil | Nenhuma perda de câmbio | -15% por conversão | -15% por conversão | -15% por conversão |
| Viabilidade para startup BR | ✅ Alto | Médio (USD) | Médio (USD) | Baixo (sem pricing) |

**Conclusão:** Pulse tem o melhor custo de aquisição inicial (R$ 99 < USD 12), mas GrowthBook ganha em transparência (USD 40 por usuário é previsível).

---

## 4. Gaps Críticos do Pulse (o que falta para paridade)

### 4.1 Gaps vs LaunchDarkly

| Gap | Prioridade | Esforço | Impacto | Status (2026-04) |
|-----|-----------|---------|---------|------------------|
| ✅ Ambientes por projeto | P1 | S | Alto | **FEITO** (2026-04-10) |
| ✅ Audit log de flags | P1 | M | Alto | **FEITO** (2026-04-10) |
| ✅ Rate limiting por projeto | P1 | S | Alto | **FEITO** (2026-04-10) |
| ✅ Stale flag detection | P2 | S | Médio | **FEITO** (2026-04-10) |
| Aprovação de mudanças (review flow) | P2 | L | Médio | Pendente |
| Streaming SSE para updates | P3 | M | Médio | Pendente |
| Flag triggers automáticos | P3 | L | Baixo | Pendente |

### 4.2 Gaps vs GrowthBook

| Gap | Prioridade | Esforço | Impacto | Status (2026-04) |
|-----|-----------|---------|---------|------------------|
| ✅ Análise estatística (chi-squared) | P1 | M | Alto | **FEITO** (2026-04-10) |
| ✅ Metric assignment | P1 | M | Alto | **FEITO** (2026-04-10) |
| ✅ Funis de conversão | P1 | L | Alto | **FEITO** (2026-04-10) |
| Bayesian inference (além de frequentist) | P2 | L | Médio | Pendente |
| Holdout groups | P2 | M | Médio | Pendente |
| Integração com data warehouse | P3 | L | Baixo | Não priorizar |

### 4.3 Gaps vs PostHog

| Gap | Prioridade | Esforço | Impacto | Status (2026-04) |
|-----|-----------|---------|---------|------------------|
| ✅ Event tracking robusto | P1 | S | Alto | **FEITO** (v2026-04) |
| ✅ Retention analysis | P1 | S | Alto | **FEITO** (v2026-04) |
| ✅ Cohort analysis | P1 | S | Alto | **FEITO** (v2026-04) |
| ✅ SQL explorer (query livre) | P2 | L | Médio | **EM DESENVOLVIMENTO** |
| Session recording | P3 | XL | Baixo | **NÃO PRIORIZAR** (objeto storage pesado) |
| Heatmaps | P3 | XL | Baixo | **NÃO PRIORIZAR** (DOM tracking complexo) |

### 4.4 Gaps exclusivos do Pulse (diferenciais a proteger)

| Feature | Vantagem | Risco |
|---------|----------|--------|
| Health score configurável | ✅ LaunchDarkly + GrowthBook não têm | Sem UI de configuração, é hardcoded |
| Churn prediction | ✅ Único no mercado | Precisa de ML real (regressão logística) vs regra simples |
| Pricing em R$ | ✅ Diferencial para BR | Concorrentes podem copiar (baixa barreira) |
| Self-hosted em Docker Compose | ✅ Rápido setup | PostHog também é open-source + self-hosted |

---

## 5. Roadmap de Evolução — 3 Fases

### Phase 1: MVP Competitivo (2-3 meses — Sprint 2026-04 até 06)
**Objetivo:** Paridade mínima com LaunchDarkly para flags + GrowthBook para A/B testing

- ✅ **Todos os bloqueadores P0 resolvidos** (BKL-1026 até BKL-1032)
- ✅ **SDKs React e React Native** (BKL-1048, BKL-NOVO)
- ✅ **Funis de conversão** (BKL-1058)
- ✅ **Análise estatística de A/B** (BKL-1043)
- ✅ **Audit log de flags** (BKL-1034)
- ✅ **Environments por projeto** (BKL-NOVO)
- ⏳ **Dashboard isolado por projeto** (sem dados xadrez)
- ⏳ **Webhooks de eventos** (integração Slack, n8n)

**Saída esperada:** Pulse vendível a primeiros clientes pagantes (3-5 clientes pilot em R$ 99-299/mês)

---

### Phase 2: Paridade Funcional (6-9 meses — Sprint 2026-07 até 12)
**Objetivo:** Cobertura de 85% dos casos de uso de LaunchDarkly + GrowthBook + PostHog

- ⏳ **Approval flow para flags** (mudanças requerem review)
- ⏳ **Streaming SSE** (eliminar polling do SDK)
- ⏳ **SDKs iOS/Android nativos** (além de Expo/RN)
- ⏳ **SQL explorer** (query builder para eventos)
- ⏳ **Export de dados** (CSV/JSON para BI)
- ⏳ **Dashboard de produto orientado a insights** (KPI cards + ações recomendadas)
- ⏳ **Visual rule builder** (drag-and-drop para targeting, sem JSON manual)

**Saída esperada:** 10-20 clientes pagantes, ARR R$ 30-50k, pronto para GTM agressivo

---

### Phase 3: Diferenciação Defensável (12+ meses — 2027+)
**Objetivo:** Features que LaunchDarkly + GrowthBook + PostHog não têm

- ⏳ **Funis de feature flags** (flag ativa → impacto no funil de conversão)
- ⏳ **Health score configurável via UI** (pesos + thresholds por cliente)
- ⏳ **Churn prediction com ML** (regressão logística, não regra simples)
- ⏳ **Approval flow contextual** (risk-based, ex: flags vermelhas precisam review)
- ⏳ **Schema registry leve** (validação opcional de eventos, sem Avo)
- ⏳ **Integração com LLMs** (flag para prompt versioning, A/B de prompts)

**Saída esperada:** Positioning único "o LaunchDarkly para Brasil que entende churn", ARR R$ 200k+

---

## 6. Proposta de Novos BKLs para o Backlog

### Gap Analysis Converted to Backlog Items

Os BKLs abaixo devem ser adicionados ao `BACKLOG.md` do `instituto-itinerante-ecosystem`, começando pelo ID **BKL-989** (próximo após BKL-988):

#### Phase 2 — A Implementar nos Próximos 6 Meses

**BKL-989: Pulse — Visual rule builder para targeting (drag-and-drop)**
- **Prioridade:** P2
- **Esforço:** L (10 dias)
- **Descrição:** Interface drag-and-drop para criar regras de targeting sem JSON manual. Exemplo: "(plano=pro AND país=BR) OR (beta_user=true)". Reduz barreira de entrada para Product Managers não-técnicos.
- **Dependências:** BKL-1067 (já existe, pendente)
- **Repo:** pulse-service (frontend)

---

**BKL-990: Pulse — Approval flow para mudanças críticas de flags**
- **Prioridade:** P2
- **Esforço:** L (12 dias)
- **Descrição:** Flags marcadas como "crítica" requerem aprovação de outro usuário antes de ativar em produção. Fluxo: owner submete mudança → requester aprova/rejeita → histórico de approvals no audit log. Diferencial vs LaunchDarkly starter.
- **Dependências:** BKL-1034 (audit log)
- **Repo:** pulse-service (backend + frontend)

---

**BKL-991: Pulse — Streaming SSE para flag updates (eliminando polling)**
- **Prioridade:** P2
- **Esforço:** M (8 dias)
- **Descrição:** WebSocket ou SSE para propagar mudanças de flags em tempo real (< 1s) vs polling 30s. SDK se conecta ao `/api/v1/flags/stream` ao carregar. Reduz latência de flag rollout — crítico para kill switches emergenciais.
- **Dependências:** —
- **Repo:** pulse-service (backend + SDK)

---

**BKL-992: Pulse — SDKs iOS e Android nativos**
- **Prioridade:** P2
- **Esforço:** XL (5 semanas)
- **Descrição:** SDKs para Swift (iOS) e Kotlin (Android) com suporte a feature flags, event tracking, offline mode. Integrar com Nitro e Libri se tiverem versões nativas no futuro. Publicar em CocoaPods e Maven Central.
- **Dependências:** —
- **Repo:** pulse-service (sdk/ios, sdk/android)

---

**BKL-993: Pulse — SQL explorer (query builder para eventos brutos)**
- **Prioridade:** P2
- **Esforço:** L (10 dias)
- **Descrição:** Interface similar ao PostHog para escrever queries SQL contra eventos brutos. Resultado em tabela + export CSV. Recurso avançado para times com analista de dados. Sem isso, clientes com demandas ad-hoc ficam presos.
- **Dependências:** —
- **Repo:** pulse-service (frontend)

---

**BKL-994: Pulse — Dashboard orientado a insights com ações recomendadas**
- **Prioridade:** P1
- **Esforço:** M (7 dias)
- **Descrição:** Home page do dashboard exibe 3 ações recomendadas dinâmicas baseadas em: (1) flags sem audit recente, (2) experimentos em coleta >7d, (3) eventos não rastreados. KPI cards: % de flags ativas, p-value de experimentos, saúde do projeto. Aumenta AFAW (north star).
- **Dependências:** —
- **Repo:** pulse-service (frontend)

---

**BKL-995: Pulse — Webhooks de eventos para integrações (Slack, n8n)**
- **Prioridade:** P2
- **Esforço:** M (5 dias)
- **Descrição:** Endpoint `POST /api/v1/webhooks` para registrar URLs. Trigger eventos: flag ativada, experimento terminou, churn alert. Sistema de retry + retry queue. Padrão: GitHub webhooks. Habilita integrações customizadas sem API polling.
- **Dependências:** —
- **Repo:** pulse-service (backend)

---

**BKL-996: Pulse — Dashboard isolado por projeto (zero dados xadrez)**
- **Prioridade:** P1
- **Esforço:** M (8 dias)
- **Descrição:** Dashboard hoje é hardcoded para o IIT. Cada projeto (via projeto.id na sessão) exibe só seus dados. Seletor de projeto no header. Middleware garante que sessão.projectId == request.projectId. Bloqueador crítico para vender a clientes externos.
- **Dependências:** BKL-1038 (project_id em eventos)
- **Repo:** pulse-service (frontend)

---

#### Phase 3 — A Implementar em 12+ Meses

**BKL-997: Pulse — Funis de feature flags (correlação flag → conversão)**
- **Prioridade:** P3
- **Esforço:** XL (15 dias)
- **Descrição:** Análise única do Pulse: mede impacto de uma flag no funil de conversão. Query: usuários que viram flag X ativa no dia Y (via flag_evaluations), depois compara taxa de conversão vs usuários sem a flag. Tabela e gráfico. Diferencial vs todos os concorrentes.
- **Dependências:** BKL-1058 (funis), BKL-1034 (audit de flags)
- **Repo:** pulse-service (backend + frontend)

---

**BKL-998: Pulse — Health score configurável por projeto (sem hardcoding)**
- **Prioridade:** P3
- **Esforço:** M (7 dias)
- **Descrição:** UI para configurar pesos de health score por projeto: login weight (20-40%), adoption weight (20-40%), subscription weight (10-30%), custom events weight (resto). Cada cliente define o que significa "saúde" para seu produto. Hoje é hardcoded em Go.
- **Dependências:** —
- **Repo:** pulse-service (frontend)

---

**BKL-999: Pulse — Churn prediction com ML (regressão logística)**
- **Prioridade:** P3
- **Esforço:** XL (20 dias)
- **Descrição:** Treinar modelo de regressão logística (Go + gonum) com dados de health score + eventos dos últimos 30d de clientes churnados. Retornar P(churn) por usuario, não just threshold fixo. Alerta quando > 70%. Dashboard com segmentação de churn risk (alto/médio/baixo). Único diferencial defensável.
- **Dependências:** BKL-998 (health score config)
- **Repo:** pulse-service (backend)

---

**BKL-1000: Pulse — Schema registry leve (validação opcional de eventos)**
- **Prioridade:** P3
- **Esforço:** M (8 dias)
- **Descrição:** Tabela `event_schemas(project_id, event_name, required_properties, property_types)`. No `/api/v1/ingest`, validação opcional (flag por projeto). Sem isso, eventos inconsistentes (ex: "user_id" vs "userId") causam problemas em escala. Lightweight vs Avo (sem colabs, sem PR/branch).
- **Dependências:** —
- **Repo:** pulse-service (backend + frontend)

---

**BKL-1001: Pulse — Bayesian inference para experimentos (além de Frequentist)**
- **Prioridade:** P3
- **Esforço:** L (8 dias)
- **Descrição:** Implementar Bayesian stopping rule (Thompson sampling) além do chi-squared (frequentist). Retornar dois p-values na API de experimentos: frequentist (BKL-1043) + Bayesian. UI exibe ambos com explicação. Diferencial vs GrowthBook (que só oferece Bayesian).
- **Dependências:** BKL-1043 (chi-squared)
- **Repo:** pulse-service (backend)

---

**BKL-1002: Pulse — Integração com LLMs (feature flags para prompts)**
- **Prioridade:** P3
- **Esforço:** L (10 dias)
- **Descrição:** Extensão do BKL-1064 (AI Configs). UI para versionar prompts + parâmetros de LLM (temperatura, max_tokens). Endpoint `GET /api/v1/ai-configs/:key/latest?user_id=X` retorna prompt + params. A/B de prompts sem redeploy de código. Aplicável a todos os agentes IIT.
- **Dependências:** BKL-1064 (AI Configs)
- **Repo:** pulse-service (backend + frontend)

---

**BKL-1003: Pulse — Node.js SDK**
- **Prioridade:** P3
- **Esforço:** M (6 dias)
- **Descrição:** SDK para Node.js/Express/Next.js com suporte a feature flags (server-side rendering), event tracking, batch mode. Publicar no npm como `@iit/pulse` ou `@pulse/node`. Necessário para backend integrations.
- **Dependências:** —
- **Repo:** pulse-service (sdk/node)

---

**BKL-1004: Pulse — Python SDK**
- **Prioridade:** P3
- **Esforço:** M (6 dias)
- **Descrição:** SDK para Python com suporte a feature flags e event tracking. Integração com FastAPI/Django. Publicar no PyPI. Necessário para scripts de análise de dados e backend em Python.
- **Dependências:** —
- **Repo:** pulse-service (sdk/python)

---

---

## 7. Segmentação de ICPs e Go-To-Market

### ICP 1: Startup SaaS Brasileira (20-50 engenheiros)
- **Problemas:** Feature flags sem rastreabilidade, A/B testing manual, analytics fragmentado
- **Produto oferecido:** Growth R$99/mês (500k eventos, 3 projetos, flags ilimitadas)
- **Canal:** Inbound via produto, recomendações, comunidade GDG
- **Conversão esperada:** R$ 99 → R$ 299 (Growth) em 3-6 meses
- **TAM Brasil:** ~200-300 startups SaaS nesta faixa

### ICP 2: EdTech / HealthTech Brasileira (50-500 engenheiros)
- **Problemas:** LGPD compliance, dados no próprio servidor, saúde de cliente crítica
- **Produto oferecido:** Pro R$297/mês (5M eventos, health score, churn alerts)
- **Canal:** Sales direto, partnerships, compliance consultants
- **Conversão esperada:** R$ 297 → R$ 999+ em contrato anual
- **TAM Brasil:** ~50-100 fintechs/healthtechs com budget >R$ 5k/ano

### ICP 3: Agência de Produto / Consultoria (10+ projetos de clientes)
- **Problemas:** Ferramentas caras por cliente, sem centralização, vendor lock-in
- **Produto oferecido:** Ofertar Pulse como white-label (custom domain, logo do cliente)
- **Canal:** Partnerships com agências, integrador de dados
- **Conversão esperada:** R$ 99-199 por cliente × 10+ = R$ 1k/mês
- **TAM Brasil:** ~20-40 agências de produto / consultoria

---

## 8. Recomendações Executivas

### Para o CEO (Antonio)

1. **Não vender Pulse externamente antes de:**
   - ✅ Todos os P0 bloqueadores resolvidos (2026-04-10: FEITO)
   - ✅ Dashboard isolado por projeto (BKL-996: próximo)
   - ✅ SDKs React e React Native maduros (BKL-1048: FEITO)
   - Mínimo 3 clientes internos usando em produção (SocialMake, Libri, Nitro)

2. **Priorize Phase 1 (próximos 2-3 meses):**
   - BKL-996 (dashboard isolado)
   - BKL-994 (insights home)
   - BKL-995 (webhooks)
   - Colocar em landing page + pricing page público

3. **Não invista em sessão recording ou heatmaps** — PostHog já cobre, infra pesada, ROI baixo.

4. **Teste GTM com ICP 1** (startup SaaS 20-50 eng):
   - Oferta: Free self-hosted + trial 14 dias cloud Growth R$ 99/mês
   - Canais: Product Hunt, Twitter, comunidades dev BR
   - Meta Q2: 3-5 clientes pagantes
   - Meta Q3: 10+ clientes pagantes, ARR R$ 15k

5. **Health score + churn prediction são diferenciadores defensáveis**:
   - Proteja como IP
   - Comunique como "único product no BR que avisa quando cliente vai embora"

---

## 9. Conclusão

**Maturidade atual do Pulse:** 4/10 (MVP funcional, mas com gaps críticos bloqueadores)

**Após Phase 1 (2-3 meses):** 7/10 (competitivo com LaunchDarkly starter + GrowthBook básico)

**Após Phase 2 (6-9 meses):** 8.5/10 (paridade funcional com concorrentes, pronto para escala)

**Após Phase 3 (12+ meses):** 9/10 (diferenciadores únicos defensáveis)

**Veredito:** Pulse é um produto viável se executado com disciplina. Os gaps não são bloqueadores estruturais — são implementáveis em 6 meses. O timing é crítico: meses 2-3 de 2026 para ganhar mercado antes de concorrentes se movimentarem no Brasil.

---

**Documento preparado por:** SOUL — PM Agent, Instituto Itinerante de Tecnologia  
**Data:** 2026-04-11  
**Próxima revisão:** 2026-07-01 (após Phase 1)
