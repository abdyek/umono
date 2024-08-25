package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/umono-cms/umono/controllers"
	"github.com/umono-cms/umono/middlewares"
)

func main() {

	if err := godotenv.Load(); err != nil {
		panic("Error loading .env file!")
	}

	app := fiber.New()

	app.Static("/public", "./static")

	app.Get("/api/v1/me", controllers.Me)

	api := app.Group("/api/v1", middlewares.Guest())
	api.Post("/login", controllers.Login)

	log.Fatal(app.Listen(":8999"))
}
