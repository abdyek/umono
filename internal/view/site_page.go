package view

import (
	"html/template"

	"github.com/umono-cms/umono/internal/models"
)

type SitePageEditorData struct {
	ID             uint
	Name           string
	Slug           string
	Content        string
	Output         template.HTML
	IsEnabled      bool
	LastModifiedAt string
	SlugErr        string
	NameErr        string
}

func SitePageEditor(sp models.SitePage, output, slugErr, nameErr string) SitePageEditorData {
	return SitePageEditorData{
		ID:             sp.ID,
		Name:           sp.Name,
		Slug:           sp.Slug,
		Content:        sp.Content,
		Output:         template.HTML(output),
		IsEnabled:      sp.Enabled,
		LastModifiedAt: RelativeTime(sp.LastModifiedAt),
		SlugErr:        slugErr,
		NameErr:        nameErr,
	}
}

type SitePageLi struct {
	ID        uint
	Title     string
	IsActive  bool
	IsEnabled bool
}

func SitePageUl(sitePages []models.SitePage, activeID uint) []SitePageLi {
	ul := []SitePageLi{}
	for _, sp := range sitePages {
		ul = append(ul, SitePageLi{
			ID:        sp.ID,
			Title:     sp.Name,
			IsActive:  sp.ID == activeID,
			IsEnabled: sp.Enabled,
		})
	}
	return ul
}
