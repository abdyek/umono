package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/umono-cms/umono/controllers"
	"github.com/umono-cms/umono/middlewares"
	"github.com/umono-cms/umono/pages"
)

func main() {

	if err := godotenv.Load(); err != nil {
		panic("Error loading .env file!")
	}

	app := fiber.New()

	app.Static("/public", "./static")

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Hello, CMS users 👋!",
		})
	})

	// TODO: If the user logged already, redirect home page
	app.Get("/login", pages.Login)

	adminPages := app.Group("/admin", middlewares.Authenticator())
	adminPages.Get("/", pages.AdminHome)

	api := app.Group("/api/v1")
	api.Post("/login", controllers.Login)

	log.Fatal(app.Listen(":8999"))
}
