# Pulse — Análise Competitiva

**Data:** Abril 2026
**Versão:** 1.0
**Autor:** IIT AI Strategy

---

## 1. Feature Matrix

Comparativo entre Pulse (estado atual, baseado nas migrations e serviços existentes) e os quatro concorrentes relevantes.

| Feature | Pulse | PostHog | Mixpanel | LaunchDarkly | Avo |
|---------|-------|---------|----------|--------------|-----|
| **Ingestion de eventos customizados** | Sim (POST /track, POST /ingest) | Sim | Sim | Não (events = contexto de flags apenas) | Não (Avo não captura, apenas valida) |
| **Identify (anon → user)** | Sim (POST /identify) | Sim | Sim | Sim | Não |
| **Funis** | Não | Sim | Sim | Não | Não |
| **Retenção (DAU/WAU/MAU)** | Sim (RetentionService + cohorts) | Sim | Sim | Não | Não |
| **Cohort analysis** | Sim (GetCohorts por semana/mês) | Sim | Sim | Não | Não |
| **Session recording** | Não | Sim | Não | Não | Não |
| **Heatmaps** | Não | Sim | Não | Não | Não |
| **Feature flags (on/off)** | Sim | Sim | Não nativo | Sim (core product) | Não |
| **Rollout percentual** | Sim (hash determinístico por user) | Sim | Não | Sim | Não |
| **Targeting por user/tenant** | Sim (JSONB targeting) | Sim | Não | Sim | Não |
| **A/B variants com peso** | Sim (BKL-094, variants JSONB) | Sim (experiments) | Sim (A/B testing) | Sim (experiments) | Não |
| **Batch flag evaluation** | Sim (POST /flags/evaluate-batch, max 50 keys) | Sim | Não | Sim | Não |
| **Scoping de flags por projeto** | Sim (project_id em feature_flags) | Sim (por projeto) | Não | Sim (por projeto/ambiente) | Não |
| **Tracking plan governance** | Não | Parcial (data management) | Não | Não | Sim (core product) |
| **Schema validation de eventos** | Não | Parcial | Não | Não | Sim |
| **NPS nativo** | Sim (tabela nps_responses, triggered surveys) | Sim | Não | Não | Não |
| **CSAT/Surveys** | Sim (csat_surveys + csat_responses) | Sim | Não | Não | Não |
| **Health score de tenant** | Sim (tenant_health_scores, score 0-100 com 4 dimensões) | Não | Não | Não | Não |
| **Churn alert automático** | Sim (ChurnAlertScheduler + email Resend, threshold 40) | Não | Não | Não | Não |
| **Custo de API (LLM tracking)** | Sim (api_costs: tokens_in/out, cost_usd por produto/modelo) | Não | Não | Não | Não |
| **Uptime/monitors** | Sim (tabela monitors, ws_handler) | Não (usa integração) | Não | Não | Não |
| **Self-hosted** | Sim (Go + PostgreSQL + Redis) | Sim (open-source) | Não | Não | Não |
| **Multi-tenant (projetos)** | Parcial (tabela projects com API key, não há billing por projeto ainda) | Sim | Sim | Sim | Sim |
| **SDK JavaScript** | Sim (sdk/pulse.js — feature flags + track no-op fase 1) | Sim (posthog-js, maduro) | Sim (mixpanel-browser) | Sim (launchdarkly-js-client) | Sim (Avo SDK por linguagem) |
| **SDK Go** | Sim (sdk/go/) | Sim | Sim | Sim | Não |
| **SDK mobile (iOS/Android/RN)** | Não | Sim | Sim | Sim | Sim |
| **Webhooks** | Não | Sim | Não (Lexer integração) | Sim | Não |
| **API REST documentada** | Parcial (README básico) | Sim (OpenAPI) | Sim | Sim (REST + gRPC SSE) | Sim |
| **Dashboard UI** | Sim (React/TS: Events, Retention, CSAT, Costs, Uptime, A/B) | Sim (maduro) | Sim (maduro) | Sim (maduro) | Sim |
| **Alertas** | Sim (email churn via Resend) | Sim (subscriptions) | Sim | Sim | Não |
| **Ambientes (staging/prod)** | Parcial (migration 013 adiciona environment em events) | Sim | Sim | Sim (core: environments por projeto) | Sim |
| **Pricing em R$** | Sim (Free / R$99 / R$297) | Não (USD) | Não (USD) | Não (USD) | Não (USD) |
| **Billing integrado** | Parcial (checkout-service/Asaas, planos definidos) | Sim (cloud) | Sim (cloud) | Sim (cloud) | Sim (cloud) |

