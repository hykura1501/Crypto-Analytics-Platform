package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/hykura1501/crypto-analytics-backend/internal/middleware"
	"github.com/hykura1501/crypto-analytics-backend/internal/models"
	"github.com/hykura1501/crypto-analytics-backend/internal/services"
	ws "github.com/hykura1501/crypto-analytics-backend/internal/websocket"
	"github.com/redis/go-redis/v9"
)

type Handler struct {
	DB             *sql.DB
	RedisClient    *redis.Client
	Hub            *ws.Hub
	BinanceService *services.BinanceService
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

func NewHandler(db *sql.DB, redis *redis.Client, hub *ws.Hub) *Handler {
	return &Handler{
		DB:             db,
		RedisClient:    redis,
		Hub:            hub,
		BinanceService: services.NewBinanceService(),
	}
}

// GetTradingPairs returns all available trading pairs
func (h *Handler) GetTradingPairs(c *gin.Context) {
	rows, err := h.DB.Query("SELECT id, symbol, base_asset, quote_asset, status, created_at, updated_at FROM trading_pairs WHERE status = 'active' ORDER BY symbol")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var pairs []models.TradingPair
	for rows.Next() {
		var pair models.TradingPair
		err := rows.Scan(&pair.ID, &pair.Symbol, &pair.BaseAsset, &pair.QuoteAsset, &pair.Status, &pair.CreatedAt, &pair.UpdatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		pairs = append(pairs, pair)
	}

	c.JSON(http.StatusOK, gin.H{"pairs": pairs})
}

// GetPairInfo returns information about a specific trading pair
func (h *Handler) GetPairInfo(c *gin.Context) {
	pair := c.Param("pair")
	var pairInfo models.TradingPair

	err := h.DB.QueryRow(
		"SELECT id, symbol, base_asset, quote_asset, status, created_at, updated_at FROM trading_pairs WHERE symbol = $1",
		pair,
	).Scan(&pairInfo.ID, &pairInfo.Symbol, &pairInfo.BaseAsset, &pairInfo.QuoteAsset, &pairInfo.Status, &pairInfo.CreatedAt, &pairInfo.UpdatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trading pair not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pairInfo)
}

// GetCurrentPrice returns current price from Binance API
func (h *Handler) GetCurrentPrice(c *gin.Context) {
	pair := c.Param("pair")

	// Try to get from cache first
	cacheKey := "price:" + pair
	cached, err := h.RedisClient.Get(c.Request.Context(), cacheKey).Result()
	if err == nil {
		var priceData map[string]interface{}
		if json.Unmarshal([]byte(cached), &priceData) == nil {
			c.JSON(http.StatusOK, priceData)
			return
		}
	}

	// Fetch from Binance API
	priceResp, err := h.BinanceService.GetCurrentPrice(pair)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch price from Binance: " + err.Error()})
		return
	}

	priceData := map[string]interface{}{
		"pair":      pair,
		"price":     priceResp.Price,
		"timestamp": time.Now().Unix(),
	}

	// Cache for 1 second
	if priceJSON, err := json.Marshal(priceData); err == nil {
		h.RedisClient.Set(c.Request.Context(), cacheKey, priceJSON, time.Second)
	}

	c.JSON(http.StatusOK, priceData)
}

// GetPriceHistory returns historical price data
func (h *Handler) GetPriceHistory(c *gin.Context) {
	pair := c.Param("pair")
	interval := c.DefaultQuery("interval", "1h")
	limit := c.DefaultQuery("limit", "100")

	limitInt, _ := strconv.Atoi(limit)

	// Get pair ID
	var pairID int
	err := h.DB.QueryRow("SELECT id FROM trading_pairs WHERE symbol = $1", pair).Scan(&pairID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trading pair not found"})
		return
	}

	rows, err := h.DB.Query(
		"SELECT id, pair_id, price, volume, high, low, open_price, close_price, timestamp, interval, created_at FROM price_history WHERE pair_id = $1 AND interval = $2 ORDER BY timestamp DESC LIMIT $3",
		pairID, interval, limitInt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var history []models.PriceHistory
	for rows.Next() {
		var h models.PriceHistory
		err := rows.Scan(&h.ID, &h.PairID, &h.Price, &h.Volume, &h.High, &h.Low, &h.OpenPrice, &h.ClosePrice, &h.Timestamp, &h.Interval, &h.CreatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		history = append(history, h)
	}

	c.JSON(http.StatusOK, gin.H{"history": history})
}

// GetKlines returns kline/candlestick data from Binance
func (h *Handler) GetKlines(c *gin.Context) {
	pair := c.Param("pair")
	interval := c.DefaultQuery("interval", "1h")
	limitStr := c.DefaultQuery("limit", "500")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 500
	}

	klines, err := h.BinanceService.GetKlines(pair, interval, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch klines from Binance: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pair":     pair,
		"interval": interval,
		"limit":    limit,
		"klines":   klines,
	})
}

// GetNews returns news articles
func (h *Handler) GetNews(c *gin.Context) {
	limit := c.DefaultQuery("limit", "20")
	limitInt, _ := strconv.Atoi(limit)

	rows, err := h.DB.Query(
		"SELECT id, source_id, title, content, url, author, published_at, crawled_at, sentiment_score, sentiment_label, keywords, created_at FROM news ORDER BY published_at DESC LIMIT $1",
		limitInt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var newsList []models.News
	for rows.Next() {
		var n models.News
		err := rows.Scan(&n.ID, &n.SourceID, &n.Title, &n.Content, &n.URL, &n.Author, &n.PublishedAt, &n.CrawledAt, &n.SentimentScore, &n.SentimentLabel, &n.Keywords, &n.CreatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		newsList = append(newsList, n)
	}

	c.JSON(http.StatusOK, gin.H{"news": newsList})
}

// GetNewsDetail returns a specific news article
func (h *Handler) GetNewsDetail(c *gin.Context) {
	id := c.Param("id")
	var news models.News

	err := h.DB.QueryRow(
		"SELECT id, source_id, title, content, url, author, published_at, crawled_at, sentiment_score, sentiment_label, keywords, created_at FROM news WHERE id = $1",
		id,
	).Scan(&news.ID, &news.SourceID, &news.Title, &news.Content, &news.URL, &news.Author, &news.PublishedAt, &news.CrawledAt, &news.SentimentScore, &news.SentimentLabel, &news.Keywords, &news.CreatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "News not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, news)
}

// GetNewsSources returns all news sources
func (h *Handler) GetNewsSources(c *gin.Context) {
	rows, err := h.DB.Query("SELECT id, name, url, selector_type, title_selector, content_selector, date_selector, link_selector, status, last_crawled_at, created_at, updated_at FROM news_sources")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var sources []models.NewsSource
	for rows.Next() {
		var s models.NewsSource
		err := rows.Scan(&s.ID, &s.Name, &s.URL, &s.SelectorType, &s.TitleSelector, &s.ContentSelector, &s.DateSelector, &s.LinkSelector, &s.Status, &s.LastCrawledAt, &s.CreatedAt, &s.UpdatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		sources = append(sources, s)
	}

	c.JSON(http.StatusOK, gin.H{"sources": sources})
}

// GetAnalysis returns AI analysis for a trading pair
func (h *Handler) GetAnalysis(c *gin.Context) {
	pair := c.Param("pair")

	var pairID int
	err := h.DB.QueryRow("SELECT id FROM trading_pairs WHERE symbol = $1", pair).Scan(&pairID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trading pair not found"})
		return
	}

	rows, err := h.DB.Query(
		"SELECT id, pair_id, analysis_type, prediction, confidence_score, reasoning, time_horizon, created_at, valid_until FROM ai_analysis WHERE pair_id = $1 ORDER BY created_at DESC LIMIT 1",
		pairID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var analysis models.AIAnalysis
	if rows.Next() {
		err := rows.Scan(&analysis.ID, &analysis.PairID, &analysis.AnalysisType, &analysis.Prediction, &analysis.ConfidenceScore, &analysis.Reasoning, &analysis.TimeHorizon, &analysis.CreatedAt, &analysis.ValidUntil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, analysis)
		return
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "No analysis found"})
}

// PredictTrend creates a new AI prediction
func (h *Handler) PredictTrend(c *gin.Context) {
	var req struct {
		Pair        string `json:"pair" binding:"required"`
		TimeHorizon string `json:"time_horizon"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.TimeHorizon == "" {
		req.TimeHorizon = "24h"
	}

	// This should call the AI service
	c.JSON(http.StatusOK, gin.H{
		"message": "Prediction request received, will be processed by AI service",
		"pair":    req.Pair,
	})
}

// Register creates a new user account
func (h *Handler) Register(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=8"`
		Username string `json:"username" binding:"required,min=3"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hash password (in production, use bcrypt)
	// For demo purposes, we'll store plain text (NOT SECURE!)
	// TODO: Implement proper password hashing

	var userID int
	err := h.DB.QueryRow(`
		INSERT INTO users (email, username, password_hash, role)
		VALUES ($1, $2, $3, 'user')
		RETURNING id
	`, req.Email, req.Username, req.Password).Scan(&userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user_id": userID,
	})
}

// Login authenticates a user
func (h *Handler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var userID int
	var username, role, passwordHash string

	err := h.DB.QueryRow(`
		SELECT id, username, password_hash, role
		FROM users
		WHERE email = $1
	`, req.Email).Scan(&userID, &username, &passwordHash, &role)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Verify password (in production, use bcrypt.CompareHashAndPassword)
	if passwordHash != req.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate tokens
	accessToken, refreshToken, err := middleware.GenerateTokenPair(userID, req.Email, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user": gin.H{
			"id":       userID,
			"email":    req.Email,
			"username": username,
			"role":     role,
		},
	})
}

// RefreshToken generates a new access token from refresh token
func (h *Handler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims, err := middleware.ValidateToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	// Generate new access token
	accessToken, err := middleware.GenerateToken(claims.UserID, claims.Email, claims.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": accessToken,
	})
}

func (h *Handler) GetProfile(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

func (h *Handler) UpdateProfile(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

func (h *Handler) GetWatchlist(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

func (h *Handler) AddToWatchlist(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

func (h *Handler) RemoveFromWatchlist(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

// HandleWebSocket handles WebSocket connections for realtime price updates
func (h *Handler) HandleWebSocket(c *gin.Context) {
	pair := c.Param("pair")

	// Upgrade connection to WebSocket
	wsConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to upgrade connection"})
		return
	}

	conn := &ws.WSConn{Conn: wsConn}
	client := &ws.Client{
		Hub:  h.Hub,
		Pair: pair,
		Conn: conn,
		Send: make(chan []byte, 256),
	}

	client.Hub.Register <- client

	// Allow collection of memory referenced by the caller by doing all work in new goroutines
	go client.WritePump()
	go client.ReadPump()
}
