# Backlog Comercial & Pricing — Pulse SaaS

**Data:** 10 de abril de 2026  
**Propriedade:** Comercial | Prioridade: Crítica para go-to-market  
**Sprint Recomendado:** Q2-Q4 2026 (imediato até fim ano)

---

## 📋 BLOQUEADORES CRÍTICOS (Implementar antes de vender)

### COMM-001 | Reescrever SDK Web (IIFE) + Implementar track() real

**Tipo:** Feature | **Esforço:** 5 dias | **Prioridade:** CRÍTICO | **Blocker:** SIM

**Descrição:**
O arquivo `sdk/pulse.js` está como módulo ES (`export { PulseSDK }`) — não funciona como `<script src="cdn">`. A função `track()` é no-op (não envia eventos). Sem isso, nenhum frontend envia eventos ao Pulse.

**Necessário:**
1. Reescrever como IIFE: `(function(window){ window.Pulse = { getFlags, track, identify } })(window)`
2. Implementar `POST /ingest` real em `track()` com `api_key`, `event_name`, `properties`
3. Auto-capture de `page_view` no load
4. Fallback offline: flags retornam `false`, eventos são descartados silenciosamente
5. Snippet de instalação (10 linhas, carrega async)

**Por quê:**
Sem isso, o Pulse é cego para dados de frontend. Toda métrica de "conversão", "clicks", "retenção" é inútil.

**Definição de Pronto:**
- [ ] `<script src="https://cdn.pulse-iit.com/v1/pulse.js"></script>` funciona em navegador
- [ ] `window.Pulse.track('button_click', { button_id: 'x' })` envia POST para `/ingest`
- [ ] 98%+ de uptime em ingestão (test com 1k req/s)
- [ ] Testes unitários em `sdk/js/pulse.test.js`
- [ ] Documentação em `sdk/js/README.md` com exemplo de instalação

**Aceitação:** Libri captura eventos de frontend; SocialMake publica no dashboard

---

### COMM-002 | Análise Estatística de Experimentos (p-value + CI)

**Tipo:** Feature | **Esforço:** 5 dias | **Prioridade:** CRÍTICO | **Blocker:** SIM

**Descrição:**
Variantes A/B existem em `feature_flags` (campo `variants: [{key, weight}]`), mas não há cálculo de significância. Sem p-value e confidence interval, "A/B testing" é apenas divisão aleatória de tráfego — não experimentação real.

**Necessário:**
1. Implementar chi-squared test em Go (stdlib)
2. Calcular p-value, confidence interval (95%), effect size
3. Endpoint `GET /experiments/{experiment_id}/stats` retorna estatísticas
4. Dashboard mostra: "Variant B: +15% conversão, p=0.03 (significante), CI=[5%, 25%]"
5. Badge visual: ✅ Significante / ⚠️ Inconclusivo / ❌ Não significante

**Por quê:**
VWO core feature; sem isso não competimos como alternativa a VWO para A/B testing.

**Definição de Pronto:**
- [ ] `GET /experiments/exp-123/stats` retorna `{p_value: 0.03, confidence_interval: [0.05, 0.25], significant_at_95: true}`
- [ ] Testes com dados sintéticos (1k vs 2k conversões)
- [ ] Dashboard mostra badge significância
- [ ] Docs: "Como ler resultados de experimento"

**Aceitação:** SocialMake A/B testing "novo-checkout" mostra significância automática

---

### COMM-003 | Audit Log de Flags (Rastreabilidade)

**Tipo:** Feature | **Esforço:** 3 dias | **Prioridade:** CRÍTICO | **Blocker:** SIM

**Descrição:**
Sem registro de quem mudou qual flag e quando, usar flags em produção é arriscado. Clientes enterprise/regulated precisam de audit trail para compliance.

**Necessário:**
1. Migration: tabela `flag_changes` com columns:
   - `flag_id`, `project_id`, `actor_user_id`, `timestamp`
   - `change_type` (created/updated/deleted)
   - `previous_state`, `new_state` (JSONB diff)
