# Pulse — Customer Success Backlog (BKL-CS-XXX)

**Status:** Ready for implementation  
**Last Updated:** 10 de abril de 2026  
**Total BKLs:** 25 (Tier 1: 6, Tier 2: 11, Tier 3: 8)  
**Estimated Effort:** Tier 1: 16-18 days | Tier 2: 35-40 days | Tier 3: 40-50 days + 1 FTE CSM

---

## 🔴 TIER 1 — MVP Support (0-30 dias) — CRÍTICO

Bloqueadores de Go-to-Market. Sem isso, não há suporte funcional para clientes pagantes.

---

### BKL-CS-001: Knowledge Base — 5 Guias Prioritários

**Descrição:** Criar KBA-001 a KBA-005 (Getting Started, SDK JS, Flag Debug, A/B Testing, Pricing)

**Objetivo:** 70% dos tickets resolvidos por KBA lookup. Reduzir volume de suporte manual em 60%.

**Aceitação:**
- [ ] KBA-001: "Comecei com Pulse — por onde começo?" (5 min read, 3 workflows)
- [ ] KBA-002: "Como instalar SDK JavaScript?" (copy-paste, CDN + npm, React hook)
- [ ] KBA-003: "Por que minha flag não está ativando?" (árvore de debug visual com 4 causas)
- [ ] KBA-004: "Como rodar A/B test + interpretar p-value?" (step-by-step + significância)
- [ ] KBA-005: "Qual plano escolher?" (matriz simples + árvore de decisão)
- [ ] Publicadas em `/docs/guides/` e linkadas no dashboard
- [ ] SEO metadata + keywords (Google indexing)
- [ ] Video snippets (5x 2 min screencasts para cada KBA)

**Esforço:** 3-4 dias (1 pessoa support/PM)  
**Dono:** @pm + support  
**Sprint:** 1  
**Prioridade:** 🔴 BLOQUEADOR

---

### BKL-CS-002: Health Score System para Clientes Pulse

**Descrição:** Migração + cron job que calcula health score 0-100 para cada tenant usando Pulse.

**Objetivo:** Detectar churn passivo (0 logins em 14d, eventos zerados) antes de cancelamento.

**Specs:**
```
Dimensões (25% cada):
  1. Atividade: logins últimos 14d (>3 = saudável, 0 = crítico)
  2. Adoção: flags criadas (>5 = saudável, 0 = crítico)
  3. Volume: eventos/mês % do limite (<70% = saudável, >90% = crítico)
  4. Engagement: dias desde último evento (<7d = saudável, >30d = crítico)

Score = (login_pts + adoption_pts + volume_pts + engagement_pts)

Threshold:
  score < 30 = Red (churn iminente) → email imediato
  30-50 = Yellow (risco) → email em 48h
  > 50 = Green (saudável) → check-in mensal
```

**Aceitação:**
- [ ] Migration: `tenant_health_scores` table com score, updated_at
- [ ] Cron job (diário 8h): calcula score para todos os tenants
- [ ] Query: agregar logins, flags, events do último período
- [ ] Dashboard: health score por tenant (admin view)
- [ ] Teste: 10 tenants fictícios, validar scores corretos
- [ ] Logging: cada score calculado em logs (audit)

**Esforço:** 2-3 dias  
**Dono:** @tl  
**Sprint:** 1  
**Prioridade:** 🔴 BLOQUEADOR

---

### BKL-CS-003: Email Sequence — Onboarding Automático

**Descrição:** 6 emails automáticos disparados baseado em milestones/dias (dias 0, 1, 7, 30, 60 + if condition).

**Objetivo:** Ativar 70% dos novos clientes em 30 dias (criar flags + enviar eventos).

**Emails:**

| Email | Trigger | Subject | CTA |
|-------|---------|---------|-----|
| 1 | Dia 0 (signup) | Welcome! Aqui estão suas credenciais | Integrar SDK em 5 min |
| 2 | Dia 1 IF zero eventos | SDK integrado? Quer ajuda? | Brio WhatsApp link |
| 3 | Dia 7 IF events > 0 | Parabéns! Primeira flag viva 🎉 | Tentar A/B testing |
| 4 | Dia 30 | Health score check | Proativo de churn |
| 5 | Dia 60 se Growth/Pro | Quer explorar Growth? 50% OFF | Upgrade link |
| 6 | Dia 60 se já pago | Como está? Feedback survey | NPS 1-10 |

