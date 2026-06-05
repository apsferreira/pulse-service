# SEÇÃO 10-11: LENTES FINANCEIRAS E LEGAIS DO PULSE
**Análise SDD 2026-06-05**

---

## LENTE 10 — FINANÇAS (Score: 6/10)

### Modelo de Billing Vigente

O pulse-service implementa um modelo de billing multi-camadas:

#### 1. **Planos por Conta (tabela `accounts`)**
- **Free**: 10.000 eventos/mês, 5 feature flags, 1 projeto
- **Starter**: 500.000 eventos/mês, flags ilimitadas, 3 projetos — R$399/mês
- **Growth**: 2.000.000 eventos/mês, flags ilimitadas, 10 projetos — R$1.099/mês
- **Pro**: Eventos ilimitados, flags ilimitadas, projetos ilimitados — R$2.399/mês
- **Enterprise/Internal**: Idem Pro, sem limite
- **Retenção de dados**: Free/Starter/Growth (90-180 dias), Pro+ (365 dias)

Fonte: `migrations/056_accounts_plan.sql`, `upgrade_gate.go` (linhas 15-25)

#### 2. **Overage Billing (BKL-1071)**
- **Mecanismo**: Projetos em plano Starter/Growth que ultrapassam quota mensal são cobrados por overage
- **Preço**: R$0,02 por 1.000 eventos acima do limite
- **Fluxo**:
  1. Middleware de quota (`quota_enforcement`) conta eventos reais em Redis
  2. Plano free: retorna HTTP 429 quando atinge 10k (bloqueio duro)
  3. Planos pagos: eventos acima da quota são registrados em `overage_billing`
  4. Cron job (`/api/v1/overage/calculate`) calcula overage no final do mês
  5. Status fluxo: `pending` → `invoiced` → `paid` (integração com Asaas via `external_charge_id`)

Tabela: `overage_billing` (migration 036)
- Campos críticos: `project_id`, `period_year`, `period_month`, `events_consumed`, `overage_events`, `amount_brl`, `status`, `external_charge_id`
- Constraint: UNIQUE(project_id, period_year, period_month) — idempotente

Fonte: `overage_handler.go`, `migrations/036_create_overage_billing.sql`

#### 3. **Licenças Self-Hosted (BKL-1031)**
- Modelo: SaaS multilocatário (cloud) + self-hosted (on-premises)
- **Geração de licença**: Webhook de checkout gera `license_key` com validade de 1 ano
- **Validação**: Pulse self-hosted valida `/api/v1/license/validate` na inicialização
- **Revogação**: Suportada (campo `revoked_at`)
- **Dados capturados**: `customer_email`, `customer_name`, `order_id`, `valid_until`, `hostname`

Fonte: `license_handler.go`

#### 4. **API Costs Tracking (BKL-1021)**
- **Tabela**: `api_costs` (migration 010)
- **O que rastreia**: LLM calls (ChatGPT, Claude), image generation, outros provedores
- **Campos**: `product`, `operation`, `provider`, `model`, `tokens_in`, `tokens_out`, `images`, `cost_usd`, `metadata`
- **Propósito**: Entender COGS do Pulse quando oferecido como SaaS (cada flag evaluation que usa LLM incorre custo)
- **Status**: Tabela criada, mas sem integração visível com billing de clientes

Fonte: `migrations/010_create_api_costs.sql`

#### 5. **Fluxo de Upgrade**
- Endpoint `/api/v1/upgrade/recommend` analisa uso real
- Triggers: flags_used > 5, events_last_30d > limite, projeto próximo de quota
- Retorna payload com `call_to_action` + `upgrade_url` para checkout-service
- URLs direcionam para: `checkout-api.institutoitinerante.com.br/checkout?plan=pulse-{starter|growth|scale}`

Fonte: `upgrade_handler.go`, `upgrade_gate.go`

---

### MRR Potencial do Pulse como Produto Standalone

#### Cenário Base (projeção 12 meses)

