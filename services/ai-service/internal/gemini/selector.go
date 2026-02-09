package gemini

import (
	"fmt"
)

var SelectorResponseSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"summary_selector": map[string]any{
			"type":        "string",
			"description": "Single CSS selector for the article summary/excerpt. Return empty string if not found.",
		},
		"content_selector": map[string]any{
			"type":        "array",
			"description": "List of selectors from generic container to specific paragraphs",
			"items": map[string]any{
				"type": "string",
			},
		},
		"author_selector": map[string]any{
			"type":        "string",
			"description": "Single CSS selector for the author name. Return empty string if not found.",
		},
		"tags_selector": map[string]any{
			"type":        "string",
			"description": "Single CSS selector for the article tags/categories. Return empty string if not found.",
		},
	},
	"required": []string{"content_selector"},
}

type SelectorResponse struct {
	SummarySelector string   `json:"summary_selector"`
	ContentSelector []string `json:"content_selector"`
	AuthorSelector  string   `json:"author_selector"`
	TagsSelector    string   `json:"tags_selector"`
}

func BuildSelectorAnalysisPrompt(htmlString string) string {
	return fmt.Sprintf(`
You are an expert Web Scraper and HTML Analyst. Your task is to analyze the provided HTML and identify the most robust, semantic CSS selectors to extract article data.

**INPUT HTML:**
%s

**CRITICAL RULES (MUST FOLLOW):**
1. **NO BARE TAGS:** NEVER return a generic tag selector like "p", "div", "span", "section" on its own. It causes data leakage.
2. **SCOPED SELECTORS:** Content paragraph selectors MUST include the parent class or ID to limit the scope (e.g., instead of "p", return ".article-body p" or "#content > p").
3. **PRIORITY:**
   - Priority 1: Semantic attributes (e.g., 'itemprop', 'data-testid', 'id="article-content"').
   - Priority 2: Stable class names (e.g., '.post-content', '.entry-body').
   - Priority 3: Path-based selectors (only if no ID/Class exists).

**FIELD INSTRUCTIONS:**

- **summary**: Selector for the short description/excerpt. Check '<meta name="description">' if not visible.

- **content**: Return an ARRAY of selectors.
  - Item 1: The **Container** wrapping the text (e.g., 'div.document-body').
  - Item 2+: **Scoped Paragraphs**. You MUST combine the container selector with the paragraph tag.
    - *Bad Example:* '["div.body", "p"]' (REJECT THIS)
    - *Good Example:* '["div.body", "div.body p"]' or '["div.body", "div.body > p"]'

- **author**: Selector for the author's name.
- **tags**: Selector for tags/categories.

**OUTPUT:**
Generate the JSON object based on the provided schema.
	`, htmlString)
}
