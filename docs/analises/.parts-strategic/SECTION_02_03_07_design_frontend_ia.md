# Pulse Service — Seções Estratégicas SDD 2026-06-05
## Lentes 02 (Design), 03 (Frontend) e 07 (IA)

---

## Lente 07 — IA (score: 7/10)

### O que está implementado

O pulse-service tem uma camada de IA surpreendentemente densa para um serviço de feature flags self-hosted. São 7 handlers de IA distintos, totalizando ~908 linhas de lógica de ML puro em Go, sem dependências de biblioteca externa de ML:

| Handler | BKL | Mecanismo | Status |
|---------|-----|-----------|--------|
| `churn_model_handler.go` | BKL-999 | Regressão logística com gradient descent (100 iter, lr=0.01), 5 features por usuário, janela 90 dias | Implementado |
| `churn_prediction_handler.go` | BKL-1065 | Heurística por projeto (complementar ao modelo treinável) | Implementado |
| `bayesian_handler.go` | BKL-1001 | Análise Bayesiana com distribuições Beta para A/B, calcula P(B > A) | Implementado |
| `mab_handler.go` | BKL-1154 | Multi-Armed Bandit epsilon-greedy com recálculo de pesos por evento | Implementado |
| `anomaly_handler.go` | BKL-865 | Média móvel + desvio padrão; Gemini 2.5 Flash gera explicação human-readable | Implementado |
| `ai_config_handler.go` | BKL-1064 | Feature flags para prompts LLM (model, temperature, max_tokens, system_prompt, versioning com rollback) | Implementado |
| `llm_observability_handler.go` | BKL-1063 | Métricas P50/P99, custo USD/BRL, error rate por provider, agrupamento por modelo e feature | Implementado |
| `insight_handler.go` | — | Insights gerados por projeto (estrutura básica) | Implementado |

### Qualidade e maturidade

**Pontos fortes:**

- O modelo de churn (BKL-999) é implementação própria em Go puro, sem dependência de scipy/sklearn — viável para rodar no homelab sem overhead de runtime Python. Label clara: churned = 1 se zero eventos nos últimos 14 dias. Features coerentes com comportamento de produto (event_count_30d, days_since_last_event, distinct_events).
- O MAB epsilon-greedy (BKL-1154) tem controle de plano integrado: free=0 experimentos, starter=2, growth/pro/enterprise=ilimitado. Isso é produto, não só feature.
- LLM Observability (BKL-1063) é o diferencial mais defensável: nenhum concorrente self-hosted (GrowthBook, Unleash, Flagsmith) tem isso. Cobre P50/P99 de latência + custo USD/BRL por provider + error rate. Dados em PostgreSQL, sem dependência de Datadog ou similar.
- AI Configs (BKL-1064) resolve um problema real dos times IIT: fazer deploy de prompt sem redeploy de código. Versioning com rollback é a feature certa.
- Anomalia usa Gemini 2.5 Flash para explicação — corretamente separado da detecção estatística. A detecção é determinística (não falha se API cair). Fallback implícito existe (explicação padrão sem LLM).

**Gaps de maturidade:**

1. O modelo de regressão logística (BKL-999) treina on-demand via endpoint POST. Não há job agendado de retreino. Se ninguém chamar `/churn-model/train`, o modelo fica obsoleto — e não há alerta para isso.
2. A análise Bayesiana (BKL-1001) é stateless: recebe os dados via query params e calcula na hora. Não persiste resultados históricos de experimentos. Isso impede análise de tendência ao longo do tempo.
3. Não há integração entre MAB e AI Configs. O caso natural seria: MAB seleciona variante de prompt, AI Config entrega o prompt — hoje são dois sistemas desconexos.
4. O `insight_handler.go` tem apenas 58 linhas e estrutura básica. Não gera insights acionáveis com LLM — é um esqueleto.
5. Nenhum dos modelos de ML tem monitoramento de drift. O modelo de churn pode degradar silenciosamente se o comportamento dos usuários mudar.

### Top 3 movimentos

**Movimento 1 — Retreino agendado de churn (baixo custo, alto valor operacional)**
Adicionar um job Go (`time.Ticker` ou cron integrado no main) que chama a lógica de treino semanalmente por projeto ativo. Sem chamada de API externa, zero custo adicional. Resolve o maior risco de degradação silenciosa. Custo: 1-2 dias de backend.

