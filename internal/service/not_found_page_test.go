package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/umono-cms/umono/internal/models"
)

func TestNotFoundPageRendersTheSystemPage(t *testing.T) {
	svc, _, db := newPageDiagnosticTestService(t)
	createNotFoundSystemPage(t, db, "# Gone")

	ctx := context.Background()
	p, err := svc.GetNotFoundPage(ctx, "Default", "# Default")
	if err != nil {
		t.Fatalf("GetNotFoundPage err = %v, want nil", err)
	}
	if got, want := p.Name, "Lost"; got != want {
		t.Fatalf("p.Name = %q, want %q", got, want)
	}
	if got, want := strings.Contains(p.Content, "Gone"), true; got != want {
		t.Fatalf("strings.Contains(p.Content, %q) = %v, want %v", "Gone", got, want)
	}
	if got, want := strings.Contains(p.Content, "Default"), false; got != want {
		t.Fatalf("strings.Contains(p.Content, %q) = %v, want %v", "Default", got, want)
	}
}

func TestNotFoundPageFallsBackToDefaults(t *testing.T) {
	svc, _, db := newPageDiagnosticTestService(t)
	sp := createNotFoundSystemPage(t, db, "# Gone")

	if err := db.Model(&models.SitePage{}).Where("id = ?", sp.ID).Updates(map[string]any{"name": "", "content": ""}).Error; err != nil {
		t.Fatalf("update system page: %v", err)
	}

	ctx := context.Background()
	p, err := svc.GetNotFoundPage(ctx, "Default", "# Default")
	if err != nil {
		t.Fatalf("GetNotFoundPage err = %v, want nil", err)
	}
	if got, want := p.Name, "Default"; got != want {
		t.Fatalf("p.Name = %q, want %q", got, want)
	}
	if got, want := strings.Contains(p.Content, "Default"), true; got != want {
		t.Fatalf("strings.Contains(p.Content, %q) = %v, want %v", "Default", got, want)
	}
}

func TestSaveNotFoundPage(t *testing.T) {
	svc, _, db := newPageDiagnosticTestService(t)
	sp := createNotFoundSystemPage(t, db, "# Gone")

	if err := svc.SaveNotFoundPage("New", "{{ MISSING }}"); err != nil {
		t.Fatalf("SaveNotFoundPage err = %v, want nil", err)
	}

	got, err := svc.GetNotFoundPageSource()
	if err != nil {
		t.Fatalf("GetNotFoundPageSource err = %v, want nil", err)
	}
	if got, want := got.ID, sp.ID; got != want {
		t.Fatalf("got.ID = %d, want %d", got, want)
	}
	if got, want := got.Name, "New"; got != want {
		t.Fatalf("got.Name = %q, want %q", got, want)
	}
	if got, want := got.Content, "{{ MISSING }}"; got != want {
		t.Fatalf("got.Content = %q, want %q", got, want)
	}
	if got, want := got.Slug, models.NotFoundSlug; got != want {
		t.Fatalf("got.Slug = %q, want %q", got, want)
	}
	if !got.Enabled {
		t.Fatalf("got.Enabled = false, want true")
	}
	if got.LastModifiedAt == nil {
		t.Fatalf("got.LastModifiedAt = nil, want non-nil")
	}
	if got.SystemKey == nil {
		t.Fatalf("got.SystemKey = nil, want non-nil")
	}

	d := svc.PublishedDiagnostics()
	if got, want := len(d), 1; got != want {
		t.Fatalf("PublishedDiagnostics len = %d, want %d", got, want)
	}
	if got, want := d[0].PageID, sp.ID; got != want {
		t.Fatalf("d[0].PageID = %d, want %d", got, want)
	}
}

func TestSaveNotFoundPageClearsIt(t *testing.T) {
	svc, _, db := newPageDiagnosticTestService(t)
	createNotFoundSystemPage(t, db, "# Gone")

	if err := svc.SaveNotFoundPage("", ""); err != nil {
		t.Fatalf("SaveNotFoundPage err = %v, want nil", err)
	}

	got, err := svc.GetNotFoundPageSource()
	if err != nil {
		t.Fatalf("GetNotFoundPageSource err = %v, want nil", err)
	}
	if got, want := got.Name, ""; got != want {
		t.Fatalf("got.Name = %q, want %q", got, want)
	}
	if got, want := got.Content, ""; got != want {
		t.Fatalf("got.Content = %q, want %q", got, want)
	}
}

func TestSaveNotFoundPageWithoutSystemPage(t *testing.T) {
	svc, _, _ := newPageDiagnosticTestService(t)

	if err := svc.SaveNotFoundPage("A", "B"); !errors.Is(err, ErrSitePageNotFound) {
		t.Fatalf("SaveNotFoundPage err = %v, want %v", err, ErrSitePageNotFound)
	}

	if _, err := svc.GetNotFoundPageSource(); !errors.Is(err, ErrSitePageNotFound) {
		t.Fatalf("GetNotFoundPageSource err = %v, want %v", err, ErrSitePageNotFound)
	}
}
