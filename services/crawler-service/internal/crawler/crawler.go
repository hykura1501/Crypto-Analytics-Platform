package crawler

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"

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

type rssArticle struct {
	SourceID    string
	URL         string
	Title       string
	PublishedAt *time.Time
	Language    string
}

type sourceMeta struct {
	name       string
	rssURL     string
	language   string
	titleTag   string
	linkTag    string
	pubDateTag string
}

// CrawlOnce thực hiện crawl 1 vòng cho tất cả nguồn
func (s *Service) CrawlOnce(ctx context.Context) (int, error) {
	totalSaved := 0

	sources := []sourceMeta{
		{
			name:       "CoinDesk",
			rssURL:     s.cfg.RSS.CoinDeskURL,
			language:   "en",
			titleTag:   "title",
			linkTag:    "link",
			pubDateTag: "pubDate",
		},
		{
			name:       "CoinTelegraph",
			rssURL:     s.cfg.RSS.CoinTelegraphURL,
			language:   "en",
			titleTag:   "title",
			linkTag:    "guid",
			pubDateTag: "pubDate",
		},
		{
			name:       "VNExpress",
			rssURL:     s.cfg.RSS.VNExpressURL,
			language:   "vi",
			titleTag:   "title",
			linkTag:    "link",
			pubDateTag: "pubDate",
		},
		{
			name:       "VnEconomy",
			rssURL:     s.cfg.RSS.VnEconomyURL,
			language:   "vi",
			titleTag:   "title",
			linkTag:    "link",
			pubDateTag: "pubDate",
		},
	}

	for _, src := range sources {
		select {
		case <-ctx.Done():
			return totalSaved, ctx.Err()
		default:
		}

		log.Printf("📰 Crawling %s...", src.name)
		articles := s.fetchRSS(ctx, src, src.rssURL)
		log.Printf("Found %d articles from %s", len(articles), src.name)

		for idx, a := range articles {
			select {
			case <-ctx.Done():
				return totalSaved, ctx.Err()
			default:
			}

			// Log full URL for debugging, but truncate for display
			log.Printf("[%d/%d] Processing %s: %s (len=%d)", idx+1, len(articles), src.name, a.URL, len(a.URL))

			content, err := s.fetchArticleContent(a, src.name)
			if err != nil || len(strings.TrimSpace(content)) < 100 {
				log.Printf("Failed to extract content from %s: %v", a.URL, err)
				continue
			}

			id, inserted, err := s.insertArticle(ctx, a, content)
			if err != nil {
				log.Printf("Error inserting article %s: %v", a.URL, err)
				continue
			}
			if !inserted {
				continue
			}

			totalSaved++
			log.Printf("✅ Saved: %s...", truncate(a.Title, 60))

			if s.producer != nil {
				if err := s.producer.PublishNews(ctx, id, a.Title, content); err != nil {
					log.Printf("Failed to publish news to Kafka: %v", err)
				}
			}
		}
	}

	log.Printf("🎉 Crawling completed. Total saved: %d", totalSaved)
	return totalSaved, nil
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
		// Get pubDate using configured tag
		pubDateStr := strings.TrimSpace(e.ChildText(src.pubDateTag))

		var publishedAt *time.Time
		if pubDateStr != "" {
			if t, err := time.Parse(time.RFC1123Z, pubDateStr); err == nil {
				publishedAt = &t
			}
		}

		if link == "" || title == "" {
			return
		}

		result = append(result, rssArticle{
			SourceID:    src.name,
			URL:         link,
			Title:       title,
			PublishedAt: publishedAt,
			Language:    src.language,
		})
	})

	if err := c.Visit(rssURL); err != nil {
		log.Printf("Error fetching RSS for %s: %v", src.name, err)
	}

	return result
}

// fetchArticleContent dùng Colly để lấy nội dung bài viết
func (s *Service) fetchArticleContent(a rssArticle, sourceName string) (string, error) {
	var contentBuilder strings.Builder

	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) "+
			"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
		colly.Async(false),
	)

	switch sourceName {
	case "CoinTelegraph":
		c.OnHTML("article", func(e *colly.HTMLElement) {
			e.ForEach("p", func(_ int, el *colly.HTMLElement) {
				text := strings.TrimSpace(el.Text)
				if text != "" {
					contentBuilder.WriteString(text)
					contentBuilder.WriteString("\n")
				}
			})
		})
	case "CoinDesk":
		// CoinDesk uses p.font-body.text-charcoal-900 for article content
		c.OnHTML("p.font-body.text-charcoal-900", func(e *colly.HTMLElement) {
			text := strings.TrimSpace(e.Text)
			if text != "" && len(text) > 20 { // Filter out short text
				contentBuilder.WriteString(text)
				contentBuilder.WriteString("\n")
			}
		})
		// Also try div.article-content-wrapper as fallback
		c.OnHTML("div.article-content-wrapper p", func(e *colly.HTMLElement) {
			text := strings.TrimSpace(e.Text)
			if text != "" && len(text) > 20 && !strings.Contains(text, "Read full story") {
				// Check if not already added
				currentContent := contentBuilder.String()
				if len(text) > 0 && (len(currentContent) == 0 || !strings.Contains(currentContent, text[:min(50, len(text))])) {
					contentBuilder.WriteString(text)
					contentBuilder.WriteString("\n")
				}
			}
		})
		// Additional fallback: try any p tag with class containing "font-body"
		c.OnHTML("p[class*='font-body']", func(e *colly.HTMLElement) {
			text := strings.TrimSpace(e.Text)
			if text != "" && len(text) > 20 && !strings.Contains(text, "Read full story") {
				currentContent := contentBuilder.String()
				if len(currentContent) == 0 || !strings.Contains(currentContent, text[:min(50, len(text))]) {
					contentBuilder.WriteString(text)
					contentBuilder.WriteString("\n")
				}
			}
		})
	case "VNExpress":
		c.OnHTML("article.fck_detail", func(e *colly.HTMLElement) {
			e.ForEach("p", func(_ int, el *colly.HTMLElement) {
				text := strings.TrimSpace(el.Text)
				if text != "" {
					contentBuilder.WriteString(text)
					contentBuilder.WriteString("\n")
				}
			})
		})
	default:
		// fallback chung
		c.OnHTML("article", func(e *colly.HTMLElement) {
			e.ForEach("p", func(_ int, el *colly.HTMLElement) {
				text := strings.TrimSpace(el.Text)
				if text != "" {
					contentBuilder.WriteString(text)
					contentBuilder.WriteString("\n")
				}
			})
		})
	}

	if err := c.Visit(a.URL); err != nil {
		return "", fmt.Errorf("failed to visit URL: %w", err)
	}

	content := contentBuilder.String()
	if content == "" {
		return "", fmt.Errorf("no content extracted")
	}

	return content, nil
}

// insertArticle ghi vào Postgres, tránh trùng URL
func (s *Service) insertArticle(ctx context.Context, a rssArticle, content string) (int64, bool, error) {
	query := `
INSERT INTO articles (source_id, url, title, content_text, language, crawled_at, created_at)
VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
ON CONFLICT (url) DO NOTHING
RETURNING id;
`

	var id int64
	err := s.db.QueryRowContext(ctx, query,
		a.SourceID,
		a.URL,
		a.Title,
		content,
		a.Language,
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

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
