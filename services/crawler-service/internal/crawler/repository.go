package crawler

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
)

// getSourcesFromDB queries all active sources from database
func (s *Service) getSourcesFromDB(ctx context.Context) ([]sourceMeta, error) {
	query := `
		SELECT source_id, rss_url, 
		       COALESCE(title_tag, 'title') as title_tag,
		       COALESCE(link_tag, 'link') as link_tag,
		       COALESCE(pub_date_tag, 'pubDate') as pub_date_tag,
		       content_selector,
		       author_selector,
		       summary_selector,
		       tags_selector
		FROM sources
		WHERE rss_url IS NOT NULL AND rss_url != ''
		ORDER BY source_id
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query sources: %w", err)
	}
	defer rows.Close()

	sources := []sourceMeta{}
	for rows.Next() {
		var src sourceMeta
		var contentSelector sql.NullString
		var authorSelector sql.NullString
		var summarySelector sql.NullString
		var tagsSelector sql.NullString

		err := rows.Scan(
			&src.name,
			&src.rssURL,
			&src.titleTag,
			&src.linkTag,
			&src.pubDateTag,
			&contentSelector,
			&authorSelector,
			&summarySelector,
			&tagsSelector,
		)
		if err != nil {
			log.Printf("Error scanning source row: %v", err)
			continue
		}

		// Set default values if not provided
		if src.titleTag == "" {
			src.titleTag = "title"
		}
		if src.linkTag == "" {
			src.linkTag = "link"
		}
		if src.pubDateTag == "" {
			src.pubDateTag = "pubDate"
		}
		if contentSelector.Valid {
			src.contentSelector = contentSelector.String
		}
		if authorSelector.Valid {
			src.authorSelector = authorSelector.String
		}
		if summarySelector.Valid {
			src.summarySelector = summarySelector.String
		}
		if tagsSelector.Valid {
			src.tagsSelector = tagsSelector.String
		}

		// Detect language from source_id (Vietnamese sources)
		if strings.Contains(strings.ToLower(src.name), "vnexpress") ||
			strings.Contains(strings.ToLower(src.name), "vneconomy") ||
			strings.Contains(strings.ToLower(src.name), "vn") {
			src.language = "vi"
		} else {
			src.language = "en"
		}

		sources = append(sources, src)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sources: %w", err)
	}

	return sources, nil
}

// insertArticle writes to Postgres, avoiding duplicate URLs
func (s *Service) insertArticle(ctx context.Context, a rssArticle, articleData *articleContent) (int64, bool, error) {
	// Check if URL exists first to avoid burning IDs on conflict
	var existingID int64
	err := s.db.QueryRowContext(ctx, "SELECT id FROM articles WHERE url = $1", a.URL).Scan(&existingID)
	if err == nil {
		// Found existing article
		return existingID, false, nil
	}
	if err != sql.ErrNoRows {
		return 0, false, err
	}

	query := `
INSERT INTO articles (source_id, url, title, content_text, language, published_at, author, summary, tags, crawled_at, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
ON CONFLICT (url) DO NOTHING
RETURNING id;
`

	var id int64
	err = s.db.QueryRowContext(ctx, query,
		a.SourceID,
		a.URL,
		a.Title,
		articleData.content,
		a.Language,
		nullString(articleData.publishedAt),
		nullString(articleData.author),
		nullString(articleData.summary),
		nullString(articleData.tags),
	).Scan(&id)

	if err != nil {
		if err == sql.ErrNoRows {
			// Already exists (ON CONFLICT DO NOTHING)
			return 0, false, nil
		}
		return 0, false, err
	}

	return id, true, nil
}
