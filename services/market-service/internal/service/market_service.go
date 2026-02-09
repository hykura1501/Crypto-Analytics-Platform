package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/crypto-platform/market-service/internal/model"
	"github.com/crypto-platform/market-service/internal/repository"
	"github.com/crypto-platform/market-service/pkg/binance"
	"github.com/crypto-platform/market-service/pkg/kafka"
	ws "github.com/crypto-platform/market-service/pkg/websocket"
)

type MarketService struct {
	repo          repository.MarketRepository
	binanceClient *binance.Client
	binanceWS     *binance.WebSocketClient
	kafkaProducer *kafka.Producer
	wsHub         *ws.Hub
}

func NewMarketService(
	repo repository.MarketRepository,
	binanceClient *binance.Client,
	binanceWS *binance.WebSocketClient,
	kafkaProducer *kafka.Producer,
	wsHub *ws.Hub,
) *MarketService {
	return &MarketService{
		repo:          repo,
		binanceClient: binanceClient,
		binanceWS:     binanceWS,
		kafkaProducer: kafkaProducer,
		wsHub:         wsHub,
	}
}

// GetHistoricalData fetches historical kline data from Binance API
func (s *MarketService) GetHistoricalData(symbol, interval string, limit int, from, to int64) ([]*model.MarketPrice, error) {
	log.Printf("Fetching historical data: symbol=%s, interval=%s, limit=%d", symbol, interval, limit)

	prices, err := s.binanceClient.GetKlines(symbol, interval, limit, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from Binance: %w", err)
	}

	// Save to database (optional - can be async)
	if len(prices) > 0 {
		go func() {
			if err := s.repo.BulkCreate(prices); err != nil {
				log.Printf("Failed to save historical data: %v", err)
			} else {
				log.Printf("Saved %d historical records to database", len(prices))
			}
		}()
	}

	return prices, nil
}

// GetFromDatabase retrieves data from database
func (s *MarketService) GetFromDatabase(symbol, interval string, from, to time.Time, limit int) ([]*model.MarketPrice, error) {
	if !from.IsZero() && !to.IsZero() {
		return s.repo.FindBySymbolAndTimeRange(symbol, interval, from, to)
	}
	return s.repo.FindLatest(symbol, interval, limit)
}

// StartRealtimeStream starts consuming from Binance WebSocket and publishing to Kafka and clients
func (s *MarketService) StartRealtimeStream(ctx context.Context) error {
	log.Println("Starting real-time market data stream...")

	// Ensure Kafka topic exists
	if err := s.kafkaProducer.EnsureTopic(ctx); err != nil {
		log.Printf("Warning: Could not ensure Kafka topic: %v", err)
	}

	// Connect to Binance WebSocket
	if err := s.binanceWS.Connect(); err != nil {
		return fmt.Errorf("failed to connect to Binance WebSocket: %w", err)
	}

	// Process messages
	go func() {
		lastLogTime := time.Now()
		messageCount := 0
		reconnectBackoff := 5 * time.Second
		maxBackoff := 5 * time.Minute

		for {
			select {
			case <-ctx.Done():
				log.Println("Stopping real-time stream...")
				s.binanceWS.Close()
				return

			case price := <-s.binanceWS.GetMessageChannel():
				if price == nil {
					continue
				}

				messageCount++
				// Log only every 60 seconds to reduce spam
				if time.Since(lastLogTime) >= 60*time.Second {
					log.Printf("Processed %d messages in last minute. Latest: %s %.2f",
						messageCount, price.Symbol, price.Close)
					messageCount = 0
					lastLogTime = time.Now()
				}

				// 1. Publish to Kafka
				go func(p *model.MarketPrice) {
					if err := s.kafkaProducer.Publish(ctx, p.Symbol, p); err != nil {
						log.Printf("❌ Kafka error: %v", err)
					} else {
						// log.Printf("✅ Kafka: %s %.2f", p.Symbol, p.Close)
					}
				}(price)

				// 2. Broadcast to WebSocket clients
				// Topic format: market:{symbol}:{interval}
				topic := fmt.Sprintf("market:%s:%s", price.Symbol, price.Interval)
				s.wsHub.BroadcastToTopic(topic, price)

				// 3. Save to database (upsert)
				go func(p *model.MarketPrice) {
					if err := s.repo.Create(p); err != nil {
						log.Printf("❌ DB save error for %s/%s at %v: %v", p.Symbol, p.Interval, p.Time, err)
					}
				}(price)

			case err := <-s.binanceWS.GetErrorChannel():
				if err != nil {
					log.Printf("WebSocket error: %v. Reconnecting in %v...", err, reconnectBackoff)
					time.Sleep(reconnectBackoff)
					if err := s.binanceWS.Connect(); err != nil {
						log.Printf("Reconnection failed: %v", err)
						// Increase backoff for next retry
						reconnectBackoff = reconnectBackoff * 2
						if reconnectBackoff > maxBackoff {
							reconnectBackoff = maxBackoff
						}
					} else {
						// Reset backoff on successful reconnection
						reconnectBackoff = 5 * time.Second
						log.Println("✅ Successfully reconnected to Binance WebSocket")
					}
				}
			}
		}
	}()

	return nil
}
