package crawler

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"
	"sync/atomic"

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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
