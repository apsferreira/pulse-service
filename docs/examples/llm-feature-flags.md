# BKL-1002: A/B Test de Prompts com Feature Flags

Feature flags com `prompt_variant` permitem testar diferentes prompts de LLM sem deploy.
O backend consulta o Pulse para decidir qual prompt usar para cada usuário.

## Configurar uma flag com prompt_variant

```bash
curl -X PUT https://pulse.institutoitinerante.com.br/api/v1/flags/attend-system-prompt \
  -H "X-Service-Token: $SERVICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "description": "A/B test do system prompt do Brio (attend-service)",
    "enabled": true,
    "rollout_percentage": 50,
    "prompt_variant": {
      "system_prompt": "Você é o Brio, assistente de atendimento da IIT. Responda de forma concisa e empática.",
      "model": "gemini-2.5-flash-preview-04-17",
      "temperature": 0.7,
      "max_tokens": 512
    }
  }'
```

## Avaliar a flag em Go (backend attend-service)

```go
package llmconfig

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
)

// PromptVariant é a configuração de LLM retornada pelo Pulse.
type PromptVariant struct {
    SystemPrompt string   `json:"system_prompt"`
    Model        string   `json:"model"`
    Temperature  *float64 `json:"temperature,omitempty"`
    MaxTokens    *int     `json:"max_tokens,omitempty"`
}

// EvaluateResponse é a resposta do endpoint /api/v1/flags/:key/evaluate.
type EvaluateResponse struct {
    Enabled       bool           `json:"enabled"`
    Variant       string         `json:"variant"`
    PromptVariant *PromptVariant `json:"prompt_variant,omitempty"`
}

// defaultPrompt é o fallback quando a flag está desabilitada ou o Pulse está indisponível.
var defaultPrompt = PromptVariant{
    SystemPrompt: "Você é um assistente útil e conciso.",
    Model:        "gemini-2.5-flash-preview-04-17",
}

// GetPromptForUser consulta o Pulse e retorna o prompt correto para o usuário.
// Fail-open: em caso de erro, retorna o defaultPrompt.
func GetPromptForUser(ctx context.Context, userID string) PromptVariant {
    pulseURL := os.Getenv("PULSE_API_URL")
    serviceToken := os.Getenv("PULSE_SERVICE_TOKEN")
    flagKey := "attend-system-prompt"

    url := fmt.Sprintf("%s/api/v1/flags/%s/evaluate?user_id=%s", pulseURL, flagKey, userID)
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if err != nil {
        return defaultPrompt
    }
    req.Header.Set("X-Service-Token", serviceToken)

    resp, err := http.DefaultClient.Do(req)
    if err != nil || resp.StatusCode != http.StatusOK {
        return defaultPrompt
    }
    defer resp.Body.Close()

    var result EvaluateResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return defaultPrompt
    }

    // Flag desabilitada ou sem prompt_variant → usa default
    if !result.Enabled || result.PromptVariant == nil {
        return defaultPrompt
    }

    return *result.PromptVariant
}
```

## Usar no handler de chat

```go
func (h *ChatHandler) HandleMessage(c *fiber.Ctx) error {
    userID := c.Get("X-User-ID")

    // Busca config de LLM do Pulse (com fallback automático)
    prompt := llmconfig.GetPromptForUser(c.Context(), userID)

    // Chama o LLM com o prompt da flag
    response, err := h.llmClient.Chat(c.Context(), llm.ChatRequest{
        Model:        prompt.Model,
        SystemPrompt: prompt.SystemPrompt,
        Temperature:  prompt.Temperature,
        MaxTokens:    prompt.MaxTokens,
        UserMessage:  c.Body(),
    })
    // ...
}
```

## Verificar resultados

Use o SQL Explorer (BKL-993) para comparar métricas entre variantes:

```bash
# Conta mensagens por grupo (flag ativa vs. inativa) nos últimos 7 dias
curl -X POST https://pulse.institutoitinerante.com.br/api/v1/projects/$PROJECT_ID/query \
  -H "X-Service-Token: $SERVICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "event": "chat_message",
    "aggregate": "count",
    "group_by": ["browser"],
    "date_range": {"from": "2026-04-05", "to": "2026-04-11"}
  }'
```

## Resposta do endpoint /evaluate

```json
{
  "enabled": true,
  "variant": "",
  "prompt_variant": {
    "system_prompt": "Você é o Brio, assistente de atendimento da IIT. Responda de forma concisa e empática.",
    "model": "gemini-2.5-flash-preview-04-17",
    "temperature": 0.7,
    "max_tokens": 512
  }
}
```

Quando a flag está desabilitada para o usuário, `prompt_variant` é omitido e o cliente usa seu fallback.
