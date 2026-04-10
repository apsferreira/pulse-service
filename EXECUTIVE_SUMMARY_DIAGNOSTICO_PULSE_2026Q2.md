# PULSE — Diagnóstico Executivo 2026 Q2
**Consolidação de análises de 13 agentes IIT**

**Data:** 10 de abril de 2026  
**Destinatário:** Antonio Pedro Ferreira (CEO)  
**Preparado por:** PM (SOUL)  
**Status:** Pronto para aprovação e priorização

---

## EXECUTIVE SUMMARY (5 minutos)

### Posição Atual
Pulse é um **MVP funcional mas frágil** que cobre ~60% dos casos de uso de LaunchDarkly/VWO/PostHog com **10 gaps críticos de segurança, analytics e SDK que bloqueiam clientes pagantes**.

### Diagnóstico em 1 linha
**"Flags e eventos funcionam, mas sem análise estatística, audit trail, multi-tenancy real, ou SDK web funcional — não vendível para fora do IIT hoje."**

### Recomendação Estratégica
1. **Próximos 30 dias:** Fechar 5 gaps críticos de segurança e analytics
2. **Próximos 90 dias:** Implementar tier de features para Starter/Growth/Pro vendível
3. **Próximos 180 dias:** Posicionar como alternativa self-hosted premium a LaunchDarkly para mercado BR

### Impacto Financeiro (12 meses)
- Hoje: R$0 ARR (produto interno)
- Com roadmap curto (6m): R$15-20k ARR (5-10 clientes Growth)
- Com roadmap completo (12m): R$60-80k ARR (30-50 clientes mix)

---

## 1. ONDE ESTAMOS HOJE (Diagnóstico por Dimensão)

### 1.1 Feature Completeness — 6/10

| Feature | Status | Nível |
|---------|--------|-------|
| **Feature Flags — On/Off simples** | ✅ Funcionando | MVP |
| **Rollout percentual com hash** | ✅ Funcionando | MVP |
| **Targeting por user/tenant JSONB** | ✅ Funcionando | MVP |
| **Variantes A/B com peso** | ✅ Implementado | MVP |
| **Event tracking por projeto** | ⚠️ Parcial (sem project_id scoping) | Frágil |
| **Retenção (cohort, DAU/MAU)** | ✅ Básico (RetentionService) | MVP |
| **CSAT/NPS nativo** | ✅ Schema | Não testado |
| **Health Score configurável** | ✅ Implementado (hardcoded) | Proof of concept |
| **Análise estatística de experimentos** | ❌ ZERO | Bloqueador |
| **Funis de conversão** | ❌ ZERO | Bloqueador |
| **Audit log de flags** | ❌ ZERO | Bloqueador |
| **SDKs React/React Native** | ❌ ZERO | Bloqueador |
| **SDK web via CDN funcional** | ⚠️ ES module (não funciona) | Inútil |
| **Ambientes por projeto em flags** | ❌ ZERO (events têm, flags não) | Crítico |

**Veredito:** O Pulse é um **feature flags incompleto + analytics skeleton sem produção**.

---

### 1.2 Segurança — 3/10 (CRÍTICO)

| Gap | Severidade | Impacto | Status |
|-----|-----------|---------|--------|
| **Cross-tenant data leak em EvaluateFlags** | CRÍTICA | flags `GetAll(ctx, nil)` sem `project_id` | Confirmado |
| **Cache Redis global para flags** | CRÍTICA | `pulse:flags:map` compartilhado entre tenants | Confirmado |
| **X-Project-Key nunca validado** | CRÍTICA | SDK envia header, backend ignora | Confirmado |
| **SQL injection em event_repository.go** | ALTA | `fmt.Sprintf` em queries | Confirmado (linhas 60+) |
| **Rate limiting por IP (não por API key)** | ALTA | Um cliente abusivo afeta todos | Por design |
| **Go 1.25.0 (versão inexistente)** | ALTA | Dockerfile quebrado em CI/CD | Confirmado |
| **Sem DPA para clientes externos** | ALTA | Requisito contratual faltando | Não existe |
| **Sem consent management (LGPD)** | MÉDIA | Direito ao esquecimento não implementado | Não existe |
| **Sem audit log de flag evaluations** | MÉDIA | Compliance impossível em produção | Não existe |
| **minReplicas: 1 em K3s** | MÉDIA | Indisponibilidade com 1 crash | Por design |

