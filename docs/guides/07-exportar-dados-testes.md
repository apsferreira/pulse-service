---
title: "Exportar dados de testes"
description: "Como exportar eventos e audit log de flags via API, o formato JSON das respostas e como conectar os dados em planilhas e ferramentas de BI."
---

# Exportar dados de testes

O Pulse oferece exportação de eventos e audit log de flags via API REST. Este guia cobre os endpoints, filtros disponíveis e como integrar os dados com ferramentas externas.

> Exportação via CSV está disponível a partir do plano **Starter**. Exportação via API está disponível a partir do **Growth**. Webhooks de exportação contínua estão disponíveis no **Pro**.

---

## Exportar eventos

### Endpoint

```
GET /api/v1/events
```

### Parâmetros de filtro

| Parâmetro | Tipo | Descrição |
|-----------|------|-----------|
| `product` | string | Obrigatório. Slug do projeto. |
| `event` | string | Filtrar por nome de evento. |
| `user_id` | string | Filtrar por usuário específico. |
| `from` | ISO 8601 | Data de início (inclusive). |
| `to` | ISO 8601 | Data de fim (inclusive). |
| `limit` | int | Máximo de registros por página (padrão: 100, máximo: 1000). |
| `cursor` | string | Cursor de paginação (retornado na resposta anterior). |

### Exemplo — últimos 7 dias

```bash
FROM=$(date -u -v-7d +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u --date="7 days ago" +"%Y-%m-%dT%H:%M:%SZ")
TO=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

curl -s "https://pulse-api.institutoitinerante.com.br/api/v1/events?product=minha-app&from=$FROM&to=$TO&limit=500" \
  -H "X-Service-Token: $SERVICE_TOKEN" | python3 -m json.tool
```

### Formato JSON da resposta

```json
{
  "data": [
    {
      "id": "evt_abc123",
      "product": "minha-app",
      "event": "checkout_completed",
      "user_id": "user-123",
      "tenant_id": "tenant-456",
      "properties": {
        "order_id": "ord_789",
        "value": 149.90,
        "currency": "BRL"
      },
      "created_at": "2026-04-11T10:30:00Z"
    },
    {
      "id": "evt_def456",
      "product": "minha-app",
      "event": "signup_completed",
      "user_id": "user-124",
      "tenant_id": null,
      "properties": {
        "plan": "free",
        "source": "organic"
      },
      "created_at": "2026-04-11T10:28:00Z"
    }
  ],
  "meta": {
    "total": 4821,
    "returned": 500,
    "has_more": true,
    "next_cursor": "eyJpZCI6ImV2dF9kZWY0NTYiLCJ0cyI6IjIwMjYtMDQtMTFUMTA6Mjg6MDBaIn0="
  }
}
```

### Paginação com cursor

```bash
# Primeira página
RESPONSE=$(curl -s "https://pulse-api.institutoitinerante.com.br/api/v1/events?product=minha-app&limit=1000" \
  -H "X-Service-Token: $SERVICE_TOKEN")

# Verificar se há mais páginas
HAS_MORE=$(echo $RESPONSE | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['meta']['has_more'])")
CURSOR=$(echo $RESPONSE | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['meta'].get('next_cursor',''))")

# Próxima página
if [ "$HAS_MORE" = "True" ]; then
  curl -s "https://pulse-api.institutoitinerante.com.br/api/v1/events?product=minha-app&limit=1000&cursor=$CURSOR" \
    -H "X-Service-Token: $SERVICE_TOKEN"
fi
```

---

## Exportar audit log de flags

O audit log registra toda criação, edição, habilitação e avaliação de flags.

### Endpoint

```
GET /api/v1/flags/audit
```

### Parâmetros

| Parâmetro | Tipo | Descrição |
|-----------|------|-----------|
| `product` | string | Obrigatório. |
| `flag` | string | Filtrar por nome de flag específica. |
| `action` | string | `created`, `updated`, `enabled`, `disabled`, `evaluated`. |
| `from` | ISO 8601 | Data de início. |
| `to` | ISO 8601 | Data de fim. |
| `limit` | int | Máximo de registros (padrão: 100). |

### Exemplo

```bash
curl -s "https://pulse-api.institutoitinerante.com.br/api/v1/flags/audit?product=minha-app&action=evaluated&limit=100" \
  -H "X-Service-Token: $SERVICE_TOKEN"
```

### Formato JSON do audit log

```json
{
  "data": [
    {
      "id": "audit_xyz789",
      "product": "minha-app",
      "flag": "nova_homepage",
      "action": "evaluated",
      "user_id": "user-123",
      "result": true,
      "reason": "rollout_hit",
      "rollout_percentage": 50,
      "user_bucket": 23,
      "actor": null,
      "created_at": "2026-04-11T11:00:00Z"
    },
    {
      "id": "audit_xyz790",
      "product": "minha-app",
      "flag": "nova_homepage",
      "action": "updated",
      "user_id": null,
      "result": null,
      "reason": null,
      "changes": {
        "rollout_percentage": { "from": 10, "to": 50 }
      },
      "actor": "antonio@institutoitinerante.com.br",
      "created_at": "2026-04-11T09:00:00Z"
    }
  ]
}
```

