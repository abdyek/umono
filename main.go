package main

import (
	"html/template"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/joho/godotenv"
	"github.com/umono-cms/umono/controllers"
	"github.com/umono-cms/umono/database"
	"github.com/umono-cms/umono/middlewares"
	"github.com/umono-cms/umono/storage"
	"github.com/umono-cms/umono/validation"
)

func main() {

	if err := godotenv.Load(); err != nil {
		panic("Error loading .env file!")
	}

	err := database.Init(os.Getenv("DSN"))
	if err != nil {
		panic(err.Error())
	}

	validation.Init()

	storage.InitPageStorage()
	storage.Page.LoadAll(database.DB)

	engine := html.New("./views", ".html")

	engine.AddFunc("safe", func(s string) template.HTML {
		return template.HTML(s)
	})

	app := fiber.New(fiber.Config{
		Views: engine,
	})

	app.Static("/public", "./static")

	if os.Getenv("DEV") == "true" {
		app.Use(func() func(*fiber.Ctx) error {
			return func(c *fiber.Ctx) error {
				// NOTE: slow response for UI development
				time.Sleep(500 * time.Millisecond)
				return c.Next()
			}
		}())
	}

	app.Post("/api/v1/login", middlewares.Guest(), controllers.Login)
	app.Get("/:slug", controllers.ServePage)
	app.Get("/api/v1/me", controllers.Me)

	api := app.Group("/api/v1", middlewares.Authenticator())

	api.Post("/logout", controllers.Logout)

	api.Post("/pages", controllers.CreatePage)
	api.Get("/pages/:ID", controllers.ReadPage)
	api.Get("/pages", controllers.ReadAllPages)
	api.Put("/pages", controllers.UpdatePage)
	api.Delete("/pages/:ID", controllers.DeletePage)

	api.Post("/converter/markdown-to-html", controllers.MDToHTML)

	log.Fatal(app.Listen(":8999"))
}
