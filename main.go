package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/umono-cms/umono/api"
	"github.com/umono-cms/umono/pages"
)

func main() {
	app := fiber.New()

	app.Static("/public", "./static")

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Hello, CMS users 👋!",
		})
	})

	app.Get("/login", pages.Login)

	app.Post("/api/v1/login", api.Login)

	log.Fatal(app.Listen(":8999"))
}