2. Middleware: toda mudança em flag → insert `flag_changes`
3. UI: Aba "Audit" em cada flag mostrando histórico completo
4. API: `GET /projects/{id}/flags/{flag_id}/audit` retorna histórico

**Por quê:**
LaunchDarkly core feature; sem isso não somos alternativa para produção.

**Definição de Pronto:**
- [ ] Migration 020 aplicada
- [ ] Dashboard mostra: "2026-04-10 14:30 — João Silva habilitou rollout 50% para feature-x"
- [ ] API retorna histórico completo
- [ ] Testes: verificar que cada mudança está em `flag_changes`

**Aceitação:** Auditor de fintech consegue rastrear todas as mudanças de flags em relatório

---

### COMM-004 | Ambientes de Feature Flags (staging/prod)

**Tipo:** Feature | **Esforço:** 1 dia | **Prioridade:** CRÍTICO | **Blocker:** SIM

**Descrição:**
Eventos já têm campo `environment` (migration 013), mas flags não. Clientes precisam de flags diferentes em staging vs produção — não pode haver "um único truth" de flag entre ambientes.

**Necessário:**
1. Migration: adicionar coluna `environment` a `feature_flags`
2. Valores: staging, production, development (customizável por projeto)
3. Índice composto: `(project_id, environment, key)` — flags isoladas por env
4. UI: seletor "Environment" no dashboard; cada environment mostra suas flags
5. SDK: passar `environment` em contexto de avaliação: `isEnabled('flag-x', {environment: 'prod'})`

**Por quê:**
Separação staging/prod é segurança básica em produção.

**Definição de Pronto:**
- [ ] Flags em staging não afetam produção
- [ ] SDK consegue avaliar flags por environment
- [ ] Dashboard permite criar flags diferentes por environment
- [ ] Migração garante que flags antigas herdam `environment='production'`

**Aceitação:** SocialMake tem feature-x desativado em staging, ativado 50% em prod

---

## 🔴 ALTA PRIORIDADE (Q2 2026)

### COMM-005 | Dashboard Multi-tenant (Isolamento por Projeto)

**Tipo:** Feature | **Esforço:** 1 semana | **Prioridade:** ALTA | **Blocker:** NÃO

**Descrição:**
Dashboard atual é hardcoded para IIT (vê todos os projetos). Em SaaS, cada cliente só vê seus próprios projetos — isolamento obrigatório.

**Necessário:**
1. Seletor de projeto no header (mostrar só projetos do usuário logado)
2. Filtros: `/projects/{project_id}/flags`, `/projects/{project_id}/events`, etc
3. Breadcrumb: "Pulse > [Projeto A] > Flags"
4. Query filters garantem `project_id` sempre (SQL security)

**Definição de Pronto:**
- [ ] Cliente Growth A vê só seus 10 projetos
- [ ] Cliente Growth A **não consegue** acessar dados de cliente Growth B (nem via API)
- [ ] Dashboard carrega rápido (<1s mesmo com 100 projetos)

**Aceitação:** Dois clientes Growth no mesmo SaaS não conseguem ver dados um do outro

---

### COMM-006 | Quota Enforcement (Bloqueio de Overage)

**Tipo:** Feature | **Esforço:** 3 dias | **Prioridade:** ALTA | **Blocker:** NÃO

**Descrição:**
Sem enforcement de quota, um cliente Growth que ultrapassar 5M eventos/mês consegue enviar unlimited — afeta SLA de todos. Precisa de middleware que bloqueia (status 429) quando limite é atingido.

**Necessário:**
1. Middleware em `/ingest` que verifica:
   - `project_id` → `plan_tier` → `max_events_per_month`
   - Redis counter: `events:month:{project_id}:{YYYY-MM}`
   - Se counter > limit → retorna 429 Quota Exceeded
