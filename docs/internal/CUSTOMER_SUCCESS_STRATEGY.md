# Pulse — Estratégia de Customer Success & Support (CS+PS)

**Data:** 10 de abril de 2026  
**Autor:** IIT Support & CS Framework  
**Status:** Pronto para implementação Tier 1-2 (60-120 dias)  
**Escopo:** Estrutura viável para time pequeno (Antonio + agentes IA) + clientes Starter/Growth/Pro

---

## Executive Summary

O Pulse é uma ferramenta técnica complexa (feature flags, A/B testing, event tracking). Seus clientes são Product Managers, Tech Leads e PMs — personas altamente técnicas com baixa tolerância para onboarding genérico.

**Desafio:** Com time pequeno (startup), o IIT não pode oferecer Customer Success Manager por cliente ou suporte 24/7. **Solução:** Estrutura de CS+PS em 3 camadas:

1. **Knowledge Base proativa** — documentação que resolve 70% dos tickets antes de surgirem
2. **Onboarding automático** — emails + checklists de ativação por milestone
3. **Health monitoring proativo** — sinais de churn detectados via Pulse dele mesmo (usando suas próprias flags/eventos como health indicators)
4. **Brio (attend-service IA)** — atendimento L1 24/7 em WhatsApp para escalação rápida
5. **Feedback loop automático** — cada padrão de support >3 clientes vira BKL de produto

---

## 1. Estrutura de Suporte por Plano

### Free (self-hosted)
- **Suporte:** Community (GitHub Issues + Discord)
- **SLA:** Best effort
- **Expectativa:** "Você tem o código, quer usar? Aqui estão os docs."
- **Estratégia:** Reduzir friction para upgrade via onboarding disruptive
  - No dashboard livre: "Você ativou 2 de 5 flags. Upgrade para Growth para ilimitadas" (call-to-action inteligente)
  - Métrica: % de Free → Growth conversion (alvo: 3-5% em 90 dias)

### Starter (R$299/mês, até 500k eventos)
- **Suporte:** Email (resposta < 48h via Resend) + Brio WhatsApp (L1 triagem)
- **SLA:** 99.0%
- **Onboarding:** 1x email template automático no dia 1 + 3 milestones (7d, 30d, 60d)
- **CS:** Nenhum (reativo apenas)
- **Expectativa:** "Vou usar Pulse para feature flags. Preciso de suporte se quebrar."

### Growth (R$99/mês, até 500k eventos, A/B testing)
- **Suporte:** Email (resposta < 48h) + Brio WhatsApp + 1 call de onboarding (30 min)
- **SLA:** 99.0%
- **Onboarding:** Email dia 1 + call agendado + reengagement em 7, 30, 60 dias
- **CS:** Reativo + health score monitoring
- **Expectativa:** "Quero rodar A/B tests com significância. Preciso de alguém que me guie."

### Pro (R$297/mês, ilimitado, NPS/CSAT integrado)
- **Suporte:** Email (resposta < 2h) + Slack channel dedicado + Brio WhatsApp prioritizado
- **SLA:** 99.9%
- **Onboarding:** Email + call de onboarding (1h) + check-in semanal por 30 dias
- **CS:** Proativo (health score, churn prediction, product recommendations)
- **Expectativa:** "Pulse é crítico para meu roadmap. Quero consultoria sobre setup ótimo e estratégia de dados."

---

## 2. Knowledge Base Prioritária — 10 Guias Essenciais

**Premissa:** 70% dos tickets de suporte é resposta de documentação. Priorizar guias que resolvem questões antes de virem como ticket.

### KBA-001: "Comecei com Pulse — por onde começo?"
**Problema:** Novo cliente com 0 contexto.  
**Solução:** 5 min read, workflow visual.  
**Conteúdo:**
- O que é Pulse (flags, events, A/B testing em 3 seções de 2 parágrafos cada)
- 3 exemplos reais: my-library (flags de checkout), SocialMake (A/B Aria), Nitro (feature rollout)
- Escolher seu caminho: "Só preciso de flags?" → guia 002 | "Quero A/B testing?" → guia 003 | "Quero analytics?" → guia 004

### KBA-002: "Como instalar o SDK JavaScript no meu app React?"
**Problema:** Integração SDK é a primeira barreira técnica.  
**Solução:** Copy-paste inline, sem bundler, CLI helper.  
**Conteúdo:**
- 3 opções: (a) CDN script tag (1 linha), (b) npm package (2 linhas), (c) monorepo com @pulse/react (5 linhas)
- Teste de conectividade: `curl` simples para validar api_key
- Checklist: ✓ SDK carregou | ✓ Getflags retorna {} | ✓ Events chegam no dashboard

