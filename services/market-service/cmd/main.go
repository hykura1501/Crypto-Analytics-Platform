package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/crypto-platform/market-service/config"
	"github.com/crypto-platform/market-service/internal/handler"
	"github.com/crypto-platform/market-service/internal/model"
	"github.com/crypto-platform/market-service/internal/repository"
	"github.com/crypto-platform/market-service/internal/router"
	"github.com/crypto-platform/market-service/internal/service"
	"github.com/crypto-platform/market-service/pkg/binance"
	"github.com/crypto-platform/market-service/pkg/kafka"
	ws "github.com/crypto-platform/market-service/pkg/websocket"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Load .env file if present
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found for market-service, using system environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Set Gin mode
	gin.SetMode(cfg.Server.GinMode)

	// Connect to database
	db, err := initDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize components
	marketRepo := repository.NewMarketRepository(db)
	binanceClient := binance.NewClient(cfg.Binance.APIURL)

	// Default interval for WebSocket (can be made configurable)
	binanceWS := binance.NewWebSocketClient(cfg.Binance.WSURL, cfg.Binance.Symbols, cfg.Binance.Intervals)

	kafkaProducer := kafka.NewProducer(cfg.Kafka.Broker, cfg.Kafka.Topic)
	defer kafkaProducer.Close()

	// Create WebSocket hub for frontend clients
	wsHub := ws.NewHub()
	go wsHub.Run()

	// Create market service
	marketService := service.NewMarketService(
		marketRepo,
		binanceClient,
		binanceWS,
		kafkaProducer,
		wsHub,
	)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start real-time data stream
	if err := marketService.StartRealtimeStream(ctx); err != nil {
		log.Fatalf("Failed to start real-time stream: %v", err)
	}

	// Initialize handlers
	marketHandler := handler.NewMarketHandler(marketService, wsHub)

	// Setup router
	r := router.SetupRouter(marketHandler)

	// Setup graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down gracefully...")
		cancel()
		kafkaProducer.Close()
		os.Exit(0)
	}()

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("Starting market-service on %s", addr)
	log.Printf("Tracking symbols: %v", cfg.Binance.Symbols)
	log.Printf("WebSocket endpoint: ws://localhost%s/ws/prices", addr)

	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func initDatabase(cfg *config.Config) (*gorm.DB, error) {
	dsn := cfg.Database.DSN()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Tắt GORM logging
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Database connection established")
	return db, nil
}

func runMigrations(db *gorm.DB) error {
	log.Println("Running database migrations...")

	if err := db.AutoMigrate(
		&model.MarketPrice{},
	); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Create index for time-series queries (TimescaleDB optimization)
	db.Exec("CREATE INDEX IF NOT EXISTS idx_market_prices_symbol_time ON market_prices(symbol, time DESC);")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_market_prices_time ON market_prices(time DESC);")

	log.Println("Migrations completed successfully")
	return nil
}
