package crawler

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/gocolly/colly/v2"
	"golang.org/x/net/html"

	"github.com/crypto-platform/crawler-service/config"
	ckafka "github.com/crypto-platform/crawler-service/internal/kafka"
)

type Service struct {
	cfg      *config.Config
	db       *sql.DB
	producer *ckafka.Producer
}

func NewService(cfg *config.Config, db *sql.DB, producer *ckafka.Producer) *Service {
	return &Service{
		cfg:      cfg,
		db:       db,
		producer: producer,
	}
}

func (s *Service) AnalyzeSource(sourceID, rssURL string) error {
	log.Printf("🔍 Analyzing source: %s (%s)", sourceID, rssURL)

	// Fetch RSS XML
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) " +
			"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)

	var firstItemXML string
	var articleLinks []string
	const maxArticles = 2

	// Parse RSS and get first 2 items
	c.OnXML("//item", func(e *colly.XMLElement) {
		if len(articleLinks) >= maxArticles {
			return
		}

		// Get the raw XML of first item only (for RSS structure analysis)
		if firstItemXML == "" {
			firstItemXML = e.Text
		}

		link := e.ChildText("link")
		if link == "" {
			link = e.ChildText("guid")
		}
		if link != "" {
			articleLinks = append(articleLinks, link)
		}
	})

	if err := c.Visit(rssURL); err != nil {
		log.Printf("Error fetching RSS for %s: %v", sourceID, err)
		return err
	}

	// Send only 1 item to Kafka for RSS structure analysis
	if s.producer != nil && firstItemXML != "" {
		message := map[string]string{
			"source_id":  sourceID,
			"xml_string": "<item>" + firstItemXML + "</item>",
		}
		if err := s.producer.SendMessage("news_analyze_rss_structure", message); err != nil {
			log.Printf("Failed to send RSS analysis message for %s: %v", sourceID, err)
		} else {
			log.Printf("✅ Sent 1 RSS item to Kafka for analysis: %s", sourceID)
		}
	}

	// Fetch HTML from first 2 articles and send for CSS selector analysis
	if s.producer != nil && len(articleLinks) > 0 {
		htmlCollector := colly.NewCollector(
			colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) " +
				"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
		)

		var htmlContents []string
		var fetchedCount int
		const maxArticles = 2

		htmlCollector.OnResponse(func(r *colly.Response) {
			htmlContent := string(r.Body)
			if htmlContent != "" {
				// Remove script and style tags content
				cleanedHTML := removeScriptAndStyleTags(htmlContent)
				htmlContents = append(htmlContents, cleanedHTML)
				fetchedCount++
				log.Printf("✅ Fetched HTML from article %d/%d: %s", fetchedCount, maxArticles, r.Request.URL)
			}
		})

		// Fetch up to 2 articles
		articlesToFetch := len(articleLinks)
		if articlesToFetch > maxArticles {
			articlesToFetch = maxArticles
		}

		log.Printf("Fetching %d article(s) HTML for CSS selector analysis", articlesToFetch)
		for i := 0; i < articlesToFetch; i++ {
			if err := htmlCollector.Visit(articleLinks[i]); err != nil {
				log.Printf("Error fetching article HTML from %s: %v", articleLinks[i], err)
			}
		}

		if len(htmlContents) > 0 {
			// Combine HTMLs with separator
			combinedHTML := ""
			for i, html := range htmlContents {
				combinedHTML += fmt.Sprintf("=== ARTICLE %d ===\n%s\n\n", i+1, html)
			}

			message := map[string]string{
				"source_id":   sourceID,
				"html_string": combinedHTML,
			}
			if err := s.producer.SendMessage("news_analyze_css_selector", message); err != nil {
				log.Printf("Failed to send CSS selector analysis message for %s: %v", sourceID, err)
			} else {
				log.Printf("✅ Sent %d cleaned HTML article(s) to Kafka for CSS selector analysis: %s", len(htmlContents), sourceID)
			}
		}
	}
	return nil
}

