package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/repository"
	"github.com/umono-cms/umono/internal/service"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newSitemapTestApp(t *testing.T) (*fiber.App, *gorm.DB, *service.OptionService) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	if err := db.AutoMigrate(&models.SitePage{}, &models.Option{}); err != nil {
		t.Fatalf("db.AutoMigrate() error = %v", err)
	}

	optionRepo := repository.NewOptionRepository(db)
	optionService := service.NewOptionService(optionRepo, nil)
	h := NewSiteHandler(service.NewSitePageService(repository.NewSitePageRepository(db), optionRepo, nil), optionService)

	app := fiber.New()
	app.Get("/sitemap.xml", h.RenderSitemap)

	return app, db, optionService
}

func TestSitemapNotFoundWithoutSiteURL(t *testing.T) {
	app, db, _ := newSitemapTestApp(t)

	if err := db.Create(&models.SitePage{Name: "About", Slug: "about", Enabled: true}).Error; err != nil {
		t.Fatalf("db.Create() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}

	if got, want := resp.StatusCode, http.StatusNotFound; got != want {
		t.Fatalf("resp.StatusCode = %d, want %d", got, want)
	}
}

func TestSitemapServesXML(t *testing.T) {
	app, db, optionService := newSitemapTestApp(t)

	if err := optionService.SaveSiteURL("https://example.com"); err != nil {
		t.Fatalf("optionService.SaveSiteURL() error = %v", err)
	}
	if err := db.Create(&models.SitePage{Name: "About", Slug: "about", Enabled: true}).Error; err != nil {
		t.Fatalf("db.Create() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}

	if got, want := resp.StatusCode, http.StatusOK; got != want {
		t.Fatalf("resp.StatusCode = %d, want %d", got, want)
	}
	if got, want := resp.Header.Get("Content-Type"), "application/xml; charset=utf-8"; got != want {
		t.Fatalf("resp.Header.Get(\"Content-Type\") = %q, want %q", got, want)
	}
	if got, want := resp.Header.Get("X-Content-Type-Options"), "nosniff"; got != want {
		t.Fatalf("resp.Header.Get(\"X-Content-Type-Options\") = %q, want %q", got, want)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("io.ReadAll() error = %v", err)
	}
	if got, want := string(body), "<loc>https://example.com/about</loc>"; !strings.Contains(string(body), want) {
		t.Fatalf("body = %q, want to contain %q", got, want)
	}
}
