package service

import (
	"context"
	"strings"
	"testing"

	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const boundContent = "{{ UMONO_LAYOUT title = \"Hello\" body = TEXT }}\n\n~ TEXT\nHi\n"

func newRenderTestServices(t *testing.T) (*SitePageService, *ThemeService, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.SitePage{}, &models.Component{}, &models.PageDiagnostic{}, &models.Contract{}, &models.Option{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}

	cc, err := NewContentCompiler(nil)
	if err != nil {
		t.Fatalf("new content compiler: %v", err)
	}

	sps := NewSitePageService(repository.NewSitePageRepository(db), repository.NewOptionRepository(db), cc)
	sps.SetPageDiagnosticRepository(repository.NewPageDiagnosticRepository(db))

	themes := NewThemeService(writeTestTheme(t), t.TempDir(), NewContractService(repository.NewContractRepository(db)), repository.NewThemeRepository(db), &fakeRegistry{})
	sps.SetPageThemes(themes)

	return sps, themes, db
}

func createPage(t *testing.T, db *gorm.DB, page models.SitePage) models.SitePage {
	t.Helper()
	if err := db.Create(&page).Error; err != nil {
		t.Fatalf("create page: %v", err)
	}
	return page
}

func TestRenderedPageWithActiveTheme(t *testing.T) {
	ctx := context.Background()
	sps, themes, db := newRenderTestServices(t)

	c, err := themes.Activate(ctx, "acme/minimal/0.1.0")
	if err != nil {
		t.Fatalf("Activate: %v", err)
	}
	layout := "post"
	createPage(t, db, models.SitePage{Name: "P", Slug: "p", Enabled: true, Content: boundContent, ContractID: &c.Contract.ID, LayoutID: &layout})

	r, err := sps.GetRenderedBySlug(ctx, "p")
	if err != nil {
		t.Fatalf("GetRenderedBySlug: %v", err)
	}
	if got, want := strings.Contains(r.Content, "Hello"), true; got != want {
		t.Fatalf("strings.Contains(r.Content, %q) = %v, want %v", "Hello", got, want)
	}
	if got, want := strings.Contains(r.Content, "Hi"), true; got != want {
		t.Fatalf("strings.Contains(r.Content, %q) = %v, want %v", "Hi", got, want)
	}
	if got, want := r.Release, c.Release; got != want {
		t.Fatalf("r.Release = %q, want %q", got, want)
	}
	if got, want := r.RootClass, "layout-post"; got != want {
		t.Fatalf("r.RootClass = %q, want %q", got, want)
	}
	if got, want := r.Name, "P"; got != want {
		t.Fatalf("r.Name = %q, want %q", got, want)
	}
}

func TestRenderedPageWithoutActiveTheme(t *testing.T) {
	ctx := context.Background()
	sps, _, db := newRenderTestServices(t)
	rec := storeBlogContract(t, db)
	layout := "post"
	createPage(t, db, models.SitePage{Name: "P", Slug: "p", Enabled: true, Content: boundContent, ContractID: &rec.ID, LayoutID: &layout})

	r, err := sps.GetRenderedBySlug(ctx, "p")
	if err != nil {
		t.Fatalf("GetRenderedBySlug: %v", err)
	}
	if got, want := strings.Contains(r.Content, "Hello"), true; got != want {
		t.Fatalf("strings.Contains(r.Content, %q) = %v, want %v", "Hello", got, want)
	}
	if got, want := r.Release, ""; got != want {
		t.Fatalf("r.Release = %q, want %q", got, want)
	}
	if got, want := r.RootClass, ""; got != want {
		t.Fatalf("r.RootClass = %q, want %q", got, want)
	}
}

func TestRenderedPageWithoutBinding(t *testing.T) {
	ctx := context.Background()
	sps, _, db := newRenderTestServices(t)
	createPage(t, db, models.SitePage{Name: "A", Slug: "a", Enabled: true, Content: "# Plain"})

	r, err := sps.GetRenderedBySlug(ctx, "a")
	if err != nil {
		t.Fatalf("GetRenderedBySlug: %v", err)
	}
	if got, want := strings.Contains(r.Content, "Plain"), true; got != want {
		t.Fatalf("strings.Contains(r.Content, %q) = %v, want %v", "Plain", got, want)
	}
	if got, want := r.Release, ""; got != want {
		t.Fatalf("r.Release = %q, want %q", got, want)
	}
	if got, want := r.RootClass, ""; got != want {
		t.Fatalf("r.RootClass = %q, want %q", got, want)
	}
}

func TestDiagnosticsUseTheLayout(t *testing.T) {
	ctx := context.Background()
	sps, themes, db := newRenderTestServices(t)
	c, err := themes.Activate(ctx, "acme/minimal/0.1.0")
	if err != nil {
		t.Fatalf("Activate: %v", err)
	}
	layout := "post"
	bound := createPage(t, db, models.SitePage{Name: "B", Slug: "b", Enabled: true, Content: boundContent, ContractID: &c.Contract.ID, LayoutID: &layout})
	unbound := createPage(t, db, models.SitePage{Name: "U", Slug: "u", Enabled: true, Content: boundContent})

	if err := sps.RecomputeDiagnostics(ctx, bound.ID); err != nil {
		t.Fatalf("RecomputeDiagnostics bound: %v", err)
	}
	if err := sps.RecomputeDiagnostics(ctx, unbound.ID); err != nil {
		t.Fatalf("RecomputeDiagnostics unbound: %v", err)
	}

	d := sps.PublishedDiagnostics()
	if len(d) == 0 {
		t.Fatalf("len(PublishedDiagnostics) = 0, want at least 1")
	}
	for _, item := range d {
		if got, want := item.PageID, unbound.ID; got != want {
			t.Fatalf("d[].PageID = %d, want %d", got, want)
		}
	}
}

func TestNotFoundPageWithLayout(t *testing.T) {
	ctx := context.Background()
	sps, themes, db := newRenderTestServices(t)
	c, err := themes.Activate(ctx, "acme/minimal/0.1.0")
	if err != nil {
		t.Fatalf("Activate: %v", err)
	}
	key := models.NotFoundSystemKey
	layout := "post"
	createPage(t, db, models.SitePage{Name: "Lost", Slug: models.NotFoundSlug, Enabled: true, Content: boundContent, SystemKey: &key, ContractID: &c.Contract.ID, LayoutID: &layout})

	r, err := sps.GetNotFoundPage(ctx, "Default", "# Default")
	if err != nil {
		t.Fatalf("GetNotFoundPage: %v", err)
	}
	if got, want := strings.Contains(r.Content, "Hello"), true; got != want {
		t.Fatalf("strings.Contains(r.Content, %q) = %v, want %v", "Hello", got, want)
	}
	if got, want := r.Release, c.Release; got != want {
		t.Fatalf("r.Release = %q, want %q", got, want)
	}
	if got, want := r.RootClass, "layout-post"; got != want {
		t.Fatalf("r.RootClass = %q, want %q", got, want)
	}
	if got, want := r.Name, "Lost"; got != want {
		t.Fatalf("r.Name = %q, want %q", got, want)
	}
}