// Format HTML: parse và serialize lại với indent
func formatHTML(htmlStr string) (string, error) {
	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	var f func(*html.Node, int)
	f = func(n *html.Node, level int) {
		indent := strings.Repeat("  ", level)
		switch n.Type {
		case html.DocumentNode:
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				f(c, level)
			}
		case html.ElementNode:
			buf.WriteString(fmt.Sprintf("%s<%s", indent, n.Data))
			for _, attr := range n.Attr {
				buf.WriteString(fmt.Sprintf(` %s="%s"`, attr.Key, attr.Val))
			}
			if n.FirstChild == nil {
				buf.WriteString("/>\n")
			} else {
				buf.WriteString(">\n")
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					f(c, level+1)
				}
				buf.WriteString(fmt.Sprintf("%s</%s>\n", indent, n.Data))
			}
		case html.TextNode:
			text := strings.TrimSpace(n.Data)
			if text != "" {
				buf.WriteString(fmt.Sprintf("%s%s\n", indent, text))
			}
		case html.CommentNode:
			buf.WriteString(fmt.Sprintf("%s<!--%s-->\n", indent, n.Data))
		}
	}
	f(doc, 0)

	return buf.String(), nil
}

func removeScriptAndStyleTags(htmlStr string) string {
	// Xóa comment
	commentRegex := regexp.MustCompile(`(?s)<!--.*?-->`)
	htmlStr = commentRegex.ReplaceAllString(htmlStr, "")

	// Xóa script, style, svg, noscript, iframe, object, embed
	blockTagsRegex := regexp.MustCompile(`(?is)<script[^>]*>.*?</script>|<style[^>]*>.*?</style>|<svg[^>]*>.*?</svg>|<noscript[^>]*>.*?</noscript>|<iframe[^>]*>.*?</iframe>|<object[^>]*>.*?</object>|<embed[^>]*>.*?</embed>`)
	htmlStr = blockTagsRegex.ReplaceAllString(htmlStr, "")

	// Xóa thẻ tự đóng không cần thiết
	selfClosingRegex := regexp.MustCompile(`(?is)<(link|input|path|rect|circle|polygon)[^>]*>`)
	htmlStr = selfClosingRegex.ReplaceAllString(htmlStr, "")

	// Xóa style và on* attributes
	styleAttrRegex := regexp.MustCompile(`(?i)\s+style="[^"]*"`)
	htmlStr = styleAttrRegex.ReplaceAllString(htmlStr, "")
	onEventRegex := regexp.MustCompile(`(?i)\s+on\w+="[^"]*"`)
	htmlStr = onEventRegex.ReplaceAllString(htmlStr, "")

	// Xóa khoảng trắng thừa
	spaceRegex := regexp.MustCompile(`\n\s*\n`)
	htmlStr = spaceRegex.ReplaceAllString(htmlStr, "\n")

	formattedHTML, err := formatHTML(htmlStr)
	if err != nil {
		return strings.TrimSpace(htmlStr)
	}
	return strings.TrimSpace(formattedHTML)
}

type rssArticle struct {
	SourceID    string
	URL         string
	Title       string
	PublishedAt string
	Author      string
	Language    string
}

type sourceMeta struct {
	name            string
	rssURL          string
	language        string
	titleTag        string
	linkTag         string
	pubDateTag      string
	contentSelector string
	authorSelector  string
	summarySelector string
	tagsSelector    string
}

type articleContent struct {
	content     string
	author      string
	summary     string
	tags        string
	publishedAt string
}

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

// CrawlOnce thực hiện crawl 1 vòng cho tất cả nguồn
func (s *Service) CrawlOnce(ctx context.Context) (int, error) {
	// Query sources from database instead of hardcoding
	sources, err := s.getSourcesFromDB(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get sources from DB: %w", err)
	}

	if len(sources) == 0 {
		log.Println("⚠️  No sources found in database")
		return 0, nil
	}

	log.Printf("📋 Found %d sources from database", len(sources))

	var wg sync.WaitGroup
	var savedCount int64 // Use atomic for thread-safe counter

	// Process each source concurrently
	for _, src := range sources {
		select {
		case <-ctx.Done():
			return int(atomic.LoadInt64(&savedCount)), ctx.Err()
		default:
		}

		wg.Add(1)
		go func(source sourceMeta) {
			defer wg.Done()

			log.Printf("📰 Crawling %s...", source.name)
			articles := s.fetchRSS(ctx, source, source.rssURL)
			log.Printf("Found %d articles from %s", len(articles), source.name)

			// Process each article sequentially (no concurrency for articles)
			for idx, article := range articles {
				select {
				case <-ctx.Done():
					return
				default:
				}

				// Log full URL for debugging, but truncate for display
				log.Printf("[%d/%d] Processing %s: %s (len=%d)", idx+1, len(articles), source.name, article.URL, len(article.URL))

				articleData, err := s.fetchArticleContent(article, source)
				if err != nil || len(strings.TrimSpace(articleData.content)) < 100 {
					log.Printf("Failed to extract content from %s: %v", article.URL, err)
					continue
				}

				// Update article with extracted author, summary, tags
				if articleData.author != "" {
					article.Author = articleData.author
				}

				id, inserted, err := s.insertArticle(ctx, article, articleData)
				if err != nil {
					log.Printf("Error inserting article %s: %v", article.URL, err)
					continue
				}
				if !inserted {
					continue
				}

				// Thread-safe increment
				atomic.AddInt64(&savedCount, 1)
				log.Printf("✅ Saved: %s...", article.Title)
				if articleData.author != "" {
					log.Printf("   Author: %s", articleData.author)
				}
				if articleData.summary != "" {
					log.Printf("   Summary: %s...", articleData.summary[:min(100, len(articleData.summary))])
				}
				if articleData.tags != "" {
					log.Printf("   Tags: %s", articleData.tags)
				}

				if s.producer != nil {
					if err := s.producer.PublishNews(ctx, id, article.Title, articleData.content); err != nil {
						log.Printf("Failed to publish news to Kafka: %v", err)
					}
				}
			}
		}(src)
	}

	// Wait for all sources to complete
	wg.Wait()

	finalCount := int(atomic.LoadInt64(&savedCount))
	log.Printf("🎉 Crawling completed. Total saved: %d", finalCount)
	return finalCount, nil
}