---

## Exportar para CSV

Para o plano Starter+, use o endpoint de exportação CSV:

```bash
# Baixar CSV de eventos
curl -s "https://pulse-api.institutoitinerante.com.br/api/v1/events/export?product=minha-app&format=csv&from=2026-04-01T00:00:00Z" \
  -H "X-Service-Token: $SERVICE_TOKEN" \
  -o eventos_abril_2026.csv

head -5 eventos_abril_2026.csv
```

```
id,product,event,user_id,tenant_id,properties,created_at
evt_abc123,minha-app,checkout_completed,user-123,tenant-456,"{""order_id"":""ord_789"",""value"":149.90}",2026-04-11T10:30:00Z
evt_def456,minha-app,signup_completed,user-124,,"{""plan"":""free""}",2026-04-11T10:28:00Z
```

---

## Conectar em Google Sheets

Use o Apps Script para importar eventos do Pulse diretamente em uma planilha:

```javascript
// Google Apps Script — Tools → Script Editor
// Cole este código e configure o trigger para rodar diariamente

const PULSE_TOKEN = PropertiesService.getScriptProperties().getProperty('PULSE_TOKEN')
const PULSE_API = 'https://pulse-api.institutoitinerante.com.br'
const PRODUCT = 'minha-app'

function importarEventos() {
  const sheet = SpreadsheetApp.getActiveSpreadsheet().getSheetByName('Eventos') 
    || SpreadsheetApp.getActiveSpreadsheet().insertSheet('Eventos')

  const ontem = new Date()
  ontem.setDate(ontem.getDate() - 1)
  const from = ontem.toISOString().split('T')[0] + 'T00:00:00Z'
  const to   = ontem.toISOString().split('T')[0] + 'T23:59:59Z'

  const url = `${PULSE_API}/api/v1/events?product=${PRODUCT}&from=${from}&to=${to}&limit=1000`
  const response = UrlFetchApp.fetch(url, {
    headers: { 'X-Service-Token': PULSE_TOKEN }
  })

  const { data } = JSON.parse(response.getContentText())

  if (sheet.getLastRow() === 0) {
    sheet.appendRow(['id', 'evento', 'user_id', 'tenant_id', 'valor', 'data'])
  }

  data.forEach(evt => {
    sheet.appendRow([
      evt.id,
      evt.event,
      evt.user_id || '',
      evt.tenant_id || '',
      evt.properties?.value || '',
      evt.created_at,
    ])
  })

  Logger.log(`Importados ${data.length} eventos de ${from}`)
}
```

Para configurar o token no Apps Script: **Arquivo → Propriedades do projeto → Adicionar propriedade → `PULSE_TOKEN`**.

---

## Conectar em ferramentas de BI (Metabase, Power BI, Looker)

Use o endpoint de eventos com paginação em um script de ETL que popula um banco de dados local:

```python
# etl_pulse.py — rodar via cron diariamente
import requests, psycopg2, json, os
from datetime import datetime, timedelta

PULSE_TOKEN = os.environ['PULSE_TOKEN']
PULSE_API = 'https://pulse-api.institutoitinerante.com.br'
DB_URL = os.environ['DATABASE_URL']

def fetch_events(product: str, from_date: str, to_date: str):
    cursor = None
    while True:
        params = {'product': product, 'from': from_date, 'to': to_date, 'limit': 1000}
        if cursor:
            params['cursor'] = cursor
        resp = requests.get(
            f'{PULSE_API}/api/v1/events',
            headers={'X-Service-Token': PULSE_TOKEN},
            params=params,
            timeout=30
        )
        resp.raise_for_status()
        body = resp.json()
        yield from body['data']
        if not body['meta']['has_more']:
            break
        cursor = body['meta']['next_cursor']

def run():
    conn = psycopg2.connect(DB_URL)
    cur = conn.cursor()
    yesterday = (datetime.utcnow() - timedelta(days=1)).strftime('%Y-%m-%dT00:00:00Z')
    today     = datetime.utcnow().strftime('%Y-%m-%dT00:00:00Z')

    for evt in fetch_events('minha-app', yesterday, today):
        cur.execute("""
            INSERT INTO pulse_events (id, product, event, user_id, properties, created_at)
            VALUES (%s, %s, %s, %s, %s, %s)
            ON CONFLICT (id) DO NOTHING
        """, (evt['id'], evt['product'], evt['event'], evt.get('user_id'),
              json.dumps(evt.get('properties', {})), evt['created_at']))
    
    conn.commit()
    cur.close()
    conn.close()

if __name__ == '__main__':
    run()
```

Após popular a tabela `pulse_events`, conecte o Metabase diretamente ao banco de dados.

---

## Próximos passos

- [04-ab-test-interpretar-p-value.md](./04-ab-test-interpretar-p-value.md) — analisar resultados de experimentos
- [06-limite-eventos-o-que-fazer.md](./06-limite-eventos-o-que-fazer.md) — quota e otimização antes de exportar
