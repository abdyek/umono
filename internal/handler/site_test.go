package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/umono-cms/umono"
	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/repository"
	"github.com/umono-cms/umono/internal/service"
	"github.com/umono-cms/umono/internal/view"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newSitePageTestApp(t *testing.T) (*fiber.App, *service.SitePageService) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	if err := db.AutoMigrate(&models.SitePage{}, &models.Option{}, &models.Component{}); err != nil {
		t.Fatalf("db.AutoMigrate() error = %v", err)
	}

	cc, err := service.NewContentCompiler(nil)
	if err != nil {
		t.Fatalf("service.NewContentCompiler() error = %v", err)
	}

	optionRepo := repository.NewOptionRepository(db)
	sps := service.NewSitePageService(repository.NewSitePageRepository(db), optionRepo, cc)
	h := NewSiteHandler(sps, service.NewOptionService(optionRepo, nil), "fallback-abc123abc123")

	engine := html.NewFileSystem(http.FS(umono.Views()), ".html")
	engine.AddFunc("t", view.Translate)

	app := fiber.New(fiber.Config{Views: engine})
	app.Get("/:slug?", h.RenderSitePage)

	return app, sps
}

func TestSitePageLinksThemeStylesheet(t *testing.T) {
	app, sps := newSitePageTestApp(t)

	if _, err := sps.Create(models.SitePage{Name: "About", Slug: "about", Enabled: true, Content: "Hello"}); err != nil {
		t.Fatalf("sps.Create() error = %v", err)
	}

	cases := []struct {
		name   string
		path   string
		status int
	}{
		{"existing page", "/about", http.StatusOK},
		{"missing page", "/missing", http.StatusNotFound},
	}

	wantStylesheet := `<link rel="stylesheet" href="/_theme/fallback-abc123abc123/theme.css">`
	notWant := "default.css"

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("app.Test() error = %v", err)
			}

			if got, want := resp.StatusCode, tc.status; got != want {
				t.Fatalf("resp.StatusCode = %d, want %d", got, want)
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("io.ReadAll() error = %v", err)
			}
			bodyStr := string(body)

			if got, want := bodyStr, wantStylesheet; !strings.Contains(bodyStr, want) {
				t.Fatalf("body = %q, want to contain %q", got, want)
			}
			if got, want := bodyStr, notWant; strings.Contains(bodyStr, want) {
				t.Fatalf("body = %q, want to not contain %q", got, want)
			}
		})
	}
}
