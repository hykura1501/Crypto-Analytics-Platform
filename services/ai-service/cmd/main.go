package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/crypto-platform/ai-service/config"
	"github.com/crypto-platform/ai-service/internal/api"
	"github.com/crypto-platform/ai-service/internal/db"
	"github.com/crypto-platform/ai-service/internal/gemini"
	"github.com/crypto-platform/ai-service/internal/handler"
	ckafka "github.com/crypto-platform/ai-service/internal/kafka"
	"github.com/crypto-platform/ai-service/internal/sentiment"
	"github.com/joho/godotenv"
)

func main() {
	log.Println("🚀 Starting AI Service...")
	log.Println("🧠 Sentiment Analysis Engine")

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

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

	// Initialize dependencies and handlers
	sentimentAnalyzer := sentiment.NewAnalyzer()
	newsHandler := handler.NewNewsHandler(db.DB, sentimentAnalyzer) // Inject global DB

	geminiClient, err := gemini.NewGeminiClient(cfg.Gemini.APIKey, cfg.Gemini.Model)
	if err != nil {
		log.Fatalf("Failed to create Gemini client: %v", err)
	}

	rssHandler := handler.NewRssHandler(db.DB, geminiClient) // Inject global DB

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

	// Create API server
	apiServer := api.NewServer(cfg, db.DB, rssHandler)
	go func() {
		if err := apiServer.Run(); err != nil {
			log.Fatalf("Failed to run API server: %v", err)
		}
	}()

	// Main consumption loop
	for {
		select {
		case <-ctx.Done():
			log.Println("AI Service stopped.")
			return
		default:
			// Read message without timeout
			msg, err := consumer.ReadMessage(ctx)

			if err != nil {
				if err == context.Canceled {
					log.Println("AI Service stopped.")
					return
				}
				log.Printf("Error reading message: %v (will retry)", err)
				time.Sleep(2 * time.Second)
				continue
			}

			// Route message based on Topic
			switch msg.Topic {
			case config.KafkaTopicNewsNewArticle:
				newsHandler.Handle(ctx, msg.Value)

			case config.KafkaTopicNewsAnalyzeRssStructure:
				var rssStructure handler.RssMessage
				err := json.Unmarshal(msg.Value, &rssStructure)
				if err != nil {
					log.Printf("Error unmarshalling message: %v", err)
					continue
				}
				rssHandler.Handle(ctx, rssStructure)

			default:
				log.Printf("⚠️ Received message from unknown topic: %s", msg.Topic)
			}
		}
	}
}
