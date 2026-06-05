# Lentes 04 e 05 — Backend e DevOps
**pulse-service | SDD Sweep 2026-06-05**

---

## Lente 04 — Backend

**Score: 7/10**

### Segurança

**Pontos fortes:**

- `ServiceTokenMiddleware` usa `crypto/subtle.ConstantTimeCompare` — imune a timing attack (BKL-334 resolvido).
- JWT HS256 implementado manualmente sem dependência externa: verifica assinatura com `hmac.Equal` e valida `exp`. Não usa `==` para comparação de strings criptográficas.
- `ValidateProjectKey` aplica isolamento cross-tenant: resolve `project_id` via `X-Project-Key` antes de qualquer handler de ingestão, com prepared statement (`$1`).
- Webhooks saintes assinados com HMAC-SHA256 (`crypto/rand` para segredo, `crypto/hmac` para payload) — compatível com padrão Stripe/GitHub.
- `SecurityHeaders()` cobre HSTS, CSP, X-Frame-Options, X-Content-Type-Options, Referrer-Policy.

**Gaps de segurança:**

1. **JWT sem claims obrigatórios além de `exp`:** `validateJWT` não verifica `iss`, `aud` nem `sub`. Um token emitido por qualquer issuer com o mesmo `JWT_SECRET` e um `exp` válido é aceito. Risco: se `JWT_SECRET` for compartilhado entre serviços (e é — `***REDACTED-SEE-TRUENAS-credentials.txt***` é o mesmo para auth/my-library/focus-hub), tokens de outros serviços autenticam no Pulse.

2. **`checkoutPayload.CustomerEmail` logado implicitamente via contexto de erro:** `checkout_webhook_handler.go:97` loga ausência do email; `project_handler.go:79` loga `req.OwnerEmail` diretamente em erro de `GetPlanByEmail`. Emails de clientes em logs são dados pessoais (LGPD Art. 5).

3. **`email_sequence_service.go:91`** loga `proj.ContactEmail` em falha de envio — PII em plaintext no log.

4. **`churn_alert_service.go:87`** loga `adminEmail` — baixo risco mas inconsistente com R5.

5. **Ausência de validação de tamanho de body:** `checkout_webhook_handler.go` e `webhook_handler.go` não limitam `c.Body()` antes do `BodyParser`. Um payload de 100 MB causa alocação na heap sem rejeição antecipada. Mitigação parcial: Traefik pode ter limite global.

6. **Sem OWASP A04 (Insecure Design) para webhooks saintes:** `WebhookHandler.httpClient` tem `Timeout: 10s` mas não limita redirects (`CheckRedirect` não configurado). Um SSRF via URL de webhook aponta para serviços internos (ex: `http://postgres:5432`).

### Performance

**OTel:** Ativo. `telemetry.Init("pulse-service")` configura `TracerProvider` com exporter OTLP gRPC; fallback para no-op quando `OTEL_EXPORTER_OTLP_ENDPOINT` não está definido. `otelfiber` instrumenta todas as rotas. Implementação correta.

**pgxpool:** `pgxpool.New(ctx, cfg.DatabaseURL)` sem `pgxpool.Config` customizado — usa defaults do pgx: `MaxConns = max(4, NumCPU)`, sem `MinConns`, sem `MaxConnLifetime` explícito. Para 50+ handlers num serviço de analytics/ML com queries potencialmente longas (Bayesian, MAB), o pool padrão pode esgotar sob carga. Falta `pgxpool.ParseConfig` + ajuste de `MaxConns`, `MinConns`, `MaxConnLifetime`, `HealthCheckPeriod`.

**Redis:** Usado corretamente para rate limiting (sliding window com INCR + EXPIRE NX), quota mensal (pipeline Incr + ExpireNX), e cache de feature flags. Namespacing consistente (`rl:pulse:*`, `quota:pulse:*`). Fail-open documentado explicitamente nos dois middlewares.

**Gaps de performance:**

1. **Sem circuit breaker para chamadas HTTP saintes:** `WebhookHandler` faz requisições para URLs externas sem padrão de circuit breaker ou retry com backoff exponencial. Falhas em cascata de webhooks exaurem goroutines.
2. **Context propagation parcial:** `ValidateProjectKey` usa `c.Context()` (context do Fiber/fasthttp), não propaga trace context do OTel para as queries de lookup de API key. Spans de banco ficam fora da árvore de trace da requisição.
3. **`log.Printf` stdlib em vez de logger estruturado:** Todo o serviço usa `log.Printf` padrão. Sem campos estruturados (`trace_id`, `project_id`, `request_id`), correlação de logs com traces é manual e sujeita a erro.

### Gaps adicionais

- Sem `recover` middleware documentado no scan — verificar se `fiber/v2/middleware/recover` está aplicado globalmente (main.go importa mas não foi confirmado no setup do app).
- Sem rate limiting para rotas admin (service token). Um token comprometido pode fazer flood de operações administrativas sem throttle.
- JWT sem rotação de chave: `PULSE_JWT_SECRET` é estático. Sem `kid` (key ID) no header, não há caminho para rotação sem downtime.

---

## Lente 05 — DevOps

**Score: 6/10**

### K8s Manifests