**Premissas:**
- Pulse já está implantado internamente em 6 produtos IIT
- Auto-faturamento habilitado para novos clientes externos (via checkout-service)
- Não há marketing agressivo — growth orgânico via network IIT

| Métrica | Free | Starter | Growth | Pro | Total MRR |
|---------|------|---------|--------|-----|-----------|
| **Clientes (mês 1)** | 8 | 2 | 1 | 0 | — |
| **MRR Base (mês 1)** | R$0 | R$798 | R$1.099 | R$0 | **R$1.897** |
| **Clientes (mês 6)** | 25 | 5 | 2 | 1 | — |
| **MRR (mês 6)** | R$0 | R$1.995 | R$2.198 | R$2.399 | **R$6.592** |
| **Clientes (mês 12)** | 45 | 12 | 5 | 2 | — |
| **MRR (mês 12)** | R$0 | R$4.788 | R$5.495 | R$4.798 | **R$15.081** |

**Overage (estimado, sem canibalizações):**
- Growth tier: ~5-8 projetos/mês excedendo quota → ~2 x R$50-200 = ~R$300/mês
- **MRR com Overage (mês 12)**: R$15.381

**Observação**: SocialMake, Nitro, Librinho, Brio, MeuIngresso já consomem Pulse internamente. Não geram MRR direto a IIT (são "clientes internos" com plano `internal`). Se vendesse Pulse como produto para terceiros, o modelo acima seria realista.

---

### Gaps e Vulnerabilidades Financeiras

#### 1. **Integração com checkout-service é parcial**
- ✅ Webhook recebe notificação de pagamento aprovado
- ✅ Licença é gerada automaticamente
- ❌ **Planos mensais não têm renovação automática** — não há crédito recorrente no checkout
- ❌ **Cancelamento e downgrade** não atualizam `accounts.plan` automaticamente
- ❌ **Reconciliação**: nenhuma tabela de `subscriptions` ou `billing_subscriptions` vinculando account → checkout order

**Ação recomendada**: Criar tabela `pulse_subscriptions(account_email, plan, checkout_subscription_id, period_start, period_end, auto_renew, status)` e webhook de renovação.

#### 2. **NFS-e não está integrado**
- `overage_billing` gera `amount_brl` mas não gera NFS-e automática
- Focus NFe (R$109/mês) ou NFE.io (R$119/mês) não estão configurados
- **Gap**: Empresas B2B clientes do Pulse exigem NFS-e — fatura em 30 dias, recebimento em 60+ dias = cash flow problem
- **LGPD context**: NFS-e em Salvador mudou para padrão Nacional (IN 08/2025) em dez/2025

**Ação recomendada**: Integrar Focus NFe + webhook → gerar NFS-e quando `overage_billing.status = 'invoiced'`

#### 3. **API Costs não fluem para preço do cliente**
- IIT rastreia `api_costs` (LLM, imagens) em `api_costs` table
- Mas **não há modelo de passthrough** de custo para clientes que usam features AI no Pulse
- **Cenário**: Um cliente Growth pagando R$1.099 usa 1M tokens/mês em LLM (custo IIT ~R$20) sem markup

**Ação recomendada**: 
- Adicionar campo `api_cost_threshold` em `accounts` (ex: R$50/mês incluído)
- Se `monthly_api_cost > threshold`, cobrar R$0,01 por R$1 excedido
- Ativar quando features AI em flag evaluations forem GA

#### 4. **Sem captura de COGS real**
- Pulse roda em K3s homelab (custo infraestrutura = ~R$400/mês amortizado)
- Banda de Cloudflare CDN (SCC/DPA aplicados) = custo desconhecido
- **Gross margin estimada**: ~85% (alto, porque infra é própria)
- **Sem tracking**: não há `monthly_revenue` vs `monthly_cogs` dashboard

**Ação recomendada**: Gerar relatório mensal `MRR vs COGS vs Gross Margin` no pulso dashboard (admin-only)

#### 5. **Sem analitics de churn**
- Tabela `plan_changes` registra transições (`from_plan` → `to_plan`)
- Mas **não há cálculo de churn** (free → null, starter → free, etc.)
- **Risco**: Não saber se clientes downgradam por preço, feature gaps ou problema técnico

