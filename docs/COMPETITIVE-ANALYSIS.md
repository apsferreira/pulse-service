# Pulse — Análise Competitiva (Revisada)

**Data:** Abril 2026
**Versão:** 2.0 — Posicionamento corrigido: Pulse como concorrente LaunchDarkly/VWO
**Autor:** IIT AI Strategy

> **Correção de posicionamento:** A versão anterior (v1.0) enquadrava o Pulse como complemento ao LaunchDarkly e concorrente do PostHog/Mixpanel. Isso estava errado. O Pulse é um **concorrente direto** do LaunchDarkly (feature flags), VWO (A/B testing / experimentation) e PostHog (product analytics) — tudo em um produto self-hosted com pricing em R$.

---

## 1. Estado Atual do Pulse (baseline real)

Inventário baseado nas 16 migrations, handlers e serviços em Go existentes.

### O que o Pulse já tem (backend operacional)

**Feature Flags**
- Tabela `feature_flags` com `key`, `enabled`, `rollout_percentage`, `targeting` (JSONB), `variants` (JSONB)
- Rollout percentual com hash determinístico por `user_id` (migration 007)
- Targeting por `user_ids` e `tenant_ids` via JSONB (migration 007)
- Variantes A/B com peso percentual — `[{key, weight}]` (migration 008, BKL-094)
- Batch evaluation: `POST /flags/evaluate-batch` (até 50 keys por request)
- Scoping por projeto via `project_id` (migration 012)
- `flag_handler.go` com CRUD completo + avaliação contextual

**Event Tracking**
- `POST /track` e `POST /ingest` para eventos customizados
- `POST /identify` para vincular usuário anônimo a user_id conhecido
- Campos: `event_name`, `user_id`, `tenant_id`, `properties` (JSONB), `environment` (migration 013)
- Índices de performance para queries de retenção (migration 005, 006)

**Analytics**
- `RetentionService` + `GetCohorts` por semana/mês
- Dashboard React/TS com: Events, Retention, CSAT, Costs, Uptime, A/B

**Health & Churn**
- `tenant_health_scores` — score 0-100 com 4 dimensões: login, content, adoption, subscription (migration 014)
- `ChurnAlertScheduler` — email automático via Resend quando score < 40

**Surveys**
- NPS: `nps_responses` + triggers configuráveis (migration 015-016)
- CSAT: `csat_surveys` + `csat_responses` (migration 011)

**Infra**
- Multi-tenant lógico via `projects` com `api_key` única (migration 009)
- Tracking de custo de API LLM: tokens_in/out, cost_usd, model (migration 010)
- Uptime monitors com WebSocket (migration 002-003)
- Go/Fiber, PostgreSQL 15, Redis 7 — pronto para self-hosted

**SDK**
- `sdk/js/pulse.js` — módulo ES com `getFlags()`, `isEnabled()`, `track()` (track é no-op)
- `sdk/go/` — SDK Go (documentação pendente)
- **`track()` no JS SDK é no-op** — events do frontend não chegam ao backend ainda

### Gaps críticos identificados (análise do código)

1. **`sdk/pulse.js` usa `export` ES module** — não funciona como `<script src="cdn">`. Para funcionar como GA4/Hotjar precisa de IIFE com `window.Pulse`.
2. **`track()` é no-op** — nenhum evento de frontend chega ao Pulse hoje.
3. **`events` não tem `project_id`** — eventos não são scoped por projeto (gap de multi-tenancy).
4. **Sem funis** — não há análise de abandono por etapa.
5. **Sem análise estatística de experimentos** — variants existem, mas não há p-value, confidence interval nem significância calculada.
6. **Sem audit log de flags** — quem mudou o quê, quando.
7. **Sem environments por projeto em flags** — migration 013 adiciona `environment` em events, mas flags não têm ambiente.
8. **Sem SDK React (`@pulse/react`)** — sem hooks como `useFlag`, `useTrack`.
9. **Sem SDK React Native (`@pulse/react-native`)**.

