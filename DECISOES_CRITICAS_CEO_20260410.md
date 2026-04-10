# PULSE — Decisões Críticas para Aprovação CEO (Antonio)
**Consolidação de análises 13 agentes IIT**

**Preparado por:** PM (SOUL)  
**Data:** 10 de abril de 2026  
**Tempo leitura:** 10 minutos  
**Ação requerida:** Aprovações em 48h para iniciar PHASE 0  

---

## O PROBLEMA EM 1 PARÁGRAFO

Pulse é um **MVP funcional mas inseguro** que não pode ser vendido para clientes externos. Existem **10 gaps críticos** (5 de segurança, 3 de analytics, 2 de SDK) que bloqueiam receita. O produto precisa de 4 meses de trabalho estruturado (PHASE 0-2) para ser vendível. A decisão hoje é: **Aprovamos esses 4 meses, definimos pricing, e escolhemos qual segmento atacar primeiro?**

---

## DIAGNÓSTICO EXECUTIVO

### Score Geral: 4/10 — NÃO VENDÍVEL HOJE

| Dimensão | Score | Status |
|----------|-------|--------|
| Feature Completeness | 6/10 | MVP incompleto (análise, SDKs faltam) |
| Segurança | 🔴 3/10 | CRÍTICO — cross-tenant leaks, SQL injection |
| Escalabilidade | 5/10 | Frágil — single PG, 1 replica K3s |
| Analytics | 2/10 | Mudo — sem p-value, sem funis |
| GTM | 4/10 | Inexistente — sem pricing page, onboarding |
| **RESULTADO** | **4/10** | **PHASE 0 OBRIGATÓRIO ANTES DE VENDAS** |

### Bloqueadores Críticos (Stop-Ship)

| Item | Severidade | Tempo Fix |
|------|-----------|----------|
| Cross-tenant data leak (Tenant A vê Tenant B flags) | CRÍTICA | 4h |
| SQL injection em event_repository.go | CRÍTICA | 3h |
| Audit log inexistente | CRÍTICA | 3 semanas |
| Chi-squared p-value ausente | ALTA | 5 semanas |
| SDK web não funciona (ES module) | ALTA | 5 semanas |

**Conclusão:** Sem PHASE 0 (30 dias), Pulse não é seguro para produção com clientes pagantes.

---

## TRÊS DECISÕES CRÍTICAS

### DECISÃO 1: Segurança — Pausar Vendas até PHASE 0?

**Status Atual:**
- Cross-tenant leaks confirmados (Tenant A vê Tenant B flags)
- SQL injection confirmada em queries de eventos
- Sem audit log de quem fez o quê
- Sem rate limiting por projeto (abuse risk)

**Opção A (RECOMENDADO — Baixo Risco):**
- **Ação:** Pausar vendas externas até 30/04/2026 (PHASE 0 complete)
- **Prazo:** 30 dias para fechar segurança + deploy em K3s (3 replicas)
- **Aprovação:** Após security audit de @tl + @qa
- **Benefício:** Zero risco LGPD/contratual
- **Custo:** 0 (time alocado já existe)

**Opção B (Alto Risco — NÃO RECOMENDADO):**
- Começar vendas beta com disclaimer "fase de desenvolvimento, dados podem vazar"
- Risco: LGPD violation, leak de dados cliente, dano reputacional
- Improvável conversão (clientes sérios não aceitam esse risco)

**RECOMENDAÇÃO:** Opção A. Aprovar PHASE 0 como **P0 obrigatório** para começar qualquer venda externa. Isso não afeta uso interno (IIT pode usar Pulse hoje).

---

### DECISÃO 2: Pricing — Elevar de R$299 → R$399 (Growth R$799 → R$1.099)?

**Análise Competitiva:**

| Plataforma | Starter | Growth | Diferencial |
|------------|---------|--------|------------|
| **LaunchDarkly** | USD 10 (R$52) | USD 200-400 (R$1.000-2.000) | Audit log, Enterprise SLA |
| **VWO** | USD 198 (R$1.030) | USD 422 (R$2.194) | Heatmaps, session recording |
| **PostHog** | USD 45-50 (R$234-260) | USD 250-1.000+ (eventos) | Events-based, open-source |
| **Pulse (ATUAL)** | **R$299** | **R$799** | Self-hosted em R$, health score |
| **Pulse (PROPOSTO)** | **R$399** | **R$1.099** | 50% abaixo VWO, premium vs PostHog |

**Impacto Financeiro (12 meses):**
- Hoje: R$0 (produto interno)
- Com preço atual (R$299/R$799): ~R$40-50k ARR (30-50 clientes)
- Com preço elevado (R$399/R$1.099): ~R$60-80k ARR (+50%)

**RECOMENDAÇÃO:** Elevar para **R$399 / R$1.099**. Mantém Pulse 50% abaixo VWO, premium vs PostHog, defensável por: self-hosted, pricing em R$, health score nativo.

---

### DECISÃO 3: ICP & GTM — Qual Segmento Atacar Primeiro?

**Três Opções:**

