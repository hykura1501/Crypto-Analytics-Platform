package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/hykura1501/crypto-analytics-backend/internal/database"
	"github.com/hykura1501/crypto-analytics-backend/internal/handlers"
	"github.com/hykura1501/crypto-analytics-backend/internal/middleware"
	"github.com/hykura1501/crypto-analytics-backend/internal/services"
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

	// Start price updater service
	tradingPairs := getActiveTradingPairs(db)
	binanceService := services.NewBinanceService()
	priceUpdater := services.NewPriceUpdater(binanceService, redisClient, hub, tradingPairs)
	go priceUpdater.Start()

	// Setup Gin router
	r := gin.Default()

	// Global middleware
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.AuditLogger())
	r.Use(middleware.IPRateLimiter(redisClient, 100)) // 100 requests per minute per IP

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
		// Public endpoints (no auth required)
		public := api.Group("")
		{
			// Trading pairs
			public.GET("/pairs", h.GetTradingPairs)
			public.GET("/pairs/:pair", h.GetPairInfo)

			// Price data
			public.GET("/price/:pair", h.GetCurrentPrice)
			public.GET("/price/:pair/history", h.GetPriceHistory)
			public.GET("/klines/:pair", h.GetKlines)

			// News
			public.GET("/news", h.GetNews)
			public.GET("/news/:id", h.GetNewsDetail)
			public.GET("/news/sources", h.GetNewsSources)

			// AI Analysis (read-only)
			public.GET("/analysis/:pair", h.GetAnalysis)
		}

		// Auth endpoints
		auth := api.Group("/auth")
		{
			auth.POST("/register", h.Register)
			auth.POST("/login", h.Login)
			auth.POST("/refresh", h.RefreshToken)
		}

		// Protected endpoints (require authentication)
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			// AI Analysis (create)
			protected.POST("/analysis/predict", h.PredictTrend)

			// Account management
			protected.GET("/account/profile", h.GetProfile)
			protected.PUT("/account/profile", h.UpdateProfile)
			protected.GET("/account/watchlist", h.GetWatchlist)
			protected.POST("/account/watchlist", h.AddToWatchlist)
			protected.DELETE("/account/watchlist/:pair", h.RemoveFromWatchlist)
		}
	}

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":      "healthy",
			"connections": hub.GetTotalConnections(),
		})
	})

	// WebSocket endpoint
	r.GET("/ws/price/:pair", h.HandleWebSocket)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Printf("Monitoring %d trading pairs", len(tradingPairs))
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

// getActiveTradingPairs retrieves active trading pairs from database
func getActiveTradingPairs(db *sql.DB) []string {
	rows, err := db.Query("SELECT symbol FROM trading_pairs WHERE status = 'active'")
	if err != nil {
		log.Printf("Error fetching trading pairs: %v", err)
		return []string{"BTCUSDT", "ETHUSDT"} // Default pairs
	}
	defer rows.Close()

	var pairs []string
	for rows.Next() {
		var symbol string
		if err := rows.Scan(&symbol); err == nil {
			pairs = append(pairs, symbol)
		}
	}

	if len(pairs) == 0 {
		return []string{"BTCUSDT", "ETHUSDT"} // Default pairs
	}

	return pairs
}
