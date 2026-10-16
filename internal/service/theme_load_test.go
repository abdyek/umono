package service

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/repository"
	"gorm.io/gorm"
)

func restart(t *testing.T, db *gorm.DB, themesDir string) (*ThemeService, *fakeRegistry) {
	t.Helper()
	reg := &fakeRegistry{}
	return NewThemeService(themesDir, t.TempDir(), NewContractService(repository.NewContractRepository(db)), repository.NewThemeRepository(db), reg), reg
}

func activated(t *testing.T) (db *gorm.DB, themesDir string, release string) {
	t.Helper()
	ctx := context.Background()
	svc, db, themesDir, _ := newPendingTestService(t, &fakeRegistry{}, nil)
	c, err := svc.Activate(ctx, "acme/minimal/0.1.0")
	if err != nil {
		t.Fatalf("Activate: %v", err)
	}
	return db, themesDir, c.Release
}

func breakLicense(t *testing.T, themesDir string) {
	t.Helper()
	p := filepath.Join(themesDir, "acme", "minimal", "0.1.0", "LICENSE")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read LICENSE: %v", err)
	}
	b = append(b, "broken\n"...)
	if err := os.WriteFile(p, b, 0o644); err != nil {
		t.Fatalf("write LICENSE: %v", err)
	}
}

func TestLoadActive(t *testing.T) {
	ctx := context.Background()
	blog := models.Contract{Org: "example", Name: "blog", Major: 1, Minor: 0}.Group()
	path := "acme/minimal/0.1.0"

	db, themesDir, release := activated(t)
	svc, reg := restart(t, db, themesDir)

	if err := svc.LoadActive(ctx); err != nil {
		t.Fatalf("LoadActive: %v", err)
	}
	l, ok := svc.Loaded(blog)
	if !ok {
		t.Fatalf("svc.Loaded(blog) = false, want true")
	}
	if l.Release != release {
		t.Fatalf("l.Release = %q, want %q", l.Release, release)
	}
	if l.Path != path {
		t.Fatalf("l.Path = %q, want %q", l.Path, path)
	}
	if reg.calls != 0 {
		t.Fatalf("reg.calls = %d, want 0", reg.calls)
	}
	_, broken := svc.Broken(blog)
	if broken {
		t.Fatalf("svc.Broken(blog) = true, want false")
	}
}

func TestLoadActiveMarksAChangedThemeBroken(t *testing.T) {
	ctx := context.Background()
	blog := models.Contract{Org: "example", Name: "blog", Major: 1, Minor: 0}.Group()
	path := "acme/minimal/0.1.0"

	db, themesDir, _ := activated(t)
	breakLicense(t, themesDir)
	svc, _ := restart(t, db, themesDir)

	if err := svc.LoadActive(ctx); err != nil {
		t.Fatalf("LoadActive: %v", err)
	}
	if _, ok := svc.Loaded(blog); ok {
		t.Fatalf("svc.Loaded(blog) = true, want false")
	}
	b, broken := svc.Broken(blog)
	if !broken {
		t.Fatalf("svc.Broken(blog) = false, want true")
	}
	if b.Path != path {
		t.Fatalf("b.Path = %q, want %q", b.Path, path)
	}
	if b.Err == nil {
		t.Fatalf("b.Err = nil, want non-nil")
	}
	if got := activeThemes(t, db); got != `[{"path":"acme/minimal/0.1.0","contract_id":"example/blog@v1.0"}]` {
		t.Fatalf("activeThemes = %q, want %q", got, `[{"path":"acme/minimal/0.1.0","contract_id":"example/blog@v1.0"}]`)
	}
}

func TestLoadActiveChecksTheStoredContract(t *testing.T) {
	ctx := context.Background()
	blog := models.Contract{Org: "example", Name: "blog", Major: 1, Minor: 0}.Group()

	db, themesDir, _ := activated(t)
	if err := db.Model(&models.Contract{}).Where("org = ? AND name = ?", "example", "blog").Update("checksum", "0000").Error; err != nil {
		t.Fatalf("update checksum: %v", err)
	}
	svc, _ := restart(t, db, themesDir)
	if err := svc.LoadActive(ctx); err != nil {
		t.Fatalf("LoadActive: %v", err)
	}
	b, broken := svc.Broken(blog)
	if !broken {
		t.Fatalf("svc.Broken(blog) = false, want true")
	}
	if !errors.Is(b.Err, ErrContractChanged) {
		t.Fatalf("b.Err = %v, want ErrContractChanged", b.Err)
	}
}

