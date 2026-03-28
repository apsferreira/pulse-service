package config

import (
	"os"
)

// Config holds all runtime configuration for pulse-service.
type Config struct {
	DatabaseURL  string
	RedisURL     string
	ServiceToken string
	Port         string
	LogLevel     string
}

// Load reads configuration from environment variables, applying defaults where appropriate.
// Default Redis URL uses DB 6 for feature flags cache.
func Load() *Config {
	return &Config{
		DatabaseURL:  getEnv("DATABASE_URL", ""),
		RedisURL:     getEnv("REDIS_URL", "redis://localhost:6379/6"),
		ServiceToken: getEnv("SERVICE_TOKEN", ""),
		Port:         getEnv("PORT", "3050"),
		LogLevel:     getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
