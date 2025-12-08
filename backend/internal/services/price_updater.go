package services

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/hykura1501/crypto-analytics-backend/internal/websocket"
	"github.com/redis/go-redis/v9"
)

// PriceUpdater fetches prices from Binance and publishes to Redis
type PriceUpdater struct {
	binanceService *BinanceService
	redisPubSub    *websocket.RedisPubSub
	pairs          []string
	interval       time.Duration
	ctx            context.Context
	cancel         context.CancelFunc
}

// NewPriceUpdater creates a new price updater service
func NewPriceUpdater(binanceService *BinanceService, redisClient *redis.Client, hub *websocket.Hub, pairs []string) *PriceUpdater {
	ctx, cancel := context.WithCancel(context.Background())
	return &PriceUpdater{
		binanceService: binanceService,
		redisPubSub:    websocket.NewRedisPubSub(redisClient, hub),
		pairs:          pairs,
		interval:       1 * time.Second, // Update every second
		ctx:            ctx,
		cancel:         cancel,
	}
}

// Start begins the price update loop
func (pu *PriceUpdater) Start() {
	log.Println("Starting price updater service...")

	// Start Redis subscription in a separate goroutine
	go pu.redisPubSub.Subscribe()

	// Start price fetching loop
	ticker := time.NewTicker(pu.interval)
	defer ticker.Stop()

	// Fetch immediately on start
	pu.updatePrices()

	for {
		select {
		case <-ticker.C:
			pu.updatePrices()
		case <-pu.ctx.Done():
			log.Println("Stopped price updater service")
			return
		}
	}
}

// updatePrices fetches current prices for all pairs and publishes to Redis
func (pu *PriceUpdater) updatePrices() {
	for _, pair := range pu.pairs {
		go func(p string) {
			priceResp, err := pu.binanceService.GetCurrentPrice(p)
			if err != nil {
				log.Printf("Error fetching price for %s: %v", p, err)
				return
			}

			priceData := map[string]interface{}{
				"pair":      p,
				"price":     priceResp.Price,
				"timestamp": time.Now().Unix(),
			}

			// Publish to Redis (will be received by all instances)
			if err := pu.redisPubSub.Publish(p, priceData); err != nil {
				log.Printf("Error publishing price for %s: %v", p, err)
			}
		}(pair)
	}
}

// Stop stops the price updater
func (pu *PriceUpdater) Stop() {
	pu.cancel()
	pu.redisPubSub.Stop()
}

// SetInterval changes the update interval
func (pu *PriceUpdater) SetInterval(interval time.Duration) {
	pu.interval = interval
}

// AddPair adds a new pair to monitor
func (pu *PriceUpdater) AddPair(pair string) {
	for _, p := range pu.pairs {
		if p == pair {
			return // Already exists
		}
	}
	pu.pairs = append(pu.pairs, pair)
	log.Printf("Added pair %s to price updater", pair)
}

// RemovePair removes a pair from monitoring
func (pu *PriceUpdater) RemovePair(pair string) {
	for i, p := range pu.pairs {
		if p == pair {
			pu.pairs = append(pu.pairs[:i], pu.pairs[i+1:]...)
			log.Printf("Removed pair %s from price updater", pair)
			return
		}
	}
}

// GetMonitoredPairs returns the list of monitored pairs
func (pu *PriceUpdater) GetMonitoredPairs() []string {
	return pu.pairs
}
