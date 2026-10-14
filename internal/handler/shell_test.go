package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/gofiber/fiber/v2"
)

func newShellTestApp(t *testing.T) *fiber.App {
	t.Helper()

	h, err := NewShellHandler(fstest.MapFS{
		"umono-shell.css": {Data: []byte("shell css")},
		"umono-root.js":   {Data: []byte("root js")},
	})
	if err != nil {
		t.Fatalf("expected NewShellHandler to succeed, got %v", err)
	}

	app := fiber.New()
	app.Get("/_umono/:file", h.Serve)
	return app
}

func TestShellServesFiles(t *testing.T) {
	app := newShellTestApp(t)

	cases := []struct {
		path        string
		wantStatus  int
		wantType    string
		wantCache   string
		wantBody    string
		wantETagSet bool
	}{
		{
			path:        "/_umono/umono-shell.css",
			wantStatus:  http.StatusOK,
			wantType:    "text/css; charset=utf-8",
			wantCache:   "no-cache",
			wantBody:    "shell css",
			wantETagSet: true,
		},
		{
			path:       "/_umono/umono-root.js",
			wantStatus: http.StatusOK,
			wantType:   "text/javascript; charset=utf-8",
			wantCache:  "no-cache",
			wantBody:   "root js",
		},
		{
			path:       "/_umono/other.css",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tc := range cases {
		req := httptest.NewRequest(http.MethodGet, tc.path, nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request %s: %v", tc.path, err)
		}

		if resp.StatusCode != tc.wantStatus {
			t.Fatalf("path %s: got status %d, want %d", tc.path, resp.StatusCode, tc.wantStatus)
		}

		if tc.wantStatus != http.StatusOK {
			continue
		}

		if got := resp.Header.Get("Content-Type"); got != tc.wantType {
			t.Fatalf("path %s: got Content-Type %q, want %q", tc.path, got, tc.wantType)
		}

		if got := resp.Header.Get("Cache-Control"); got != tc.wantCache {
			t.Fatalf("path %s: got Cache-Control %q, want %q", tc.path, got, tc.wantCache)
		}

		if tc.wantETagSet {
			if got := resp.Header.Get("ETag"); got == "" {
				t.Fatalf("path %s: got empty ETag, want non-empty", tc.path)
			}
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("path %s: read body: %v", tc.path, err)
		}
		if got := string(body); got != tc.wantBody {
			t.Fatalf("path %s: got body %q, want %q", tc.path, got, tc.wantBody)
		}
	}
}

func TestShellAnswersNotModified(t *testing.T) {
	app := newShellTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/_umono/umono-shell.css", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	etag := resp.Header.Get("ETag")
	if etag == "" {
		t.Fatal("expected non-empty ETag")
	}

	jsReq := httptest.NewRequest(http.MethodGet, "/_umono/umono-root.js", nil)
	jsResp, err := app.Test(jsReq)
	if err != nil {
		t.Fatal(err)
	}
	if jsResp.StatusCode != http.StatusOK {
		t.Fatalf("expected js status 200, got %d", jsResp.StatusCode)
	}
	jsETag := jsResp.Header.Get("ETag")
	if jsETag == "" {
		t.Fatal("expected non-empty ETag for js")
	}

	cases := []struct {
		name        string
		ifNoneMatch string
		wantStatus  int
	}{
		{"exact match", etag, http.StatusNotModified},
		{"weak prefix", "W/" + etag, http.StatusNotModified},
		{"star", "*", http.StatusNotModified},
		{"list with match", "other, " + etag, http.StatusNotModified},
		{"mismatching tag", jsETag, http.StatusOK},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/_umono/umono-shell.css", nil)
			req.Header.Set("If-None-Match", tc.ifNoneMatch)

			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tc.wantStatus {
				t.Fatalf("got status %d, want %d", resp.StatusCode, tc.wantStatus)
			}

			if tc.wantStatus == http.StatusNotModified {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatal(err)
				}
				if len(body) != 0 {
					t.Fatalf("got body %q, want empty", string(body))
				}
			}
		})
	}
}

func TestNewShellHandlerNeedsBothFiles(t *testing.T) {
	_, err := NewShellHandler(fstest.MapFS{
		"umono-shell.css": {Data: []byte("shell css")},
	})
	if err == nil {
		t.Fatal("got nil error, want non-nil error when umono-root.js is missing")
	}
}
