package handler

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
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

// pageContract reads the page-block.txtar fixture and returns the text
// between the "-- contract.toml --" line and the "-- page.comp --" marker.
func pageContract(t *testing.T) string {
	t.Helper()
	const path = "../themeapi/testdata/dom/scenarios/contracts/umono/page/v0.1/page-block.txtar"
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) error = %v", path, err)
	}
	text := string(b)
	const start = "-- contract.toml --\n"
	const end = "-- page.comp --"
	i := strings.Index(text, start)
	if i < 0 {
		t.Fatalf("start marker %q not found in %s", start, path)
	}
	j := strings.Index(text, end)
	if j < 0 {
		t.Fatalf("end marker %q not found in %s", end, path)
	}
	return text[i+len(start) : j]
}

func newBindingTestApp(t *testing.T) (*fiber.App, *gorm.DB, models.Contract) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	if err := db.AutoMigrate(
		&models.SitePage{},
		&models.Component{},
		&models.PageDiagnostic{},
		&models.Contract{},
		&models.Option{},
	); err != nil {
		t.Fatalf("db.AutoMigrate() error = %v", err)
	}

	cc, err := service.NewContentCompiler(nil)
	if err != nil {
		t.Fatalf("service.NewContentCompiler() error = %v", err)
	}

	optionRepo := repository.NewOptionRepository(db)
	sps := service.NewSitePageService(repository.NewSitePageRepository(db), optionRepo, cc)
	sps.SetPageDiagnosticRepository(repository.NewPageDiagnosticRepository(db))

	cs := service.NewComponentService(repository.NewComponentRepository(db))
	cs.SetContentCompiler(cc)

	ts := service.NewThemeService(
		writeHandlerTheme(t),
		t.TempDir(),
		service.NewContractService(repository.NewContractRepository(db)),
		repository.NewThemeRepository(db),
		&fakeThemeRegistry{},
	)
	sps.SetPageThemes(ts)

	rec := models.Contract{
		Org:      "umono",
		Name:     "page",
		Major:    0,
		Minor:    1,
		Source:   pageContract(t),
		Checksum: "x",
	}
	if err := db.Create(&rec).Error; err != nil {
		t.Fatalf("db.Create(contract) error = %v", err)
	}

	key := models.NotFoundSystemKey
	if err := db.Create(&models.SitePage{
		Name:      "",
		Slug:      models.NotFoundSlug,
		Enabled:   true,
		SystemKey: &key,
	}).Error; err != nil {
		t.Fatalf("db.Create(not-found page) error = %v", err)
	}

	h := NewSitePageHandler(sps, cs, ts)

	engine := html.NewFileSystem(http.FS(umono.Views()), ".html")
	engine.AddFunc("t", view.Translate)

	app := fiber.New(fiber.Config{Views: engine})
	app.Post("/admin/site-pages", h.Create)
	app.Get("/admin/site-pages/layout-info", h.LayoutInfo)
	app.Post("/admin/options/not-found-page-option", h.SaveNotFoundPage)

	return app, db, rec
}

func formPost(t *testing.T, app *fiber.App, path string, form url.Values) (*http.Response, string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("io.ReadAll() error = %v", err)
	}
	return resp, string(body)
}

const boundPageContent = "{{ UMONO_LAYOUT site-name = \"S\" home-url = \"/\" menu-label = \"M\" menu = [] content = BODY footer = FOOT }}\n\n~ BODY\nHi\n\n~ FOOT\nBye\n"

func TestCreatePageWithALayout(t *testing.T) {
	app, db, rec := newBindingTestApp(t)

	layout := fmt.Sprintf("%d:page", rec.ID)
	resp, _ := formPost(t, app, "/admin/site-pages", url.Values{
		"name":    {"Home"},
		"slug":    {"home"},
		"enabled": {"1"},
		"content": {boundPageContent},
		"layout":  {layout},
	})
	if got, want := resp.StatusCode, http.StatusOK; got != want {
		t.Fatalf("resp.StatusCode = %d, want %d", got, want)
	}

	var page models.SitePage
	if err := db.Where("slug = ?", "home").First(&page).Error; err != nil {
		t.Fatalf("db.Where(slug=home).First() error = %v", err)
	}
	if page.ContractID == nil || *page.ContractID != rec.ID {
		t.Fatalf("page.ContractID = %v, want %d", page.ContractID, rec.ID)
	}
	if page.LayoutID == nil || *page.LayoutID != "page" {
		t.Fatalf("page.LayoutID = %v, want %q", page.LayoutID, "page")
	}
}

