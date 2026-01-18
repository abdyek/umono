package handler

import (
	"strconv"

	"github.com/umono-cms/umono/internal/models"
)

type SitePageLi struct {
	Title     string
	PageURL   string
	HxGet     string
	HxTarget  string
	HxPushURL string
	IsActive  bool
	IsEnabled bool
}

func BuildSitePageUl(pages []models.SitePage, activeID uint) []SitePageLi {
	var ul []SitePageLi
	for _, sp := range pages {
		idStr := strconv.FormatUint(uint64(sp.ID), 10)
		ul = append(ul, SitePageLi{
			Title:     sp.Name,
			PageURL:   "/admin/site-pages/" + idStr,
			HxGet:     "/admin/site-pages/" + idStr + "/editor",
			HxTarget:  "#editor-area",
			HxPushURL: "/admin/site-pages/" + idStr,
			IsActive:  sp.ID == activeID,
			IsEnabled: sp.Enabled,
		})
	}
	return ul
}

type SitePageEditor struct {
	Name           string
	Slug           string
	Content        string
	Output         string
	IsEnabled      bool
	LastModifiedAt string
}

func BuildSitePageEditor(page models.SitePage) SitePageEditor {
	return SitePageEditor{
		Name:           page.Name,
		Slug:           page.Slug,
		Content:        page.Content,
		Output:         "here is output to preview",
		IsEnabled:      page.Enabled,
		LastModifiedAt: "2 hours ago", // TODO: get relative time string from page.LastModifiedAt
	}
}
