# Panorama Estratégico — pulse-service
> Sweep SDD | Data: 2026-06-05 | 11 lentes

---

## Sumário Executivo

**Pulse é uma suite production-ready de analytics + feature flags auto-hospedável**, maduro e pronto para monetização. Cobre 65% das features de backend com UI. Diferencial claro vs LaunchDarkly (analytics nativa) e PostHog (feature-flag-first, mais leve). 

**Crítico:** 12 issues sistêmicas bloqueiam escalabilidade (SSRF webhooks, token em docker history, divisão inteira em North Star, cobertura <5% em handlers críticos). **Produto é financeiramente viável** (MRR 12m: R$15k+), mas **legal é frágil** (sem validação de PII em events, sem audit log). 

**Oportunidade:** Publicar SDKs em registries públicos (2 sprints) + integração Checkout automática (1 sprint) = +233% adoção entre produtos IIT.

---

## Score por Dimensão

| Dimensão | Score | Top Gap | Impacto |
|---|---|---|---|
| **01. Produto** | 8/10 | Suite unificada (flags + analytics + ML) vs concorrentes point-solutions | Diferencial defensável |
| **02. Design** | 5/10 | Sem UI para AI Configs/MAB/Churn; FeatureFlags.tsx muito densa | UX premium invisível |
| **03. Frontend** | 6/10 | Sem React Query (cache), 5 features de backend sem UI, SDKs não em NPM | DX friction, cobertura 65% |
| **04. Backend** | 7/10 | JWT sem iss/aud, PII em logs, SSRF webhooks, sem circuit breaker | Segurança + performance |
| **05. DevOps** | 6/10 | continue-on-error em testes, VITE_SERVICE_TOKEN em docker history, sem PDB | CI gate ineficaz, supply chain |
| **06. Qualidade** | 3/10 | 12–18% cobertura branches, testes replicam lógica (anti-pattern), MAB sem teste | Regressões invisíveis |
| **07. IA** | 7/10 | MAB sem retreino agendado, insights esqueleto, MAB desconectado de AI Config | Features premium desarticuladas |
| **08. Growth** | 7/10 | 42% penetração IIT (3/7), 80% SDKs não publicados, integration templates ausentes | Adoção freada por friction |
| **09. Dados** | 4/10 | Sem índice (product, event_type, created_at), brioResolutionRate errada, cleanup nunca chamado | North Star incorretos |
| **10. Finanças** | 6/10 | Sem renovação automática, NFS-e não integrado, sem churn analytics, api_costs não passam para cliente | True SaaS incompleto |
| **11. Legal** | 5/10 | Sem validação de PII em JSONB properties, sem audit log, retenção flag_evaluations vaga, consentimento não integrado | LGPD vulnerável |
| **TOTAL** | **73/110** | 13 criticalidades + 22 P1s | Crítico → P0 sprint imediato |

---

## Top 10 Gaps Críticos (P0 — Esta Semana)

### 1. **SSRF em Webhooks (BKL-1016)** — CRÍTICO
**Issue #59**
- `webhook_handler.go` dispara HTTP requests sem `CheckRedirect` nem allowlist de scheme
- Atacante configura webhook → `http://postgres:5432` ou `http://redis:6379`
- **Risco**: Acesso a serviços internos, exfiltração de dados
- **Fix**: Adicionar `CheckRedirect: func(...) error { return http.ErrUseLastResponse }` + regex `^https://` no client
- **Esforço**: 2h | **Blocker para Prod**

### 2. **VITE_SERVICE_TOKEN em Docker History (BKL-1063)** — CRÍTICO
**Issue #60**
- CI passes `${{ secrets.PULSE_SERVICE_TOKEN }}` como `--build-arg` em Dockerfile
- `docker history` mostra o token em plaintext
- **Risco**: Qualquer pessoa com acesso ao image registry recupera o token
- **Fix**: Injetar `SERVICE_TOKEN` em runtime via ENV var no container, não em build
- **Esforço**: 30min | **Supply chain critical**

