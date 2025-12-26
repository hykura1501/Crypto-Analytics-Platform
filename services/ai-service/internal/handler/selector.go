package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/crypto-platform/ai-service/internal/gemini"
	"google.golang.org/genai"
)

type SelectorMessage struct {
	SourceID   string `json:"source_id"`
	HTMLString string `json:"html_string"`
}

type SelectorHandler struct {
	DB           *sql.DB
	GeminiClient *gemini.GeminiClient
}

func NewSelectorHandler(db *sql.DB, geminiClient *gemini.GeminiClient) *SelectorHandler {
	return &SelectorHandler{
		DB:           db,
		GeminiClient: geminiClient,
	}
}
func (handler *SelectorHandler) Handle(ctx context.Context, msg SelectorMessage) (*gemini.SelectorResponse, error) {
	prompt := gemini.BuildSelectorAnalysisPrompt(msg.HTMLString)

	// Write prompt to file for debugging/improvement
	promptDir := "/tmp/ai-prompts/selector"
	os.MkdirAll(promptDir, 0755)
	timestamp := time.Now().Format("20060102-150405")
	promptFile := filepath.Join(promptDir, msg.SourceID+"_"+timestamp+".txt")
	if err := os.WriteFile(promptFile, []byte(prompt), 0644); err != nil {
		log.Printf("⚠️  Failed to write prompt to file: %v", err)
	} else {
		log.Printf("📝 Prompt written to: %s", promptFile)
	}

	content, err := handler.GeminiClient.GenerateContent(ctx, prompt, &genai.GenerateContentConfig{
		ResponseMIMEType:   "application/json",
		ResponseJsonSchema: gemini.SelectorResponseSchema,
	})

	if err != nil {
		log.Printf("❌ Error generating content: %v", err)
		return nil, err
	}

	var response gemini.SelectorResponse
	if err := json.Unmarshal([]byte(content), &response); err != nil {
		log.Printf("❌ Error decoding response: %v", err)
		return nil, err
	}

	log.Printf("CSS selector analysis result: %v", response)

	// Update database with CSS selectors
	query := `UPDATE sources SET summary_selector = $1, content_selector = $2, author_selector = $3, tags_selector = $4 WHERE source_id = $5`
	_, err = handler.DB.Exec(query, response.SummarySelector, response.ContentSelector, response.AuthorSelector, response.TagsSelector, msg.SourceID)
	if err != nil {
		log.Printf("Error updating CSS selectors for source %s: %v", msg.SourceID, err)
		return nil, err
	}

	log.Printf("✅ Updated CSS selectors for source %s", msg.SourceID)
	return &response, nil
}