**Movimento 2 — Insights acionáveis com Haiku via AI Config (custo explícito)**
O `insight_handler.go` deve gerar 3-5 insights por projeto usando Haiku, consumindo os dados já calculados (anomalias, churn score, MAB weights). Prompt fixo no AI Config do próprio Pulse (dogfooding). Estimativa: ~2K tokens input + ~300 tokens output por projeto por dia. A $1/MTok input + $5/MTok output = $0.003/projeto/dia. Para 50 projetos ativos: $0.15/dia = $4.50/mês. Requer aprovação CEO para ativar (R4).

**Movimento 3 — Conectar MAB a AI Configs (diferencial de produto)**
Quando um experimento MAB tem variantes do tipo "prompt", o sistema deve buscar o AI Config correspondente e injetar o prompt correto na resposta do `SelectVariant`. Isso cria um loop fechado: MAB seleciona qual prompt performou melhor, AI Config entrega o prompt, evento de conversão alimenta o MAB de volta. Nenhum concorrente tem isso. Custo: 3-4 dias de backend + 1 dia de SDK.

---

## Lente 02 — Design (score: 5/10)

### Dashboard admin: estado atual

O frontend tem 22 páginas, cobrindo uma superfície muito maior do que a maioria dos feature flag tools self-hosted. As principais:

- `FeatureFlags.tsx` (43.8KB) — a maior página, sugere complexidade de UX concentrada em um único arquivo
- `ABTesting.tsx` (36.4KB) — A/B testing com `SampleSizeCalculator` e `RetentionMatrix` como componentes dedicados
- `Users.tsx` (34.4KB) e `Events.tsx` (33.1KB) — análise de usuários e eventos
- `CSAT.tsx` (31.7KB) — satisfação de clientes
- `Billing.tsx` e `Pricing.tsx` — gestão de plano

Sem acesso a screenshots, a avaliação parte de indícios estruturais: páginas com 30-43KB de TSX são densas demais para uma única responsabilidade. O componente `FlagTable.tsx` (14.4KB) + `FlagModal.tsx` (21.9KB) + `RuleBuilder.tsx` (11.5KB) como componentes separados é um bom sinal de decomposição. O `OnboardingWizard.tsx` (6.2KB) indica atenção ao primeiro contato.

### Gaps vs LaunchDarkly

| Dimensão | LaunchDarkly | Pulse (estimado) | Gap |
|----------|-------------|-----------------|-----|
| Targeting visual | Builder drag-and-drop com preview em tempo real | RuleBuilder component — provavelmente form-based | Provável gap de interatividade |
| Flag lifecycle | Status visual claro (archived, deprecated, stale) com alertas automáticos | Não confirmado | Desconhecido |
| Audit trail | Timeline visual por flag com diff de mudança | `FlagAudit.tsx` existe (10KB) | Provavelmente básico |
| Approval flow | Workflow nativo com reviewers e comentários | `Approvals.tsx` (18.8KB) existe | Provavelmente funcional |
| Search e filtros | Busca full-text + filtros multi-dimensão + tags | Não confirmado na estrutura | Provável gap |
| Dark launch preview | Simulação de quem veria qual variante | Não encontrado na estrutura | Gap real |

### Gaps de UX críticos identificados

1. `FeatureFlags.tsx` com 43.8KB em arquivo único indica que targeting rules, variantes, segmentos e configurações estão misturados. Para o usuário, isso provavelmente significa uma modal longa e densa — o fluxo de criação de flag não está guiando, está cobrando.
2. Não há página dedicada a `AIConfigs` no frontend. O handler BKL-1064 existe no backend mas não aparece nas páginas listadas. Feature premium sem UI = feature invisível.
3. Não há página de `MAB` no frontend. O handler BKL-1154 também sem UI. Dois dos maiores diferenciais do Pulse são backend-only.
4. `StatusPage.tsx` (9.2KB) existe, mas `Uptime.tsx` (25.4KB) é uma página separada — separação de conceitos que pode confundir o usuário sobre onde ver status de serviço.

### Top 3 movimentos de design

1. Criar página "AI Studio" no dashboard que unifica AI Configs + LLM Observability + MAB. Hoje são três backends sem UI. Uma página única consolida o diferencial competitivo e o torna tangível para o usuário.
2. Quebrar `FeatureFlags.tsx` em wizard de criação passo-a-passo: (1) nome/descrição, (2) tipo de variante, (3) targeting rules, (4) rollout strategy. Reduz a carga cognitiva na criação de flag.
3. Adicionar "stale flag detector" visual: flags sem mudança há 90 dias aparecem com badge de alerta. Implementação trivial com dados já disponíveis (audit trail). Cria hábito de limpeza de débito técnico de feature flags.

---

## Lente 03 — Frontend (score: 6/10)

### Stack