**Aceitação:**
- [ ] Tabela `email_sequences` com template, trigger, variables
- [ ] Cron job (horário 9h): detecta segmentos, dispara via Resend
- [ ] Templates (HTML) responsivos
- [ ] Variáveis dinâmicas: {{firstName}}, {{flagCount}}, {{healthScore}}, {{planName}}
- [ ] Unsubscribe link + compliance LGPD
- [ ] Open/click tracking (Resend nativo)
- [ ] A/B test subject line (Email 1): "Welcome!" vs "Let's Go! 🚀" (50/50, mede CTR)

**Esforço:** 2 dias  
**Dono:** @pm  
**Sprint:** 1  
**Prioridade:** 🔴 BLOQUEADOR

---

### BKL-CS-004: Brio Integration — WhatsApp L1 Support

**Descrição:** Integrar Brio (attend-service) em Pulse support. Brio triagens e resolve L1 via KBA lookup.

**Objetivo:** Suporte 24/7 sem custo. <15 min resposta para triage. 60% dos tickets resolvidos em L1.

**Capabilities:**
- Understands: feature flags, A/B testing, events, SDK issues, plan questions
- Resolves: KBA lookups, upgrade recommender, API validation, escalation
- Does NOT: Configure flags, debug SDK issues, enterprise pricing

**Integration:**
```
[Cliente envia: "por que minha flag não ativa?"]
     ↓
[Brio classifica: flag_not_activating]
     ↓
[Busca BKL-CS-003 config → KBA-003]
     ↓
[Oferece 4 causas + debug steps]
     ↓
[Cliente testa → resolvido OU precisa de mais ajuda]
     ↓
[Se unresolved → escalação: "Antonio pode ajudar em 30 min? Quer?"]
     ↓
[Se sim → criar ticket interno + notificar @antonio via Telegram]
```

**Aceitação:**
- [ ] Brio SOUL.md atualizado (Pulse-specific identity)
- [ ] KBA index embarcado em Brio context (5 primeiros guias, Tier 1)
- [ ] Webhook: ticket suporte → POST para Brio inbox
- [ ] Escalation: Brio envia summary + customer choice para Telegram (@antonio)
- [ ] Rating system: "Brio resolveu?" sim/não (treina modelo)
- [ ] Analytics: % de tickets resolvidos em L1 vs escalados

**Esforço:** 3 dias (2 eng + 1 PM para treino)  
**Dono:** @ai (Brio) + support  
**Sprint:** 2  
**Prioridade:** 🔴 BLOQUEADOR

---

### BKL-CS-005: Churn Alert — Automático via Health Score

**Descrição:** Quando health score < 30, dispara email + Brio alerta automático.

**Objetivo:** Reduzir churn reativo detectando sinais 7-14 dias antes.

**Fluxo:**
```
[Cron diário: calcula health scores]
  ↓ [BKL-CS-002 calcula score]
  ↓
[IF score < 30 AND antes não era < 30]
  ↓
[Cria alert]
  ↓
[Email: "Detectamos inatividade. Tudo bem? Quer ajuda?"]
  ↓
[Brio WhatsApp: "Vimos você não usa Pulse. Algo quebrou? Link direto: kba-003"]
  ↓
[CEO notificado via Telegram]
  ↓
[Se resposta negativa ou nenhuma → escalação manual]
```

**Aceitação:**
- [ ] Tabela `churn_alerts` (tenant_id, alert_type, status, created_at)
- [ ] Email template + Brio message
- [ ] Integração com Resend + Telegram API
- [ ] Deduplication: não enviar 2x em 7 dias mesmo alert
- [ ] Tracking: quantos clientes alertados, quantos responderam, quantos churned

**Esforço:** 1 dia  
**Dono:** @tl  
**Sprint:** 1  
**Prioridade:** 🔴 BLOQUEADOR

---

### BKL-CS-006: Knowledge Base — 5 Guias Adicionais

**Descrição:** Criar KBA-006 a KBA-010 (Limite eventos, Export, Self-hosted, NPS/CSAT, CI/CD integration).

**Objetivo:** Completar cobertura 100% de casos de uso documentados.

**KBAs:**
- [ ] KBA-006: "Atingi limite 10k eventos. O que fazer?" (overage pricing + upgrade)
- [ ] KBA-007: "Como exportar dados de testes?" (JSON, CSV, API endpoint)
- [ ] KBA-008: "Pulse tem self-hosted? E LGPD?" (modo cloud vs self-hosted trade-offs)
- [ ] KBA-009: "Por que NPS baixo? Quando dispara survey?" (timing, configuração, análise)
- [ ] KBA-010: "Como integrar Pulse em CI/CD?" (GitHub Actions + feature branch flags)

