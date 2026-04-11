package middleware

// BKL-463: Rate limiting por project_id usando Redis (sliding window) em vez de memória.
// Corrige o rate limit anterior que era por IP e em memória — inadequado para SaaS multi-tenant
// com múltiplas réplicas horizontais.
//
// BKL-1032: Limites por plano (free/growth/pro).
//
// Limites:
//   - Free:    1.000 req/min por projeto
//   - Growth: 10.000 req/min por projeto
//   - Pro:    sem limite (bypass)
//
// Implementação: fixed window via INCR + EXPIRE com NX (seta TTL apenas na criação da chave).
// Isso garante que a janela não seja resetada a cada request.
//
// Chaves Redis: rl:pulse:project:<project_id> | rl:pulse:ip:<ip>
//
// Headers de resposta:
//   - X-RateLimit-Limit:     limite do plano
//   - X-RateLimit-Remaining: requisições restantes na janela
//   - X-RateLimit-Reset:     unix timestamp em que a janela expira
//   - Retry-After:           segundos até retry (apenas em 429)
//
// Fallback: se Redis estiver indisponível → fail-open (logar warning, permitir request).

import (
	"context"
	"fmt"
	"log"
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
// BKL-463: Redis-backed, seguro para múltiplas réplicas.
func RateLimit(rdb *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Projetos Pro não têm rate limit — bypass imediato
		plan := getProjectPlan(c)
		if plan == "pro" {
			return c.Next()
		}

		limit := rateLimitByPlan(plan)

		// Obtém project_id injetado pelo ValidateProjectKey (BKL-1029)
		projectID, _ := c.Locals("project_id").(string)

		// Chave por project_id; fallback para IP se rota não autenticada
		key := rateLimitKey(projectID, c.IP())
		ctx := context.Background()

		// Calcula o unix timestamp em que a janela atual expira.
		// Usamos uma janela fixa alinhada ao minuto para previsibilidade.
		now := time.Now()
		windowStart := now.Truncate(rateLimitWindow)
		windowEnd := windowStart.Add(rateLimitWindow)
		ttl := time.Until(windowEnd)
		resetAt := windowEnd.Unix()

		// Chave inclui o início da janela para garantir isolamento entre janelas
		windowKey := fmt.Sprintf("%s:%d", key, windowStart.Unix())

		// INCR atômico: incrementa e obtém valor atual
		current, err := rdb.Incr(ctx, windowKey).Result()
		if err != nil {
			// Fail-open: Redis indisponível não deve derrubar o serviço
			log.Printf("ratelimit: redis INCR error (fail-open): %v", err)
			return c.Next()
		}

		// Seta TTL apenas se a chave acabou de ser criada (EXPIRE com NX).
		// Se já existia, o TTL não é alterado — janela não é estendida.
		if current == 1 {
			if err := rdb.Expire(ctx, windowKey, ttl).Err(); err != nil {
				// Não fatal: chave pode expirar depois, mas funciona corretamente
				log.Printf("ratelimit: redis EXPIRE error: %v", err)
			}
		}

		remaining := int64(limit) - current
		if remaining < 0 {
			remaining = 0
		}

		// Seta headers informativos em todos os casos
		c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		c.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Set("X-RateLimit-Reset", fmt.Sprintf("%d", resetAt))

		if current > int64(limit) {
			c.Set("Retry-After", fmt.Sprintf("%d", int(ttl.Seconds())))
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":       "rate limit exceeded — upgrade your plan for higher limits",
				"limit":       limit,
				"reset_at":    resetAt,
				"retry_after": int(ttl.Seconds()),
			})
		}

		return c.Next()
	}
}

// rateLimitKey retorna a chave Redis base para rate limiting.
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
// Retorna "free" para qualquer projeto sem plano configurado.
func getProjectPlan(c *fiber.Ctx) string {
	if plan, ok := c.Locals("project_plan").(string); ok && plan != "" {
		return plan
	}
	return "free"
}
