package crawler

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"golang.org/x/net/html"
)

// setupBrowserLikeCollector configures a collector with browser-like headers and settings
func setupBrowserLikeCollector() *colly.Collector {
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
		colly.Async(false),
	)

	// Add realistic browser headers
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
		r.Headers.Set("Accept-Language", "en-US,en;q=0.9")
		r.Headers.Set("Connection", "keep-alive")
		r.Headers.Set("Upgrade-Insecure-Requests", "1")
	})

	// Handle errors
	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Error fetching %s: %v (Status: %d)", r.Request.URL, err, r.StatusCode)
	})

	return c
}

// formatHTML parses and serializes HTML with indentation
func formatHTML(htmlStr string) (string, error) {
	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer

	// Helper function to check if a node has only text content (no element children)
	hasOnlyText := func(n *html.Node) bool {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode {
				return false
			}
		}
		return true
	}

	// Helper function to get all text content from children
	getTextContent := func(n *html.Node) string {
		var text strings.Builder
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.TextNode {
				text.WriteString(c.Data)
			}
		}
		return strings.TrimSpace(text.String())
	}

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
				// Self-closing tag
				buf.WriteString("/>\n")
			} else if hasOnlyText(n) {
				// Element with only text content - keep it inline
				text := getTextContent(n)
				if text != "" {
					buf.WriteString(fmt.Sprintf(">%s</%s>\n", text, n.Data))
				} else {
					buf.WriteString(fmt.Sprintf("></%s>\n", n.Data))
				}
			} else {
				// Element with child elements
				buf.WriteString(">\n")
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					if c.Type != html.TextNode || strings.TrimSpace(c.Data) != "" {
						f(c, level+1)
					}
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
	// Remove comments
	commentRegex := regexp.MustCompile(`(?s)<!--.*?-->`)
	htmlStr = commentRegex.ReplaceAllString(htmlStr, "")

	// Remove script, style, svg, noscript, iframe, object, embed tags
	blockTagsRegex := regexp.MustCompile(`(?is)<script[^>]*>.*?</script>|<style[^>]*>.*?</style>|<svg[^>]*>.*?</svg>|<noscript[^>]*>.*?</noscript>|<iframe[^>]*>.*?</iframe>|<object[^>]*>.*?</object>|<embed[^>]*>.*?</embed>`)
	htmlStr = blockTagsRegex.ReplaceAllString(htmlStr, "")

	// Remove unnecessary self-closing tags
	selfClosingRegex := regexp.MustCompile(`(?is)<(link|input|path|rect|circle|polygon)[^>]*>`)
	htmlStr = selfClosingRegex.ReplaceAllString(htmlStr, "")

	// Remove style and on* attributes
	styleAttrRegex := regexp.MustCompile(`(?i)\s+style="[^"]*"`)
	htmlStr = styleAttrRegex.ReplaceAllString(htmlStr, "")
	onEventRegex := regexp.MustCompile(`(?i)\s+on\w+="[^"]*"`)
	htmlStr = onEventRegex.ReplaceAllString(htmlStr, "")

	// Remove extra whitespace
	spaceRegex := regexp.MustCompile(`\n\s*\n`)
	htmlStr = spaceRegex.ReplaceAllString(htmlStr, "\n")

	formattedHTML, err := formatHTML(htmlStr)
	if err != nil {
		return strings.TrimSpace(htmlStr)
	}
	return strings.TrimSpace(formattedHTML)
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

// fetchHTMLWithBrowser uses colly to fetch HTML content (imitating browser behavior)
func fetchHTMLWithBrowser(url string) (string, error) {
	var htmlContent string
	var errFetch error

	c := setupBrowserLikeCollector()

	// Set a reasonable timeout
	c.SetRequestTimeout(30 * time.Second)

	// Capture the response body
	c.OnResponse(func(r *colly.Response) {
		htmlContent = string(r.Body)
	})

	// Capture errors
	c.OnError(func(r *colly.Response, err error) {
		errFetch = err
	})

	err := c.Visit(url)
	if err != nil {
		return "", fmt.Errorf("failed to visit URL: %w", err)
	}

	if errFetch != nil {
		return "", fmt.Errorf("failed to fetch content: %w", errFetch)
	}

	if htmlContent == "" {
		return "", fmt.Errorf("empty response from %s", url)
	}

	log.Printf("✅ Fetched HTML from %s (%d bytes)", url, len(htmlContent))
	return htmlContent, nil
}
