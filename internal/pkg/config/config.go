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
}

// Load reads configuration from environment variables, applying defaults where appropriate.
// Default Redis URL uses DB 6 for feature flags cache.
func Load() *Config {
	return &Config{
		DatabaseURL:    getEnv("DATABASE_URL", ""),
		RedisURL:       getEnv("REDIS_URL", "redis://localhost:6379/6"),
		ServiceToken:   getEnv("SERVICE_TOKEN", ""),
		Port:           getEnv("PORT", "3050"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		AllowedOrigins: getEnv("PULSE_ALLOWED_ORIGINS", "*"),
	}
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
