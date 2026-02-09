package handler

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/crypto-platform/market-service/internal/model"
	"github.com/crypto-platform/market-service/internal/service"
	ws "github.com/crypto-platform/market-service/pkg/websocket"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var allowedOrigins = strings.Split(
	getEnvDefault("WS_ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:5173"),
	",",
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		for _, allowed := range allowedOrigins {
			if strings.TrimSpace(allowed) == origin {
				return true
			}
		}
		return false
	},
}

func getEnvDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

type MarketHandler struct {
	service *service.MarketService
	wsHub   *ws.Hub
}

func NewMarketHandler(service *service.MarketService, wsHub *ws.Hub) *MarketHandler {
	return &MarketHandler{
		service: service,
		wsHub:   wsHub,
	}
}

// GetHistory godoc
// @Summary Get historical market data
// @Description Fetch historical candlestick data from Binance
// @Tags market
// @Accept json
// @Produce json
// @Param symbol query string true "Trading pair symbol (e.g., BTCUSDT)"
// @Param interval query string true "Kline interval (1m, 5m, 15m, 1h, 4h, 1d, etc.)"
// @Param limit query int false "Number of records (default 100, max 1000)"
// @Param from query int64 false "Start time in milliseconds"
// @Param to query int64 false "End time in milliseconds"
// @Success 200 {object} model.SuccessResponse
// @Failure 400 {object} model.ErrorResponse
// @Router /history [get]
func (h *MarketHandler) GetHistory(c *gin.Context) {
	var req model.HistoryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "Invalid parameters",
			Message: err.Error(),
		})
		return
	}

	// Set default limit
	if req.Limit <= 0 {
		req.Limit = 100
	}
	if req.Limit > 1000 {
		req.Limit = 1000
	}

	// Fetch from Binance API
	prices, err := h.service.GetHistoricalData(req.Symbol, req.Interval, req.Limit, req.From, req.To)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "Failed to fetch historical data",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse{
		Success: true,
		Data:    prices,
	})
}

// GetFromDB godoc
// @Summary Get market data from database
// @Description Retrieve stored market data from database
// @Tags market
// @Produce json
// @Param symbol query string true "Trading pair symbol"
// @Param interval query string true "Kline interval"
// @Param limit query int false "Number of records (default 100)"
// @Param from query string false "Start time (RFC3339 format)"
// @Param to query string false "End time (RFC3339 format)"
// @Success 200 {object} model.SuccessResponse
// @Failure 400 {object} model.ErrorResponse
// @Router /data [get]
func (h *MarketHandler) GetFromDB(c *gin.Context) {
	symbol := c.Query("symbol")
	interval := c.Query("interval")
	limitStr := c.DefaultQuery("limit", "100")
	fromStr := c.Query("from")
	toStr := c.Query("to")

	if symbol == "" || interval == "" {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error: "Symbol and interval are required",
		})
		return
	}

	limit, _ := strconv.Atoi(limitStr)

	var from, to time.Time
	var err error

	if fromStr != "" {
		from, err = time.Parse(time.RFC3339, fromStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error:   "Invalid from time format",
				Message: "Use RFC3339 format (e.g., 2024-01-01T00:00:00Z)",
			})
			return
		}
	}

	if toStr != "" {
		to, err = time.Parse(time.RFC3339, toStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error:   "Invalid to time format",
				Message: "Use RFC3339 format",
			})
			return
		}
	}

	prices, err := h.service.GetFromDatabase(symbol, interval, from, to, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "Failed to retrieve data",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse{
		Success: true,
		Data:    prices,
	})
}

// WebSocketHandler handles WebSocket connections for real-time price streaming
// @Summary WebSocket endpoint for real-time prices
// @Description Connect to receive real-time market price updates
// @Tags websocket
// @Router /ws/prices [get]
func (h *MarketHandler) WebSocketHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "Failed to upgrade connection",
			Message: err.Error(),
		})
		return
	}

	// Register client
	client := h.wsHub.RegisterClient(conn)

	// Start pumps
	go client.WritePump()
	go client.ReadPump()
}