// fetchRSS dùng Colly để đọc RSS feed với config tags từ sourceMeta
func (s *Service) fetchRSS(ctx context.Context, src sourceMeta, rssURL string) []rssArticle {
	result := make([]rssArticle, 0, s.cfg.RSS.MaxArticlesPerRun)

	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) " +
			"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)

	c.OnXML("//item", func(e *colly.XMLElement) {
		if len(result) >= s.cfg.RSS.MaxArticlesPerRun {
			return
		}

		// Get title using configured tag
		title := strings.TrimSpace(e.ChildText(src.titleTag))
		// Clean CDATA if present (CoinDesk uses CDATA for title)
		if title != "" {
			title = strings.TrimPrefix(title, "<![CDATA[")
			title = strings.TrimSuffix(title, "]]>")
			title = strings.TrimSpace(title)
		}

		// Get link using configured tag
		link := strings.TrimSpace(e.ChildText(src.linkTag))
		// Clean CDATA if present (CoinTelegraph uses CDATA for link)
		if link != "" {
			link = strings.TrimPrefix(link, "<![CDATA[")
			link = strings.TrimSuffix(link, "]]>")
			link = strings.TrimSpace(link)
		}
		// Get pubDate using configured tag (store as string, no parsing)
		pubDateStr := strings.TrimSpace(e.ChildText(src.pubDateTag))

		if link == "" || title == "" {
			return
		}

		result = append(result, rssArticle{
			SourceID:    src.name,
			URL:         link,
			Title:       title,
			PublishedAt: pubDateStr, // Store as string directly
			Language:    src.language,
		})
	})

	if err := c.Visit(rssURL); err != nil {
		log.Printf("Error fetching RSS for %s: %v", src.name, err)
	}

	return result
}

