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
	intervals   []string
	conn        *websocket.Conn
	messageChan chan *model.MarketPrice
	errorChan   chan error
	done        chan struct{}
}

func NewWebSocketClient(wsURL string, symbols []string, intervals []string) *WebSocketClient {
	return &WebSocketClient{
		wsURL:       wsURL,
		symbols:     symbols,
		intervals:   intervals,
		messageChan: make(chan *model.MarketPrice, 300),
		errorChan:   make(chan error, 10),
		done:        make(chan struct{}),
	}
}

// Connect establishes WebSocket connection to Binance
func (ws *WebSocketClient) Connect() error {
	// Build stream URL
	// Format: wss://stream.binance.com:9443/stream?streams=btcusdt@kline_1m/ethusdt@kline_1m
	streams := make([]string, len(ws.symbols)*len(ws.intervals))
	for i, symbol := range ws.symbols {
		for j, interval := range ws.intervals {
			streams[i*len(ws.intervals)+j] = fmt.Sprintf("%s@kline_%s", strings.ToLower(symbol), interval)
		}
	}
	streamURL := fmt.Sprintf("%s/stream?streams=%s", ws.wsURL, strings.Join(streams, "/"))

	log.Printf("Connecting to Binance WebSocket: %s", streamURL)

	conn, _, err := websocket.DefaultDialer.Dial(streamURL, nil)
	if err != nil {
		return fmt.Errorf("websocket connection failed: %w", err)
	}

	conn.SetPingHandler(func(appData string) error {
		// 1. Log để biết là Binance vừa Ping mình
		// log.Println("📡 Received PING from Binance, sending PONG...", appData)

		// 2. Quan trọng: Gia hạn thời gian sống cho kết nối
		// Nếu nhận được Ping nghĩa là mạng vẫn ngon -> Reset timeout thêm 60s
		if err := conn.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
			return err
		}

		// 3. Gửi trả lại PONG (Bắt buộc theo docs Binance)
		// WriteControl dùng để gửi các frame điều khiển (Ping/Pong/Close)
		err := conn.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(10*time.Second))
		if err == websocket.ErrCloseSent {
			return nil
		} else if e, ok := err.(interface{ Temporary() bool }); ok && e.Temporary() {
			return nil
		}
		return err
	})

	ws.conn = conn
	log.Println("Successfully connected to Binance WebSocket")

	// Start reading messages
	go ws.readMessages()

	return nil
}

func (ws *WebSocketClient) readMessages() {
	defer func() {
		ws.conn.Close()
		// Do not close channels here as they are reused on reconnection
		// close(ws.messageChan)
		// close(ws.errorChan)
	}()

	// Set initial read deadline
	ws.conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	for {
		select {
		case <-ws.done:
			return
		default:
			_, message, err := ws.conn.ReadMessage()
			if err != nil {
				if strings.Contains(err.Error(), "i/o timeout") {
					log.Println("⚠️ WebSocket connection timed out (no PING/Data from Binance)")
				}
				ws.errorChan <- fmt.Errorf("read error: %w", err)
				return
			}

			// Reset read deadline on successful message
			ws.conn.SetReadDeadline(time.Now().Add(60 * time.Second))

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