2. Reset automático: 00:00 UTC primeira do mês
3. Logs: quadro de comando mostra "Growth-x ultrapassou quota no dia 15"

**Definição de Pronto:**
- [ ] Teste: Growth com 5M limit envia 5.1M eventos → 429 no 5.1M-ésimo
- [ ] Redis counter está sincronizado
- [ ] Reset funciona corretamente (UTC midnight)

**Aceitação:** Cliente Growth recebe 429 quando atinge 5M de eventos; página explica como fazer upgrade

---

### COMM-007 | Rate Limiting por API Key

**Tipo:** Feature | **Esforço:** 2 dias | **Prioridade:** ALTA | **Blocker:** NÃO

**Descrição:**
DDoS acidental ou malicioso: evitar um cliente que ataca `/ingest` com 10k req/s afete outros. Token bucket por API key.

**Necessário:**
1. Middleware em todos endpoints
2. Limites por plan:
   - Starter: 100 req/s
   - Growth: 500 req/s
   - Pro: 2000 req/s
3. Redis: `rate_limit:{api_key}` com token bucket
4. Resposta 429: retry-after header

**Definição de Pronto:**
- [ ] Starter throttled em 100 req/s; Growth em 500
- [ ] Teste de carga: 1000 req/s com 2 keys (500 cada) funciona
- [ ] Página de erro 429 explica rate limiting

**Aceitação:** Um cliente não consegue derrubar o serviço de outro via rate limiting

---

### COMM-008 | Webhooks para Integração

**Tipo:** Feature | **Esforço:** 3 dias | **Prioridade:** ALTA | **Blocker:** NÃO

**Descrição:**
Cliente Growth quer notificação em Slack quando feature é ativada, ou quando evento chega. Webhooks permitem integrações com n8n, Zapier, sistemas internos.

**Necessário:**
1. Admin UI: "Settings > Webhooks" — adicionar URL + eventos (flag_enabled, event_tracked, etc)
2. Roteador: quando evento dispara, POST JSON para URL configurada
3. Retry logic: exponential backoff se falhar (3 tentativas)
4. Logging: quadro de comando mostra "Webhook POST para slack.com falhou com 500"

**Definição de Pronto:**
- [ ] Customer Growth configura webhook para Slack
- [ ] Quando flag é ativado, Slack recebe POST JSON em tempo real
- [ ] Falhas são logadas e retried

**Aceitação:** Slack do cliente recebe mensagem: "Feature feature-x foi habilitado 50% por João Silva"

---

### COMM-009 | SDKs React e React Native

**Tipo:** Feature | **Esforço:** 2 semanas | **Prioridade:** ALTA | **Blocker:** NÃO

**Descrição:**
Libri e Nitro precisam de SDKs idiomáticos. React hooks para DX melhor; React Native com AsyncStorage para cache offline.

**Necessário:**
1. **@pulse/react** (NPM package):
   - `<PulseProvider apiKey="..." >` context provider
   - `useFlags()` hook → `{flagKey: bool}`
   - `useTrack()` hook → `(eventName, props) => void`
   - `useExperiment()` hook → `{variant: 'a'|'b', isLoading}`
2. **@pulse/react-native**:
   - Mesmo API que React
   - AsyncStorage para cache flags offline
   - Background sync de eventos (quando conecta)

**Definição de Pronto:**
- [ ] SocialMake usa: `const flag = useFlags().featureX`
- [ ] Libri coleta eventos: `const track = useTrack(); track('purchase')`
- [ ] Testes: hooks funcionam em React/RN
- [ ] Documentação em `sdk/react/README.md` e `sdk/react-native/README.md`

**Aceitação:** Libri publica produtos via Pulse flags; analytics coleta eventos

---

## 🟡 MÉDIA PRIORIDADE (Q3 2026)

### COMM-010 | Funis de Conversão

**Tipo:** Feature | **Esforço:** 2 semanas | **Prioridade:** MÉDIA | **Blocker:** NÃO

