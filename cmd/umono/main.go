package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/joho/godotenv"
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

	db.AutoMigrate(&models.Component{}, &models.SitePage{})

	sitePageRepo := repository.NewSitePageRepository(db)
	sitePageService := service.NewSitePageService(sitePageRepo)

	componentRepo := repository.NewComponentRepository(db)
	componentService := service.NewComponentService(componentRepo)

	pageHandler := handler.NewPageHandler(
		sitePageService,
		componentService,
	)

	engine := html.New("./views", ".html")

	store := config.NewSessionStore()

	loginHandler := handler.NewLoginHandler(store)

	app := fiber.New(fiber.Config{
		Views: engine,
	})

	app.Static("/static", "./public")

	admin := app.Group("/admin")
	admin.Use(middleware.HTMXContext())

	admin.Get("/login", middleware.Guest(store), pageHandler.RenderLogin)

	adminProtected := admin.Group("/", middleware.Logged(store))

	adminProtected.Get("/", pageHandler.RenderAdmin)

	adminProtected.Get("/site-pages/:id", pageHandler.RenderAdminSitePage)
	adminProtected.Get("/site-pages/:id/editor",
		middleware.OnlyHTMX(),
		pageHandler.RenderAdminSitePageEditor,
	)

	adminProtected.Get("/components/:id", pageHandler.RenderAdminComponent)
	adminProtected.Get("/components/:id/editor",
		middleware.OnlyHTMX(),
		pageHandler.RenderAdminComponentEditor,
	)

	app.Post("/login", loginHandler.Login)

	app.Post("/get-joy", pageHandler.GetJoy)
	app.Get("/:slug?", pageHandler.RenderPage)

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
