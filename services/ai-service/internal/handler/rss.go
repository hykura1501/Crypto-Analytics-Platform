package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"

	"github.com/crypto-platform/ai-service/internal/gemini"
	"google.golang.org/genai"
)

type RssMessage struct {
	SourceID  string `json:"source_id"`
	XMLString string `json:"xml_string"`
}

type RssHandler struct {
	DB           *sql.DB
	GeminiClient *gemini.GeminiClient
}

func NewRssHandler(db *sql.DB, geminiClient *gemini.GeminiClient) *RssHandler {
	return &RssHandler{
		DB:           db,
		GeminiClient: geminiClient,
	}
}

func (handler *RssHandler) Handle(ctx context.Context, msgValue []byte) {
	var msg RssMessage
	if err := json.Unmarshal(msgValue, &msg); err != nil {
		log.Printf("❌ Error decoding RSS message: %v", err)
		return
	}

	prompt := gemini.BuildRssStructureAnalysisPrompt()

	content, err := handler.GeminiClient.GenerateContent(ctx, prompt, &genai.GenerateContentConfig{
		ResponseMIMEType:   "application/json",
		ResponseJsonSchema: gemini.RssResponseSchema,
	})

	if err != nil {
		log.Printf("❌ Error generating content: %v", err)
		return
	}

	var response gemini.RssResponse
	if err := json.Unmarshal([]byte(content), &response); err != nil {
		log.Printf("❌ Error decoding response: %v", err)
		return
	}

	log.Printf("RSS structure analysis result: %v", response)

	// Update database with RSS structure
	query := `UPDATE rss_sources SET title_tag = $1, link_tag = $2, pub_date_tag = $3 WHERE source_id = $4`
	_, err = handler.DB.Exec(query, response.TitleTag, response.LinkTag, response.PubDateTag, msg.SourceID)
	if err != nil {
		log.Printf("Error updating RSS structure for source %s: %v", msg.SourceID, err)
		return
	}

	log.Printf("✅ Updated RSS structure for source %s", msg.SourceID)
}
