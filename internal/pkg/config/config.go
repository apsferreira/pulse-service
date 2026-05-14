package config

import (
	"os"
)

// Config holds all runtime configuration for pulse-service.
type Config struct {
	DatabaseURL    string
	RedisURL       string
	ServiceToken   string
	Port           string
	LogLevel       string
	// AllowedOrigins lista as origens CORS permitidas.
	// SEC-005: configurável via PULSE_ALLOWED_ORIGINS (CSV).
	// Default "*" mantido pois o SDK Pulse é embeddable em qualquer domínio de cliente.
	// Adicionar Vary: Origin é obrigatório quando o valor não for literal "*".
	AllowedOrigins string
	// BillingURL é a URL exibida no HTTP 429 quando a quota mensal é atingida.
	// Vazio = omite o campo (adequado para self-hosted sem portal de billing).
	BillingURL string
	// AuthMode controla o mecanismo de autenticação do dashboard.
	// "external" (padrão): redireciona para auth-service externo.
	// "local": email+senha na tabela pulse_users — adequado para self-hosted.
	AuthMode string
	// JWTSecret é usado para assinar tokens no modo local.
	// Se vazio, usa ServiceToken como fallback.
	JWTSecret string
}

// Load reads configuration from environment variables, applying defaults where appropriate.
// Default Redis URL uses DB 6 for feature flags cache.
func Load() *Config {
	cfg := &Config{
		DatabaseURL:    getEnv("DATABASE_URL", ""),
		RedisURL:       getEnv("REDIS_URL", "redis://localhost:6379/6"),
		ServiceToken:   getEnv("SERVICE_TOKEN", ""),
		Port:           getEnv("PORT", "3050"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		AllowedOrigins: getEnv("PULSE_ALLOWED_ORIGINS", "*"),
		BillingURL:     getEnv("PULSE_BILLING_URL", ""),
		AuthMode:       getEnv("PULSE_AUTH_MODE", "external"),
		JWTSecret:      getEnv("PULSE_JWT_SECRET", ""),
	}
	// Fallback: se JWTSecret não foi definido, usa ServiceToken como secret.
	// Só entra em vigor no modo local; não expõe o ServiceToken em nenhum log.
	if cfg.JWTSecret == "" {
		cfg.JWTSecret = cfg.ServiceToken
	}
	return cfg
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
