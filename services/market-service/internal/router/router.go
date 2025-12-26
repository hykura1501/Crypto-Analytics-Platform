package router

import (
	"net/http"

	"github.com/crypto-platform/market-service/internal/handler"
	"github.com/gin-gonic/gin"
)

func SetupRouter(marketHandler *handler.MarketHandler) *gin.Engine {
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
			c.JSON(http.StatusOK, gin.H{
				"status":  "ok",
				"service": "market-service",
			})
		})

		market := v1.Group("/market")
		{
			market.GET("/history", marketHandler.GetHistory)
			market.GET("/data", marketHandler.GetFromDB)
		}
	}

	// WebSocket endpoint
	router.GET("/ws/prices", marketHandler.WebSocketHandler)

	return router
}