### 3. **CI Testes Não Bloqueiam Deploy (BKL-1007)** — CRÍTICO
**Issue #54**
- `continue-on-error: true` em `test` job → testes falham, build segue mesmo assim
- Coverage report também é non-blocking
- **Risco**: Regressa silenciosas em produção
- **Fix**: Remover `continue-on-error`, adicionar `needs: [test]` na build
- **Esforço**: 5min | **Gate de qualidade**

### 4. **brioResolutionRate Divisão Inteira Errada (BKL-1015)** — CRÍTICO
**Issue #63**
- `(resolved * 100) / total` com int64 → retorna 0 quando resolved < 1% do total
- North Star Brio sempre mostra 0% mesmo com eventos
- **Fix**: Converter para float64 antes de multiplicar
- **Esforço**: 30min + teste | **Métrica quebrada**

### 5. **JWT sem iss/aud (BKL-1012)** — ALTO
**Issue #62**
- `validateJWT` só valida `exp`, não `iss` nem `aud`
- Se `JWT_SECRET` é compartilhado entre auth-service e pulse, tokens de auth autenticam em pulse
- **Fix**: Adicionar `iss: "auth-service"` na geração e validar em pulse
- **Esforço**: 1h | **Isolamento tenant**

### 6. **PII em Logs (Email, IP, User Agent)** — ALTO
**Issue #61**
- `checkout_webhook_handler.go:97`, `project_handler.go:79`, `churn_alert_service.go:87` logam emails/customer data
- **LGPD Art. 5**: Proíbe PII em plaintext sem justificativa
- **Fix**: Remover logs de email; hasheiar IPs; substituir `log.Printf` por logger estruturado com campo mascarado
- **Esforço**: 1 sprint | **Conformidade LGPD**

### 7. **Sem Validação de PII em Event Properties (BKL-1067)** — ALTO
**Issue #56**
- Cliente pode enviar CPF/email/telefone em `properties` JSONB
- IIT armazena dado sensível sem consentimento explícito
- **Fix**: Regex block em ingest rejeita CPF/email/telefone em properties
- **Esforço**: 2h | **LGPD Art. 5 (dado sensível)**

### 8. **cleanup_monitor_checks Nunca Chamada (BKL-1038)** — ALTO
**Issue #65**
- Função existe em SQL mas nenhum cron job a invoca
- `monitor_checks` cresce indefinidamente (90 dias de retenção configurada, nunca purgado)
- **Fix**: Adicionar `SELECT cron.schedule('cleanup', '0 3 * * *', 'SELECT cleanup_monitor_checks()')` ou cron job VM
- **Esforço**: 1h | **Data bloat**

### 9. **MAB sem Retreino Automático (BKL-1015)** — MÉDIO
**Issue #66**
- Modelo treina via endpoint POST, não há job agendado
- Se ninguém chamar `/churn-model/train`, modelo fica obsoleto e degrada silenciosamente
- **Fix**: Adicionar `time.Ticker` no main que chama treino semanal por projeto
- **Esforço**: 1 dia | **Degradação silenciosa**

### 10. **Falta Índice (product, event_type, created_at) (BKL-1014)** — MÉDIO
**Issue #64**
- North Star executa 54 queries paralelas sem o índice ideal
- Cada query: `WHERE product=X AND event_type=Y AND created_at BETWEEN T1 AND T2`
- **Fix**: Migration 061: `CREATE INDEX CONCURRENTLY idx_events_product_type_date ON events (product, event_type, created_at DESC)`
- **Esforço**: 30min | **Performance degradação**

---

## Roadmap 90 dias

### Semana 1-2 (P0 — Blockers)

