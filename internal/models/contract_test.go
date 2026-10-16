package models

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestContractIdentityIsUnique(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}
	if err := db.AutoMigrate(&Contract{}); err != nil {
		t.Fatalf("failed to auto-migrate Contract: %v", err)
	}

	if err := db.Create(&Contract{Org: "umono", Name: "page", Major: 0, Minor: 1, Source: "a", Checksum: "x"}).Error; err != nil {
		t.Fatalf("failed to create first contract: %v", err)
	}

	if err := db.Create(&Contract{Org: "umono", Name: "page", Major: 0, Minor: 1, Source: "b", Checksum: "y"}).Error; err == nil {
		t.Fatal("expected error creating duplicate identity contract, got nil")
	}

	if err := db.Create(&Contract{Org: "umono", Name: "page", Major: 0, Minor: 2, Source: "a", Checksum: "x"}).Error; err != nil {
		t.Fatalf("failed to create contract with different Minor: %v", err)
	}
	if err := db.Create(&Contract{Org: "umono", Name: "page", Major: 1, Minor: 1, Source: "a", Checksum: "x"}).Error; err != nil {
		t.Fatalf("failed to create contract with different Major: %v", err)
	}
	if err := db.Create(&Contract{Org: "acme", Name: "page", Major: 0, Minor: 1, Source: "a", Checksum: "x"}).Error; err != nil {
		t.Fatalf("failed to create contract with different Org: %v", err)
	}

	var count int64
	if err := db.Model(&Contract{}).Count(&count).Error; err != nil {
		t.Fatalf("failed to count contracts: %v", err)
	}
	if count != 4 {
		t.Fatalf("expected 4 contracts, got %d", count)
	}
}
