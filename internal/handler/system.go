package handler

import (
	"path"

	"github.com/gofiber/fiber/v2"
	"github.com/umono-cms/umono/internal/service"
	"github.com/umono-cms/umono/internal/view"
)

type systemHandler struct {
	systemService *service.SystemService
}

func NewSystemHandler(ss *service.SystemService) *systemHandler {
	return &systemHandler{
		systemService: ss,
	}
}

func (h *systemHandler) Index(c *fiber.Ctx) error {
	menuItems := h.systemService.MenuItems()
	if len(menuItems) == 0 {
		return c.SendStatus(fiber.StatusOK)
	}

	return c.Redirect("/admin/system/" + menuItems[0].Slug)
}

func (h *systemHandler) RenderJobs(c *fiber.Ctx) error {
	partial := "partials/system-jobs"
	layouts := []string{"layouts/system", "layouts/admin"}

	if isSystemContentSwap(c) {
		partial = "partials/htmx/system-jobs"
		layouts = []string{}
	}

	return Render(c, partial, fiber.Map{
		"SystemUl": h.systemUl(c, path.Base(c.Path())),
	}, layouts...)
}

func (h *systemHandler) systemUl(c *fiber.Ctx, activeSlug string) []view.SettingsLi {
	ul := []view.SettingsLi{}
	translator, _ := c.Locals("I18n").(interface{ T(string) string })
	for _, mi := range h.systemService.MenuItems() {
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

func isSystemContentSwap(c *fiber.Ctx) bool {
	if c.Get("HX-Request") != "true" {
		return false
	}

	switch c.Get("HX-Target") {
	case "system-content", "system-jobs-content":
		return true
	default:
		return false
	}
}
