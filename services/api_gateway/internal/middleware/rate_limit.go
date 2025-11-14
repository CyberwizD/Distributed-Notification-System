package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// RateLimitMiddleware creates a middleware for rate limiting.
func RateLimitMiddleware(redisClient *redis.Client, limit int, duration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := "rate_limit:" + ip

		// Use a pipeline to efficiently execute multiple commands
		pipe := redisClient.Pipeline()
		pipe.Incr(c, key)
		pipe.Expire(c, key, duration)
		cmds, err := pipe.Exec(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to execute redis pipeline"})
			return
		}

		count := cmds[0].(*redis.IntCmd).Val()
		if count > int64(limit) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}

		c.Next()
	}
}