**Veredito:** Pulse **NÃO é seguro para clientes pagantes**. Bloqueadores de produção: cross-tenant leaks, SQL injection, falta de auditoria.

---

### 1.3 Escalabilidade e Infra — 5/10

| Métrica | Target | Atual | Status |
|---------|--------|-------|--------|
| **Flag evaluation latência p99** | <100ms | ~50ms (cached) | ✅ |
| **Event ingestão throughput** | 1M eventos/min | ~800 eventos/s (~48k/min) | ⚠️ Falta batch INSERT |
| **RPO (Recovery Point Objective)** | <1h | 24h (pg_dump nightly) | ⚠️ Sem WAL archiving |
| **Replicação de read** | Sim | NÃO | ❌ Single PostgreSQL |
| **minReplicas em K3s** | 3+ | 1 | ❌ Risco de downtime |
| **Redis persistence** | RDB + AOF | RDB apenas | ⚠️ Pode perder dados |

**Veredito:** Infraestrutura **frágil para produção com SLAs**. Sem replicação PG, single replica K3s, sem WAL.

---

### 1.4 Analytics — 2/10

| Feature | LaunchDarkly | VWO | PostHog | Pulse | Bloqueador? |
|---------|:---:|:---:|:---:|:---:|:---:|
| Chi-squared p-value | Sim | Sim | Sim | ❌ | SIM |
| Confidence interval | Sim | Sim | Sim | ❌ | SIM |
| Funis | Não | Sim | Sim | ❌ | SIM |
| Metric assignment (flag→métrica) | Sim | Sim | Sim | ❌ | SIM |
| User journeys/paths | Não | Sim | Sim | ❌ | NÃO |
| SQL explorer | Não | Não | Sim | ❌ | NÃO |
| Property filtering em queries | Não | Sim | Sim | ❌ | SIM |

**Veredito:** Pulse é **analytics mudo**. A/B testing sem significância = divisão de tráfego, não experimentação.

---

### 1.5 GTM / Comercial — 4/10

| Item | Status | Impacto |
|------|--------|--------|
| **Pricing page pública** | ❌ ZERO | Sem conversão self-serve |
| **Self-service onboarding** | ❌ ZERO | Sem free trial |
| **ICP definido** | ⚠️ Vago (startups BR) | Mensaging fraca |
| **Documentação de SDK** | ⚠️ Mínima | DX ruim |
| **Comparação vs concorrentes** | ✅ Análise completa | Uso interno |
| **Pricing por plano** | ✅ Definido | Starter R$299, Growth R$799, Pro R$1.999 |
| **Free tier policy** | ❌ Indefinido | Abuso possível |
| **Suporte tiered** | ❌ ZERO | Clientes sem SLA |

**Veredito:** Pulse é um **produto técnico sem go-to-market**. Ninguém fora do IIT sabe que existe.

---

## 2. ONDE QUEREMOS CHEGAR (Visão 12 meses)

### 2.1 Posicionamento Estratégico (Nicho Defendível)

**Tagline proposta:**
> "Feature flags, A/B testing e product analytics em um produto self-hosted com pricing em R$. Para startups e scale-ups brasileiras que querem controlar seus dados sem pagar em USD."

**Diferencial vs concorrentes:**