**Aceitação:**
- [ ] 5 guias publicados em `/docs/guides/`
- [ ] Screencasts: 2-3 min por KBA
- [ ] Exemplos de código prontos (copy-paste)
- [ ] Troubleshooting section em cada guia

**Esforço:** 3 dias  
**Dono:** @pm + support  
**Sprint:** 2  
**Prioridade:** 🔴 BLOQUEADOR

---

## 🟠 TIER 2 — Diferenciadores (30-120 dias) — RECOMENDADO

Implementar após Tier 1. Transformam suporte reativo em proativo.

---

### BKL-CS-010: Dashboard Isolado por Projeto

**Descrição:** Cada cliente vê APENAS seus dados no dashboard (não mistura com outros tenants).

**Objetivo:** Pronto para venda real. Hoje o dashboard é hardcoded para IIT.

**Specs:**
```
[Cliente faz login via auth-service]
  ↓
[Redireciona para `/pulse/dashboard?project={project_id}`]
  ↓
[Dashboard filtra por project_id]
  ↓
[Vê flags próprias, eventos próprios, A/B tests próprios]
  ↓
[Sidebar com projeto selector (se tem multiple)]
  ↓
[Settings: API key, notifications, team members (Pro)]
```

**Aceitação:**
- [ ] Auth: via token de auth-service (bearer token)
- [ ] Query filters: project_id em todas as queries de flags, events, experiments
- [ ] UI: "Project: My App" selector (top-left)
- [ ] Isolation test: client A não vê dados de client B mesmo com acesso direto
- [ ] Performance: queries < 200ms (cache Redis)

**Esforço:** 1 semana  
**Dono:** @tl + @pm  
**Sprint:** 3-4  
**Prioridade:** 🟠 MUITO IMPORTANTE

---

### BKL-CS-011: Community Discord + Bot FAQ

**Descrição:** Servidor Discord público com canais de suporte + bot que responde FAQs.

**Objetivo:** Reduzir 30% do suporte repetitivo. Engajamento cliente.

**Estrutura:**
```
#welcome — onboarding + links rápidos (pinned)
#announcements — releases, status page, downtime notices
#getting-started — curador linkar KBA-001, KBA-002, etc
#feature-flags — questões técnicas sobre flags
#ab-testing — experimentos, significância, interpretação
#self-hosted — deployment em K3s/Docker
#showcase — clientes compartilham case studies
#feedback — feature requests (linkado ao GitHub backlog)
#random — off-topic
```

**Bot capabilities:**
```
?help → lista de comandos
?kba flag-not-activating → retorna KBA-003 summary + link
?cheatsheet flags → dá 5 exemplos curtos de flags
?pricing → matriz de planos
?status → link status page
```

**Aceitação:**
- [ ] Discord server criado (invite link em docs)
- [ ] 7+ canais conforme estrutura
- [ ] Bot (Discord.py) implementado
- [ ] Moderator (@pm, @support) designado
- [ ] Semanal: pinned message com "semana's top questions"
- [ ] Monthly: Pulse team responde thread de feedback

**Esforço:** 3 dias  
**Dono:** @ai (bot) + @pm (moderation)  
**Sprint:** 3  
**Prioridade:** 🟠 IMPORTANTE

---

### BKL-CS-012: Support Patterns Detector

**Descrição:** Cron que agrupa tickets por tema. Se >5 em 90d → escalação BKL automática.

**Objetivo:** Feedback loop sistemático (tickets → roadmap).

**Implementação:**
```sql
-- Tabela de patterns
CREATE TABLE support_patterns (
  id UUID PRIMARY KEY,
  pattern_name TEXT,
  keywords TEXT[],  -- ["flag", "not activating", "targeting"]
  escalation_threshold INT DEFAULT 5,
  period_days INT DEFAULT 90,
  associated_bkl_id UUID,
  created_at TIMESTAMP
);

-- Cron (semanal, quinta-feira 10h):
SELECT 
  pattern_name, 
  COUNT(*) as count
FROM support_tickets st
JOIN support_patterns sp ON st.body LIKE ANY(sp.keywords)
WHERE st.created_at > NOW() - INTERVAL '{sp.period_days} days'
  AND st.status = 'resolved'
GROUP BY pattern_name
HAVING COUNT(*) >= sp.escalation_threshold
ORDER BY count DESC;

-- Se encontra pattern, cria BKL e notifica @pm
```