// fetchArticleContent dùng Colly để lấy nội dung bài viết với selector động từ DB
// Trả về articleContent bao gồm content, author, summary, tags
func (s *Service) fetchArticleContent(a rssArticle, src sourceMeta) (*articleContent, error) {
	result := &articleContent{}
	var contentBuilder strings.Builder
	var authorBuilder strings.Builder
	var summaryBuilder strings.Builder
	var tagsBuilder strings.Builder

	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) "+
			"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
		colly.Async(false),
	)

	// Extract content using content_selector from database
	if src.contentSelector != "" {
		selector := strings.TrimSpace(src.contentSelector)

		// Check if selector targets paragraphs
		if strings.Contains(selector, " p") || strings.HasSuffix(selector, "p") {
			// Extract paragraphs from the selector
			c.OnHTML(selector, func(e *colly.HTMLElement) {
				e.ForEach("p", func(_ int, el *colly.HTMLElement) {
					text := strings.TrimSpace(el.Text)
					if text != "" && len(text) > 20 {
						// Avoid duplicates
						currentContent := contentBuilder.String()
						if len(currentContent) == 0 || !strings.Contains(currentContent, text[:min(50, len(text))]) {
							contentBuilder.WriteString(text)
							contentBuilder.WriteString("\n")
						}
					}
				})
			})
		} else {
			// Extract all text from the selector
			c.OnHTML(selector, func(e *colly.HTMLElement) {
				text := strings.TrimSpace(e.Text)
				if text != "" && len(text) > 20 {
					// Avoid duplicates
					currentContent := contentBuilder.String()
					if len(currentContent) == 0 || !strings.Contains(currentContent, text[:min(50, len(text))]) {
						contentBuilder.WriteString(text)
						contentBuilder.WriteString("\n")
					}
				}
			})
		}
	} else {
		// Fallback: use default "article" selector if no content_selector in DB
		log.Printf("⚠️  No content_selector for %s, using default 'article'", src.name)
		c.OnHTML("article", func(e *colly.HTMLElement) {
			e.ForEach("p", func(_ int, el *colly.HTMLElement) {
				text := strings.TrimSpace(el.Text)
				if text != "" && len(text) > 20 {
					currentContent := contentBuilder.String()
					if len(currentContent) == 0 || !strings.Contains(currentContent, text[:min(50, len(text))]) {
						contentBuilder.WriteString(text)
						contentBuilder.WriteString("\n")
					}
				}
			})
		})
	}

	// Extract author using author_selector from database
	if src.authorSelector != "" {
		c.OnHTML(src.authorSelector, func(e *colly.HTMLElement) {
			authorText := strings.TrimSpace(e.Text)
			if authorText != "" && authorBuilder.Len() == 0 {
				authorBuilder.WriteString(authorText)
			}
		})
	}

	// Extract summary using summary_selector from database
	if src.summarySelector != "" {
		c.OnHTML(src.summarySelector, func(e *colly.HTMLElement) {
			summaryText := strings.TrimSpace(e.Text)
			if summaryText != "" && summaryBuilder.Len() == 0 {
				summaryBuilder.WriteString(summaryText)
			}
		})
	}

	// Extract tags using tags_selector from database
	if src.tagsSelector != "" {
		c.OnHTML(src.tagsSelector, func(e *colly.HTMLElement) {
			// First try to extract from child elements (a, span, li) which are common for tags
			hasChildTags := false
			e.ForEach("a, span, li", func(_ int, el *colly.HTMLElement) {
				tagText := strings.TrimSpace(el.Text)
				if tagText != "" {
					currentTags := tagsBuilder.String()
					if !strings.Contains(currentTags, tagText) {
						if tagsBuilder.Len() > 0 {
							tagsBuilder.WriteString(", ")
						}
						tagsBuilder.WriteString(tagText)
						hasChildTags = true
					}
				}
			})

			// If no child tags found, use the element's text directly
			if !hasChildTags {
				tagText := strings.TrimSpace(e.Text)
				if tagText != "" {
					currentTags := tagsBuilder.String()
					if !strings.Contains(currentTags, tagText) {
						if tagsBuilder.Len() > 0 {
							tagsBuilder.WriteString(", ")
						}
						tagsBuilder.WriteString(tagText)
					}
				}
			}
		})
	}

	if err := c.Visit(a.URL); err != nil {
		return nil, fmt.Errorf("failed to visit URL: %w", err)
	}

	result.content = strings.TrimSpace(contentBuilder.String())
	result.author = strings.TrimSpace(authorBuilder.String())
	result.summary = strings.TrimSpace(summaryBuilder.String())
	result.tags = strings.TrimSpace(tagsBuilder.String())
	result.publishedAt = a.PublishedAt // Already a string
	if result.content == "" {
		return nil, fmt.Errorf("no content extracted")
	}

	return result, nil
}

// insertArticle ghi vào Postgres, tránh trùng URL
func (s *Service) insertArticle(ctx context.Context, a rssArticle, articleData *articleContent) (int64, bool, error) {
	query := `
INSERT INTO articles (source_id, url, title, content_text, language, published_at, author, summary, tags, crawled_at, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
ON CONFLICT (url) DO NOTHING
RETURNING id;
`

	var id int64
	err := s.db.QueryRowContext(ctx, query,
		a.SourceID,
		a.URL,
		a.Title,
		articleData.content,
		a.Language,
		nullString(articleData.publishedAt), // publishedAt is now string, use nullString helper
		nullString(articleData.author),
		nullString(articleData.summary),
		nullString(articleData.tags),
	).Scan(&id)

	if err != nil {
		if err == sql.ErrNoRows {
			// Đã tồn tại (ON CONFLICT DO NOTHING)
			return 0, false, nil
		}
		return 0, false, err
	}

	return id, true, nil
}

// nullString converts empty string to sql.NullString
func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
