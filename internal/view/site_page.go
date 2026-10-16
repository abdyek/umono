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
	Diagnostics    []DiagnosticItem
	IsEnabled      bool
	LastModifiedAt string
	SlugErr        string
	NameErr        string
	// Layout is the Layout field and PublicURL the page's public address,
	// "" for a new page.
	Layout    LayoutField
	PublicURL string
	I18n      any
}

func SitePageEditor(sp models.SitePage, output template.HTML, diagnostics []DiagnosticItem, slugErr, nameErr string, translator any) SitePageEditorData {
	return SitePageEditorData{
		ID:             sp.ID,
		Name:           sp.Name,
		Slug:           sp.Slug,
		Content:        sp.Content,
		Output:         output,
		Diagnostics:    diagnostics,
		IsEnabled:      sp.Enabled,
		LastModifiedAt: RelativeTimeWithTranslator(sp.LastModifiedAt, translator),
		SlugErr:        slugErr,
		NameErr:        nameErr,
		I18n:           translator,
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

// SitePageList is one page of the site page sidebar. Next is the offset of
// the next page, or 0 when there is none.
type SitePageList struct {
	Items    []SitePageLi
	Next     int
	ActiveID uint
}