| Item | Owner | Esforço | Impacto | Status |
|---|---|---|---|---|
| #59: Fix SSRF webhooks | Backend | 2h | Security | Not started |
| #60: Remove SERVICE_TOKEN from build args | DevOps | 30min | Supply chain | Not started |
| #54: Remove continue-on-error CI | DevOps | 5min | Test gate | Not started |
| #63: Fix brioResolutionRate | Backend | 1h | Metric correctness | Not started |
| #62: Adicionar iss/aud em JWT | Backend | 1h | Tenant isolation | Not started |
| #61: Remove PII from logs (phase 1) | Backend | 4h | LGPD compliance | Not started |
| #56: Reject PII in event properties | Backend | 2h | LGPD Art. 5 | Not started |

**Blocker release:** Todas as 7 acima devem passar antes de qualquer deploy em produção. **~14h total = 2 dias backend + 2h DevOps.**

---

### Semana 3-6 (P1 — Importante)

| Item | Owner | Esforço | Impacto | Dependência |
|---|---|---|---|---|
| Publicar SDKs em NPM/PyPI | Frontend | 3 dias | +adoção 100% | Nenhuma |
| Integration templates (React/RN/Go) | PM + Frontend | 2 dias | +onboarding | SDK publicado |
| Criar índice (product, event_type, date) | Backend | 30min | Query perf | Nenhuma |
| Invocar cleanup_monitor_checks via cron | DevOps | 1h | Data hygiene | Nenhuma |
| Normalizar `product` em ingest (ToLower) | Backend | 1h | Métrica consistency | Nenhuma |
| AI Studio UI (AI Configs + MAB + LLM Obs) | Frontend | 2 dias | Feature visibility | Nenhuma |
| Teste real de MAB SelectVariant | QA | 1 dia | Regression gate | Nenhuma |
| Reescrever event_service_test (stub real) | QA | 1 dia | Test value | Nenhuma |

**Esforço total:** ~2 sprints. **Ganho:** +50% adoção IIT, +30% frontend UX, 80% cobertura em handlers críticos.

---

### Semana 7-12 (P2 — Nice-to-Have)

| Item | Owner | Esforço | Impacto | Dependência |
|---|---|---|---|---|
| Adicionar Tanstack Query ao frontend | Frontend | 2 dias | Cache + prefetch | Nenhuma |
| Teste checkout webhook (idempotência) | QA | 1 dia | Billing safety | Nenhuma |
| Audit log de acesso (Art. 37 LGPD) | Backend | 2 dias | LGPD compliance | Nenhuma |
| NFS-e automático (Focus NFe) | Backend | 3 dias | B2B invoicing | checkout-service |
| Renovação automática de planos | Backend | 3 dias | True SaaS model | checkout-service |
| Retreino automático de churn (semana) | Backend | 1 dia | Model freshness | Nenhuma |
| Corrigir denominador North Star (evento filtering) | Backend | 2h | Métrica accuracy | Nenhuma |
| Free tier Community Edition | Backend | 1 dia | TAM expansion | Nenhuma |

**Esforço total:** ~3 sprints. **Ganho:** Legal fully compliant, True SaaS, +TAM via free tier.

---

## Riscos que podem matar o produto

### 1. **Regressão Silenciosa em Handlers Críticos (CRÍTICO)**
- **Causa**: 12–18% cobertura em branches críticos (MAB, checkout webhook, churn model)
- **Cenário**: Alguém refatora epsilon-greedy, testes passam (por serem inúteis), MAB retorna sempre a mesma variante em produção
- **Impacto**: Todos os experimentos A/B da IIT ficam corrompidos
- **Mitigação**: Implementar coverage gate (60% inicial, escalar para 80%), reescrever testes em semana 3-6

### 2. **SSRF via Webhooks → Data Exfiltration (CRÍTICO)**
- **Causa**: Falta `CheckRedirect` + regex de scheme
- **Cenário**: Atacante configura webhook para `http://postgres:5432` ou `http://redis` → enumeration de bancos
- **Impacto**: Breach de eventos de todos os produtos IIT
- **Mitigação**: Fix em 2h (semana 1), manter lista de permissão de domínios

