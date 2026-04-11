package middleware

// BKL-1032: Rate limiting por project_id com limites por plano.
// Substitui rate limit por IP (inadequado para SaaS multi-tenant).
//
// Limites:
//   - Free:    1.000 req/min
//   - Growth: 10.000 req/min
//   - Pro:    sem limite (bypass)
//
// Fallback: se project_id não estiver em Locals (rota sem ValidateProjectKey), usa IP com limite Free.
// Implementação: sliding window via Redis INCR + EXPIRE em pipeline atômico.

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

const (
	rateLimitFree   = 1_000
	rateLimitGrowth = 10_000
	rateLimitWindow = time.Minute
)

// RateLimit retorna middleware de rate limiting por project_id com limites por plano.
// BKL-1032: substitui rate limit por IP.
// BKL-463: usa Redis para funcionar com múltiplas réplicas.
func RateLimit(rdb *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Obtém project_id injetado pelo ValidateProjectKey (BKL-1029)
		projectID, _ := c.Locals("project_id").(string)

		// Projetos Pro não têm rate limit
		plan := getProjectPlan(c)
		if plan == "pro" {
			return c.Next()
		}

		limit := rateLimitByPlan(plan)

		// Chave por project_id; fallback para IP se não autenticado
		key := rateLimitKey(projectID, c.IP())
		ctx := context.Background()

		pipe := rdb.Pipeline()
		incr := pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, rateLimitWindow)
		if _, err := pipe.Exec(ctx); err != nil {
			// Falha no Redis — fail open (não bloquear por erro de infra)
			return c.Next()
		}

		current := incr.Val()
		if current > int64(limit) {
			c.Set("Retry-After", "60")
			c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
			c.Set("X-RateLimit-Remaining", "0")
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "rate limit exceeded — upgrade your plan for higher limits",
				"limit": limit,
			})
		}

		remaining := int64(limit) - current
		if remaining < 0 {
			remaining = 0
		}
		c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		c.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		return c.Next()
	}
}

// rateLimitKey retorna a chave Redis para sliding window.
// Usa project_id quando disponível; fallback para IP.
func rateLimitKey(projectID, ip string) string {
	if projectID != "" {
		return "rl:pulse:project:" + projectID
	}
	return "rl:pulse:ip:" + ip
}

// rateLimitByPlan retorna o limite de requisições por minuto para o plano.
func rateLimitByPlan(plan string) int {
	switch plan {
	case "growth":
		return rateLimitGrowth
	default:
		return rateLimitFree
	}
}

// getProjectPlan lê o plano do projeto a partir do Locals injetado pelo ValidateProjectKey.
// Por enquanto retorna "free" para qualquer projeto sem plano configurado.
// TODO: quando billing-service integrado, adicionar campo plan ao Project e injetar aqui.
func getProjectPlan(c *fiber.Ctx) string {
	// Verifica se há plano explícito nos Locals
	if plan, ok := c.Locals("project_plan").(string); ok && plan != "" {
		return plan
	}
	return "free"
}
