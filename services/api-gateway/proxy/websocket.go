package proxy

import (
	"log"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for WebSocket
	},
}

// WebSocketProxy proxies WebSocket connections to the target service
func WebSocketProxy(targetURL string) gin.HandlerFunc {
	target, err := url.Parse(targetURL)
	if err != nil {
		log.Fatalf("Invalid WebSocket target URL: %s", targetURL)
	}

	return func(c *gin.Context) {
		// Upgrade HTTP connection to WebSocket
		clientConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("Failed to upgrade WebSocket: %v", err)
			return
		}
		defer clientConn.Close()

		// Connect to target WebSocket server
		targetWsURL := "ws://" + target.Host + c.Request.URL.Path

		targetConn, _, err := websocket.DefaultDialer.Dial(targetWsURL, nil)
		if err != nil {
			log.Printf("Failed to connect to target WebSocket: %v", err)
			return
		}
		defer targetConn.Close()

		// Bidirectional proxy
		done := make(chan struct{}, 2)

		// Client -> Target
		go func() {
			defer func() { done <- struct{}{} }()
			for {
				messageType, message, err := clientConn.ReadMessage()
				if err != nil {
					return
				}
				if err := targetConn.WriteMessage(messageType, message); err != nil {
					return
				}
			}
		}()

		// Target -> Client
		go func() {
			defer func() { done <- struct{}{} }()
			for {
				messageType, message, err := targetConn.ReadMessage()
				if err != nil {
					return
				}
				if err := clientConn.WriteMessage(messageType, message); err != nil {
					return
				}
			}
		}()

		// Wait for either direction to close
		<-done
	}
}