---

## 2. Gap Analysis

### 2.1 O que cada concorrente tem e o Pulse não tem

**PostHog**
- Session recording e replay — captura cliques, scrolls, erros com playback
- Heatmaps de cliques e scroll depth
- Surveys in-product mais avançados (com targeting por cohort)
- Integrations marketplace (Hubspot, Sentry, Slack, Intercom, ~40+ integrações)
- SQL explorer (query builder sobre eventos brutos)
- SDK mobile maduro (iOS, Android, React Native, Flutter)
- Dashboard sharing público/privado com embed
- Data warehouse export nativo (BigQuery, Snowflake, Redshift)

**Mixpanel**
- Funis com análise de abandono por etapa — ausência crítica no Pulse
- Impact analysis (como uma feature afetou uma métrica)
- Signal (correlação automática entre propriedades e conversão)
- Fórmulas de métricas customizadas
- Reports colaborativos com comentários
- Lexer/integração CDP

**LaunchDarkly**
- Feature flag audit log completo (quem mudou, quando, qual valor)
- Aprovações de mudança de flag (fluxo de review antes de ativar)
- Flag triggers automáticos por threshold de erro (integração DataDog/New Relic)
- Environments por projeto com herança de configuração
- Flag stale detection (flags antigas que deveriam ser removidas)
- SDK edge computing (Cloudflare Workers, Vercel Edge)
- Mobile SDK com streaming SSE para atualizações em tempo real (sem polling)

**Avo**
- Tracking plan centralizado (schema de eventos com nome, propriedades, tipos)
- Code-generated analytics SDK (tipagem forte, sem typos em nomes de eventos)
- Branch e pull request para mudanças no tracking plan
- Inspector (detecta desvios entre o que o código envia e o que o schema define)
- Colaboração entre product, eng e data (comentários, aprovações no tracking plan)

### 2.2 O que o Pulse tem e nenhum concorrente tem

**Estes são os diferenciadores únicos — proteger e expandir:**

1. **Tenant Health Score (0-100) com algoritmo customizável**
   - 4 dimensões: login (30pts), content (40pts), adoption (20pts), subscription (10pts)
   - Calculado diariamente via cron, armazenado por tenant_id
   - Nenhum concorrente tem health score pre-calculado e multi-dimensional nativo
   - PostHog e Mixpanel exigem que o cliente monte isso via fórmulas/SQL

2. **Churn Alert Scheduler integrado**
   - Email automático para admin quando tenant cai abaixo de threshold (40 pts)
   - Resend API com HTML dinâmico, sem configuração de terceiros
   - PostHog tem "subscriptions" mas não tem churn scoring nativo

3. **Rastreamento de custo de API de IA por produto/modelo**
   - Tabela `api_costs` com tokens_in, tokens_out, cost_usd, model
   - Produto específico do IIT que reflete o custo real de features de IA (SocialMake, Libri)
   - Nenhum analytics SaaS inclui tracking de custo LLM — é diferencial para produtos AI-first

4. **All-in-one: flags + events + retention + NPS/CSAT + health + uptime em um serviço**
   - PostHog é o mais próximo mas não tem health score nem custo de IA
   - LaunchDarkly é flags-only; Mixpanel é events-only; Avo é tracking plan-only
   - Pulse unifica sem necessidade de integrações entre ferramentas

5. **Preço em R$ para mercado brasileiro**
   - Free → R$99 → R$297 (sem conversão de câmbio, sem IOF, faturamento via Asaas)
   - Mixpanel e PostHog cobram em USD — a depreciação cambial é dor real para startups BR

---

## 3. Modelo de Negócio dos Concorrentes

### PostHog
- **Modelo:** freemium + usage-based
- **Free:** 1M eventos/mês + 15k session replays (generoso)
- **Pago:** USD 0,00045/evento acima do free (escalona)
- **Estimativa:** empresa com 10M eventos/mês paga ~USD 3.820/mês
- **Self-hosted:** open-source (MIT), sem limite de eventos — financiam via PostHog Cloud
- **Receita 2024:** estimado USD 25M ARR (série B, a2z.vc)

### Mixpanel
- **Modelo:** freemium + por MTU (Monthly Tracked User)
- **Free:** 20M eventos/mês
- **Growth:** USD 28/mês (100 usuários) → escalona por MTU
- **Enterprise:** USD 800+/mês (customizado)
- **Problema:** pricing confuso — MTU vs eventos vs projetos; escala mal para B2C
- **Sem self-hosted**

