package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds push service configuration loaded from the environment.
type Config struct {
	AppName             string
	LogLevel            string
	HTTPPort            string
	MetricsAddr         string
	RabbitURL           string
	PushQueue           string
	DeadLetterQueue     string
	PrefetchCount       int
	WorkerCount         int
	TemplateServiceURL  string
	DatabaseURL         string
	RedisURL            string
	StatusTable         string
	FCMServerKey        string
	FCMEndpoint         string
	ProviderTimeout     time.Duration
	RetryMaxAttempts    int
	RetryInitialBackoff time.Duration
	RetryMaxBackoff     time.Duration
}

// Load loads configuration and performs basic validation.
func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		AppName:             getEnv("APP_NAME", "push_service"),
		LogLevel:            getEnv("LOG_LEVEL", "info"),
		HTTPPort:            getEnv("HTTP_PORT", "8082"),
		MetricsAddr:         getEnv("METRICS_ADDR", ":9092"),
		RabbitURL:           getEnv("RABBITMQ_URL", ""),
		PushQueue:           getEnv("PUSH_QUEUE", "push.queue"),
		DeadLetterQueue:     getEnv("PUSH_DLQ", "failed.queue"),
		PrefetchCount:       getEnvAsInt("PUSH_PREFETCH", 100),
		WorkerCount:         getEnvAsInt("WORKER_COUNT", 5),
		TemplateServiceURL:  getEnv("TEMPLATE_SERVICE_URL", ""),
		DatabaseURL:         getEnv("DATABASE_URL", ""),
		RedisURL:            getEnv("REDIS_URL", ""),
		StatusTable:         getEnv("STATUS_TABLE", "notification_statuses"),
		FCMServerKey:        getEnv("FCM_SERVER_KEY", ""),
		FCMEndpoint:         getEnv("FCM_ENDPOINT", "https://fcm.googleapis.com/fcm/send"),
		ProviderTimeout:     getEnvAsDuration("PROVIDER_TIMEOUT", 10*time.Second),
		RetryMaxAttempts:    getEnvAsInt("RETRY_MAX_ATTEMPTS", 4),
		RetryInitialBackoff: getEnvAsDuration("RETRY_INITIAL_BACKOFF", time.Second),
		RetryMaxBackoff:     getEnvAsDuration("RETRY_MAX_BACKOFF", 15*time.Second),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) validate() error {
	var missing []string
	if c.RabbitURL == "" {
		missing = append(missing, "RABBITMQ_URL")
	}
	if c.DatabaseURL == "" {
		missing = append(missing, "DATABASE_URL")
	}
	if c.TemplateServiceURL == "" {
		missing = append(missing, "TEMPLATE_SERVICE_URL")
	}
	if c.FCMServerKey == "" {
		missing = append(missing, "FCM_SERVER_KEY")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %v", missing)
	}
	return nil
}

func getEnv(key, def string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return def
	}
	return value
}

func getEnvAsInt(key string, def int) int {
	if value, ok := os.LookupEnv(key); ok {
		i, err := strconv.Atoi(value)
		if err != nil {
			log.Printf("invalid int for %s, using default %d: %v", key, def, err)
			return def
		}
		return i
	}
	return def
}

func getEnvAsDuration(key string, def time.Duration) time.Duration {
	if value, ok := os.LookupEnv(key); ok {
		d, err := time.ParseDuration(value)
		if err != nil {
			log.Printf("invalid duration for %s, using default %s: %v", key, def, err)
			return def
		}
		return d
	}
	return def
}
