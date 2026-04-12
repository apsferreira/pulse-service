# @pulse-iit/node

SDK Node.js server-side para o pulse-service IIT.

Buffer local com flush automatico a cada 5s ou ao atingir 50 eventos — ideal para APIs e workers backend.

> Para uso no browser ou React, use `@iit/pulse-sdk` (sdk/js).

## Instalacao

```bash
npm install @pulse-iit/node
# ou
pnpm add @pulse-iit/node
```

## Uso rapido — Express.js

```typescript
import express from "express";
import { PulseClient } from "@pulse-iit/node";

const pulse = new PulseClient({
  projectId: process.env.PULSE_PROJECT_ID!,
  apiUrl: process.env.PULSE_API_URL!,
  // flushInterval: 5000,  // default: 5s
  // maxBatchSize: 50,     // default: 50 eventos
});

const app = express();

// Middleware de tracking por rota
app.use((req, _res, next) => {
  pulse.track("http_request", {
    method: req.method,
    path: req.path,
    userAgent: req.headers["user-agent"],
  });
  next();
});

app.post("/api/checkout", async (req, res) => {
  const { userId } = req.body as { userId: string };

  // Identificar usuario
  pulse.identify(userId, { plan: "pro" });

  // Evento de negocio
  pulse.track("checkout_started", { cartValue: 99.9 }, userId);

  res.json({ ok: true });
});

// Graceful shutdown — garante que eventos em buffer sao enviados
process.on("SIGTERM", async () => {
  await pulse.shutdown();
  process.exit(0);
});

app.listen(3000);
```

## API

### `new PulseClient(options)`

| Opcao | Tipo | Default | Descricao |
|-------|------|---------|-----------|
| `projectId` | `string` | obrigatorio | ID do projeto no pulse-service |
| `apiUrl` | `string` | obrigatorio | URL base do pulse-service |
| `flushInterval` | `number` | `5000` | Intervalo de flush em ms |
| `maxBatchSize` | `number` | `50` | Eventos por batch antes de flush forcado |
| `requestTimeout` | `number` | `10000` | Timeout por requisicao HTTP em ms |
| `fetch` | `function` | `globalThis.fetch` | Custom fetch (util para testes) |

### `pulse.track(event, properties?, userId?)`

Adiciona evento de tracking ao buffer. Fire-and-forget — nao bloqueia.

### `pulse.identify(userId, traits?)`

Adiciona evento de identificacao de usuario ao buffer.

### `pulse.flush(): Promise<void>`

Envia todos os eventos em buffer imediatamente. Seguro para chamar multiplas vezes.

### `pulse.shutdown(): Promise<void>`

Flush final + para o timer automatico. Chamar no `SIGTERM`/`SIGINT` do processo.

## Requisitos

- Node.js >= 18 (fetch nativo)
- Para Node < 18: passar `fetch` customizado via opcoes (ex: `node-fetch`)

## Endpoint esperado no pulse-service

`POST /api/v1/events/batch`

```json
{
  "projectId": "proj_xxx",
  "events": [
    {
      "type": "track",
      "event": "checkout_started",
      "userId": "user_123",
      "properties": { "cartValue": 99.9 },
      "projectId": "proj_xxx",
      "timestamp": "2026-04-11T10:00:00.000Z"
    }
  ]
}
```
