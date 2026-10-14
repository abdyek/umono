package service

import (
	"context"
	"testing"

	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newPageDiagnosticTestService(t *testing.T) (*SitePageService, *ContentCompiler, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.SitePage{}, &models.Component{}, &models.PageDiagnostic{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	cc, err := NewContentCompiler(nil)
	if err != nil {
		t.Fatalf("new content compiler: %v", err)
	}

	svc := NewSitePageService(repository.NewSitePageRepository(db), nil, cc)
	svc.SetPageDiagnosticRepository(repository.NewPageDiagnosticRepository(db))

	return svc, cc, db
}

func countPageDiagnostics(t *testing.T, db *gorm.DB) int64 {
	t.Helper()
	var n int64
	if err := db.Model(&models.PageDiagnostic{}).Count(&n).Error; err != nil {
		t.Fatalf("count diagnostics: %v", err)
	}
	return n
}

func TestSitePageServiceCreateStoresDiagnostics(t *testing.T) {
	svc, _, _ := newPageDiagnosticTestService(t)

	created, errs := svc.Create(models.SitePage{
		Name:    "About",
		Slug:    "about",
		Enabled: true,
		Content: "{{ MISSING }}",
	})
	if len(errs) > 0 {
		t.Fatalf("Create errors: %v", errs)
	}

	diagnostics := svc.PublishedDiagnostics()
	if got, want := len(diagnostics), 1; got != want {
		t.Fatalf("PublishedDiagnostics len = %d, want %d", got, want)
	}
	if got, want := diagnostics[0].PageID, created.ID; got != want {
		t.Fatalf("diagnostics[0].PageID = %d, want %d", got, want)
	}
	if got, want := diagnostics[0].PageName, "About"; got != want {
		t.Fatalf("diagnostics[0].PageName = %q, want %q", got, want)
	}
	if got, want := diagnostics[0].Code, "unknown-component"; got != want {
		t.Fatalf("diagnostics[0].Code = %q, want %q", got, want)
	}
	if got, want := diagnostics[0].Source, ""; got != want {
		t.Fatalf("diagnostics[0].Source = %q, want %q", got, want)
	}
	if diagnostics[0].Message == "" {
		t.Fatalf("diagnostics[0].Message is empty")
	}
}

func TestSitePageServiceUpdateReplacesDiagnostics(t *testing.T) {
	svc, _, db := newPageDiagnosticTestService(t)

	created, errs := svc.Create(models.SitePage{
		Name:    "About",
		Slug:    "about",
		Enabled: true,
		Content: "{{ MISSING }}",
	})
	if len(errs) > 0 {
		t.Fatalf("Create errors: %v", errs)
	}

	created.Content = "Hello"
	_, errs = svc.Update(created)
	if len(errs) > 0 {
		t.Fatalf("Update errors: %v", errs)
	}

	if got := len(svc.PublishedDiagnostics()); got != 0 {
		t.Fatalf("PublishedDiagnostics len = %d, want 0", got)
	}
	if got := countPageDiagnostics(t, db); got != 0 {
		t.Fatalf("countPageDiagnostics = %d, want 0", got)
	}
}

func TestSitePageServicePublishedDiagnosticsSkipsDisabledPages(t *testing.T) {
	svc, _, db := newPageDiagnosticTestService(t)

	_, errs := svc.Create(models.SitePage{
		Name:    "Draft",
		Slug:    "draft",
		Enabled: false,
		Content: "{{ MISSING }}",
	})
	if len(errs) > 0 {
		t.Fatalf("Create errors: %v", errs)
	}

	if got := len(svc.PublishedDiagnostics()); got != 0 {
		t.Fatalf("PublishedDiagnostics len = %d, want 0", got)
	}
	if got, want := countPageDiagnostics(t, db), int64(1); got != want {
		t.Fatalf("countPageDiagnostics = %d, want %d", got, want)
	}
}

func TestSitePageServiceDeleteRemovesDiagnostics(t *testing.T) {
	svc, _, db := newPageDiagnosticTestService(t)

	created, errs := svc.Create(models.SitePage{
		Name:    "About",
		Slug:    "about",
		Enabled: true,
		Content: "{{ MISSING }}",
	})
	if len(errs) > 0 {
		t.Fatalf("Create errors: %v", errs)
	}

	if err := svc.Delete(created.ID); err != nil {
		t.Fatalf("Delete error: %v", err)
	}

	if got := countPageDiagnostics(t, db); got != 0 {
		t.Fatalf("countPageDiagnostics = %d, want 0", got)
	}
}

func TestSitePageServiceRecomputeAllDiagnostics(t *testing.T) {
	svc, cc, _ := newPageDiagnosticTestService(t)

	_, errs := svc.Create(models.SitePage{
		Name:    "Home",
		Slug:    "",
		Enabled: true,
		Content: "{{ FOO }}",
	})
	if len(errs) > 0 {
		t.Fatalf("Create errors: %v", errs)
	}

	if got, want := len(svc.PublishedDiagnostics()), 1; got != want {
		t.Fatalf("PublishedDiagnostics len = %d, want %d", got, want)
	}

	if err := cc.LoadGlobalComponent(models.Component{Name: "FOO", Content: "Foo"}); err != nil {
		t.Fatalf("LoadGlobalComponent: %v", err)
	}

	if got, want := len(svc.PublishedDiagnostics()), 1; got != want {
		t.Fatalf("PublishedDiagnostics len after LoadGlobalComponent = %d, want %d", got, want)
	}

	if err := svc.RecomputeAllDiagnostics(context.Background()); err != nil {
		t.Fatalf("RecomputeAllDiagnostics: %v", err)
	}

	if got := len(svc.PublishedDiagnostics()); got != 0 {
		t.Fatalf("PublishedDiagnostics len after RecomputeAllDiagnostics = %d, want 0", got)
	}
}

func TestSitePageServiceDiagnosticSourceOfGlobal(t *testing.T) {
	svc, cc, _ := newPageDiagnosticTestService(t)

	if err := cc.LoadGlobalComponent(models.Component{Name: "FOO", Content: "{{ MISSING }}"}); err != nil {
		t.Fatalf("LoadGlobalComponent: %v", err)
	}

	_, errs := svc.Create(models.SitePage{
		Name:    "Home",
		Slug:    "",
		Enabled: true,
		Content: "{{ FOO }}",
	})
	if len(errs) > 0 {
		t.Fatalf("Create errors: %v", errs)
	}

	diagnostics := svc.PublishedDiagnostics()
	if got, want := len(diagnostics), 1; got != want {
		t.Fatalf("PublishedDiagnostics len = %d, want %d", got, want)
	}
	if got, want := diagnostics[0].Code, "unknown-component"; got != want {
		t.Fatalf("diagnostics[0].Code = %q, want %q", got, want)
	}
	if got, want := diagnostics[0].Source, "FOO"; got != want {
		t.Fatalf("diagnostics[0].Source = %q, want %q", got, want)
	}
}

func TestSitePageServicePublishedDiagnosticsOrder(t *testing.T) {
	svc, _, _ := newPageDiagnosticTestService(t)

	_, errs := svc.Create(models.SitePage{
		Name:    "Beta",
		Slug:    "beta",
		Enabled: true,
		Content: "{{ MISSING }}",
	})
	if len(errs) > 0 {
		t.Fatalf("Create Beta errors: %v", errs)
	}

	_, errs = svc.Create(models.SitePage{
		Name:    "Alpha",
		Slug:    "alpha",
		Enabled: true,
		Content: "{{ MISSING }}",
	})
	if len(errs) > 0 {
		t.Fatalf("Create Alpha errors: %v", errs)
	}

	diagnostics := svc.PublishedDiagnostics()
	if got, want := len(diagnostics), 2; got != want {
		t.Fatalf("PublishedDiagnostics len = %d, want %d", got, want)
	}
	if got, want := diagnostics[0].PageName, "Alpha"; got != want {
		t.Fatalf("diagnostics[0].PageName = %q, want %q", got, want)
	}
	if got, want := diagnostics[1].PageName, "Beta"; got != want {
		t.Fatalf("diagnostics[1].PageName = %q, want %q", got, want)
	}
}

func TestSitePageServiceFillDiagnosticsOnce(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.SitePage{}, &models.Component{}, &models.PageDiagnostic{}, &models.Option{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	cc, err := NewContentCompiler(nil)
	if err != nil {
		t.Fatalf("new content compiler: %v", err)
	}

	optRepo := repository.NewOptionRepository(db)
	svc := NewSitePageService(repository.NewSitePageRepository(db), optRepo, cc)
	svc.SetPageDiagnosticRepository(repository.NewPageDiagnosticRepository(db))

	if err := db.Create(&models.SitePage{Name: "Old", Slug: "old", Enabled: true, Content: "{{ MISSING }}"}).Error; err != nil {
		t.Fatalf("create Old page: %v", err)
	}

	if got, want := len(svc.PublishedDiagnostics()), 0; got != want {
		t.Fatalf("PublishedDiagnostics len = %d, want %d", got, want)
	}

	if err := svc.FillDiagnosticsOnce(context.Background()); err != nil {
		t.Fatalf("FillDiagnosticsOnce: %v", err)
	}

	if got, want := len(svc.PublishedDiagnostics()), 1; got != want {
		t.Fatalf("PublishedDiagnostics len = %d, want %d", got, want)
	}

	filledOpt := optRepo.GetOptionByKey("page_diagnostics_filled")
	if filledOpt.ID == 0 {
		t.Fatalf("optRepo.GetOptionByKey(\"page_diagnostics_filled\").ID = 0, want non-zero")
	}

	if err := db.Create(&models.SitePage{Name: "Later", Slug: "later", Enabled: true, Content: "{{ MISSING }}"}).Error; err != nil {
		t.Fatalf("create Later page: %v", err)
	}

	if err := svc.FillDiagnosticsOnce(context.Background()); err != nil {
		t.Fatalf("FillDiagnosticsOnce second call: %v", err)
	}

	if got, want := len(svc.PublishedDiagnostics()), 1; got != want {
		t.Fatalf("PublishedDiagnostics len = %d, want %d", got, want)
	}
}
