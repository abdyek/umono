package service

import (
	"bytes"
	"context"
	"testing"

	"github.com/umono-cms/themecore"
	"github.com/umono-cms/umono/internal/models"
)

func TestThemeServiceAsset(t *testing.T) {
	ctx := context.Background()
	reg := &fakeRegistry{}
	svc, _ := newThemeTestService(t, reg)
	blog := models.Contract{Org: "example", Name: "blog", Major: 1, Minor: 0}.Group()

	c, err := svc.Activate(ctx, "acme/minimal/0.1.0")
	if err != nil {
		t.Fatalf("Activate: %v", err)
	}

	b, ok := svc.Asset(c.Release, "theme.css")
	if !ok {
		t.Fatalf("Asset(release, theme.css) = false, want true")
	}
	if !bytes.Equal(b, themecore.FallbackCSS()) {
		t.Fatalf("Asset(release, theme.css) = %q, want %q", b, themecore.FallbackCSS())
	}

	if _, ok := svc.Asset(c.Release, "built.toml"); ok {
		t.Fatalf("Asset(release, built.toml) = true, want false")
	}

	if _, ok := svc.Asset(c.Release, "../built.toml"); ok {
		t.Fatalf("Asset(release, ../built.toml) = true, want false")
	}

	if _, ok := svc.Asset("000000000000", "theme.css"); ok {
		t.Fatalf("Asset(unknown-release, theme.css) = true, want false")
	}

	if err := svc.Deactivate(blog); err != nil {
		t.Fatalf("Deactivate: %v", err)
	}
	if _, ok := svc.Asset(c.Release, "theme.css"); ok {
		t.Fatalf("Asset(release, theme.css) after Deactivate = true, want false")
	}
}
