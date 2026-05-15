# Pulse — Getting Started

Bem-vindo ao Pulse! Este guia leva você do zero até feature flags e event tracking funcionando em minutos.

---

## Pré-requisitos

- Instância Pulse rodando (ver [self-hosted.md](../self-hosted.md)) ou acesso a uma instância existente
- `SERVICE_TOKEN` (visível no `.env` ou nos secrets do Kubernetes)
- Node.js 18+ (SDK JavaScript), Go 1.24+ (SDK Go) ou Flutter 3.x+ (SDK Flutter)

---

## 1. Criar um projeto

Cada sistema/produto que vai usar o Pulse precisa de um projeto próprio. O projeto tem uma `api_key` usada nos SDKs.

```bash
export PULSE_API=https://seu-pulse.seudominio.com
export SERVICE_TOKEN=seu-service-token-aqui

curl -X POST $PULSE_API/api/v1/projects \
  -H "X-Service-Token: $SERVICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Meu App",
    "slug": "meu-app",
    "plan": "free"
  }'
```

**Resposta:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Meu App",
  "slug": "meu-app",
  "api_key": "pulse_xxxxxxxxxxxxxxxxxxxxxxxx",
  "plan": "free"
}
```

Guarde o `api_key` — é ele que vai nas variáveis de ambiente do seu produto.

> Crie **um projeto por sistema**. Não compartilhe `api_key` entre produtos diferentes — isso mistura métricas e flags.

---

## 2. Criar uma feature flag

```bash
export PROJECT_ID=550e8400-e29b-41d4-a716-446655440000

curl -X POST $PULSE_API/api/v1/flags \
  -H "X-Service-Token: $SERVICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "key": "nova_homepage",
    "description": "Nova versão da homepage",
    "enabled": false,
    "rollout_percentage": 0,
    "project_id": "'$PROJECT_ID'"
  }'
```

---

## 3. Integrar o SDK JavaScript (frontend)

### Instalação

```bash
npm install @pulse/js
```

### Uso básico

```typescript
import { PulseSDK } from '@pulse/js'

// Inicializar uma vez (ex: main.tsx ou App.tsx)
PulseSDK.init({
  projectKey: import.meta.env.VITE_PULSE_KEY,    // api_key do projeto
  apiUrl: import.meta.env.VITE_PULSE_URL,        // URL da instância Pulse
})

// Verificar flag
if (await PulseSDK.getFeatureFlag('nova_homepage', false)) {
  renderV2()
} else {
  renderV1()
}

// Rastrear evento
PulseSDK.track('page_viewed', { page: '/home' })
```

### Variáveis de ambiente (Vite)

```dotenv
# .env.local
VITE_PULSE_KEY=pulse_xxxxxxxxxxxxxxxxxxxxxxxx
VITE_PULSE_URL=https://seu-pulse.seudominio.com
```

### Hook React

```typescript
import { useFeatureFlag } from '@pulse/js/react'

function Homepage() {
  const isV2 = useFeatureFlag('nova_homepage')
  return isV2 ? <HomepageV2 /> : <HomepageV1 />
}
```

---

## 4. Integrar o SDK Flutter (mobile)

```yaml
# pubspec.yaml
dependencies:
  pulse_flutter:
    path: ../pulse-service/sdk/flutter  # dev local
```

```dart
// main.dart
import 'package:flutter/foundation.dart';
import 'package:pulse_flutter/pulse_flutter.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await Pulse.init(
    projectKey: 'meu-app',
    debug: kDebugMode,
  );
  runApp(const MyApp());
}

// Rastrear evento
Pulse.instance.track('page_viewed', properties: {'page': '/home'});

// Feature flag
final enabled = await Pulse.instance.isEnabled('nova_homepage');
```

Consulte [sdk-reference.md](./sdk-reference.md#flutter--dart) para a referência completa do SDK Flutter.

---

## 5. Integrar o SDK Go (backend)

```bash
go get github.com/seu-org/pulse-service/sdk/go
```

```go
import pulse "github.com/seu-org/pulse-service/sdk/go"

client, _ := pulse.NewClient(os.Getenv("PULSE_URL"))

// Feature flag
enabled, err := client.IsEnabled(ctx, "nova_feature")
if err != nil {
    enabled = false // fail-open: nunca quebrar o sistema por causa do Pulse
}

// Evento
client.TrackEvent(ctx, pulse.TrackEventRequest{
    EventName:  "user_signup",
    UserID:     &userID,
    Properties: map[string]interface{}{"plan": "free"},
})
```

```bash
# Variável de ambiente
PULSE_URL=https://seu-pulse.seudominio.com
```

---

## 6. Padrões de uso

### Flag guard — rollout seguro

```typescript
// O mesmo usuário sempre cai no mesmo bucket (hash por userId)
// rollout_percentage: 10 → 10% dos usuários veem a nova versão
if (await PulseSDK.getFeatureFlag('nova_checkout', false)) {
  return <CheckoutV2 />
}
return <CheckoutV1 />
```

### Kill switch — rollback sem redeploy

```typescript
// Flag 'maintenance_mode' desativa o sistema instantaneamente pelo dashboard
// sem precisar de deploy ou rollback de código
const inMaintenance = await PulseSDK.getFeatureFlag('maintenance_mode', false)
if (inMaintenance) {
  return <MaintenancePage />
}
```

### Targeting por usuário

```bash
# Forçar flag ativa só para usuários beta
curl -X PUT $PULSE_API/api/v1/flags/nova_feature \
  -H "X-Service-Token: $SERVICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "targeting": {
      "user_ids": ["user-123", "user-456"]
    }
  }'
```

---

## 7. Monitorar no dashboard

1. Acesse `https://seu-pulse.seudominio.com` (porta `3016` em self-hosted)
2. Selecione o projeto no seletor do topo
3. Veja flags ativas, rollout % e última modificação
4. Edite em tempo real — mudanças propagam em <10s via Redis

---

## 8. Próximos passos

- [sdk-reference.md](./sdk-reference.md) — referência completa de todos os SDKs
- [ab-testing.md](./ab-testing.md) — criar e analisar testes A/B
- [self-hosted.md](../self-hosted.md) — instalar e operar sua própria instância
- [validation-checklist.md](./validation-checklist.md) — validar que tudo funciona após setup

---

## Troubleshooting

**SDK não conecta na API**
```
Verificar: CORS (apiUrl deve ser acessível do browser), SERVICE_TOKEN correto,
           instância Pulse rodando (GET /health deve retornar {"status":"ok"})
```

**Flag alterada no dashboard mas app não reflete**
```
Cache JavaScript padrão: 60s. Para debug, reduza: PulseSDK.init({ ..., flushInterval: 1000 })
Ou chame PulseSDK.flush() manualmente.
```

**Eventos não aparecem no dashboard**
```
Verificar: X-Project-Key correto (api_key do projeto, não o SERVICE_TOKEN)
           POST /api/v1/events retorna 200?
           Plano Free tem limite de 10k eventos/mês.
```