### LaunchDarkly
- **Modelo:** por MAU (Monthly Active User) que recebe flags
- **Starter:** USD 10/mês (1k MAU)
- **Pro:** USD 20/MAU/mês (estimado para 10k MAU = USD 200/mês)
- **Enterprise:** customizado (USD 2.000+/mês)
- **Cuidado:** cobra por MAU que recebe qualquer flag — escala fast e caro
- **Receita 2024:** ~USD 150M ARR (série D, Salesforce Ventures)
- **Sem self-hosted**

### Avo
- **Modelo:** por seat (usuário da plataforma de tracking plan)
- **Free:** 1 workspace, 1 fonte, 2 usuários
- **Team:** USD 249/mês (5 usuários, múltiplas fontes) — verificar site
- **Business:** USD 750+/mês (customizado)
- **Modelo correto para o IIT:** Avo não é concorrente direto — é ferramenta de engenharia/data, não de observabilidade
- **Sem self-hosted**

---

## 4. Posicionamento Estratégico

### Onde o Pulse NÃO deve competir

- **Session recording/heatmaps:** requer infra pesada (object storage de vídeo, processamento), ROI baixo para o IIT
- **Data warehouse connectors:** PostHog já faz isso bem; não é vantagem competitiva
- **SDKs para 15 linguagens:** não escala para time pequeno; focar em JS/Go/React Native

### Onde o Pulse DEVE competir — nicho defendível

**Pulse = Analytics + Feature Flags + Customer Health para produtos SaaS brasileiros.**

O nicho é: **startup ou empresa brasileira que lança um produto SaaS** e precisa de:
1. Rastrear eventos sem pagar em USD
2. Feature flags com rollout seguro
3. Saber quem vai churnar antes de acontecer
4. Um serviço só, não 3 ferramentas integradas

### Posicionamento recomendado

> "PostHog para equipes que não podem pagar em dólar. Mais simples, com health score de tenant nativo e custo em reais."

Não tentar ser Mixpanel (analytics puro) nem LaunchDarkly (feature flags enterprise). Ser o **all-in-one de analytics + flags para startups BR** com o health score como killer feature de retenção.

### Feature flags vs LaunchDarkly: faz sentido competir?

Sim, mas apenas no segmento SMB/startup. As features atuais do Pulse (rollout percentual, targeting por user/tenant, variants, batch evaluation, scoping por projeto) cobrem 80% dos casos de uso de LaunchDarkly para times de até 20 engenheiros. LaunchDarkly começa a ser indispensável quando o cliente precisa de: audit log de aprovações, flag triggers por métricas de erro, ou edge SDK para Cloudflare Workers. Esses casos são raros em startups brasileiras em estágio inicial.

### Tracking plan (Avo): faz sentido incluir no Pulse?

Não agora. Avo resolve um problema de maturidade de data — quando a empresa tem múltiplos times enviando eventos inconsistentes e precisa de governance. Para o IIT no estágio atual (produtos internos, time pequeno, eventos definidos pelo próprio time que consome o Pulse), a complexidade de um tracking plan não gera valor proporcional. Revisar quando o Pulse tiver 10+ clientes externos enviando eventos.

---

## 5. Conclusão Estratégica

### Prioridade 1 — Completar o que já começou
As funcionalidades mais avançadas (health score, churn alert, CSAT/NPS, A/B variants, cost tracking) já existem no backend mas precisam de maturidade no SDK JavaScript e documentação de integração. Completar antes de adicionar novas features.

### Prioridade 2 — Funis
É a lacuna mais sentida por quem usa analytics de produto. Sem funis, não é possível saber onde usuários dropam no onboarding. Implementar antes de session recording ou heatmaps. Custo de implementação: SQL com window functions no PostgreSQL + UI no frontend.

### Prioridade 3 — SDK React Native
Libri e Nitro são mobile-first. Sem SDK React Native, o Pulse não captura dados dos apps mais importantes do IIT. Deve ser feito em conjunto com o desenvolvimento mobile dos produtos.

### Prioridade 4 — Audit log de flags
Quem habilitou qual flag, quando, com qual rollout. Requisito mínimo para confiar feature flags em produção. LaunchDarkly tem isso como feature central; o Pulse precisa para ser credível.

### Não priorizar agora
Session recording, heatmaps, data warehouse export, tracking plan governance, edge SDK.