func TestCreatePageShowsTheBindingError(t *testing.T) {
	app, db, rec := newBindingTestApp(t)

	layout := fmt.Sprintf("%d:page", rec.ID)
	_, body := formPost(t, app, "/admin/site-pages", url.Values{
		"name":    {"Home"},
		"slug":    {"home"},
		"enabled": {"1"},
		"content": {"# Not a layout call"},
		"layout":  {layout},
	})

	if !strings.Contains(body, "admin.pages.layout.errors.root") {
		t.Fatalf("body = %q, want to contain %q", body, "admin.pages.layout.errors.root")
	}
	if !strings.Contains(body, "{{ UMONO_LAYOUT site-name") {
		t.Fatalf("body = %q, want to contain %q", body, "{{ UMONO_LAYOUT site-name")
	}

	var count int64
	if err := db.Model(&models.SitePage{}).Where("slug = ?", "home").Count(&count).Error; err != nil {
		t.Fatalf("db.Count() error = %v", err)
	}
	if got, want := count, int64(0); got != want {
		t.Fatalf("count of pages with slug \"home\" = %d, want %d", got, want)
	}
}

func TestLayoutInfoListsTheParameters(t *testing.T) {
	app, _, rec := newBindingTestApp(t)

	layout := fmt.Sprintf("%d:page", rec.ID)
	req := httptest.NewRequest(http.MethodGet, "/admin/site-pages/layout-info?layout="+url.QueryEscape(layout), nil)
	req.Header.Set("HX-Request", "true")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if got, want := resp.StatusCode, http.StatusOK; got != want {
		t.Fatalf("resp.StatusCode = %d, want %d", got, want)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("io.ReadAll() error = %v", err)
	}
	bodyStr := string(body)

	for _, want := range []string{"site-name", "footer", "admin.pages.layout.required"} {
		if !strings.Contains(bodyStr, want) {
			t.Fatalf("body = %q, want to contain %q", bodyStr, want)
		}
	}
}

func TestNotFoundPageShowsTheBindingError(t *testing.T) {
	app, db, rec := newBindingTestApp(t)

	layout := fmt.Sprintf("%d:page", rec.ID)
	resp, body := formPost(t, app, "/admin/options/not-found-page-option", url.Values{
		"title":   {"Lost"},
		"content": {"# Not a layout call"},
		"layout":  {layout},
	})

	if got, want := resp.Header.Get("HX-Retarget"), "#not-found-layout-error"; got != want {
		t.Fatalf("resp.Header.Get(\"HX-Retarget\") = %q, want %q", got, want)
	}
	if !strings.Contains(body, "admin.pages.layout.errors.root") {
		t.Fatalf("body = %q, want to contain %q", body, "admin.pages.layout.errors.root")
	}

	var page models.SitePage
	if err := db.Where("system_key = ?", models.NotFoundSystemKey).First(&page).Error; err != nil {
		t.Fatalf("db.Where(system_key=not_found).First() error = %v", err)
	}
	if page.ContractID != nil {
		t.Fatalf("page.ContractID = %v, want nil", page.ContractID)
	}
}

func TestNotFoundPageSavesItsLayout(t *testing.T) {
	app, db, rec := newBindingTestApp(t)

	layout := fmt.Sprintf("%d:page", rec.ID)
	resp, _ := formPost(t, app, "/admin/options/not-found-page-option", url.Values{
		"title":   {"Lost"},
		"content": {boundPageContent},
		"layout":  {layout},
	})
	if got, want := resp.StatusCode, http.StatusOK; got != want {
		t.Fatalf("resp.StatusCode = %d, want %d", got, want)
	}

	var page models.SitePage
	if err := db.Where("system_key = ?", models.NotFoundSystemKey).First(&page).Error; err != nil {
		t.Fatalf("db.Where(system_key=not_found).First() error = %v", err)
	}
	if page.ContractID == nil || *page.ContractID != rec.ID {
		t.Fatalf("page.ContractID = %v, want %d", page.ContractID, rec.ID)
	}
}
