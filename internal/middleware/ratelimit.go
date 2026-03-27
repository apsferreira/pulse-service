package middleware

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

type visitor struct {
	count   int
	resetAt time.Time
}

var (
	visitors sync.Map
)

func init() {
	go cleanupVisitors()
}

func cleanupVisitors() {
	for {
		time.Sleep(1 * time.Minute)
		now := time.Now()
		visitors.Range(func(key, value any) bool {
			v := value.(*visitor)
			if now.After(v.resetAt) {
				visitors.Delete(key)
			}
			return true
		})
	}
}

// RateLimit allows up to 100 requests per minute per IP.
func RateLimit() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ip := c.IP()

		now := time.Now()
		val, _ := visitors.LoadOrStore(ip, &visitor{count: 0, resetAt: now.Add(1 * time.Minute)})
		v := val.(*visitor)

		if now.After(v.resetAt) {
			v.count = 0
			v.resetAt = now.Add(1 * time.Minute)
		}

		v.count++
		if v.count > 100 {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "rate limit exceeded",
			})
		}

		return c.Next()
	}
}
