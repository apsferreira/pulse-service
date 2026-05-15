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
	"crypto/sha256"
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
// billingURL é incluído na resposta 429; se vazio, o campo é omitido.
func QuotaEnforcement(rdb *redis.Client, billingURL string) fiber.Handler {
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

		// Pro/enterprise/internal: sem limite de quota
		plan := getProjectPlan(c)
		if plan == "pro" || plan == "enterprise" || plan == "internal" {
			return c.Next()
		}

		// Chave de quota por account_email: todos os projetos da conta somam para a mesma cota.
		accountEmail, _ := c.Locals("account_email").(string)
		if accountEmail == "" {
			accountEmail = projectID // fallback seguro para rotas sem account_email
		}

		limit := quotaByPlan(plan)
		key := quotaKey(accountEmail)
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
			resp := fiber.Map{
				"error":    "quota mensal atingida — faça upgrade do seu plano",
				"plan":     plan,
				"limit":    limit,
				"consumed": current - 1,
				"reset_at": endOfMonth().Format(time.RFC3339),
			}
			if billingURL != "" {
				resp["upgrade_url"] = billingURL
			}
			return c.Status(fiber.StatusTooManyRequests).JSON(resp)
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

// QuotaStatus retorna o consumo atual da quota da conta sem incrementar.
// Útil para endpoints de status/dashboard.
func QuotaStatus(rdb *redis.Client, accountEmail, plan string) (consumed, limit int64, resetAt time.Time) {
	resetAt = endOfMonth()
	if plan == "pro" || plan == "enterprise" || plan == "internal" {
		return 0, -1, resetAt // -1 = ilimitado
	}

	limit = int64(quotaByPlan(plan))
	key := quotaKey(accountEmail)
	ctx := context.Background()

	val, err := rdb.Get(ctx, key).Int64()
	if err != nil {
		return 0, limit, resetAt
	}
	return val, limit, resetAt
}

// quotaKey retorna a chave Redis para o contador mensal da conta.
// Usa os primeiros 8 bytes do SHA-256 do email para evitar caracteres especiais na chave.
// Inclui ano+mês para reset automático no início de cada mês.
func quotaKey(accountEmail string) string {
	now := time.Now().UTC()
	h := sha256.Sum256([]byte(accountEmail))
	return fmt.Sprintf("quota:pulse:account:%x:%d-%02d", h[:8], now.Year(), now.Month())
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