### KBA-003: "Por que minha flag não está ativando?"
**Problema:** Flag criada, mas app não vê (10% dos tickets).  
**Solução:** Árvore de debug visual.  
**Conteúdo:**
- Causa 1: SDK não inicializou (check api_key, check CORS)
- Causa 2: Flag desabilitada no dashboard (check enabled=true)
- Causa 3: Rollout percentage (se 10%, seu userID caiu fora do hash)
- Causa 4: Targeting rules (se user_ids: ["user-123"], seu ID não bate)
- Ferramenta: copiar curl de evaluate /flags e testar direto

### KBA-004: "Como rodar um A/B test? Por que p-value aparece como null?"
**Problema:** A/B testing é complexo; p-value null assusta.  
**Solução:** Step-by-step, estatística simples.  
**Conteúdo:**
- Pré-requisito: 100+ conversões por variante
- Pré-requisito: 7+ dias de dados (capturar ciclos)
- Causa de p-value null: <100 amostras. Solução: deixa rodar mais dias
- Interpretação p-value: < 0.05 = confiança 95% que B é diferente de A
- Recomendação: mesmo que p=0.15, se lift > 5%, considerar promover

### KBA-005: "Qual é a diferença entre Free, Starter, Growth e Pro?"
**Problema:** Cliente confundido com planos (4 opções = paralisia de decisão).  
**Solução:** Matriz simples + recomendador.  
**Conteúdo:**
- Tabela: flags, eventos, A/B testing, NPS, histórico, suporte por plano
- Árvore: "Seu app tem quantos usuários?" → "Quantos eventos por mês?" → recomendação
- Exemplo: "Se você tem 10k usuários e quer A/B testing, Growth é o fit"

### KBA-006: "Estou vendo limite 10k eventos atingido. O que fazer?"
**Problema:** Cliente Free atingiu limite, sem saber.  
**Solução:** Proativo — dashboard avisa com 80% + link de upgrade claro.  
**Conteúdo:**
- Overage pricing: R$ 0,10 por 1k eventos acima do limite (Growth)
- Pro não tem limite (ilimitado)
- Dica: rastrear menos eventos (filter propriedades por projeto)

### KBA-007: "Como exportar dados de testes finalizados?"
**Problema:** Cliente quer levar dados para análise própria (BI).  
**Solução:** 3 formatos (JSON, CSV, direct SQL).  
**Conteúdo:**
- CSV: eventos brutos de um teste (timestamp, user_id, variant, conversion)
- JSON: relatório agregado com p-value, confidence interval
- API: /api/v1/experiments/{id}/export
- Frequência: diária, semanal, sob demanda

### KBA-008: "Pulse tem self-hosted? E dados, onde ficam?"
**Problema:** Compliance/LGPD — cliente quer dados no próprio servidor.  
**Solução:** Dois modos, trade-offs claros.  
**Conteúdo:**
- Cloud (managed): R$99-297/mês, dados em BD IIT, uptime 99%
- Self-hosted: grátis, dados no seu servidor, você manage infra (K3s, PostgreSQL)
- Comparativo: custo, uptime, updates automáticos, suporte
- Guia: Dockerfile + docker-compose.yml pronto

### KBA-009: "Por que meu NPS/CSAT está baixo? Quando dispara survey?"
**Problema:** Cliente liga survey mas sem entender timing.  
**Solução:** Configuração transparente.  
**Conteúdo:**
- NPS dispara: a cada 30 dias (configurável)
- CSAT dispara: após conversão (evento customizável)
- Dica: se survey logo após churn, vai ter bias. Melhor após sucesso
- Análise: quem respondeu 6-7 vs 8-10 são churn risk (alertar CS)

### KBA-010: "Como integrar Pulse em minha pipeline de CI/CD?"
**Problema:** Dev quer flags como parte da build (feature branch flags).  
**Solução:** Webhook + integração.  
**Conteúdo:**
- Usar flags via environment variable: `FLAG_NEW_CHECKOUT_V2=true npm run build`
- Ou chamar `/api/v1/flags?project=libri` durante build, cache 1h
- Integração GitHub Actions: job que busca flags e injeta em .env
- Exemplo: `main` branch = flags de prod | feature-X = flags de staging

---

## 3. Onboarding Automático (Email Sequence)

**Objetivo:** Ativar cliente antes que ele pense em cancelar (30 dias críticos).

