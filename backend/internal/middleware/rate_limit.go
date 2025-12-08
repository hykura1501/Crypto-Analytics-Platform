package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"
)

// RateLimiter creates a simple rate limiting middleware
func RateLimiter(requestsPerMinute int) gin.HandlerFunc {
	limiter := rate.NewLimiter(rate.Every(time.Minute/time.Duration(requestsPerMinute)), requestsPerMinute)

	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please try again later.",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// IPRateLimiter creates per-IP rate limiting using Redis
func IPRateLimiter(redisClient *redis.Client, requestsPerMinute int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := fmt.Sprintf("rate_limit:%s", ip)

		ctx := context.Background()
		count, err := redisClient.Incr(ctx, key).Result()
		if err != nil {
			// If Redis fails, allow the request (fail open)
			c.Next()
			return
		}

		if count == 1 {
			// Set expiration on first request
			redisClient.Expire(ctx, key, time.Minute)
		}

		if count > int64(requestsPerMinute) {
			retryAfter := 60 - int(time.Now().Unix()%60)
			c.Header("Retry-After", strconv.Itoa(retryAfter))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"retry_after": retryAfter,
			})
			c.Abort()
			return
		}

		// Add rate limit headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(requestsPerMinute))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(int64(requestsPerMinute)-count, 10))

		c.Next()
	}
}

// UserRateLimiter creates per-user rate limiting
func UserRateLimiter(redisClient *redis.Client, requestsPerMinute int) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			// Not authenticated, skip user rate limit
			c.Next()
			return
		}

		key := fmt.Sprintf("rate_limit:user:%v", userID)

		ctx := context.Background()
		count, err := redisClient.Incr(ctx, key).Result()
		if err != nil {
			c.Next()
			return
		}

		if count == 1 {
			redisClient.Expire(ctx, key, time.Minute)
		}

		if count > int64(requestsPerMinute) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "User rate limit exceeded",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// TokenBucket implements token bucket rate limiting
type TokenBucket struct {
	redis      *redis.Client
	capacity   int
	refillRate int // tokens per second
}

// NewTokenBucket creates a new token bucket rate limiter
func NewTokenBucket(redisClient *redis.Client, capacity, refillRate int) *TokenBucket {
	return &TokenBucket{
		redis:      redisClient,
		capacity:   capacity,
		refillRate: refillRate,
	}
}

// Allow checks if a request is allowed
func (tb *TokenBucket) Allow(key string) bool {
	ctx := context.Background()
	now := time.Now().Unix()

	// Lua script for atomic token bucket check
	script := `
		local key = KEYS[1]
		local capacity = tonumber(ARGV[1])
		local refill_rate = tonumber(ARGV[2])
		local now = tonumber(ARGV[3])
		
		local bucket = redis.call('HGETALL', key)
		local tokens = capacity
		local last_refill = now
		
		if #bucket > 0 then
			tokens = tonumber(bucket[2])
			last_refill = tonumber(bucket[4])
		end
		
		-- Refill tokens
		local elapsed = now - last_refill
		tokens = math.min(capacity, tokens + (elapsed * refill_rate))
		
		if tokens >= 1 then
			tokens = tokens - 1
			redis.call('HSET', key, 'tokens', tokens, 'last_refill', now)
			redis.call('EXPIRE', key, 3600)
			return 1
		else
			return 0
		end
	`

	result, err := tb.redis.Eval(ctx, script, []string{key}, tb.capacity, tb.refillRate, now).Int()
	if err != nil {
		return true // Fail open
	}

	return result == 1
}

// TokenBucketMiddleware creates a token bucket rate limiting middleware
func TokenBucketMiddleware(redisClient *redis.Client, capacity, refillRate int) gin.HandlerFunc {
	bucket := NewTokenBucket(redisClient, capacity, refillRate)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := fmt.Sprintf("token_bucket:%s", ip)

		if !bucket.Allow(key) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
