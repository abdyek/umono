package handler

import (
	"github.com/gofiber/fiber/v2"
)

func Render(c *fiber.Ctx, view string, data fiber.Map, layouts ...string) error {
	if data == nil {
		data = fiber.Map{}
	}
	data["IsHTMX"] = c.Locals("IsHTMX")
	return c.Render(view, data, layouts...)
}
