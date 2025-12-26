package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/joho/godotenv"

	"github.com/crypto-platform/crawler-service/config"
	"github.com/crypto-platform/crawler-service/internal/api"
	"github.com/crypto-platform/crawler-service/internal/crawler"
	"github.com/crypto-platform/crawler-service/internal/db"
	"github.com/crypto-platform/crawler-service/internal/handler"
	ckafka "github.com/crypto-platform/crawler-service/internal/kafka"
)

type CrawlerSource struct {
	SourceID string `db:"source_id"`
	RssURL   string `db:"rss_url"`
}

func InitializeCrawlerSource(cfg *config.Config, database *sql.DB, crawlService *crawler.Service) {
	// Insert some soure_id to db if not exist
	crawlerSource := []CrawlerSource{
		{
			SourceID: "CoinDesk",
			RssURL:   cfg.RSS.CoinDeskURL,
		},
		{
			SourceID: "CoinTelegraph",
			RssURL:   cfg.RSS.CoinTelegraphURL,
		},
		{
			SourceID: "VNExpress",
			RssURL:   cfg.RSS.VNExpressURL,
		},
		{
			SourceID: "VnEconomy",
			RssURL:   cfg.RSS.VnEconomyURL,
		},
	}
	for _, source := range crawlerSource {
		result, err := database.Exec(`
			INSERT INTO sources (source_id, rss_url)
			VALUES ($1, $2)
			ON CONFLICT (source_id) DO NOTHING
		`, source.SourceID, source.RssURL)
		if err != nil {
			log.Fatalf("Failed to insert crawler source %s: %v", source.SourceID, err)
		}

		// Check if row was actually inserted (not a conflict)
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			log.Printf("⏭️  Source %s already exists, skipping Kafka analysis", source.SourceID)
			continue
		}

		log.Printf("Initialized crawler source: %s", source.SourceID)

		// Trigger analysis
		if err := crawlService.AnalyzeSource(source.SourceID, source.RssURL); err != nil {
			log.Printf("Failed to analyze source %s: %v", source.SourceID, err)
		}
	}
}

func main() {
	// Load .env file if present
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found for crawler-service, using system environment variables")
	}

	cfg := config.Load()

	// Kết nối DB (sẽ tự động chạy migrations)
	database, err := db.Connect(&cfg.Database)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	defer database.Close()

	// Kafka producer
	var producer *ckafka.Producer
	if cfg.Kafka.Broker != "" && cfg.Kafka.Topic != "" {
		producer = ckafka.NewProducer(cfg.Kafka.Broker, cfg.Kafka.Topic)
		defer producer.Close()
	}

	crawlService := crawler.NewService(cfg, database, producer)
	sourceHandler := handler.NewSourceHandler(database)

	// Initialize crawler sources and send RSS XML to Kafka
	go InitializeCrawlerSource(cfg, database, crawlService)

	// Run scheduler
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	time.Sleep(30 * time.Second)
	log.Println("Waiting for 30 seconds to initialize crawler sources...")

	go startScheduler(ctx, crawlService, cfg.Crawler.IntervalMinutes)

	// Initialize API Server
	server := api.NewServer(cfg, database, crawlService, sourceHandler)

	log.Printf("Starting crawler-service (Go + Colly) on :%s", cfg.Server.Port)
	if err := server.Run(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func startScheduler(ctx context.Context, svc *crawler.Service, intervalMinutes int) {
	if intervalMinutes <= 0 {
		intervalMinutes = 10
	}

	ticker := time.NewTicker(time.Duration(intervalMinutes) * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		log.Println("Starting scheduled crawl...")
		if _, err := svc.CrawlOnce(ctx); err != nil {
			log.Printf("Scheduled crawl error: %v", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
