---
title: "Qual plano Pulse escolher"
description: "Árvore de decisão e tabela comparativa para escolher entre Free, Starter, Growth, Scale e Self-Hosted com base no seu volume e necessidades."
---

# Qual plano Pulse escolher

O Pulse tem cinco planos. Este guia ajuda você a escolher o certo sem pagar por capacidade que não vai usar.

---

## Tabela comparativa

| | Free | Starter | Growth | Scale | Self-Hosted |
|---|---|---|---|---|---|
| **Preço/mês** | R$ 0 | R$ 399 | R$ 1.099 | R$ 2.399 | — |
| **Preço anual (por mês)** | — | R$ 319 | R$ 879 | R$ 1.919 | R$ 50.000/ano (único) |
| **Eventos/mês** | 10.000 | 500.000 | 2.000.000 | 10.000.000 | Ilimitado |
| **Projetos** | 1 | 5 | 15 | Ilimitado | Ilimitado |
| **Feature flags** | 10 | 50 | 200 | Ilimitado | Ilimitado |
| **A/B Testing** | Não | Sim (básico) | Sim (avançado) | Sim (avançado) | Sim (avançado) |
| **Insights agent** | Não | Não | Sim | Sim | Sim |
| **Audit log** | 7 dias | 30 dias | 90 dias | 365 dias | Ilimitado |
| **Retenção de eventos** | 30 dias | 90 dias | 180 dias | 1 ano | Ilimitado |
| **Exportação de dados** | Não | CSV | CSV + API | CSV + API + Webhook | CSV + API + Webhook |
| **SSO / SAML** | Não | Não | Não | Sim | Sim |
| **SLA** | Sem SLA | 99,5% | 99,9% | 99,95% + suporte prioritário | — |
| **Dados on-premise** | Não | Não | Não | Não | Sim |
| **PULSE_LICENSE_KEY** | — | — | — | — | Obrigatório |
| **Overage (por 1k eventos extras)** | Bloqueado | R$ 4,00 | R$ 2,50 | R$ 1,50 | N/A |

---

## Árvore de decisão

### 1. Você está em fase de experimentação ou MVP?

**Sim** → comece com **Free**. Sem custo, sem cartão. Limite de 10.000 eventos/mês é suficiente para validar a integração e rodar testes internos com uma equipe pequena.

**Não** (produto em produção com usuários reais) → continue na árvore.

---

### 2. Quantos eventos únicos você rastreia por mês?

Estime somando todos os `pulse.track()` disparados. Se não sabe ainda, comece no Free e observe o medidor de quota no dashboard.

| Volume | Plano indicado |
|--------|---------------|
| Até 10.000 | Free |
| 10.001 – 500.000 | Starter |
| 500.001 – 2.000.000 | Growth |
| 2.000.001 – 10.000.000 | Scale |
| Acima de 10.000.000 | Self-Hosted |

> Dica: o Pulse mostra uma projeção de consumo do mês atual em **Settings → Quota**. Use isso para calibrar antes de fazer upgrade.

---

### 3. Você precisa de A/B testing com análise estatística?

**Não** (só feature flags on/off) → o **Starter** atende com flags básicas e 5 projetos.

**Sim, básico** (1-2 experimentos simultâneos, equipe pequena) → **Starter**.

**Sim, avançado** (múltiplos experimentos, segmentação de audiência, p-value automático, Insights agent) → **Growth** ou **Scale**.

---

### 4. Quantos projetos separados você tem?

| Projetos | Plano mínimo |
|----------|-------------|
| 1 | Free ou Starter |
| 2 – 5 | Starter |
| 6 – 15 | Growth |
| Mais de 15 | Scale |
| Dados on-premise | Self-Hosted |

No contexto do IIT, um "projeto" corresponde a um produto: Librinho, Nitro, Brio, SocialMake, etc. Cada um tem seu próprio conjunto de flags e eventos isolados.

---

### 5. Você precisa de compliance ou auditoria estendida?

**LGPD/auditoria com 1 ano de histórico** → **Scale** (audit log 365 dias).

**Revisão periódica de 3 meses** → **Growth**.

**Apenas debug recente** → Free (7 dias) ou Starter (30 dias).

---

### 6. Você tem requisitos de SSO ou SAML?

**Sim** (empresa com IdP corporativo) → **Scale** ou **Self-Hosted** (únicos planos com SSO/SAML).

---

### 7. Você precisa de instalação self-hosted (dados nunca saem da sua infra)?

**Sim** → **Self-Hosted** (R$ 50.000/ano, pagamento único anual, requer `PULSE_LICENSE_KEY`). Veja [08-self-hosted-como-funciona.md](./08-self-hosted-como-funciona.md).

---

## Resumo por perfil

| Perfil | Plano recomendado |
|--------|-------------------|
| Side project / MVP sem usuários | Free |
| Startup early-stage (1-2 produtos, <500k eventos) | Starter |
| Produto em crescimento (3-5 produtos, A/B testing ativo, Insights agent) | Growth |
| Empresa / múltiplos produtos / compliance / SSO | Scale |
| Enterprise com dados on-premise | Self-Hosted |

---

## Overage — quando estourar o limite vale a pena

Se você está no **Starter** e estoura 500.000 eventos em um mês específico (ex: uma campanha), o custo de overage é R$ 4,00 por mil eventos extras. Se você consistentemente usa 600.000 eventos/mês, o Growth (R$ 1.099) sai mais barato do que Starter (R$ 399) + overage recorrente de R$ 400.

Fórmula para calcular o ponto de troca:

```
Plano atual + overage mensal > Plano superior?
  → Fazer upgrade

Starter (R$399) + (100 * R$4,00) = R$799
Growth (R$1.099)

Se overage > R$700/mês → upgrade compensa
```

> Planos anuais têm 20% de desconto: Starter R$ 3.830/ano (R$ 319/mês), Growth R$ 10.550/ano (R$ 879/mês), Scale R$ 23.030/ano (R$ 1.919/mês).

---

## Fazer upgrade ou downgrade

- **Upgrade:** efeito imediato, cobrança proporcional ao período restante do ciclo.
- **Downgrade:** efeito no início do próximo ciclo de cobrança. Dados acima do limite do plano inferior ficam em modo somente leitura por 30 dias.

Para alterar o plano: **Dashboard → Settings → Plano → Alterar**.

---

## Próximos passos

- [06-limite-eventos-o-que-fazer.md](./06-limite-eventos-o-que-fazer.md) — recebeu HTTP 429, como otimizar e quando fazer upgrade
- [08-self-hosted-como-funciona.md](./08-self-hosted-como-funciona.md) — detalhes da instalação self-hosted
