# Pulse vs LaunchDarkly — Análise Competitiva e Pitch

**Data:** 2026-04-13  
**Contexto:** Análise comparativa baseada em screenshots reais da LP e console do LD.  
**Uso:** Referência para o time comercial, POC com cliente potencial self-hosted, e priorização de roadmap.

---

## 1. Posicionamento do LaunchDarkly

O LD saiu de "feature flag tool" para **"runtime control platform"** — controle em tempo real de features, AI behavior e customer experience, sem redeploy.

Headline atual: *"Move at AI speed. Stay in control."*

O que está funcionando na abordagem deles:
- **Credibilidade em camadas:** logos (Volvo, Samsung, Hulu) → números (45T avaliações/dia) → cases (Paramount 100X produtividade)
- **Três pilares por persona:** Built for developers / Backed by data / Trusted in production
- **Tangibilização:** code snippet real + interface screenshot + mascote interativo (Toggle)
- **Dado de impacto de negócio:** 48% aumento em confiabilidade de software (não é métrica de produto, é ROI)

---

## 2. Gap Analysis — Pulse vs LaunchDarkly

### 2.1 Feature Flags

| Feature | Pulse | LD | Criticidade |
|---------|-------|----|-------------|
| On/off + rollout percentual (SHA-256) | ✅ | ✅ | — |
| Targeting por user_id / tenant_id | ✅ básico | ✅ completo | Alta |
| Regras compostas (atributo + operador) | ❌ | ✅ | P1 — BKL-1153 |
| Múltiplos context kinds | ❌ | ✅ | Média |
| **Guarded Rollouts (auto-rollback)** | ❌ | ✅ | **P0 — BKL-1151** |
| Live Events stream | ❌ | ✅ | Média |
| Flag lifecycle / cleanup | ❌ | ✅ | Baixa |

### 2.2 Experiments

| Feature | Pulse | LD | Criticidade |
|---------|-------|----|-------------|
| A/B Testing básico | ✅ | ✅ | — |
| Análise bayesiana | ✅ | ✅ | — |
| **Multi-armed bandits** | ❌ | ✅ | P1 — BKL-1154 |
| Métricas definidas por experiment | ⚠️ | ✅ | Alta |
| Event Explorer | ❌ | ✅ | Média |

### 2.3 Observability / Analytics

| Feature | Pulse | LD | Criticidade |
|---------|-------|----|-------------|
| Session Replay | ✅ | ✅ | — |
| Correlação Replay + Flag Variant | ✅ único | ❌ | **Diferencial Pulse** |
| Error tracking | ❌ | ✅ | P1 |
| Logs / Traces | ❌ | ✅ | Baixa |
| Dashboards configuráveis | ❌ | ✅ | Média |
| Alerts | ❌ | ✅ | Alta |

### 2.4 Integrações e Plataforma

| Feature | Pulse | LD | Criticidade |
|---------|-------|----|-------------|
| **BigQuery export** | ❌ | ✅ | **P0 — BKL-1150** |
| Slack / Datadog / webhook | ❌ | ✅ | P1 — BKL-1152 |
| 60+ integrações | ❌ | ✅ | Baixa |
| **Self-hosted total** | ✅ único | ❌ | **Diferencial Pulse** |
| Relay Proxy | ❌ | ✅ | P1 |
| SDK: JS/React/Node | ✅ | ✅ | — |
| SDK: Flutter/PHP/Ruby | ❌ | ✅ | P1 |
| Pricing sem penalidade por MAU | ✅ único | ❌ | **Diferencial Pulse** |

---

## 3. Dois P0s que Bloqueiam Qualquer Conversa com a Loft

O time da Loft usa **LaunchDarkly + BigQuery** com 46 projetos e 2M+ MAU. Esses são os dois bloqueadores reais:

### P0-A: BigQuery Export (BKL-1150)

O time de dados tem pipelines inteiros construídos em cima da exportação de flag evaluations e experiment events para BigQuery. Sem isso:
- Qualquer A/B test fica sem análise de dados pelo time de Data Science
- A prova de impacto de negócio de cada feature não funciona
- O dashboard de experimentação da Loft fica vazio

**Implementação mínima viável:**
1. Buffer de eventos em Redis (flag evaluation, experiment assignment, webhook event)
2. Worker que drena para BigQuery Data Write API ou Pub/Sub topic (cliente configura destino)
3. Schema fixo e documentado
4. Self-hosted: variável de ambiente `BIGQUERY_PROJECT` / `PUBSUB_TOPIC`

**Critério:** Eventos chegam no BigQuery em <5min de latência. Funciona em cloud e self-hosted.

### P0-B: Guarded Rollouts / Auto-rollback (BKL-1151)

Para um time de 2M+ MAU, rollout manual sem guarda automática é inaceitável. O LD detecta regressão de error rate / latência e reverte a flag automaticamente. Sem isso:
- O eng on-call precisa monitorar manualmente cada rollout
- Incidentes noturnos com feature recém-lançada não têm proteção automática
- O argumento de "mais confiança para fazer deploy" cai

**Implementação mínima viável:**
1. Na criação/edição de rollout, campo "Métrica de guarda" (error_rate, latência_p99, evento customizado)
2. Worker que coleta a métrica via Prometheus/API interna a cada 60s durante o rollout
3. Se threshold excedido: rollback automático + notificação (webhook/email)
4. Tela de histórico de rollbacks automáticos