**Descrição:**
Product analytics core: rastrear abandono por etapa (signup → verify email → first login → payment). PostHog tem; Pulse não.

**Necessário:**
1. UI: "Funnels" aba no dashboard
2. Configurar: "Etapa 1: event_name='signup', Etapa 2: event_name='verify_email', etc"
3. Visualizar: "Step 1: 1000 usuarios, Step 2: 800 (80% retention), Step 3: 400 (50%), Step 4: 100 (25%)"
4. SQL: window functions `PARTITION BY user_id ORDER BY timestamp` para calcular retenção entre etapas

**Definição de Pronto:**
- [ ] Dashboard mostra funil com % de conversão por etapa
- [ ] Detalhe: nome de usuários que caíram em cada etapa (privacy: anonimizado)
- [ ] Performance: funil com 10k usuários carrega em <2s

**Aceitação:** Customer Growth vê que apenas 25% dos usuários completam onboarding; identifica bottleneck

---

### COMM-011 | Metric Assignment (Flag → Conversão)

**Tipo:** Feature | **Esforço:** 1 semana | **Prioridade:** MÉDIA | **Blocker:** NÃO

**Descrição:**
Vincular um experimento (flag A/B) a uma métrica de negócio. Ex: "Flag novo-checkout" está mapeada a "taxa de conversão" — dashboard mostra impacto automático.

**Necessário:**
1. UI: ao criar experimento, seletor "Métrica" (lista de eventos já trackados)
2. Join: quando evento chega, verificar: "usuário tem flag novo-checkout ativado?" → registrar impacto
3. Dashboard: mostra "Variante B (+15% conversão, p=0.03)"

**Definição de Pronto:**
- [ ] Customer Growth cria experimento "novo-checkout", seleciona métrica "purchase"
- [ ] Dashboard calcula: "Grupo A: 10% conversion, Grupo B: 12% conversion, +2pp diferença"
- [ ] p-value é calculado automaticamente

**Aceitação:** Experimentação real: não é só divisão de tráfego, é impacto em métrica de negócio

---

### COMM-012 | Health Score Configurável por Cliente

**Tipo:** Feature | **Esforço:** 1 semana | **Prioridade:** MÉDIA | **Blocker:** NÃO

**Descrição:**
Health score hoje é hardcoded em Go (4 dimensões: login, content, adoption, subscription). Clientes SaaS Growth/Pro precisam customizar pesos — Libri tem outras métricas que importam.

**Necessário:**
1. Admin UI: "Settings > Health Score Configuration"
2. Sliders: "Login: 40%, Content: 30%, Adoption: 20%, Custom metric: 10%"
3. Custom metric: campo texto que mapeia a um evento (ex: "library_viewed")
4. Recalcular: score é recompilado mensalmente com novos pesos

**Definição de Pronto:**
- [ ] Customer Pro customiza pesos via UI
- [ ] Score é recalculado com novos pesos
- [ ] Histórico mostra scores antigos vs novos (não retroativo)

**Aceitação:** Libri configura health score com peso 20% em "content_viewed"; churn predictions melhoram

---

## 🟢 BAIXA PRIORIDADE (Q4 2026+)

### COMM-013 | Session Recording e Heatmaps

**Tipo:** Feature | **Esforço:** 4-6 semanas | **Prioridade:** BAIXA | **Status:** Postergar

**Descrição:**
VWO diferencial; PostHog oferece. Captura DOM mutations e randa playback; heatmaps de clicks/scrolls. Alto esforço, alto custo storage, ROI baixo para early-stage.

**Nota:** Considerar apenas se 5+ clientes Growth pedirem explicitamente. Senão, adiar para ano 2.

---

### COMM-014 | Approval Workflow para Flags

**Tipo:** Feature | **Esforço:** 2 semanas | **Prioridade:** BAIXA | **Status:** Postergar até 10+ clientes Growth pedirem