| vs LaunchDarkly | vs VWO | vs PostHog |
|---|---|---|
| ✅ Self-hosted | ✅ Pricing em R$ | ✅ Health score nativo |
| ✅ Preço 50% menor | ✅ Analytics integrado | ✅ Churn prediction |
| ✅ Health score | ✅ Feature flags integrado | ✅ Sem overhead infra pesada |
| ❌ Sem Enterprise SLA | ❌ Sem audit log | ❌ Sem funis (PostHog tem) |

**ICP (Ideal Customer Profile):**
- **Persona:** Engineering Manager ou CTO de startup/scale-up com 20-500 engenheiros
- **Budget:** R$300-2.000/mês (vs USD 200-5.000 de LD/VWO)
- **Pain:** Dados sensíveis no servidor, múltiplas ferramentas, falta de analytics integrada
- **Segmento:** FinTech, HealthTech, EdTech, PropTech BR com requisito LGPD ou vendor lock-in

**Market Size Potencial:**
- Startups de Series A/B em BR: ~500 (Crunchbase)
- Scale-ups com 50+ eng: ~100
- Enterprises com requisito self-hosted: ~50-100
- TAM: ~650 potenciais clientes × R$1.000/mês (mix) = **R$7,8M TAM em 5 anos**

---

### 2.2 Roadmap de Funcionalidades 12 Meses

#### Tier 1 (Próximos 60 dias) — MVP Vendível
**Objetivo:** Feature parity básica com LaunchDarkly starter + análise estatística

- **Security:** Corrigir cross-tenant leaks, SQL injection, Go version
- **Analytics:** Chi-squared p-value + confidence interval para A/B tests
- **SDK Web:** Reescrever `pulse.js` como IIFE + implementar `track()` real
- **Audit:** Tabela `flag_changes` com actor/timestamp/diff
- **Ambientes:** Campo `environment` em flags (staging/prod)
- **Project-scoping:** Adicionar `project_id` a events e filtrar queries

**Resultado:** Pulse seguro para 1-5 clientes em beta com feature flags rastreáveis + A/B testing com significância.

---

#### Tier 2 (Próximos 180 dias) — Paridade Competitiva
**Objetivo:** Funcionalidade completa para crescimento linear de MRR

- **SDKs idiomáticos:** `@pulse/react` + `@pulse/react-native` com hooks
- **Funis:** SQL com window functions para análise de conversão
- **Metric assignment:** Vincular flags a eventos para experimentação real de negócio
- **Rate limiting + Quota:** Enforcement por projeto com bloqueia no Free
- **Dashboard multi-projeto:** Seletor de projeto + isolamento de dados
- **Webhooks:** POST para Slack, n8n, sistemas internos
- **SDK Go maduro:** Documentação, versionamento, pkg.go.dev

**Resultado:** Pulse com 80% de paridade com LaunchDarkly/VWO. Pronto para 20-30 clientes pagantes.

---

#### Tier 3 (180-365 dias) — Diferenciação Premium
**Objetivo:** Features únicas que LaunchDarkly/VWO não têm

- **Funis de feature flags:** Ver se ativar uma flag melhorou conversão no funil
- **Health score configurável via UI:** Clientes customizam pesos e thresholds
- **Churn prediction:** ML leve (regressão logística) sobre health score + comportamento
- **Schema registry leve:** Validação opcional de evento schema (anti-Avo)
- **Approval flow:** Request de review antes de ativar flags em prod
- **NPS triggered por evento:** Disparar survey após evento customizável
- **Data export:** CSV/JSON de eventos brutos para BI próprio
- **Streaming SSE:** Flag updates <1s via SSE (vs polling ~30s)

**Resultado:** Pulse com diferenciadores únicos. Posicionamento competitivo agressivo no mercado BR premium.

---

### 2.3 North Star Metric (NSM) Recomendada

**Métrica:** **Activated Features per Active User per Week (AFAW)**

**Definição:**
- **Active User:** qualquer usuário que teve ≥1 flag evaluation no último dia
- **Activated Feature:** flag foi avaliada como `true` para o usuário (independent de enabled/disabled status)
- **Métrica:** (sum de evaluations=true) / (unique users) / 7 dias

