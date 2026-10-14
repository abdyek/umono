package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
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

// imageCall returns an IMAGE call with a media literal whose url is url.
func imageCall(url string) string {
	return "{{ IMAGE media = { url: " + strconv.Quote(url) + ", width: 1, height: 1, mime-type: " + strconv.Quote("image/png") + " } alt = " + strconv.Quote("") + " }}"
}

func nonceOf(t *testing.T, body string) string {
	t.Helper()
	const quote = "\""
	prefix := "<script nonce=" + quote
	i := strings.Index(body, prefix)
	if i < 0 {
		t.Fatalf("body = %q, want to contain %q", body, prefix)
	}
	rest := body[i+len(prefix):]
	j := strings.Index(rest, quote)
	if j < 0 {
		t.Fatalf("body after nonce prefix = %q, want to contain %q", rest, quote)
	}
	nonce := rest[:j]
	if nonce == "" {
		t.Fatalf("nonce is empty")
	}
	return nonce
}

func TestSitePageSendsContentSecurityPolicy(t *testing.T) {
	app, sps := newSitePageTestApp(t)

	if _, err := sps.Create(models.SitePage{
		Name:    "Gallery",
		Slug:    "gallery",
		Enabled: true,
		Content: imageCall("https://cdn.example.net/a.png"),
	}); err != nil {
		t.Fatalf("sps.Create() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/gallery", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() first error = %v", err)
	}
	if got, want := resp.StatusCode, http.StatusOK; got != want {
		t.Fatalf("resp.StatusCode = %d, want %d", got, want)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("io.ReadAll() first error = %v", err)
	}
	bodyStr := string(body)

	const quote = "\""
	nonce := nonceOf(t, bodyStr)

	wantScriptTag := "<script nonce=" + quote + nonce + quote + " src=" + quote + "/_umono/umono-root.js" + quote + " defer></script>"
	if got, want := bodyStr, wantScriptTag; !strings.Contains(bodyStr, want) {
		t.Fatalf("body = %q, want to contain %q", got, want)
	}

	wantCSP := "default-src 'none'; script-src 'nonce-" + nonce + "'; style-src 'self'; img-src 'self' https://cdn.example.net; font-src 'self'; base-uri 'none'; form-action 'none'; frame-ancestors 'none'"
	if got, want := resp.Header.Get("Content-Security-Policy"), wantCSP; got != want {
		t.Fatalf("resp.Header.Get(\"Content-Security-Policy\") = %q, want %q", got, want)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/gallery", nil)
	resp2, err := app.Test(req2)
	if err != nil {
		t.Fatalf("app.Test() second error = %v", err)
	}
	body2, err := io.ReadAll(resp2.Body)
	if err != nil {
		t.Fatalf("io.ReadAll() second error = %v", err)
	}
	nonce2 := nonceOf(t, string(body2))
	if got, want := nonce2, nonce; got == want {
		t.Fatalf("second nonce = %q, want different from %q", got, want)
	}

	req3 := httptest.NewRequest(http.MethodGet, "/missing", nil)
	resp3, err := app.Test(req3)
	if err != nil {
		t.Fatalf("app.Test() missing error = %v", err)
	}
	if got, want := resp3.StatusCode, http.StatusNotFound; got != want {
		t.Fatalf("resp3.StatusCode = %d, want %d", got, want)
	}
	body3, err := io.ReadAll(resp3.Body)
	if err != nil {
		t.Fatalf("io.ReadAll() missing error = %v", err)
	}
	body3Str := string(body3)
	nonce3 := nonceOf(t, body3Str)
	wantCSP3 := "default-src 'none'; script-src 'nonce-" + nonce3 + "'; style-src 'self'; img-src 'self'; font-src 'self'; base-uri 'none'; form-action 'none'; frame-ancestors 'none'"
	if got, want := resp3.Header.Get("Content-Security-Policy"), wantCSP3; got != want {
		t.Fatalf("resp3.Header.Get(\"Content-Security-Policy\") = %q, want %q", got, want)
	}
}

func TestSitePageETag(t *testing.T) {
	app, sps := newSitePageTestApp(t)

	const quote = "\""

	page, errs := sps.Create(models.SitePage{Name: "About", Slug: "about", Enabled: true, Content: "Hello"})
	if errs != nil {
		t.Fatalf("sps.Create() errs = %v", errs)
	}

	get := func(t *testing.T, app *fiber.App, path, ifNoneMatch string) (*http.Response, string) {
		t.Helper()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		if ifNoneMatch != "" {
			req.Header.Set("If-None-Match", ifNoneMatch)
		}
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

	resp1, _ := get(t, app, "/about", "")
	if got, want := resp1.StatusCode, http.StatusOK; got != want {
		t.Fatalf("resp1.StatusCode = %d, want %d", got, want)
	}
	if got, want := resp1.Header.Get("Cache-Control"), "no-cache"; got != want {
		t.Fatalf("resp1.Header.Get(\"Cache-Control\") = %q, want %q", got, want)
	}
	if got, want := resp1.Header.Get("Content-Type"), "text/html; charset=utf-8"; got != want {
		t.Fatalf("resp1.Header.Get(\"Content-Type\") = %q, want %q", got, want)
	}
	etag := resp1.Header.Get("ETag")
	if !strings.HasPrefix(etag, "W/"+quote) {
		t.Fatalf("resp1.Header.Get(\"ETag\") = %q, want to start with %q", etag, "W/"+quote)
	}
	if !strings.HasSuffix(etag, quote) {
		t.Fatalf("resp1.Header.Get(\"ETag\") = %q, want to end with %q", etag, quote)
	}

	resp2, _ := get(t, app, "/about", "")
	if got, want := resp2.Header.Get("ETag"), etag; got != want {
		t.Fatalf("resp2.Header.Get(\"ETag\") = %q, want %q", got, want)
	}

	resp3, body3 := get(t, app, "/about", etag)
	if got, want := resp3.StatusCode, http.StatusNotModified; got != want {
		t.Fatalf("resp3.StatusCode = %d, want %d", got, want)
	}
	if got, want := body3, ""; got != want {
		t.Fatalf("resp3 body = %q, want %q", got, want)
	}
	if got, want := resp3.Header.Get("Content-Security-Policy"), ""; got != want {
		t.Fatalf("resp3.Header.Get(\"Content-Security-Policy\") = %q, want %q", got, want)
	}
	if got, want := resp3.Header.Get("ETag"), etag; got != want {
		t.Fatalf("resp3.Header.Get(\"ETag\") = %q, want %q", got, want)
	}

	resp4, _ := get(t, app, "/about", "W/"+quote+"other"+quote)
	if got, want := resp4.StatusCode, http.StatusOK; got != want {
		t.Fatalf("resp4.StatusCode = %d, want %d", got, want)
	}

	page.Content = "Changed"
	if _, err := sps.Update(page); err != nil {
		t.Fatalf("sps.Update() err = %v", err)
	}

	resp5, _ := get(t, app, "/about", "")
	if got, want := resp5.Header.Get("ETag"), etag; got == want {
		t.Fatalf("resp5.Header.Get(\"ETag\") = %q, want different from %q", got, want)
	}

	resp6, _ := get(t, app, "/about", etag)
	if got, want := resp6.StatusCode, http.StatusOK; got != want {
		t.Fatalf("resp6.StatusCode = %d, want %d", got, want)
	}

	resp7, _ := get(t, app, "/missing", "")
	if got, want := resp7.StatusCode, http.StatusNotFound; got != want {
		t.Fatalf("resp7.StatusCode = %d, want %d", got, want)
	}
	if got, want := resp7.Header.Get("ETag"), ""; got != want {
		t.Fatalf("resp7.Header.Get(\"ETag\") = %q, want %q", got, want)
	}

	resp8, _ := get(t, app, "/missing", "*")
	if got, want := resp8.StatusCode, http.StatusNotFound; got != want {
		t.Fatalf("resp8.StatusCode = %d, want %d", got, want)
	}
}
