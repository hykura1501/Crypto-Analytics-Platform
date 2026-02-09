package binance

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/crypto-platform/market-service/internal/model"
	"github.com/gorilla/websocket"
)

type WebSocketClient struct {
	wsURL       string
	symbols     []string
	interval    string
	conn        *websocket.Conn
	messageChan chan *model.MarketPrice
	errorChan   chan error
	done        chan struct{}
}

func NewWebSocketClient(wsURL string, symbols []string, interval string) *WebSocketClient {
	return &WebSocketClient{
		wsURL:       wsURL,
		symbols:     symbols,
		interval:    interval,
		messageChan: make(chan *model.MarketPrice, 100),
		errorChan:   make(chan error, 10),
		done:        make(chan struct{}),
	}
}

// Connect establishes WebSocket connection to Binance
func (ws *WebSocketClient) Connect() error {
	// Build stream URL
	// Format: wss://stream.binance.com:9443/stream?streams=btcusdt@kline_1m/ethusdt@kline_1m
	streams := make([]string, len(ws.symbols))
	for i, symbol := range ws.symbols {
		streams[i] = fmt.Sprintf("%s@kline_%s", strings.ToLower(symbol), ws.interval)
	}
	streamURL := fmt.Sprintf("%s/stream?streams=%s", ws.wsURL, strings.Join(streams, "/"))

	log.Printf("Connecting to Binance WebSocket: %s", streamURL)

	conn, _, err := websocket.DefaultDialer.Dial(streamURL, nil)
	if err != nil {
		return fmt.Errorf("websocket connection failed: %w", err)
	}

	ws.conn = conn
	log.Println("Successfully connected to Binance WebSocket")

	// Start reading messages
	go ws.readMessages()

	return nil
}

func (ws *WebSocketClient) readMessages() {
	defer func() {
		ws.conn.Close()
		close(ws.messageChan)
		close(ws.errorChan)
	}()

	for {
		select {
		case <-ws.done:
			return
		default:
			_, message, err := ws.conn.ReadMessage()
			if err != nil {
				ws.errorChan <- fmt.Errorf("read error: %w", err)
				return
			}

			// Parse message
			var streamData struct {
				Stream string                 `json:"stream"`
				Data   model.BinanceWSMessage `json:"data"`
			}

			if err := json.Unmarshal(message, &streamData); err != nil {
				log.Printf("Failed to parse message: %v", err)
				continue
			}

			// Convert to MarketPrice
			price, err := ws.parseWSMessage(&streamData.Data)
			if err != nil {
				log.Printf("Failed to parse kline data: %v", err)
				continue
			}

			// Send to channel
			select {
			case ws.messageChan <- price:
			default:
				log.Println("Message channel full, dropping message")
			}
		}
	}
}

func (ws *WebSocketClient) parseWSMessage(msg *model.BinanceWSMessage) (*model.MarketPrice, error) {
	open, _ := strconv.ParseFloat(msg.Kline.Open, 64)
	high, _ := strconv.ParseFloat(msg.Kline.High, 64)
	low, _ := strconv.ParseFloat(msg.Kline.Low, 64)
	close, _ := strconv.ParseFloat(msg.Kline.Close, 64)
	volume, _ := strconv.ParseFloat(msg.Kline.Volume, 64)

	return &model.MarketPrice{
		Symbol:   msg.Symbol,
		Time:     time.Unix(0, msg.Kline.StartTime*int64(time.Millisecond)),
		Interval: msg.Kline.Interval,
		Open:     open,
		High:     high,
		Low:      low,
		Close:    close,
		Volume:   volume,
	}, nil
}

// GetMessageChannel returns the channel for receiving market prices
func (ws *WebSocketClient) GetMessageChannel() <-chan *model.MarketPrice {
	return ws.messageChan
}

// GetErrorChannel returns the channel for receiving errors
func (ws *WebSocketClient) GetErrorChannel() <-chan error {
	return ws.errorChan
}

// Close closes the WebSocket connection
func (ws *WebSocketClient) Close() {
	close(ws.done)
	if ws.conn != nil {
		ws.conn.Close()
	}
}
