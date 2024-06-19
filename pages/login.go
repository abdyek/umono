package pages

import "github.com/gofiber/fiber/v2"

func Login(c *fiber.Ctx) error {
	return c.SendFile("./html/login.html")
}
