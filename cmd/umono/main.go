package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/template/html/v2"
	"github.com/joho/godotenv"
	"github.com/umono-cms/compono"
	"github.com/umono-cms/umono"
	"github.com/umono-cms/umono/internal/config"
	"github.com/umono-cms/umono/internal/handler"
	"github.com/umono-cms/umono/internal/handler/middleware"
	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/repository"
	"github.com/umono-cms/umono/internal/service"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	if err := godotenv.Load(); err != nil {
		panic("Error loading .env file!")
	}

	if os.Getenv("SECRET") == "" {

		bytes := make([]byte, 32)
		_, err := rand.Read(bytes)
		if err != nil {
			panic("SECRET could not generate.")
		}

		os.Setenv("SECRET", hex.EncodeToString(bytes))
	}

	if os.Getenv("USERNAME") != "" && os.Getenv("PASSWORD") != "" {
		err := updateEnvFile()
		if err != nil {
			panic("Error updating .env file" + err.Error())
		}
	}

	db, err := gorm.Open(sqlite.Open("umono.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("db err", err)
	}

	comp := compono.New()

	db.AutoMigrate(&models.Component{}, &models.SitePage{})

	// TODO: Refactor: move DI another file
	sitePageRepo := repository.NewSitePageRepository(db)
	sitePageService := service.NewSitePageService(sitePageRepo, comp)

	componentRepo := repository.NewComponentRepository(db)
	componentService := service.NewComponentService(componentRepo, comp)
	componentService.LoadAsGlobalComponent()

	adminHandler := handler.NewAdminHandler(
		sitePageService,
		componentService,
	)

	previewHandler := handler.NewPreviewHandler(sitePageService, componentService)
	siteHandler := handler.NewSiteHandler(sitePageService)
	sitePageHandler := handler.NewSitePageHandler(sitePageService, componentService)
	componentHandler := handler.NewComponentHandler(componentService, sitePageService)

	engine := html.NewFileSystem(http.FS(umono.Views()), ".html")

	store := config.NewSessionStore()

	authHandler := handler.NewAuthHandler(store)

	app := fiber.New(fiber.Config{
		Views: engine,
	})

	app.Use(func(c *fiber.Ctx) error {
		c.Set("X-Powered-By", "Umono")
		return c.Next()
	})

	app.Use("/static", func(c *fiber.Ctx) error {
		c.Set("Cache-Control", "public, max-age=31536000, immutable")
		return c.Next()
	}, filesystem.New(filesystem.Config{
		Root:   http.FS(umono.Public()),
		Browse: false,
	}))

	if os.Getenv("APP_ENV") == "dev" {
		app.Use(func(c *fiber.Ctx) error {
			if c.Get("HX-Request") == "true" {
				time.Sleep(400 * time.Millisecond)
			}
			return c.Next()
		})
	}
	admin := app.Group("/admin")
	admin.Use(middleware.HTMXContext())

	admin.Get("/login", middleware.Guest(store), authHandler.RenderLogin)

	adminProtected := admin.Group("/", middleware.Logged(store))

	adminProtected.Get("/", adminHandler.Index)

	adminProtected.Get("/site-pages/new", sitePageHandler.RenderNewPageSiteEditor)
	adminProtected.Get("/components/new", componentHandler.RenderComponentEditor)

	adminProtected.Get("/site-pages/check-slug",
		middleware.OnlyHTMX(),
		sitePageHandler.CheckSlug,
	)

	adminProtected.Get("/site-pages/:id", adminHandler.RenderAdminSitePage)
	adminProtected.Get("/site-pages/:id/editor",
		middleware.OnlyHTMX(),
		adminHandler.RenderAdminSitePageEditor,
	)

	adminProtected.Post("/site-pages",
		middleware.OnlyHTMX(),
		sitePageHandler.Create,
	)

	adminProtected.Put("/site-pages/:id",
		middleware.OnlyHTMX(),
		sitePageHandler.Update,
	)

	adminProtected.Delete("/site-pages/:id",
		middleware.OnlyHTMX(),
		sitePageHandler.Delete,
	)

	adminProtected.Get("/components/:id", adminHandler.RenderAdminComponent)
	adminProtected.Get("/components/:id/editor",
		middleware.OnlyHTMX(),
		adminHandler.RenderAdminComponentEditor,
	)

	adminProtected.Post("/components",
		middleware.OnlyHTMX(),
		componentHandler.Create,
	)

	adminProtected.Put("/components/:id",
		middleware.OnlyHTMX(),
		componentHandler.Update,
	)

	adminProtected.Delete("/components/:id",
		middleware.OnlyHTMX(),
		componentHandler.Delete,
	)

	adminProtected.Post("/site-pages/preview",
		middleware.OnlyHTMX(),
		previewHandler.RenderSitePagePreview,
	)

	adminProtected.Post("/components/preview",
		middleware.OnlyHTMX(),
		previewHandler.RenderComponentPreview,
	)

	app.Post("/login", authHandler.Login)
	app.Post("/logout", middleware.Logged(store), authHandler.Logout)

	app.Get("/:slug?", siteHandler.RenderSitePage)

	port := ":8999"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = ":" + envPort
	}

	log.Fatal(app.Listen(port))
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
	content += "APP_ENV=" + os.Getenv("APP_ENV") + "\n"
	content += "SESSION_DRIVER=" + os.Getenv("SESSION_DRIVER") + "\n\n"

	content += "PORT=" + os.Getenv("PORT") + "\n"
	content += "DSN=" + os.Getenv("DSN") + "\n\n"

	content += "USERNAME=\n"
	content += "PASSWORD=\n\n"

	content += "HASHED_USERNAME=" + base64.StdEncoding.EncodeToString([]byte(hashedUsername)) + "\n"
	content += "HASHED_PASSWORD=" + base64.StdEncoding.EncodeToString([]byte(hashedPassword)) + "\n\n"

	file, err := os.OpenFile(".env", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o666)
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
