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
			HxGet:     "/admin/site-pages/" + idStr + "/edit",
			HxTarget:  "#editor-area",
			HxPushURL: "/admin/site-pages/" + idStr,
			IsActive:  sp.ID == activeID,
			IsEnabled: sp.Enabled,
		})
	}
	return ul
}