---

## 2. Feature Matrix Revisada

| Feature | Pulse (atual) | LaunchDarkly | VWO | Avo | PostHog |
|---------|:---:|:---:|:---:|:---:|:---:|
| **FEATURE FLAGS** | | | | | |
| On/off simples | Sim | Sim | Sim | Não | Sim |
| Rollout percentual | Sim | Sim | Sim | Não | Sim |
| Targeting por user/tenant | Sim (JSONB) | Sim (regras avançadas) | Sim | Não | Sim |
| Multivariate flags (N variantes) | Sim (variants JSONB) | Sim | Sim | Não | Sim |
| Ambientes por projeto (staging/prod) | Parcial (events sim, flags não) | Sim (core) | Sim | Não | Sim |
| Audit log de flags | Não | Sim | Não | Não | Sim |
| Aprovação de mudança de flag | Não | Sim (Pro+) | Não | Não | Não |
| Flag stale detection | Não | Sim | Não | Não | Não |
| Flag triggers por métrica de erro | Não | Sim | Não | Não | Não |
| Streaming SSE para flag updates | Não | Sim | Não | Não | Sim |
| **A/B TESTING / EXPERIMENTS** | | | | | |
| A/B com variantes e peso | Sim (JSONB, sem análise) | Sim | Sim | Não | Sim |
| Análise estatística (p-value, CI) | Não | Sim | Sim | Não | Sim (bayesiano) |
| Metric assignment (flag → métrica) | Não | Sim | Sim | Não | Sim |
| Holdout groups | Não | Sim | Não | Não | Não |
| Multi-variate experiments | Parcial (via variants) | Sim | Sim | Não | Sim |
| Heatmaps | Não | Não | Sim | Não | Sim |
| Session recording | Não | Não | Sim | Não | Sim |
| **EVENT TRACKING** | | | | | |
| Ingestão de eventos customizados | Sim | Contexto apenas | Sim | Não (valida, não captura) | Sim |
| Identify (anon → user) | Sim | Sim | Sim | Não | Sim |
| Propriedades via JSONB | Sim | Sim | Sim | Não | Sim |
| Batch ingest | Sim | Sim | Sim | Não | Sim |
| Funis de conversão | Não | Não | Sim | Não | Sim |
| Retenção (DAU/WAU/MAU) | Sim | Não | Parcial | Não | Sim |
| Cohort analysis | Sim | Não | Sim | Não | Sim |
| SQL explorer (query livre) | Não | Não | Não | Não | Sim |
| **TRACKING PLAN / SCHEMA** | | | | | |
| Schema registry de eventos | Não | Não | Não | Sim (core) | Parcial |
| Validação de propriedades | Não | Não | Não | Sim | Não |
| SDK tipado gerado por schema | Não | Não | Não | Sim | Não |
| Branch/PR para tracking plan | Não | Não | Não | Sim | Não |
| Inspector (desvios em runtime) | Não | Não | Não | Sim | Não |
| **SDKs** | | | | | |
| Web via CDN script tag | Não (ES module, não IIFE) | Sim | Sim | Parcial | Sim |
| SDK React (hooks) | Não | Sim | Sim | Sim | Sim |
| SDK React Native | Não | Sim | Não | Sim | Sim |
| SDK iOS nativo | Não | Sim | Sim | Sim | Sim |
| SDK Android nativo | Não | Sim | Sim | Sim | Sim |
| SDK Go | Sim (básico) | Sim | Não | Não | Sim |
| SDK Node.js | Não | Sim | Não | Sim | Sim |
| **INFRA & OPERAÇÃO** | | | | | |
| Self-hosted | Sim | Não | Não | Não | Sim (open-source) |
| Multi-tenant lógico | Sim (projects + api_key) | Sim | Sim | Sim | Sim |
| Rate limiting por projeto | Não | Sim | Sim | Sim | Sim |
| Webhooks de eventos | Não | Sim | Sim | Não | Sim |
| Export de dados (CSV/JSON) | Não | Sim | Sim | Não | Sim |
| Data warehouse export | Não | Sim | Não | Não | Sim |
| **FEATURES EXCLUSIVAS DO PULSE** | | | | | |
| Tenant health score (0-100) | Sim | Não | Não | Não | Não |
| Churn alert automático | Sim | Não | Não | Não | Não |
| LLM API cost tracking | Sim | Não | Não | Não | Não |
| NPS/CSAT nativo integrado | Sim | Não | Não | Não | Sim |
| Uptime monitors integrado | Sim | Não | Não | Não | Não |
| **COMERCIAL** | | | | | |
| Pricing em R$ | Sim | Não (USD) | Não (USD) | Não (USD) | Não (USD) |
| Free tier self-hosted | Sim | Não | Não | Não | Sim |
| Plano pago inicial | R$99/mês | USD 10/mês (1k MAU) | USD 199/mês | USD 249/mês | USD 0 (events-based) |

