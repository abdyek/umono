package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"reflect"
	"testing"

	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/theme"
	"gorm.io/gorm"
)

func activeThemes(t *testing.T, db *gorm.DB) string {
	t.Helper()
	var rows []models.Option
	if err := db.Where("key = ?", "active_themes").Limit(1).Find(&rows).Error; err != nil {
		t.Fatalf("query active_themes: %v", err)
	}
	if len(rows) == 0 {
		return ""
	}
	return rows[0].Value
}

func countContracts(t *testing.T, db *gorm.DB) int64 {
	t.Helper()
	var n int64
	if err := db.Model(&models.Contract{}).Count(&n).Error; err != nil {
		t.Fatalf("count contracts: %v", err)
	}
	return n
}

func storeBlogContract(t *testing.T, db *gorm.DB) models.Contract {
	t.Helper()
	src := contractSource(t)
	sum := sha256.Sum256(src)
	rec := models.Contract{
		Org:      "example",
		Name:     "blog",
		Major:    1,
		Minor:    0,
		Source:   string(src),
		Checksum: hex.EncodeToString(sum[:]),
	}
	if err := db.Create(&rec).Error; err != nil {
		t.Fatalf("create contract: %v", err)
	}
	return rec
}

func TestThemeServiceActivate(t *testing.T) {
	ctx := context.Background()
	reg := &fakeRegistry{}
	svc, db := newThemeTestService(t, reg)

	c, err := svc.Activate(ctx, "acme/minimal/0.1.0")
	if err != nil {
		t.Fatalf("Activate: %v", err)
	}
	if n := countContracts(t, db); n != 1 {
		t.Fatalf("countContracts = %d, want 1", n)
	}
	if c.Contract.ID == 0 {
		t.Fatalf("c.Contract.ID = 0, want non-zero")
	}
	var stored models.Contract
	if err := db.First(&stored, c.Contract.ID).Error; err != nil {
		t.Fatalf("load contract: %v", err)
	}
	if stored.Source != string(contractSource(t)) {
		t.Fatalf("stored.Source = %q, want %q", stored.Source, string(contractSource(t)))
	}
	wantThemes := `[{"path":"acme/minimal/0.1.0","contract_id":"example/blog@v1.0"}]`
	if got := activeThemes(t, db); got != wantThemes {
		t.Fatalf("activeThemes = %q, want %q", got, wantThemes)
	}
	blog := models.Contract{Org: "example", Name: "blog", Major: 1, Minor: 0}.Group()
	l, ok := svc.Loaded(blog)
	if !ok {
		t.Fatalf("svc.Loaded(blog) = false, want true")
	}
	if l.Path != "acme/minimal/0.1.0" {
		t.Fatalf("l.Path = %q, want %q", l.Path, "acme/minimal/0.1.0")
	}
	if l.Release != c.Release {
		t.Fatalf("l.Release = %q, want %q", l.Release, c.Release)
	}
	if l.ContractID != c.Contract.ID {
		t.Fatalf("l.ContractID = %d, want %d", l.ContractID, c.Contract.ID)
	}
	if l.Contract.ID != "example/blog" {
		t.Fatalf("l.Contract.ID = %q, want %q", l.Contract.ID, "example/blog")
	}
	wantAttrs := map[string]string{"class": "layout-post"}
	if got := l.Attributes.Root("post"); !reflect.DeepEqual(got, wantAttrs) {
		t.Fatalf("l.Attributes.Root(\"post\") = %v, want %v", got, wantAttrs)
	}
}

func TestThemeServiceActivateReplacesTheThemeOfTheGroup(t *testing.T) {
	ctx := context.Background()
	reg := &fakeRegistry{}
	svc, db := newThemeTestService(t, reg)

	rec := storeBlogContract(t, db)

	if err := db.Create(&models.Option{
		Key:   "active_themes",
		Value: `[{"path":"acme/old/0.1.0","contract_id":"example/blog@v1.0"},{"path":"jane/folio/1.0.0","contract_id":"umono/page@v0.1"}]`,
	}).Error; err != nil {
		t.Fatalf("create option: %v", err)
	}

	if _, err := svc.Activate(ctx, "acme/minimal/0.1.0"); err != nil {
		t.Fatalf("Activate: %v", err)
	}
	if reg.calls != 0 {
		t.Fatalf("reg.calls = %d, want 0", reg.calls)
	}
	if n := countContracts(t, db); n != 1 {
		t.Fatalf("countContracts = %d, want 1", n)
	}
	wantThemes := `[{"path":"acme/minimal/0.1.0","contract_id":"example/blog@v1.0"},{"path":"jane/folio/1.0.0","contract_id":"umono/page@v0.1"}]`
	if got := activeThemes(t, db); got != wantThemes {
		t.Fatalf("activeThemes = %q, want %q", got, wantThemes)
	}
	blog := models.Contract{Org: "example", Name: "blog", Major: 1, Minor: 0}.Group()
	l, ok := svc.Loaded(blog)
	if !ok {
		t.Fatalf("svc.Loaded(blog) = false, want true")
	}
	if l.ContractID != rec.ID {
		t.Fatalf("l.ContractID = %d, want %d", l.ContractID, rec.ID)
	}
}

func TestThemeServiceActivateWritesNothingWhenACheckFails(t *testing.T) {
	ctx := context.Background()
	reg := &fakeRegistry{err: theme.ErrContractNotPublished}
	svc, db := newThemeTestService(t, reg)

	_, err := svc.Activate(ctx, "acme/minimal/0.1.0")
	if err == nil {
		t.Fatalf("Activate error = nil, want non-nil")
	}
	if !errors.Is(err, theme.ErrContractNotPublished) {
		t.Fatalf("Activate error = %v, want theme.ErrContractNotPublished", err)
	}
	if n := countContracts(t, db); n != 0 {
		t.Fatalf("countContracts = %d, want 0", n)
	}
	if got := activeThemes(t, db); got != "" {
		t.Fatalf("activeThemes = %q, want %q", got, "")
	}
	blog := models.Contract{Org: "example", Name: "blog", Major: 1, Minor: 0}.Group()
	_, ok := svc.Loaded(blog)
	if ok {
		t.Fatalf("svc.Loaded(blog) = true, want false")
	}
}

func TestThemeServiceActivateIsOneTransaction(t *testing.T) {
	ctx := context.Background()
	reg := &fakeRegistry{}
	svc, db := newThemeTestService(t, reg)

	if err := db.Create(&models.Option{Key: "active_themes", Value: "not json"}).Error; err != nil {
		t.Fatalf("create option: %v", err)
	}

	_, err := svc.Activate(ctx, "acme/minimal/0.1.0")
	if err == nil {
		t.Fatalf("Activate error = nil, want non-nil")
	}
	if n := countContracts(t, db); n != 0 {
		t.Fatalf("countContracts = %d, want 0", n)
	}
	if got := activeThemes(t, db); got != "not json" {
		t.Fatalf("activeThemes = %q, want %q", got, "not json")
	}
	blog := models.Contract{Org: "example", Name: "blog", Major: 1, Minor: 0}.Group()
	_, ok := svc.Loaded(blog)
	if ok {
		t.Fatalf("svc.Loaded(blog) = true, want false")
	}
}
