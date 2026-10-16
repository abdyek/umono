package models

import (
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type sitePageV08 struct {
	ID             uint `gorm:"primaryKey"`
	Name           string
	Slug           string `gorm:"uniqueIndex"`
	Content        string
	LastModifiedAt *time.Time
	Enabled        bool
}

func (sitePageV08) TableName() string { return "pages" }

func TestSitePageMigrationLeavesExistingPagesUnbound(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&sitePageV08{}); err != nil {
		t.Fatalf("automigrate v08: %v", err)
	}
	if err := db.Create(&sitePageV08{Name: "Home", Slug: "", Content: "# Hi", Enabled: true}).Error; err != nil {
		t.Fatalf("create v08 page: %v", err)
	}
	if err := db.AutoMigrate(&Contract{}, &SitePage{}); err != nil {
		t.Fatalf("automigrate contract and sitepage: %v", err)
	}
	var p SitePage
	if err := db.First(&p, "slug = ?", "").Error; err != nil {
		t.Fatalf("load page: %v", err)
	}
	if p.Name != "Home" {
		t.Fatalf("Name: got %q, want %q", p.Name, "Home")
	}
	if p.Content != "# Hi" {
		t.Fatalf("Content: got %q, want %q", p.Content, "# Hi")
	}
	if !p.Enabled {
		t.Fatalf("Enabled: got %v, want true", p.Enabled)
	}
	if p.ContractID != nil {
		t.Fatalf("ContractID: got %v, want nil", *p.ContractID)
	}
	if p.LayoutID != nil {
		t.Fatalf("LayoutID: got %v, want nil", *p.LayoutID)
	}
}

func TestSitePageBindsToContract(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&Contract{}, &SitePage{}); err != nil {
		t.Fatalf("automigrate contract and sitepage: %v", err)
	}
	c := Contract{Org: "umono", Name: "page", Major: 0, Minor: 1, Source: "s", Checksum: "x"}
	if err := db.Create(&c).Error; err != nil {
		t.Fatalf("create contract: %v", err)
	}
	layout := "page"
	if err := db.Create(&SitePage{Name: "About", Slug: "about", Enabled: true, ContractID: &c.ID, LayoutID: &layout}).Error; err != nil {
		t.Fatalf("create sitepage: %v", err)
	}
	var p SitePage
	if err := db.Preload("Contract").First(&p, "slug = ?", "about").Error; err != nil {
		t.Fatalf("load page: %v", err)
	}
	if p.ContractID == nil {
		t.Fatalf("ContractID: got nil, want non-nil")
	}
	if *p.ContractID != c.ID {
		t.Fatalf("ContractID: got %d, want %d", *p.ContractID, c.ID)
	}
	if p.LayoutID == nil {
		t.Fatalf("LayoutID: got nil, want non-nil")
	}
	if *p.LayoutID != "page" {
		t.Fatalf("LayoutID: got %q, want %q", *p.LayoutID, "page")
	}
	if p.Contract == nil {
		t.Fatalf("Contract: got nil, want non-nil")
	}
	if p.Contract.Org != "umono" {
		t.Fatalf("Contract.Org: got %q, want %q", p.Contract.Org, "umono")
	}
	if p.Contract.Name != "page" {
		t.Fatalf("Contract.Name: got %q, want %q", p.Contract.Name, "page")
	}
}
