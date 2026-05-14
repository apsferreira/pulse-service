# Pulse — Checklist de Validação Funcional

Use este guia após subir uma instância self-hosted para confirmar que tudo está funcionando antes de integrar seus produtos.

---

## Variáveis de referência

```bash
export PULSE_API=http://localhost:3015      # ou sua URL pública
export SERVICE_TOKEN=seu-service-token
```

---

## Fase 1 — Infraestrutura (2 min)

```bash
# 1. Backend responde
curl $PULSE_API/health
# Esperado: {"status":"ok","service":"pulse-service"}

# 2. Frontend carrega
curl -I http://localhost:3016
# Esperado: HTTP/1.1 200 OK

# 3. Containers rodando
docker compose -f docker-compose.selfhosted.yml ps
# Esperado: pulse-backend (healthy), pulse-frontend (running),
#           pulse-postgres (healthy), pulse-redis (healthy)
```

---

## Fase 2 — Auth (2 min)

### Modo local (`PULSE_AUTH_MODE=local`)

```bash
# Criar admin (só funciona se não há usuários)
curl -X POST $PULSE_API/api/v1/auth/setup \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@empresa.com","password":"senha123","name":"Admin"}'
# Esperado: {"token":"eyJ...","user":{"email":"admin@empresa.com"}}

# Login
TOKEN=$(curl -s -X POST $PULSE_API/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@empresa.com","password":"senha123"}' | python3 -c "import sys,json; print(json.load(sys.stdin)['token'])")

# Validar token
curl $PULSE_API/api/v1/auth/me \
  -H "Authorization: Bearer $TOKEN"
# Esperado: {"id":"...","email":"admin@empresa.com","role":"admin"}
```

---

## Fase 3 — Projetos (2 min)

```bash
# Criar projeto
curl -X POST $PULSE_API/api/v1/projects \
  -H "X-Service-Token: $SERVICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Meu App","slug":"meu-app","plan":"free"}'
# Esperado: {"id":"...","api_key":"pulse_...","slug":"meu-app"}

export PROJECT_ID=<id-retornado>
export PROJECT_KEY=<api_key-retornado>

# Listar projetos
curl $PULSE_API/api/v1/projects \
  -H "X-Service-Token: $SERVICE_TOKEN"
# Esperado: {"projects":[{"slug":"meu-app",...}],"total":1}
```

---

## Fase 4 — Feature Flags (3 min)

```bash
# Criar flag
curl -X POST $PULSE_API/api/v1/flags \
  -H "X-Service-Token: $SERVICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"key\":\"nova_feature\",\"description\":\"Teste\",\"enabled\":false,\"rollout_percentage\":0,\"project_id\":\"$PROJECT_ID\"}"
# Esperado: {"id":"...","key":"nova_feature","enabled":false}

# Avaliar flag (como se fosse o SDK — desativada)
curl "$PULSE_API/flags?user_id=user-123&project_id=$PROJECT_ID" \
  -H "X-Project-Key: $PROJECT_KEY"
# Esperado: {"flags":{"nova_feature":false}}

# Ativar a flag
curl -X PUT $PULSE_API/api/v1/flags/nova_feature \
  -H "X-Service-Token: $SERVICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"enabled":true,"rollout_percentage":100}'
# Esperado: {"enabled":true}

# Avaliar novamente (agora ativa)
curl "$PULSE_API/flags?user_id=user-123&project_id=$PROJECT_ID" \
  -H "X-Project-Key: $PROJECT_KEY"
# Esperado: {"flags":{"nova_feature":true}}
```

**Ponto de validação crítico:** o valor muda de `false` para `true` sem nenhum redeploy. Isso é o core value do produto.

---

## Fase 5 — Event Tracking (2 min)

```bash
# Enviar evento
curl -X POST $PULSE_API/api/v1/events \
  -H "Content-Type: application/json" \
  -H "X-Project-Key: $PROJECT_KEY" \
  -d "{
    \"event_name\": \"feature_tested\",
    \"user_id\": \"user-123\",
    \"environment\": \"production\",
    \"properties\": {
      \"flag\": \"nova_feature\",
      \"valor\": true
    }
  }"
# Esperado: {"id":"...","status":"ok"} ou HTTP 200

# Confirmar que chegou (endpoint admin)
curl "$PULSE_API/api/v1/events?limit=5" \
  -H "X-Service-Token: $SERVICE_TOKEN"
# Esperado: {"items":[{"event_name":"feature_tested",...}]}
```

---

## Fase 6 — Dashboard visual (1 min)

1. Abrir `http://localhost:3016`
2. Fazer login com as credenciais criadas na Fase 2
3. Selecionar o projeto no seletor do topo
4. Verificar que a flag `nova_feature` aparece como **Ativa**
5. Desativar a flag pelo toggle e verificar que muda imediatamente

---

## Resumo: o que cada fase valida

| Fase | O que valida |
|------|-------------|
| 1 — Infraestrutura | Containers, portas, healthchecks |
| 2 — Auth | Login funcional, token JWT válido |
| 3 — Projetos | Isolamento por produto, geração de api_key |
| 4 — Feature Flags | Criar, ativar, avaliar via SDK endpoint |
| 5 — Events | Ingest de eventos, visibilidade no dashboard |
| 6 — Dashboard | UI funcional, toggle em tempo real |

---

## Pendências conhecidas (self-hosted)

| Item | Status | Detalhe |
|------|--------|---------|
| migrations automáticas no startup | ❌ Não implementado | Aplicar manualmente via `psql` |
| `POST /api/v1/projects` via API | ⚠️ 500 em alguns deploys | Usar SQL direto como workaround (ver nota abaixo) |
| Frontend: `VITE_*` em runtime | ⚠️ Baked no build | Variáveis precisam estar definidas em build-time |
| Webhook de notificação | ❌ Não documentado | Endpoint existe mas sem guia |

### Workaround: criar projeto via SQL

Se o `POST /api/v1/projects` retornar 500 (bug no INSERT de `health_config`):

```sql
INSERT INTO projects (id, name, slug, api_key, description, plan, data_retention_days, is_active)
VALUES (
  gen_random_uuid(),
  'Meu App',
  'meu-app',
  'pulse_' || encode(gen_random_bytes(32), 'hex'),
  'Descrição do projeto',
  'free',
  90,
  true
);
```

---

## Teste end-to-end com SDK real (JavaScript)

```html
<!-- teste.html — abrir no browser apontando para sua instância -->
<!DOCTYPE html>
<html>
<head>
  <script src="http://localhost:3015/sdk/pulse.js"></script>
</head>
<body>
  <div id="resultado">Carregando...</div>
  <script>
    PulseSDK.init({
      projectKey: 'COLOQUE_SUA_PROJECT_KEY_AQUI',
      apiUrl: 'http://localhost:3015',
      debug: true,
    })

    PulseSDK.getFeatureFlag('nova_feature', false).then(enabled => {
      document.getElementById('resultado').textContent =
        `nova_feature = ${enabled}`
    })

    PulseSDK.track('sdk_test', { origem: 'validation_checklist' })
  </script>
</body>
</html>
```

Altere a flag no dashboard e recarregue a página — o valor deve mudar sem nenhum deploy.
