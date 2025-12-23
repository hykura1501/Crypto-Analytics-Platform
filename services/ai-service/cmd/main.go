package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/crypto-platform/ai-service/config"
	"github.com/crypto-platform/ai-service/internal/causal"
	"github.com/crypto-platform/ai-service/internal/db"
	ckafka "github.com/crypto-platform/ai-service/internal/kafka"
	"github.com/crypto-platform/ai-service/internal/sentiment"
)

func main() {
	log.Println("🚀 Starting AI Service...")
	log.Println("🧠 Sentiment Analysis Engine")

	// Load configuration
	cfg := config.Load()

	// Initialize database
	log.Println("Initializing database...")
	if err := db.Init(cfg); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create Kafka consumer
	log.Println("Connecting to Kafka...")
	consumer, err := ckafka.NewConsumer(cfg)
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer: %v", err)
	}
	defer consumer.Close()

	// Initialize sentiment analyzer
	sentimentAnalyzer := sentiment.NewAnalyzer()

	log.Println(strings.Repeat("=", 60))
	log.Println("AI Service is ready. Waiting for news messages...")
	log.Println(strings.Repeat("=", 60))

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down AI Service...")
		cancel()
	}()

	// Main consumption loop
	for {
		select {
		case <-ctx.Done():
			log.Println("AI Service stopped.")
			return
		default:
			// Read message without timeout - this will block until a message is received
			msg, err := consumer.ReadMessage(ctx)

			if err != nil {
				if err == context.Canceled {
					// Service is shutting down
					log.Println("AI Service stopped.")
					return
				}
				// Log error but continue
				log.Printf("Error reading message: %v (will retry)", err)
				time.Sleep(2 * time.Second)
				continue
			}

			processNewsMessage(ctx, msg, sentimentAnalyzer)
		}
	}
}

func processNewsMessage(ctx context.Context, msg *ckafka.NewsMessage, analyzer *sentiment.Analyzer) {
	if msg.NewsID == 0 || msg.Content == "" {
		log.Printf("Invalid message: news_id=%d", msg.NewsID)
		return
	}

	log.Printf("📰 Processing news #%d: %s...", msg.NewsID, truncate(msg.Title, 50))

	// Analyze sentiment
	sentimentResult := analyzer.Analyze(msg.Title + " " + msg.Content)
	log.Printf(
		"🎭 Sentiment: %s (score: %.3f)",
		sentimentResult.Label,
		sentimentResult.Compound,
	)

	// Update database with sentiment score
	query := `UPDATE articles SET sentiment_score = $1 WHERE id = $2`
	_, err := db.DB.Exec(query, sentimentResult.Compound, msg.NewsID)
	if err != nil {
		log.Printf("Error updating sentiment score for article #%d: %v", msg.NewsID, err)
		return
	}

	log.Printf("✅ Updated article #%d with sentiment score", msg.NewsID)

	// Get article details for causal analysis
	var article struct {
		ID          int
		Entities    sql.NullString // Use NullString to handle NULL values
		PublishedAt sql.NullTime
	}

	query = `SELECT id, entities, published_at FROM articles WHERE id = $1`
	err = db.DB.QueryRow(query, msg.NewsID).Scan(
		&article.ID,
		&article.Entities,
		&article.PublishedAt,
	)

	if err == sql.ErrNoRows {
		log.Printf("Article #%d not found in database", msg.NewsID)
		return
	}
	if err != nil {
		log.Printf("Error fetching article #%d: %v", msg.NewsID, err)
		return
	}

	// Causal analysis (if article has entities and published time)
	if article.Entities.Valid && article.Entities.String != "" && article.PublishedAt.Valid {
		var entities map[string]interface{}
		if err := json.Unmarshal([]byte(article.Entities.String), &entities); err == nil {
			symbols := causal.ExtractSymbolsFromEntities(entities)

			for _, symbol := range symbols {
				causalResult, err := causal.AlignNewsWithPrice(
					db.DB,
					msg.NewsID,
					msg.Title,
					article.PublishedAt.Time,
					symbol,
				)

				if err != nil {
					log.Printf("Error in causal analysis: %v", err)
					continue
				}

				if causalResult != nil {
					log.Printf(
						"📈 %s: %.2f → %.2f (%.2f%%, %s)",
						causalResult.Symbol,
						causalResult.PriceBefore,
						causalResult.PriceAfter,
						causalResult.ChangePct,
						causalResult.Direction,
					)
				}
			}
		}
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}