**Descrição:**
Fluxo de review: flag change é proposto, reviewer aprova antes de ativar em produção. LaunchDarkly Pro feature. Relevante apenas para times com 15+ engenheiros.

---

### COMM-015 | Schema Registry com Validação

**Tipo:** Feature | **Esforço:** 1 semana | **Prioridade:** BAIXA | **Status:** Postergar

**Descrição:**
Governance leve: tabela `event_schemas` com validação opcional. Avo-style. Útil apenas com 10+ clientes multi-team.

---

## 🎯 ESTRUTURA DE PRICING (Pendente Aprovação CEO)

### Proposta de Reajuste

| Tier | Preço Atual | Preço Proposto | Justificativa |
|---|---|---|---|
| **Starter** | R$299/mês | **R$399/mês** (+33%) | Margem atual 40%; 30% abaixo VWO USD 198 |
| **Growth** | R$799/mês | **R$1.099/mês** (+38%) | Margem atual 35% → 75%; ainda 50% abaixo VWO Growth |
| **Pro** | R$1.999/mês | **R$2.399/mês** (+20%) | Margem 47% → 70%; 85% abaixo LD Enterprise |
| **Self-Hosted** | R$24k/ano | **R$50k/ano** (+108%) | Model A: Licença + Support separado; ainda 75% abaixo LD |

**Impacto esperado:**
- ARR multiplica ~1.5-2x com mesmo volume
- CAC payback reduz de 6m → 2-3m
- Gross margin: 45% → 65-75%

---

## 📅 ROADMAP DE EXECUÇÃO

### Q2 2026 (Abr-Jun)
- [ ] COMM-001: SDK IIFE + track() — **SEMANAL 1**
- [ ] COMM-002: Stats experimentos — **SEMANAL 2**
- [ ] COMM-003: Audit log — **SEMANAL 2**
- [ ] COMM-004: Ambientes flags — **SEMANAL 2**
- [ ] **Primeira venda:** 5 clientes Starter @ R$399
- [ ] Publicar pricing page públicamente

### Q3 2026 (Jul-Set)
- [ ] COMM-005: Dashboard multi-tenant — **SEMANAL 1**
- [ ] COMM-006: Quota enforcement — **SEMANAL 1**
- [ ] COMM-007: Rate limiting — **SEMANAL 1**
- [ ] COMM-008: Webhooks — **SEMANAL 1**
- [ ] COMM-009: SDKs React/RN — **SEMANAL 2**
- [ ] **Beta com 10-20 clientes Growth** @ R$1.099
- [ ] Publicar case study cliente Série A

### Q4 2026 (Out-Dez)
- [ ] COMM-010: Funis conversão
- [ ] COMM-011: Metric assignment
- [ ] COMM-012: Health score customizável
- [ ] **50+ clientes pagos**
- [ ] **Primeira venda self-hosted** (R$50k/ano)
- [ ] Auditar compliance LGPD

### 2027 (Opcional)
- [ ] COMM-013: Session recording (se demanda)
- [ ] COMM-014: Approval workflows
- [ ] COMM-015: Schema registry
- [ ] ISO 27001, SOC 2 Type II

---

## 💾 Backlog Template para Adicionar ao Repositório

**Para cada BKL, usar estrutura:**

```markdown
### BKL-{numero} | {Título}

- **Status:** To Do
- **Prioridade:** {CRITICAL/HIGH/MEDIUM/LOW}
- **Esforço:** {5/8/13/21 dias}
- **Sprint:** {Q2/Q3/Q4}
- **Descrição:** {Resumo 1-3 linhas}
- **Definição de Pronto:**
  - [ ] Task 1
  - [ ] Task 2
  - [ ] Task 3
- **Aceitação:** {Como cliente vai validar}
```

---

**Documento finalizado:** 10 de abril de 2026  
**Próximo review:** 20 de abril (após aprovação de pricing pelo CEO)