**Ação recomendada**: Query mensal: `SELECT COUNT(*) FROM plan_changes WHERE created_at >= now() - interval '30 days' AND from_plan IN ('starter', 'growth') AND to_plan IN ('free', null)` → medir churn rate

---

## LENTE 11 — LEGAL (Score: 5/10)

### Conformidade LGPD

#### Status de Implementação

| Artigo LGPD | Requirement | Status | Detalhe |
|---|---|---|---|
| Art. 7° VI | Contrato válido como base legal | ✅ | DPA_PULSE_TEMPLATE.md criado (BKL-1033) |
| Art. 7° IX | Legítimo interesse (analytics) | ✅ | Documentado no DPA Sec. 2 |
| Art. 18 V | Direito de acesso | ✅ | `GET /api/v1/users/:user_id/export` (BKL-1036) |
| Art. 18 VI | Direito ao esquecimento | ✅ | `DELETE /api/v1/users/:user_id/data` (BKL-1036) |
| Art. 37 | Segurança e boas práticas | ⚠️ | LUKS em homologação, não em produção |
| Art. 46 | Medidas de segurança | ⚠️ | Incompleto (ver abaixo) |
| Art. 48 | Notificação de incidente (72h) | ❌ | Sem processo formal definido |
| Art. 50 | Impacto de segurança notificado | ❌ | Sem template de comunicação |

Fonte: `user_export_handler.go`, `docs/legal/DPA_PULSE_TEMPLATE.md`

#### Eventos Coletados — Dados Pessoais Transferidos

Tabela `events` (migration 004):
```sql
- user_id (UUID) — identificador direto do usuário ✅ PII
- session_id (VARCHAR) — correlação de comportamento ✅ quasi-PII
- properties (JSONB) — pode conter PII se cliente não filtrar ⚠️
- ip_address (INET) — geolocalização + identificação ✅ PII
- user_agent (TEXT) — fingerprinting + device detection ✅ quasi-PII
- created_at (TIMESTAMPTZ) — padrão necessário
```

**Problema Legal**: Não há validação de schema para `properties` JSONB. Cliente pode enviar:
```json
{
  "user_id": "u-123",
  "properties": {
    "cpf": "123.456.789-00",
    "phone": "+55 71 98765-4321",
    "home_address": "Rua X, Salvador"
  }
}
```

**Análise de Risco**:
- Se cliente envia CPF/email/telefone em `properties`, IIT está armazenando dado sensível **sem consentimento explícito**
- Art. 5° II LGPD proíbe processamento de dados sensíveis sem consentimento + justificativa
- **IIT não pode alegar ignorância** — é responsável por implementar guarda-chuva tecnológico

**Ação recomendada**: 
- Adicionar regex block em ingest para rejeitar CPF, email, telefone em `properties`
- Truncar `user_agent` para primeiros 50 chars (reduz fingerprinting)
- Documentar no SDK guide: "NÃO envie CPF, email, telefone, endereço em event properties"

---

### Retenção de Dados e Purge Automático

#### Política Atual
```
Free:           90 dias    (data_retention_days = 90)
Starter:        180 dias   (data_retention_days = 180)
Growth:         180 dias   (data_retention_days = 180)
Pro+:           365 dias   (data_retention_days = 365)
Internal (IIT): 365 dias
```

Fonte: `migrations/056_accounts_plan.sql` (linha 50), `upgrade_handler.go` (planLimits)

#### Purge Automático (BKL-1037)
- **Job**: `retentionpurge/purger.go` — rodado 1x/dia via cron
- **Lógica**: `DELETE FROM events WHERE project_id IN (...) AND created_at < NOW() - INTERVAL '${data_retention_days} days'`
- **Status**: ✅ Implementado e testado

#### Gaps Legais

