package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sony/gobreaker"
)

// CircuitBreakerMiddleware creates a middleware for the circuit breaker.
func CircuitBreakerMiddleware(cb *gobreaker.CircuitBreaker) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, err := cb.Execute(func() (interface{}, error) {
			c.Next()
			return nil, nil
		})

		if err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "service is unavailable"})
		}
	}
}