**Padrões iniciais:**
| Pattern | Keywords | BKL | Threshold |
|---------|----------|-----|-----------|
| flag_not_activating | ["flag", "not activating", "disabled"] | BKL-217 | 5 |
| p_value_confusion | ["p-value", "null", "null", "significance"] | BKL-180 | 4 |
| multi_environment | ["staging", "prod", "environment"] | BKL-180 | 3 |
| sdk_integration | ["SDK", "install", "npm", "error"] | Doc-X | 6 |

**Aceitação:**
- [ ] Padrões table populada (20+ patterns)
- [ ] Cron job funcionando (sem crashes)
- [ ] Auto-creates GitHub issues quando threshold atingido
- [ ] Email para @pm com resumo semanal
- [ ] Rastreamento: quantos BKLs criados, quantos já resolvidos

**Esforço:** 2 dias  
**Dono:** @tl  
**Sprint:** 4  
**Prioridade:** 🟠 IMPORTANTE

---

### BKL-CS-013: Churn Prediction Model

**Descrição:** Regressão logística que prediz prob de churn nos próximos 30d. Features: health score, engagement, etc.

**Objetivo:** Detectar clientes em risco 2-4 semanas antes e oferecer intervenção.

**Modelo:**
```
Features (5):
  1. health_score (0-100)
  2. days_since_last_event
  3. logins_14d
  4. flags_created (trend: crescente vs estável vs decrescente)
  5. event_volume_trend (decrescente = risco)

Target: churned = 1 (cancelou no mês seguinte), 0 (manteve)

Training data: 6 meses histórico de tenants (depois de Tier 1)
Performance: alvo precision >= 80% (menos false positives)
Threshold: se pred >= 0.70 → trigger reengagement offer
```

**Aceitação:**
- [ ] Dataset: 6 meses histórico de tenant metrics + churn labels
- [ ] Train/test split: 70/30
- [ ] Precision >= 80%, recall >= 60%
- [ ] Predictions: rodam diário, armazenam em `tenant_churn_predictions`
- [ ] Dashboard: CS vê lista de "churn risk" tenants ordenados por risk
- [ ] Ação automática: IF pred >= 0.70 AND notificado_antes != true
       → enviar email "Notamos mudança. Quer ajuda?" + Brio escala

**Esforço:** 3 dias (dados + Python sklearn + Go wrapper)  
**Dono:** @data + @tl  
**Sprint:** 5  
**Prioridade:** 🟠 IMPORTANTE

---

### BKL-CS-014: Self-Hosted Health Beacon (Opt-in)

**Descrição:** Cliente self-hosted pode enviar health check semanal para Pulse cloud (opt-in).

**Objetivo:** CS insights sem acesso ao servidor do cliente.

**Fluxo:**
```
[Self-hosted Pulse (v1.0.5) roda em cliente]
  ↓
[Weekly cron: POST /api/v1/health-beacon]
  ↓
[Payload: version, uptime%, event_count_7d, error_count, etc]
  ↓
[Pulse cloud recebe, armazena em beacon_reports]
  ↓
[CS observa: "Cliente rodando v1.0.5 que tem bug X. Email: atualizar"]
  ↓
[Detecta: uptime 95% (baixo). Email: quer ajuda com infra?]
```

**Payload:**
```json
{
  "deployment_id": "hash(license_key + install_date)",
  "version": "1.0.5",
  "timestamp": "2026-04-10T14:30:00Z",
  "uptime_percentage": 99.97,
  "event_count_7d": 2500000,
  "user_count": 350,
  "flag_count": 28,
  "errors_7d": 0,
  "db_connection_pool_used": 45,
  "db_query_p95_ms": 234
}
```

**Aceitação:**
- [ ] Endpoint `/api/v1/health-beacon` (autenticado via license key)
- [ ] Tabela `beacon_reports` armazena histórico
- [ ] Cron no self-hosted (opt-in via env var: ENABLE_HEALTH_BEACON=true)
- [ ] Dashboard: CS vê lista de self-hosted instances com version, uptime, etc
- [ ] Alert: se version é outdated → email "upgrade to X (bugfix)"
- [ ] Alert: se uptime < 95% → email "quer ajuda?"

**Esforço:** 2 dias  
**Dono:** @tl  
**Sprint:** 4  
**Prioridade:** 🟠 IMPORTANTE

---

### BKL-CS-015: ROI Calculator

**Descrição:** Tool interativo no landing page que mostra economia (em R$) de usar Pulse vs concorrentes.

**Objetivo:** Convince buyers: "Você economiza R$ 3.500/ano com Pulse vs LD".

