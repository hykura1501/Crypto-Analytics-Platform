package websocket

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisPubSub manages Redis publish/subscribe for distributed WebSocket
type RedisPubSub struct {
	client  *redis.Client
	hub     *Hub
	ctx     context.Context
	cancel  context.CancelFunc
	channel string
}

// PriceUpdate represents a price update message
type PriceUpdate struct {
	Pair      string      `json:"pair"`
	Price     string      `json:"price"`
	Timestamp int64       `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// NewRedisPubSub creates a new Redis pub/sub manager
func NewRedisPubSub(redisClient *redis.Client, hub *Hub) *RedisPubSub {
	ctx, cancel := context.WithCancel(context.Background())
	return &RedisPubSub{
		client:  redisClient,
		hub:     hub,
		ctx:     ctx,
		cancel:  cancel,
		channel: "price_updates",
	}
}

// Publish sends a price update to all instances via Redis
func (r *RedisPubSub) Publish(pair string, priceData interface{}) error {
	update := PriceUpdate{
		Pair:      pair,
		Data:      priceData,
		Timestamp: time.Now().Unix(),
	}

	data, err := json.Marshal(update)
	if err != nil {
		return err
	}

	return r.client.Publish(r.ctx, r.channel, data).Err()
}

// Subscribe listens for price updates from Redis and broadcasts to local clients
func (r *RedisPubSub) Subscribe() {
	pubsub := r.client.Subscribe(r.ctx, r.channel)
	defer pubsub.Close()

	log.Println("Started Redis subscription for price updates")

	ch := pubsub.Channel()
	for {
		select {
		case msg := <-ch:
			r.handleMessage(msg.Payload)
		case <-r.ctx.Done():
			log.Println("Stopped Redis subscription")
			return
		}
	}
}

// handleMessage processes incoming Redis messages
func (r *RedisPubSub) handleMessage(payload string) {
	var update PriceUpdate
	if err := json.Unmarshal([]byte(payload), &update); err != nil {
		log.Printf("Error unmarshaling price update: %v", err)
		return
	}

	// Broadcast to local WebSocket clients
	r.hub.BroadcastPrice(update.Pair, update.Data)
}

// Stop stops the Redis subscription
func (r *RedisPubSub) Stop() {
	r.cancel()
}

// GetConnectedClients returns the number of clients connected to this instance
func (h *Hub) GetConnectedClients() map[string]int {
	counts := make(map[string]int)
	for pair, clients := range h.clients {
		counts[pair] = len(clients)
	}
	return counts
}

// GetTotalConnections returns total connections across all pairs
func (h *Hub) GetTotalConnections() int {
	total := 0
	for _, clients := range h.clients {
		total += len(clients)
	}
	return total
}
