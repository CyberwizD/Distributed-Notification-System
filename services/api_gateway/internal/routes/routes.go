package routes

import (
	"time"

	"github.com/CyberwizD/Distributed-Notification-System/services/api_gateway/internal/handlers"
	"github.com/CyberwizD/Distributed-Notification-System/services/api_gateway/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/sony/gobreaker"
)

// SetupRoutes configures the routes for the application.
func SetupRoutes(
	router *gin.Engine,
	notificationHandler *handlers.NotificationHandler,
	statusHandler *handlers.StatusHandler,
	redisClient *redis.Client,
) {
	router.Use(middleware.CorrelationIDMiddleware())

	// Initialize circuit breaker
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{})

	// Setup routes
	v1 := router.Group("/v1")
	v1.Use(middleware.AuthMiddleware())
	v1.Use(middleware.RateLimitMiddleware(redisClient, 100, time.Minute))
	v1.Use(middleware.CircuitBreakerMiddleware(cb))
	{
		notifications := v1.Group("/notifications")
		{
			notifications.POST("/send", notificationHandler.SendNotification)
			notifications.GET("/:request_id/status", statusHandler.GetStatus)
		}
	}

	// Health check endpoint
	router.GET("/health", handlers.HealthCheck)
}
