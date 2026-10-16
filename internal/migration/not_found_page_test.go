package migration

import (
	"testing"

	"github.com/umono-cms/umono/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newNotFoundTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&models.Option{}, &models.Contract{}, &models.SitePage{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func loadNotFoundPages(t *testing.T, db *gorm.DB) []models.SitePage {
	t.Helper()
	var pages []models.SitePage
	if err := db.Where("system_key = ?", models.NotFoundSystemKey).Find(&pages).Error; err != nil {
		t.Fatalf("load not found pages: %v", err)
	}
	return pages
}

func TestNotFoundPageMovesTheOption(t *testing.T) {
	db := newNotFoundTestDB(t)

	if err := db.Create(&models.Option{
		Key:   "not_found_page_option",
		Value: `{"title":"Lost","content":"# Gone"}`,
	}).Error; err != nil {
		t.Fatalf("create option: %v", err)
	}

	id, err := NotFoundPage(db)
	if err != nil {
		t.Fatalf("NotFoundPage: %v", err)
	}

	pages := loadNotFoundPages(t, db)
	if len(pages) != 1 {
		t.Fatalf("not found pages count: got %d, want 1", len(pages))
	}
	p := pages[0]

	if id != p.ID {
		t.Errorf("id: got %d, want %d", id, p.ID)
	}
	if id == 0 {
		t.Errorf("id is 0")
	}
	if p.Name != "Lost" {
		t.Errorf("Name: got %q, want %q", p.Name, "Lost")
	}
	if p.Content != "# Gone" {
		t.Errorf("Content: got %q, want %q", p.Content, "# Gone")
	}
	if p.Slug != models.NotFoundSlug {
		t.Errorf("Slug: got %q, want %q", p.Slug, models.NotFoundSlug)
	}
	if !p.Enabled {
		t.Errorf("Enabled: got false, want true")
	}
	if p.LastModifiedAt == nil {
		t.Errorf("LastModifiedAt is nil")
	}
	if p.ContractID != nil {
		t.Errorf("ContractID: got %v, want nil", *p.ContractID)
	}
	if p.LayoutID != nil {
		t.Errorf("LayoutID: got %v, want nil", *p.LayoutID)
	}

	var optionCount int64
	if err := db.Model(&models.Option{}).Where("key = ?", "not_found_page_option").Count(&optionCount).Error; err != nil {
		t.Fatalf("count options: %v", err)
	}
	if optionCount != 0 {
		t.Errorf("option count: got %d, want 0", optionCount)
	}
}

func TestNotFoundPageWithoutOptionIsEmpty(t *testing.T) {
	db := newNotFoundTestDB(t)

	id, err := NotFoundPage(db)
	if err != nil {
		t.Fatalf("NotFoundPage: %v", err)
	}
	if id == 0 {
		t.Errorf("id is 0")
	}

	pages := loadNotFoundPages(t, db)
	if len(pages) != 1 {
		t.Fatalf("not found pages count: got %d, want 1", len(pages))
	}
	p := pages[0]

	if p.ID != id {
		t.Errorf("ID: got %d, want %d", p.ID, id)
	}
	if p.Name != "" {
		t.Errorf("Name: got %q, want empty", p.Name)
	}
	if p.Content != "" {
		t.Errorf("Content: got %q, want empty", p.Content)
	}
	if p.Slug != models.NotFoundSlug {
		t.Errorf("Slug: got %q, want %q", p.Slug, models.NotFoundSlug)
	}
	if !p.Enabled {
		t.Errorf("Enabled: got false, want true")
	}
}

func TestNotFoundPageRunsOnce(t *testing.T) {
	db := newNotFoundTestDB(t)

	if _, err := NotFoundPage(db); err != nil {
		t.Fatalf("first NotFoundPage: %v", err)
	}

	if err := db.Create(&models.Option{
		Key:   "not_found_page_option",
		Value: `{"title":"Again","content":"x"}`,
	}).Error; err != nil {
		t.Fatalf("create option: %v", err)
	}

	id, err := NotFoundPage(db)
	if err != nil {
		t.Fatalf("second NotFoundPage: %v", err)
	}
	if id != 0 {
		t.Errorf("id: got %d, want 0", id)
	}

	pages := loadNotFoundPages(t, db)
	if len(pages) != 1 {
		t.Fatalf("not found pages count: got %d, want 1", len(pages))
	}
	if pages[0].Name != "" {
		t.Errorf("Name: got %q, want empty", pages[0].Name)
	}

	var optionCount int64
	if err := db.Model(&models.Option{}).Where("key = ?", "not_found_page_option").Count(&optionCount).Error; err != nil {
		t.Fatalf("count options: %v", err)
	}
	if optionCount != 1 {
		t.Errorf("option count: got %d, want 1", optionCount)
	}
}

func TestNotFoundPageKeepsOrdinaryPages(t *testing.T) {
	db := newNotFoundTestDB(t)

	if err := db.Create(&models.SitePage{
		Name:    "Home",
		Slug:    "",
		Content: "# Hi",
		Enabled: true,
	}).Error; err != nil {
		t.Fatalf("create home page: %v", err)
	}

	if _, err := NotFoundPage(db); err != nil {
		t.Fatalf("NotFoundPage: %v", err)
	}

	var total int64
	if err := db.Model(&models.SitePage{}).Count(&total).Error; err != nil {
		t.Fatalf("count pages: %v", err)
	}
	if total != 2 {
		t.Errorf("total pages: got %d, want 2", total)
	}

	var home models.SitePage
	if err := db.Where("slug = ?", "").First(&home).Error; err != nil {
		t.Fatalf("load home page: %v", err)
	}
	if home.Name != "Home" {
		t.Errorf("Name: got %q, want %q", home.Name, "Home")
	}
	if home.SystemKey != nil {
		t.Errorf("SystemKey: got %v, want nil", *home.SystemKey)
	}
}
