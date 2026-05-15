---
title: "Flag não ativa — debug"
description: "Checklist sistemático para diagnosticar por que uma feature flag não está sendo avaliada como true no seu app."
---

# Flag não ativa — debug

Você habilitou a flag no dashboard, deployou o código, mas a feature não aparece. Este guia cobre as causas mais comuns em ordem de probabilidade.

---

## Checklist rápido

Execute cada item em sequência. A maioria dos casos é resolvida nos primeiros três pontos.

---

### 1. A flag está habilitada no dashboard?

Acesse **Pulse → Flags → [nome da flag]** e confirme:

- `enabled: true` (toggle ligado)
- `rollout_percentage > 0`

Se `rollout_percentage` for `0`, **a flag está tecnicamente habilitada mas nenhum usuário a recebe.** Aumente para pelo menos `1`.

---

### 2. O `SERVICE_TOKEN` está correto?

```bash
# Teste direto na API — deve retornar 200, não 401
curl -s -o /dev/null -w "%{http_code}" \
  https://pulse-api.institutoitinerante.com.br/api/v1/flags?product=minha-app \
  -H "X-Service-Token: seu-token-aqui"
```

Se retornar `401`: token errado ou expirado. Gere um novo em **Settings → API → Regenerar Token**.

Se retornar `403`: o token não tem permissão para o produto/projeto solicitado.

---

### 3. O `product` (slug do projeto) bate com o da flag?

A flag é criada com um `product` específico. Se o SDK for inicializado com um `product` diferente, a flag não será retornada.

```typescript
// Flag criada com product: "libri"
// SDK inicializado com product: "library"  ← errado, não vai retornar a flag

const pulse = new PulseSDK({
  product: 'libri',  // deve ser idêntico ao slug no dashboard
  // ...
})
```

Confirme o slug exato em **Settings → Projeto → Slug**.

---

### 4. O SDK foi inicializado antes da chamada `isEnabled`?

`getFlags()` é assíncrono. Chamar `isEnabled()` antes de `await getFlags()` retorna sempre `false`.

```typescript
// Errado — isEnabled chamado antes de getFlags resolver
const pulse = new PulseSDK({ ... })
if (pulse.isEnabled('minha_flag')) { ... }  // sempre false

// Correto
const pulse = new PulseSDK({ ... })
await pulse.getFlags()
if (pulse.isEnabled('minha_flag')) { ... }
```

Em React com o `PulseProvider`, use o estado de loading:

```typescript
const flagValue = useFeatureFlag('minha_flag')
if (flagValue === undefined) return <Skeleton />  // ainda carregando
```

---

### 5. O `rollout_percentage` cobre o seu `userId`?

O rollout usa hash consistente do `userId`. Para um usuário específico, o hash pode cair fora da faixa do rollout — isso é intencional.

**Para forçar um grupo específico sem alterar o backend**, use o parâmetro de URL:

```
# Forçar flag ON para quem acessar esse link
https://seuapp.com/pagina?pulse_override[nome-da-flag]=on

# Forçar flag OFF
https://seuapp.com/pagina?pulse_override[nome-da-flag]=off

# Forçar uma variante específica de A/B test
https://seuapp.com/pagina?pulse_override[nome-da-flag]=challenger
```

O SDK JS lê esse parâmetro automaticamente — funciona sem nenhuma configuração extra. Útil para QA, demos e validar os dois grupos de um A/B test.

**Para checar em qual bucket um `userId` específico cai**, use o endpoint de preview:

```bash
curl "https://pulse-api.institutoitinerante.com.br/api/v1/flags/nome-da-flag/preview?user_id=user-123" \
  -H "X-Service-Token: $SERVICE_TOKEN"
# → {"bucket": 34, "rollout_percentage": 50, "in_rollout": true, "reason": "rollout"}
```

Ou use o painel **Testar avaliação** diretamente no modal da flag no dashboard.

---

### 6. O `userId` está sendo passado ao SDK?