**Por que esta métrica?**
1. **Alinhada com o núcleo do produto:** Feature flags são o diferencial do Pulse
2. **Indicadora de saúde:** Crescimento de AFAW = adoção real, não apenas sign-ups
3. **Prevê retenção:** Times que experimentam com flags tendem a renoar (ARPU preditor)
4. **Não é métrica de vanidade:** Não cresce com spam de eventos; requer decisões de produto

**Targets por Trimestre:**

| Trimestre | AFAW Target | Clientes Ativos | MRR |
|-----------|-----------|---|---|
| Q2 2026 (atual) | 0.5 (interno) | 1 (IIT) | R$0 |
| Q3 2026 | 1.2 | 5-8 clientes | R$3-5k |
| Q4 2026 | 2.1 | 15-25 clientes | R$12-20k |
| Q1 2027 | 3.5 | 35-50 clientes | R$35-50k |

**Como medir:**
```sql
SELECT 
  DATE_TRUNC('day', created_at) as day,
  COUNT(DISTINCT user_id) as active_users,
  COUNT(*) FILTER (WHERE evaluated_as = true) as activated_features,
  ROUND(COUNT(*) FILTER (WHERE evaluated_as = true) / COUNT(DISTINCT user_id) / 7.0, 2) as AFAW
FROM flag_evaluations
WHERE project_id IN (SELECT id FROM projects WHERE status = 'active')
  AND created_at >= NOW() - INTERVAL '7 days'
GROUP BY 1
ORDER BY 1 DESC;
```

---

## 3. TOP 10 BACKLOG RICE (Priorização)

### RICE Framework
- **Reach:** # usuários/semana afetados (1-3 pontos)
- **Impact:** Efeito por usuário (3=huge, 2=high, 1=medium, 0.5=low)
- **Confidence:** % confiança na estimativa (0.5-1.0)
- **Effort:** semanas de desenvolvimento
- **RICE Score:** (Reach × Impact × Confidence) / Effort

---

### Ranking Final

| # | Item | Reach | Impact | Conf | Effort | RICE | Prioridade | Bloqueador? |
|---|------|:---:|:---:|:---:|:---:|:---:|---|:---:|
| 1 | **Corrigir cross-tenant data leak (EvaluateFlags)** | 20 | 3 | 1.0 | 0.5 | 120 | 🔴 P0 | SIM |
| 2 | **Corrigir SQL injection (event_repository.go)** | 20 | 3 | 1.0 | 0.5 | 120 | 🔴 P0 | SIM |
| 3 | **Audit log de flag changes** | 30 | 2 | 1.0 | 3 | 20 | 🔴 P0 | SIM |
| 4 | **Chi-squared p-value + CI para A/B tests** | 25 | 2.5 | 1.0 | 5 | 12.5 | 🟠 P1 | SIM |
| 5 | **SDK web JS (IIFE + track real)** | 40 | 3 | 1.0 | 5 | 24 | 🟠 P1 | SIM |
| 6 | **@pulse/react hooks (useFlag, useTrack)** | 30 | 2.5 | 0.9 | 7 | 9.6 | 🟠 P1 | NÃO |
| 7 | **Project scoping em events (project_id)** | 25 | 2.5 | 1.0 | 2 | 31.25 | 🟡 P2 | SIM |
| 8 | **Funis de conversão (SQL window functions)** | 35 | 2 | 0.9 | 14 | 4.5 | 🟡 P2 | NÃO |
| 9 | **Rate limiting + Quota enforcement** | 20 | 2 | 0.8 | 3 | 10.7 | 🟡 P2 | NÃO |
| 10 | **@pulse/react-native (Expo/RN)** | 15 | 2.5 | 0.8 | 7 | 4.3 | 🟡 P3 | NÃO |

---

### Detalhamento dos Top 5 Críticos

