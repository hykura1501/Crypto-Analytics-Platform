package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/gocolly/colly/v2"
	"github.com/joho/godotenv"

	"github.com/crypto-platform/crawler-service/config"
	"github.com/crypto-platform/crawler-service/internal/crawler"
	"github.com/crypto-platform/crawler-service/internal/db"
	ckafka "github.com/crypto-platform/crawler-service/internal/kafka"
)

type CrawlerSource struct {
	SourceID string `db:"source_id"`
	RssURL   string `db:"rss_url"`
}

func InitializeCrawlerSource(cfg *config.Config, database *sql.DB, producer *ckafka.Producer) {
	// Insert some soure_id to db if not exist
	crawlerSource := []CrawlerSource{
		{
			SourceID: "CoinDesk",
			RssURL:   cfg.RSS.CoinDeskURL,
		},
		{
			SourceID: "CoinTelegraph",
			RssURL:   cfg.RSS.CoinTelegraphURL,
		},
		{
			SourceID: "VNExpress",
			RssURL:   cfg.RSS.VNExpressURL,
		},
		{
			SourceID: "VnEconomy",
			RssURL:   cfg.RSS.VnEconomyURL,
		},
	}
	for _, source := range crawlerSource {
		result, err := database.Exec(`
			INSERT INTO sources (source_id, rss_url)
			VALUES ($1, $2)
			ON CONFLICT (source_id) DO NOTHING
		`, source.SourceID, source.RssURL)
		if err != nil {
			log.Fatalf("Failed to insert crawler source %s: %v", source.SourceID, err)
		}

		// Check if row was actually inserted (not a conflict)
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			log.Printf("⏭️  Source %s already exists, skipping Kafka analysis", source.SourceID)
			continue
		}

		log.Printf("Initialized crawler source: %s", source.SourceID)

		// Fetch RSS XML
		c := colly.NewCollector(
			colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) " +
				"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
		)

		var firstItemXML string
		var firstArticleLink string

		// Parse RSS and get only the first item
		c.OnXML("//item", func(e *colly.XMLElement) {
			if firstItemXML != "" {
				return
			}
			// Get the raw XML of this item element
			firstItemXML = e.Text

			link := e.ChildText("link")
			if link == "" {
				link = e.ChildText("guid")
			}
			if link != "" {
				firstArticleLink = link
			}
		})

		if err := c.Visit(source.RssURL); err != nil {
			log.Printf("Error fetching RSS for %s: %v", source.SourceID, err)
			continue
		}

		// Send only 1 item to Kafka for RSS structure analysis
		if producer != nil && firstItemXML != "" {
			message := map[string]string{
				"source_id":  source.SourceID,
				"xml_string": "<item>" + firstItemXML + "</item>",
			}
			if err := producer.SendMessage("news_analyze_rss_structure", message); err != nil {
				log.Printf("Failed to send RSS analysis message for %s: %v", source.SourceID, err)
			} else {
				log.Printf("✅ Sent 1 RSS item to Kafka for analysis: %s", source.SourceID)
			}
		}

		// Fetch HTML from first article and send for CSS selector analysis
		if producer != nil && firstArticleLink != "" {
			htmlCollector := colly.NewCollector(
				colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) " +
					"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
			)

			var htmlContent string
			htmlCollector.OnResponse(func(r *colly.Response) {
				htmlContent = string(r.Body)
				log.Printf("✅ Fetched HTML from: %s", r.Request.URL)
			})

			log.Printf("Fetching article HTML for CSS selector analysis: %s", firstArticleLink)

			if err := htmlCollector.Visit(firstArticleLink); err != nil {
				log.Printf("Error fetching article HTML from %s: %v", firstArticleLink, err)
			} else if htmlContent != "" {
				// Remove script and style tags content
				cleanedHTML := removeScriptAndStyleTags(htmlContent)

				message := map[string]string{
					"source_id":   source.SourceID,
					"html_string": cleanedHTML,
				}
				if err := producer.SendMessage("news_analyze_css_selector", message); err != nil {
					log.Printf("Failed to send CSS selector analysis message for %s: %v", source.SourceID, err)
				} else {
					log.Printf("✅ Sent cleaned HTML to Kafka for CSS selector analysis: %s", source.SourceID)
				}
			}
		}
	}
}

func removeScriptAndStyleTags(html string) string {
	// Remove script tags and their content
	scriptRegex := regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`)
	html = scriptRegex.ReplaceAllString(html, "")

	// Remove style tags and their content
	styleRegex := regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`)
	html = styleRegex.ReplaceAllString(html, "")

	// Remove svg tags and their content
	svgRegex := regexp.MustCompile(`(?is)<svg[^>]*>.*?</svg>`)
	html = svgRegex.ReplaceAllString(html, "")

	return html
}

func main() {
	// Load .env file if present
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found for crawler-service, using system environment variables")
	}

	cfg := config.Load()

	// Kết nối DB (sẽ tự động chạy migrations)
	database, err := db.Connect(&cfg.Database)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	defer database.Close()

	// Kafka producer
	var producer *ckafka.Producer
	if cfg.Kafka.Broker != "" && cfg.Kafka.Topic != "" {
		producer = ckafka.NewProducer(cfg.Kafka.Broker, cfg.Kafka.Topic)
		defer producer.Close()
	}

	crawlService := crawler.NewService(cfg, database, producer)

	// Initialize crawler sources and send RSS XML to Kafka
	go InitializeCrawlerSource(cfg, database, producer)

	// Run scheduler
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go startScheduler(ctx, crawlService, cfg.Crawler.IntervalMinutes)

	// HTTP server (health + manual trigger)
	r := gin.Default()

	// API v1 Group
	v1 := r.Group("/api/v1/news")
	{
		// Infrastructure Health Check
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":  "ok",
				"service": "crawler-service-go",
			})
		})

		v1.POST("/crawl/once", func(c *gin.Context) {
			count, err := crawlService.CrawlOnce(c.Request.Context())
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"saved": count,
			})
		})
	}

	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("Starting crawler-service (Go + Colly) on %s", addr)

	// graceful shutdown
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutting down crawler-service...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
}

func startScheduler(ctx context.Context, svc *crawler.Service, intervalMinutes int) {
	if intervalMinutes <= 0 {
		intervalMinutes = 10
	}

	ticker := time.NewTicker(time.Duration(intervalMinutes) * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		log.Println("Starting scheduled crawl...")
		if _, err := svc.CrawlOnce(ctx); err != nil {
			log.Printf("Scheduled crawl error: %v", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