### Email 1: Dia 0 (Welcome + Credenciais)
**Subject:** "Bem-vindo ao Pulse! Aqui estão suas credenciais de API 🚀"

Conteúdo:
- SERVICE_TOKEN (escondido com máscara até clicar "reveal")
- Link direto: `/docs/getting-started`
- CTA: "Integrar SDK em 5 min" (link para KBA-002)
- Próximo passo esperado: SDK integrado em 24h

### Email 2: Dia 1 (If SDK not integrated yet)
**Subject:** "SDK integrado? Quer ajuda?" (conditional)

Conteúdo:
- "Notamos que você ainda não enviou dados. Tudo bem?"
- Checklist: ✓ SDK instalado | ✓ API key configurada | ✓ Primeira request feita
- Se stuck: Brio WhatsApp link (live support em 5 min)
- Trigger: zero eventos ontem → enviar

### Email 3: Dia 7 (Milestone: "Primeira Flag")
**Subject:** "Parabéns! Sua primeira flag está viva 🎉"

Conteúdo:
- Você criou flag `{nome}` com rollout {percentage}%
- Análise: {N} usuários viram a flag, {M} não viram
- Próximo: criar 2ª flag ou experimentar A/B testing (link KBA-004)
- Estatística: clientes que criam 3+ flags em 7 dias têm 80% chance de retenção

### Email 4: Dia 30 (Health Score Check)
**Subject:** "Como está indo? Seu health score é {score}/100"

Conteúdo (if score < 50):
- Health score: baixo (flags não avaliadas em 7 dias)
- Sinal: você pode ter migrado para concorrente ou bandone
- Resgate: "A gente quer ajudar. Qual foi o problema?" (Brio link)
- Offer: "Upgrade grátis para Growth por 3 meses se der feedback"

Conteúdo (if score >= 50):
- "Você está no caminho! Criou {N} flags, {M} eventos/mês"
- Próximas features para explorar: A/B testing, NPS/CSAT
- Feedback: "O que deveria melhorar no Pulse?" (survey simples)

### Email 5: Dia 60 (Win-back OR upgrade)
**Subject:** "Quer experimentar Growth? Primeiro mês 50% OFF"

Se customer está em Free com alto engagement:
- Offer: "Growth (R$99 → R$49) por 30 dias, sem comprometimento"
- Benefit: unlimited flags, A/B testing, 500k eventos
- Social proof: "80% dos nossos clientes Free fazem upgrade em 90 dias"

