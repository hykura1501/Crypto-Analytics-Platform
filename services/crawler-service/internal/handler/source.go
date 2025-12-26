package handler

import (
	"database/sql"
	"fmt"
)

type Source struct {
	SourceID        string `json:"source_id"`
	RssURL          string `json:"rss_url"`
	TitleTag        string `json:"title_tag,omitempty"`
	LinkTag         string `json:"link_tag,omitempty"`
	PubDateTag      string `json:"pub_date_tag,omitempty"`
	SummarySelector string `json:"summary_selector,omitempty"`
	ContentSelector string `json:"content_selector,omitempty"`
	AuthorSelector  string `json:"author_selector,omitempty"`
	TagsSelector    string `json:"tags_selector,omitempty"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

type CreateSourceRequest struct {
	SourceID string `json:"source_id"`
	RssURL   string `json:"rss_url"`
}

type UpdateSourceRequest struct {
	RssURL          *string `json:"rss_url"`
	TitleTag        *string `json:"title_tag"`
	LinkTag         *string `json:"link_tag"`
	PubDateTag      *string `json:"pub_date_tag"`
	SummarySelector *string `json:"summary_selector"`
	ContentSelector *string `json:"content_selector"`
	AuthorSelector  *string `json:"author_selector"`
	TagsSelector    *string `json:"tags_selector"`
}

type SourceHandler struct {
	DB *sql.DB
}

func NewSourceHandler(db *sql.DB) *SourceHandler {
	return &SourceHandler{DB: db}
}

func (h *SourceHandler) ListSources() ([]Source, error) {
	rows, err := h.DB.Query(`
		SELECT source_id, rss_url, title_tag, link_tag, pub_date_tag, 
		       summary_selector, content_selector, author_selector, tags_selector,
		       created_at, updated_at
		FROM sources
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sources := []Source{}
	for rows.Next() {
		var s Source
		var titleTag, linkTag, pubDateTag sql.NullString
		var summarySelector, contentSelector, authorSelector, tagsSelector sql.NullString

		err := rows.Scan(
			&s.SourceID, &s.RssURL, &titleTag, &linkTag, &pubDateTag,
			&summarySelector, &contentSelector, &authorSelector, &tagsSelector,
			&s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Convert sql.NullString to string
		if titleTag.Valid {
			s.TitleTag = titleTag.String
		}
		if linkTag.Valid {
			s.LinkTag = linkTag.String
		}
		if pubDateTag.Valid {
			s.PubDateTag = pubDateTag.String
		}
		if summarySelector.Valid {
			s.SummarySelector = summarySelector.String
		}
		if contentSelector.Valid {
			s.ContentSelector = contentSelector.String
		}
		if authorSelector.Valid {
			s.AuthorSelector = authorSelector.String
		}
		if tagsSelector.Valid {
			s.TagsSelector = tagsSelector.String
		}

		sources = append(sources, s)
	}
	return sources, nil
}

func (h *SourceHandler) GetSource(sourceID string) (*Source, error) {
	var s Source
	var titleTag, linkTag, pubDateTag sql.NullString
	var summarySelector, contentSelector, authorSelector, tagsSelector sql.NullString

	err := h.DB.QueryRow(`
		SELECT source_id, rss_url, title_tag, link_tag, pub_date_tag, 
		       summary_selector, content_selector, author_selector, tags_selector,
		       created_at, updated_at
		FROM sources
		WHERE source_id = $1
	`, sourceID).Scan(
		&s.SourceID, &s.RssURL, &titleTag, &linkTag, &pubDateTag,
		&summarySelector, &contentSelector, &authorSelector, &tagsSelector,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Convert sql.NullString to string
	if titleTag.Valid {
		s.TitleTag = titleTag.String
	}
	if linkTag.Valid {
		s.LinkTag = linkTag.String
	}
	if pubDateTag.Valid {
		s.PubDateTag = pubDateTag.String
	}
	if summarySelector.Valid {
		s.SummarySelector = summarySelector.String
	}
	if contentSelector.Valid {
		s.ContentSelector = contentSelector.String
	}
	if authorSelector.Valid {
		s.AuthorSelector = authorSelector.String
	}
	if tagsSelector.Valid {
		s.TagsSelector = tagsSelector.String
	}

	return &s, nil
}

func (h *SourceHandler) CreateSource(req CreateSourceRequest) error {
	_, err := h.DB.Exec(`
		INSERT INTO sources (source_id, rss_url)
		VALUES ($1, $2)
	`, req.SourceID, req.RssURL)
	return err
}

func (h *SourceHandler) UpdateSource(sourceID string, req UpdateSourceRequest) error {
	query := `UPDATE sources SET updated_at = NOW()`
	args := []interface{}{}
	argIndex := 1

	if req.RssURL != nil {
		query += fmt.Sprintf(`, rss_url = $%d`, argIndex)
		args = append(args, *req.RssURL)
		argIndex++
	}
	if req.TitleTag != nil {
		query += fmt.Sprintf(`, title_tag = $%d`, argIndex)
		args = append(args, *req.TitleTag)
		argIndex++
	}
	if req.LinkTag != nil {
		query += fmt.Sprintf(`, link_tag = $%d`, argIndex)
		args = append(args, *req.LinkTag)
		argIndex++
	}
	if req.PubDateTag != nil {
		query += fmt.Sprintf(`, pub_date_tag = $%d`, argIndex)
		args = append(args, *req.PubDateTag)
		argIndex++
	}
	if req.SummarySelector != nil {
		query += fmt.Sprintf(`, summary_selector = $%d`, argIndex)
		args = append(args, *req.SummarySelector)
		argIndex++
	}
	if req.ContentSelector != nil {
		query += fmt.Sprintf(`, content_selector = $%d`, argIndex)
		args = append(args, *req.ContentSelector)
		argIndex++
	}
	if req.AuthorSelector != nil {
		query += fmt.Sprintf(`, author_selector = $%d`, argIndex)
		args = append(args, *req.AuthorSelector)
		argIndex++
	}
	if req.TagsSelector != nil {
		query += fmt.Sprintf(`, tags_selector = $%d`, argIndex)
		args = append(args, *req.TagsSelector)
		argIndex++
	}

	query += fmt.Sprintf(` WHERE source_id = $%d`, argIndex)
	args = append(args, sourceID)

	result, err := h.DB.Exec(query, args...)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (h *SourceHandler) DeleteSource(sourceID string) error {
	result, err := h.DB.Exec(`DELETE FROM sources WHERE source_id = $1`, sourceID)
	if err != nil {
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}
