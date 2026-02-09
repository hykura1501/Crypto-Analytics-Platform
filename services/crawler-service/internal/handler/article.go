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
	cursorID := 0         // cursor ID from previous page
	cursorSortValue := "" // cursor sort value (published_at or created_at) from previous page
	limit := 50
	sourceID := ""
	language := ""
	sortBy := "published_at" // default sort by published_at
	sortOrder := "DESC"      // default descending (newest first)

	// Parse composite cursor: "sortValue|id" format
	if cursorStr := c.Query("cursor"); cursorStr != "" {
		parts := strings.Split(cursorStr, "|")
		if len(parts) == 2 {
			cursorSortValue = parts[0]
			if val, err := strconv.Atoi(parts[1]); err == nil && val > 0 {
				cursorID = val
			}
		} else {
			// Backward compatibility: try parsing as single ID
			if val, err := strconv.Atoi(cursorStr); err == nil && val > 0 {
				cursorID = val
			}
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

	// Parse sort parameters
	if sortByParam := c.Query("sort_by"); sortByParam != "" {
		// Allow sorting by published_at or created_at
		if sortByParam == "published_at" || sortByParam == "created_at" {
			sortBy = sortByParam
		}
	}

	if sortOrderParam := c.Query("sort_order"); sortOrderParam != "" {
		if sortOrderParam == "ASC" || sortOrderParam == "DESC" {
			sortOrder = sortOrderParam
		}
	}

	// Build query with cursor pagination
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

	// Cursor pagination: use composite cursor (sortBy value + id) for accurate pagination
	// This ensures we don't skip records when multiple articles have the same sortBy value
	if cursorID > 0 && cursorSortValue != "" {
		// Handle NULL values: if cursorSortValue is "NULL", treat it as NULL
		if cursorSortValue == "NULL" {
			if sortOrder == "DESC" {
				// For DESC with NULL: (sortBy IS NOT NULL) OR (sortBy IS NULL AND id < cursorID)
				query += ` AND ((` + sortBy + ` IS NOT NULL) OR (` + sortBy + ` IS NULL AND id < $` + strconv.Itoa(argIndex) + `))`
			} else {
				// For ASC with NULL: (sortBy IS NULL AND id > cursorID)
				query += ` AND (` + sortBy + ` IS NULL AND id > $` + strconv.Itoa(argIndex) + `)`
			}
			args = append(args, cursorID)
			argIndex++
		} else {
			// Normal case: cursorSortValue is not NULL
			if sortOrder == "DESC" {
				// For DESC: (sortBy < cursorSortValue) OR (sortBy IS NULL) OR (sortBy = cursorSortValue AND id < cursorID)
				query += ` AND ((` + sortBy + ` < $` + strconv.Itoa(argIndex) + `) OR (` + sortBy + ` IS NULL) OR (` + sortBy + ` = $` + strconv.Itoa(argIndex) + ` AND id < $` + strconv.Itoa(argIndex+1) + `))`
			} else {
				// For ASC: (sortBy > cursorSortValue) OR (sortBy = cursorSortValue AND id > cursorID)
				query += ` AND ((` + sortBy + ` > $` + strconv.Itoa(argIndex) + `) OR (` + sortBy + ` = $` + strconv.Itoa(argIndex) + ` AND id > $` + strconv.Itoa(argIndex+1) + `))`
			}
			args = append(args, cursorSortValue, cursorID)
			argIndex += 2
		}
	} else if cursorID > 0 {
		// Fallback: backward compatibility with old cursor format (ID only)
		// This is less accurate but works for basic cases
		if sortOrder == "DESC" {
			query += ` AND id < $` + strconv.Itoa(argIndex)
		} else {
			query += ` AND id > $` + strconv.Itoa(argIndex)
		}
		args = append(args, cursorID)
		argIndex++
	}

	// Sort by published_at or created_at
	query += ` ORDER BY ` + sortBy + ` ` + sortOrder + `, id ` + sortOrder
	query += ` LIMIT $` + strconv.Itoa(argIndex)
	args = append(args, limit+1) // Fetch one extra to check if there's more

	rows, err := h.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to query articles: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	articles := []Article{}
	hasMore := false
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

	// Check if there are more results (we fetched limit+1)
	if len(articles) > limit {
		hasMore = true
		articles = articles[:limit] // Remove the extra item
	}

	// Get next cursor (composite: sortBy value + id)
	// Format: "sortValue|id" or "NULL|id" for NULL values
	nextCursor := ""
	if len(articles) > 0 {
		lastArticle := articles[len(articles)-1]
		var sortValue string
		if sortBy == "published_at" {
			sortValue = lastArticle.PublishedAt
		} else {
			sortValue = lastArticle.CreatedAt
		}
		// Use "NULL" for empty/NULL values
		if sortValue == "" {
			sortValue = "NULL"
		}
		nextCursor = sortValue + "|" + strconv.Itoa(lastArticle.ID)
	}

	// Always return response with pagination metadata for better frontend support
	c.JSON(http.StatusOK, gin.H{
		"articles":    articles,
		"next_cursor": nextCursor,
		"has_more":    hasMore,
	})
}
