package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// AuditLogger logs all requests for security analysis
func AuditLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Log after request
		latency := time.Since(start)
		statusCode := c.Writer.Status()

		userID, exists := c.Get("user_id")
		if !exists {
			userID = "anonymous"
		}

		userEmail, _ := c.Get("user_email")

		log.Printf("[AUDIT] method=%s path=%s status=%d latency=%v user_id=%v email=%v ip=%s user_agent=%s",
			c.Request.Method,
			c.Request.URL.Path,
			statusCode,
			latency,
			userID,
			userEmail,
			c.ClientIP(),
			c.Request.UserAgent(),
		)

		// Log errors separately
		if statusCode >= 400 {
			log.Printf("[ERROR] method=%s path=%s status=%d user_id=%v ip=%s errors=%v",
				c.Request.Method,
				c.Request.URL.Path,
				statusCode,
				userID,
				c.ClientIP(),
				c.Errors.String(),
			)
		}
	}
}

// SecurityHeaders adds security headers to responses
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString generates a random string of given length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
