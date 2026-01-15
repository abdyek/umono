package middleware

import "github.com/gofiber/fiber/v2"

func SmartRedirect(c *fiber.Ctx, url string) error {
	if c.Get("HX-Request") == "true" {
		c.Set("HX-Redirect", url)
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	return c.Redirect(url, fiber.StatusSeeOther)
}