---

## 3. Gap Analysis Priorizado

### 3.1 Gaps críticos vs LaunchDarkly (feature flags)

**O que o LaunchDarkly tem que o Pulse não tem:**

| Gap | Impacto | Esforço | Detalhe |
|-----|---------|---------|---------|
| Ambientes por projeto em flags | Alto | Baixo | Migration + campo `environment` em `feature_flags`. Events já têm (migration 013). Falta replicar em flags. |
| Audit log de flags | Alto | Médio | Tabela `flag_changes` com actor, timestamp, diff do estado anterior. Requisito para usar flags em produção com segurança. Sem isso, não há rastreabilidade. |
| Streaming SSE para updates de flags | Médio | Médio | Polling no SDK é aceitável para MVP. SSE reduz latência de propagação de flag de ~30s para <1s. LaunchDarkly cobra isso como vantagem core. |
| Aprovação de mudança de flag | Médio | Alto | Fluxo de review antes de ativar em prod. Relevante para times com múltiplos engenheiros. Não é bloqueador de MVP. |
| Flag triggers por métrica de erro | Baixo | Alto | Desligar flag automaticamente quando error rate sobe. Requer integração com observabilidade externa. Não priorizar. |
| Flag stale detection | Baixo | Baixo | Cron que marca flags não avaliadas em 30d. Fácil de implementar, boa prática. |

**Veredito LaunchDarkly:** O Pulse cobre ~75% dos casos de uso do LaunchDarkly para times de até 30 engenheiros. O audit log é o único gap bloqueador de produção. O ambiente por projeto é estrutural mas implementável em horas.

### 3.2 Gaps críticos vs VWO (A/B testing / experimentation)

**O que o VWO tem que o Pulse não tem:**

| Gap | Impacto | Esforço | Detalhe |
|-----|---------|---------|---------|
| Análise estatística de experimentos | Muito alto | Alto | O Pulse tem variantes com peso, mas não calcula p-value, confidence interval nem potência estatística. Sem isso, não é experimentação — é só divisão aleatória de tráfego. |
| Metric assignment (flag → métrica de negócio) | Alto | Médio | Vincular um experimento a uma métrica (ex: flag `novo-checkout` → taxa de conversão). Requer JOIN entre avaliações de flag e eventos. |
| Heatmaps | Médio | Muito alto | Requer captura de cliques/scrolls no DOM via SDK JS e renderização. Infra pesada. Não priorizar no curto prazo. |
| Session recording | Médio | Muito alto | Requer gravação de DOM mutations (rrweb ou similar) + object storage de sessões. Fora do escopo Tier 1-2. |

**Veredito VWO:** O gap mais crítico é a análise estatística. Ter variantes sem significância calculada não é A/B testing — é feature flag com split. Para ser alternativa ao VWO no core, precisa de pelo menos: p-value, confidence interval e indicador visual de significância.

### 3.3 SDK web via CDN — o que falta para ter um `<script>` funcional

