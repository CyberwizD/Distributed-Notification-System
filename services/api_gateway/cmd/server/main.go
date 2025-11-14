package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/CyberwizD/Distributed-Notification-System/services/api_gateway/internal/config"
	"github.com/CyberwizD/Distributed-Notification-System/services/api_gateway/internal/handlers"
	"github.com/CyberwizD/Distributed-Notification-System/services/api_gateway/internal/repository"
	"github.com/CyberwizD/Distributed-Notification-System/services/api_gateway/internal/routes"
	"github.com/CyberwizD/Distributed-Notification-System/services/api_gateway/internal/services"
	"github.com/CyberwizD/Distributed-Notification-System/services/api_gateway/pkg/logger"
	"github.com/CyberwizD/Distributed-Notification-System/services/api_gateway/pkg/metrics"
	"github.com/CyberwizD/Distributed-Notification-System/services/api_gateway/pkg/rabbitmq"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	logr := logger.New(cfg.LogLevel)

	// Initialize database
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		logr.Error("failed to connect to database", slog.Any("error", err))
		os.Exit(1)
	}

	// Initialize Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.RedisURL,
	})
	defer redisClient.Close()

	metricsCollector := metrics.New()

	// Initialize RabbitMQ
	mqManager, err := rabbitmq.NewManager(cfg.RabbitMQURL, logr)
	if err != nil {
		logr.Error("failed to connect to RabbitMQ", slog.Any("error", err))
		os.Exit(1)
	}
	defer mqManager.Close()

	if err := mqManager.DeclareNotificationTopology(
		"notifications.direct",
		map[string]string{
			"email.queue": "email",
			"push.queue":  "push",
		},
		"failed.queue",
	); err != nil {
		logr.Error("failed to declare rabbitmq topology", slog.Any("error", err))
		os.Exit(1)
	}
	amqpConn := mqManager.Connection()

	// Initialize repositories
	statusStore := repository.NewStatusStore(db)
	redisRepo := repository.NewRedisRepository(redisClient)

	// Initialize services
	idempotencyService := services.NewIdempotencyService(redisRepo)
	userClient := services.NewUserClient(cfg.UserServiceURL, cfg.UserServiceAPIKey, redisRepo, cfg.UserPrefCacheTTL)
	publisher := services.NewPublisher(amqpConn)
	templateClient := services.NewTemplateClient(cfg.TemplateServiceURL)

	// Initialize handlers
	notificationHandler := handlers.NewNotificationHandler(idempotencyService, userClient, publisher, statusStore, templateClient)
	statusHandler := handlers.NewStatusHandler(statusStore)

	// Initialize router
	router := gin.Default()
	router.Use(metricsCollector.GinMiddleware())
	router.GET("/metrics", gin.WrapH(metricsCollector.Handler()))

	// Setup routes
	routes.SetupRoutes(router, notificationHandler, statusHandler, redisClient)

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logr.Error("server listen failed", slog.Any("error", err))
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logr.Info("shutting down server")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logr.Error("server forced to shutdown", slog.Any("error", err))
	}

	logr.Info("server exiting")
}