**Inputs:**
```
1. # de engineers: slider 1-100
2. # de eventos/mês: slider 10k-50M
3. Duração setup (horas): slider 0-100
4. Escolha concorrente: LaunchDarkly vs VWO vs PostHog
```

**Outputs:**
```
LaunchDarkly: "10 engineers × USD 10/engineer/mês = USD 100 → R$ 500/mês"
Pulse Growth: "R$ 99/mês"
Economia: "R$ 401/mês = R$ 4.812/ano"

Adicional: setup time savings (LD = 40h → Pulse = 5h)
           = 35h × R$ 250/h = R$ 8.750 one-time savings
```

**Aceitação:**
- [ ] React component (inputs, math, output)
- [ ] Pricing data: atualizado com LD/VWO/PostHog pricing real
- [ ] Shareability: gera link com inputs salvos (UTM tracking)
- [ ] Analytics: rastreia submissions + leads

**Esforço:** 2 dias (React + data)  
**Dono:** @pm  
**Sprint:** 3  
**Prioridade:** 🟠 IMPORTANTE

---

### BKL-CS-016: Upgrade Recommender

**Descrição:** Quiz interativo 1 min que pergunta 5 perguntas e recomenda plano fit.

**Objetivo:** Reduzir paralisia de decisão (Free vs Growth vs Pro).

**Flow:**
```
1. "Quantos engenheiros vai usar?" → 1-5 / 6-20 / 20+
2. "Quantos eventos/mês estima?" → <10k / 100k-500k / >500k
3. "Quer A/B testing?" → sim/não
4. "Quer CSAT/NPS nativo?" → sim/não
5. "Self-hosted ou cloud?" → cloud / self-hosted

Resultado: "Recomendamos Growth" (com explicação)
```

**Aceitação:**
- [ ] React component (quiz, UI bonita)
- [ ] Logic: rules-based routing (simples, não ML)
- [ ] Recomendação mostra plano + 3 benefits principais
- [ ] CTA: "Comece com Growth" → checkout
- [ ] Analytics: quiz completion rate + plano recomendado vs escolhido

**Esforço:** 1 dia  
**Dono:** @pm  
**Sprint:** 3  
**Prioridade:** 🟠 IMPORTANTE

---

### BKL-CS-017: Win-Back Campaign

**Descrição:** Email sequência automática para clientes Free lapsed (não ativam há 30d+).

**Objetivo:** Reativar 5-10% de Free → Growth.

**Sequence:**
```
Email 1 (Day 30): "Sentimos sua falta. Pulse ainda está aqui. Tudo bem?"
  CTA: "Volte e veja novidades"

Email 2 (Day 45): "50% OFF Growth por 1 mês se voltar agora"
  CTA: "Usar desconto" → checkout (auto-apply code)

Email 3 (Day 60): Last chance — desconto expira em 3 dias"
  CTA: "Aproveitar desconto"
```

**Aceitação:**
- [ ] Segment: tenants com last_event_date < 30 days ago AND plan = Free
- [ ] Cron: executa semanal (segunda-feira 10h)
- [ ] Email templates + discount code gerado automático
- [ ] Tracking: open/click/conversion rate
- [ ] Resultado: % que voltaram e upgradearam

**Esforço:** 1 dia  
**Dono:** @pm  
**Sprint:** 5  
**Prioridade:** 🟠 IMPORTANTE

---

### BKL-CS-018: Support Ticket System Upgrade

**Descrição:** Migrar de email para ticket system (ex: Chatwoot, Freshdesk ou homemade no Pulse).

**Objetivo:** Rastreamento, SLA, escalonamento claro.

**Mínimo viable:**
- Email → ticket automático (headers: from, subject, body)
- Status: new, assigned, waiting_customer, resolved
- SLA: Growth <48h, Pro <2h (auto-reminder se SLA broken)
- Integração com Brio: escalação automática cria ticket

**Aceitação:**
- [ ] Tabela `support_tickets` com fields padrão
- [ ] API: GET /tickets, PATCH /tickets/{id}
- [ ] Dashboard: lista de tickets por status
- [ ] SLA timer: mostra tempo desde abertura
- [ ] Integração email: inbound via Resend + outbound

**Esforço:** 2-3 dias (pode usar sistema existente ou build light)  
**Dono:** @tl  
**Sprint:** 4  
**Prioridade:** 🟠 IMPORTANTE

---

### BKL-CS-019: First-Contact Resolution (FCR) Tracking

**Descrição:** Métrica de Brio resolvendo ticket em primeira interação (sem escalação).

**Objetivo:** Medir efetividade de L1 automation.

