package middleware

// BKL-1077: Quota enforcement por plano com hard limit.
// Verifica eventos consumidos no mês corrente antes de aceitar novos.
//
// Limites mensais:
//   - free:    10.000 eventos/mês
//   - starter: 500.000 eventos/mês
//   - growth:  2.000.000 eventos/mês
//   - pro:     ilimitado (bypass)
//
// Implementação: Redis INCR com TTL até fim do mês (chave única por projeto/mês).
// Fail-open: qualquer erro de Redis permite a requisição (nunca bloquear por infra).
// Retorna HTTP 429 com link de upgrade quando limite atingido.

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

const (
	quotaFree    = 10_000
	quotaStarter = 500_000
	quotaGrowth  = 2_000_000
	// Pro: sem limite
)

// QuotaEnforcement retorna middleware que verifica e decrementa quota mensal do projeto.
// Deve ser aplicado APÓS ValidateProjectKey (depende de Locals("project_id") e Locals("project_plan")).
// Se rdb for nil, retorna middleware no-op (fail-open).
func QuotaEnforcement(rdb *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Sem Redis — fail-open
		if rdb == nil {
			return c.Next()
		}

		projectID, _ := c.Locals("project_id").(string)

		// Sem project_id (rota interna/admin) — bypass
		if projectID == "" {
			return c.Next()
		}

		// Pro: sem limite
		plan := getProjectPlan(c)
		if plan == "pro" {
			return c.Next()
		}

		limit := quotaByPlan(plan)
		key := quotaKey(projectID)
		ctx := context.Background()

		// Incrementa contador e garante TTL até fim do mês
		pipe := rdb.Pipeline()
		incr := pipe.Incr(ctx, key)
		// Apenas seta TTL na primeira vez (para não redefinir no meio do mês)
		pipe.ExpireNX(ctx, key, timeUntilEndOfMonth())
		if _, err := pipe.Exec(ctx); err != nil {
			// Fail-open: erro de Redis não bloqueia ingestão
			return c.Next()
		}

		current := incr.Val()
		if current > int64(limit) {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":      "quota mensal atingida — faça upgrade do seu plano",
				"plan":       plan,
				"limit":      limit,
				"consumed":   current - 1,
				"reset_at":   endOfMonth().Format(time.RFC3339),
				"upgrade_url": "https://pulse.institutoitinerante.com.br/billing",
			})
		}

		// Injeta headers informativos
		remaining := int64(limit) - current
		if remaining < 0 {
			remaining = 0
		}
		c.Set("X-Quota-Limit", fmt.Sprintf("%d", limit))
		c.Set("X-Quota-Remaining", fmt.Sprintf("%d", remaining))
		c.Set("X-Quota-Reset", endOfMonth().Format(time.RFC3339))

		return c.Next()
	}
}

// QuotaStatus retorna o consumo atual da quota do projeto sem incrementar.
// Útil para endpoints de status/dashboard.
func QuotaStatus(rdb *redis.Client, projectID, plan string) (consumed, limit int64, resetAt time.Time) {
	resetAt = endOfMonth()
	if plan == "pro" {
		return 0, -1, resetAt // -1 = ilimitado
	}

	limit = int64(quotaByPlan(plan))
	key := quotaKey(projectID)
	ctx := context.Background()

	val, err := rdb.Get(ctx, key).Int64()
	if err != nil {
		return 0, limit, resetAt
	}
	return val, limit, resetAt
}

// quotaKey retorna a chave Redis para o contador mensal do projeto.
// Inclui ano+mês para reset automático no início de cada mês.
func quotaKey(projectID string) string {
	now := time.Now().UTC()
	return fmt.Sprintf("quota:pulse:project:%s:%d-%02d", projectID, now.Year(), now.Month())
}

// quotaByPlan retorna o limite mensal de eventos para o plano.
func quotaByPlan(plan string) int {
	switch plan {
	case "starter":
		return quotaStarter
	case "growth":
		return quotaGrowth
	default:
		return quotaFree
	}
}

// endOfMonth retorna o instante UTC de fim do mês corrente (23:59:59).
func endOfMonth() time.Time {
	now := time.Now().UTC()
	firstOfNextMonth := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.UTC)
	return firstOfNextMonth.Add(-time.Second)
}

// timeUntilEndOfMonth retorna a duração até o fim do mês corrente.
func timeUntilEndOfMonth() time.Duration {
	return time.Until(endOfMonth())
}
