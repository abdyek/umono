package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/umono-cms/umono"
)

// TestStaticCacheControl checks that a release caches /static for a year,
// while a local build, whose URLs always carry ?v=dev, revalidates it.
func TestStaticCacheControl(t *testing.T) {
	cases := []struct {
		version string
		want    string
	}{
		{version: "dev", want: "no-cache"},
		{version: "v0.9.0", want: "public, max-age=31536000, immutable"},
	}

	for _, tc := range cases {
		t.Run(tc.version, func(t *testing.T) {
			old := umono.Version
			umono.Version = tc.version
			t.Cleanup(func() { umono.Version = old })

			app, _ := newTestApp(t)
			res, err := app.Test(httptest.NewRequest(http.MethodGet, "/static/js/admin.js?v="+tc.version, nil), -1)
			if err != nil {
				t.Fatalf("test: %v", err)
			}
			if res.StatusCode != http.StatusOK {
				t.Fatalf("status = %d", res.StatusCode)
			}
			if got := res.Header.Get("Cache-Control"); got != tc.want {
				t.Errorf("Cache-Control = %q, want %q", got, tc.want)
			}
		})
	}
}
