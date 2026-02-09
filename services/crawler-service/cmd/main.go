package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/crypto-platform/crawler-service/config"
	"github.com/crypto-platform/crawler-service/internal/crawler"
	"github.com/crypto-platform/crawler-service/internal/db"
	ckafka "github.com/crypto-platform/crawler-service/internal/kafka"
)

func main() {
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

	// Run scheduler
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go startScheduler(ctx, crawlService, cfg.Crawler.IntervalMinutes)

	// HTTP server (health + manual trigger)
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "crawler-service-go",
		})
	})

	r.POST("/crawl/once", func(c *gin.Context) {
		count, err := crawlService.CrawlOnce(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"saved": count,
		})
	})

	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("Starting crawler-service (Go + Colly) on %s", addr)

	// graceful shutdown
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutting down crawler-service...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
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
