package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/institutoitinerante/pulse-service/internal/domain"
)

// ServiceTokenMiddleware validates the X-Service-Token header against the configured token.
// Returns 401 Unauthorized if the header is missing or does not match.
func ServiceTokenMiddleware(serviceToken string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Get("X-Service-Token")
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
				Error: "missing X-Service-Token header",
			})
		}
		if token != serviceToken {
			return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
				Error: "invalid service token",
			})
		}
		return c.Next()
	}
}
