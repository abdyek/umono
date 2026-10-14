package view

import (
	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/service"
)

type MediaListItem struct {
	ID          string
	Name        string
	Alias       string
	URL         string
	IsActive    bool
	ContentType string
	Size        int64
}

func MediaList(items []models.Media, activeID string, publicURL func(models.Media) string) []MediaListItem {
	out := make([]MediaListItem, 0, len(items))
	for _, item := range items {
		out = append(out, MediaListItem{
			ID:          item.ID,
			Name:        item.OriginalName,
			Alias:       service.MediaAlias(item),
			URL:         publicURL(item),
			IsActive:    item.ID == activeID,
			ContentType: item.MimeType,
			Size:        item.Size,
		})
	}
	return out
}

// MediaSidebar is one page of the media sidebar. Next is the offset of
// the next page, or 0 when there is none.
type MediaSidebar struct {
	Items    []MediaListItem
	Next     int
	ActiveID string
}
