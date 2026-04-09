# Pulse — Planos de Preços

## Visão Geral

Pulse é um serviço de feature flags e event tracking para aplicações na escala do Instituto Itinerante. Oferecemos 3 planos: **Free**, **Growth** e **Pro**, calibrados para startups brasileiras e empresas em escala.

---

## Planos

### Free
**Sem custo | Sem cartão de crédito**

Ideal para protótipos, MVPs e pequenas aplicações.

- **Eventos/mês:** até 10k
- **Feature flags:** até 5
- **Projetos:** 1
- **Usuários:** 1
- **A/B testing:** não
- **Histórico de eventos:** 7 dias
- **SLA:** Best effort
- **Suporte:** Community (Discord/GitHub Issues)

**Limite:** 1 projeto, 5 flags, 10k eventos/mês.

---

### Growth
**R$ 99/mês | Faturamento mensal**

Para startups e equipes em crescimento.

- **Eventos/mês:** até 500k
- **Feature flags:** ilimitadas
- **Projetos:** até 5
- **Usuários:** até 5
- **A/B testing:** sim
- **Histórico de eventos:** 30 dias
- **Rollout percentual:** sim
- **Overrides:** por usuário/tenant
- **Eventos por segundo:** ~5.8 (500k/mês ÷ 30 dias ÷ 86.4k segundos)
- **SLA:** 99.0%
- **Suporte:** Email (resposta < 48h)

**Upgrade triggers:** Quando usar >10k eventos/mês OU precisar de >1 projeto OU A/B testing.

---

### Pro
**R$ 297/mês | Faturamento mensal**

Para empresas que precisam de confiabilidade e escala.

- **Eventos/mês:** ilimitados
- **Feature flags:** ilimitadas
- **Projetos:** ilimitados
- **Usuários:** ilimitados
- **A/B testing:** sim
- **Histórico de eventos:** 90 dias
- **CSAT/NPS tracking:** sim
- **Retention analysis:** sim
- **Webhooks customizados:** sim
- **API priority:** SLA 1h
- **SLA:** 99.9%
- **Suporte:** Email + Slack channel (resposta < 2h)

**Upgrade triggers:** Quando usar >500k eventos/mês OU precisar de >5 projetos OU dados históricos 90 dias OU CSAT/NPS.

---

## Pricing Justificativa

### Comparativo com Concorrentes

| Recurso | Pulse Free | Pulse Growth | Pulse Pro | PostHog | LaunchDarkly | Amplitude |
|---------|-----------|-------------|----------|---------|--------------|-----------|
| Feature flags | ✅ 5 | ✅ Ilimitado | ✅ Ilimitado | ✅ Ilimitado | ✅ Ilimitado | ✅ (via Ampli) |
| Event tracking | ✅ 10k/mês | ✅ 500k/mês | ✅ Ilimitado | ✅ Ilimitado | ❌ | ✅ Ilimitado |
| A/B testing | ❌ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Preço entry | Grátis | $99/mês | $297/mês | $20-$475+ | $10-$500+ | $495-$1995+ |
| **Posicionamento** | MVP/dev | Startup BR | Empresa BR | SaaS maduro | Enterprise | Enterprise analytics |

**Rationale:**
- **PostHog:** USD 20/mês (≈ R$ 100) para Growth equivalente, mas hosted dedicado = maior overhead operacional
- **LaunchDarkly:** USD 10/mês entry, mas $50-500 em features = escalabilidade cara. Pulse 99/mês all-in.
- **Amplitude:** $495+ (fora do alcance de startups BR). Pulse Pro $297 = 40% do preço entry deles.

**Diferenciador Pulse:**
1. **Preço local:** em R$, com base no mercado brasileiro
2. **All-in:** flags + events + analytics num só plano
3. **Simplicidade:** sem limites ocultos ("pagamento por evento")
4. **Deploy flexível:** self-hosted ou managed

---

## Modelo de Faturamento

- **Cobrança:** mensal, automática via Asaas
- **Ciclo:** 1º ao último dia do mês
- **Cancelamento:** a qualquer momento, sem penalidade
- **Upgrade:** imediato, diferença pro-rata
- **Downgrade:** no mês seguinte

### Exemplo: Upgrade Mid-month
Usuário em **Free** (R$ 0) e faz upgrade para **Growth** (R$ 99) no dia 15 de um mês com 30 dias.
- Custo pro-rata: 99 × (16 dias ÷ 30) = R$ 52,80
- Próxima cobrança: R$ 99 (ciclo cheio)

---

## Overage Pricing

Se cliente em **Growth** ultrapassar 500k eventos:
- **Overage:** R$ 0,10 por 1k eventos acima do limite
- **Exemplo:** 600k eventos = 99 + (100k ÷ 1k × 0,10) = 99 + 10 = R$ 109

Clientes em **Pro** não têm overage (ilimitado).

---

## Migrações e Suporte

- **Free → Growth:** automático, após confirmar plan
- **Growth → Pro:** automático, após confirmar plan
- **Pro → Growth:** possível no 1º dia do mês seguinte
- **Suporte gratuito:** migração de dados, integração SDK

---

## Roadmap Futuro

- [ ] **Tier Enterprise** (R$ 1.990/mês): SLA 99.99%, dedicated account manager, custom integrations
- [ ] **Pay-as-you-go:** para picos de traffic (ex: campanha viral)
- [ ] **Adicional: Audit logs** (R$ 29/mês em Growth+)
- [ ] **Adicional: Custom retention** (R$ 49/mês em Growth+)

---

## FAQ

**P: Posso usar Free indefinidamente?**
R: Sim. Sem limite de tempo. Quando ultrapassar 10k eventos/mês, sugerimos upgrade para Growth.

**P: E se eu exceder 500k eventos em Growth?**
R: Você paga overage (R$ 0,10/1k) até o final do mês. Recomendamos fazer upgrade para Pro se for padrão recorrente.

**P: Vocês oferecem desconto anual?**
R: Não por enquanto, mas clientes Pro têm prioridade em futuras promoções.

**P: Como funciona a migração de dados entre planos?**
R: Automática e instantânea. Nenhuma perda de flags ou histórico.

**P: Qual a diferença entre Free e Growth em flags?**
R: Free: 5 flags máximo. Growth: ilimitadas. Pro: ilimitadas + CSAT/NPS tracking.

**P: Qual SLA tem o Free?**
R: Best effort (sem garantia formal). Growth tem 99.0%, Pro tem 99.9%.