func TestRetry(t *testing.T) {
	ctx := context.Background()
	blog := models.Contract{Org: "example", Name: "blog", Major: 1, Minor: 0}.Group()

	db, themesDir, release := activated(t)
	original := readThemeFile(t, themesDir, "LICENSE")
	breakLicense(t, themesDir)
	svc, _ := restart(t, db, themesDir)
	if err := svc.LoadActive(ctx); err != nil {
		t.Fatalf("LoadActive: %v", err)
	}

	calls := 0
	svc.OnRenderInputChange(func(ctx context.Context) {
		calls++
	})

	p := filepath.Join(themesDir, "acme", "minimal", "0.1.0", "LICENSE")
	if err := os.WriteFile(p, original, 0o644); err != nil {
		t.Fatalf("write LICENSE: %v", err)
	}

	if err := svc.Retry(ctx, blog); err != nil {
		t.Fatalf("Retry: %v", err)
	}
	l, ok := svc.Loaded(blog)
	if !ok {
		t.Fatalf("svc.Loaded(blog) = false, want true")
	}
	if l.Release != release {
		t.Fatalf("l.Release = %q, want %q", l.Release, release)
	}
	if _, broken := svc.Broken(blog); broken {
		t.Fatalf("svc.Broken(blog) = true, want false")
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}
}

func TestRetryStillBroken(t *testing.T) {
	ctx := context.Background()
	blog := models.Contract{Org: "example", Name: "blog", Major: 1, Minor: 0}.Group()

	db, themesDir, _ := activated(t)
	breakLicense(t, themesDir)
	svc, _ := restart(t, db, themesDir)
	if err := svc.LoadActive(ctx); err != nil {
		t.Fatalf("LoadActive: %v", err)
	}

	calls := 0
	svc.OnRenderInputChange(func(ctx context.Context) {
		calls++
	})

	if err := svc.Retry(ctx, blog); err == nil {
		t.Fatalf("Retry error = nil, want non-nil")
	}
	b, broken := svc.Broken(blog)
	if !broken {
		t.Fatalf("svc.Broken(blog) = false, want true")
	}
	if b.Err == nil {
		t.Fatalf("b.Err = nil, want non-nil")
	}
	if _, ok := svc.Loaded(blog); ok {
		t.Fatalf("svc.Loaded(blog) = true, want false")
	}
	if calls != 0 {
		t.Fatalf("calls = %d, want 0", calls)
	}
}

func TestRetryWithoutActiveTheme(t *testing.T) {
	ctx := context.Background()
	blog := models.Contract{Org: "example", Name: "blog", Major: 1, Minor: 0}.Group()

	svc, _, _, _ := newPendingTestService(t, &fakeRegistry{}, nil)

	if err := svc.Retry(ctx, blog); !errors.Is(err, ErrNoActiveTheme) {
		t.Fatalf("Retry: err = %v, want ErrNoActiveTheme", err)
	}
}

func TestDeactivateABrokenTheme(t *testing.T) {
	ctx := context.Background()
	blog := models.Contract{Org: "example", Name: "blog", Major: 1, Minor: 0}.Group()

	db, themesDir, _ := activated(t)
	breakLicense(t, themesDir)
	svc, _ := restart(t, db, themesDir)
	if err := svc.LoadActive(ctx); err != nil {
		t.Fatalf("LoadActive: %v", err)
	}

	if err := svc.Deactivate(blog); err != nil {
		t.Fatalf("Deactivate: %v", err)
	}
	if _, broken := svc.Broken(blog); broken {
		t.Fatalf("svc.Broken(blog) = true, want false")
	}
	if got := activeThemes(t, db); got != "[]" {
		t.Fatalf("activeThemes = %q, want %q", got, "[]")
	}
}
