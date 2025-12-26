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
	return fmt.Sprintf(`You are an expert Web Scraper and HTML Analyst. Analyze the provided HTML articles and extract COMMON CSS selectors that work for ALL articles following STRICT rules.

**INPUT HTML:**
The HTML may contain 1-2 articles separated by "=== ARTICLE N ===" markers. Analyze ALL articles to find COMMON selectors.

%s

**CRITICAL: MULTI-ARTICLE ANALYSIS:**

1. **If multiple articles are provided:**
   - Compare the HTML structure of ALL articles
   - Find CSS selectors that are CONSISTENT across ALL articles
   - The selector must work for EVERY article, not just one
   - Look for common patterns, classes, or IDs that appear in all articles
   - If a selector only works for one article, it's WRONG - find the common pattern

2. **If only one article is provided:**
   - Analyze that single article
   - Still follow all format requirements below

**STRICT FORMAT REQUIREMENTS:**

1. **SELECTOR FORMAT:**
   - MUST be valid CSS selectors (e.g., '.class-name', '#id-name', 'tag.class', 'parent > child')
   - MUST be a SINGLE selector string (no commas, no multiple selectors)
   - MUST NOT contain spaces at start/end
   - MUST NOT be generic tags alone ('body', 'html', 'div', 'p', 'span' without classes/ids)
   - MUST be specific enough to target the exact element

2. **CONTENT SELECTOR (REQUIRED):**
   - Find the MAIN container that wraps ALL article paragraphs/text
   - Look for: article tags, divs with classes like 'content', 'article-body', 'post-content', 'entry-content'
   - Priority: article > [class*="content"] > [class*="article"] > [id*="content"]
   - Examples: 'article', '.article-body', '#main-content', '.post-content', 'div.entry-content'
   - NEVER return: 'body', 'html', 'div', 'p', or multiple selectors

3. **SUMMARY SELECTOR (OPTIONAL):**
   - Find the excerpt/lead/summary paragraph (usually before main content)
   - Look for: elements with classes like 'excerpt', 'summary', 'lead', 'intro'
   - Return empty string "" if not found
   - Examples: '.article-excerpt', '.summary', 'p.lead', '.intro-text'

4. **AUTHOR SELECTOR (OPTIONAL):**
   - Find the author name element
   - Look for: elements with classes like 'author', 'byline', 'writer', 'posted-by'
   - Return empty string "" if not found
   - Examples: '.author-name', '.byline', 'span.author', '.article-author'

5. **TAGS SELECTOR (OPTIONAL):**
   - Find the container/list that holds tags/categories
   - Look for: elements with classes like 'tags', 'categories', 'topics', 'labels'
   - Return empty string "" if not found
   - Examples: '.tags', '.categories', 'ul.tag-list', '.article-tags'

**VALIDATION RULES:**

✓ CORRECT: 'article', '.article-content', '#post-body', 'div.main-content'
✗ WRONG: 'body', 'html', 'div', 'p', 'article, .content', ' article ', ''

**STEP-BY-STEP PROCESS:**

1. **If multiple articles:**
   - Compare HTML structure of ALL articles
   - Identify COMMON patterns (same classes, same IDs, same structure)
   - Find selectors that match ALL articles, not just one
   - Verify the selector works consistently across all articles

2. **If single article:**
   - Scan the HTML structure
   - Identify the main article content container

3. **For all cases:**
   - Find the MAIN container that wraps ALL article paragraphs/text (must be consistent across articles if multiple)
   - Find summary/excerpt if present (must be consistent if multiple articles)
   - Locate author information if present (must be consistent if multiple articles)
   - Find tags/categories container if present (must be consistent if multiple articles)
   - Return selectors in the exact JSON format specified

**VALIDATION FOR MULTI-ARTICLE:**
- If you see "=== ARTICLE 1 ===" and "=== ARTICLE 2 ===", you MUST find selectors that work for BOTH
- Test your selectors mentally: would they select the same element type in both articles?
- If unsure, prefer more specific selectors that are likely to be consistent

**OUTPUT FORMAT:**
Return ONLY valid JSON matching this exact structure:
{
  "summary_selector": "string or empty",
  "content_selector": "string (REQUIRED)",
  "author_selector": "string or empty",
  "tags_selector": "string or empty"
}

**CRITICAL:** 
- content_selector is MANDATORY and must be a valid, specific CSS selector
- All other fields can be empty string "" if not found
- Do NOT include any explanation, only return the JSON object
`, htmlString)
}
