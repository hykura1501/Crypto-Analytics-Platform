package crawler

type rssArticle struct {
	SourceID    string
	URL         string
	Title       string
	PublishedAt string
	Author      string
	Language    string
}

type sourceMeta struct {
	name            string
	rssURL          string
	language        string
	titleTag        string
	linkTag         string
	pubDateTag      string
	contentSelector string
	authorSelector  string
	summarySelector string
	tagsSelector    string
}

type articleContent struct {
	content     string
	author      string
	summary     string
	tags        string
	publishedAt string
}