- React 18.2 + TypeScript + Vite 6.3 + TailwindCSS 3.4 + shadcn/ui (inferido pelo tailwind.config.js e uso de clsx/tailwind-merge)
- React Router v6, React Hook Form + Zod, Axios, Lucide React, react-hot-toast
- Sem Tanstack Query / React Query — estado de servidor provavelmente gerenciado com useState + useEffect + Axios direto
- Sem biblioteca de charts mencionada no package.json — visualizações provavelmente construídas com CSS/SVG puro ou inline

Stack coerente com o padrão IIT. A ausência de React Query é o maior gap técnico: sem cache de servidor, sem invalidação automática, sem loading/error states padronizados. Para um dashboard de analytics com muitas queries paralelas, isso causa UX inconsistente (loading states ad-hoc por página).

### Cobertura de features via UI

| Feature de backend | UI disponível |
|-------------------|--------------|
| Feature Flags (CRUD + rules) | Sim (FeatureFlags.tsx + FlagModal + RuleBuilder) |
| A/B Testing | Sim (ABTesting.tsx) |
| Events tracking | Sim (Events.tsx) |
| Users | Sim (Users.tsx) |
| CSAT | Sim (CSAT.tsx) |
| Retention | Sim (Retention.tsx + RetentionMatrix) |
| Billing | Sim (Billing.tsx + Pricing.tsx) |
| Uptime / Status | Sim (Uptime.tsx + StatusPage.tsx) |
| LLM Observability | Nao |
| AI Configs | Nao |
| MAB Experiments | Nao |
| Churn Model | Nao |
| Anomaly Detection | Nao |
| Query Explorer | Sim (QueryExplorer.tsx) |
| Session Replay | Sim (Replay.tsx) |

Cobertura: ~65% das features de backend têm UI. As 5 features sem UI são exatamente as de IA — o diferencial competitivo do produto.

### SDK — Developer Experience (DX)

**Breadth:** 9 SDKs declarados (JS, React, React Native, Go, Node, Python, iOS, Android, Flutter). Para um produto early-stage, isso é overextension de surface sem profundidade.

**JS SDK (`@iit/pulse-sdk`):**
- Pacote bem estruturado: dual output (ESM + CJS), tipos TypeScript, publicação via GHCR
- `index.ts` (17KB) sugere cobertura ampla: feature flags + event tracking + session replay (`session-replay.ts` separado)
- Sem testes automatizados visíveis no package.json (sem jest/vitest no devDependencies)
- `badge.ts` (4.1KB) — provavelmente geração de badge de status, diferencial visual interessante

**React SDK (`@pulse/react`):**
- Nome de pacote inconsistente: JS é `@iit/pulse-sdk`, React é `@pulse/react` — isso quebra a experiência do desenvolvedor que precisa procurar no registro
- Sem dependências além de React peer — leve, correto
- `index.tsx` (18.9KB) — hooks e providers

**Go SDK:**
- `pulse.go` (9.3KB) + `go.mod` — básico mas funcional para integração server-side

**Gaps de DX:**
- Sem playground/sandbox interativo para testar flags sem integrar o SDK
- Sem documentação de referência gerada (sem typedoc no JS SDK)
- Inconsistência de naming entre pacotes quebra a percepção de suite coesa

### Top 3 movimentos de frontend

**Movimento 1 — AI Studio UI (impacto alto, urgente)**
Criar `AIStudio.tsx` com três tabs: "AI Configs" (CRUD de prompts com editor de código), "LLM Observability" (gráficos de latência P50/P99 e custo), "Experimentos MAB" (lista + criação). Adicionar ao `Layout.tsx` com ícone de estrela para destacar como feature premium. Os endpoints já existem — é puro frontend.

**Movimento 2 — Adicionar Tanstack Query (impacto médio, fundação)**
Substituir o padrão useState + useEffect + Axios por `@tanstack/react-query`. Benefícios diretos: cache automático entre páginas, refetch em background, loading/error states unificados, prefetch de dados críticos. Dashboard de analytics sem cache de cliente é lento e inconsistente. Estimativa: 2-3 dias para migrar as páginas principais.

**Movimento 3 — Unificar naming dos SDKs e adicionar Quickstart interativo**
Renomear `@pulse/react` para `@iit/pulse-react` (consistência com `@iit/pulse-sdk`). Criar `snippet/` SDK como base para um Quickstart de 3 linhas no onboarding do dashboard — quando o usuário cria o primeiro projeto, ele vê o código de integração gerado com o seu project_id preenchido. Reduz time-to-value do SDK de horas para minutos.
