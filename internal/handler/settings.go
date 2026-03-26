package handler

import (
	"path"

	"github.com/gofiber/fiber/v2"
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

	if isSettingsContentSwap(c) {
		partial = "partials/htmx/settings-404-page"
		layouts = []string{}
	}

	title, content := h.optionService.GetNotFoundPageOption()
	defaultTitle, defaultContent := localizedNotFoundDefaults(c)

	previewContent := content
	if content == "" {
		previewContent = defaultContent
	}

	return Render(c, partial, fiber.Map{
		"SettingsUl":     h.settingsUl(c, path.Base(c.Path())),
		"DefaultTitle":   defaultTitle,
		"DefaultContent": defaultContent,
		"Title":          title,
		"Content":        content,
		"Output":         mustPreviewHTML(h.sitePageService.MustPreview(previewContent)),
	}, layouts...)
}

func (h *settingsHandler) RenderGeneral(c *fiber.Ctx) error {
	partial := "partials/settings-general"
	layouts := []string{"layouts/settings", "layouts/admin"}

	if isSettingsContentSwap(c) {
		partial = "partials/htmx/settings-general"
		layouts = []string{}
	}

	return Render(c, partial, fiber.Map{
		"SettingsUl": h.settingsUl(c, path.Base(c.Path())),
		"Language":   h.optionService.GetLanguage(),
		"Languages":  h.optionService.SupportedLanguages(),
	}, layouts...)
}

func (h *settingsHandler) RenderAbout(c *fiber.Ctx) error {
	partial := "partials/settings-about"
	layouts := []string{"layouts/settings", "layouts/admin"}

	if isSettingsContentSwap(c) {
		partial = "partials/htmx/settings-about"
		layouts = []string{}
	}

	return Render(c, partial, fiber.Map{
		"SettingsUl": h.settingsUl(c, path.Base(c.Path())),
	}, layouts...)
}

func (h *settingsHandler) settingsUl(c *fiber.Ctx, activeSlug string) []view.SettingsLi {
	ul := []view.SettingsLi{}
	translator, _ := c.Locals("I18n").(interface{ T(string) string })
	for _, mi := range h.settingsService.MenuItems() {
		title := mi.TitleKey
		if translator != nil {
			title = translator.T(mi.TitleKey)
		}

		ul = append(ul, view.SettingsLi{
			Title:    title,
			Slug:     mi.Slug,
			IsActive: mi.Slug == activeSlug,
		})
	}
	return ul
}

func isSettingsContentSwap(c *fiber.Ctx) bool {
	if c.Get("HX-Request") != "true" {
		return false
	}

	switch c.Get("HX-Target") {
	case "settings-content", "settings-general-content", "settings-404-page-content", "settings-about-content":
		return true
	default:
		return false
	}
}
