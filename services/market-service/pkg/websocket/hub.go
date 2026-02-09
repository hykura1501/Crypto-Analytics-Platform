package websocket

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// SubscriptionMessage represents a message from client to subscribe/unsubscribe
type SubscriptionMessage struct {
	Action string   `json:"action"` // "subscribe", "unsubscribe"
	Topics []string `json:"topics"`
}

// BroadcastMessage represents a message to be broadcasted to a specific topic
type BroadcastMessage struct {
	Topic   string
	Payload interface{}
}

// Hub maintains the set of active clients and broadcasts messages to them
type Hub struct {
	// clients map[*Client]bool // Removed global broadcast list
	topics      map[string]map[*Client]bool // Topic -> Set of Clients
	broadcast   chan *BroadcastMessage
	register    chan *Client
	unregister  chan *Client
	subscribe   chan *Subscription
	unsubscribe chan *Subscription
	mu          sync.RWMutex
}

// Subscription represents a client subscribing to a topic
type Subscription struct {
	Client *Client
	Topics []string
}

// Client represents a WebSocket client connection
type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan interface{}
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		topics:      make(map[string]map[*Client]bool),
		broadcast:   make(chan *BroadcastMessage, 256),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		subscribe:   make(chan *Subscription),
		unsubscribe: make(chan *Subscription),
	}
}

// Run starts the hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			// Just register the connection, no topics yet
			log.Printf("New client connected: %v", client.conn.RemoteAddr())

		case client := <-h.unregister:
			h.mu.Lock()
			// Remove client from all topics
			for topic, clients := range h.topics {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					if len(clients) == 0 {
						delete(h.topics, topic)
					}
				}
			}
			close(client.send)
			h.mu.Unlock()
			log.Printf("Client disconnected: %v", client.conn.RemoteAddr())

		case sub := <-h.subscribe:
			h.mu.Lock()
			for _, topic := range sub.Topics {
				if h.topics[topic] == nil {
					h.topics[topic] = make(map[*Client]bool)
				}
				h.topics[topic][sub.Client] = true
				log.Printf("Client %v subscribed to %s", sub.Client.conn.RemoteAddr(), topic)
			}
			h.mu.Unlock()

		case sub := <-h.unsubscribe:
			h.mu.Lock()
			for _, topic := range sub.Topics {
				if clients, ok := h.topics[topic]; ok {
					delete(clients, sub.Client)
					if len(clients) == 0 {
						delete(h.topics, topic)
					}
				}
				log.Printf("Client %v unsubscribed from %s", sub.Client.conn.RemoteAddr(), topic)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			if clients, ok := h.topics[message.Topic]; ok {
				deadClients := []*Client{}
				for client := range clients {
					select {
					case client.send <- message.Payload:
					default:
						// Client buffer full — mark for cleanup
						deadClients = append(deadClients, client)
					}
				}
				h.mu.RUnlock()

				// Clean up dead clients (need write lock)
				if len(deadClients) > 0 {
					h.mu.Lock()
					for _, client := range deadClients {
						// Remove client from all topics
						for topic, topicClients := range h.topics {
							delete(topicClients, client)
							if len(topicClients) == 0 {
								delete(h.topics, topic)
							}
						}
						close(client.send)
						log.Printf("Removed unresponsive client: %v", client.conn.RemoteAddr())
					}
					h.mu.Unlock()
				}
			} else {
				h.mu.RUnlock()
			}
		}
	}
}

// BroadcastToTopic sends a message to clients subscribed to the topic
func (h *Hub) BroadcastToTopic(topic string, payload interface{}) {
	h.broadcast <- &BroadcastMessage{
		Topic:   topic,
		Payload: payload,
	}
}

// RegisterClient registers a new client
func (h *Hub) RegisterClient(conn *websocket.Conn) *Client {
	client := &Client{
		hub:  h,
		conn: conn,
		send: make(chan interface{}, 256),
	}
	h.register <- client
	return client
}

// ReadPump pumps messages from the client
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		// Handle subscription messages
		var subMsg SubscriptionMessage
		if err := json.Unmarshal(message, &subMsg); err != nil {
			log.Printf("Invalid message format: %v", err)
			continue
		}

		switch subMsg.Action {
		case "subscribe":
			c.hub.subscribe <- &Subscription{Client: c, Topics: subMsg.Topics}
		case "unsubscribe":
			c.hub.unsubscribe <- &Subscription{Client: c, Topics: subMsg.Topics}
		}
	}
}

// WritePump pumps messages to the client
func (c *Client) WritePump() {
	defer func() {
		c.conn.Close()
	}()

	for message := range c.send {
		err := c.conn.WriteJSON(message)
		if err != nil {
			log.Printf("Error writing to client: %v", err)
			return
		}
	}
}
