package websocket

import (
	"encoding/json"
	"io"
	"log"
	"time"
)

// Hub maintains the set of active clients and broadcasts messages to them
type Hub struct {
	// Registered clients per trading pair
	clients map[string]map[*Client]bool

	// Inbound messages from the clients
	broadcast chan *Message

	// Register requests from the clients
	Register chan *Client

	// Unregister requests from clients
	Unregister chan *Client
}

// Message represents a WebSocket message
type Message struct {
	Pair string      `json:"pair"`
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// Conn wraps the websocket connection
type Conn interface {
	ReadJSON(v interface{}) error
	WriteJSON(v interface{}) error
	Close() error
	SetReadDeadline(t time.Time) error
	SetWriteDeadline(t time.Time) error
	ReadMessage() (messageType int, p []byte, err error)
	WriteMessage(messageType int, data []byte) error
	NextWriter(messageType int) (io.WriteCloser, error)
	SetPongHandler(h func(appData string) error)
}

// Client is a middleman between the websocket connection and the hub
type Client struct {
	Hub  *Hub
	Pair string
	Conn Conn
	Send chan []byte
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]map[*Client]bool),
		broadcast:  make(chan *Message, 256),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			if h.clients[client.Pair] == nil {
				h.clients[client.Pair] = make(map[*Client]bool)
			}
			h.clients[client.Pair][client] = true
			log.Printf("Client registered for pair: %s", client.Pair)

		case client := <-h.Unregister:
			if clients, ok := h.clients[client.Pair]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.Send)
					if len(clients) == 0 {
						delete(h.clients, client.Pair)
					}
					log.Printf("Client unregistered for pair: %s", client.Pair)
				}
			}

		case message := <-h.broadcast:
			clients, ok := h.clients[message.Pair]
			if !ok {
				continue
			}

			data, err := json.Marshal(message)
			if err != nil {
				log.Printf("Error marshaling message: %v", err)
				continue
			}

			for client := range clients {
				select {
				case client.Send <- data:
				default:
					close(client.Send)
					delete(clients, client)
				}
			}
		}
	}
}

func (h *Hub) BroadcastPrice(pair string, priceData interface{}) {
	message := &Message{
		Pair: pair,
		Type: "price",
		Data: priceData,
	}
	select {
	case h.broadcast <- message:
	default:
		log.Println("Broadcast channel full, dropping message")
	}
}
