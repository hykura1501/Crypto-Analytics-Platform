package middleware

import (
	"log"
	"net/http"
	"strings"

	"api-gateway/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTAuth middleware to validate JWT tokens
func JWTAuth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip JWT validation for login, register endpoints, and WebSocket connections
		path := c.Request.URL.Path
		upgradeHeader := c.GetHeader("Upgrade")
		connectionHeader := c.GetHeader("Connection")

		// Debug log
		log.Printf("JWT Middleware: path=%s, upgrade=%s, connection=%s", path, upgradeHeader, connectionHeader)

		// Check for WebSocket upgrade header or /ws/ path
		// WebSocket requests have "Upgrade: websocket" header
		isWebSocket := strings.HasPrefix(path, "/ws/") ||
			upgradeHeader == "websocket" ||
			(strings.ToLower(connectionHeader) == "upgrade" && upgradeHeader != "")

		if strings.HasPrefix(path, "/auth/login") ||
			strings.HasPrefix(path, "/auth/register") ||
			strings.HasPrefix(path, "/health") ||
			isWebSocket {
			log.Printf("JWT Middleware: Skipping auth for path=%s (isWebSocket=%v)", path, isWebSocket)
			c.Next()
			return
		}

		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			c.Abort()
			return
		}

		// Parse and validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Extract claims and set user context
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Set("user_id", claims["user_id"])
			c.Set("username", claims["username"])
		}

		c.Next()
	}
}
