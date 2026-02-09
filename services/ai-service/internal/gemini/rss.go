package gemini

import (
	"fmt"
)

var RssResponseSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"title_tag": map[string]any{
			"type":        "string",
			"description": "Title tag",
		},
		"link_tag": map[string]any{
			"type":        "string",
			"description": "Link tag",
		},
		"pub_date_tag": map[string]any{
			"type":        "string",
			"description": "Pub date tag",
		},
	},
	"required": []string{"title_tag", "link_tag", "pub_date_tag"},
}

type RssResponse struct {
	TitleTag   string `json:"title_tag"`
	LinkTag    string `json:"link_tag"`
	PubDateTag string `json:"pub_date_tag"`
}

func BuildRssStructureAnalysisPrompt() string {
	return fmt.Sprintf(`
		
	`)
}
