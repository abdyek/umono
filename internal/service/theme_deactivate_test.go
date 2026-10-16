package service

import (
	"context"
	"testing"

	"github.com/umono-cms/umono/internal/models"
)

func TestThemeServiceDeactivate(t *testing.T) {
	ctx := context.Background()
	reg := &fakeRegistry{}
	svc, db := newThemeTestService(t, reg)

	if err := db.Create(&models.Option{
		Key:   "active_themes",
		Value: `[{"path":"jane/folio/1.0.0","contract_id":"umono/page@v0.1"}]`,
	}).Error; err != nil {
		t.Fatalf("create option: %v", err)
	}

	if _, err := svc.Activate(ctx, "acme/minimal/0.1.0"); err != nil {
		t.Fatalf("Activate: %v", err)
	}

	blog := models.Contract{Org: "example", Name: "blog", Major: 1, Minor: 0}.Group()

	if err := svc.Deactivate(blog); err != nil {
		t.Fatalf("Deactivate: %v", err)
	}

	wantThemes := `[{"path":"jane/folio/1.0.0","contract_id":"umono/page@v0.1"}]`
	if got := activeThemes(t, db); got != wantThemes {
		t.Fatalf("activeThemes = %q, want %q", got, wantThemes)
	}

	if _, ok := svc.Loaded(blog); ok {
		t.Fatalf("svc.Loaded(blog) = true, want false")
	}

	if n := countContracts(t, db); n != 1 {
		t.Fatalf("countContracts = %d, want 1", n)
	}
}

func TestThemeServiceDeactivateWithoutActiveTheme(t *testing.T) {
	reg := &fakeRegistry{}
	svc, db := newThemeTestService(t, reg)

	blog := models.Contract{Org: "example", Name: "blog", Major: 1, Minor: 0}.Group()

	if err := svc.Deactivate(blog); err != nil {
		t.Fatalf("Deactivate (no option): %v", err)
	}
	if got := activeThemes(t, db); got != "" {
		t.Fatalf("activeThemes = %q, want %q", got, "")
	}

	if err := db.Create(&models.Option{
		Key:   "active_themes",
		Value: `[{"path":"jane/folio/1.0.0","contract_id":"umono/page@v0.1"}]`,
	}).Error; err != nil {
		t.Fatalf("create option: %v", err)
	}

	wantThemes := `[{"path":"jane/folio/1.0.0","contract_id":"umono/page@v0.1"}]`
	if err := svc.Deactivate(blog); err != nil {
		t.Fatalf("Deactivate (with other group): %v", err)
	}
	if got := activeThemes(t, db); got != wantThemes {
		t.Fatalf("activeThemes = %q, want %q", got, wantThemes)
	}
}

func TestThemeServiceNotifiesRenderInputChange(t *testing.T) {
	ctx := context.Background()
	reg := &fakeRegistry{}
	svc, _ := newThemeTestService(t, reg)

	calls := 0
	svc.OnRenderInputChange(func(ctx context.Context) {
		calls++
	})

	if _, err := svc.Activate(ctx, "acme/minimal/0.1.0"); err != nil {
		t.Fatalf("Activate: %v", err)
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}

	if _, err := svc.Activate(ctx, "../acme/minimal/0.1.0"); err == nil {
		t.Fatalf("Activate error = nil, want non-nil")
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}

	blog := models.Contract{Org: "example", Name: "blog", Major: 1, Minor: 0}.Group()
	if err := svc.Deactivate(blog); err != nil {
		t.Fatalf("Deactivate: %v", err)
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
}
