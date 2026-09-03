package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"go-cp-boilerplate/internal/cache"
	"go-cp-boilerplate/internal/pkg/response"
)

// RateLimit implements a fixed-window counter backed by the shared Redis
// instance (INCR + EXPIRE), so limits hold across multiple app instances,
// not just per-process like an in-memory limiter would.
func RateLimit(redisClient *cache.RedisClient, scope string, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := redisClient.Key("ratelimit", scope, c.ClientIP())

		count, err := redisClient.Incr(c.Request.Context(), key).Result()
		if err != nil {
			// Fail open: don't block traffic if Redis is unreachable.
			c.Next()
			return
		}
		if count == 1 {
			redisClient.Expire(c.Request.Context(), key, window)
		}

		if count > int64(limit) {
			response.Error(c, http.StatusTooManyRequests, "Too many requests, please try again later", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}