**Problema atual:** `sdk/pulse.js` usa `export { PulseSDK }` (ES module). Não funciona com `<script src="cdn">` — navegadores antigos não suportam e a sintaxe exige bundler ou `type="module"`.

**O que falta:**

1. **Reescrever como IIFE** — `(function(window){ window.Pulse = {...} })(window)` — o mesmo padrão do GA4, Hotjar e Segment.
2. **Implementar `track()` de verdade** — hoje é no-op. A função precisa enviar um POST para o backend com `event_name`, `properties`, `user_id` e `api_key`.
3. **Auto-captura de pageviews** — ao carregar, capturar `page_view` com URL, referrer e user agent automaticamente.
4. **Identify** — `window.Pulse.identify(userId, traits)` para resolver anônimo → usuário.
5. **Flag evaluation com contexto** — `window.Pulse.isEnabled(flagKey, context)` passando `userId` para que o backend faça hash determinístico e targeting.
6. **Fallback offline** — se o endpoint não responder, flags retornam `false` (safe default). Eventos são descartados silenciosamente (não bloquear o site do cliente).
7. **Snippet de instalação** — como GA4, um snippet de 10 linhas que carrega o script assincronamente sem bloquear o parser HTML.

**Esforço estimado:** 3-5 dias de desenvolvimento para um SDK funcional sem heatmaps/recording.

### 3.4 Gaps vs PostHog (product analytics — referência)

O PostHog é self-hosted open-source e é a referência mais próxima do Pulse em modelo. Gaps principais:

- **Funis** — ausência mais crítica. Sem funis não há análise séria de onboarding ou conversão.
- **SDK React com hooks** — PostHog tem `usePostHog()`, `useFeatureFlagEnabled()`. O Pulse precisa de `@pulse/react` com hooks equivalentes.
- **SDK React Native maduro** — PostHog tem SDK com auto-capture para Expo. Fundamental para Libri e Nitro.
- **SQL explorer** — query builder sobre eventos brutos. Diferencial para times de dados. Não priorizar agora.

---

## 4. Decisão sobre Avo (tracking plan governance)

**Contexto:** Avo é um produto de governance para tracking plans — define o schema de eventos (nome canônico, propriedades obrigatórias, tipos) e gera SDKs tipados para que engenheiros não cometam erros como `user_signed_up` vs `UserSignedUp` vs `user-signed-up`.

**Faz sentido incluir no Pulse?**

**Resposta curta: não agora, mas está no caminho natural do produto.**

**Argumento contra (prioridade atual):**
- Avo resolve um problema de maturidade de dados — relevante quando há múltiplos times enviando eventos inconsistentes e um time de data separado tentando governar.
- Para o IIT (produtos internos, time pequeno), o schema de eventos é simples e controlado pelo mesmo time que lê o Pulse.
- Implementar tracking plan governance antes de ter funis e SDK web funcional é inverter a ordem de prioridades.
- Avo é um produto B2B de engenharia — seu comprador é o Head of Data ou Staff Engineer, não o fundador de uma startup BR.

**Argumento a favor (futuro):**
- Quando o Pulse tiver clientes externos com múltiplos engenheiros enviando eventos, a inconsistência de schema vai aparecer como problema.
- Um schema registry leve (sem a complexidade de branch/PR do Avo) pode ser implementado como tabela `event_schemas` com validação opcional no `POST /ingest`.
- Seria um diferencial vs PostHog, que tem gestão de dados mas não schema governance nativa.

**Decisão:** incluir um **schema registry leve** no Tier 3 (12 meses) — não como clone do Avo, mas como validação opcional de schema no `POST /ingest`. O Pulse não precisa dos features de colaboração e branching do Avo para ser útil.

---

## 5. Roadmap Revisado por Tier

### Tier 1 — MVP competitivo (60 dias)
Mínimo para ser uma alternativa real ao LaunchDarkly para SaaS brasileiros. Foco em feature flags robustas + SDK web funcional.