#### 1. Corrigir Cross-Tenant Data Leak (RICE 120) — P0 IMEDIATO
**Problema:** `EvaluateFlags` chama `GetAll(ctx, nil)` sem project_id filter. Tenant A vê flags de Tenant B.

**Solução:**
- Adicionar `project_id` a cada flag evaluation
- Filtrar queries com `WHERE project_id = $1`
- Validar X-Project-Key header vs projeto na request

**Esforço:** 0.5 semana (4h código + 4h testes)

**Critério de Aceite:**
- Dado tenant A faz GET `/flags?project_id=proj-a`
- Quando backend valida project_id na requisição
- Então flags de proj-b não aparecem no resultado
- E flags são cacheadas por (projeto, user) na chave Redis

---

#### 2. Corrigir SQL Injection (RICE 120) — P0 IMEDIATO
**Problema:** `event_repository.go` usa `fmt.Sprintf` em SQL queries ao invés de prepared statements.

**Solução:**
- Converter todas queries de `fmt.Sprintf(...)` para parametrizadas `($1, $2, ...)`
- Adicionar test case com payload malicioso: `"; DROP TABLE events; --`

**Esforço:** 0.5 semana (3h código + 4h testes)

---

#### 3. Audit Log de Flag Changes (RICE 20) — P0
**Problema:** Ninguém sabe quem ativou/desativou qual flag, quando. Requisito legal para produção.

**Solução:**
```sql
CREATE TABLE flag_changes (
  id UUID PRIMARY KEY,
  flag_id UUID REFERENCES feature_flags,
  actor_id VARCHAR NOT NULL,  -- user que fez a mudança
  action VARCHAR(32),  -- create, update, delete
  previous_state JSONB,  -- estado anterior (enabled, rollout%, variants)
  new_state JSONB,  -- estado novo
  timestamp TIMESTAMPTZ DEFAULT NOW(),
  project_id UUID REFERENCES projects
);
```
- Trigger PostgreSQL em UPDATE de flags → insert em flag_changes
- Dashboard mostra histórico de mudanças com actor + timestamp

**Esforço:** 3 semanas (1h DB + 2h backend + 2h frontend + 2h testes)

---

#### 4. Chi-Squared P-Value + Confidence Interval (RICE 12.5) — P1
**Problema:** A/B test mostra "Variant A: 30% conversão vs Variant B: 28%" sem indicador se diferença é significante.

**Solução:**
1. Coletar eventos com `variant` tag durante experiment
2. Na visualização, calcular chi-squared test:
   ```
   H0: não há diferença entre variantes
   p-value = P(dados | H0 verdadeiro)
   CI 95% = intervalo onde valor verdadeiro cai com 95% prob
   ```
3. Usar `gonum` (biblioteca Go) para cálculos estatísticos
4. Dashboard mostra: "Variant A 30% ± 2% vs Variant B 28% ± 2%, p=0.18 (não significante)"

**Esforço:** 5 semanas (2h análise + 3h código + 3h testes + 2h validação com estatístico)

**Critério de Aceite:**
- Dado experimento com 100 conversões em 1000 usuários em cada variante
- Quando calculo chi-squared
- Então p-value ≈ 0.047 (à beira da significância em α=0.05)
- E CI 95% = [24%, 36%] para Variant A

---

#### 5. SDK Web JavaScript (IIFE + Track Real) (RICE 24) — P1
**Problema:** `sdk/pulse.js` é ES module — não funciona com `<script src="cdn">`. `track()` é no-op.

**Solução:**
```javascript
// Novo formato IIFE
(function(window) {
  window.Pulse = {
    init: function(apiKey, projectId) { ... },
    isEnabled: function(flagKey, context) { ... },
    track: function(eventName, properties) { 
      // POST /api/v1/ingest com batch
    },
    identify: function(userId, traits) { ... }
  };
})(window);
```

