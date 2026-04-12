# pulse-iit

SDK Python server-side para o pulse-service IIT.

Buffer thread-safe com flush automatico a cada 5s ou ao atingir 50 eventos — ideal para APIs e workers backend.

## Instalacao

```bash
pip install pulse-iit
```

Ou diretamente do repositorio:

```bash
pip install git+https://github.com/apsferreira/pulse-service.git#subdirectory=sdk/python
```

## Uso rapido — FastAPI

```python
from contextlib import asynccontextmanager
from fastapi import FastAPI, Request
from pulse_iit import PulseClient
import os

pulse = PulseClient(
    project_id=os.environ["PULSE_PROJECT_ID"],
    api_url=os.environ["PULSE_API_URL"],
    # flush_interval=5.0,  # default: 5s
    # max_batch_size=50,   # default: 50 eventos
)

@asynccontextmanager
async def lifespan(app: FastAPI):
    yield
    # Graceful shutdown — garante que eventos em buffer sao enviados
    pulse.shutdown()

app = FastAPI(lifespan=lifespan)

@app.middleware("http")
async def tracking_middleware(request: Request, call_next):
    response = await call_next(request)
    pulse.track("http_request", {
        "method": request.method,
        "path": request.url.path,
        "status": response.status_code,
    })
    return response

@app.post("/checkout")
async def checkout(user_id: str):
    pulse.identify(user_id, {"plan": "pro"})
    pulse.track("checkout_started", {"cart_value": 99.9}, user_id=user_id)
    return {"ok": True}
```

## Uso rapido — Django

```python
# settings.py ou apps.py
import os
from pulse_iit import PulseClient

pulse = PulseClient(
    project_id=os.environ["PULSE_PROJECT_ID"],
    api_url=os.environ["PULSE_API_URL"],
)

# views.py
from myproject.settings import pulse

def checkout_view(request):
    user_id = str(request.user.id)
    pulse.identify(user_id, {"username": request.user.username})
    pulse.track("checkout_started", {"path": request.path}, user_id=user_id)
    return JsonResponse({"ok": True})
```

## Graceful shutdown (SIGTERM)

```python
import signal
import sys

def handle_sigterm(sig, frame):
    pulse.shutdown()  # flush final antes de encerrar
    sys.exit(0)

signal.signal(signal.SIGTERM, handle_sigterm)
signal.signal(signal.SIGINT, handle_sigterm)
```

## API

### `PulseClient(project_id, api_url, **kwargs)`

| Parametro | Tipo | Default | Descricao |
|-----------|------|---------|-----------|
| `project_id` | `str` | obrigatorio | ID do projeto no pulse-service |
| `api_url` | `str` | obrigatorio | URL base do pulse-service |
| `flush_interval` | `float` | `5.0` | Intervalo de flush em segundos |
| `max_batch_size` | `int` | `50` | Eventos por batch antes de flush forcado |
| `request_timeout` | `float` | `10.0` | Timeout por requisicao HTTP em segundos |

### `pulse.track(event, properties=None, user_id=None)`

Adiciona evento de tracking ao buffer. Nao bloqueia.

### `pulse.identify(user_id, traits=None)`

Adiciona evento de identificacao ao buffer. Nao bloqueia.

### `pulse.flush()`

Envia todos os eventos em buffer imediatamente. Bloqueia ate completar.

### `pulse.shutdown()`

Flush final + encerra a thread de flush automatico. Chamar no encerramento da aplicacao.

## Requisitos

- Python >= 3.9
- Sem dependencias externas (usa apenas stdlib: `urllib`, `threading`, `json`)

## Endpoint esperado no pulse-service

`POST /api/v1/events/batch`

```json
{
  "project_id": "proj_xxx",
  "events": [
    {
      "type": "track",
      "event": "checkout_started",
      "user_id": "user_123",
      "properties": { "cart_value": 99.9 },
      "project_id": "proj_xxx",
      "timestamp": "2026-04-11T10:00:00.000000+00:00"
    }
  ]
}
```
