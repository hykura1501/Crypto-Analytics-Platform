package crawler

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/gocolly/colly/v2"
	"golang.org/x/net/html"
)

// setupBrowserLikeCollector configures a collector with browser-like headers and settings
func setupBrowserLikeCollector() *colly.Collector {
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
		colly.Async(false),
	)
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
