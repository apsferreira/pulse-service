# A/B Testing no Pulse

Referência técnica interna do Pulse: [`pulse-service/docs/guides/ab-testing.md`](../../../../pulse-service/docs/guides/ab-testing.md)

---

## O que é um A/B test no Pulse?

Um A/B test é uma feature flag com variantes nomeadas. O Pulse distribui usuários entre as variantes de forma **determinística**: o mesmo `user_id` sempre recebe a mesma variante, garantindo experiência consistente entre sessões e dispositivos.

```
flag: cta_checkout
  ├── variante "control"    → botão "Comprar"      (50% dos usuários)
  └── variante "treatment"  → botão "Assinar agora" (50% dos usuários)
```

A distribuição é feita por hash consistente do `user_id` + `flag_key`. Não há estado salvo por usuário no banco — o mesmo cálculo sempre produz o mesmo resultado.

---

## Criar uma flag com variantes

**Auth:** `X-Service-Token` (token admin do serviço)
**Base URL (prod):** `https://pulse-api.institutoitinerante.com.br`

```bash
curl -X POST https://pulse-api.institutoitinerante.com.br/api/v1/flags \
  -H "X-Service-Token: $SERVICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "cta_checkout",
    "product": "socialmake",
    "description": "A/B test: botão de CTA — comprar vs assinar agora",
    "enabled": true,
    "type": "multivariate",
    "variants": [
      { "name": "control",   "weight": 50 },
      { "name": "treatment", "weight": 50 }
    ]
  }'
```

Campos obrigatórios:

| Campo | Tipo | Descrição |
|-------|------|-----------|
| `name` | string | Chave da flag (snake_case, único por produto) |
| `product` | string | Slug do projeto no Pulse |
| `type` | string | `"multivariate"` para A/B, `"boolean"` para liga/desliga |
| `variants[].name` | string | Identificador da variante |
| `variants[].weight` | int | Peso relativo (soma deve ser 100) |

---

## Como o SDK determina a variante do usuário

O cálculo é local — não há round-trip ao banco por usuário:

```
variante = hash(user_id + flag_key) mod soma_dos_pesos
```

Consequências práticas:
- Mesmo usuário, mesma variante sempre (determinístico)
- Usuário não autenticado pode usar `session_id` como `user_id`
- Alterar os pesos das variantes pode mover usuários de grupo — evitar isso após o teste começar

---

## Avaliar variante via API (SDK)

**Auth:** `X-Project-Key` (api_key do projeto)

```bash
curl "https://pulse-api.institutoitinerante.com.br/flags?user_id=user-123&project_id=<uuid>" \
  -H "X-Project-Key: pulse_sm_xxxxxxxxxxxx"
```

Resposta:

```json
{
  "flags": {
    "cta_checkout": {
      "enabled": true,
      "variant": "treatment"
    }
  }
}
```

---

## Exemplo real: botão "Comprar" vs "Assinar agora"

**Contexto:** SocialMake quer testar qual CTA gera mais conversões no plano Growth.

### 1. Criar a flag

```bash
curl -X POST https://pulse-api.institutoitinerante.com.br/api/v1/flags \
  -H "X-Service-Token: $SERVICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "cta_plano_growth",
    "product": "socialmake",
    "description": "CTA plano Growth: comprar vs assinar agora",
    "enabled": true,
    "type": "multivariate",
    "variants": [
      { "name": "control",   "weight": 50 },
      { "name": "treatment", "weight": 50 }
    ]
  }'
```

### 2. Usar no frontend (React)

```tsx
import { usePulse } from '@pulse/react'

function CTAButton({ userId }: { userId: string }) {
  const { getVariant } = usePulse()
  const variant = getVariant('cta_plano_growth', userId)

  return (
    <button onClick={handleCheckout}>
      {variant === 'treatment' ? 'Assinar agora' : 'Comprar'}
    </button>
  )
}
```

### 3. Registrar o evento de conversão

Quando o usuário converte, rastrear com o mesmo `user_id`:

```bash
curl -X POST https://pulse-api.institutoitinerante.com.br/api/v1/ingest \
  -H "X-Project-Key: pulse_sm_xxxxxxxxxxxx" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": "checkout.completed",
    "product": "socialmake",
    "user_id": "user-123",
    "properties": {
      "flag": "cta_plano_growth",
      "variant": "treatment",
      "plan": "growth",
      "value": 197.00
    }
  }'
```

---

## Ver resultados no dashboard

Acesse o dashboard em `pulse.institutoitinerante.com.br` (ou `localhost:5173` em dev).

Navegue em: **Projetos → socialmake → A/B Testing → cta_plano_growth**

O dashboard exibe:
- Distribuição de usuários por variante
- Taxa de conversão por variante
- P-value (significância estatística) — considerar resultado válido quando p < 0.05
- Uplift relativo entre variantes

Para interpretar o p-value e saber quando encerrar o teste: [`pulse-service/docs/guides/04-ab-test-interpretar-p-value.md`](../../../../pulse-service/docs/guides/04-ab-test-interpretar-p-value.md)

---

## Encerrar o teste

Após atingir significância estatística, promova a variante vencedora:

```bash
# Fixar 100% no treatment (vencedor)
curl -X PUT https://pulse-api.institutoitinerante.com.br/api/v1/flags/cta_plano_growth \
  -H "X-Service-Token: $SERVICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "variants": [
      { "name": "treatment", "weight": 100 }
    ]
  }'
```

Depois remova o código de A/B do frontend e mantenha só o texto vencedor.

---

## Produtos IIT com A/B tests ativos

| Produto | Flag | Variantes |
|---------|------|-----------|
| SocialMake | `cta_plano_growth` | control / treatment |
| Catraca Virtual | `checkin_flow_v2` | legacy / streamlined |
| Librinho | `isbn_scan_mode` | camera / manual |

---

Ver tambem: [`eventos.md`](eventos.md) — como rastrear eventos de conversão para medir os resultados.