**Implementação:**
```
Ticket é "FCR" se:
  - Brio respondeu e cliente não replicou (resolvido)
  - OU cliente respondeu "thanks, that fixed it" (resolvido)
  - OU Brio escalou (não é FCR)

Target: 60% de FCR em Tier 1 (Brio + KBA)

Dashboard:
  - FCR rate (% de tickets resolvidos sem escalação)
  - Avg response time (Brio vs @antonio)
  - Ticket volume by category
```

**Aceitação:**
- [ ] Métrica: FCR = resolved after 1 brio message / total tickets
- [ ] Tracking: Brio marca tickets como resolved automaticamente
- [ ] Dashboard: mostra FCR por semana
- [ ] Alvo: >60% FCR

**Esforço:** 1 dia  
**Dono:** @ai  
**Sprint:** 5  
**Prioridade:** 🟠 IMPORTANTE

---

## 🟡 TIER 3 — Premium Services (6+ meses)

Implementar quando tiver produto-market fit (50+ clientes, NPS >= 50).

---

### BKL-CS-020: Customer Success Manager (CSM) Hiring

**Descrição:** Contratar 1 FTE Customer Success Manager dedicado a clientes Pro.

**Objetivo:** Retenção 99%+ em tier premium. Oportunidades de upsell.

**Responsibilities:**
- 1x call mensal com cada Pro customer
- Proactive: "Vi que não usou A/B testing. Quer consultoria?"
- Reports: health score trends, product adoption, NPS
- Escalation: qualquer issue Pro é prioridade
- Upsell: Enterprise, custom integrations

**Critérios de Contratação:**
- 50+ Pro customers OU ARR > R$ 150k/ano em tier premium
- NPS >= 50 (produto stável)
- Churn rate < 5% (trending down)

**Job Description:**
- Tenure: full-time, brasileiro (timezone)
- Salary: R$ 6-8k (SMB CSM BR market rate)
- Manager: @pm

**Aceitação:**
- [ ] JD publicado
- [ ] Candidate screened
- [ ] Onboarded com product training
- [ ] KPIs: customer retention rate, NPS trend, upsell rate

**Esforço:** Ongoing (recruiting + onboarding)  
**Dono:** @pm + HR  
**Sprint:** 6+  
**Prioridade:** 🟡 PREMIUM

---

### BKL-CS-021: Onboarding Call em Produção

**Descrição:** 1x 30-min call agendado automaticamente para Growth customers após signup.

**Objetivo:** Activation rate 80%+ (hoje é ~40% sem call).

**Flow:**
```
[Customer completa signup]
  ↓
[Calendly link enviado: "Agendar onboarding 30min?"]]
  ↓
[Customer escolhe slot (próximas 48h)]
  ↓
[Google Meet link automático]
  ↓
[Call: setup SDK + primeira flag + Q&A]
  ↓
[Zoom recording salvo + notes no ticket]
  ↓
[Follow-up email com SDK snippet + KBAs]
```

**Aceitação:**
- [ ] Calendly integrado com auth-service
- [ ] Email automático com Calendly link (dia 0)
- [ ] Follow-up email com transcrição/notes
- [ ] Metric: % de Growth customers que agendaram call
- [ ] Outcome: activation rate pós-call vs sem call

**Esforço:** 2 dias (Calendly + email automation)  
**Dono:** @pm + support  
**Sprint:** 6  
**Prioridade:** 🟡 PREMIUM

---

### BKL-CS-022: Health Report Mensal (PDF)

**Descrição:** Relatório automatizado em PDF enviado 1x/mês para Pro customers com health score, adoption, NPS trends.

**Objetivo:** Proativo de insights. Engagement.

**Conteúdo do PDF:**
```
1. Health Score: 75/100 (🟢 saudável)
2. Atividade: 8 logins em 30d (média 6)
3. Feature Adoption:
   - Flags criadas: 12 (trend: +2 vs mês passado)
   - A/B tests rodados: 1 (conclusão: variante B ganhou +5%)
   - Events/mês: 850k (95% do limite Pro)
4. NPS: 8/10 (promoter) vs média clientes: 7/10
5. Recomendações:
   - "Você está perto do limite de eventos. Quer ir Pro Unlimited?"
   - "A/B testing adoption é baixa (1 test/mês). Quer consultoria?"
```

**Aceitação:**
- [ ] Template: HTML → PDF (wkhtmltopdf ou similar)
- [ ] Data: aggregated do Pulse backend
- [ ] Distribution: email automatizado (1º dia útil do mês)
- [ ] Analytics: open rate, click throughs
- [ ] Customization: CSM pode adicionar notas pessoais