| Item | Esforço | Impacto | Justificativa |
|------|---------|---------|---------------|
| **`pulse.js` como IIFE via CDN** — reescrever como script tag, implementar `track()` real | 3-5 dias | Muito alto | Sem isso o Pulse não captura nenhum evento de frontend. Todo o tracking é cego para apps web. |
| **Ambientes por projeto em flags** — migration + campo `environment` em `feature_flags` | 1 dia | Alto | LaunchDarkly core feature. Ter staging/prod separados é requisito básico para usar flags com segurança. |
| **Audit log de flags** — tabela `flag_changes` + UI | 3 dias | Alto | Quem ativou qual flag, quando. Requisito de rastreabilidade para produção. |
| **`track()` real no SDK** — POST para `/ingest` com `api_key`, `event_name`, `properties` | 1 dia | Alto | Necessário junto com a reescrita IIFE. Sem isso, o SDK é inútil para analytics. |
| **Análise estatística básica de experimentos** — p-value + confidence interval no backend | 5 dias | Alto | Sem significância calculada, variantes A/B são divisão de tráfego, não experimentação. Usa chi-squared test (implementável em Go puro). |
| **`project_id` em events** — migration + filtros nos handlers | 1 dia | Alto | Gap estrutural de multi-tenancy. Eventos não scoped = dados misturados entre projetos. |
| **Middleware de autenticação por API key** — resolver `project_id` via `X-Pulse-Key` | 2 dias | Alto | Pré-requisito para qualquer cliente externo usar o SDK. Já existe na migration 009, falta enforcement. |

**Resultado do Tier 1:** Pulse com feature flags rastreáveis por ambiente, SDK web funcional, ingestão de eventos por projeto e A/B testing com significância. Alternativa real ao LaunchDarkly starter para times de até 10 engenheiros.

### Tier 2 — Paridade funcional (6 meses)
Paridade com LaunchDarkly core + VWO básico + PostHog analytics essencial.

| Item | Esforço | Impacto | Justificativa |
|------|---------|---------|---------------|
| **`@pulse/react`** — pacote NPM com hooks: `useFlag`, `useTrack`, `useExperiment` | 1 semana | Muito alto | React é o stack do IIT e de 90% dos SaaS BR. Sem hooks, a DX é ruim. |
| **`@pulse/react-native`** — wrapper com AsyncStorage para Expo/RN | 1 semana | Muito alto | Libri e Nitro são mobile. Hoje o Pulse não captura dados desses apps. |
| **Funis de conversão** — SQL com window functions + UI | 2 semanas | Alto | Funcionalidade mais pedida em product analytics. Sem funis, não há análise de onboarding. |
| **Metric assignment** — vincular experimento a evento de conversão | 1 semana | Alto | Transformar A/B testing em experimentação real de negócio. |
| **Rate limiting por projeto** — Redis token bucket | 2 dias | Alto | Sem rate limiting, um cliente abusivo afeta todos no modo SaaS. |
| **Quota enforcement** — bloquear eventos acima do limite do plano | 3 dias | Alto | Pré-requisito para plano Free público sem risco de abuso. |
| **Environments por projeto em flags (UI)** — seletor de ambiente no dashboard | 3 dias | Médio | Backend feito no Tier 1, falta frontend. |
| **Flag stale detection** — cron que marca flags não avaliadas em 30d | 1 dia | Médio | Boa prática, fácil de implementar. |
| **Webhooks de eventos** — POST para URL configurada pelo cliente | 3 dias | Médio | Permite integrações com Slack, n8n, sistemas internos do cliente. |
| **Dashboard por projeto** — seletor + overview isolado | 1 semana | Alto | Hoje o dashboard é hardcoded para o IIT. Para SaaS, cada cliente vê só seus dados. |
| **SDK Go maduro** — publish em pkg.go.dev + exemplos | 2 dias | Médio | SDK existe, falta documentação e versionamento semântico. |