Sem `userId`, o Pulse usa o `tenantId` ou, como fallback, um ID de sessão aleatório. Rollouts baseados em usuário ficam inconsistentes.

```typescript
// Sem userId — comportamento imprevisível no rollout
const pulse = new PulseSDK({ product: 'minha-app', serviceToken: '...' })

// Com userId — rollout determinístico
const pulse = new PulseSDK({
  product: 'minha-app',
  serviceToken: '...',
  userId: currentUser.id,  // passe sempre que o usuário estiver autenticado
})
```

---

### 7. O cache local está servindo uma versão antiga?

O SDK cacheia as flags por `cacheTime` (padrão: 60 segundos). Se você habilitou a flag agora, aguarde o TTL ou force um reload:

```typescript
// Forçar recarga imediata (ignora cache)
await pulse.getFlags({ forceRefresh: true })
```

Ou reduza o `cacheTime` durante o debug:

```typescript
const pulse = new PulseSDK({
  // ...
  cacheTime: 1000,  // 1s — apenas para debug, não use em produção
})
```

---

### 8. Usar o audit log para investigar

O audit log registra toda avaliação de flag, incluindo por que foi negada.

```bash
# Ver últimas 20 avaliações da flag
curl "https://pulse-api.institutoitinerante.com.br/api/v1/flags/audit?product=minha-app&flag=minha_flag&limit=20" \
  -H "X-Service-Token: $SERVICE_TOKEN"
```

**Resposta de exemplo:**

```json
[
  {
    "timestamp": "2026-04-11T10:00:00Z",
    "flag": "minha_flag",
    "user_id": "user-123",
    "result": false,
    "reason": "rollout_miss",   // usuário não caiu no percentual
    "rollout_percentage": 10,
    "user_bucket": 73
  },
  {
    "timestamp": "2026-04-11T09:55:00Z",
    "flag": "minha_flag",
    "user_id": "user-456",
    "result": true,
    "reason": "rollout_hit",
    "rollout_percentage": 10,
    "user_bucket": 8
  }
]
```

| `reason` | Causa | Solução |
|----------|-------|---------|
| `flag_disabled` | Toggle desligado | Habilitar no dashboard |
| `rollout_miss` | User bucket fora do range | Aumentar percentual ou adicionar override |
| `token_invalid` | SERVICE_TOKEN incorreto | Regenerar token |
| `product_mismatch` | Slug do produto diferente | Corrigir inicialização do SDK |
| `flag_not_found` | Flag não existe neste produto | Verificar nome e produto da flag |
| `quota_exceeded` | Plano Free atingiu limite | Fazer upgrade ou aguardar reset mensal |

---

## Diagnóstico em um comando

```bash
# Script de diagnóstico rápido
PULSE_API=https://pulse-api.institutoitinerante.com.br
TOKEN=seu-token
PRODUCT=minha-app
FLAG=minha_flag

echo "=== 1. Autenticação ==="
curl -s -o /dev/null -w "HTTP %{http_code}\n" \
  "$PULSE_API/api/v1/flags?product=$PRODUCT" \
  -H "X-Service-Token: $TOKEN"

echo "=== 2. Flag existe? ==="
curl -s "$PULSE_API/api/v1/flags?product=$PRODUCT" \
  -H "X-Service-Token: $TOKEN" | \
  python3 -c "import sys,json; flags=json.load(sys.stdin); \
  f=[x for x in flags if x['name']=='$FLAG']; \
  print(f[0] if f else 'FLAG NAO ENCONTRADA')"

echo "=== 3. Últimas avaliações ==="
curl -s "$PULSE_API/api/v1/flags/audit?product=$PRODUCT&flag=$FLAG&limit=5" \
  -H "X-Service-Token: $TOKEN" | python3 -m json.tool
```

---

## Ainda com problema?

Se nenhum item do checklist resolveu:

1. Abra uma issue em https://github.com/apsferreira/pulse-service/issues com o output do script de diagnóstico acima.
2. Inclua o `project_id` (não o token), o nome da flag e o `userId` de teste.
3. Email: support@institutoitinerante.com.br