**Esforço:** 3 dias (template + data aggregation)  
**Dono:** @data + @pm  
**Sprint:** 6  
**Prioridade:** 🟡 PREMIUM

---

### BKL-CS-023: Enterprise SLA 99.99% + Monitoring

**Descrição:** Para tier Enterprise (futuro), garantir 99.99% uptime com status page público + alerting.

**Objetivo:** Confiabilidade contratual.

**Implementação:**
- [ ] Status page pública (status.pulse.institutoitinerante.com.br)
- [ ] Monitoramento: 5 distributed checks/min (Uptime Robot ou internal)
- [ ] Alerting: if downtime > 5min, notifica ops + customer
- [ ] SLA credits: se violado SLA, 10% de desconto no mês seguinte
- [ ] Incident postmortems: publicado <24h após incident

**Aceitação:**
- [ ] Status page em produção
- [ ] Uptime >= 99.99% comprovado em 3 meses
- [ ] SLA document publicado (termos claros)

**Esforço:** 1 semana  
**Dono:** @devops + @tl  
**Sprint:** 7  
**Prioridade:** 🟡 PREMIUM

---

### BKL-CS-024: Private Slack Channel para Pro

**Descrição:** Invitar cada Pro customer em Slack private channel dedicado para suporte direto.

**Objetivo:** Premium feeling. Resposta <1h em vez de <2h via email.

**Setup:**
```
[Customer upgrades para Pro]
  ↓
[Automático: Slack invite gerado para {email}@slack.com]
  ↓
[Channel: #pulse-{company-slug}]
  ↓
[@tl + support monitora 24/7]
  ↓
[SLA: resposta <1h (vs <2h email)]
```

**Aceitação:**
- [ ] Slack workspace criado
- [ ] Bot integrado (Auto-populates channel on Pro signup)
- [ ] @support monitorado 24/7 (rotação)
- [ ] Analytics: response time, resolution rate

**Esforço:** 1 dia (bot + channel template)  
**Dono:** @ai (bot) + @tl  
**Sprint:** 6  
**Prioridade:** 🟡 PREMIUM

---

### BKL-CS-025: Case Studies + Success Stories

**Descrição:** 3-5 case studies publicados de clientes usando Pulse (SocialMake, my-library, etc).

**Objetivo:** Social proof. Marketing.

**Exemplo:**
```
Title: "Como SocialMake usou Pulse para aumentar conversão de Aria em 25%"

Content:
- Desafio: Aria model era custoso. Precisava validar ROI antes de scale.
- Solução: A/B test via Pulse (Aria vs baseline)
- Resultado: Aria convertia 5% mais. Economia: R$ 50k/mês em tokens.
- Quote: "Pulse foi essencial para tomar essa decisão" — Paulo, PM SocialMake

Métricas:
- Duração do teste: 4 semanas
- Amostra: 10k usuários
- Confidence: 99% (p=0.0001)
```

**Aceitação:**
- [ ] 3 case studies: SocialMake, my-library, +1 cliente externo (quando tiver)
- [ ] Publicados em blog + landing page
- [ ] Video testimonial (5 min) de cliente
- [ ] Shareability: PDF downloadável, social cards

**Esforço:** 4 dias (interviews + writing + design)  
**Dono:** @pm + marketing  
**Sprint:** 7  
**Prioridade:** 🟡 PREMIUM

---

### BKL-CS-026: Advanced Onboarding Workshop (Tier 3+)

**Descrição:** 2-hour workshop para Pro/Enterprise customers sobre A/B testing best practices.

**Objetivo:** Elevate customer success. Certify advanced usage.

**Agenda:**
```
1. A/B Testing Fundamentals (20 min)
   - Hypothesis design
   - Sample size calculation
   - P-value + confidence interval

2. Real case study: SocialMake Aria test (30 min)
   - Setup, results, interpretation

3. Hands-on: Design your own test (50 min)
   - 3-4 groups, each designs an A/B test
   - Live feedback

4. Q&A + Certification (20 min)
   - Attendees get "Pulse Certified" badge
```

**Format:**
- Monthly virtual workshop (Zoom)
- Quarterly in-person (São Paulo? Salvador?)

**Aceitação:**
- [ ] Slide deck + workshop materials
- [ ] Zoom recording + follow-up materials
- [ ] Attendee survey (NPS)
- [ ] Certification badge (shareable)