**Mudanças:**
- Reescrever `sdk/js` como IIFE (não module)
- Implementar batch queue (100 events ou 5s timeout)
- POST para `/api/v1/ingest` com `api_key` header
- Fallback offline: salvar em localStorage, sincronizar quando online
- Snippet de instalação: `<script async src="https://cdn.pulse.io/v1/pulse.js"></script>`

**Esforço:** 5 semanas (2h design + 3h IIFE rewrite + 2h batching + 3h testing + 1h CDN setup)

---

## 4. ROADMAP 90-180-365 DIAS

### PHASE 0 — Security Lockdown (Próximos 30 dias) 🔴
**Objetivo:** Fechar 5 bloqueadores de segurança

- ✅ Corrigir cross-tenant leak
- ✅ Corrigir SQL injection
- ✅ Adicionar audit log (schema + trigger)
- ✅ Validar X-Project-Key header
- ✅ Corrigir Go version 1.25.0 → 1.24.0
- ✅ Deploy em K3s com 3 replicas

**Entrada para clientes:** Nenhuma. Internal only.

---

### PHASE 1 — MVP Vendível (Dias 30-90) 🟠

**Objetivo:** Feature parity com LaunchDarkly starter. Pronto para 5-10 clientes em beta.

- ✅ Chi-squared A/B test analysis
- ✅ SDK web via CDN (track real)
- ✅ Ambientes por projeto (staging/prod)
- ✅ Project scoping em events
- ✅ Dashboard de A/B tests com significância

**Entrada para clientes:** Growth plan (R$799/mês) — 5 projetos, 500k eventos, análise de A/B

---

### PHASE 2 — Paridade Competitiva (Dias 90-180) 🟡

**Objetivo:** Funcionalidade completa. Crescimento agressivo de MRR.

- ✅ `@pulse/react` com hooks
- ✅ Funis de conversão
- ✅ Metric assignment (flag → métrica)
- ✅ Rate limiting + Quota enforcement
- ✅ Dashboard multi-projeto
- ✅ Webhooks (Slack, n8n)
- ✅ Pricing page pública + free tier

**Entrada para clientes:** Free (10k eventos/mês) → Growth (R$799) → Pro (R$1.999)

**Target:** 20-30 clientes, R$15-25k MRR

---

### PHASE 3 — Premium Differentiation (Dias 180-365) 🟢

**Objetivo:** Diferenciadores únicos. Posicionamento defensável.

- ✅ Funis de feature flags (único mercado)
- ✅ Health score configurável
- ✅ Churn prediction com ML
- ✅ Schema registry leve
- ✅ Approval flow para flags
- ✅ NPS triggered por evento
- ✅ Data export (BigQuery, S3)
- ✅ Streaming SSE para flags

**Entrada para clientes:** Pro+ (R$3.000+/mês) — features enterprise

**Target:** 50+ clientes, R$60-80k MRR

---

## 5. DECISÕES CRÍTICAS PARA APROVAÇÃO DO CEO

### 5.1 Segurança: Go/No-Go para Clientes Pagantes

**DECISÃO NECESSÁRIA:**

- **Opção A (Recomendado):** Pausar vendas externas até PHASE 0 completo (30 dias). Usar período para garantir Pulse é seguro antes de qualquer cliente beta.
  - Custo: 0 (time alocado já)
  - Benefício: Zero risco de leak de dados
  - Timeline: 09/05 (pronto para vendas)

- **Opção B (Alto Risco):** Começar vendas com disclaimer de "beta, sem SLA, dados em risco". NÃO RECOMENDADO.
  - Risco: LGPD violation, reputacional, contrato quebrado
  - Benefício: MRR mais cedo (improvável — clientes sérios não entram nesse risco)

**Recomendação:** Opção A. Aprovar PHASE 0 como P0 obrigatório.

---

### 5.2 Pricing: Elevar ou Manter?

**Análise Competitiva:**

