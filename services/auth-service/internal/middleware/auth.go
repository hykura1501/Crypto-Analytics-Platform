package middleware

import (
	"net/http"
	"strings"

	"github.com/crypto-platform/auth-service/pkg/utils"
	"github.com/gin-gonic/gin"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
	authorizationPayloadKey = "authorization_payload"
)

// AuthMiddleware validates JWT tokens
func AuthMiddleware(jwtManager *utils.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authorizationHeader := c.GetHeader(authorizationHeaderKey)

		if len(authorizationHeader) == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header is not provided"})
			c.Abort()
			return
		}

		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		authorizationType := strings.ToLower(fields[0])
		if authorizationType != authorizationTypeBearer {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unsupported authorization type"})
			c.Abort()
			return
		}

		accessToken := fields[1]
		claims, err := jwtManager.ValidateToken(accessToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		role := claims.Role
		if role == "" {
			role = "NORMAL"
		}
		c.Set(authorizationPayloadKey, claims)
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", role)
		c.Next()
	}
}

// GetUserID retrieves user ID from context
func GetUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	return userID.(uint), true
}

// GetEmail retrieves email from context
func GetEmail(c *gin.Context) (string, bool) {
	email, exists := c.Get("email")
	if !exists {
		return "", false
	}
	return email.(string), true
}

// GetRole retrieves role from context
func GetRole(c *gin.Context) (string, bool) {
	role, exists := c.Get("role")
	if !exists {
		return "", false
	}
	return role.(string), true
}

// RequireRole returns a middleware that ensures the user has one of the allowed roles
func RequireRole(allowed ...string) gin.HandlerFunc {
	set := make(map[string]bool)
	for _, r := range allowed {
		set[r] = true
	}
	return func(c *gin.Context) {
		role, _ := GetRole(c)
		if !set[role] {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden", "message": "insufficient permissions"})
			c.Abort()
			return
		}
		c.Next()
	}
}
