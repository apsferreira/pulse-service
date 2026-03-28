# @iit/pulse-sdk

SDK JS/TS para integração com o pulse-service (Feature Flags + Event Tracking).

## Instalação

```bash
npm install @iit/pulse-sdk
```

## Uso

```typescript
import { PulseClient } from "@iit/pulse-sdk";

const pulse = new PulseClient({
  baseURL: "https://pulse.institutoitinerante.com.br",
});

// Feature Flags
const flags = await pulse.getFlags();
const isDarkMode = await pulse.isEnabled("dark-mode");

// Event Tracking
await pulse.trackEvent({
  event_name: "page_view",
  user_id: "user-123",
  properties: { page: "/home" },
});
```