**Critério:** Rollout com guarda configurada que ao atingir threshold reverte automaticamente e notifica.

---

## 4. Diferenciais do Pulse (Onde Ganhamos)

### Self-hosted total
O LD oferece apenas um Relay Proxy — os dados ainda vão para a nuvem deles. O Pulse pode rodar 100% on-prem:
- Nenhum dado de flag evaluation ou sessão de usuário sai do ambiente do cliente
- Argumento LGPD direto para PropTech (CPF, dados financeiros, localização)
- Para a Loft: dados de usuários de imóveis têm sensibilidade regulatória real

### Pricing sem punição por volume
LD cobra por MAU. A 2M+ MAU no plano enterprise:
- Estimativa conservadora: $8-15 / 1000 MAU → $16k-30k+/mês
- Custo de infra Pulse self-hosted equivalente: R$2k-8k/mês (3-5 nós K8s)
- **A diferença de custo paga o desenvolvimento dos P0s em semanas**

### Correlação Replay + Flag Variant (único no mercado)
O LD não conecta session replay com qual variante o usuário estava vendo. No Pulse:
- Um usuário reporta bug → você abre o replay e vê exatamente qual flag estava ativa
- Debug de "qual variante causou o problema" vai de horas para minutos
- Nenhum concorrente direto tem isso integrado nativamente

---

## 5. Pitch para o Cliente Self-Hosted

### Abertura (o problema real)
> "Você paga por MAU numa plataforma onde não controla onde seus dados de usuário ficam. Cada flag evaluation envia um evento para os servidores do LD. A 2M MAU, isso é um invoice grande e uma superfície de risco de dados desnecessária — especialmente para uma PropTech com dados financeiros de usuários."

### Proposta de valor central
> "Pulse roda inteiramente na sua infraestrutura. Nenhum dado de flag evaluation, nenhum dado de sessão de usuário sai do seu ambiente. Você tem acesso total ao banco, pode auditar qualquer query, e o custo de operar é o custo da sua infra — não uma mensalidade variável por volume."

### Honestidade sobre gaps
> "Não temos paridade total com o LD hoje — e seria desonesto dizer que temos. Os dois gaps que importam para vocês são BigQuery export e Guarded Rollouts. BigQuery export está no roadmap imediato (BKL-1150) e podemos co-desenvolver se houver interesse. Guarded Rollouts também (BKL-1151). Tudo que é core — flags, segments, environments, A/B testing, approval workflow, session replay — está operacional."

### Diferencial de correlação
> "Uma coisa que o LD não faz bem: correlacionar o replay de sessão do usuário com a variante de flag que ele estava vendo. No Pulse, quando um usuário tem uma experiência ruim, você abre o replay e vê exatamente qual variante estava ativa. Isso muda o speed de debug de horas para minutos."

### Estratégia de POC recomendada
- **Escopo:** 1 projeto piloto menor, 30 dias
- **3-5 flags reais** do menor projeto da Loft (evitar migração de core)
- **Objetivo:** validar performance, approval workflow e — se BKL-1150 estiver pronto — o pipeline BigQuery
- **Sem migração total:** POC é prova de conceito, não compromisso
- **Critério de sucesso do POC:** latência p99 < 100ms, approval workflow funcional, dados chegando no BigQuery

### Argumento de custo direto
```
LD Enterprise @ 2M MAU:     ~R$80k-150k/ano (estimativa)
Pulse self-hosted @ 2M MAU: ~R$24k-96k/ano (infra + licença)
Diferença:                   R$56k-54k/ano → ROI positivo em < 6 meses
```

---

## 6. Roadmap Priorizado

### Fase 1 — Próximas 4-6 semanas (desbloquear o POC)

| BKL | Item | Esforço | Impacto |
|-----|------|---------|---------|
| BKL-1150 | BigQuery Export via streaming | XL | P0 — crítico |
| BKL-1151 | Guarded Rollouts v1 | XL | P0 — crítico |
| BKL-1152 | Webhook outbound (Datadog, Slack) | M | P1 — alto |
| BKL-1155 | Benchmark k6 documentado | S | P1 — alto |
| BKL-1153 | Segments com regras compostas | L | P1 — alto |

### Fase 2 — Q2/2026

| BKL | Item | Esforço |
|-----|------|---------|
| BKL-1154 | Multi-armed bandits (epsilon-greedy) | XL |
| — | Live Events stream (SSE/WebSocket) | M |
| — | Error tracking correlacionado com flags | L |
| — | SDK React Native / Flutter | M |

### Fase 3 — Q3-Q4/2026

| Item | Esforço |
|------|---------|
| AI Configs (prompts como flags) | XL |
| Relay Proxy para deploy em edge | XL |
| Marketplace de integrações | XL |
| SAML/SSO enterprise | M |

---

## 7. Materiais de Referência

- **LP pública:** pulse.institutoitinerante.com.br/landing
- **Doc self-hosted:** pulse-service/docs/self-hosted.md
- **Análise financeira profunda:** pulse-service/docs/ANALISE_FINANCEIRA_PROFUNDA_PULSE.md
- **Análise competitiva ampla (PostHog, GrowthBook):** pulse-service/docs/ANALISE-COMPETITIVA-PULSE-2026.md
- **Backlog completo:** instituto-itinerante-ecosystem/BACKLOG.md (BKL-1150 a BKL-1155)
