package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/umono-cms/umono/internal/models"
)

func Render(c *fiber.Ctx, view string, data fiber.Map, layouts ...string) error {
	if data == nil {
		data = fiber.Map{}
	}
	data["IsHTMX"] = c.Locals("IsHTMX")
	return c.Render(view, data, layouts...)
}

// TODO: Move Hx.. and PageURL into the template
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

type ComponentLi struct {
	ID       string
	Name     string
	IsActive bool
}

func BuildComponentUl(comps []models.Component, activeID uint) []ComponentLi {
	var ul []ComponentLi
	for _, c := range comps {
		idStr := strconv.FormatUint(uint64(c.ID), 10)
		ul = append(ul, ComponentLi{
			ID:       idStr,
			Name:     c.Name,
			IsActive: c.ID == activeID,
		})
	}
	return ul
}

type ComponentEditor struct {
	Name           string
	Content        string
	Output         string
	LastModifiedAt string
}

func BuildComponentEditor(comp models.Component) ComponentEditor {
	return ComponentEditor{
		Name:           comp.Name,
		Content:        comp.Content,
		Output:         "here is output to preview",
		LastModifiedAt: "2 hours ago", // TODO: get relative time string from component.LastModifiedAt
	}
}
