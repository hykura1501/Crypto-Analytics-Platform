package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"

	"github.com/crypto-platform/ai-service/internal/sentiment"
)

type NewsMessage struct {
	NewsID  int    `json:"news_id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type NewsHandler struct {
	DB       *sql.DB
	Analyzer *sentiment.Analyzer
}

func NewNewsHandler(db *sql.DB, analyzer *sentiment.Analyzer) *NewsHandler {
	return &NewsHandler{
		DB:       db,
		Analyzer: analyzer,
	}
}

func (h *NewsHandler) Handle(ctx context.Context, msgValue []byte) {
	var msg NewsMessage
	if err := json.Unmarshal(msgValue, &msg); err != nil {
		log.Printf("❌ Error decoding news message: %v", err)
		return
	}

	if msg.NewsID == 0 || msg.Content == "" {
		log.Printf("Invalid message: news_id=%d", msg.NewsID)
		return
	}

	log.Printf("📰 Processing news #%d: %s...", msg.NewsID, truncate(msg.Title, 50))

	// Analyze sentiment
	sentimentResult := h.Analyzer.Analyze(msg.Title + " " + msg.Content)
	log.Printf(
		"🎭 Sentiment: %s (score: %.3f)",
		sentimentResult.Label,
		sentimentResult.Compound,
	)

	// Update database with sentiment score
	query := `UPDATE articles SET sentiment_score = $1 WHERE id = $2`
	_, err := h.DB.Exec(query, sentimentResult.Compound, msg.NewsID)
	if err != nil {
		log.Printf("Error updating sentiment score for article #%d: %v", msg.NewsID, err)
		return
	}

	log.Printf("✅ Updated article #%d with sentiment score", msg.NewsID)

	// Get article details for causal analysis
	var article struct {
		ID          int
		PublishedAt sql.NullTime
	}

	query = `SELECT id, published_at FROM articles WHERE id = $1`
	err = h.DB.QueryRow(query, msg.NewsID).Scan(
		&article.ID,
		&article.PublishedAt,
	)

	if err == sql.ErrNoRows {
		log.Printf("Article #%d not found in database", msg.NewsID)
		return
	}
	if err != nil {
		log.Printf("Error fetching article #%d: %v", msg.NewsID, err)
		return
	}

	// Causal analysis skipped: no entities column in articles table.
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}
