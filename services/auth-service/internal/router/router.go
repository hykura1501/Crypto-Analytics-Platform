package router

import (
	"github.com/crypto-platform/auth-service/internal/handler"
	"github.com/crypto-platform/auth-service/internal/middleware"
	"github.com/crypto-platform/auth-service/pkg/utils"
	"github.com/gin-gonic/gin"
)

func SetupRouter(authHandler *handler.AuthHandler, jwtManager *utils.JWTManager) *gin.Engine {
	router := gin.Default()

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Infrastructure Health Check
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok", "service": "auth-service"})
		})

		// Public routes
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/logout", authHandler.Logout)
		}

		// Protected routes
		protected := v1.Group("/auth")
		protected.Use(middleware.AuthMiddleware(jwtManager))
		{
			protected.GET("/validate", authHandler.Validate)
			protected.GET("/me", authHandler.Me)
		}
	}

	return router
}