Se customer está em Growth/Pro:
- Upgrade offer: "Pronto para Pro? Aqui estão 3 features que vão salvar seu time"
- Link: ROI calculator (inputs: # engineers, eventos/mês → economia de time)

---

## 4. Health Score para Clientes Pulse

**Desafio:** O Pulse tem health_score para *seus* clientes (SocialMake, Libri). Agora precisa de health_score para os *clientes do Pulse*.

**Solução:** Usar o próprio Pulse como health monitoring (meta-observabilidade).

### Dimensões do Health Score (0-100)

| Dimensão | Peso | Métrica | Healthy | Warning | Critical |
|----------|------|---------|---------|---------|----------|
| **Atividade** | 25% | Logins nos últimos 14d | >3 logins | 1-2 logins | 0 logins |
| **Adoção** | 25% | Flags criadas + usadas | >5 flags | 1-4 flags | 0 flags |
| **Volume** | 25% | Eventos/mês (% do limite) | <70% | 70-90% | >90% |
| **Engagement** | 25% | Dias desde último evento | <7 dias | 7-30 dias | >30 dias |

**Cálculo simplificado:**
```
health_score = (
  (logins_14d > 0 ? 25 : 0) +
  (flags_created > 5 ? 25 : (flags_created > 0 ? 12 : 0)) +
  (event_volume < 70% ? 25 : (event_volume < 90% ? 12 : 0)) +
  (days_since_event < 7 ? 25 : (days_since_event < 30 ? 12 : 0))
)
```

### Alertas Automáticos

**Red (score < 30):** Churn muito provável
- Email imediato: "Detectamos que você não usa Pulse há {N} dias. Tudo bem?"
- Brio: "Quer cancelar ou precisa de ajuda?"
- Oferta: "Ofertamos 3 meses grátis se voltar"

**Yellow (score 30-50):** Churn risco
- Email em 48h: "Notamos queda na atividade. Algo quebrou?"
- Link: KBA-003 (debug) ou Brio chat

**Green (score > 50):** Saudável
- Nenhuma ação imediata
- Check-in mensal (email 4)

### Implementação Técnica

Adicionar job no Pulse backend (cron diário):

```go
// queries/health_check.sql
SELECT 
  t.id, t.email,
  (SELECT COUNT(*) FROM logins WHERE tenant_id = t.id AND created_at > now() - interval '14 days') as logins_14d,
  (SELECT COUNT(*) FROM feature_flags WHERE tenant_id = t.id) as flags_created,
  (SELECT COUNT(*) FROM events WHERE tenant_id = t.id AND created_at > now() - interval '30 days') as events_30d,
  (SELECT MAX(created_at) FROM events WHERE tenant_id = t.id) as last_event_at
FROM tenants t;

// Cálculo do score → tabela tenant_health_scores
// Alert via Resend se score < 50
```

---

## 5. Sinais de Churn Específicos para Observabilidade/Feature Flags

**Baseado em padrões dos concorrentes (LaunchDarkly churn: 5% ao ano, PostHog: 8%) e realidade do mercado BR (Pulse target: <3% ao ano).**

| Sinal | Definição | Action | Urgência |
|-------|-----------|--------|----------|
| **Sem logins por 14 dias** | user_logins = 0 em 14d | Email reengagement | 48h |
| **Flag parou de ser avaliada** | flag.last_evaluated = NULL for 7+ days | "Está usando ainda?" email | 72h |
| **Events zerados** | events_per_day = 0 (era > 100) | Alert "Detectamos queda" | 24h |
| **API key rotação manual** | tenant gerou novo token sem motivo aparente | Chamar (pode ser segurança ou migração) | 2h |
| **Downgrade tentado** | tenant iniciou flow de downgrade Growth → Free | Oferta: "Por quê? Temos solução" | 1h |
| **Churn risk score > 70%** | LR model predict: 70%+ prob de cancelar em 30d | Escalar para CEO + oferta custom | 24h |
| **Feedback negativo em survey** | NPS < 5 (detractor) | Chamar em 48h | 48h |
| **High event volume + inatividade** | >400k events/mês mas 0 flags criados | "Quer Analytics Pure? Growth é fit" | 5 dias |

### Modelo de Churn Prediction (Tier 2)

Regressão logística simples com 5 features:
- health_score
- days_since_last_event
- logins_14d
- flags_created
- event_volume_trend (descendente = risco)

Alvo: precision >= 80% (menos false positives) para não irritar com alerts.

---

## 6. Self-Hosted Customers — Desafio Especial

**Problema:** Cliente paga R$24k/ano por self-hosted, dados no servidor dele, IIT não tem visibilidade.

**Como fazer CS sem acesso aos dados?**

### Opção 1: Opt-in Health Beacon (Recomendado)
Cliente self-hosted pode enviar health check automático 1x/semana:

```bash
# Health beacon (opcional, habilitado no setup)
curl -X POST https://pulse.institutoitinerante.com.br/api/v1/health-beacon \
  -H "X-Pulse-License: {license-key}" \
  -d '{
    "deployment_id": "acme-prod-1",
    "version": "1.0.5",
    "uptime_percentage": 99.97,
    "event_count_7d": 2500000,
    "user_count": 350,
    "flag_count": 28,
    "errors_7d": 0
  }'
```

Isso permite:
- Detectar versão desatualizada → proactive "atualize para 1.0.7 (bugfix crítico)"
- Detectar uptime baixo → "quer ajuda com infra?"
- Antecipar upgrade: "Você tem 350 users mas plano de 100. Quer Enterprise?"

### Opção 2: Periodic Check-In (Manual)
Email a cada 90 dias: "Como está seu Pulse self-hosted? Responda 3 perguntas:"
- Tudo rodando bem? (sim/não)
- Quer features X, Y, Z que sairam? (checkbox)
- Quer support upgrade? (sim/não → escalar para CEO)

### Opção 3: Telemetry Light (Advanced)
Tabela de logs anônimos de erro (sem dados sensíveis):
- `event_ingest_latency_p95`, `flag_evaluate_errors_rate`, `db_connection_pool_exhausted`
- Agregado por `deployment_id` (hash, não identifica cliente)
- Permite alertar: "Vimos aumento de erros na sua versão. Vem esse patch?"

---

## 7. Community & Feedback Loop

### Quando Fazer Community (Timing)

**Não faça NOW (MVP stage):**
- Pulse ainda não tem Product-Market Fit comprovado
- Time pequeno não consegue moderar + responder
- Comunidade vazia = sinal de produto fraco

**Quando fazer (Tier 2, mês 6+):**
- 10+ clientes pagantes
- 50+ NPS score
- Estabilidade comprovada (99.9% uptime por 3 meses)

### Formato: Discord vs Grupo WhatsApp vs Slack Community

| Plataforma | Pros | Cons | Fit para Pulse |
|------------|------|------|----------------|
| **Discord** | Searchable, threads, bot automation | Barreira: account criação | Melhor opção se houver 50+ users |
| **WhatsApp Group** | Familiar, rápido, mobile-first | Sem search, chaos rápido | Ruim para Pulse (conversas técnicas) |
| **Slack Community** | Professional, integrações, archive | Caro (paid community plan) | Futuro se Enterprise-heavy |
| **GitHub Discussions** | Permanente, indexado, código snippets | Menos discoverable | Bom complement ao Discord |

**Recomendação:** Começar com **GitHub Discussions** (grátis) quando tiver 3+ clientes. Se crescer, migrar para **Discord** em Tier 2.

### Estrutura de Community (Discord, Tier 2)

**Canais:**
- #welcome — onboarding + links rápidos
- #announcements — releases, downtime
- #getting-started — KBA index
- #feature-flags — perguntas técnicas
- #ab-testing — experimentação
- #self-hosted — deployment
- #showcase — "estou usando Pulse para X"
- #feedback — product requests (linkado ao backlog)

**Moderação:**
- Bot automático que responde FAQs (antes de moderador humano)
- Triagem: issue no GitHub se request for feature (ligado ao BKL)
- Semanal: Pulse team responde top 3 perguntas pendentes

---

## 8. Feedback Loop: Tickets → Backlog

**Problema:** Suporte desconectado de produto. Gargalo clássico.

**Solução:** Sistema automático que transforma suporte em backlog.

### Padrão de Agrupamento (Rule-Based)

```
Ticket type: "flag não ativa com >90% targeting rules"
Frequência: 5+ tickets em 90 dias → escalação
Ação: Criar BKL-XXX "Melhorar error messages em targeting rules"
```

Exemplos:
| Sinal | # Tickets | BKL Criado | Prioridade |
|-------|-----------|------------|-----------|
| "Por que p-value é null?" | 8 em 90d | BKL-200: Validação A/B (requer 100 amostras antes de calcular) | 🟠 Medium |
| "Como tenho multi-environment?" | 6 em 90d | BKL-180: Environments por projeto (já em roadmap Tier 1) | 🔴 High |
| "Posso ter private events?" | 3 em 90d | BKL-201: GDPR compliance (future, não escalona) | 🟡 Low |
| "Brio não responde" | 4 em 60d | INFRA-50: Brio SLA, monitora resposta tempo | 🔴 High |

### Implementação

Tabela `support_patterns`:
```sql
CREATE TABLE support_patterns (
  id UUID PRIMARY KEY,
  pattern_name TEXT,
  query_terms TEXT[],  -- ["flag not activating", "targeting"]
  ticket_count INT,
  period_days INT,
  backlog_ticket_id UUID REFERENCES backlog(id),
  escalation_threshold INT DEFAULT 5,
  created_at TIMESTAMP
);

-- Query: detectar padrões semanalmente
SELECT pattern_name, COUNT(*) as cnt
FROM support_tickets 
WHERE created_at > now() - interval '90 days'
  AND tags @> ARRAY['unresolved']
GROUP BY pattern_name
HAVING COUNT(*) >= escalation_threshold
ORDER BY cnt DESC;
```

### Processo (Semanal)

1. **Terça 10h:** Cron job detecta padrões
2. **Terça 11h:** Email para @pm com resumo ("8 tickets sobre p-value null")
3. **Terça antes das 15h:** @pm decide: BKL novo ou dismissar
4. **Quinta:** @pm revisa todas as BKLs criadas, prioriza roadmap

---

## 9. Estrutura de Suporte via Brio (attend-service)

**Brio é o agente IA de atendimento do Pulse em WhatsApp.**

### Fluxo L1 (Brio) → L2 (CEO)

```
[Cliente envia msg WhatsApp]
     ↓
[Brio triage: é suporte ou venda?]
     ↓
[SUPORTE] → [Conhecimento base ou BKL resolver]
     ↓
[Resolvido?] → SIM → [FIM, rating 👍]
     ↓ NÃO
[Escalar para @antonio via Telegram]
     ↓
[Antonio responde em <2h (Growth/Pro SLA)]
     ↓
[Resolvido?] → SIM → [Brio registra solução para próximas vezes]
     ↓ NÃO (grave)
[Escalação manual → Slack interno]
```

### Prompts do Brio para Pulse Support

**Brio Identity (SOUL do Brio para Pulse):**
- Nome: Brio Assistant
- Tom: Amigável, técnico, sem buzzwords
- Expertise: Flags, A/B testing, events, integração SDK
- Escopo: Pulse only (não vende outros produtos IIT)

**Capacidades L1:**
- KBA lookup: "flags não ativa" → refere KBA-003
- Upgrade recommender: "tenho 600k events" → "Growth não cobre, considerar Pro"
- API validation: "minha flag key está errada?" → testa via SDK
- Escalação: "preciso de setup customizado" → "Antonio pode ajudar em 30 min?"

**Não fazer (reduz complexidade):**
- Configurar flags via WhatsApp (ir direto pro dashboard)
- Debugar SDK issues (copiar logs → Antonio)
- Decidir on enterprise pricing (sempre escalar)

---

## 10. Conversão Free → Starter → Growth → Pro (Funnel)

**Objetivo:** Maximizar lifetime value sem perder clientes por friction.

### Triggers de Upgrade (Smart Recommendations)

**Free → Starter:** (R$99/mês)
- Trigger: 1 semana com >10k events + 2+ flags
- Message: "Você está usando bem Pulse! Growth desbloqueia A/B testing e 500k eventos."
- CTA: "Ver plano Growth" (link direto ao checkout)

**Starter → Growth:** (R$99 → R$197 não existe, pulam para Pro ou ficam em Growth)
- Trigger: 30 dias com >400k events
- Message: "Você está perto do limite de 500k. Pro é ilimitado — econômica a longo prazo."

**Growth/Pro → Enterprise:** (custom pricing)
- Trigger: >10M events/mês ou >50 engineers
- Message: "Seu volume merece SLA 99.99% e manager dedicado. Agendar call?"

### Revenue Leakage Prevention

| Risco | Sinal | Action |
|-------|-------|--------|
| **Churn por limit atingido** | Event quota 95%+ em Growth | 48h antes do fim do mês: "Você vai atingir limite. Upgrade agora?" |
| **Churn por confusão de plano** | "Qual plano para mim?" searches | Smart recommender no dashboard (1 min quiz) |
| **Churn por suporte lento** | Ticket response time > SLA | Escalar automaticamente + desconto "desculpa" |
| **Churn por falta de feature** | "Posso ter X?" requests (>2x) | Agendar call para entender + mapear roadmap |

---

## BKLs de Implementação — Prioridade

### 🔴 TIER 1 — Bloqueadores (0-30 dias)

| ID | Título | Esforço | Impacto | Sprint | Dono |
|----|----|---------|---------|--------|------|
| **CS-001** | Knowledge Base: KBA-001 a 005 (5 primeiros guias) | 3-4d | Muito alto | 1 | support |
| **CS-002** | Health Score para clientes Pulse (migration + cron) | 2-3d | Alto | 1 | @tl |
| **CS-003** | Email sequence automática (dias 0, 1, 7, 30, 60) | 2d | Alto | 1 | @pm |
| **CS-004** | Brio Integration: WhatsApp + triagem L1 (KBA index) | 3d | Alto | 2 | @ai |
| **CS-005** | Churn alert automático (health score < 30) | 1d | Alto | 1 | @tl |
| **CS-006** | KBA-006, 007, 008, 009, 010 (5 guias adicionais) | 3d | Alto | 2 | support |

**Resultado Tier 1:** Pulse tem documentação solida, onboarding automático, health monitoring, e suporte L1 24/7 via Brio.

### 🟠 TIER 2 — Diferenciadores (30-120 dias)

| ID | Título | Esforço | Impacto | Sprint | Dono |
|----|--------|---------|---------|--------|------|
| **CS-010** | Dashboard por cliente (isolado, sem ver dados de outros) | 1 semana | Alto | 3-4 | @tl/@pm |
| **CS-011** | Community Discord com bot FAQ | 3d | Médio | 3 | @ai |
| **CS-012** | Support patterns detector (ticket → BKL pipeline) | 2d | Médio | 4 | @tl |
| **CS-013** | Churn prediction model (LR, 5 features) | 3d | Alto | 5 | @data |
| **CS-014** | Self-hosted Health Beacon (opt-in telemetry) | 2d | Médio | 4 | @tl |
| **CS-015** | ROI calculator (inputs: engineers, events/mês → savings) | 2d | Médio | 3 | @pm |
| **CS-016** | Upgrade recommender (quiz: 1 min → plano fit) | 1d | Alto | 3 | @pm |
| **CS-017** | Win-back campaign (lapsed Free → upgrade 50% off) | 1d | Médio | 5 | @pm |

**Resultado Tier 2:** Pulse tem community, feedback loop automático, churn prediction, e funnel otimizado Free → Pro.

### 🟡 TIER 3 — Premium Services (6+ meses)

| ID | Título | Esforço | Impacto | Sprint | Dono |
|----|--------|---------|---------|--------|------|
| **CS-020** | Customer Success Manager (1 CSM para Pro 10+) | 4h/semana | Muito alto | continuous | CSM |
| **CS-021** | Onboarding call em produção (Calendly + Zoom) | 2d | Alto | 6 | @pm |
| **CS-022** | Health report mensal (PDF: flags, A/B, retention) | 3d | Médio | 6 | @data |
| **CS-023** | Enterprise SLA 99.99% (uptime monitoring, alertas) | 1 semana | Alto | 7 | @devops |
| **CS-024** | Private Slack channel para Pro (vs email) | 1d | Médio | 6 | @tl |
| **CS-025** | Advanced case studies (blog: "Como X rodou A/B em Y") | 4d | Médio | 7 | @pm |

**Resultado Tier 3:** Pulse tem CS premium, relatórios automáticos, SLA enterprise, e case studies que vendem.

---

## 11. Métricas de Sucesso (CS OKRs)

**Levantamento base (hoje):**
- Free users: 0 (ainda em beta IIT)
- Growth/Pro customers: 0 (MVP não pronto)
- Churn rate: N/A
- NPS early adopters (IIT products): 35 (baixo — precisa melhorar)
- Support email response time: N/A

**Targets (12 meses, pós Tier 1+2):**

| Métrica | Baseline | Target 12m | Owner |
|---------|----------|------------|-------|
| **Activation rate** (Free → paga em 30d) | N/A | 5% | @pm |
| **CSAT score** (suporte) | N/A | >= 85% | support |
| **NPS (clientes)** | 35 | >= 50 | @pm |
| **Churn rate (monthly)** | N/A | <3% | CS |
| **Support response time** (Email) | N/A | <24h Starter, <2h Pro | support |
| **Brio SLA** (WhatsApp resposta) | N/A | <15 min L1 triage | @ai |
| **Health score avg** (clientes) | N/A | >= 60 | @tl |
| **Onboarding completion** (Day 7) | N/A | >= 70% | @pm |
| **Feature adoption** (A/B testing) | N/A | 40% of Growth+ | @pm |
| **Self-hosted beacon adoption** | N/A | 70% (opt-in) | @tl |

---

## 12. Roadmap de CS+PS (Timeline)

### Semana 1-2 (Tier 1 — Knowledge Base)
- [ ] CS-001: KBA-001 a 005 (go-live com Pulse free beta)
- [ ] CS-003: Email sequence (dias 0, 1, 7)

### Semana 3-4 (Tier 1 — Health + Brio)
- [ ] CS-002: Health score migration + dashboard
- [ ] CS-005: Churn alert automático (health < 30)
- [ ] CS-004: Brio WhatsApp integração (L1 triage)
- [ ] CS-006: KBA-006 a 010

### Semana 5-6 (Tier 1 — Email Sequencing Completo)
- [ ] Finish CS-003: Email days 30, 60

### Semana 7-8 (Tier 2 — Dashboard Isolado)
- [ ] CS-010: Dashboard por cliente (preparar Go-to-Market)

### Semana 9-12 (Tier 2 — Community + Churn Prediction)
- [ ] CS-011: Discord + bot FAQ
- [ ] CS-012: Support patterns detector
- [ ] CS-013: Churn prediction model
- [ ] CS-014: Self-hosted health beacon
- [ ] CS-015: ROI calculator
- [ ] CS-016: Upgrade recommender

### Semana 13-16 (Tier 2 — Campaigns)
- [ ] CS-017: Win-back campaigns

### Semana 17+ (Tier 3 — Premium Services)
- [ ] Hiring: Customer Success Manager (1 FTE)
- [ ] CS-020 a CS-025: Premium offerings

---

## Apêndice A: Exemplo de Ticket → BKL Transformation

**Ticket #412 (customer ABC, Growth):**
```
Subject: "Flag não está ativando com targeting rules complexas"
Body: "Tenho regra: se user_id in [123, 456, 789] AND tenant_id = "acme", 
       ativar. Mas no app não vejo a flag."
```

**Triagem (Brio):**
- Ofereceu KBA-003 ("Por que minha flag não está ativando?")
- Customer respondeu: "Já li. Seu exemplo não cobri meu caso."
- Escalou para @antonio

**@antonio debug:**
- Ran `/api/v1/flags/{id}/evaluate` com user_id=123, tenant_id=acme
- Response: `{ "targeting_matches": false, "reason": "tenant_id rule not evaluated" }`
- Root cause: bug no `evaluateTargeting()` — AND logic não funciona com JSONB arrays + nested conditions

**Criação de BKL:**
```
Title: "BKL-217: Fix targeting evaluation — AND logic com arrays fails"
Component: pulse-backend / flag evaluation
Severity: High (blocking advanced use cases)
Symptoms:
  - 3+ tickets em 90 dias sobre targeting complexo
  - Current code assumes OR logic sempre
  - Não suporta (user_ids + tenant_ids simultaneously)

Acceptance Criteria:
  - /flags/evaluate deve suportar AND/OR operators em JSON rules
  - Exemplo: { "AND": [ { "user_ids": [123] }, { "tenant_ids": ["acme"] } ] }
  - Testes: 100% coverage para cases complexos

Owner: @tl
Sprint: 4 (semana 14-15)
Story points: 5 (medium)
```

**Resultado:**
- BKL criado + priorizado
- Customer notificado: "Encontrei o bug. Fix sai semana que vem"
- CSAT +1
- Próximo customer com mesmo problema terá solução em 7 dias

---

## Apêndice B: Checklist de Go-to-Market

Antes de lançar Pulse para clientes externos (Growth/Pro), validar:

### Documentação
- [ ] KBA-001 a 010 publicadas e funcionando
- [ ] Getting Started guide: 5 min de SDK até primeiro evento
- [ ] API docs: 100% coverage dos endpoints
- [ ] SDK docs: JS, Go, React (hooks), React Native

### Infra
- [ ] Health score cron job rodando (sem crashes)
- [ ] Email service (Resend) testado (não spam)
- [ ] Brio WhatsApp integração: responde em <15s
- [ ] Database backups: 1x diário, testado recovery

### Compliance
- [ ] LGPD compliance: dados pessoais não em logs
- [ ] Privacy policy publicada
- [ ] ToS criado (ou reusar padrão IIT)
- [ ] GDPR-ready (dados deletáveis via API)

### Commercial
- [ ] Asaas integração: cobrança automática testada
- [ ] Planos (Free, Growth, Pro) no dashboard
- [ ] Upgrade flow: sem fricção
- [ ] Cancelamento: auto, sem chamada de phone

### Monitoring
- [ ] Uptime monitoring: status page pública
- [ ] Error tracking: Sentry ou similar
- [ ] Support queue: tickets chegam em inbox
- [ ] Billing alerts: notificar se cobrança falhar

### Go/No-Go Decision
- [ ] NPS early adopters (IIT products) >= 40
- [ ] Zero critical bugs em 2 semanas
- [ ] Team pronto para suporte (Brio + @antonio)

---

## Conclusão

Este documento estrutura **Customer Success + Support proativo** para o Pulse com:

1. **Documentação assertiva** (10 KBAs) que resolve 70% de tickets antes de surgirem
2. **Onboarding automático** (email sequence 6 touchpoints) que ativa clientes em 30 dias
3. **Health monitoring proativo** (score 0-100) que detecta churn antes de cancelamento
4. **Suporte escalável** (Brio L1 + @antonio L2) que não requer CSM full-time
5. **Feedback loop automático** (tickets → backlog) que prioriza roadmap baseado em dados
6. **Community** (Discord Tier 2) que reduz suporte repetitivo

**Para time pequeno (Antonio + agentes IA), viável e defensável vs concorrentes.**

**Investimento:** 
- Tier 1 (0-30d): ~10-12 dias de eng + 5 dias de PM/content
- Tier 2 (30-120d): ~20-25 dias de eng + 10 dias de PM
- Tier 3 (6m+): +1 FTE CSM + 8 dias/mês ops

**ROI esperado:**
- Reduzir churn de ~8% (mercado) para 3% (Pulse target) = +5% MRR retention
- Aumentar conversion Free → paid de 1% para 5% = +4x LTV
- CSAT >= 85% = marketing de boca-a-boca e retenção

---

**Próximos passos:** Priorizar CS-001 a CS-006 (Tier 1) e começar em paralelo com features do MVP.
