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
	"github.com/umono-cms/umono/internal/i18n"
	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/repository"
	"github.com/umono-cms/umono/internal/service"
	"github.com/umono-cms/umono/internal/view"
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

	db.AutoMigrate(
		&models.Component{},
		&models.SitePage{},
		&models.Option{},
		&models.Storage{},
		&models.Media{},
		&models.MediaVariant{},
	)

	// TODO: Refactor: move DI another file
	optionRepo := repository.NewOptionRepository(db)
	storageRepo := repository.NewStorageRepository(db)
	mediaRepo := repository.NewMediaRepository(db)

	sitePageRepo := repository.NewSitePageRepository(db)
	sitePageService := service.NewSitePageService(sitePageRepo, optionRepo, comp)

	componentRepo := repository.NewComponentRepository(db)
	componentService := service.NewComponentService(componentRepo, comp)
	componentService.LoadAsGlobalComponent()

	settingsService := service.NewSettingsService()
	storageService := service.NewStorageService(storageRepo)
	bundle, err := i18n.LoadBundle(umono.Locales(), service.DefaultLanguage)
	if err != nil {
		log.Fatal("i18n err", err)
	}
	optionService := service.NewOptionService(optionRepo, bundle)
	mediaService := service.NewMediaService(mediaRepo, storageRepo, optionRepo, "uploads/.pending")
	if err := mediaService.EnsureDefaultLocalStorage("uploads"); err != nil {
		log.Fatal("media storage err", err)
	}

	adminHandler := handler.NewAdminHandler(
		sitePageService,
		componentService,
	)

	previewHandler := handler.NewPreviewHandler(sitePageService, componentService)
	siteHandler := handler.NewSiteHandler(sitePageService)
	sitePageHandler := handler.NewSitePageHandler(sitePageService, componentService)
	componentHandler := handler.NewComponentHandler(componentService, sitePageService)
	mediaHandler := handler.NewMediaHandler(mediaService, storageService, optionService)
	settingsHandler := handler.NewSettingsHandler(settingsService, optionService, sitePageService, storageService)
	optionHandler := handler.NewOptionHandler(optionService)

	engine := html.NewFileSystem(http.FS(umono.Views()), ".html")
	engine.AddFunc("t", view.Translate)

	store := config.NewSessionStore()

	authHandler := handler.NewAuthHandler(store)

	app := fiber.New(fiber.Config{
		Views: engine,
	})

	app.Use(func(c *fiber.Ctx) error {
		c.Set("X-Powered-By", "Umono")
		return c.Next()
	})
	app.Use(middleware.I18nContext(optionService, bundle))

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

	adminProtected.Post("/not-found-page/preview",
		middleware.OnlyHTMX(),
		previewHandler.NotFoundPagePreview,
	)

	adminProtected.Get("/settings", settingsHandler.Index)
	adminProtected.Get("/settings/general", settingsHandler.RenderGeneral)
	adminProtected.Get("/settings/404-page", settingsHandler.Render404Page)
	adminProtected.Get("/settings/storage", settingsHandler.RenderStorageIndex)
	adminProtected.Get("/settings/storage/new", settingsHandler.RenderStorageNew)
	adminProtected.Get("/settings/storage/:id", settingsHandler.RenderStorageShow)
	adminProtected.Post("/settings/storage",
		middleware.OnlyHTMX(),
		settingsHandler.CreateStorage,
	)
	adminProtected.Put("/settings/storage/:id",
		middleware.OnlyHTMX(),
		settingsHandler.UpdateStorage,
	)
	adminProtected.Post("/settings/storage/:id/test",
		middleware.OnlyHTMX(),
		settingsHandler.TestStorage,
	)
	adminProtected.Delete("/settings/storage/:id",
		middleware.OnlyHTMX(),
		settingsHandler.DeleteStorage,
	)
	adminProtected.Get("/settings/about", settingsHandler.RenderAbout)
	adminProtected.Get("/media", mediaHandler.Index)
	adminProtected.Get("/media/new", mediaHandler.RenderNew)
	adminProtected.Get("/media/pending/:token/preview", mediaHandler.ServePending)
	adminProtected.Get("/media/:id", mediaHandler.RenderShow)
	adminProtected.Post("/media",
		middleware.OnlyHTMX(),
		mediaHandler.Upload,
	)
	adminProtected.Post("/media/presign",
		middleware.OnlyHTMX(),
		mediaHandler.PresignUpload,
	)
	adminProtected.Post("/media/complete",
		middleware.OnlyHTMX(),
		mediaHandler.CompleteUpload,
	)
	adminProtected.Post("/media/confirm",
		middleware.OnlyHTMX(),
		mediaHandler.ConfirmUpload,
	)
	adminProtected.Post("/media/cancel",
		middleware.OnlyHTMX(),
		mediaHandler.CancelUpload,
	)
	adminProtected.Put("/media/:id/alias",
		middleware.OnlyHTMX(),
		mediaHandler.UpdateAlias,
	)
	adminProtected.Delete("/media/:id",
		middleware.OnlyHTMX(),
		mediaHandler.Delete,
	)

	adminProtected.Post("/options/language",
		middleware.OnlyHTMX(),
		optionHandler.SaveLanguage,
	)

	adminProtected.Post("/options/not-found-page-option",
		middleware.OnlyHTMX(),
		optionHandler.SaveNotFoundPageOption,
	)

	app.Post("/login", authHandler.Login)
	app.Post("/logout", middleware.Logged(store), authHandler.Logout)
	app.Get("/uploads/:filename", mediaHandler.Serve)

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
