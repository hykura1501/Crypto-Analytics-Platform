package main

import (
	"log"
	"strings"
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

	// Register WebSocket routes FIRST, before any middleware
	// This ensures they are not affected by any middleware
	// Note: Only register /ws/*path, it will match /ws/prices and other paths
	wsProxy := proxy.WebSocketProxy(cfg.MarketServiceURL)
	r.GET("/ws/*path", wsProxy)

	// CORS configuration - must be before other middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Requested-With", "Accept"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health check (no auth required)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"service":   "api-gateway",
			"timestamp": time.Now().Unix(),
		})
	})

	// Route: /auth/* -> auth-service:8081/api/v1/auth/* (no auth required for login/register)
	authGroup := r.Group("/auth")
	{
		authProxy := proxy.ReverseProxy(cfg.AuthServiceURL)
		authGroup.Any("/*path", func(c *gin.Context) {
			// Rewrite path: /auth/login -> /api/v1/auth/login
			originalPath := c.Request.URL.Path
			c.Request.URL.Path = "/api/v1" + originalPath
			authProxy(c)
			c.Request.URL.Path = originalPath // Restore for logging
		})
	}

	// Protected routes (require JWT auth) - apply middleware to specific groups only
	// Route: /market/* -> market-service:8082/api/v1/market/*
	marketGroup := r.Group("/market")
	marketGroup.Use(middleware.JWTAuth(cfg))
	{
		marketProxy := proxy.ReverseProxy(cfg.MarketServiceURL)
		marketGroup.Any("/*path", func(c *gin.Context) {
			// Rewrite path: /market/api/v1/market/history -> /api/v1/market/history
			// or /market/history -> /api/v1/market/history
			originalPath := c.Request.URL.Path
			if strings.HasPrefix(originalPath, "/market/api/v1/market") {
				// Already has full path, just remove /market prefix
				c.Request.URL.Path = strings.TrimPrefix(originalPath, "/market")
			} else if strings.HasPrefix(originalPath, "/market/") {
				// /market/history -> /api/v1/market/history
				c.Request.URL.Path = "/api/v1" + originalPath
			} else {
				// Fallback
				c.Request.URL.Path = "/api/v1/market" + originalPath
			}
			marketProxy(c)
			c.Request.URL.Path = originalPath // Restore
		})
	}

	// Route: /news/* -> crawler-service:8083
	newsGroup := r.Group("/news")
	newsGroup.Use(middleware.JWTAuth(cfg))
	{
		newsProxy := proxy.ReverseProxy(cfg.CrawlerServiceURL)
		newsGroup.Any("/*path", proxy.StripPrefix("/news", newsProxy))
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