| Plano | Atual | Proposto | Concorrente | Recomendação |
|-------|-------|----------|-------------|---|
| Starter | R$299 | R$399 | VWO R$1.030 | ✅ Elevar |
| Growth | R$799 | R$1.099 | VWO R$2.194 | ✅ Elevar |
| Pro | R$1.999 | R$2.399 | LD R$4.160 | ⚠️ Considerar |

**Impacto:** +35% ARR com mesmo volume.

**DECISÃO NECESSÁRIA:**
- Elevar para R$399 / R$1.099 / R$2.399 em 01/05/2026?
- Ou manter e usar como argumento de "underpriced" para crescimento agressivo?

**Recomendação:** Elevar a R$399/R$1.099. Manter Pro em R$1.999 por agora (elasticidade desconhecida).

---

### 5.3 Self-Hosted: Posicionar ou Descartar?

**Modelo Self-Hosted Proposto:** R$24.000/ano (Opção B: R$24k/ano + R$5k/mês suporte)

**DECISÃO NECESSÁRIA:**
- Incluir self-hosted como plano oficial desde PHASE 2 (dia 90)?
- Ou focar em cloud primeiro, self-hosted em PHASE 3 (dia 180)?

**Recomendação:** PHASE 2. Self-hosted é diferencial vs LD/VWO. Inclua no Growth/Pro como opção "On-Premise Deploy" por R$5k/mês de suporte.

---

### 5.4 ICP e Go-to-Market: Qual Segmento Atacar Primeiro?

**Três Opções:**

| Segmento | Viabilidade | Timing | Esforço Sales |
|----------|-----------|--------|---|
| **Fintech BR** (Pix, pagamento) | Alta | Médio (3m) | Alto (vendas técnicas) |
| **EdTech** (Curso online, bootcamp) | Média | Curto (1m) | Baixo (product-led) |
| **SaaS genérico BR** (agências, consultoria) | Muito Alta | Imediato | Médio |

**DECISÃO NECESSÁRIA:**
- Qual ICP primeira prioridade?
- Alocar 1 full-time Growth/Sales para esse segmento?

**Recomendação:** **EdTech primeiro** (PHASE 1, dia 30-90). Razão:
- Público open para experimentação (não risk-averse)
- Problema claro: A/B testing de onboarding
- Vendas product-led (demo simples)
- Migração de Amplitude/Mixpanel é comum (já conhecem categoria)

---

## 6. MATRIZ DE DEPENDÊNCIAS

```
PHASE 0 (Seg 1-4)
├── Security fixes (cross-tenant, SQL injection)
├── Audit log schema
└── K3s hardening (3 replicas)

PHASE 1 (Seg 5-12)
├── Depends on: PHASE 0 ✅
├── Chi-squared analysis
├── SDK web IIFE + track
└── A/B dashboard

PHASE 2 (Seg 13-24)
├── Depends on: PHASE 1 ✅
├── @pulse/react (depends on SDK web)
├── Funis (depends on project scoping)
└── Webhooks (independent)

PHASE 3 (Seg 25-52)
├── Depends on: PHASE 2 ✅
├── Health score UI
├── Churn prediction (depends on health score)
└── Schema registry (independent)
```

**Caminho crítico:** PHASE 0 → PHASE 1 → PHASE 2 (16 semanas = 4 meses para MVP competitivo)

---

## 7. RESOURCE ALLOCATION RECOMENDADA

### Para Executar PHASE 0-1 (Próximos 90 dias)

| Função | Alocação | Responsável | Esforço |
|--------|----------|-------------|---------|
| **Backend Lead** | 1.0 FTE | @tl | Security + Analysis |
| **Frontend Lead** | 0.8 FTE | @frontend | A/B dashboard + SDK web |
| **QA** | 0.5 FTE | @qa | Testes security + regressão |
| **PM** | 0.3 FTE | @pm | Specs + priorização |

**Estrutura de sprints:**
- Semana 1-2: PHASE 0 security (2 devs, 1 QA)
- Semana 3-4: PHASE 0 hardening + PHASE 1 design (specs)
- Semana 5-8: PHASE 1 development (chi-squared, SDK web, A/B dashboard)
- Semana 9-12: PHASE 1 testing + launch prep

