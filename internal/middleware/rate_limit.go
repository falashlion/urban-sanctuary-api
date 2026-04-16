package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/falashlion/urban-sanctuary-api/internal/platform/cache"
	"github.com/falashlion/urban-sanctuary-api/pkg/response"
	"github.com/gin-gonic/gin"
)

// RateLimiter creates a Redis-backed sliding window rate limiter.
func RateLimiter(redis *cache.RedisClient, maxRequests int64, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if redis == nil {
			c.Next()
			return
		}

		ip := c.ClientIP()
		key := fmt.Sprintf("rate_limit:%s", ip)

		count, err := redis.Incr(c.Request.Context(), key)
		if err != nil {
			// If Redis is down, allow the request
			c.Next()
			return
		}

		if count == 1 {
			_ = redis.Expire(c.Request.Context(), key, window)
		}

		// Set rate limit headers
		remaining := maxRequests - count
		if remaining < 0 {
			remaining = 0
		}
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", maxRequests))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		ttl, _ := redis.TTL(c.Request.Context(), key)
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(ttl).Unix()))

		if count > maxRequests {
			c.Header("Retry-After", fmt.Sprintf("%d", int(ttl.Seconds())))
			response.Error(c, http.StatusTooManyRequests, "RATE_LIMITED",
				fmt.Sprintf("Too many requests. Please retry after %d seconds", int(ttl.Seconds())), nil)
			return
		}

		c.Next()
	}
}

// AuthRateLimiter applies a stricter rate limit for auth endpoints.
func AuthRateLimiter(redis *cache.RedisClient) gin.HandlerFunc {
	return RateLimiter(redis, 5, time.Minute)
}

// GlobalRateLimiter applies a global rate limit.
func GlobalRateLimiter(redis *cache.RedisClient) gin.HandlerFunc {
	return RateLimiter(redis, 100, time.Minute)
}
