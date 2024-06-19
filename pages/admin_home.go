package pages

import "github.com/gofiber/fiber/v2"

func AdminHome(c *fiber.Ctx) error {
	return c.SendFile("./html/admin_home.html")
}
