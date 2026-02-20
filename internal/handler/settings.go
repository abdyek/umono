package handler

import (
	"path"

	"github.com/gofiber/fiber/v2"
	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/service"
	"github.com/umono-cms/umono/internal/view"
)

type settingsHandler struct {
	settingsService *service.SettingsService
	optionService   *service.OptionService
	sitePageService *service.SitePageService
}

func NewSettingsHandler(ss *service.SettingsService, os *service.OptionService, sps *service.SitePageService) *settingsHandler {
	return &settingsHandler{
		settingsService: ss,
		optionService:   os,
		sitePageService: sps,
	}
}

func (h *settingsHandler) Index(c *fiber.Ctx) error {
	menuItems := h.settingsService.MenuItems()
	if len(menuItems) == 0 {
		return c.SendStatus(fiber.StatusOK)
	}

	return c.Redirect("/admin/settings/" + menuItems[0].Slug)
}

func (h *settingsHandler) Render404Page(c *fiber.Ctx) error {
	partial := "partials/settings-404-page"
	layouts := []string{"layouts/settings", "layouts/admin"}

	if c.Get("HX-Request") == "true" {
		if c.Get("HX-Target") == "settings-content" || c.Get("HX-Target") == "settings-404-page-content" {
			partial = "partials/htmx/settings-404-page"
			layouts = []string{}
		}
	}

	title, content := h.optionService.GetNotFoundPageOption()

	previewContent := content
	if content == "" {
		previewContent = models.DefaultNotFoundContent
	}

	return Render(c, partial, fiber.Map{
		"SettingsUl":     h.settingsUl(path.Base(c.Path())),
		"DefaultTitle":   models.DefaultNotFoundTitle,
		"DefaultContent": models.DefaultNotFoundContent,
		"Title":          title,
		"Content":        content,
		"Output":         h.sitePageService.MustPreview(previewContent),
	}, layouts...)
}

func (h *settingsHandler) RenderAbout(c *fiber.Ctx) error {
	partial := "partials/settings-about"
	layouts := []string{"layouts/settings", "layouts/admin"}

	if c.Get("HX-Request") == "true" {
		partial = "partials/htmx/settings-about"
		layouts = []string{}
	}

	return Render(c, partial, fiber.Map{
		"SettingsUl": h.settingsUl(path.Base(c.Path())),
	}, layouts...)
}

func (h *settingsHandler) settingsUl(activeSlug string) []view.SettingsLi {
	ul := []view.SettingsLi{}
	for _, mi := range h.settingsService.MenuItems() {
		ul = append(ul, view.SettingsLi{
			Title:    mi.Title,
			Slug:     mi.Slug,
			IsActive: mi.Slug == activeSlug,
		})
	}
	return ul
}
