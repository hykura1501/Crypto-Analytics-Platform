package crawler

import (
	"context"
	"database/sql"
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

// CrawlOnce thực hiện crawl 1 vòng cho tất cả nguồn
func (s *Service) CrawlOnce(ctx context.Context) (int, error) {
	totalSaved := 0

	sources := []struct {
		name     string
		rssURL   string
		language string
	}{
		{"CoinDesk", s.cfg.RSS.CoinDeskURL, "en"},
		{"CoinTelegraph", s.cfg.RSS.CoinTelegraphURL, "en"},
		{"VNExpress", s.cfg.RSS.VNExpressURL, "vi"},
		{"VnEconomy", s.cfg.RSS.VnEconomyURL, "vi"},
	}

	for _, src := range sources {
		select {
		case <-ctx.Done():
			return totalSaved, ctx.Err()
		default:
		}

		log.Printf("📰 Crawling %s...", src.name)
		articles := s.fetchRSS(ctx, src.name, src.rssURL, src.language)
		log.Printf("Found %d articles from %s", len(articles), src.name)

		for idx, a := range articles {
			select {
			case <-ctx.Done():
				return totalSaved, ctx.Err()
			default:
			}

			log.Printf("[%d/%d] Processing %s: %s", idx+1, len(articles), src.name, truncate(a.URL, 80))

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

// fetchRSS dùng Colly để đọc RSS feed
func (s *Service) fetchRSS(ctx context.Context, sourceName, rssURL, language string) []rssArticle {
	result := make([]rssArticle, 0, s.cfg.RSS.MaxArticlesPerRun)

	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) " +
			"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)

	c.OnXML("//item", func(e *colly.XMLElement) {
		if len(result) >= s.cfg.RSS.MaxArticlesPerRun {
			return
		}

		link := strings.TrimSpace(e.ChildText("link"))
		title := strings.TrimSpace(e.ChildText("title"))
		pubDateStr := strings.TrimSpace(e.ChildText("pubDate"))

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
			SourceID:    sourceName,
			URL:         link,
			Title:       title,
			PublishedAt: publishedAt,
			Language:    language,
		})
	})

	if err := c.Visit(rssURL); err != nil {
		log.Printf("Error fetching RSS for %s: %v", sourceName, err)
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
		c.OnHTML("article", func(e *colly.HTMLElement) {
			e.ForEach("p", func(_ int, el *colly.HTMLElement) {
				text := strings.TrimSpace(el.Text)
				if text != "" {
					contentBuilder.WriteString(text)
					contentBuilder.WriteString("\n")
				}
			})
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
		return "", err
	}

	return contentBuilder.String(), nil
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
