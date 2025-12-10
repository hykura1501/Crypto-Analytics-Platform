package main

import (
	"log"
	"time"

	"api-gateway/config"
	"api-gateway/middleware"
	"api-gateway/proxy"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize Gin router
	r := gin.Default()

	// CORS configuration
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Apply JWT middleware globally
	r.Use(middleware.JWTAuth(cfg))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"service":   "api-gateway",
			"timestamp": time.Now().Unix(),
		})
	})

	// Route: /auth/* -> auth-service:8081
	authGroup := r.Group("/auth")
	{
		authProxy := proxy.ReverseProxy(cfg.AuthServiceURL)
		authGroup.Any("/*path", proxy.StripPrefix("/auth", authProxy))
	}

	// Route: /market/* -> market-service:8082
	marketGroup := r.Group("/market")
	{
		marketProxy := proxy.ReverseProxy(cfg.MarketServiceURL)
		marketGroup.Any("/*path", proxy.StripPrefix("/market", marketProxy))
	}

	// Route: /news/* -> crawler-service:8083
	newsGroup := r.Group("/news")
	{
		newsProxy := proxy.ReverseProxy(cfg.CrawlerServiceURL)
		newsGroup.Any("/*path", proxy.StripPrefix("/news", newsProxy))
	}

	// WebSocket Route: /ws/* -> market-service:8082
	wsGroup := r.Group("/ws")
	{
		wsProxy := proxy.WebSocketProxy(cfg.MarketServiceURL)
		wsGroup.GET("/*path", wsProxy)
	}

	// Start server
	addr := ":" + cfg.Port
	log.Printf("🚀 API Gateway starting on %s", addr)
	log.Printf("📡 Proxying:")
	log.Printf("   /auth/*   → %s", cfg.AuthServiceURL)
	log.Printf("   /market/* → %s", cfg.MarketServiceURL)
	log.Printf("   /news/*   → %s", cfg.CrawlerServiceURL)
	log.Printf("   /ws/*     → %s (WebSocket)", cfg.MarketServiceURL)
	log.Printf("🔐 JWT Auth: Enabled (except /auth/login, /auth/register)")
	log.Printf("🌐 CORS: %v", cfg.AllowedOrigins)

	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