**Resultado do Tier 2:** Pulse com paridade funcional para 80% dos casos de uso do LaunchDarkly + VWO básico + PostHog analytics essencial. Pronto para primeiros clientes pagantes externos.

### Tier 3 — Diferenciação (12 meses)
O que o Pulse pode ter que LaunchDarkly e VWO **não têm** — diferenciação defensável.

| Item | Esforço | Impacto | Justificativa |
|------|---------|---------|---------------|
| **Funis de feature flags** — correlação bandeira ativa → conversão no funil | 2 semanas | Muito alto | Único no mercado. Ver se ativar uma flag aumentou conversão. LD e VWO não têm isso nativo. |
| **Health score configurável via UI** — pesos e thresholds por projeto | 1 semana | Alto | Hoje hardcoded em Go. Clientes SaaS precisam de health score para o perfil do seu produto. |
| **Churn prediction por ML leve** — regressão logística sobre health score + eventos | 3 semanas | Alto | Evoluir o churn alert de threshold fixo para probabilidade calculada. Diferencial para produtos B2B. |
| **Schema registry leve** — tabela `event_schemas` com validação opcional no `/ingest` | 1 semana | Médio | Tracking plan governance leve, sem a complexidade do Avo. Evita eventos inconsistentes em escala. |
| **Approval flow para flags** — request de review antes de ativar em prod | 2 semanas | Médio | LaunchDarkly Pro feature. Para o segmento SMB BR, relevante apenas acima de 15 engenheiros. |
| **NPS triggered por evento** — disparar survey após evento customizável | 3 dias | Médio | Hoje trigger é `after_30_days` hardcoded. Clientes precisam triggerar após primeira compra, ativação, etc. |
| **Export de dados** — CSV/JSON dos eventos brutos | 3 dias | Médio | Clientes querem levar dados para BI próprio. Requisito de confiança para adoção. |
| **Streaming SSE para flag updates** — eliminar polling no SDK | 1 semana | Médio | Reduz latência de propagação de ~30s para <1s. Útil para flags de kill switch emergencial. |

**Resultado do Tier 3:** Pulse com diferenciadores únicos que LaunchDarkly e VWO não cobrem nativamente — especialmente a correlação flag→funil e o health score configurável. Pronto para posicionamento competitivo agressivo no mercado BR.

---

## 6. Posicionamento Estratégico Revisado

### Onde o Pulse compete diretamente

| Concorrente | Segmento de overlap | Vantagem do Pulse |
|------------|---------------------|-------------------|
| LaunchDarkly | Feature flags para SaaS SMB | Self-hosted, pricing em R$, analytics integrado |
| VWO | A/B testing para SaaS | Self-hosted, flags integradas, sem USD |
| PostHog | Product analytics self-hosted | Health score, churn alert, custo LLM — features B2B que PostHog não tem |

### Onde o Pulse não compete (deliberadamente)

- **Session recording / heatmaps** — infra pesada (object storage de DOM snapshots), ROI baixo. PostHog e VWO cobrem quem precisa.
- **Data warehouse connectors** — segmento enterprise. PostHog já faz; não é vantagem competitiva no mercado BR early-stage.
- **Edge SDK (Cloudflare Workers)** — não é demanda do mercado BR early-stage.
- **Tracking plan governance completo (Avo-style)** — útil apenas com 10+ clientes com times de data separados.

### Nicho defendível

> **"LaunchDarkly + VWO + PostHog em um produto, self-hosted, com pricing em R$ e health score de cliente nativo."**

Mercado-alvo: **startup ou scale-up brasileira com produto SaaS** que precisa de:
1. Feature flags com rollout seguro e rastreabilidade
2. A/B testing com significância estatística
3. Analytics de produto sem pagar em USD
4. Saber quem vai churnar antes que aconteça
5. Um serviço só, sem 3-4 ferramentas integradas

### Por que self-hosted é o diferencial de distribuição

