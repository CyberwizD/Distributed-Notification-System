package main

import (
	"context"
	"log"
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
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/streadway/amqp"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	// Initialize database
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// Initialize Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.RedisURL,
	})

	// Initialize RabbitMQ
	amqpConn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("failed to connect to RabbitMQ: %v", err)
	}
	defer amqpConn.Close()

	// Initialize repositories
	statusStore := repository.NewStatusStore(db)
	redisRepo := repository.NewRedisRepository(redisClient)

	// Initialize services
	idempotencyService := services.NewIdempotencyService(redisRepo)
	userClient := services.NewUserClient(cfg.UserServiceURL, redisRepo, cfg.UserPrefCacheTTL)
	publisher := services.NewPublisher(amqpConn)
	templateClient := services.NewTemplateClient(cfg.TemplateServiceURL)

	// Initialize handlers
	notificationHandler := handlers.NewNotificationHandler(idempotencyService, userClient, publisher, statusStore, templateClient)
	statusHandler := handlers.NewStatusHandler(statusStore)

	// Initialize router
	router := gin.Default()

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
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
