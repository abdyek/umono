package main

import (
	"encoding/base64"
	"html/template"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	"github.com/gofiber/template/html/v2"
	"github.com/joho/godotenv"
	"github.com/umono-cms/umono/controllers"
	"github.com/umono-cms/umono/database"
	"github.com/umono-cms/umono/middlewares"
	"github.com/umono-cms/umono/storage"
	"github.com/umono-cms/umono/umono"
	"github.com/umono-cms/umono/validation"
	"golang.org/x/crypto/bcrypt"
)

func main() {

	if err := godotenv.Load(); err != nil {
		panic("Error loading .env file!")
	}

	if os.Getenv("USERNAME") != "" && os.Getenv("PASSWORD") != "" {
		err := updateEnvFile()
		if err != nil {
			panic("Error updating .env file" + err.Error())
		}
	}

	err := database.Init(os.Getenv("DSN"))
	if err != nil {
		panic(err.Error())
	}

	umono.InitLang()
	umono.SetGlobalComponents(database.DB)

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

	app.Use("/admin/*", func(c *fiber.Ctx) error {
		return proxy.Do(c, "http://admin-ui:80"+c.OriginalURL())
	})

	app.Use("/assets/*", func(c *fiber.Ctx) error {
		return proxy.Do(c, "http://admin-ui:80"+c.OriginalURL())
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
	app.Get("/:slug?", controllers.ServePage)
	app.Get("/api/v1/me", controllers.Me)

	api := app.Group("/api/v1", middlewares.Authenticator())

	api.Post("/logout", controllers.Logout)

	api.Post("/pages", controllers.CreatePage)
	api.Get("/pages/:ID", controllers.ReadPage)
	api.Get("/pages", controllers.ReadAllPages)
	api.Put("/pages", controllers.UpdatePage)
	api.Delete("/pages/:ID", controllers.DeletePage)
	api.Get("/slug/check-usable/:slug?", controllers.SlugCheckUsable)

	// NOTE: Deprecated - Remove it for v1
	api.Post("/converter/markdown-to-html", controllers.MDToHTML)
	api.Post("/converter/umono-lang-to-html", controllers.UmonoLangToHTML)

	api.Post("/components", controllers.CreateComponent)
	api.Get("/components/:ID", controllers.ReadComponent)
	api.Get("/components", controllers.ReadAllComponents)
	api.Put("/components", controllers.UpdateComponent)
	api.Delete("/components/:ID", controllers.DeleteComponent)

	log.Fatal(app.Listen(":8999"))
}

func updateEnvFile() error {

	hashedUsername, err := hashData(os.Getenv("USERNAME"))
	if err != nil {
		return err
	}

	hashedPassword, err := hashData(os.Getenv("PASSWORD"))
	if err != nil {
		return err
	}

	content := ""

	content += "DEV=" + os.Getenv("DEV") + "\n\n"
	content += "USERNAME=\n"
	content += "PASSWORD=\n\n"
	content += "HASHED_USERNAME=" + base64.StdEncoding.EncodeToString([]byte(hashedUsername)) + "\n"
	content += "HASHED_PASSWORD=" + base64.StdEncoding.EncodeToString([]byte(hashedPassword)) + "\n\n"
	content += "DSN=" + os.Getenv("DSN")

	file, err := os.OpenFile(".env", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return err
	}

	if err := godotenv.Overload(); err != nil {
		return err
	}

	return nil
}

func hashData(data string) (string, error) {
	hashedData, err := bcrypt.GenerateFromPassword([]byte(data), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedData), nil
}
