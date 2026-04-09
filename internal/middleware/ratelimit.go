package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

// RateLimit retorna middleware de rate limiting com sliding window via Redis.
// Limite padrão: 100 req/min por IP. Usa INCR + EXPIRE em pipeline atômico.
// BKL-463: substituído sync.Map (in-memory, não funciona em múltiplas réplicas) por Redis.
func RateLimit(rdb *redis.Client) fiber.Handler {
	const (
		limit  = 100
		window = time.Minute
	)

	return func(c *fiber.Ctx) error {
		ip := c.IP()
		key := fmt.Sprintf("rl:pulse:%s", ip)
		ctx := context.Background()

		pipe := rdb.Pipeline()
		incr := pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, window)
		if _, err := pipe.Exec(ctx); err != nil {
			// Falha no Redis — fail open (não bloquear usuário por erro de infra)
			return c.Next()
		}

		if incr.Val() > int64(limit) {
			c.Set("Retry-After", "60")
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "rate limit exceeded",
			})
		}

		return c.Next()
	}
}
