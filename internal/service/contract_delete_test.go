package service

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/umono-cms/umono/internal/models"
	"gorm.io/gorm"
)

func storePageContract(t *testing.T, db *gorm.DB) models.Contract {
	t.Helper()
	rec := models.Contract{
		Org:      "umono",
		Name:     "page",
		Major:    0,
		Minor:    1,
		Source:   "s",
		Checksum: "x",
	}
	if err := db.Create(&rec).Error; err != nil {
		t.Fatalf("create page contract: %v", err)
	}
	return rec
}

func contractExists(t *testing.T, db *gorm.DB, id uint) bool {
	t.Helper()
	var n int64
	if err := db.Model(&models.Contract{}).Where("id = ?", id).Count(&n).Error; err != nil {
		t.Fatalf("count contract: %v", err)
	}
	return n > 0
}

func TestDeleteContract(t *testing.T) {
	svc, db, _, _ := newPendingTestService(t, &fakeRegistry{}, nil)
	rec := storePageContract(t, db)
	page := models.Contract{Org: "umono", Name: "page", Major: 0, Minor: 1}
	if err := svc.DeleteContract(page); err != nil {
		t.Fatalf("DeleteContract: %v", err)
	}
	if contractExists(t, db, rec.ID) {
		t.Fatalf("contract %d still exists", rec.ID)
	}
}

func TestDeleteContractWithAnInstalledTheme(t *testing.T) {
	svc, db, _, _ := newPendingTestService(t, &fakeRegistry{}, nil)
	rec := storeBlogContract(t, db)
	blog := models.Contract{Org: "example", Name: "blog", Major: 1, Minor: 0}
	if !errors.Is(svc.DeleteContract(blog), ErrContractHasThemes) {
		t.Fatalf("DeleteContract err = %v, want ErrContractHasThemes", svc.DeleteContract(blog))
	}
	if !contractExists(t, db, rec.ID) {
		t.Fatalf("contract %d removed", rec.ID)
	}
}

func TestDeleteContractWithABoundPage(t *testing.T) {
	svc, db, _, _ := newPendingTestService(t, &fakeRegistry{}, nil)
	rec := storePageContract(t, db)
	layout := "page"
	page := models.SitePage{Name: "P", Slug: "p", Enabled: false, ContractID: &rec.ID, LayoutID: &layout}
	if err := db.Create(&page).Error; err != nil {
		t.Fatalf("create site page: %v", err)
	}
	ref := models.Contract{Org: "umono", Name: "page", Major: 0, Minor: 1}
	if !errors.Is(svc.DeleteContract(ref), ErrContractHasPages) {
		t.Fatalf("DeleteContract err = %v, want ErrContractHasPages", svc.DeleteContract(ref))
	}
	if !contractExists(t, db, rec.ID) {
		t.Fatalf("contract %d removed", rec.ID)
	}
}

func TestDeleteContractWithoutRecord(t *testing.T) {
	svc, _, _, _ := newPendingTestService(t, &fakeRegistry{}, nil)
	page := models.Contract{Org: "umono", Name: "page", Major: 0, Minor: 1}
	if !errors.Is(svc.DeleteContract(page), ErrContractNotFound) {
		t.Fatalf("DeleteContract err = %v, want ErrContractNotFound", svc.DeleteContract(page))
	}
}

func TestDeleteContractWithAnUnreadableTheme(t *testing.T) {
	svc, db, themesDir, _ := newPendingTestService(t, &fakeRegistry{}, nil)
	rec := storePageContract(t, db)
	brokenDir := filepath.Join(themesDir, "acme", "broken", "1.0.0")
	if err := os.MkdirAll(brokenDir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", brokenDir, err)
	}
	if err := os.WriteFile(filepath.Join(brokenDir, "built.toml"), []byte("not toml"), 0o644); err != nil {
		t.Fatalf("write broken built.toml: %v", err)
	}
	page := models.Contract{Org: "umono", Name: "page", Major: 0, Minor: 1}
	if err := svc.DeleteContract(page); err == nil {
		t.Fatalf("DeleteContract err = nil, want non-nil")
	}
	if !contractExists(t, db, rec.ID) {
		t.Fatalf("contract %d removed", rec.ID)
	}
}
