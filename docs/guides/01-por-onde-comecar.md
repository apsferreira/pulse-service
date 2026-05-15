---
title: "Por onde começar com o Pulse"
description: "Fluxo completo do zero ao primeiro evento rastreado e primeira feature flag ativa."
---

# Por onde começar com o Pulse

Este guia cobre o fluxo mínimo para você rastrear seu primeiro evento e ativar sua primeira feature flag em menos de 15 minutos.

---

## Visão geral do fluxo

```
Criar conta → Criar projeto → Instalar SDK → Primeiro evento → Primeira flag
```

Cada etapa é independente — você pode parar em qualquer ponto e continuar depois.

---

## Passo 1 — Criar conta

Acesse [pulse.institutoitinerante.com.br](https://pulse.institutoitinerante.com.br) e crie sua conta via auth-service. O plano **Free** não exige cartão de crédito e inclui até 10.000 eventos/mês.

Após o login, você cai direto na tela de criação de projeto.

---

## Passo 2 — Criar projeto

No dashboard, clique em **Novo Projeto** e preencha:

| Campo | Exemplo |
|-------|---------|
| Nome | `Minha App` |
| Slug | `minha-app` (gerado automaticamente) |
| Ambiente | `production` ou `staging` |

Ao salvar, o Pulse gera seu `project_id` e `SERVICE_TOKEN`. Guarde o token — ele é exibido apenas uma vez. Se perder, vá em **Settings → API → Regenerar Token**.

---

## Passo 3 — Instalar o SDK

Para aplicações JavaScript/TypeScript:

```bash
npm install @pulse/js
```

Para Go (backend):

```bash
go get github.com/apsferreira/pulse-service/sdk/go
```

Detalhes completos de instalação, incluindo React hook, Next.js SSR guard e Astro: [02-instalar-sdk-js.md](./02-instalar-sdk-js.md).

---

## Passo 4 — Rastrear o primeiro evento

Com o SDK instalado, inicialize e dispare um evento:

```typescript
import { PulseSDK } from '@pulse/js'

const pulse = new PulseSDK({
  apiUrl: 'https://pulse.institutoitinerante.com.br',
  product: 'minha-app',       // slug do projeto
  serviceToken: process.env.PULSE_SERVICE_TOKEN,
  userId: 'user-123',         // opcional, mas recomendado
})

// Rastrear evento
await pulse.track('signup_completed', {
  plan: 'free',
  source: 'landing_page',
})
```

Vá ao dashboard → **Eventos** e confirme que o evento aparece na lista. Se não aparecer em 30 segundos, consulte [03-debug-flag-nao-ativa.md](./03-debug-flag-nao-ativa.md).

---

## Passo 5 — Criar a primeira flag

Via API:

```bash
export PULSE_API=https://pulse.institutoitinerante.com.br
export SERVICE_TOKEN=seu-token-aqui

curl -X POST $PULSE_API/api/v1/flags \
  -H "X-Service-Token: $SERVICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "nova_homepage",
    "product": "minha-app",
    "description": "Teste do novo layout da homepage",
    "enabled": true,
    "rollout_percentage": 10
  }'
```

Ou pelo dashboard: **Flags → Nova Flag → preencher campos → Salvar**.

No código, consumir a flag:

```typescript
await pulse.getFlags()

if (pulse.isEnabled('nova_homepage')) {
  return <HomepageV2 />
}
return <HomepageV1 />
```

O mesmo usuário (`userId`) sempre recebe a mesma experiência — o rollout usa hash consistente.

---

## Próximos passos

| Guia | Quando usar |
|------|-------------|
| [02-instalar-sdk-js.md](./02-instalar-sdk-js.md) | Configuração avançada: React hook, Next.js, Astro |
| [03-debug-flag-nao-ativa.md](./03-debug-flag-nao-ativa.md) | Flag habilitada mas não aparece no app |
| [04-ab-test-interpretar-p-value.md](./04-ab-test-interpretar-p-value.md) | Rodar experimentos com análise estatística |
| [05-qual-plano-escolher.md](./05-qual-plano-escolher.md) | Comparar Free / Starter / Growth / Pro |
| [10-integrar-pulse-ci-cd.md](./10-integrar-pulse-ci-cd.md) | Usar flags como gates de deploy em CI/CD |

---

## Suporte

- **Email:** support@institutoitinerante.com.br
- **Issues:** https://github.com/apsferreira/pulse-service/issues