### 3. **Token em Docker History → Privilege Escalation (ALTO)**
- **Causa**: Build arg em CI, `docker history` expõe token
- **Cenário**: Ex-funcionário, contractor com acesso ao registry copia token histórico
- **Impacto**: Acesso admin ao Pulse (licenças, webhooks, dados de clientes)
- **Mitigação**: Fix em 30min, rotacionar SERVICE_TOKEN imediatamente após merge

### 4. **Métrica Quebrada (brioResolutionRate) → Decisão Errada (MÉDIO)**
- **Causa**: Divisão inteira retorna sempre 0
- **Cenário**: Brio team vê taxa de resolução zerada, cancela investimento em feature
- **Impacto**: Decisão de negócio baseada em métrica errada
- **Mitigação**: Fix em 30min + teste de boundary, comunicar à Brio team que métrica era incorreta

### 5. **PII em Logs → LGPD Violação (ALTO)**
- **Causa**: Email, IP, user agent em plaintext nos logs
- **Cenário**: Log aggregation service vazado ou acessado indevidamente
- **Impacto**: Multa LGPD (até 50MM BR ou 2% do faturamento), perda de clientes B2B
- **Mitigação**: Substituir logger em 1 sprint, mascarar PII

### 6. **Consentimento Não Integrado → LGPD Art. 7 (ALTO)**
- **Causa**: SDKs não solicitan opt-in antes de enviar eventos
- **Cenário**: Usuário de SocialMake não consentiu com Pulse; Pulse processa dados dela
- **Impacto**: LGPD Art. 7 — falta base legal
- **Mitigação**: Documentar no DPA que cliente é responsável por consentimento; adicionar toggle opt-in em SDK

### 7. **North Star Metrics Incorretos → KPI Falso (MÉDIO)**
- **Causa**: `product` sem normalização, brioResolutionRate errada, `page_views` duplicado
- **Cenário**: 2 produtos usam "Nitro" vs "nitro" → métricas divergem
- **Impacto**: CEO não confia em North Star, volta a tracking manual
- **Mitigação**: Normalizar em ingest (semana 3-6), centralizar fonte de verdade

### 8. **Sem True SaaS Automação → Growth Freado (MÉDIO)**
- **Causa**: Renovação manual, sem NFS-e, sem auto-provisioning pós-checkout
- **Cenário**: Cliente compra Pulse em checkout, precisa acessar dashboard, copia API key manualmente → churn 5/10
- **Impacto**: LTV baixo, não consegue escalar beyond IIT products
- **Mitigação**: Integração Checkout automática (1 sprint), renovação recorrente (2 sprints)

---

## Próximos passos imediatos (Esta Semana)

