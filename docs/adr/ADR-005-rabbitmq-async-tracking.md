# ADR-005 — RabbitMQ para Tracking Assíncrono no Pulse

**Status:** Accepted  
**Data:** 2026-04-11  
**BKL:** BKL-1116  
**Autor:** Antonio Pedro Ferreira

---

## Contexto

O endpoint `POST /api/v1/events` do pulse-service recebe eventos de tracking e executa `INSERT` direto no PostgreSQL de forma síncrona. Cada request HTTP aguarda a confirmação da escrita no banco antes de retornar.

Esse modelo funciona bem em baixo volume. Sob carga elevada — múltiplos projetos ativos publicando eventos simultaneamente — surgem dois problemas:

1. **Contenção no banco:** muitas conexões concorrentes competindo por lock e I/O no PostgreSQL
2. **Latência na resposta:** o cliente fica bloqueado esperando a escrita completar, aumentando o tempo de resposta percebido

O padrão já adotado em outros serviços do ecossistema (notification-service, auth-service) usa RabbitMQ para desacoplar produtores de consumidores. O RabbitMQ já está disponível na infraestrutura compartilhada.

---

## Decisao

Adicionar uma fila RabbitMQ entre o endpoint `/api/v1/events` e a escrita no PostgreSQL.

O endpoint publica o evento na fila e retorna `202 Accepted` imediatamente. Um consumer worker consome a fila, acumula eventos em buffer e executa `INSERT` em batch.

### Arquitetura

```
HTTP Request (POST /api/v1/events)
    │
    ▼
pulse-service handler
    │  publish (fire-and-forget)
    ▼
Exchange: pulse.events (type: direct)
    │
    ▼
Queue: pulse.events.tracking
    │
    ▼
Consumer Worker (goroutine)
    │  buffer: 50 eventos OU 5s (o que vier primeiro)
    ▼
PostgreSQL — batch INSERT (single statement com multiple VALUES)
```

### Detalhes de implementacao

- **Exchange:** `pulse.events`, tipo `direct`, durable
- **Queue:** `pulse.events.tracking`, durable, dead-letter para `pulse.events.dlq`
- **Consumer:** goroutine dedicada com channel Go de buffer
- **Batch:** flush quando buffer atinge 50 eventos OU timer de 5s expira
- **Dead Letter Queue:** `pulse.events.dlq` — eventos que falharam apos 3 retentativas ficam para analise
- **Serialização:** JSON (mesmo formato atual do payload de evento)
- **Graceful shutdown:** consumer drena o buffer antes de fechar (flush final)

### Mudancas no contrato HTTP

| Antes | Depois |
|-------|--------|
| `201 Created` com evento persistido | `202 Accepted` com `{"queued": true}` |

O endpoint `/api/v1/events/batch` (usado pelos SDKs) segue o mesmo padrao.

---

## Consequencias positivas

- Desacopla pico de trafego HTTP da escrita no banco
- Batch inserts reduzem significativamente o numero de transacoes no PostgreSQL
- Retry automatico para eventos que falham na escrita (via RabbitMQ requeue)
- Latencia de resposta ao cliente cai de ~20ms (insert sync) para ~2ms (publish async)
- Padrao consistente com notification-service e auth-service

## Consequencias negativas

- **Latencia eventual nos dashboards:** eventos podem demorar ate 5s para aparecer nas queries de analytics. Aceitavel para o caso de uso de analytics; dashboards em tempo real podem precisar de websocket separado
- **Complexidade adicional:** mais um componente na stack — consumer worker, dead-letter queue, monitoramento do RabbitMQ
- **Dependencia de infraestrutura:** RabbitMQ precisa estar UP; se cair, os eventos ficam sem rota. Mitigacao: dead-letter + alertas no Grafana
- **Requer RabbitMQ no K3s:** ja disponivel no shared-infra, mas precisa de manifest dedicado no namespace `pulse`

---

## Alternativas consideradas

### Sem fila (status quo)
INSERT síncrono por request. Simples, sem dependencias extras. Nao escala alem de ~5k eventos/dia por projeto sem degradacao perceptivel.

### Channel Go interno (sem RabbitMQ)
Goroutine consumer com channel Go para bufferizar antes do INSERT. Mais simples que RabbitMQ, sem infra adicional. Contra: sem persistencia — eventos em buffer sao perdidos se o processo reiniciar. Sem retry automatico. Descartado por falta de durabilidade.

### TimescaleDB / ClickHouse
Banco de dados orientado a séries temporais que absorveria o volume de inserts sem bottleneck. Valido a longo prazo mas troca de banco e mudanca arquitetural muito maior para o volume atual.

---

## Decisao de implementacao

**Implementar no Q3 2026** quando o volume de eventos exceder 10.000 eventos/dia por projeto de forma sustentada.

Ate esse threshold, o INSERT síncrono e suficiente e menos complexo. O prazo de Q3 2026 permite:
1. Validar volume real apos lancamento (Q2 2026)
2. Implementar com dados de producao como referencia de sizing
3. Evitar over-engineering prematuro

Se o volume ultrapassar 10k/dia antes do Q3, acionar implementacao antecipada via BKL de urgência.

---

## Referencias

- [notification-service consumer pattern](../../notification-service/internal/consumer/)
- [auth-service RabbitMQ config](../../auth-service/internal/messaging/)
- [ANALISE-COMPETITIVA-PULSE-2026.md](../ANALISE-COMPETITIVA-PULSE-2026.md)
- BKL-1038 — cross-tenant leak via WebSocket (relacionado: event_handler usa string livre como discriminador)
