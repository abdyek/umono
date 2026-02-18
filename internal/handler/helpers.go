package handler

import (
	"strings"

	umono "github.com/umono-cms/umono"

	"github.com/gofiber/fiber/v2"
)

func Render(c *fiber.Ctx, view string, data fiber.Map, layouts ...string) error {
	if data == nil {
		data = fiber.Map{}
	}
	data["IsHTMX"] = c.Locals("IsHTMX")
	data["Version"] = umono.Version
	data["IsAlreadyOnSettings"] = strings.Contains(c.Get("HX-Current-URL"), "/admin/settings")
	data["IsAlreadyOnSitePages"] = strings.Contains(c.Get("HX-Current-URL"), "/admin/site-pages")
	data["IsAlreadyOnComponents"] = strings.Contains(c.Get("HX-Current-URL"), "/admin/components")
	return c.Render(view, data, layouts...)
}