**Total:** 4 pessoas × 90 dias = 1.440 pessoa-horas

---

## 8. SUCCESS CRITERIA (Go/No-Go Gates)

### PHASE 0 Exit Criteria (30 dias)
- [ ] Cross-tenant leak corrigido + tested
- [ ] SQL injection corrigido + tested
- [ ] Audit log implementado + tested
- [ ] K3s rodando 3 replicas
- [ ] Penetration test interno OK
- **Gate:** Security sign-off de @tl + @qa

### PHASE 1 Exit Criteria (90 dias)
- [ ] A/B analysis com p-value + CI
- [ ] SDK web IIFE funcional em 3 apps
- [ ] Ambientes por projeto implementados
- [ ] 0 event losses em teste de carga 1M/min
- [ ] 5+ clientes em beta usando Growth plano
- [ ] NPS >= 40 entre beta users
- **Gate:** Product readiness sign-off de @pm + @tl

### PHASE 2 Exit Criteria (180 dias)
- [ ] `@pulse/react` com ≥10k downloads/mês
- [ ] Funis implementados + usados em 50% de clientes
- [ ] 20-30 clientes ativos
- [ ] R$15-25k MRR
- [ ] NPS >= 50
- **Gate:** Business readiness sign-off de @pm + @cfo

---

## 9. RISKS E MITIGAÇÕES

| Risk | Prob | Impact | Mitigação |
|------|------|--------|-----------|
| **Cross-tenant leak encontrado depois de clientes beta** | Alta | Crítica | PHASE 0 penetration test completo |
| **A/B analysis sem estatístico valida p-value incorreto** | Média | Alta | Revisar fórmulas com especialista antes |
| **SDK web cria performance issue em frontend cliente** | Média | Alta | Load test em 3 apps reais (Libri, Nitro, SocialMake) |
| **Falta de demanda por Pulse como produto (SaaS)** | Média | Média | MVP learning: 3 clientes alpha = validação mínima |
| **Concorrente reduz preço (PostHog eventos mais baratos)** | Média | Média | Diferenciador: health score + audit log, não preço |

---

## 10. PRÓXIMOS PASSOS (Dentro de 48h)

1. **CEO Decision:** Aprovar PHASE 0 como P0? (Sim/Não)
   - Se SIM → alocar @tl + 1 backend dev + @qa para 30 dias

2. **CEO Decision:** Elevar pricing R$299→R$399?
   - Se SIM → atualizar site, docs, pricing page antes de PHASE 1

3. **CEO Decision:** Qual ICP primeira: FinTech, EdTech, ou SaaS genérico?
   - Resultado: 1 person Growth alocado

4. **PM Action:** Criar BKLs no Linear com RICE scores + acceptance criteria

5. **TL Action:** Iniciar security review de cross-tenant leak (horas, não dias)

---

## APÊNDICE: SUMMARY TABLE

| Dimensão | Score | Veredito | Bloqueador? |
|----------|-------|---------|-----------|
| **Feature Completeness** | 6/10 | MVP incompleto | SIM (analytics, SDKs) |
| **Segurança** | 3/10 | CRÍTICO | SIM (cross-tenant, SQL injection) |
| **Escalabilidade** | 5/10 | Frágil | SIM (single PG, 1 replica K3s) |
| **Analytics** | 2/10 | Mudo | SIM (sem p-value, funis) |
| **GTM** | 4/10 | Inexistente | SIM (sem pricing page, onboarding) |
| **RESULTADO GERAL** | 4/10 | **NÃO VENDÍVEL HOJE** | **PHASE 0 OBRIGATÓRIO** |

---

**Documento preparado por:** PM (SOUL)  
**Data:** 10 de abril de 2026  
**Status:** Aguardando aprovação CEO  
**Próxima revisão:** 30 de abril de 2026 (pós PHASE 0)

