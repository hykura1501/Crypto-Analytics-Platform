package handler

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type ArticleHandler struct {
	DB *sql.DB
}

func NewArticleHandler(db *sql.DB) *ArticleHandler {
	return &ArticleHandler{
		DB: db,
	}
}

type Article struct {
	ID             int      `json:"id"`
	SourceID       string   `json:"source_id"`
	URL            string   `json:"url"`
	Title          string   `json:"title"`
	Author         *string  `json:"author,omitempty"`
	PublishedAt    string   `json:"published_at"`
	CrawledAt      string   `json:"crawled_at"`
	ContentText    string   `json:"content_text"`
	Language       string   `json:"language"`
	Tags           []string `json:"tags,omitempty"`
	Summary        *string  `json:"summary,omitempty"`
	SentimentScore *float64 `json:"sentiment_score,omitempty"`
	CreatedAt      string   `json:"created_at"`
}

func (h *ArticleHandler) ListArticles(c *gin.Context) {
	// Parse query parameters
	skip := 0
	limit := 50
	sourceID := ""
	language := ""

	if skipStr := c.Query("skip"); skipStr != "" {
		if val, err := strconv.Atoi(skipStr); err == nil && val >= 0 {
			skip = val
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil && val > 0 && val <= 100 {
			limit = val
		}
	}

	if sourceIDStr := c.Query("source_id"); sourceIDStr != "" {
		sourceID = sourceIDStr
	}

	if languageStr := c.Query("language"); languageStr != "" {
		language = languageStr
	}

	// Note: event_type filtering is not implemented as it's not in the articles table

	// Build query
	query := `
		SELECT id, source_id, url, title, author, published_at, 
		       crawled_at, content_text, language, tags, summary, 
		       sentiment_score, created_at
		FROM articles
		WHERE 1=1
	`
	args := []interface{}{}
	argIndex := 1

	if sourceID != "" {
		query += ` AND source_id = $` + strconv.Itoa(argIndex)
		args = append(args, sourceID)
		argIndex++
	}

	if language != "" {
		query += ` AND language = $` + strconv.Itoa(argIndex)
		args = append(args, language)
		argIndex++
	}

	// Note: event_type filtering would require additional logic
	// For now, we'll skip it as it's not in the articles table directly

	query += ` ORDER BY created_at DESC LIMIT $` + strconv.Itoa(argIndex) + ` OFFSET $` + strconv.Itoa(argIndex+1)
	args = append(args, limit, skip)

	rows, err := h.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to query articles: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	articles := []Article{}
	for rows.Next() {
		var article Article
		var author sql.NullString
		var publishedAt sql.NullString
		var tags sql.NullString
		var summary sql.NullString
		var sentimentScore sql.NullFloat64

		err := rows.Scan(
			&article.ID,
			&article.SourceID,
			&article.URL,
			&article.Title,
			&author,
			&publishedAt,
			&article.CrawledAt,
			&article.ContentText,
			&article.Language,
			&tags,
			&summary,
			&sentimentScore,
			&article.CreatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to scan article: " + err.Error(),
			})
			return
		}

		// Convert sql.NullString to *string
		if author.Valid {
			article.Author = &author.String
		}
		if publishedAt.Valid {
			article.PublishedAt = publishedAt.String
		} else {
			article.PublishedAt = ""
		}
		// Parse tags from comma-separated string to array
		if tags.Valid && tags.String != "" {
			tagList := strings.Split(tags.String, ",")
			article.Tags = make([]string, 0, len(tagList))
			for _, tag := range tagList {
				trimmed := strings.TrimSpace(tag)
				if trimmed != "" {
					article.Tags = append(article.Tags, trimmed)
				}
			}
		}
		if summary.Valid {
			article.Summary = &summary.String
		}
		if sentimentScore.Valid {
			article.SentimentScore = &sentimentScore.Float64
		}

		articles = append(articles, article)
	}

	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error iterating articles: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, articles)
}
