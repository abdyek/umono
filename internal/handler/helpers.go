package handler

import (
	"fmt"
	"html/template"
	"strings"

	umono "github.com/umono-cms/umono"
	"github.com/umono-cms/umono/internal/runtime"

	"github.com/gofiber/fiber/v2"
)

func Render(c *fiber.Ctx, view string, data fiber.Map, layouts ...string) error {
	if data == nil {
		data = fiber.Map{}
	}
	data["I18n"] = c.Locals("I18n")
	data["Lang"] = c.Locals("Lang")
	data["Dir"] = c.Locals("Dir")
	data["IsHTMX"] = c.Locals("IsHTMX")
	data["Version"] = umono.Version
	data["IsAlreadyOnSettings"] = strings.Contains(c.Get("HX-Current-URL"), "/admin/settings")
	data["IsAlreadyOnSitePages"] = strings.Contains(c.Get("HX-Current-URL"), "/admin/site-pages")
	data["IsAlreadyOnComponents"] = strings.Contains(c.Get("HX-Current-URL"), "/admin/components")
	return c.Render(view, data, layouts...)
}

func localizedNotFoundDefaults(c *fiber.Ctx) (string, string) {
	translator, _ := c.Locals("I18n").(interface{ T(string) string })
	if translator == nil {
		return "Page Not Found", "# Page Not Found\nThe page you're looking for may have been moved, deleted, or never existed."
	}

	return translator.T("settings.404_page.default_title"), translator.T("settings.404_page.default_content")
}

func buildPreviewHTML(renderedHTML string) string {
	gridCSS, _ := runtime.GenerateGridCSS(renderedHTML)
	if gridCSS == "" {
		return renderedHTML
	}

	return fmt.Sprintf("<style>%s</style>%s", gridCSS, renderedHTML)
}

func mustPreviewHTML(renderedHTML string) template.HTML {
	return template.HTML(buildPreviewHTML(renderedHTML))
}
