# Pulse — Dashboard Analytics Unificado (Todos os Produtos IIT)

> BKL-424 | Spec v1.0 | Abril 2026
> Acesso: Admin IIT only (role `iit_admin`)

## Objetivo

View única no Pulse que consolida os KPIs de todos os 6 produtos do ecossistema IIT. Permite ao CEO e time de produto enxergar saúde do ecossistema em um único painel, sem trocar de contexto entre ferramentas.

---

## Wireframe Textual

```
┌─────────────────────────────────────────────────────────────────────┐
│  PULSE  /  Ecossistema IIT                  [Últimos 30 dias ▼]     │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  RESUMO GERAL                                                       │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌───────────┐ │
│  │ Usuários    │  │ MRR Total   │  │ Eventos/dia │  │ Churn     │ │
│  │ Ativos      │  │             │  │ (todos)     │  │ Rate      │ │
│  │  1.247      │  │ R$ 18.420   │  │  42.810     │  │  2,3%     │ │
│  │ ↑ 12% mês  │  │ ↑ 8% mês   │  │ ↑ 31% mês  │  │ ↓ 0,4pp  │ │
│  └─────────────┘  └─────────────┘  └─────────────┘  └───────────┘ │
│                                                                     │
│  POR PRODUTO  ────────────────────────────────────────────────────  │
│                                                                     │
│  ┌──────────────────┬──────────┬──────────┬─────────┬──────────┐   │
│  │ Produto          │ Usuários │ Eventos  │ MRR     │ Saúde    │   │
│  ├──────────────────┼──────────┼──────────┼─────────┼──────────┤   │
│  │ SocialMake       │    312   │  18.400  │ R$6.180 │  ●● ótima│   │
│  │ Nitro (FocusHub) │    428   │  12.200  │ R$4.280 │  ●● ótima│   │
│  │ Libri (MyLib.)   │    291   │   8.100  │ R$2.910 │  ●  ok   │   │
│  │ MeuIngresso      │     87   │   2.300  │ R$1.740 │  ●  ok   │   │
│  │ Brio (Attend)    │     94   │   1.480  │ R$4.700 │  ▲ atenc.│   │
│  │ Pulse            │     35   │     330  │ R$  610 │  ▲ atenc.│   │
│  └──────────────────┴──────────┴──────────┴─────────┴──────────┘   │
│                                                                     │
│  SÉRIE TEMPORAL  [Usuários ▼]  [1S | 1M | 3M | 6M | 1A]           │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │                                              ____SocialMake   │   │
│  │                                         ____Nitro            │   │
│  │                                    _____Libri                │   │
│  │  ___________________________________________outros           │   │
│  │  Jan  Fev  Mar  Abr  Mai  Jun  Jul  Ago  Set  Out  Nov  Dez  │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  EVENTOS RECENTES (cross-produto)              [ver todos →]        │
│  ● 14:32  SocialMake   post_published          tenant: agencia_x    │
│  ● 14:28  Nitro        session_completed       user: maria@...      │
│  ● 14:15  Brio         escalation_triggered    tenant: loja_y       │
│  ● 13:58  Libri        book_added              user: joao@...       │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## KPIs por Produto

### SocialMake
| KPI | Descrição | Evento Pulse |
|-----|-----------|--------------|
| Posts publicados | Total de publicações realizadas | `post_published` |
| Conexões ativas | Contas de redes conectadas | `account_connected` |
| Taxa de aprovação | Posts aprovados / total submetidos | `post_approved` / `post_submitted` |
| Custo de geração IA | Tokens consumidos × tarifa | `ai_generation_completed` (metadata: cost) |
| Tenants ativos | Tenants com ao menos 1 post nos últimos 30d | derivado |

### Nitro (FocusHub)
| KPI | Descrição | Evento Pulse |
|-----|-----------|--------------|
| Sessões Pomodoro concluídas | Total de sessões finalizadas | `session_completed` |
| Tarefas criadas | Volume de tasks adicionadas | `task_created` |
| Tarefas concluídas | Volume de tasks completadas | `task_completed` |
| Taxa de conclusão | Tarefas concluídas / criadas | derivado |
| DAU (usuários ativos/dia) | Usuários com ao menos 1 sessão | `session_started` |

### Libri (MyLibrary)
| KPI | Descrição | Evento Pulse |
|-----|-----------|--------------|
| Livros cadastrados | Total de livros na coleção | `book_added` |
| Livros lidos | Marcados como concluídos | `book_completed` |
| Empréstimos ativos | Livros em posse de outro usuário | `loan_created` |
| Scans de ISBN | Uso do leitor de código de barras | `isbn_scanned` |

### MeuIngresso
| KPI | Descrição | Evento Pulse |
|-----|-----------|--------------|
| Eventos criados | Eventos publicados por organizadores | `event_created` |
| Ingressos vendidos | Transações de venda concluídas | `ticket_sold` |
| Check-ins realizados | Validações na porta | `checkin_done` |
| Receita processada | Volume financeiro (R$) | `ticket_sold` (metadata: amount) |

### Brio (Attend-Service)
| KPI | Descrição | Evento Pulse |
|-----|-----------|--------------|
| Conversas iniciadas | Sessões de atendimento abertas | `conversation_started` |
| Taxa de resolução | Resolvidas pelo agente / total | `conversation_resolved` / total |
| Escalações | Transferências para humano | `escalation_triggered` |
| Tempo médio de resposta | Latência média do agente (ms) | `message_sent` (metadata: latency) |
| CSAT | Nota média pós-atendimento | `csat_submitted` |

### Pulse (self-monitoring)
| KPI | Descrição | Evento Pulse |
|-----|-----------|--------------|
| Tenants ativos | Tenants com eventos nos últimos 30d | derivado |
| Eventos ingeridos | Total de eventos recebidos | interno (tabela `events`) |
| Feature flags ativos | Flags no estado `enabled` | derivado |
| Latência de ingestão | p99 de tempo de ingestão (ms) | interno |

---

## Filtros Temporais

| Opção | Período |
|-------|---------|
| Hoje | Últimas 24h |
| 7 dias | D-7 até hoje |
| 30 dias | D-30 até hoje (padrão) |
| 3 meses | D-90 até hoje |
| 6 meses | D-180 até hoje |
| 1 ano | D-365 até hoje |
| Customizado | Date picker — qualquer intervalo |

Todos os filtros são aplicados globalmente e por produto simultaneamente.

---

## Controle de Acesso

**Acesso restrito a `iit_admin`.**

```
Regra de autorização:
- User precisa ter claim `role = iit_admin` no JWT
- OU pertencer ao tenant slug = `iit` com plan = `internal`
- O dashboard unificado NÃO é acessível a tenants externos
- Tenants externos veem apenas seus próprios dados (comportamento atual do Pulse)
```

Implementação sugerida: middleware `RequireIITAdmin()` no router do Pulse, verificando o claim antes de servir os endpoints `/api/v1/ecosystem/*`.

---

## Dependências Técnicas

### Pré-requisito: todos os produtos precisam ter `project_id` configurado no Pulse

| Produto | project_id (produção) | Status |
|---------|-----------------------|--------|
| SocialMake | `proj_socialmake_prod` | A confirmar |
| Nitro | `proj_nitro_prod` | A confirmar |
| Libri | `proj_libri_prod` | A confirmar |
| MeuIngresso | `proj_meuingresso_prod` | A configurar |
| Brio | `proj_brio_prod` | A configurar |
| Pulse | `proj_pulse_self` | A configurar |

### Endpoints necessários no Pulse

```
GET /api/v1/ecosystem/summary
  → Totais consolidados (usuários, eventos, MRR estimado)
  → Autenticação: iit_admin only

GET /api/v1/ecosystem/by-product
  → KPIs por produto (tabela)
  → Autenticação: iit_admin only

GET /api/v1/ecosystem/timeseries?metric=users&from=&to=&granularity=day
  → Série temporal por produto
  → Autenticação: iit_admin only

GET /api/v1/ecosystem/events/recent?limit=50
  → Feed de eventos recentes cross-produto
  → Autenticação: iit_admin only
```

### Queries de banco (PostgreSQL)

MRR estimado é calculado cruzando o `project_id` com a tabela de tenants/planos do `billing-service`. Precisa de uma view materializada ou chamada à API do billing.

Alternativa simples: cada produto envia um evento `mrr_snapshot` diário com o valor atual — o Pulse agrega sem precisar de integração com billing.

### Frontend

- Novo módulo no Pulse frontend: `/ecosystem` (rota protegida por role)
- Componentes a criar: `EcosystemSummaryCards`, `ProductTable`, `EcosystemTimeseries`, `CrossProductEventFeed`
- Reutilizar componentes de chart existentes no Pulse

---

## Critérios de Aceite

- [ ] Dashboard carrega em menos de 2 segundos com dados dos últimos 30 dias
- [ ] Dados por produto corretos para ao menos 3 produtos com events reais
- [ ] Filtro temporal funcional (mudar intervalo atualiza todos os cards)
- [ ] Acesso bloqueado para usuários sem role `iit_admin`
- [ ] Série temporal exibe linhas separadas por produto com legenda
- [ ] Feed de eventos recentes atualiza sem reload (polling a cada 60s)

---

## Estimativa de Esforço

| Tarefa | Esforço |
|--------|---------|
| Endpoints no Pulse backend | 3 dias |
| Configurar project_ids em todos os produtos | 1 dia |
| Frontend — dashboard completo | 4 dias |
| Testes + ajustes | 1 dia |
| **Total** | **~9 dias** |