#### Opção A: FinTech (Pix, pagamento)
- **Viabilidade:** Muito Alta (LGPD-sensitive, willingness to pay)
- **Tempo Sales:** 3-6 meses (ciclo longo, due diligence)
- **Esforço Sales:** Alto (vendas técnicas, compliance)
- **Vencedor para:** Enterprise, long-term ARR

#### Opção B: EdTech (Cursos online, bootcamp) ⭐ RECOMENDADO
- **Viabilidade:** Alta (open to experimentation, A/B testing mindset)
- **Tempo Sales:** 1-2 meses (short sales cycle)
- **Esforço Sales:** Baixo (product-led growth possível)
- **Vencedor para:** Crescimento rápido early-stage, feedback loop

#### Opção C: SaaS Genérico (agências, consultoria)
- **Viabilidade:** Muito Alta (numeroso mercado)
- **Tempo Sales:** 2-3 meses (médio)
- **Esforço Sales:** Médio (mix product-led + outbound)
- **Vencedor para:** Volume, diversificação

**RECOMENDAÇÃO:** **EdTech (Opção B)** como primeira prioridade. Razão:
1. Problema claro: A/B testing de onboarding (qual estrutura de aula converte mais)
2. Público open para experimentation (não risk-averse como FinTech)
3. Sales product-led (demo simples em 15 min)
4. Feedback loop rápido (ajustes em 2 semanas vs 3 meses FinTech)

**Ação:** Alocar **1 person full-time Growth** para EdTech (~300 MQLs no BR identificáveis em 30 dias).

---

## ROADMAP APROVADO (90-365 dias)

### PHASE 0 — Security Lockdown (Próximos 30 dias) 🔴
**Objetivo:** Fechar segurança + disponibilidade antes de qualquer cliente

**Itens:**
- Corrigir cross-tenant leak (4h)
- Corrigir SQL injection (3h)
- Audit log de flag changes (3 semanas)
- Deploy em K3s 3 replicas (2 dias)
- Security audit de @tl (1 semana)

**Resultado:** Pulse seguro para produção interna. Pronto para beta clientes.

**Entrada para clientes:** Nenhuma (internal only).

---

### PHASE 1 — MVP Vendível (Dias 30-90) 🟠
**Objetivo:** Feature parity com LaunchDarkly starter. Pronto para 5-10 clientes Growth.

**Itens:**
- Chi-squared p-value + confidence interval para A/B tests (5 semanas)
- SDK web via CDN (IIFE + track real) (5 semanas)
- Ambientes por projeto em flags (2 dias)
- Project scoping em events (2 semanas)

**Resultado:** Pulse com análise estatística de A/B + SDK web funcional.

**Entrada para clientes:** Growth plan (R$799/mês) — 5 projetos, 500k eventos/mês, A/B analysis.

**Target:** 5-10 clientes beta, R$4-8k MRR.

---

### PHASE 2 — Paridade Competitiva (Dias 90-180) 🟡
**Objetivo:** 80% de funcionalidade vs LaunchDarkly/VWO. Crescimento agressivo.

**Itens:**
- `@pulse/react` com hooks (useFlag, useTrack) (7 semanas)
- Funis de conversão (SQL window functions) (14 semanas)
- Metric assignment (flag → métrica) (1 semana)
- Rate limiting + Quota enforcement (3 semanas)
- Dashboard multi-projeto (1 semana)
- Webhooks (Slack, n8n) (3 semanas)
- Pricing page pública + free tier (1 semana)

**Resultado:** Pulse com paridade funcional para 80% casos de uso de LD/VWO.

**Entrada para clientes:** Free tier (10k eventos/mês) → Growth (R$1.099) → Pro (R$2.399).

**Target:** 20-30 clientes, R$20-30k MRR.

---

### PHASE 3 — Premium Differentiation (Dias 180-365) 🟢
**Objetivo:** Diferenciadores únicos. Posicionamento defensável.

**Itens:**
- Funis de feature flags (único mercado)
- Health score configurável via UI
- Churn prediction com ML leve
- Schema registry leve
- NPS triggered por evento
- Data export (BigQuery, S3, CSV)
- Streaming SSE para flags (<1s)

**Resultado:** Pulse com features exclusivas vs LaunchDarkly/VWO.

**Entrada para clientes:** Pro+ (R$3.000+/mês) — Enterprise features.

**Target:** 50+ clientes, R$60-80k MRR.

---

## MÉTRICAS DE SUCESSO (12 MESES)

### North Star Metric: Activated Features per Active User per Week (AFAW)

**Definição:**
- Activated Feature = flag avaliada como `true` para usuário
- Active User = ≥1 flag evaluation no último dia
- AFAW = (total activations) / (unique users) / 7 dias

**Targets por Trimestre:**

| Trimestre | AFAW | Clientes | MRR | Nota |
|-----------|------|----------|-----|------|
| Q2 (atual) | 0.5 | 1 (IIT) | R$0 | Internal only |
| Q3 | 1.2 | 5-8 | R$3-5k | PHASE 1 complete |
| Q4 | 2.1 | 15-25 | R$15-25k | PHASE 2 growing |
| Q1 2027 | 3.5 | 35-50 | R$40-60k | PHASE 3 benefits |

