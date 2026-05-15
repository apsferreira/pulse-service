# Eventos no Pulse

Eventos registram ações dos usuários em tempo real. São a base para métricas de produto, análise de funil, A/B testing e auditoria de uso.

---

## O que são eventos no Pulse?

Um evento é qualquer ação rastreável de um usuário em um produto:

```
page_view          → usuário abriu uma página
button_click       → clicou em um botão
checkout.completed → finalizou uma compra
checkin.completed  → fez check-in em um evento (Catraca Virtual)
ai.generated       → IA gerou conteúdo (SocialMake)
```

Eventos são imutáveis: uma vez criados, não são editados. O histórico completo fica acessível no dashboard e via API.

---

## Estrutura de um evento

| Campo | Tipo | Obrigatório | Descrição |
|-------|------|-------------|-----------|
| `event_type` | string | sim | Nome do evento (snake_case ou dot.notation) |
| `product` | string | sim | Slug do projeto no Pulse |
| `user_id` | string | sim | ID do usuário (UUID, email ou session_id) |
| `tenant_id` | string | não | ID do tenant (para produtos multi-tenant) |
| `properties` | object | não | JSON livre — qualquer dado relevante |

Campos gerados pelo servidor (não enviar):

| Campo | Tipo | Descrição |
|-------|------|-----------|
| `id` | UUID | Identificador único do evento |
| `ip_address` | string | IP de origem (extraído do request) |
| `created_at` | timestamp | Data/hora UTC de recebimento |

---

## Enviar evento via API

**Auth:** `X-Project-Key` (api_key do projeto — mesma usada para avaliar flags)
**Endpoint:** `POST /api/v1/ingest`
**Base URL (prod):** `https://pulse.institutoitinerante.com.br`

```bash
curl -X POST https://pulse.institutoitinerante.com.br/api/v1/ingest \
  -H "X-Project-Key: pulse_libri_xxxxxxxxxxxx" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": "page_view",
    "product": "librinho",
    "user_id": "user-456",
    "properties": {
      "page": "/catalog",
      "referrer": "home"
    }
  }'
```

Resposta de sucesso (`202 Accepted`):

```json
{
  "id": "evt_01hx...",
  "status": "accepted"
}
```

O endpoint retorna `202` — o evento é processado de forma assíncrona. Falhas de rede são silenciosas do lado do usuário (fail-open).

---

## Enviar evento via SDK JavaScript

```html
<!-- Incluir o snippet no <head> -->
<script>
  window.PulseSDK_config = {
    projectKey: 'pulse_site_xxxxxxxxxxxx',
    apiUrl: 'https://pulse.institutoitinerante.com.br',
    userId: '{{ user.id }}' // ou session_id para usuários anônimos
  }
</script>
<script src="https://pulse.institutoitinerante.com.br/sdk/pulse.min.js" async></script>
```

Após o carregamento:

```javascript
// Rastrear page view
PulseSDK.track('page_view', {
  page: window.location.pathname,
  referrer: document.referrer
})

// Rastrear clique em botão
document.getElementById('btn-cta').addEventListener('click', () => {
  PulseSDK.track('button_click', {
    button: 'cta_hero',
    page: 'landing'
  })
})
```

Em React/TypeScript:

```tsx
import { usePulse } from '@pulse/react'

function CheckoutButton() {
  const { track } = usePulse()

  const handleClick = () => {
    track('checkout.started', {
      plan: 'growth',
      source: 'pricing_page'
    })
    // continuar com o fluxo...
  }

  return <button onClick={handleClick}>Assinar agora</button>
}
```

SDK completo com React Native e Go: [`pulse-service/docs/guides/sdk-reference.md`](../../../../pulse-service/docs/guides/sdk-reference.md)

---

## Casos de uso reais no IIT

### Catraca Virtual — check-in em eventos

```bash
curl -X POST https://pulse.institutoitinerante.com.br/api/v1/ingest \
  -H "X-Project-Key: pulse_catraca_xxxxxxxxxxxx" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": "checkin.completed",
    "product": "catraca",
    "user_id": "participant-789",
    "tenant_id": "tenant-iit",
    "properties": {
      "event_id": "evt-workshop-go-2026",
      "ticket_type": "presencial",
      "checkin_method": "qrcode"
    }
  }'
```

Usado para: auditoria de presença, relatório pós-evento, análise de horário de pico.

### SocialMake — geração de conteúdo por IA

```bash
curl -X POST https://pulse.institutoitinerante.com.br/api/v1/ingest \
  -H "X-Project-Key: pulse_sm_xxxxxxxxxxxx" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": "ai.generated",
    "product": "socialmake",
    "user_id": "user-123",
    "tenant_id": "tenant-iit",
    "properties": {
      "model": "claude-sonnet",
      "content_type": "instagram_caption",
      "word_count": 87,
      "campaign_id": "camp-maio-2026"
    }
  }'
```

Usado para: rastrear consumo de IA por tenant, billing, detecção de abuso.

### Site do Instituto — page views

```javascript
// Em cada página do site institutoitinerante.com.br
PulseSDK.track('page_view', {
  page: window.location.pathname,
  utm_source: new URLSearchParams(window.location.search).get('utm_source'),
  utm_campaign: new URLSearchParams(window.location.search).get('utm_campaign')
})
```

Usado para: funil de aquisição, análise de campanhas, páginas mais visitadas.

### Librinho — scan de ISBN

```bash
curl -X POST https://pulse.institutoitinerante.com.br/api/v1/ingest \
  -H "X-Project-Key: pulse_libri_xxxxxxxxxxxx" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": "book.scanned",
    "product": "librinho",
    "user_id": "user-456",
    "properties": {
      "isbn": "9788535908583",
      "scan_method": "camera",
      "found_in_catalog": true
    }
  }'
```

---

## Convenção de nomenclatura de eventos

Usar `dominio.acao` em snake_case/dot.notation:

| Dominio | Eventos |
|---------|---------|
| `page` | `page_view`, `page_exit` |
| `checkout` | `checkout.started`, `checkout.completed`, `checkout.abandoned` |
| `checkin` | `checkin.completed`, `checkin.failed` |
| `ai` | `ai.generated`, `ai.error` |
| `auth` | `auth.login`, `auth.logout`, `auth.signup` |
| `book` | `book.scanned`, `book.added`, `book.finished` |

Evitar nomes genéricos como `click` ou `action` sem contexto.

---

## Visualizar eventos no dashboard

Acesse `pulse.institutoitinerante.com.br` → **Projetos → [produto] → Eventos**

O dashboard exibe:
- Feed de eventos em tempo real
- Filtros por `event_type`, `user_id`, `tenant_id` e intervalo de datas
- Contagem por tipo de evento ao longo do tempo
- Propriedades de cada evento em JSON expansível

Para exportar dados de eventos: [`pulse-service/docs/guides/07-exportar-dados-testes.md`](../../../../pulse-service/docs/guides/07-exportar-dados-testes.md)

---

Ver também: [`ab-testing.md`](ab-testing.md) — usar eventos de conversão para medir variantes de A/B tests.
