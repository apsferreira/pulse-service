---
title: "Qual plano Pulse escolher"
description: "Árvore de decisão e tabela comparativa para escolher entre Free, Starter, Growth e Pro com base no seu volume e necessidades."
---

# Qual plano Pulse escolher

O Pulse tem quatro planos. Este guia ajuda você a escolher o certo sem pagar por capacidade que não vai usar.

---

## Tabela comparativa

| | Free | Starter | Growth | Pro |
|---|---|---|---|---|
| **Preço/mês** | R$ 0 | R$ 399 | R$ 999 | R$ 2.499 |
| **Eventos/mês** | 10.000 | 100.000 | 500.000 | Ilimitado |
| **Projetos** | 1 | 3 | 10 | Ilimitado |
| **Feature flags** | 10 | 50 | 200 | Ilimitado |
| **A/B Testing** | Não | Sim (básico) | Sim (avançado) | Sim (avançado) |
| **Audit log** | 7 dias | 30 dias | 90 dias | 365 dias |
| **Retenção de eventos** | 30 dias | 90 dias | 180 dias | 1 ano |
| **Exportação de dados** | Não | CSV | CSV + API | CSV + API + Webhook |
| **SSO / SAML** | Não | Não | Não | Sim |
| **SLA** | Sem SLA | 99,5% | 99,9% | 99,95% + suporte prioritário |
| **Self-hosted** | Não | Não | Não | Sob consulta |
| **Overage (por 1k eventos extras)** | Bloqueado | R$ 4,00 | R$ 2,50 | Incluso |

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
| 10.001 – 100.000 | Starter |
| 100.001 – 500.000 | Growth |
| Acima de 500.000 | Pro |

> Dica: o Pulse mostra uma projeção de consumo do mês atual em **Settings → Quota**. Use isso para calibrar antes de fazer upgrade.

---

### 3. Você precisa de A/B testing com análise estatística?

**Não** (só feature flags on/off) → o **Starter** atende com flags básicas e 3 projetos.

**Sim, básico** (1-2 experimentos simultâneos, equipe pequena) → **Starter**.

**Sim, avançado** (múltiplos experimentos, segmentação de audiência, p-value automático) → **Growth** ou **Pro**.

---

### 4. Quantos projetos separados você tem?

| Projetos | Plano mínimo |
|----------|-------------|
| 1 | Free ou Starter |
| 2 – 3 | Starter |
| 4 – 10 | Growth |
| Mais de 10 | Pro |

No contexto do IIT, um "projeto" corresponde a um produto: Libri, Nitro, Brio, SocialMake, etc. Cada um tem seu próprio conjunto de flags e eventos isolados.

---

### 5. Você precisa de compliance ou auditoria estendida?

**LGPD/auditoria com 1 ano de histórico** → **Pro** (audit log 365 dias).

**Revisão periódica de 3 meses** → **Growth**.

**Apenas debug recente** → Free (7 dias) ou Starter (30 dias).

---

### 6. Você tem requisitos de SSO ou SAML?

**Sim** (empresa com IdP corporativo) → **Pro** (único plano com SSO/SAML).

---

### 7. Você precisa de instalação self-hosted (dados nunca saem da sua infra)?

**Sim** → **Pro** (self-hosted disponível sob consulta). Veja [08-self-hosted-como-funciona.md](./08-self-hosted-como-funciona.md).

---

## Resumo por perfil

| Perfil | Plano recomendado |
|--------|-------------------|
| Side project / MVP sem usuários | Free |
| Startup early-stage (1 produto, <100k eventos) | Starter |
| Produto em crescimento (3-5 produtos, A/B testing ativo) | Growth |
| Empresa / múltiplos produtos / compliance / SSO | Pro |
| Enterprise com dados on-premise | Pro + Self-hosted |

---

## Overage — quando estourar o limite vale a pena

Se você está no **Starter** e estoura 100.000 eventos em um mês específico (ex: uma campanha), o custo de overage é R$ 4,00 por mil eventos extras. Se você consistentemente usa 150.000 eventos/mês, o Growth (R$ 999) sai mais barato do que Starter (R$ 399) + overage recorrente de R$ 200.

Fórmula para calcular o ponto de troca:

```
Plano atual + overage mensal > Plano superior?
  → Fazer upgrade

Starter (R$399) + (50 * R$4,00) = R$599
Growth (R$999)

Se overage > R$600/mês → upgrade compensa
```

---

## Fazer upgrade ou downgrade

- **Upgrade:** efeito imediato, cobrança proporcional ao período restante do ciclo.
- **Downgrade:** efeito no início do próximo ciclo de cobrança. Dados acima do limite do plano inferior ficam em modo somente leitura por 30 dias.

Para alterar o plano: **Dashboard → Settings → Plano → Alterar**.

---

## Próximos passos

- [06-limite-eventos-o-que-fazer.md](./06-limite-eventos-o-que-fazer.md) — recebeu HTTP 429, como otimizar e quando fazer upgrade
- [08-self-hosted-como-funciona.md](./08-self-hosted-como-funciona.md) — detalhes da instalação self-hosted (Pro)