- LaunchDarkly e VWO não têm opção self-hosted
- PostHog tem, mas requer Kubernetes ou infra pesada para escala
- O Pulse roda com `docker compose up` em um VPS de USD 20/mês
- Para startups BR preocupadas com LGPD, dados no próprio servidor é argumento de venda
- Para clientes enterprise BR (fintechs, healthtechs), self-hosted pode ser requisito contratual

---

## 7. Modelo de Negócio dos Concorrentes (atualizado)

### LaunchDarkly
- **Modelo:** por MAU que recebe flags (Monthly Active User)
- **Starter:** USD 10/mês (1k MAU)
- **Pro:** USD 20/MAU/mês estimado — 10k MAU = USD 200/mês
- **Enterprise:** customizado (USD 2.000+/mês) com audit logs, aprovações, SSO
- **Sem self-hosted**
- **Receita 2024:** ~USD 150M ARR

### VWO
- **Modelo:** por MAU testado
- **Starter:** USD 199/mês (10k sessões/mês)
- **Growth:** USD 399/mês (50k sessões)
- **Enterprise:** customizado
- **Sem self-hosted**
- **Heatmaps e session recording incluídos no Growth+**

### Avo
- **Modelo:** por seat (usuário do tracking plan)
- **Free:** 1 workspace, 2 usuários
- **Team:** USD 249/mês (5 usuários)
- **Business:** USD 750+/mês
- **Sem self-hosted**
- **Comprador:** Head of Data ou Staff Engineer — não é o mesmo ICP do Pulse

### PostHog
- **Modelo:** freemium + usage-based por evento
- **Free:** 1M eventos/mês + 15k session replays
- **Pago:** USD 0,00045/evento acima do free
- **Self-hosted:** open-source (MIT), sem limite, sem suporte
- **Receita 2024:** estimado USD 25M ARR

### Pulse (proposta)
- **Modelo:** Free (self-hosted ilimitado) + planos cloud por projeto/eventos
- **Free Self-Hosted:** sem limite, sem suporte (como PostHog)
- **Growth Cloud:** R$99/mês (500k eventos, 3 projetos, 10k MAU flags)
- **Pro Cloud:** R$297/mês (5M eventos, projetos ilimitados, 100k MAU flags, health score)
- **Vantagem:** pricing em R$ sem conversão de câmbio, billing via Asaas

---

## 8. Conclusão Estratégica

### Prioridade 1 — SDK web funcional (Tier 1, semana 1-2)
O `pulse.js` atual é um módulo ES que não funciona como `<script>` de CDN e tem `track()` como no-op. Sem corrigir isso, o Pulse não captura nenhum dado de frontend. É o gap mais bloqueador — torna toda a infra de tracking inútil para apps web.

### Prioridade 2 — Análise estatística de experimentos (Tier 1, semana 3-4)
Variantes A/B existem no backend, mas sem p-value e confidence interval não é experimentação — é divisão de tráfego. Implementar chi-squared test em Go puro para calcular significância. Sem isso, o Pulse não pode se posicionar como alternativa ao VWO.

### Prioridade 3 — Audit log e ambientes de flags (Tier 1, semana 2-3)
Dois itens de baixo esforço e alto impacto para paridade com LaunchDarkly. Sem audit log, usar flags em produção é arriscado (quem mudou o quê?). Sem ambientes, não há separação staging/prod.

### Prioridade 4 — SDKs React e React Native (Tier 2, mês 2-3)
Sem `@pulse/react` e `@pulse/react-native`, os próprios produtos do IIT (SocialMake, Libri, Nitro) não conseguem usar o Pulse de forma idiomática. É o pré-requisito para dogfooding real.

### Prioridade 5 — Funis de conversão (Tier 2, mês 3-4)
A lacuna mais sentida por qualquer equipe de produto. Sem funis, não há análise de onboarding ou abandono. Implementação é SQL com window functions + UI — esforço controlado, impacto alto.

### Não priorizar (horizonte 12+ meses)
Session recording, heatmaps, data warehouse export, edge SDK, tracking plan governance completo.
