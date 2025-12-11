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
		log.Printf("WebSocketProxy: Received request for path=%s, method=%s", c.Request.URL.Path, c.Request.Method)
		log.Printf("WebSocketProxy: Headers - Upgrade=%s, Connection=%s, Origin=%s", c.GetHeader("Upgrade"), c.GetHeader("Connection"), c.GetHeader("Origin"))

		// Set CORS headers before upgrade
		origin := c.GetHeader("Origin")
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		// Upgrade HTTP connection to WebSocket
		clientConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("Failed to upgrade WebSocket: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to upgrade connection"})
			return
		}
		log.Printf("WebSocketProxy: Successfully upgraded connection")
		defer clientConn.Close()

		// Connect to target WebSocket server
		// Use the original path from the request
		targetPath := c.Request.URL.Path
		targetWsURL := "ws://" + target.Host + targetPath
		log.Printf("WebSocketProxy: Connecting to target: %s", targetWsURL)

		targetConn, _, err := websocket.DefaultDialer.Dial(targetWsURL, nil)
		if err != nil {
			log.Printf("Failed to connect to target WebSocket: %v", err)
			clientConn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "Failed to connect to backend"))
			return
		}
		defer targetConn.Close()
		log.Printf("WebSocketProxy: Connected to target WebSocket")

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
		log.Printf("WebSocketProxy: Connection closed")
	}
}
