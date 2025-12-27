package crawler

import (
	"fmt"
	"log"

	"github.com/gocolly/colly/v2"
)

// AnalyzeSource fetches and analyzes RSS feed and article HTML structure
func (s *Service) AnalyzeSource(sourceID, rssURL string) error {
	log.Printf("🔍 Analyzing source: %s (%s)", sourceID, rssURL)

	// Fetch RSS XML
	c := setupBrowserLikeCollector()

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
		htmlCollector := setupBrowserLikeCollector()

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
