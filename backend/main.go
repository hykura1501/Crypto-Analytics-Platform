package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/hykura1501/crypto-analytics-backend/internal/database"
	"github.com/hykura1501/crypto-analytics-backend/internal/handlers"
	"github.com/hykura1501/crypto-analytics-backend/internal/websocket"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Initialize database
	db, err := database.Initialize()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Initialize Redis for caching
	redisClient := database.InitializeRedis()

	// Initialize WebSocket hub
	hub := websocket.NewHub()
	go hub.Run()

	// Setup Gin router
	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Initialize handlers
	h := handlers.NewHandler(db, redisClient, hub)

	// API routes
	api := r.Group("/api/v1")
	{
		// Trading pairs
		api.GET("/pairs", h.GetTradingPairs)
		api.GET("/pairs/:pair", h.GetPairInfo)

		// Price data
		api.GET("/price/:pair", h.GetCurrentPrice)
		api.GET("/price/:pair/history", h.GetPriceHistory)
		api.GET("/klines/:pair", h.GetKlines)

		// News
		api.GET("/news", h.GetNews)
		api.GET("/news/:id", h.GetNewsDetail)
		api.GET("/news/sources", h.GetNewsSources)

		// AI Analysis
		api.GET("/analysis/:pair", h.GetAnalysis)
		api.POST("/analysis/predict", h.PredictTrend)

		// Account management
		api.POST("/auth/register", h.Register)
		api.POST("/auth/login", h.Login)
		api.GET("/account/profile", h.GetProfile)
		api.PUT("/account/profile", h.UpdateProfile)
		api.GET("/account/watchlist", h.GetWatchlist)
		api.POST("/account/watchlist", h.AddToWatchlist)
		api.DELETE("/account/watchlist/:pair", h.RemoveFromWatchlist)
	}

	// WebSocket endpoint
	r.GET("/ws/price/:pair", h.HandleWebSocket)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
