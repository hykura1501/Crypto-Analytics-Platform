package gemini

import (
	"fmt"
)

var SelectorResponseSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"summary_selector": map[string]any{
			"type":        "string",
			"description": "Single CSS selector for article summary/excerpt. Must be valid CSS selector (e.g., '.excerpt', 'p.lead'). Return empty string \"\" if not found. Must NOT be generic tags like 'body', 'div', 'p' without classes.",
			"pattern":     "^$|^[a-zA-Z][a-zA-Z0-9_-]*(\\.[a-zA-Z][a-zA-Z0-9_-]*|#[a-zA-Z][a-zA-Z0-9_-]*|\\[.*\\]|>|\\s|\\+)*$",
		},
		"content_selector": map[string]any{
			"type":        "string",
			"description": "REQUIRED: Single CSS selector for main article content container. Must be specific (e.g., 'article', '.article-body', '#content-main'). Must NOT be generic tags like 'body', 'html', 'div', 'p' without classes/ids.",
			"minLength":   1,
			"pattern":     "^[a-zA-Z][a-zA-Z0-9_-]*(\\.[a-zA-Z][a-zA-Z0-9_-]*|#[a-zA-Z][a-zA-Z0-9_-]*|\\[.*\\]|>|\\s|\\+)*$",
		},
		"author_selector": map[string]any{
			"type":        "string",
			"description": "Single CSS selector for author name (e.g., '.author', 'span.byline'). Return empty string \"\" if not found. Must NOT be generic tags.",
			"pattern":     "^$|^[a-zA-Z][a-zA-Z0-9_-]*(\\.[a-zA-Z][a-zA-Z0-9_-]*|#[a-zA-Z][a-zA-Z0-9_-]*|\\[.*\\]|>|\\s|\\+)*$",
		},
		"tags_selector": map[string]any{
			"type":        "string",
			"description": "Single CSS selector for tags/categories container (e.g., '.tags', 'ul.tag-list'). Return empty string \"\" if not found. Must NOT be generic tags.",
			"pattern":     "^$|^[a-zA-Z][a-zA-Z0-9_-]*(\\.[a-zA-Z][a-zA-Z0-9_-]*|#[a-zA-Z][a-zA-Z0-9_-]*|\\[.*\\]|>|\\s|\\+)*$",
		},
	},
	"required":             []string{"content_selector"},
	"additionalProperties": false,
}

type SelectorResponse struct {
	SummarySelector string `json:"summary_selector"`
	ContentSelector string `json:"content_selector"`
	AuthorSelector  string `json:"author_selector"`
	TagsSelector    string `json:"tags_selector"`
}

func BuildSelectorAnalysisPrompt(htmlString string) string {
	return fmt.Sprintf(`
You are an expert Web Scraper and HTML Analyst.

Your task: analyze the provided HTML article(s) and return CSS selectors that work CORRECTLY and CONSISTENTLY.

================================
INPUT HTML
================================
- The HTML may contain ONE or MULTIPLE articles
- Multiple articles are separated by:
  === ARTICLE N ===
- If multiple articles exist, selectors MUST work for ALL of them

%s

================================
CORE RULES (VERY IMPORTANT)
================================

### MULTI-ARTICLE RULE
- If multiple articles are provided:
  - Compare ALL article structures
  - Find selectors COMMON to ALL articles
  - If a selector works for only one article → IT IS INVALID

### SELECTOR RULES
- Must be valid CSS selector
- Must be ONE selector only (no commas)
- No leading/trailing spaces
- NOT allowed:
  - 'body', 'html', 'div', 'p', 'span' alone
  - multiple selectors
- Must be specific enough to select the correct element

================================
CONTENT SELECTOR (REQUIRED)
================================
Find the MAIN container that wraps ALL article paragraphs/text.

CRITICAL SIMPLICITY RULE:
- Use the SIMPLEST selector that fully contains the article content
- Prefer:
  - article
  - article.class-name
  - div.content / .article-body / .post-content
- DO NOT over-nest:
  - ❌ article.fck_detail > div.main-content
  - ✅ article.fck_detail
- Use nested selectors ONLY if:
  1. Parent contains unrelated content AND
  2. Child is the only real article content

NEVER return:
- 'body', 'html', 'div', 'p'
- empty string

================================
OPTIONAL SELECTORS
================================

### summary_selector
- Excerpt / lead paragraph before main content
- Classes like: excerpt, summary, lead, intro
- Return "" if not found or inconsistent

### author_selector
- Author / byline
- Classes like: author, byline, writer
- Return "" if not found or inconsistent

### tags_selector
- Tags / categories container
- Classes like: tags, categories, topics
- Return "" if not found or inconsistent

================================
VALIDATION CHECK
================================
Before answering, verify:
- content_selector exists in ALL articles
- selector is as SIMPLE as possible
- JSON format is EXACT

================================
OUTPUT FORMAT (JSON ONLY)
================================
Return ONLY this JSON structure:

{
  "summary_selector": "string or empty",
  "content_selector": "string (REQUIRED)",
  "author_selector": "string or empty",
  "tags_selector": "string or empty"
}

DO NOT add explanations.
DO NOT add markdown.
DO NOT add comments.
`, htmlString)
}
