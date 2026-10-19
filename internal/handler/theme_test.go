package handler

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/umono-cms/umono/internal/theme"
)

type fakeThemeAssets map[string][]byte

func (f fakeThemeAssets) Asset(release, name string) ([]byte, bool) {
	b, ok := f[release+"/"+name]
	return b, ok
}

// DevAsset serves the entries under dev/ as the theme in development; an
// entry under dev-error/ fails its check.
func (f fakeThemeAssets) DevAsset(name string) ([]byte, bool, error) {
	if _, ok := f["dev-error/"+name]; ok {
		return nil, false, errors.New("check failed")
	}
	b, ok := f["dev/"+name]
	return b, ok, nil
}

func newThemeTestApp(t *testing.T) *fiber.App {
	t.Helper()

	h := NewThemeHandler(theme.Fallback{
		Release: "fallback-abc123abc123",
		CSS:     []byte("umono-root{}"),
	}, fakeThemeAssets{
		"abc123abc123/theme.css":     []byte("a{}"),
		"abc123abc123/fonts/a.woff2": []byte("wOF2"),
		"dev/fonts/a.woff2":          []byte("wOF2"),
		"dev-error/theme.css":        []byte("x"),
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
		"/_theme/abc123abc123/other.woff2",
		"/_theme/000000000000/theme.css",
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

func TestThemeServesActiveThemeAssets(t *testing.T) {
	app := newThemeTestApp(t)

	cases := []struct {
		path        string
		contentType string
		body        string
	}{
		{"/_theme/abc123abc123/theme.css", "text/css; charset=utf-8", "a{}"},
		{"/_theme/abc123abc123/fonts/a.woff2", "font/woff2", "wOF2"},
	}

	for _, c := range cases {
		c := c
		t.Run(c.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, c.path, nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("request %s: %v", c.path, err)
			}

			if got, want := resp.StatusCode, http.StatusOK; got != want {
				t.Fatalf("path %s: got status %d, want %d", c.path, got, want)
			}

			if got, want := resp.Header.Get("Content-Type"), c.contentType; got != want {
				t.Fatalf("path %s: got Content-Type %q, want %q", c.path, got, want)
			}

			if got, want := resp.Header.Get("Cache-Control"), "public, max-age=31536000, immutable"; got != want {
				t.Fatalf("path %s: got Cache-Control %q, want %q", c.path, got, want)
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("path %s: read body: %v", c.path, err)
			}
			if got, want := string(body), c.body; got != want {
				t.Fatalf("path %s: got body %q, want %q", c.path, got, want)
			}
		})
	}
}

func TestThemeServesDevAssetsWithoutCache(t *testing.T) {
	app := newThemeTestApp(t)

	path := "/_theme/dev/fonts/a.woff2"
	req := httptest.NewRequest(http.MethodGet, path, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request %s: %v", path, err)
	}

	if got, want := resp.StatusCode, http.StatusOK; got != want {
		t.Fatalf("path %s: got status %d, want %d", path, got, want)
	}

	if got, want := resp.Header.Get("Cache-Control"), "no-store"; got != want {
		t.Fatalf("path %s: got Cache-Control %q, want %q", path, got, want)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("path %s: read body: %v", path, err)
	}
	if got, want := string(body), "wOF2"; got != want {
		t.Fatalf("path %s: got body %q, want %q", path, got, want)
	}
}

func TestThemeDevAssetErrors(t *testing.T) {
	app := newThemeTestApp(t)

	cases := []struct {
		path   string
		status int
	}{
		{"/_theme/dev/theme.css", http.StatusInternalServerError},
		{"/_theme/dev/missing.png", http.StatusNotFound},
	}

	for _, c := range cases {
		c := c
		t.Run(c.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, c.path, nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("request %s: %v", c.path, err)
			}

			if got, want := resp.StatusCode, c.status; got != want {
				t.Fatalf("path %s: got status %d, want %d", c.path, got, want)
			}

			if got, want := resp.Header.Get("Cache-Control"), "no-store"; got != want {
				t.Fatalf("path %s: got Cache-Control %q, want %q", c.path, got, want)
			}
		})
	}
}
