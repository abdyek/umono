package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"

	"github.com/umono-cms/umono/internal/theme"
)

func newThemeTestApp(t *testing.T) *fiber.App {
	t.Helper()

	h := NewThemeHandler(theme.Fallback{
		Release: "fallback-abc123abc123",
		CSS:     []byte("umono-root{}"),
	})

	app := fiber.New()
	app.Get("/_theme/:release/*", h.Serve)
	return app
}

func TestThemeServesFallbackCSS(t *testing.T) {
	app := newThemeTestApp(t)

	path := "/_theme/fallback-abc123abc123/theme.css"
	req := httptest.NewRequest(http.MethodGet, path, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request %s: %v", path, err)
	}

	if got, want := resp.StatusCode, http.StatusOK; got != want {
		t.Fatalf("path %s: got status %d, want %d", path, got, want)
	}

	if got, want := resp.Header.Get("Content-Type"), "text/css; charset=utf-8"; got != want {
		t.Fatalf("path %s: got Content-Type %q, want %q", path, got, want)
	}

	if got, want := resp.Header.Get("Cache-Control"), "public, max-age=31536000, immutable"; got != want {
		t.Fatalf("path %s: got Cache-Control %q, want %q", path, got, want)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("path %s: read body: %v", path, err)
	}
	if got, want := string(body), "umono-root{}"; got != want {
		t.Fatalf("path %s: got body %q, want %q", path, got, want)
	}
}

func TestThemeNotFound(t *testing.T) {
	app := newThemeTestApp(t)

	cases := []string{
		"/_theme/fallback-abc123abc123/other.css",
		"/_theme/fallback-abc123abc123/fonts/theme.css",
		"/_theme/fallback-000000000000/theme.css",
		"/_theme/abc123abc123/theme.css",
	}

	for _, path := range cases {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request %s: %v", path, err)
		}

		if got, want := resp.StatusCode, http.StatusNotFound; got != want {
			t.Fatalf("path %s: got status %d, want %d", path, got, want)
		}
	}
}
