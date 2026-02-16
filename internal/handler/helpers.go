package handler

import (
	umono "github.com/umono-cms/umono"

	"github.com/gofiber/fiber/v2"
)

func Render(c *fiber.Ctx, view string, data fiber.Map, layouts ...string) error {
	if data == nil {
		data = fiber.Map{}
	}
	data["IsHTMX"] = c.Locals("IsHTMX")
	data["Version"] = umono.Version
	return c.Render(view, data, layouts...)
}
