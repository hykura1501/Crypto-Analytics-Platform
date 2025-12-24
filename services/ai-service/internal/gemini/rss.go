package gemini

import (
	"fmt"
)

var RssResponseSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"title_tag": map[string]any{
			"type":        "string",
			"description": "The specific XML tag name used for the title of an item (e.g. 'title'). Return only the tag name.",
		},
		"link_tag": map[string]any{
			"type":        "string",
			"description": "The specific XML tag name used for the URL/permalink of an item. Prefer 'link'. If 'guid' is explicitly a better permalink (isPermaLink='true') and 'link' is missing, use 'guid'. Return only the tag name.",
		},
		"pub_date_tag": map[string]any{
			"type":        "string",
			"description": "The specific XML tag name used for the publication date of an item (e.g. 'pubDate', 'dc:date'). Return only the tag name.",
		},
	},
	"required": []string{"title_tag", "link_tag", "pub_date_tag"},
}

type RssResponse struct {
	TitleTag   string `json:"title_tag"`
	LinkTag    string `json:"link_tag"`
	PubDateTag string `json:"pub_date_tag"`
}

func BuildRssStructureAnalysisPrompt(xmlString string) string {
	return fmt.Sprintf(`
Analyze the provided XML content (RSS/Atom feed) to identify the correct structure.
Your goal is to determine the XML tag names for:
1. Title
2. Link (The clean, persistent URL to the article)
3. Publication Date

Instruction for Link Extraction:
- Your goal is to identify the tag containing the cleanest, most direct URL to the article.
- Default to 'link' if it appears clean.
- However, if 'link' contains tracking parameters (e.g., utm_source, utm_medium, long query strings) or complex CDATA wrapping, AND 'guid' offers a cleaner, shorter permalink (isPermaLink='true' or default), then PREFER 'guid'.
- Example: If <link> has '...utm_source=rss...' and <guid> is clean, return 'guid'.
- Do NOT return 'guid' if isPermaLink='false'.

XML Content:
%s

Return the specific tag names found.
`, xmlString)
}
