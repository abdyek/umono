package service

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/umono-cms/compono"
	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func bindPage(t *testing.T, db *gorm.DB, contractID uint, layout string) models.SitePage {
	t.Helper()
	page := models.SitePage{
		Name:       "P",
		Slug:       "p",
		Enabled:    true,
		Content:    "",
		ContractID: &contractID,
		LayoutID:   &layout,
	}
	if err := db.Create(&page).Error; err != nil {
		t.Fatalf("create page: %v", err)
	}
	return page
}

func TestPageRenderWithoutBinding(t *testing.T) {
	svc, _ := newThemeTestService(t, &fakeRegistry{})

	r, err := svc.PageRender(models.SitePage{Name: "A", Slug: "a"})
	if err != nil {
		t.Fatalf("PageRender: %v", err)
	}
	if r.LayoutID != "" {
		t.Fatalf("r.LayoutID = %q, want %q", r.LayoutID, "")
	}
	if r.Release != "" {
		t.Fatalf("r.Release = %q, want %q", r.Release, "")
	}
	if r.Attributes != nil {
		t.Fatalf("r.Attributes = %v, want nil", r.Attributes)
	}
	if got := len(r.Options()); got != 0 {
		t.Fatalf("len(r.Options()) = %d, want 0", got)
	}
	if r.RootClass() != "" {
		t.Fatalf("r.RootClass() = %q, want %q", r.RootClass(), "")
	}
}

func TestPageRenderWithoutActiveTheme(t *testing.T) {
	svc, db := newThemeTestService(t, &fakeRegistry{})
	rec := storeBlogContract(t, db)
	page := bindPage(t, db, rec.ID, "post")

	r, err := svc.PageRender(page)
	if err != nil {
		t.Fatalf("PageRender: %v", err)
	}
	if r.LayoutID != "post" {
		t.Fatalf("r.LayoutID = %q, want %q", r.LayoutID, "post")
	}
	if !strings.Contains(r.Layout, "{{ POST_BODY body = body }}") {
		t.Fatalf("r.Layout does not contain %q", "{{ POST_BODY body = body }}")
	}
	if len(r.Components) != 1 {
		t.Fatalf("len(r.Components) = %d, want 1", len(r.Components))
	}
	if r.Components[0].Name != "POST_BODY" {
		t.Fatalf("r.Components[0].Name = %q, want %q", r.Components[0].Name, "POST_BODY")
	}
	if r.Release != "" {
		t.Fatalf("r.Release = %q, want %q", r.Release, "")
	}
	if r.Attributes != nil {
		t.Fatalf("r.Attributes = %v, want nil", r.Attributes)
	}
	if r.RootClass() != "" {
		t.Fatalf("r.RootClass() = %q, want %q", r.RootClass(), "")
	}
	if got := len(r.Options()); got != 1 {
		t.Fatalf("len(r.Options()) = %d, want 1", got)
	}
}

func TestPageRenderWithActiveTheme(t *testing.T) {
	ctx := context.Background()
	svc, db := newThemeTestService(t, &fakeRegistry{})

	c, err := svc.Activate(ctx, "acme/minimal/0.1.0")
	if err != nil {
		t.Fatalf("Activate: %v", err)
	}
	page := bindPage(t, db, c.Contract.ID, "post")

	r, err := svc.PageRender(page)
	if err != nil {
		t.Fatalf("PageRender: %v", err)
	}
	if r.LayoutID != "post" {
		t.Fatalf("r.LayoutID = %q, want %q", r.LayoutID, "post")
	}
	if r.Release != c.Release {
		t.Fatalf("r.Release = %q, want %q", r.Release, c.Release)
	}
	if r.Attributes == nil {
		t.Fatalf("r.Attributes = nil, want non-nil")
	}
	if r.RootClass() != "layout-post" {
		t.Fatalf("r.RootClass() = %q, want %q", r.RootClass(), "layout-post")
	}
	if got := len(r.Options()); got != 2 {
		t.Fatalf("len(r.Options()) = %d, want 2", got)
	}
}

