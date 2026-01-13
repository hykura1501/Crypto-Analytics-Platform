package crawler

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
)

// fetchRSS uses Colly to read RSS feed with config tags from sourceMeta
func (s *Service) fetchRSS(ctx context.Context, src sourceMeta, rssURL string) []rssArticle {
	result := make([]rssArticle, 0, s.cfg.RSS.MaxArticlesPerRun)

	c := setupBrowserLikeCollector()

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

// fetchArticleContent fetches article with browser rendering then extracts content using selectors
// Returns articleContent including content, author, summary, tags
func (s *Service) fetchArticleContent(a rssArticle, src sourceMeta) (*articleContent, error) {
	result := &articleContent{}

	// Fetch HTML with browser rendering to execute JavaScript
	log.Printf("🌐 Fetching %s with browser rendering...", a.URL)
	htmlContent, err := fetchHTMLWithBrowser(a.URL)
	if err != nil {
		log.Printf("❌ Failed to fetch HTML with browser for %s: %v", a.URL, err)
		return nil, fmt.Errorf("failed to fetch HTML: %w", err)
	}

	// Parse HTML with goquery
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var contentBuilder strings.Builder
	var authorBuilder strings.Builder
	var summaryBuilder strings.Builder
	var tagsBuilder strings.Builder

	// Extract content using content_selector from database
	if src.contentSelector != "" {
		selector := strings.TrimSpace(src.contentSelector)

		// Check if selector targets paragraphs
		if strings.Contains(selector, " p") || strings.HasSuffix(selector, "p") {
			// Extract paragraphs from the selector
			doc.Find(selector).Each(func(i int, s *goquery.Selection) {
				s.Find("p").Each(func(j int, p *goquery.Selection) {
					text := strings.TrimSpace(p.Text())
					if text != "" && len(text) > 20 {
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
			doc.Find(selector).Each(func(i int, s *goquery.Selection) {
				text := strings.TrimSpace(s.Text())
				if text != "" && len(text) > 20 {
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
		doc.Find("article").Each(func(i int, s *goquery.Selection) {
			s.Find("p").Each(func(j int, p *goquery.Selection) {
				text := strings.TrimSpace(p.Text())
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
		doc.Find(src.authorSelector).Each(func(i int, s *goquery.Selection) {
			authorText := strings.TrimSpace(s.Text())
			if authorText != "" && authorBuilder.Len() == 0 {
				authorBuilder.WriteString(authorText)
			}
		})
	}

	// Extract summary using summary_selector from database
	if src.summarySelector != "" {
		doc.Find(src.summarySelector).Each(func(i int, s *goquery.Selection) {
			summaryText := strings.TrimSpace(s.Text())
			if summaryText != "" && summaryBuilder.Len() == 0 {
				summaryBuilder.WriteString(summaryText)
			}
		})
	}

	// Extract tags using tags_selector from database
	if src.tagsSelector != "" {
		doc.Find(src.tagsSelector).Each(func(i int, s *goquery.Selection) {
			// First try to extract from child elements (a, span, li) which are common for tags
			hasChildTags := false
			s.Find("a, span, li").Each(func(j int, tag *goquery.Selection) {
				tagText := strings.TrimSpace(tag.Text())
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
				tagText := strings.TrimSpace(s.Text())
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