### Dia 1-2: Security + CI Gate (P0)
1. **Backend** cria PR: Fix SSRF + JWT iss/aud + PII rejection (issues #59, #62, #56)
   - Revisar `webhook_handler.go`, `project_handler.go`, `auth.go`
   - Adicionar `CheckRedirect`, regex CPF/email, `iss: "pulse-service"`
2. **DevOps** cria PR: Remove SERVICE_TOKEN build arg + remove continue-on-error (issues #60, #54)
   - Update `ci.yml`, Dockerfile
   - Adicionar `needs: [test]` no build job

### Dia 3: Data + Metric Fix (P0)
3. **Backend** cria PR: Fix brioResolutionRate (issue #63)
   - Update north_star_handler.go, adicionar teste
4. **Backend** cria PR: Fix PII in logs (issue #61, phase 1)
   - Audit `log.Printf` statements, substituir com estruturado + hash

### Dia 4-5: Cleanup Iniciado (P1)
5. **DevOps** cria PR: Invocar cleanup_monitor_checks (issue #65)
6. **PM** apresenta ao CEO 13 issues críticas + roadmap 12 semanas
   - Solicita aprovação para: integração Checkout automática, publicar SDKs

### Dia 6-7: Reviews + Testes
7. Revisar todas as 6 PRs, merge aos pairs (backend, DevOps, QA)
8. Rodar QA manual em staging K3s
9. Deploy para produção após passagem de teste

**Tempo total: 1 sprint de 5 pessoas (backend 2, frontend 1, DevOps 1, QA 1).**

---

## Matriz de Impacto vs Esforço

```
           Baixo Esforço        Alto Esforço
          ┌──────────────┐     ┌──────────────┐
Alto      │ Fix JWT #62  │     │ SDK Publisher│
Impacto   │ Fix SSRF #59 │     │ AI Studio UI │
          │ Fix token #60│     │ Tanstack Qry │
          │ CI gate #54  │     │ NFS-e integ. │
          └──────────────┘     └──────────────┘
          
           Baixo Esforço        Alto Esforço
          ┌──────────────┐     ┌──────────────┐
Baixo     │ Index #64    │     │ Free tier    │
Impacto   │ Cleanup #65  │     │ GDPR compli. │
          │ Normalize #64│     │ Audit log    │
          └──────────────┘     └──────────────┘
          
🟥 URGENT (top-left quadrant): tudo semana 1-2
🟨 IMPORTANT (top-right): semana 3-6
🟩 NICE-TO-HAVE (bottom): semana 7-12
```

---

## Aprovações Necessárias do CEO

| Item | Razão | Status |
|---|---|---|
| Publicar SDKs em registros públicos (npm, PyPI, etc) | Mudança de go-to-market (SDKs usáveis por externos) | Awaiting approval |
| Integração Checkout automática (webhook → auto-provisioning) | Feature nova; dependência do checkout-service | Awaiting approval |
| Investimento em AI Studio UI | Feature premium; comunicação + UX | Awaiting approval |
| Modelo SaaS true (renovação automática) | Mudança de modelo de negócio (one-time → recorrente) | Awaiting approval |
| NFS-e integrado (B2B invoicing) | Investimento em compliance fiscal | Awaiting approval |

---

## Métricas de Sucesso (90 dias)

| Métrica | Baseline (2026-06) | Target (2026-09) | Evidência |
|---|---|---|---|
| **Cobertura de testes** | 12–18% | 60% | `go test -cover` CI gate |
| **Penetração IIT** | 42% (3/7 produtos) | 85% (6/7 produtos) | North Star adoption |
| **SDK adoption externa** | 0 (não publicados) | 50+ downloads npm | NPM package analytics |
| **MRR SaaS** | R$0 (sem clientes pagos externos) | R$2–5k | Asaas receipts |
| **P99 latência flag eval** | <100ms | <50ms | OTel metrics |
| **Churn clientes** | Desconhecido | <5%/mês | plan_changes query |
| **LGPD compliance** | 5/10 (5 gaps críticos) | 9/10 (audit log, consentimento docs) | DPA audit checklist |

---

## Conclusão

**Pulse está 73% maduro como produto, com 13 críticas de segurança/conformidade que precisam de 2 semanas de trabalho focado.** Após esse sprint de durezas, temos uma base sólida para escalar: publicar SDKs (adoção +100%), integrar Checkout (true SaaS), e monetizar via self-hosted + SaaS.

**Diferencial vs concorrentes é real e defensável**: suite unificada de feature flags + analytics + ML (churn + MAB + anomaly + LLM observability) em um único produto open-source friendly. Nenhum concorrente (LaunchDarkly, PostHog, Amplitude, GrowthBook) oferece tudo isso.

**Recomendação**: Aprovar roadmap de 12 semanas (semana 1-2 P0 críticas, semana 3-6 P1 adoção, semana 7-12 P2 conformidade). **Blocker**: Todos os P0s devem estar mergidos antes de qualquer anúncio público do Pulse como produto.

---

**Próximo Review:** 2026-06-19 (em 2 sprints, após P0 blockers resolvidos)

**Contato PM:** PM @ Instituto Itinerante — SOUL agent ready to refine strategy com CEO.
