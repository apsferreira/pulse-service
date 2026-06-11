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

// Defaults para verificação de claims iss/aud em JWTs Bearer.
// Issue #62: tokens emitidos pelo auth-service IIT para outros serviços (focus-hub etc.)
// não devem autenticar no pulse-service — exigimos aud conhecido e iss conhecido.
// Issue #77: auth-service emite aud="iit" (audience comum a todo o ecossistema IIT);
// tokens locais (modo dev/teste) continuam aceitos via mesma audience.
const (
	defaultExpectedIssuers  = "auth.institutoitinerante.com.br,pulse-service"
	defaultExpectedAudience = "iit"
)

// ServiceTokenMiddleware validates X-Service-Token OR Authorization Bearer JWT.
// Aceita ambos para retrocompatibilidade (inter-service usa token, dashboard usa JWT).
// BKL-CSAT-fix: lê PULSE_JWT_SECRET (nome canônico da env var) com fallback para JWT_SECRET.
// Issue #62: além da assinatura, exige claims iss/aud (configuráveis via env).
func ServiceTokenMiddleware(serviceToken string) fiber.Handler {
	jwtSecret := os.Getenv("PULSE_JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = os.Getenv("JWT_SECRET")
	}

	expectedIssuers := parseCSV(getEnvDefault("PULSE_JWT_EXPECTED_ISSUER", defaultExpectedIssuers))
	expectedAudience := getEnvDefault("PULSE_JWT_EXPECTED_AUDIENCE", defaultExpectedAudience)

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
			if validateJWT(tokenStr, jwtSecret, expectedIssuers, expectedAudience) {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{
			Error: "autenticacao necessaria — envie X-Service-Token ou Authorization Bearer JWT",
		})
	}
}

// validateJWT verifica um JWT HS256 sem dependencias externas.
// Issue #62: além de assinatura e expiração, valida iss (issuer) e aud (audience)
// para impedir que tokens emitidos para outros serviços IIT autentiquem aqui.
func validateJWT(tokenStr, secret string, expectedIssuers []string, expectedAudience string) bool {
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

	// Decodificar payload
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}
	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return false
	}

	// Expiração — obrigatória
	expVal, ok := claims["exp"].(float64)
	if !ok || time.Now().Unix() > int64(expVal) {
		return false
	}

	// Issuer — obrigatório quando expectedIssuers configurado
	if len(expectedIssuers) > 0 {
		iss, _ := claims["iss"].(string)
		if !containsString(expectedIssuers, iss) {
			return false
		}
	}

	// Audience — obrigatória quando expectedAudience configurado
	if expectedAudience != "" {
		if !audienceMatches(claims["aud"], expectedAudience) {
			return false
		}
	}

	return true
}

// audienceMatches aceita aud como string ou []string (conforme RFC 7519).
func audienceMatches(audClaim interface{}, expected string) bool {
	switch v := audClaim.(type) {
	case string:
		return v == expected
	case []interface{}:
		for _, item := range v {
			if s, ok := item.(string); ok && s == expected {
				return true
			}
		}
	}
	return false
}

func containsString(list []string, s string) bool {
	for _, item := range list {
		if item == s {
			return true
		}
	}
	return false
}

func parseCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func getEnvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