**Por quê essa métrica?**
- Alinhada com núcleo do produto (flags)
- Indicadora de saúde real (não vanity metric)
- Preditor de retenção (times que experimentam renovam)

---

## ALOCAÇÃO DE RECURSOS (90 dias)

| Função | Alocação | Responsável | Esforço |
|--------|----------|------------|---------|
| Backend Lead | 1.0 FTE | @tl | Security + analysis |
| Frontend Lead | 0.8 FTE | @frontend | A/B dashboard + SDK |
| QA | 0.5 FTE | @qa | Testes security |
| PM | 0.3 FTE | @pm | Specs + priorização |
| Growth | 1.0 FTE (PHASE 1) | @growth | EdTech outreach |

**Timeline de Sprints:**
- Semana 1-2: PHASE 0 security (2 devs, 1 QA)
- Semana 3-4: PHASE 0 hardening + specs PHASE 1
- Semana 5-8: PHASE 1 development (chi-squared, SDK, dashboard)
- Semana 9-12: PHASE 1 launch + beta sales (Growth)

---

## RISCOS CRÍTICOS E MITIGAÇÕES

| Risk | Prob | Impact | Mitigação |
|------|------|--------|-----------|
| **Cross-tenant leak encontrado em produção** | Alta | CRÍTICA | Penetration test completo em PHASE 0 |
| **A/B analysis calcula p-value errado** | Média | Alta | Validar fórmula com estatístico antes de launch |
| **SDK web cria performance issue em frontend** | Média | Alta | Load test em 3 apps reais (Libri, Nitro, SocialMake) |
| **Demanda por Pulse é zero** | Média | Média | 3 clientes alpha em PHASE 1 = validação |
| **Concorrente reduz preço (PostHog)** | Média | Média | Diferencial é health score + audit log, não preço |

---

## DECISÕES REQUERIDAS (Próximas 48 horas)

### ✅ DECISÃO 1: Segurança
**Pergunta:** Pausamos vendas externas até 30/04 (PHASE 0)?
- [ ] SIM — Opção A (RECOMENDADO, zero risk)
- [ ] NÃO — Opção B (alto risk, não recomendado)

**Default:** SIM

---

### ✅ DECISÃO 2: Pricing
**Pergunta:** Elevar para R$399 / R$1.099 em 01/05?
- [ ] SIM — Elevamos agora, defensável vs mercado
- [ ] NÃO — Mantemos R$299/R$799, uso como argumento underpriced

**Default:** SIM (R$399 / R$1.099)

---

### ✅ DECISÃO 3: ICP
**Pergunta:** Qual segmento ataca primeiro?
- [ ] EdTech (RECOMENDADO) — rápido, product-led
- [ ] FinTech — higher-touch, longer ARR
- [ ] SaaS genérico — volume

**Default:** EdTech

---

### ✅ DECISÃO 4: Alocação
**Pergunta:** Aprovamos 4 pessoas × 4 meses para PHASE 0-1?
- [ ] SIM — @tl, @frontend, @qa, @growth alocados
- [ ] NÃO — Reduzir escopo ou timeline

**Default:** SIM

---

## PRÓXIMAS AÇÕES (Imediatas)

**Hoje (10/04):**
- [ ] CEO aprova decisões 1-4 (48h window)

**Segunda (15/04) — Kick-off PHASE 0:**
- [ ] @tl inicia code review de cross-tenant leak + SQL injection
- [ ] @qa estrutura teste de penetration
- [ ] @pm cria BKLs no Linear com RICE scores
- [ ] @growth identifica 10 EdTech prospects no BR

**30/04 — Gate PHASE 0:**
- [ ] Security audit sign-off de @tl
- [ ] K3s rodando 3 replicas
- [ ] Pricing page atualizada com R$399/R$1.099
- [ ] 1 cliente EdTech em early conversation

---

## SUMMARY EXECUTIVO (1 página para passagem)

**Pulse Status:**
- MVP funcional internamente, inseguro para clientes
- 10 gaps críticos bloqueando receita
- Score 4/10 — precisa 4 meses (PHASE 0-2)

**Plano:**
- PHASE 0 (30d): Segurança — corrigir cross-tenant leak, SQL injection, audit log
- PHASE 1 (60d): MVP vendível — A/B analysis + SDK web
- PHASE 2 (90d): Paridade — React hooks, funis, quota

**Pricing:**
- Elevar: R$299 → R$399 (Starter), R$799 → R$1.099 (Growth)
- Defensável: 50% abaixo VWO, premium vs PostHog

**ICP:**
- EdTech (cursos, bootcamps) como primeira prioridade
- Problema claro (A/B testing onboarding), sales rápido

**Alocação:**
- 4 pessoas × 4 meses
- @tl (@backend), @frontend, @qa, @growth (EdTech)

**Resultado 12 meses:**
- 50+ clientes, R$60-80k MRR
- Posição defensável vs LD/VWO no mercado BR

---

**Documento:** DECISOES_CRITICAS_CEO_20260410.md  
**Preparado por:** PM (SOUL)  
**Aguardando:** Aprovação CEO (48h)  