1. **Flag evaluations não têm purge explícito**
   - Tabela `flag_evaluations` armazena: `project_id`, `flag_id`, `user_id`, `context` (JSONB), `result` (boolean), `created_at`
   - **Problema**: Não há política de retenção — historicamente pode acumular anos
   - **LGPD**: Art. 15 exige "retenção não mais longa que o necessário"
   - **Ação**: Aplicar `data_retention_days` também a `flag_evaluations` no purge job

2. **Dados de conta (administrador email) retidos 5 anos**
   - Doc DPA Sec. 4: "Email do administrador: Duração do contrato + 5 anos (obrigação legal fiscal)"
   - **Risco**: Se cliente solicita exclusão (Art. 18 VI), há conflito entre LGPD (direito ao esquecimento) e fiscal (obrigação de retenção)
   - **Ação**: Criar log separado (`audit_log` com hash do email, sem email em claro) para fiscal; limpar email real após 30 dias de cancelamento

3. **Nenhum sistema de auditoria de acesso** (Art. 37)
   - `api_costs` table existe, mas **não há `access_log`**
   - Quando alguém chama `GET /api/v1/users/:user_id/export`, ninguém registra quem acessou
   - **LGPD Art. 37**: "A autoridade nacional poderá solicitar relatório de operação"
   - **Ação**: Criar `access_audit_log(actor_email, action, resource_type, resource_id, timestamp)` e manter 2 anos

---

### Transferências Internacionais e SCC (Cláusulas Contratuais Padrão)

#### Subprocessadores Listados no DPA

| Processador | País | Problema Legal |
|---|---|---|
| Cloudflare Inc. | EUA | ✅ DPA + SCC aprovadas (Comissão Europeia) |
| Resend Inc. | EUA | ⚠️ Verificar DPA em https://resend.com/legal |
| Redis Labs | EUA | ⚠️ Sem menção a DPA em https://redis.com |
| Hetzner Online | DE | ✅ Europeia, GDPR nativa |
| Homelab IIT | BR | ✅ On-premise Brasil |

**Gap**: DPA template menciona "SCC aprovadas pela Comissão Europeia" mas:
- ❌ Não há links aos SCCs específicos
- ❌ Pulse não oferece ferramentas para cliente executar SCC próprio (consentimento)
- ❌ Transferência via Cloudflare CDN não é documentada em ToS do cliente

**Ação recomendada**:
- Adicionar apêndice "Transferências de Dados Internacionais" ao DPA
- Se cliente é europeu (GDPR), oferecer opção de "EU-only" (sem CDN Cloudflare, maior latência)
- Manter registro de assinatura de SCC por cliente

---

### Consentimento e Consentimento Prévio

#### Lacuna Crítica: Nenhuma ferramenta de consentimento integrada

**Cenário**:
1. SocialMake (produto IIT) coleta eventos de usuários finais (creators, followers)
2. SocialMake chama `/api/v1/events` do Pulse com user_id + properties
3. **Pergunta**: O usuário final consentiu com Pulse? Ou só com SocialMake?

**LGPD Art. 7°**: Base legal exige consentimento **explícito** para finalidade específica. "Analytics via Pulse" é finalidade distinta de "publicar post no Instagram".

**Status de Implementação**:
- ❌ Nenhuma cookie banner ou pop-up no frontend Pulse
- ❌ Nenhuma integração com `customer-service` para verificar consentimento
- ❌ SDKs (JS, React Native) não solicitam opt-in antes de enviar eventos

**Ação recomendada**:
- Se Pulse é oferecido como SaaS público, **cliente é responsável** por consentimento (DPA está claro nisso — Sec. 10)
- Se IIT coleta eventos dos próprios produtos (SocialMake, Nitro, etc.), **IIT é responsável** por consentimento → adicionar toggle `analytics_consent` em account settings, solicitar opt-in em mobile apps
- Documentar responsabilidade no DPA: "Controlador garante ter base legal"

---

### Conformidade GDPR (se tiver clientes EU/EEA)

#### Status Atual
- **Pulse já roda em K3s Brasil** (homelab)
- **Cloudflare CDN** está entre cliente EU e Pulse
- **DPA template** menciona "Regulamento (UE) 2016/679 (GDPR), quando aplicável"

