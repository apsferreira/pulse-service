package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/institutoitinerante/pulse-service/internal/domain"
)

// ServiceTokenMiddleware validates X-Service-Token OR Authorization Bearer JWT.
// Aceita ambos para retrocompatibilidade (inter-service usa token, dashboard usa JWT).
// BKL-CSAT-fix: lê PULSE_JWT_SECRET (nome canônico da env var) com fallback para JWT_SECRET.
func ServiceTokenMiddleware(serviceToken string) fiber.Handler {
	jwtSecret := os.Getenv("PULSE_JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = os.Getenv("JWT_SECRET")
	}

	return func(c *fiber.Ctx) error {
		// 1. Tenta Service Token primeiro — BKL-334: comparação em tempo constante
		svcToken := c.Get("X-Service-Token")
		if svcToken != "" && subtle.ConstantTimeCompare([]byte(svcToken), []byte(serviceToken)) == 1 {
			return c.Next()
		}

		// 2. Tenta JWT Bearer (HMAC-SHA256)
		authHeader := c.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") && jwtSecret != "" {
			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			if validateJWT(tokenStr, jwtSecret) {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error: "autenticacao necessaria — envie X-Service-Token ou Authorization Bearer JWT",
		})
	}
}

// validateJWT verifica um JWT HS256 sem dependencias externas.
func validateJWT(tokenStr, secret string) bool {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return false
	}

	// Verificar assinatura HMAC-SHA256
	signingInput := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signingInput))
	expectedSig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expectedSig), []byte(parts[2])) {
		return false
	}

	// Verificar expiração
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}
	var claims struct {
		Exp int64 `json:"exp"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return false
	}
	if claims.Exp > 0 && time.Now().Unix() > claims.Exp {
		return false
	}

	return true
}
