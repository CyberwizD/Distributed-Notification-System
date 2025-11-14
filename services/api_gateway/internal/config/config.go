package config

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// Config holds the application configuration.
type Config struct {
	Port               string
	DatabaseURL        string
	RedisURL           string
	RabbitMQURL        string
	UserServiceURL     string
	UserServiceAPIKey  string
	TemplateServiceURL string
	UserPrefCacheTTL   time.Duration
	LogLevel           string
}

// Load loads the configuration from environment variables.
func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	return &Config{
		Port:               getEnv("PORT", "8080"),
		DatabaseURL:        getEnv("DATABASE_URL", ""),
		RedisURL:           getEnv("REDIS_URL", ""),
		RabbitMQURL:        getEnv("RABBITMQ_URL", ""),
		UserServiceURL:     getEnv("USER_SERVICE_URL", ""),
		UserServiceAPIKey:  getEnv("USER_SERVICE_INTERNAL_API_KEY", ""),
		TemplateServiceURL: getEnv("TEMPLATE_SERVICE_URL", ""),
		UserPrefCacheTTL:   getEnvAsDuration("USER_PREF_CACHE_TTL", 5*time.Minute),
		LogLevel:           getEnv("LOG_LEVEL", "info"),
	}, nil
}

// getEnv retrieves an environment variable or returns a default value.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
		log.Printf("invalid duration for %s; using default %s", key, defaultValue)
	}
	return defaultValue
}