#### Gaps GDPR

1. **Sem opção de "residência de dados"** — se cliente alemão exige dados em Hetzner (EU), Pulse não oferece
2. **Sem "direito a ser esquecido" automático** — `DELETE /api/v1/users/:user_id/data` existe, mas é manual (cliente chama API)
3. **Sem relatório de AIPD** (Avaliação de Impacto de Proteção de Dados) — exigido se processamento tem alto risco

**Ação recomendada**:
- Documentar GDPR compliance como "Aplica-se apenas se cliente é EU/EEA"
- Redirecionar para lawyer se cliente é GDPR-subject
- Ou: implementar flag `gdpr_jurisdiction=true` → força purge após 30 dias (sem retenção longa)

---

### Dados de Pagamento e PCI-DSS

#### Gap: Pulse não processa cartão de crédito

**Status**:
- ✅ Pulse gera licença após webhook do checkout-service
- ✅ checkout-service integra com Asaas (payment processor)
- ❌ **Pulse não vê dados de cartão** — Asaas é responsável por PCI-DSS

**Risco residual**: 
- `overage_billing.external_charge_id` armazena ID da cobrança no Asaas
- Se Asaas sofre breach, clientes do Pulse podem ser impactados indiretamente

**Ação recomendada**: 
- Manter contato com Asaas sobre compliance PCI-DSS
- Documentar em ToS: "Pulse não processa dados de pagamento; Asaas é responsável"

---

## Resumo de Scores

### Lente 10 — Finanças: 6/10

**Justificativa**:
- ✅ Modelo de billing está funcional (free + 3 planos pagos + overage)
- ✅ Integração com checkout-service para geração de licença
- ✅ Rastreamento de API costs para futuro markUp
- ❌ Sem renovação automática de planos (não é SaaS verdadeiro ainda)
- ❌ Sem NFS-e integrado (gap fiscal para B2B)
- ❌ Sem analitics de churn
- ❌ Sem passthrough de api_costs para clientes

**Upgrade para 8/10**: Implementar renovação automática + NFS-e + churn metrics

### Lente 11 — Legal: 5/10

**Justificativa**:
- ✅ DPA template completo (BKL-1033)
- ✅ Direitos de acesso/exclusão implementados (BKL-1036)
- ✅ Purge automático de eventos (BKL-1037)
- ⚠️ Retenção de `flag_evaluations` sem política explícita
- ❌ Sem validação de schema para JSONB `properties` (risco de PII)
- ❌ Sem audit log de quem acessa dados
- ❌ Sem consentimento integrado (SaaS público)
- ❌ Sem GDPR compliance claro (se houver clientes EU)

**Upgrade para 8/10**: 
1. Validação de schema `properties` (reject CPF/email)
2. Audit log de acesso
3. Estender retenção a `flag_evaluations`
4. Clarificar responsabilidade de consentimento no DPA

---

## Próximas Ações (Prioridade)

### P0 — Críticas para FY2026
1. **Validação de PII em `properties`** — evita breach de CPF/email
2. **Renovação automática de planos** — true SaaS model
3. **Audit log de acesso** — compliance Art. 37

### P1 — Importantes Q3 2026
4. **NFS-e automático** — faturamento B2B legal
5. **Retenção para `flag_evaluations`** — completar LGPD Art. 15
6. **Churn analytics** — entender produto-market fit

### P2 — Nice-to-Have
7. API cost passthrough (quando AI em flags for GA)
8. GDPR compliance clarity (se contratos GDPR assinados)
9. Reconciliação checkout ↔ Pulse subscriptions

---

## Referências

- BKL-1033: Data Processing Agreement (DPA)
- BKL-1036: User data export/delete (LGPD)
- BKL-1037: Retention purge (automated)
- BKL-1071: Overage billing
- BKL-1031: License generation (self-hosted)
- BKL-1021: API costs tracking

Documentação: `/docs/legal/DPA_PULSE_TEMPLATE.md`, handlers em `/internal/handler/`
