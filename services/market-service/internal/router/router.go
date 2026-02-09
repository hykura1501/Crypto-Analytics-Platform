package router

import (
	"net/http"

	"github.com/crypto-platform/market-service/internal/handler"
	"github.com/gin-gonic/gin"
)

func SetupRouter(marketHandler *handler.MarketHandler) *gin.Engine {
	router := gin.Default()

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