| Componente | Status | Detalhe |
|---|---|---|
| Imagem com tag SHA | OK | `ghcr.io/apsferreira/pulse-service:sha-d19a051` — sem `:latest` no deployment backend |
| Imagem frontend | RISCO | `ci.yml` publica tag `latest` para frontend: `type=raw,value=latest` — frontend-deployment.yaml pode estar usando `:latest` |
| Resource limits | OK | `requests: cpu=100m, mem=128Mi` / `limits: cpu=200m, mem=256Mi` — configurados |
| Liveness/Readiness | OK | Ambos em `/health:3050` com delays adequados |
| HPA | OK | `minReplicas: 2`, `maxReplicas: 4`, CPU 70%, Memory 80% — autoscaling/v2 correto |
| PDB (PodDisruptionBudget) | AUSENTE | Nenhum arquivo de PDB no `deploy/k8s/`. Com `minReplicas: 2`, uma drenagem de nó pode derrubar 100% dos pods |
| NetworkPolicy | AUSENTE | Sem restrição de egress/ingress a nível de pod. Qualquer pod no namespace `production` pode alcançar o pulse-service e vice-versa |
| Secrets | OK | Todos via `secretKeyRef` (database-url, redis-url, service-token, jwt-secret) — sem hardcode |
| ArgoCD | OK | `automated.prune: true`, `selfHeal: true` — GitOps funcional. `CreateNamespace: false` previne criação acidental |

### CI/CD

**Gaps críticos:**

1. **`continue-on-error: true` no job `test` (linha 17):** O job de testes inteiro é non-blocking. Build e deploy ocorrem mesmo com todos os testes falhando. Isso torna o pipeline de teste decorativo — não há gate de qualidade.

2. **`continue-on-error: true` no step `Coverage report` (linha 78):** Coverage report falha silenciosamente. Não há gate de cobertura mínima — qualquer regressão de 80% para 0% passa sem alarme.

3. **Sem `gosec`:** Nenhuma etapa de análise de segurança estática Go no pipeline. Requisito da Definição de Pronto não atendido.

4. **Sem `govulncheck`:** Nenhuma verificação de vulnerabilidades conhecidas em dependências.

5. **Sem `golangci-lint`:** Linter não roda no CI — inconsistente com Definição de Pronto do SOUL.

6. **`VITE_SERVICE_TOKEN` exposto como build-arg:** `ci.yml:188` passa `${{ secrets.PULSE_SERVICE_TOKEN }}` como `--build-arg` do Docker. Build args aparecem em `docker history` e no cache do BuildKit. Para tokens, o correto é injetar em runtime via env var no container, não em build time.

7. **Self-hosted runner `iit-pulse`:** Jobs rodam em runner próprio. Sem isolamento de job (mesma máquina para test, build e push) — um step malicioso tem acesso ao socket Docker e às credenciais do runner. Aceitável para homelab, mas documentar o risco.

### Top 5 Ações Prioritárias

**P0 — Bloqueadores imediatos:**

1. **Remover `continue-on-error: true` do job `test`** e adicionar `needs: [test]` no `build-and-push`. Testes devem ser gate real de deploy. Estimativa: 5 min.

2. **Adicionar PodDisruptionBudget** com `minAvailable: 1` para garantir que drenagens de nó não derrubem todos os pods simultaneamente. Estimativa: 10 min.

**P1 — Segurança:**

3. **Adicionar `gosec` e `govulncheck` ao CI** como steps obrigatórios (não `continue-on-error`). Integrar antes do build. Estimativa: 30 min.

4. **Corrigir SSRF em webhooks saintes:** Configurar `CheckRedirect` no `httpClient` para rejeitar redirects e adicionar allowlist de esquemas (`https://` apenas) antes de disparar. Estimativa: 2h.

**P2 — Observabilidade e Qualidade:**

5. **Substituir `log.Printf` por logger estruturado** (ex: `log/slog` stdlib Go 1.21+) com campos `trace_id` propagados do contexto OTel. Remove PII acidental de emails em logs e habilita correlação automática com traces. Estimativa: 1 sprint.

---

**Arquivos auditados:**
- `/Users/antonio.ferreira/projects/pulse-service/internal/handler/webhook_handler.go`
- `/Users/antonio.ferreira/projects/pulse-service/internal/handler/checkout_webhook_handler.go`
- `/Users/antonio.ferreira/projects/pulse-service/internal/middleware/auth.go`
- `/Users/antonio.ferreira/projects/pulse-service/internal/middleware/ratelimit.go`
- `/Users/antonio.ferreira/projects/pulse-service/internal/middleware/project_key.go`
- `/Users/antonio.ferreira/projects/pulse-service/internal/middleware/quota.go`
- `/Users/antonio.ferreira/projects/pulse-service/internal/middleware/security_headers.go`
- `/Users/antonio.ferreira/projects/pulse-service/deploy/k8s/deployment.yaml`
- `/Users/antonio.ferreira/projects/pulse-service/deploy/k8s/hpa.yaml`
- `/Users/antonio.ferreira/projects/pulse-service/deploy/k8s/ingress.yaml`
- `/Users/antonio.ferreira/projects/pulse-service/deploy/argocd/application.yaml`
- `/Users/antonio.ferreira/projects/pulse-service/.github/workflows/ci.yml`
- `/Users/antonio.ferreira/projects/pulse-service/cmd/server/main.go` (parcial)