**Esforço:** 3 dias initial + 4h/month ongoing  
**Dono:** @pm + @tl  
**Sprint:** 7+  
**Prioridade:** 🟡 PREMIUM

---

## Summary Table

| BKL | Título | Esforço | Sprint | Status |
|-----|--------|---------|--------|--------|
| **CS-001** | KBA 1-5 | 3-4d | 1 | 🔴 CRITICAL |
| **CS-002** | Health Score System | 2-3d | 1 | 🔴 CRITICAL |
| **CS-003** | Email Sequence | 2d | 1-2 | 🔴 CRITICAL |
| **CS-004** | Brio Integration | 3d | 2 | 🔴 CRITICAL |
| **CS-005** | Churn Alert | 1d | 1 | 🔴 CRITICAL |
| **CS-006** | KBA 6-10 | 3d | 2 | 🔴 CRITICAL |
| **CS-010** | Isolated Dashboard | 5-7d | 3-4 | 🟠 IMPORTANT |
| **CS-011** | Community Discord | 3d | 3 | 🟠 IMPORTANT |
| **CS-012** | Patterns Detector | 2d | 4 | 🟠 IMPORTANT |
| **CS-013** | Churn Prediction ML | 3d | 5 | 🟠 IMPORTANT |
| **CS-014** | Health Beacon (Self-hosted) | 2d | 4 | 🟠 IMPORTANT |
| **CS-015** | ROI Calculator | 2d | 3 | 🟠 IMPORTANT |
| **CS-016** | Upgrade Recommender | 1d | 3 | 🟠 IMPORTANT |
| **CS-017** | Win-Back Campaign | 1d | 5 | 🟠 IMPORTANT |
| **CS-018** | Ticket System | 2-3d | 4 | 🟠 IMPORTANT |
| **CS-019** | FCR Tracking | 1d | 5 | 🟠 IMPORTANT |
| **CS-020** | CSM Hiring | Ongoing | 6+ | 🟡 PREMIUM |
| **CS-021** | Onboarding Call | 2d | 6 | 🟡 PREMIUM |
| **CS-022** | Health Report PDF | 3d | 6 | 🟡 PREMIUM |
| **CS-023** | Enterprise SLA 99.99% | 5-7d | 7 | 🟡 PREMIUM |
| **CS-024** | Private Slack | 1d | 6 | 🟡 PREMIUM |
| **CS-025** | Case Studies | 4d | 7 | 🟡 PREMIUM |
| **CS-026** | Advanced Workshop | 3d + 4h/mo | 7+ | 🟡 PREMIUM |

---

## Implementation Roadmap (Timeline)

### Week 1-2 (Tier 1 Core)
```
Mon: CS-001 start (KBA writing)
Tue: CS-002 start (health score migration)
Wed: CS-003 start (email templates)
Thu-Fri: Review, test, deploy
```

### Week 3-4 (Tier 1 Complete)
```
Mon: CS-004 start (Brio integration)
Tue: CS-005 deploy (churn alerts live)
Wed-Fri: CS-006 (remaining KBAs), polish
Weekend: QA all Tier 1
```

### Week 5-8 (Tier 2 Start)
```
Week 5: CS-010 (dashboard isolation)
Week 6: CS-011 (Discord), CS-012 (patterns detector)
Week 7: CS-015 (ROI calc), CS-016 (recommender)
Week 8: CS-014 (beacon), CS-017 (win-back), CS-018 (tickets)
```

### Week 9-12 (Tier 2 Advanced)
```
Week 9: CS-013 (churn ML)
Week 10: CS-019 (FCR tracking), review Tier 2
Week 11-12: Polish, testing, documentation
```

### Month 5+ (Tier 3 Premium)
```
Month 5: CS-020 (CSM hiring), CS-021 (onboarding calls)
Month 6: CS-022 (health reports), CS-023 (SLA)
Month 6-7: CS-024 (Slack), CS-025 (case studies), CS-026 (workshop)
```

---

**Total Investment:**
- **Tier 1:** ~16-18 days engineering + 5 days PM/content
- **Tier 2:** ~35-40 days engineering + 10 days PM/content + @data (3 days)
- **Tier 3:** ~40-50 days eng + 1 FTE CSM + 15 days PM

**Expected ROI:**
- Tier 1: Enabled go-to-market (0 churn risk due to poor support)
- Tier 2: 3% churn rate (vs 5-8% industry) = +2-5% MRR retained
- Tier 3: 70% Net Revenue Retention (Pro tier) = sustainable growth + upsell

---

**Next Step:** Prioritize CS-001 to CS-006 and start execution Sprint 1 (Monday).