func TestPageRenderUsesTheActiveThemesRecord(t *testing.T) {
	ctx := context.Background()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.Contract{}, &models.SitePage{}, &models.Option{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
	themesDir := writeTestThemeWith(t, contractVersion("v1.1"))
	svc := NewThemeService(themesDir, t.TempDir(), NewContractService(repository.NewContractRepository(db)), repository.NewThemeRepository(db), &fakeRegistry{})

	c, err := svc.Activate(ctx, "acme/minimal/0.1.0")
	if err != nil {
		t.Fatalf("Activate: %v", err)
	}

	old := models.Contract{
		Org:      "example",
		Name:     "blog",
		Major:    1,
		Minor:    0,
		Source:   "not a contract",
		Checksum: "x",
	}
	if err := db.Create(&old).Error; err != nil {
		t.Fatalf("create old contract: %v", err)
	}
	page := bindPage(t, db, old.ID, "post")

	r, err := svc.PageRender(page)
	if err != nil {
		t.Fatalf("PageRender: %v", err)
	}
	if r.Release != c.Release {
		t.Fatalf("r.Release = %q, want %q", r.Release, c.Release)
	}
	if r.LayoutID != "post" {
		t.Fatalf("r.LayoutID = %q, want %q", r.LayoutID, "post")
	}
}

func TestPageRenderWithBrokenTheme(t *testing.T) {
	ctx := context.Background()

	db, themesDir, _ := activated(t)
	breakLicense(t, themesDir)
	svc, _ := restart(t, db, themesDir)
	if err := svc.LoadActive(ctx); err != nil {
		t.Fatalf("LoadActive: %v", err)
	}

	var rec models.Contract
	if err := db.Where("org = ? AND name = ? AND major = ? AND minor = ?", "example", "blog", 1, 0).First(&rec).Error; err != nil {
		t.Fatalf("load contract: %v", err)
	}
	page := bindPage(t, db, rec.ID, "post")

	r, err := svc.PageRender(page)
	if err != nil {
		t.Fatalf("PageRender: %v", err)
	}
	if r.LayoutID != "post" {
		t.Fatalf("r.LayoutID = %q, want %q", r.LayoutID, "post")
	}
	if r.Release != "" {
		t.Fatalf("r.Release = %q, want %q", r.Release, "")
	}
	if r.Attributes != nil {
		t.Fatalf("r.Attributes = %v, want nil", r.Attributes)
	}
}

func TestPageRenderLayoutNotFound(t *testing.T) {
	svc, db := newThemeTestService(t, &fakeRegistry{})
	rec := storeBlogContract(t, db)
	page := bindPage(t, db, rec.ID, "nope")

	_, err := svc.PageRender(page)
	if !errors.Is(err, ErrLayoutNotFound) {
		t.Fatalf("PageRender error = %v, want ErrLayoutNotFound", err)
	}
}

func TestPageRenderOptionsRenderTheLayout(t *testing.T) {
	ctx := context.Background()
	svc, db := newThemeTestService(t, &fakeRegistry{})

	c, err := svc.Activate(ctx, "acme/minimal/0.1.0")
	if err != nil {
		t.Fatalf("Activate: %v", err)
	}
	page := bindPage(t, db, c.Contract.ID, "post")

	r, err := svc.PageRender(page)
	if err != nil {
		t.Fatalf("PageRender: %v", err)
	}

	var buf bytes.Buffer
	if _, err := compono.New().Convert([]byte(`{{ UMONO_LAYOUT title = "Hello" }}`), &buf, r.Options()...); err != nil {
		t.Fatalf("Convert: %v", err)
	}
	if !strings.Contains(buf.String(), "Hello") {
		t.Fatalf("buf = %q, want contains %q", buf.String(), "Hello")
	}
}
