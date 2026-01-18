package middleware

import "github.com/gofiber/fiber/v2"

func OnlyHTMX() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Get("HX-Request") != "true" {
			return c.SendStatus(fiber.StatusNotFound)
		}
		return c.Next()
	}
}
