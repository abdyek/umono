package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/umono-cms/umono/internal/models"
	"gorm.io/gorm"
)

func createNotFoundSystemPage(t *testing.T, db *gorm.DB, content string) models.SitePage {
	t.Helper()
	key := models.NotFoundSystemKey
	p := models.SitePage{Name: "Lost", Slug: models.NotFoundSlug, Content: content, Enabled: true, SystemKey: &key}
	if err := db.Create(&p).Error; err != nil {
		t.Fatalf("create not found system page: %v", err)
	}
	return p
}

func TestSystemPageIsNotResolvedBySlug(t *testing.T) {
	svc, _, db := newPageDiagnosticTestService(t)

	createNotFoundSystemPage(t, db, "# Gone")

	if _, err := svc.GetBySlug(models.NotFoundSlug); !errors.Is(err, ErrSitePageNotFound) {
		t.Fatalf("GetBySlug err = %v, want %v", err, ErrSitePageNotFound)
	}

	if _, err := svc.GetRenderedBySlug(context.Background(), models.NotFoundSlug); !errors.Is(err, ErrSitePageNotFound) {
		t.Fatalf("GetRenderedBySlug err = %v, want %v", err, ErrSitePageNotFound)
	}
}

func TestSystemPageIsNotListed(t *testing.T) {
	svc, _, db := newPageDiagnosticTestService(t)

	createNotFoundSystemPage(t, db, "# Gone")

	if _, errs := svc.Create(models.SitePage{Name: "About", Slug: "about", Enabled: true, Content: "# About"}); len(errs) > 0 {
		t.Fatalf("Create errors: %v", errs)
	}

	items, _ := svc.GetPage(0)
	if got, want := len(items), 1; got != want {
		t.Fatalf("GetPage len = %d, want %d", got, want)
	}
	if got, want := items[0].Slug, "about"; got != want {
		t.Fatalf("items[0].Slug = %q, want %q", got, want)
	}

	body, err := svc.Sitemap("https://example.com")
	if err != nil {
		t.Fatalf("Sitemap err = %v, want nil", err)
	}
	if got, want := strings.Contains(string(body), "https://example.com/about"), true; got != want {
		t.Fatalf("strings.Contains(body, %q) = %v, want %v", "https://example.com/about", got, want)
	}
	if got, want := strings.Contains(string(body), models.NotFoundSlug), false; got != want {
		t.Fatalf("strings.Contains(body, %q) = %v, want %v", models.NotFoundSlug, got, want)
	}
}

func TestSystemPageCannotBeOpenedAsAPage(t *testing.T) {
	svc, _, db := newPageDiagnosticTestService(t)

	p := createNotFoundSystemPage(t, db, "# Gone")

	if _, err := svc.GetByID(p.ID); !errors.Is(err, ErrSitePageNotFound) {
		t.Fatalf("GetByID err = %v, want %v", err, ErrSitePageNotFound)
	}
}

func TestSystemPageCannotBeDeleted(t *testing.T) {
	svc, _, db := newPageDiagnosticTestService(t)

	p := createNotFoundSystemPage(t, db, "# Gone")

	if err := svc.Delete(p.ID); !errors.Is(err, ErrSystemPage) {
		t.Fatalf("Delete err = %v, want %v", err, ErrSystemPage)
	}

	var count int64
	if err := db.Model(&models.SitePage{}).Where("id = ?", p.ID).Count(&count).Error; err != nil {
		t.Fatalf("count page: %v", err)
	}
	if got, want := count, int64(1); got != want {
		t.Fatalf("page count = %d, want %d", got, want)
	}
}

func TestSystemPageCannotBeUpdatedAsAPage(t *testing.T) {
	svc, _, db := newPageDiagnosticTestService(t)

	p := createNotFoundSystemPage(t, db, "# Gone")

	_, errs := svc.Update(models.SitePage{ID: p.ID, Name: "X", Slug: "x", Enabled: false, Content: "y"})
	if got, want := len(errs), 1; got != want {
		t.Fatalf("Update errs len = %d, want %d", got, want)
	}
	if !errors.Is(errs[0], ErrSystemPage) {
		t.Fatalf("Update errs[0] = %v, want %v", errs[0], ErrSystemPage)
	}

	var got models.SitePage
	if err := db.First(&got, p.ID).Error; err != nil {
		t.Fatalf("First: %v", err)
	}
	if got.Name != "Lost" {
		t.Fatalf("got.Name = %q, want %q", got.Name, "Lost")
	}
	if got.Slug != models.NotFoundSlug {
		t.Fatalf("got.Slug = %q, want %q", got.Slug, models.NotFoundSlug)
	}
	if !got.Enabled {
		t.Fatalf("got.Enabled = false, want true")
	}
	if got.Content != "# Gone" {
		t.Fatalf("got.Content = %q, want %q", got.Content, "# Gone")
	}
	if got.SystemKey == nil {
		t.Fatalf("got.SystemKey = nil, want non-nil")
	}
	if *got.SystemKey != models.NotFoundSystemKey {
		t.Fatalf("*got.SystemKey = %q, want %q", *got.SystemKey, models.NotFoundSystemKey)
	}
}

func TestSystemPageDiagnosticsArePublished(t *testing.T) {
	svc, _, db := newPageDiagnosticTestService(t)

	p := createNotFoundSystemPage(t, db, "{{ MISSING }}")

	if err := svc.RecomputeAllDiagnostics(context.Background()); err != nil {
		t.Fatalf("RecomputeAllDiagnostics: %v", err)
	}

	d := svc.PublishedDiagnostics()
	if got, want := len(d), 1; got != want {
		t.Fatalf("PublishedDiagnostics len = %d, want %d", got, want)
	}
	if got, want := d[0].PageID, p.ID; got != want {
		t.Fatalf("d[0].PageID = %d, want %d", got, want)
	}
	if d[0].SystemKey == nil {
		t.Fatalf("d[0].SystemKey = nil, want non-nil")
	}
	if *d[0].SystemKey != models.NotFoundSystemKey {
		t.Fatalf("*d[0].SystemKey = %q, want %q", *d[0].SystemKey, models.NotFoundSystemKey)
	}
}

func TestReservedSlugIsInvalid(t *testing.T) {
	svc, _, _ := newPageDiagnosticTestService(t)

	if err := svc.CheckSlug(models.NotFoundSlug, 0); !errors.Is(err, ErrInvalidSlug) {
		t.Fatalf("CheckSlug err = %v, want %v", err, ErrInvalidSlug)
	}

	_, errs := svc.Create(models.SitePage{Name: "N", Slug: models.NotFoundSlug, Enabled: true})
	if len(errs) == 0 {
		t.Fatalf("Create errs is empty, want at least one error")
	}
	if !ErrorsIs(errs, ErrInvalidSlug) {
		t.Fatalf("Create errs = %v, want one matching %v", errs, ErrInvalidSlug)
	}
}
